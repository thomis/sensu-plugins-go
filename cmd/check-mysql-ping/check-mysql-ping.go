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

func main() {
	var (
		connection common.Connection
	)

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
	}

	c.Ok(fmt.Sprint("Server version ", version))
}

func selectVersion(connection common.Connection) (string, error) {
	var info string

	source := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		connection.User,
		connection.Password,
		connection.Host,
		connection.Port,
		connection.Database)
	db, err := sql.Open("mysql", source)
	if err != nil {
		return "", err
	}
	defer db.Close()

	err = db.QueryRow("select version()").Scan(&info)
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile(`([0-9\.]+)`)
	return re.FindStringSubmatch(info)[1], nil
}
