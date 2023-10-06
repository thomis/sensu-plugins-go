package main

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/godror/godror"
	"github.com/portertech/sensu-plugins-go/lib/check"
)

type fileParams struct {
	file    string
	timeout time.Duration
}

type connection struct {
	label    string
	username string
	password string
	database string
	timeout  time.Duration
}

type task struct {
	connection *connection
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
		response, err = filePing(fileParams{file: file, timeout: timeout})
	} else {
		connection := connection{username: username, password: password, database: database, timeout: timeout}
		response, err = singlePing(connection)
	}

	if err != nil {
		c.Critical(err.Error())
		return
	}

	c.Ok(response)
}

func filePing(fileParams fileParams) (string, error) {
	connections, err := parseConnectionsFromFile(fileParams)
	if err != nil {
		return "", err
	}

	channel := make(chan (task))
	for _, c := range *connections {
		go func(c connection) {
			_, err := singlePing(c)

			channel <- task{
				connection: &c,
				err:        err}
		}(c)
	}

	total := len(*connections)
	success := 0
	timeout := time.After(fileParams.timeout)
	details := []string{}

	for i := 0; i < total; i++ {
		select {
		case task := <-channel:
			if task.err == nil {
				success++
			} else {
				details = append(details, fmt.Sprintf("- %s (%s@%s): %s", task.connection.label, task.connection.username, task.connection.database, task.err.Error()))
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

func singlePing(connection connection) (string, error) {
	params := godror.ConnectionParams{}
	params.Username = connection.username
	params.Password = godror.NewPassword(connection.password)
	params.Timezone = time.UTC
	params.ConnectString = connection.database

	db, err := sql.Open("godror", params.StringWithPassword())
	if err != nil {
		return "", extractOracleError(err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), connection.timeout)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		if ctx.Err() != nil {
			return "", fmt.Errorf("timeout reached")
		}
		return "", extractOracleError(err)
	}

	return "Connection is pingable", nil
}

func parseConnectionsFromFile(fileParams fileParams) (*[]connection, error) {
	connections := []connection{}

	readFile, err := os.Open(fileParams.file)
	if err != nil {
		return &connections, err
	}
	defer readFile.Close()

	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)

	reConnection := regexp.MustCompile(`(.+),(.+)/(.+)@(.+)`)

	i := 0
	for fileScanner.Scan() {
		i++
		line := strings.TrimSpace(fileScanner.Text())

		// empty line
		if len(line) == 0 {
			continue
		}

		// comment line
		if line[0] == '#' {
			continue
		}

		result := reConnection.FindSubmatch([]byte(line))
		if len(result) == 0 {
			return &connections, fmt.Errorf("connection string on line [%d] does not match pattern [label,username/password@database]", i)
		}

		connection := connection{
			label:    string(result[1]),
			username: string(result[2]),
			password: string(result[3]),
			database: string(result[4]),
			timeout:  fileParams.timeout}
		connections = append(connections, connection)
	}

	return &connections, nil
}

func extractOracleError(err error) error {
	if err == nil {
		return err
	}

	oraErr, isOraErr := godror.AsOraErr(err)
	if isOraErr {
		return fmt.Errorf("ORA-%d: %s", oraErr.Code(), oraErr.Message())
	}

	return err
}
