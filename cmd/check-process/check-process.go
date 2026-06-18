package main

import (
	"fmt"
	"os"
	"regexp"

	"github.com/shirou/gopsutil/v4/process"
	"github.com/thomis/sensu-plugins-go/pkg/check"
)

func main() {
	var pattern string

	c := check.New("CheckProcess")
	c.Option.StringVarP(&pattern, "regexp_pattern", "p", "a_process_name", "PATTERN")
	c.Init()

	count, err := countProcess(pattern)
	if err != nil {
		c.Error(err)
		return
	}

	level, message := describe(pattern, count)
	report(c, level, message)
}

// report maps a level (ok|critical) to the matching check result.
func report(c *check.CheckStruct, level string, message string) {
	switch level {
	case "critical":
		c.Critical(message)
	default:
		c.Ok(message)
	}
}

// describe turns a match count into a level and message.
func describe(pattern string, count int) (string, string) {
	if count == 0 {
		return "critical", fmt.Sprintf("Unable to find process [%s]", pattern)
	}
	return "ok", fmt.Sprintf("Process [%s]: %d occurence(s)", pattern, count)
}

func countProcess(pattern string) (int, error) {
	count := 0

	re, err := regexp.Compile(pattern)
	if err != nil {
		return count, err
	}

	processes, err := process.Processes()
	if err != nil {
		return count, err
	}

	pid := os.Getpid()

	for _, process := range processes {
		if int32(pid) == process.Pid {
			continue
		}
		cmdLine, _ := process.Cmdline()
		if re.Match([]byte(cmdLine)) {
			fmt.Printf(" - (%d) %s\n", process.Pid, cmdLine)
			count += 1
		}
	}

	return count, nil
}
