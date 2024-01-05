package main

import (
	"database/sql"
	"fmt"
	"regexp"

	_ "github.com/lib/pq"
	"github.com/thomis/sensu-plugins-go/pkg/check"
	"github.com/thomis/sensu-plugins-go/pkg/common"
)

func main() {
	var (
		connection common.Connection
	)

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
	}

	c.Ok(fmt.Sprint("Server version ", version))
}

func selectVersion(connection common.Connection) (string, error) {
	var info string

	source := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		connection.Host,
		connection.Port,
		connection.User,
		connection.Password,
		connection.Database)
	db, err := sql.Open("postgres", source)
	if err != nil {
		return "", err
	}
	defer db.Close()

	err = db.QueryRow("select version()").Scan(&info)
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile(`PostgreSQL ([0-9\.]+)`)
	return re.FindStringSubmatch(info)[1], nil
}
