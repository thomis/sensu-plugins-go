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
		username     string
		password     string
		database     string
		file         string
		timeout      time.Duration
		excludeTypes []string

		response string
		err      error
	)

	c := check.New("check-oracle-validity")
	c.Option.SortFlags = false
	c.Option.StringVarP(&username, "username", "u", "", "Oracle username")
	c.Option.StringVarP(&password, "password", "p", "", "Oracle password")
	c.Option.StringVarP(&database, "database", "d", "", "Database name")
	c.Option.StringVarP(&file, "file", "f", "", "File with connection strings to check. Line format: label,username/password@database")
	c.Option.DurationVarP(&timeout, "timeout", "T", 30*time.Second, "Timeout")
	c.Option.StringArrayVarP(&excludeTypes, "exclude-types", "t", []string{}, "Exclude given object types from validity check")
	c.Init()

	if len(file) > 0 {
		response, err = fileValidity(oracle.FileParams{File: file, Timeout: timeout, ExcludeTypes: excludeTypes})
	} else {
		connection := oracle.Connection{Username: username, Password: password, Database: database, Timeout: timeout, ExcludeTypes: excludeTypes}
		response, err = singleValidity(connection)
	}

	if err != nil {
		c.Critical(err.Error())
		return
	}

	c.Ok(response)
}

func fileValidity(fileParams oracle.FileParams) (string, error) {
	connections, err := oracle.ParseConnectionsFromFile(fileParams)
	if err != nil {
		return "", err
	}

	channel := make(chan (task))
	for _, c := range *connections {
		go func(c oracle.Connection) {
			_, err := singleValidity(c)

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
		return "", fmt.Errorf("%d/%d connections are fine\n"+strings.Join(details, "\n\n"), success, total)
	}

	return fmt.Sprintf("%d/%d connections are fine", success, total), nil
}

func singleValidity(connection oracle.Connection) (string, error) {
	params := godror.ConnectionParams{}
	params.Username = connection.Username
	params.Password = godror.NewPassword(connection.Password)
	params.Timezone = time.UTC
	params.ConnectString = connection.Database

	db, err := sql.Open("godror", params.StringWithPassword())
	if err != nil {
		return "", extractOracleError(err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), connection.Timeout)
	defer cancel()

	var (
		objectType string
		objectName string

		objectsInvalid int64
		buffer         []string
	)

	stmt := "select object_type, object_name from user_objects where status = 'INVALID'"
	if len(connection.ExcludeTypes) > 0 {
		stmt = fmt.Sprintf("%s and object_type not in ('%s')", stmt, strings.Join(connection.ExcludeTypes, "','"))
	}
	stmt += " order by object_type, object_name"

	rows, err := db.QueryContext(ctx, stmt)
	if err != nil {
		if ctx.Err() != nil {
			return "", fmt.Errorf("timeout reached")
		}
		return "", extractOracleError(err)
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&objectType, &objectName)
		if err != nil {
			return "", extractOracleError((err))
		}

		objectsInvalid++
		buffer = append(buffer, fmt.Sprintf("%-40s%s", objectType, objectName))

	}

	if objectsInvalid > 0 {
		return "", fmt.Errorf("invalid objects: %d\n%s", objectsInvalid, strings.Join(buffer, "\n"))
	}

	return "All objects are valid", nil
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
