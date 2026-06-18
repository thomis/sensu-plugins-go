package main

import (
	"database/sql"
	"fmt"
	"os"
	"regexp"

	_ "github.com/go-sql-driver/mysql"
	"github.com/thomis/sensu-plugins-go/pkg/check"
	"github.com/thomis/sensu-plugins-go/pkg/common"
)

var versionRe = regexp.MustCompile(`([0-9\.]+)`)

func main() {
	var connection common.Connection

	c := check.New("CheckMySQLPing")
	c.Option.StringVarP(&connection.Host, "host", "h", "localhost", "MySQL host to connect to")
	c.Option.IntVarP(&connection.Port, "port", "P", 3306, "MySQL tcp port to connect to")
	c.Option.StringVarP(&connection.User, "user", "u", os.Getenv("MYSQL_USER"), "MySQL User")
	c.Option.StringVarP(&connection.Password, "password", "p", os.Getenv("MYSQL_PASSWORD"), "MySQL user password")
	c.Option.StringVarP(&connection.Database, "database", "d", "mysql", "MySQL database")
	c.Init()

	version, err := selectVersion(connection)
	if err != nil {
		c.Error(err)
		return
	}

	c.Ok(fmt.Sprint("Server version ", version))
}

func selectVersion(connection common.Connection) (string, error) {
	db, err := sql.Open("mysql", buildSource(connection))
	if err != nil {
		return "", err
	}
	defer db.Close()

	return queryVersion(db)
}

// buildSource assembles the go-sql-driver DSN from the connection parameters.
func buildSource(connection common.Connection) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		connection.User,
		connection.Password,
		connection.Host,
		connection.Port,
		connection.Database)
}

// queryVersion reads the server version from an open database handle.
func queryVersion(db *sql.DB) (string, error) {
	var info string
	if err := db.QueryRow("select version()").Scan(&info); err != nil {
		return "", err
	}

	return parseVersion(info)
}

// parseVersion extracts the numeric version from a MySQL version banner.
func parseVersion(info string) (string, error) {
	matches := versionRe.FindStringSubmatch(info)
	if matches == nil {
		return "", fmt.Errorf("could not parse version from %q", info)
	}

	return matches[1], nil
}
