package main

import (
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/thomis/sensu-plugins-go/pkg/common"
)

func TestBuildSource(t *testing.T) {
	tests := []struct {
		name       string
		connection common.Connection
		expected   string
	}{
		{
			name: "Standard connection",
			connection: common.Connection{
				Host: "localhost", Port: 5432, User: "testuser", Password: "testpass", Database: "testdb",
			},
			expected: "host=localhost port=5432 user=testuser password=testpass dbname=testdb sslmode=disable",
		},
		{
			name: "Connection with special characters",
			connection: common.Connection{
				Host: "db.example.com", Port: 5433, User: "user@domain", Password: "p@ss!word", Database: "my-db",
			},
			expected: "host=db.example.com port=5433 user=user@domain password=p@ss!word dbname=my-db sslmode=disable",
		},
		{
			name: "Connection with empty password",
			connection: common.Connection{
				Host: "localhost", Port: 5432, User: "postgres", Password: "", Database: "postgres",
			},
			expected: "host=localhost port=5432 user=postgres password= dbname=postgres sslmode=disable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, buildSource(tt.connection))
		})
	}
}

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name          string
		versionString string
		expected      string
		shouldMatch   bool
	}{
		{"Standard PostgreSQL version", "PostgreSQL 13.1 on x86_64-pc-linux-gnu", "13.1", true},
		{"PostgreSQL with patch version", "PostgreSQL 12.5.1 (Ubuntu 12.5.1-0ubuntu0.20.04.1)", "12.5.1", true},
		{"PostgreSQL major version only", "PostgreSQL 15 on darwin", "15", true},
		{"Non-PostgreSQL database", "MySQL 8.0.23", "", false},
		{"Empty string", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, err := parseVersion(tt.versionString)
			if tt.shouldMatch {
				assert.Nil(t, err)
				assert.Equal(t, tt.expected, version)
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}

func TestQueryVersionSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"version"}).
		AddRow("PostgreSQL 13.1 on x86_64-pc-linux-gnu, compiled by gcc")
	mock.ExpectQuery("select version").WillReturnRows(rows)

	version, err := queryVersion(db)
	assert.Nil(t, err)
	assert.Equal(t, "13.1", version)
	assert.Nil(t, mock.ExpectationsWereMet())
}

func TestQueryVersionQueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	mock.ExpectQuery("select version").WillReturnError(fmt.Errorf("connection refused"))

	_, err = queryVersion(db)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "connection refused")
}

func TestQueryVersionUnparseable(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"version"}).AddRow("Some other database 1.2.3")
	mock.ExpectQuery("select version").WillReturnRows(rows)

	_, err = queryVersion(db)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "could not parse")
}
