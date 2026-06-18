package main

import (
	"database/sql"
	"fmt"
	"regexp"

	_ "github.com/lib/pq"
	"github.com/thomis/sensu-plugins-go/pkg/check"
	"github.com/thomis/sensu-plugins-go/pkg/common"
)

var versionRe = regexp.MustCompile(`PostgreSQL ([0-9\.]+)`)

func main() {
	var connection common.Connection

	c := check.New("CheckPostgres")
	c.Option.StringVarP(&connection.Host, "host", "h", "localhost", "HOST")
	c.Option.IntVarP(&connection.Port, "port", "P", 5432, "PORT")
	c.Option.StringVarP(&connection.User, "user", "u", "", "USER")
	c.Option.StringVarP(&connection.Password, "password", "p", "", "PASSWORD")
	c.Option.StringVarP(&connection.Database, "database", "d", "test", "DATABASE")
	c.Init()

	version, err := selectVersion(connection)
	if err != nil {
		c.Error(err)
		return
	}

	c.Ok(fmt.Sprint("Server version ", version))
}

func selectVersion(connection common.Connection) (string, error) {
	db, err := sql.Open("postgres", buildSource(connection))
	if err != nil {
		return "", err
	}
	defer db.Close()

	return queryVersion(db)
}

// buildSource assembles the lib/pq connection string from the connection
// parameters.
func buildSource(connection common.Connection) string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		connection.Host,
		connection.Port,
		connection.User,
		connection.Password,
		connection.Database)
}

// queryVersion reads the server version from an open database handle. It is
// separated from connection handling so it can be tested with a mocked database.
func queryVersion(db *sql.DB) (string, error) {
	var info string
	if err := db.QueryRow("select version()").Scan(&info); err != nil {
		return "", err
	}

	return parseVersion(info)
}

// parseVersion extracts the numeric version from a PostgreSQL version banner.
func parseVersion(info string) (string, error) {
	matches := versionRe.FindStringSubmatch(info)
	if matches == nil {
		return "", fmt.Errorf("could not parse PostgreSQL version from %q", info)
	}

	return matches[1], nil
}
