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

func TestExecValidityAllValid(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	mock.ExpectQuery("object_type").WillReturnRows(
		sqlmock.NewRows([]string{"object_type", "object_name"}))

	response, err := execValidity(context.Background(), db, nil)
	assert.Nil(t, err)
	assert.Equal(t, "All objects are valid", response)
}

func TestExecValidityInvalidObjects(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"object_type", "object_name"}).
		AddRow("PROCEDURE", "PROC_DAILY").
		AddRow("VIEW", "V_SALES")
	mock.ExpectQuery("object_type").WillReturnRows(rows)

	// Non-empty exclude list also exercises the exclusion branch.
	_, err = execValidity(context.Background(), db, []string{"INDEX", "SYNONYM"})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "invalid objects: 2")
	assert.Contains(t, err.Error(), "PROCEDURE")
	assert.Contains(t, err.Error(), "V_SALES")
}

func TestExecValidityQueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	mock.ExpectQuery("object_type").WillReturnError(fmt.Errorf("ORA-00942: table or view does not exist"))

	_, err = execValidity(context.Background(), db, nil)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "ORA-00942")
}

func TestExecValidityTimeout(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	mock.ExpectQuery("object_type").WillReturnRows(
		sqlmock.NewRows([]string{"object_type", "object_name"}))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = execValidity(ctx, db, nil)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "timeout")
}

func TestFileValidityAllOk(t *testing.T) {
	file, cleanup := writeConnectionsFile(t, "a,u1/p@DB1\nb,u2/p@DB2\n")
	defer cleanup()

	validity := func(c oracle.Connection) (string, error) { return "All objects are valid", nil }

	response, err := fileValidity(oracle.FileParams{File: file, Timeout: 5 * time.Second}, validity)
	assert.Nil(t, err)
	assert.Equal(t, "2/2 connections are fine", response)
}

func TestFileValidityWithFailure(t *testing.T) {
	file, cleanup := writeConnectionsFile(t, "ok,u1/p@DB1\nbad,u2/p@DB2\n")
	defer cleanup()

	validity := func(c oracle.Connection) (string, error) {
		if c.Username == "u2" {
			return "", fmt.Errorf("invalid objects: 1")
		}
		return "All objects are valid", nil
	}

	_, err := fileValidity(oracle.FileParams{File: file, Timeout: 5 * time.Second}, validity)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "1/2 connections are fine")
	assert.Contains(t, err.Error(), "bad (u2@DB2): invalid objects: 1")
}

func TestFileValidityParseError(t *testing.T) {
	file, cleanup := writeConnectionsFile(t, "not a valid line\n")
	defer cleanup()

	validity := func(c oracle.Connection) (string, error) { return "All objects are valid", nil }

	_, err := fileValidity(oracle.FileParams{File: file, Timeout: 5 * time.Second}, validity)
	assert.NotNil(t, err)
}

func TestFileValidityTimeout(t *testing.T) {
	file, cleanup := writeConnectionsFile(t, "a,u1/p@DB1\n")
	defer cleanup()

	validity := func(c oracle.Connection) (string, error) {
		time.Sleep(200 * time.Millisecond)
		return "All objects are valid", nil
	}

	_, err := fileValidity(oracle.FileParams{File: file, Timeout: time.Millisecond}, validity)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "timeout")
}
