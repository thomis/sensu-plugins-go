package main

import (
	"fmt"
	"strings"

	"github.com/shirou/gopsutil/v3/host"
	"github.com/thomis/sensu-plugins-go/pkg/check"
)

func main() {
	c := check.New("CheckUptime")
	c.Init()

	uptime, _ := host.Uptime()

	days := uptime / (60 * 60 * 24)
	hours := (uptime - (days * 60 * 60 * 24)) / (60 * 60)
	minutes := ((uptime - (days * 60 * 60 * 24)) - (hours * 60 * 60)) / 60
	seconds := uptime - (days * 60 * 60 * 24) - (hours * 60 * 60) - (minutes * 60)

	elements := []string{}

	if days > 1 {
		elements = append(elements, fmt.Sprintf("%d days", days))
	}

	if days == 1 {
		elements = append(elements, fmt.Sprintf("%d day", days))
	}

	if hours > 1 {
		elements = append(elements, fmt.Sprintf("%d hours", hours))
	}

	if hours == 1 {
		elements = append(elements, fmt.Sprintf("%d hour", hours))
	}

	if minutes > 1 {
		elements = append(elements, fmt.Sprintf("%d minutes", minutes))
	}

	if minutes == 1 {
		elements = append(elements, fmt.Sprintf("%d minute", minutes))
	}

	if seconds == 1 {
		elements = append(elements, fmt.Sprintf("%d second", seconds))
	} else {
		elements = append(elements, fmt.Sprintf("%d seconds", seconds))
	}

	c.Ok(fmt.Sprintf("Uptime is %s", strings.Join(elements, ", ")))
}
