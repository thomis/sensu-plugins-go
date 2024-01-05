package main

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/thomis/sensu-plugins-go/pkg/check"
)

func main() {
	var (
		host     string
		port     int
		user     string
		password string
		critical string
		warning  string
		critMin  int64
		critMax  int64
		warnMin  int64
		warnMax  int64
		err      error
	)

	c := check.New("CheckMySQLProceses")
	c.Option.StringVarP(&host, "host", "h", "localhost", "MySQL host to connect to")
	c.Option.IntVarP(&port, "port", "P", 3306, "MySQL tcp port to connect to")
	c.Option.StringVarP(&user, "user", "u", os.Getenv("MYSQL_USER"), "MySQL User")
	c.Option.StringVarP(&password, "password", "p", os.Getenv("MYSQL_PASSWORD"), "MySQL user password")
	c.Option.StringVarP(&critical, "critical", "c", "", "Critical min:max threshold, max is optional")
	c.Option.StringVarP(&warning, "warning", "w", "", "Warning min:max threshold, max is optional")
	c.Init()

	crits := strings.Split(critical, ":")
	warns := strings.Split(warning, ":")

	critMin, err = strconv.ParseInt(crits[0], 10, 64)
	if err != nil {
		c.Error(err)
	}
	if len(crits) > 1 {
		critMax, err = strconv.ParseInt(crits[1], 10, 64)
		if err != nil {
			c.Error(err)
		}
	} else {
		critMax = 0
	}
	warnMin, err = strconv.ParseInt(warns[0], 10, 64)
	if err != nil {
		c.Error(err)
	}
	if len(warns) > 1 {
		warnMax, err = strconv.ParseInt(warns[1], 10, 64)
		if err != nil {
			c.Error(err)
		}
	} else {
		warnMax = 0
	}
	if critMax > 0 && critMin > critMax {
		c.Error(fmt.Errorf("critical argument %s invalid, min %d is greater than max %d", critical, critMin, critMax))
	}
	if warnMax > 0 && warnMin > warnMax {
		c.Error(fmt.Errorf("warning argument %s invalid, min %d is greater than max %d", warning, warnMin, warnMax))
	}

	processCount, err := selectProcessCount(host, port, user, password, "mysql")
	if err != nil {
		c.Error(err)
	}

	switch {
	case critMax > 0 && processCount >= critMax:
		c.Critical(fmt.Sprintf("%d MySQL processes exceed threshold of %d | mysql_processes=%d;%s;%s;0", processCount, critMax, processCount, warning, critical))
	case processCount <= critMin:
		c.Critical(fmt.Sprintf("%d MySQL processes are below threshold of %d | mysql_processes=%d;%s;%s;0", processCount, critMin, processCount, warning, critical))
	case warnMax > 0 && processCount >= warnMax:
		c.Warning(fmt.Sprintf("%d MySQL processes exceed threshold of %d | mysql_processes=%d;%s;%s;0", processCount, warnMax, processCount, warning, critical))
	case processCount <= warnMin:
		c.Warning(fmt.Sprintf("%d MySQL processes are below threshold of %d | mysql_processes=%d;%s;%s;0", processCount, warnMin, processCount, warning, critical))
	default:
		c.Ok(fmt.Sprintf("MySQL process Count %d | mysql_processes=%d;%s;%s;0", processCount, processCount, warning, critical))
	}
}

func selectProcessCount(host string, port int, user string, password string, database string) (int64, error) {
	var count int64

	source := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", user, password, host, port, database)
	db, err := sql.Open("mysql", source)
	if err != nil {
		return 0, err
	}
	defer db.Close()

	err = db.QueryRow("select count(*) from information_schema.PROCESSLIST").Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}
