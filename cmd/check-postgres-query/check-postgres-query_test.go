package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/thomis/sensu-plugins-go/pkg/check"
)

func TestReport(t *testing.T) {
	cases := map[string]int{"ok": 0, "warning": 1, "critical": 2, "anything-else": 2}
	for status, expected := range cases {
		var got int
		c := check.New("check-postgres-query")
		c.ExitFn = func(code int) { got = code }
		report(c, status, "msg")
		assert.Equal(t, expected, got, "status %q", status)
	}
}

func TestExecQuerySuccess(t *testing.T) {
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

func TestExecQueryNoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	mock.ExpectQuery("select").WillReturnRows(sqlmock.NewRows([]string{"status", "message"}))

	_, _, err = execQuery(context.Background(), db, "select status, message from health")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "no rows")
}

func TestExecQueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	mock.ExpectQuery("select").WillReturnError(fmt.Errorf("pq: relation \"missing\" does not exist"))

	_, _, err = execQuery(context.Background(), db, "select status, message from missing")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

func TestExecQueryTimeout(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	mock.ExpectQuery("select").WillReturnRows(
		sqlmock.NewRows([]string{"status", "message"}).AddRow("ok", "x"))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _, err = execQuery(ctx, db, "select status, message from health")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "timeout")
}
