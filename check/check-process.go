package main

import (
	"fmt"
	"regexp"

	"github.com/mitchellh/go-ps"
	"github.com/portertech/sensu-plugins-go/lib/check"
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

	processes, err := ps.Processes()
	if err != nil {
		return count, err
	}

	for _, process := range processes {
		if re.Match([]byte(process.Executable())) {
			fmt.Printf(" - %s\n", process.Executable())
			count += 1
		}
	}

	return count, nil
}
