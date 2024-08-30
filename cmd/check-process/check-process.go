package main

import (
	"fmt"
	"os"
	"regexp"

	"github.com/shirou/gopsutil/v4/process"
	"github.com/thomis/sensu-plugins-go/pkg/check"
)

func main() {
	var (
		pattern string
		count   int
	)

	c := check.New("CheckProcess")
	c.Option.StringVarP(&pattern, "regexp_pattern", "p", "a_process_name", "PATTERN")
	c.Init()

	count, err := countProcess(pattern)
	if err != nil {
		c.Error(err)
	}

	switch {
	case count == 0:
		c.Critical(fmt.Sprintf("Unable to find process [%s]", pattern))
	default:
		c.Ok(fmt.Sprintf("Process [%s]: %d occurence(s)", pattern, count))
	}
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
