package main

import (
    "fmt"
    "os"
    "path/filepath"

    "github.com/portertech/sensu-plugins-go/lib/check"
)

func main() {
	var (
		queue string
		warn int
		crit int
	)

	c := check.New("CheckPostfixQueue")
	c.Option.StringVarP(&queue, "queue", "q", "deferred", "Postfix queue to check")
	c.Option.IntVarP(&warn, "warn", "w", 5, "Warning threshold")
	c.Option.IntVarP(&crit, "crit", "c", 10, "Critical threshold")
	c.Init()

        queueDir := fmt.Sprintf("/var/spool/postfix/%s", queue)

        if _, err := os.Stat(queueDir); os.IsNotExist(err) || os.IsPermission(err) {
            c.Error(fmt.Errorf("Cannot access queue directory %s: %v", queueDir, err))
        }

	queueLength, err := mailQueue(queueDir)
	if err != nil {
		c.Error(err)
	}

	switch {
	case queueLength > crit:
		c.Critical(fmt.Sprintf("%d messages in the postfix mail queue", queueLength))
	case queueLength > warn:
		c.Warning(fmt.Sprintf("%d messages in the postfix mail queue", queueLength))
	default:
		c.Ok(fmt.Sprintf("%d messages in the postfix mail queue", queueLength))
        }
}

func mailQueue(path string) (int, error) {

    var files []string

    err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
        if info.IsDir() {
            return nil
        }
        files = append(files, path)
        return nil
    })

    if err != nil {
        panic(err)
    }

    return len(files), nil
}

