package main

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/thomis/sensu-plugins-go/pkg/oracle"
)

func writeConnectionsFile(t *testing.T, content string) (string, func()) {
	t.Helper()
	f, err := os.CreateTemp("", "conns-*.txt")
	assert.Nil(t, err)
	_, err = f.WriteString(content)
	assert.Nil(t, err)
	assert.Nil(t, f.Close())
	return f.Name(), func() { os.Remove(f.Name()) }
}

func TestExecPingSuccess(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	assert.Nil(t, err)
	defer db.Close()

	mock.ExpectPing()

	response, err := execPing(context.Background(), db)
	assert.Nil(t, err)
	assert.Equal(t, "Connection is pingable", response)
}

func TestExecPingError(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	assert.Nil(t, err)
	defer db.Close()

	mock.ExpectPing().WillReturnError(fmt.Errorf("ORA-12541: TNS no listener"))

	_, err = execPing(context.Background(), db)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "ORA-12541")
}

func TestExecPingTimeout(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	assert.Nil(t, err)
	defer db.Close()

	mock.ExpectPing()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = execPing(ctx, db)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "timeout")
}

func TestFilePingAllOk(t *testing.T) {
	file, cleanup := writeConnectionsFile(t, "a,u1/p@DB1\nb,u2/p@DB2\n")
	defer cleanup()

	ping := func(c oracle.Connection) (string, error) { return "Connection is pingable", nil }

	response, err := filePing(oracle.FileParams{File: file, Timeout: 5 * time.Second}, ping)
	assert.Nil(t, err)
	assert.Equal(t, "2/2 connections are pingable", response)
}

func TestFilePingWithFailure(t *testing.T) {
	file, cleanup := writeConnectionsFile(t, "ok,u1/p@DB1\nbad,u2/p@DB2\n")
	defer cleanup()

	ping := func(c oracle.Connection) (string, error) {
		if c.Username == "u2" {
			return "", fmt.Errorf("ORA-12541: TNS no listener")
		}
		return "Connection is pingable", nil
	}

	_, err := filePing(oracle.FileParams{File: file, Timeout: 5 * time.Second}, ping)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "1/2 connections are pingable")
	assert.Contains(t, err.Error(), "bad (u2@DB2): ORA-12541")
}

func TestFilePingParseError(t *testing.T) {
	file, cleanup := writeConnectionsFile(t, "not a valid line\n")
	defer cleanup()

	ping := func(c oracle.Connection) (string, error) { return "Connection is pingable", nil }

	_, err := filePing(oracle.FileParams{File: file, Timeout: 5 * time.Second}, ping)
	assert.NotNil(t, err)
}

func TestFilePingTimeout(t *testing.T) {
	file, cleanup := writeConnectionsFile(t, "a,u1/p@DB1\n")
	defer cleanup()

	ping := func(c oracle.Connection) (string, error) {
		time.Sleep(200 * time.Millisecond)
		return "Connection is pingable", nil
	}

	_, err := filePing(oracle.FileParams{File: file, Timeout: time.Millisecond}, ping)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "timeout")
}
