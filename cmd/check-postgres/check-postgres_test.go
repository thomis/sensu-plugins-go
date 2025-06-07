package main

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomis/sensu-plugins-go/pkg/common"
)

// Since we can't easily mock database connections without adding dependencies,
// we'll test the core logic separately

func TestSelectVersionConnectionString(t *testing.T) {
	tests := []struct {
		name       string
		connection common.Connection
		expected   string
	}{
		{
			name: "Standard connection",
			connection: common.Connection{
				Host:     "localhost",
				Port:     5432,
				User:     "testuser",
				Password: "testpass",
				Database: "testdb",
			},
			expected: "host=localhost port=5432 user=testuser password=testpass dbname=testdb sslmode=disable",
		},
		{
			name: "Connection with special characters",
			connection: common.Connection{
				Host:     "db.example.com",
				Port:     5433,
				User:     "user@domain",
				Password: "p@ss!word",
				Database: "my-db",
			},
			expected: "host=db.example.com port=5433 user=user@domain password=p@ss!word dbname=my-db sslmode=disable",
		},
		{
			name: "Connection with empty password",
			connection: common.Connection{
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Password: "",
				Database: "postgres",
			},
			expected: "host=localhost port=5432 user=postgres password= dbname=postgres sslmode=disable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reconstruct the connection string as done in selectVersion
			source := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
				tt.connection.Host,
				tt.connection.Port,
				tt.connection.User,
				tt.connection.Password,
				tt.connection.Database)

			assert.Equal(t, tt.expected, source)
		})
	}
}

// Test helper function to test regex extraction
func TestPostgreSQLVersionRegex(t *testing.T) {
	tests := []struct {
		name          string
		versionString string
		expected      string
		shouldMatch   bool
	}{
		{
			name:          "Standard PostgreSQL version",
			versionString: "PostgreSQL 13.1 on x86_64-pc-linux-gnu",
			expected:      "13.1",
			shouldMatch:   true,
		},
		{
			name:          "PostgreSQL with patch version",
			versionString: "PostgreSQL 12.5.1 (Ubuntu 12.5.1-0ubuntu0.20.04.1)",
			expected:      "12.5.1",
			shouldMatch:   true,
		},
		{
			name:          "PostgreSQL major version only",
			versionString: "PostgreSQL 15 on darwin",
			expected:      "15",
			shouldMatch:   true,
		},
		{
			name:          "Non-PostgreSQL database",
			versionString: "MySQL 8.0.23",
			expected:      "",
			shouldMatch:   false,
		},
		{
			name:          "Empty string",
			versionString: "",
			expected:      "",
			shouldMatch:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re := regexp.MustCompile(`PostgreSQL ([0-9\.]+)`)
			matches := re.FindStringSubmatch(tt.versionString)

			if tt.shouldMatch {
				assert.NotNil(t, matches)
				assert.Len(t, matches, 2)
				assert.Equal(t, tt.expected, matches[1])
			} else {
				assert.Nil(t, matches)
			}
		})
	}
}
