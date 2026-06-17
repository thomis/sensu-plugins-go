package main

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/godror/godror"
	"github.com/thomis/sensu-plugins-go/pkg/check"
	"github.com/thomis/sensu-plugins-go/pkg/oracle"
)

type task struct {
	connection *oracle.Connection
	err        error
}

func main() {
	var (
		username string
		password string
		database string
		file     string
		timeout  time.Duration

		response string
		err      error
	)

	c := check.New("check-oracle-ping")
	c.Option.SortFlags = false
	c.Option.StringVarP(&username, "username", "u", "", "Oracle username")
	c.Option.StringVarP(&password, "password", "p", "", "Oracle password")
	c.Option.StringVarP(&database, "database", "d", "", "Database name")
	c.Option.StringVarP(&file, "file", "f", "", "File with connection strings to check. Line format: label,username/password@database")
	c.Option.DurationVarP(&timeout, "timeout", "T", 30*time.Second, "Timeout")
	c.Init()

	if len(file) > 0 {
		response, err = filePing(oracle.FileParams{File: file, Timeout: timeout}, singlePing)
	} else {
		connection := oracle.Connection{Username: username, Password: password, Database: database, Timeout: timeout}
		response, err = singlePing(connection)
	}

	if err != nil {
		c.Critical(err.Error())
		return
	}

	c.Ok(response)
}

// pingRunner pings a single connection. main wires in singlePing; tests inject a
// fake to exercise the batch orchestration without a database.
type pingRunner func(connection oracle.Connection) (string, error)

func filePing(fileParams oracle.FileParams, ping pingRunner) (string, error) {
	connections, err := oracle.ParseConnectionsFromFile(fileParams)
	if err != nil {
		return "", err
	}

	// Buffered so stragglers can send and exit even after an overall timeout,
	// avoiding leaked goroutines.
	channel := make(chan task, len(*connections))
	for _, c := range *connections {
		go func(c oracle.Connection) {
			_, err := ping(c)

			channel <- task{
				connection: &c,
				err:        err}
		}(c)
	}

	total := len(*connections)
	success := 0
	timeout := time.After(fileParams.Timeout)
	details := []string{}

	for i := 0; i < total; i++ {
		select {
		case task := <-channel:
			if task.err == nil {
				success++
			} else {
				details = append(details, fmt.Sprintf("- %s (%s@%s): %s", task.connection.Label, task.connection.Username, task.connection.Database, task.err.Error()))
			}
		case <-timeout:
			return "", fmt.Errorf("timeout reached while testing [%d] connections", total)
		}
	}

	if success < total {
		return "", fmt.Errorf("%d/%d connections are pingable\n"+strings.Join(details, "\n"), success, total)
	}

	return fmt.Sprintf("%d/%d connections are pingable", success, total), nil
}

func singlePing(connection oracle.Connection) (string, error) {
	params := godror.ConnectionParams{}
	params.Username = connection.Username
	params.Password = godror.NewPassword(connection.Password)
	params.Timezone = time.UTC
	params.ConnectString = connection.Database

	db, err := sql.Open("godror", params.StringWithPassword())
	if err != nil {
		return "", oracle.ExtractOracleError(err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), connection.Timeout)
	defer cancel()

	return execPing(ctx, db)
}

// execPing pings an open database handle. It is separated from connection
// handling so it can be tested with a mocked database.
func execPing(ctx context.Context, db *sql.DB) (string, error) {
	err := db.PingContext(ctx)
	if err != nil {
		if ctx.Err() != nil {
			return "", fmt.Errorf("timeout reached")
		}
		return "", oracle.ExtractOracleError(err)
	}

	return "Connection is pingable", nil
}
