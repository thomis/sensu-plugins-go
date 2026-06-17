package main

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/thomis/sensu-plugins-go/pkg/check"
	"github.com/thomis/sensu-plugins-go/pkg/oracle"
)

// writeConnectionsFile writes a temporary connections file and returns its name
// together with a cleanup function.
func writeConnectionsFile(t *testing.T, content string) (string, func()) {
	t.Helper()
	f, err := os.CreateTemp("", "conns-*.txt")
	assert.Nil(t, err)
	_, err = f.WriteString(content)
	assert.Nil(t, err)
	assert.Nil(t, f.Close())
	return f.Name(), func() { os.Remove(f.Name()) }
}

func TestBatchQueryWorstStatusWins(t *testing.T) {
	file, cleanup := writeConnectionsFile(t,
		"prod,okuser/pass@DB1\nstage,warnuser/pass@DB2\ndr,failuser/pass@DB3\n")
	defer cleanup()

	run := func(c oracle.Connection, stmt string) (string, string, error) {
		switch c.Username {
		case "okuser":
			return "ok", "fine", nil
		case "warnuser":
			return "warn", "busy", nil
		default:
			return "", "", fmt.Errorf("ORA-12541: TNS no listener")
		}
	}

	status, output, err := batchQuery(oracle.FileParams{File: file, Timeout: 5 * time.Second}, "select 1", run)
	assert.Nil(t, err)
	assert.Equal(t, "critical", status)
	assert.Contains(t, output, "1 critical, 1 warning, 1 ok (of 3)")
	assert.Contains(t, output, "stage (warnuser@DB2): WARNING busy")
	assert.Contains(t, output, "dr (failuser@DB3): CRITICAL ORA-12541")
}

func TestBatchQueryAllOk(t *testing.T) {
	file, cleanup := writeConnectionsFile(t, "a,u1/p@DB1\nb,u2/p@DB2\n")
	defer cleanup()

	run := func(c oracle.Connection, stmt string) (string, string, error) {
		return "ok", "fine", nil
	}

	status, output, err := batchQuery(oracle.FileParams{File: file, Timeout: 5 * time.Second}, "select 1", run)
	assert.Nil(t, err)
	assert.Equal(t, "ok", status)
	assert.Contains(t, output, "0 critical, 0 warning, 2 ok (of 2)")
}

func TestBatchQueryUnknownStatusIsCritical(t *testing.T) {
	file, cleanup := writeConnectionsFile(t, "a,u1/p@DB1\n")
	defer cleanup()

	run := func(c oracle.Connection, stmt string) (string, string, error) {
		return "weird", "??", nil
	}

	status, _, err := batchQuery(oracle.FileParams{File: file, Timeout: 5 * time.Second}, "select 1", run)
	assert.Nil(t, err)
	assert.Equal(t, "critical", status)
}

func TestBatchQueryParseError(t *testing.T) {
	file, cleanup := writeConnectionsFile(t, "this is not a valid connection line\n")
	defer cleanup()

	run := func(c oracle.Connection, stmt string) (string, string, error) {
		return "ok", "fine", nil
	}

	_, _, err := batchQuery(oracle.FileParams{File: file, Timeout: 5 * time.Second}, "select 1", run)
	assert.NotNil(t, err)
}

func TestBatchQueryTimeout(t *testing.T) {
	file, cleanup := writeConnectionsFile(t, "a,u1/p@DB1\n")
	defer cleanup()

	run := func(c oracle.Connection, stmt string) (string, string, error) {
		time.Sleep(200 * time.Millisecond)
		return "ok", "fine", nil
	}

	_, _, err := batchQuery(oracle.FileParams{File: file, Timeout: time.Millisecond}, "select 1", run)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "timeout")
}

func TestReport(t *testing.T) {
	cases := map[string]int{"ok": 0, "warning": 1, "critical": 2, "anything-else": 2}
	for status, expected := range cases {
		var got int
		c := check.New("check-oracle-query")
		c.ExitFn = func(code int) { got = code }
		report(c, status, "msg")
		assert.Equal(t, expected, got, "status %q", status)
	}
}

func TestExecQuerySQLSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"status", "message"}).AddRow("ok", "all good")
	mock.ExpectQuery("select").WillReturnRows(rows)

	status, message, err := execQuery(context.Background(), db, "select status, message from health")
	assert.Nil(t, err)
	assert.Equal(t, "ok", status)
	assert.Equal(t, "all good", message)
	assert.Nil(t, mock.ExpectationsWereMet())
}

func TestExecQuerySQLNoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	mock.ExpectQuery("select").WillReturnRows(sqlmock.NewRows([]string{"status", "message"}))

	_, _, err = execQuery(context.Background(), db, "select status, message from health")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "no rows")
}

func TestExecQuerySQLError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	mock.ExpectQuery("select").WillReturnError(fmt.Errorf("ORA-00942: table or view does not exist"))

	_, _, err = execQuery(context.Background(), db, "select status, message from missing")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "ORA-00942")
}

func TestExecQuerySQLTimeout(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	mock.ExpectQuery("select").WillReturnRows(
		sqlmock.NewRows([]string{"status", "message"}).AddRow("ok", "x"))

	// A cancelled context makes database/sql abort before returning a row.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _, err = execQuery(ctx, db, "select status, message from health")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "timeout")
}

func TestExecQueryPLSQLError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	mock.ExpectExec("begin").WillReturnError(fmt.Errorf("ORA-06550: PL/SQL compilation error"))

	// Non-cancelled context: the PL/SQL branch returns the execution error.
	_, _, err = execQuery(context.Background(), db, "begin my_proc(:status, :message); end;")
	assert.NotNil(t, err)
}

func TestExecQueryPLSQLTimeout(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	mock.ExpectExec("begin")

	// Routing into the PL/SQL branch; a cancelled context yields the timeout path.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _, err = execQuery(ctx, db, "begin my_proc(:status, :message); end;")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "timeout")
}
