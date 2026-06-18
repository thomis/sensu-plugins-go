package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/godror/godror"
	"github.com/thomis/sensu-plugins-go/pkg/check"
	"github.com/thomis/sensu-plugins-go/pkg/dbquery"
	"github.com/thomis/sensu-plugins-go/pkg/oracle"
)

func main() {
	var (
		username  string
		password  string
		database  string
		query     string
		queryFile string
		file      string
		timeout   time.Duration
	)

	c := check.New("check-oracle-query")
	c.Option.SortFlags = false
	c.Option.StringVarP(&username, "username", "u", "", "Oracle username")
	c.Option.StringVarP(&password, "password", "p", "", "Oracle password")
	c.Option.StringVarP(&database, "database", "d", "", "Database name")
	c.Option.StringVarP(&query, "query", "q", "", "Inline query returning two values: status (ok|warn|warning|error) and message")
	c.Option.StringVar(&queryFile, "query-file", "", "File containing the query (alternative to -q)")
	c.Option.StringVarP(&file, "file", "f", "", "File with connection strings for batch mode. Line format: label,username/password@database")
	c.Option.DurationVarP(&timeout, "timeout", "T", 30*time.Second, "Timeout")
	c.Init()

	// Resolving the query text is a configuration concern; failures here are
	// usage errors (exit 3) rather than a database problem.
	stmt, err := dbquery.ReadQuery(query, queryFile)
	if err != nil {
		c.Error(err)
		return
	}

	// Batch mode: run the same query against every connection in the file.
	if len(file) > 0 {
		status, output, err := batchQuery(oracle.FileParams{File: file, Timeout: timeout}, stmt, runQuery)
		if err != nil {
			c.Critical(err.Error())
			return
		}
		report(c, status, output)
		return
	}

	// Single connection mode.
	connection := oracle.Connection{Username: username, Password: password, Database: database, Timeout: timeout}
	status, message, err := runQuery(connection, stmt)
	if err != nil {
		c.Critical(err.Error())
		return
	}

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

// queryRunner executes the statement against a single connection. main wires in
// runQuery; tests inject a fake to exercise the batch orchestration without a
// database.
type queryRunner func(connection oracle.Connection, stmt string) (string, string, error)

// batchQuery runs the statement against all connections in the file concurrently
// and reduces the per-connection outcomes to a single overall result
// (worst-status-wins). A connection or query failure, or an unrecognized status,
// counts as critical for that connection.
func batchQuery(fileParams oracle.FileParams, stmt string, run queryRunner) (string, string, error) {
	connections, err := oracle.ParseConnectionsFromFile(fileParams)
	if err != nil {
		return "", "", err
	}

	// Buffered so stragglers can send and exit even after an overall timeout,
	// avoiding leaked goroutines.
	channel := make(chan dbquery.QueryOutcome, len(*connections))
	for _, c := range *connections {
		go func(c oracle.Connection) {
			outcome := dbquery.QueryOutcome{Label: fmt.Sprintf("%s (%s@%s)", c.Label, c.Username, c.Database)}

			status, message, err := run(c, stmt)
			switch {
			case err != nil:
				outcome.Status = "critical"
				outcome.Message = err.Error()
			default:
				normalized, nerr := dbquery.NormalizeStatus(status)
				if nerr != nil {
					outcome.Status = "critical"
					outcome.Message = nerr.Error()
				} else {
					outcome.Status = normalized
					outcome.Message = message
				}
			}

			channel <- outcome
		}(c)
	}

	total := len(*connections)
	outcomes := make([]dbquery.QueryOutcome, 0, total)
	timeout := time.After(fileParams.Timeout)

	for i := 0; i < total; i++ {
		select {
		case outcome := <-channel:
			outcomes = append(outcomes, outcome)
		case <-timeout:
			return "", "", fmt.Errorf("timeout reached while running query on [%d] connections", total)
		}
	}

	status, output := dbquery.AggregateQueryOutcomes(outcomes)
	return status, output, nil
}

// runQuery connects to the database and executes the statement against a single
// connection.
func runQuery(connection oracle.Connection, stmt string) (string, string, error) {
	params := godror.ConnectionParams{}
	params.Username = connection.Username
	params.Password = godror.NewPassword(connection.Password)
	params.Timezone = time.UTC
	params.ConnectString = connection.Database

	db, err := sql.Open("godror", params.StringWithPassword())
	if err != nil {
		return "", "", oracle.ExtractOracleError(err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), connection.Timeout)
	defer cancel()

	return execQuery(ctx, db, stmt)
}

// execQuery runs the resolved statement against an open database handle. A plain
// SQL statement is expected to return a single row with two columns (status,
// message); a PL/SQL block is executed with two OUT bind variables (:status,
// :message). It is separated from connection handling so it can be tested with a
// mocked database.
func execQuery(ctx context.Context, db *sql.DB, stmt string) (string, string, error) {
	var (
		status  string
		message string
	)

	if dbquery.IsPLSQL(stmt) {
		_, err := db.ExecContext(ctx, stmt,
			sql.Named("status", sql.Out{Dest: &status}),
			sql.Named("message", sql.Out{Dest: &message}))
		if err != nil {
			if ctx.Err() != nil {
				return "", "", fmt.Errorf("timeout reached")
			}
			return "", "", oracle.ExtractOracleError(err)
		}
		return status, message, nil
	}

	err := db.QueryRowContext(ctx, stmt).Scan(&status, &message)
	if err != nil {
		if ctx.Err() != nil {
			return "", "", fmt.Errorf("timeout reached")
		}
		if err == sql.ErrNoRows {
			return "", "", fmt.Errorf("query returned no rows (expected one row with two columns: status, message)")
		}
		return "", "", oracle.ExtractOracleError(err)
	}

	return status, message, nil
}
