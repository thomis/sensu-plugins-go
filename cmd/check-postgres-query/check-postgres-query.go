package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/thomis/sensu-plugins-go/pkg/check"
	"github.com/thomis/sensu-plugins-go/pkg/common"
	"github.com/thomis/sensu-plugins-go/pkg/dbquery"
)

func main() {
	var (
		connection common.Connection
		query      string
		queryFile  string
		timeout    time.Duration
	)

	c := check.New("check-postgres-query")
	c.Option.SortFlags = false
	c.Option.StringVarP(&connection.Host, "host", "h", "localhost", "Host")
	c.Option.IntVarP(&connection.Port, "port", "P", 5432, "Port")
	c.Option.StringVarP(&connection.User, "user", "u", "", "User")
	c.Option.StringVarP(&connection.Password, "password", "p", "", "Password")
	c.Option.StringVarP(&connection.Database, "database", "d", "test", "Database")
	c.Option.StringVarP(&query, "query", "q", "", "Inline query returning two values: status (ok|warn|warning|error) and message")
	c.Option.StringVar(&queryFile, "query-file", "", "File containing the query (alternative to -q)")
	c.Option.DurationVarP(&timeout, "timeout", "T", 30*time.Second, "Timeout")
	c.Init()

	// Resolving the query text is a configuration concern; failures here are
	// usage errors (exit 3) rather than a database problem.
	stmt, err := dbquery.ReadQuery(query, queryFile)
	if err != nil {
		c.Error(err)
		return
	}

	status, message, err := runQuery(connection, stmt, timeout)
	if err != nil {
		c.Critical(err.Error())
		return
	}

	// The query decides the outcome; an unrecognized status is treated as a
	// usage/contract error (exit 3).
	status, err = dbquery.NormalizeStatus(status)
	if err != nil {
		c.Error(err)
		return
	}

	report(c, status, message)
}

// report maps a normalized status (ok|warning|critical) to the matching check
// result and exit code.
func report(c *check.CheckStruct, status string, output string) {
	switch status {
	case "ok":
		c.Ok(output)
	case "warning":
		c.Warning(output)
	default:
		c.Critical(output)
	}
}

// runQuery connects to PostgreSQL and executes the statement.
func runQuery(connection common.Connection, stmt string, timeout time.Duration) (string, string, error) {
	source := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		connection.Host,
		connection.Port,
		connection.User,
		connection.Password,
		connection.Database)

	db, err := sql.Open("postgres", source)
	if err != nil {
		return "", "", err
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return execQuery(ctx, db, stmt)
}

// execQuery runs the statement against an open database handle. The statement is
// expected to return a single row with two columns (status, message). It is
// separated from connection handling so it can be tested with a mocked database.
func execQuery(ctx context.Context, db *sql.DB, stmt string) (string, string, error) {
	var (
		status  string
		message string
	)

	err := db.QueryRowContext(ctx, stmt).Scan(&status, &message)
	if err != nil {
		if ctx.Err() != nil {
			return "", "", fmt.Errorf("timeout reached")
		}
		if err == sql.ErrNoRows {
			return "", "", fmt.Errorf("query returned no rows (expected one row with two columns: status, message)")
		}
		return "", "", err
	}

	return status, message, nil
}
