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

type session struct {
	Connection common.Connection
	Critical   string
	Warning    string
	CritMin    int64
	CritMax    int64
	WarnMin    int64
	WarnMax    int64
	Check      *check.CheckStruct
}

func main() {
	var (
		session session
		err     error
	)

	session.Connection.Database = "mysql"
	session.handleArguments()

	crits := strings.Split(session.Critical, ":")
	warns := strings.Split(session.Warning, ":")

	session.CritMin, err = strconv.ParseInt(crits[0], 10, 64)
	if err != nil {
		session.Check.Error(err)
	}
	if len(crits) > 1 {
		session.CritMax, err = strconv.ParseInt(crits[1], 10, 64)
		if err != nil {
			session.Check.Error(err)
		}
	} else {
		session.CritMax = 0
	}
	session.WarnMin, err = strconv.ParseInt(warns[0], 10, 64)
	if err != nil {
		session.Check.Error(err)
	}
	if len(warns) > 1 {
		session.WarnMax, err = strconv.ParseInt(warns[1], 10, 64)
		if err != nil {
			session.Check.Error(err)
		}
	} else {
		session.WarnMax = 0
	}
	if session.CritMax > 0 && session.CritMin > session.CritMax {
		session.Check.Error(fmt.Errorf("critical argument %s invalid, min %d is greater than max %d", session.Critical, session.CritMin, session.CritMax))
	}
	if session.WarnMax > 0 && session.WarnMin > session.WarnMax {
		session.Check.Error(fmt.Errorf("warning argument %s invalid, min %d is greater than max %d", session.Warning, session.WarnMin, session.WarnMax))
	}

	processCount, err := selectProcessCount(session.Connection)
	if err != nil {
		session.Check.Error(err)
	}

	switch {
	case session.CritMax > 0 && processCount >= session.CritMax:
		session.Check.Critical(fmt.Sprintf("%d MySQL processes exceed threshold of %d | mysql_processes=%d;%s;%s;0", processCount, session.CritMax, processCount, session.Warning, session.Critical))
	case processCount <= session.CritMin:
		session.Check.Critical(fmt.Sprintf("%d MySQL processes are below threshold of %d | mysql_processes=%d;%s;%s;0", processCount, session.CritMin, processCount, session.Warning, session.Critical))
	case session.WarnMax > 0 && processCount >= session.WarnMax:
		session.Check.Warning(fmt.Sprintf("%d MySQL processes exceed threshold of %d | mysql_processes=%d;%s;%s;0", processCount, session.WarnMax, processCount, session.Warning, session.Critical))
	case processCount <= session.WarnMin:
		session.Check.Warning(fmt.Sprintf("%d MySQL processes are below threshold of %d | mysql_processes=%d;%s;%s;0", processCount, session.WarnMin, processCount, session.Warning, session.Critical))
	default:
		session.Check.Ok(fmt.Sprintf("MySQL process Count %d | mysql_processes=%d;%s;%s;0", processCount, processCount, session.Warning, session.Critical))
	}
}

func (s *session) handleArguments() {
	s.Check = check.New("CheckMySQLProceses")
	s.Check.Option.StringVarP(&s.Connection.Host, "host", "h", "localhost", "MySQL host to connect to")
	s.Check.Option.IntVarP(&s.Connection.Port, "port", "P", 3306, "MySQL tcp port to connect to")
	s.Check.Option.StringVarP(&s.Connection.User, "user", "u", os.Getenv("MYSQL_USER"), "MySQL User")
	s.Check.Option.StringVarP(&s.Connection.Password, "password", "p", os.Getenv("MYSQL_PASSWORD"), "MySQL user password")
	s.Check.Option.StringVarP(&s.Critical, "critical", "c", "", "Critical min:max threshold, max is optional")
	s.Check.Option.StringVarP(&s.Warning, "warning", "w", "", "Warning min:max threshold, max is optional")
	s.Check.Init()
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
