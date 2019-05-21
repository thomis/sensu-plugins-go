package main

import (
	"database/sql"
	"fmt"
	"os"
	"regexp"

	_ "github.com/go-sql-driver/mysql"
	"github.com/portertech/sensu-plugins-go/lib/check"
)

func main() {
	var (
		host     string
		port     int
		database string
		user     string
		password string
	)

	c := check.New("CheckMySQLPing")
	c.Option.StringVarP(&host, "host", "h", "localhost", "MySQL host to connect to")
	c.Option.IntVarP(&port, "port", "P", 3306, "MySQL tcp port to connect to")
	c.Option.StringVarP(&user, "user", "u", os.Getenv("MYSQL_USER"), "MySQL User")
	c.Option.StringVarP(&password, "password", "p", os.Getenv("MYSQL_PASSWORD"), "MySQL user password")
	c.Option.StringVarP(&database, "database", "d", "mysql", "MySQL database")
	c.Init()

	version, err := selectVersion(host, port, user, password, database)
	if err != nil {
		c.Error(err)
	}

	c.Ok(fmt.Sprint("Server version ", version))
}

func selectVersion(host string, port int, user string, password string, database string) (string, error) {
	var info string

	source := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", user, password, host, port, database)
	db, err := sql.Open("mysql", source)
	if err != nil {
		return "", err
	}
	defer db.Close()

	err = db.QueryRow("select version()").Scan(&info)
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile("([0-9\\.]+)")
	return re.FindStringSubmatch(info)[1], nil
}
