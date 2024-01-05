package main

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/thomis/sensu-plugins-go/pkg/check"
	"github.com/thomis/sensu-plugins-go/pkg/common"
)

func main() {
	var (
		connection common.Connection
		critical string
		warning  string
		critMin  int64
		critMax  int64
		warnMin  int64
		warnMax  int64
		err      error
	)

	connection.Database = "mysql"

	c := check.New("CheckMySQLProceses")
	c.Option.StringVarP(&connection.Host, "host", "h", "localhost", "MySQL host to connect to")
	c.Option.IntVarP(&connection.Port, "port", "P", 3306, "MySQL tcp port to connect to")
	c.Option.StringVarP(&connection.User, "user", "u", os.Getenv("MYSQL_USER"), "MySQL User")
	c.Option.StringVarP(&connection.Password, "password", "p", os.Getenv("MYSQL_PASSWORD"), "MySQL user password")
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

	processCount, err := selectProcessCount(connection)
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

func selectProcessCount(connection common.Connection) (int64, error) {
	var count int64

	source := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", connection.User, connection.Password, connection.Host, connection.Port, connection.Database)
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
