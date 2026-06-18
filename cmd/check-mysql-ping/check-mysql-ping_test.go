package main

import (
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/thomis/sensu-plugins-go/pkg/common"
)

func TestBuildSource(t *testing.T) {
	source := buildSource(common.Connection{
		Host: "db.example.com", Port: 3306, User: "monitor", Password: "secret", Database: "appdb",
	})
	assert.Equal(t, "monitor:secret@tcp(db.example.com:3306)/appdb", source)
}

func TestParseVersion(t *testing.T) {
	v, err := parseVersion("8.0.23-0ubuntu0.20.04.1")
	assert.NoError(t, err)
	assert.Equal(t, "8.0.23", v)

	v, err = parseVersion("5.7.31")
	assert.NoError(t, err)
	assert.Equal(t, "5.7.31", v)

	_, err = parseVersion("no digits here")
	assert.Error(t, err)
}

func TestQueryVersionSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	mock.ExpectQuery("select version").WillReturnRows(
		sqlmock.NewRows([]string{"version()"}).AddRow("8.0.23"))

	v, err := queryVersion(db)
	assert.NoError(t, err)
	assert.Equal(t, "8.0.23", v)
}

func TestQueryVersionError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	mock.ExpectQuery("select version").WillReturnError(fmt.Errorf("connection refused"))

	_, err = queryVersion(db)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection refused")
}
