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
	Connection   common.Connection
	Critical     string
	Warning      string
	CritMin      int64
	CritMax      int64
	WarnMin      int64
	WarnMax      int64
	ProcessCount int64
	Check        *check.CheckStruct
}

func main() {
	var (
		session session
		err     error
	)

	session.Connection.Database = "mysql"
	session.handleArguments()

	session.CritMin, session.CritMax, session.WarnMin, session.WarnMax, err = parseThresholds(session.Critical, session.Warning)
	if err != nil {
		session.Check.Error(err)
		return
	}

	session.ProcessCount, err = selectProcessCount(session.Connection)
	if err != nil {
		session.Check.Error(err)
		return
	}

	session.report()
}

// parseThresholds parses "min:max" critical and warning arguments (max is
// optional) and validates that min does not exceed max.
func parseThresholds(critical, warning string) (critMin, critMax, warnMin, warnMax int64, err error) {
	crits := strings.Split(critical, ":")
	warns := strings.Split(warning, ":")

	if critMin, err = strconv.ParseInt(crits[0], 10, 64); err != nil {
		return
	}
	if len(crits) > 1 {
		if critMax, err = strconv.ParseInt(crits[1], 10, 64); err != nil {
			return
		}
	}
	if warnMin, err = strconv.ParseInt(warns[0], 10, 64); err != nil {
		return
	}
	if len(warns) > 1 {
		if warnMax, err = strconv.ParseInt(warns[1], 10, 64); err != nil {
			return
		}
	}

	if critMax > 0 && critMin > critMax {
		err = fmt.Errorf("critical argument %s invalid, min %d is greater than max %d", critical, critMin, critMax)
		return
	}
	if warnMax > 0 && warnMin > warnMax {
		err = fmt.Errorf("warning argument %s invalid, min %d is greater than max %d", warning, warnMin, warnMax)
		return
	}

	return
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

func (s *session) report() {
	switch {
	case s.CritMax > 0 && s.ProcessCount >= s.CritMax:
		s.Check.Critical(fmt.Sprintf("%d MySQL processes exceed threshold of %d | mysql_processes=%d;%s;%s;0", s.ProcessCount, s.CritMax, s.ProcessCount, s.Warning, s.Critical))
	case s.ProcessCount <= s.CritMin:
		s.Check.Critical(fmt.Sprintf("%d MySQL processes are below threshold of %d | mysql_processes=%d;%s;%s;0", s.ProcessCount, s.CritMin, s.ProcessCount, s.Warning, s.Critical))
	case s.WarnMax > 0 && s.ProcessCount >= s.WarnMax:
		s.Check.Warning(fmt.Sprintf("%d MySQL processes exceed threshold of %d | mysql_processes=%d;%s;%s;0", s.ProcessCount, s.WarnMax, s.ProcessCount, s.Warning, s.Critical))
	case s.ProcessCount <= s.WarnMin:
		s.Check.Warning(fmt.Sprintf("%d MySQL processes are below threshold of %d | mysql_processes=%d;%s;%s;0", s.ProcessCount, s.WarnMin, s.ProcessCount, s.Warning, s.Critical))
	default:
		s.Check.Ok(fmt.Sprintf("MySQL process Count %d | mysql_processes=%d;%s;%s;0", s.ProcessCount, s.ProcessCount, s.Warning, s.Critical))
	}
}

func selectProcessCount(connection common.Connection) (int64, error) {
	source := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", connection.User, connection.Password, connection.Host, connection.Port, connection.Database)
	db, err := sql.Open("mysql", source)
	if err != nil {
		return 0, err
	}
	defer db.Close()

	return execProcessCount(db)
}

// execProcessCount reads the process count from an open database handle.
func execProcessCount(db *sql.DB) (int64, error) {
	var count int64
	if err := db.QueryRow("select count(*) from information_schema.PROCESSLIST").Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}
