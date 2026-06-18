package main

import (
	"fmt"
	"time"

	"github.com/thomis/sensu-plugins-go/pkg/check"
	"github.com/thomis/sensu-plugins-go/pkg/common"
)

func main() {
	var (
		warn  int
		crit  int
		sleep int
	)

	c := check.New("CheckCPU")
	c.Option.IntVarP(&warn, "warn", "w", 80, "Warning threshold")
	c.Option.IntVarP(&crit, "crit", "c", 90, "Critical threshold")
	c.Option.IntVarP(&sleep, "sleep", "s", 1, "Sleep time for sampling")
	c.Init()

	usage, err := cpuUsage(sleep)
	if err != nil {
		c.Error(err)
		return
	}

	level, message := evaluate(usage, warn, crit)
	report(c, level, message)
}

// report maps a level (ok|warning|critical) to the matching check result.
func report(c *check.CheckStruct, level string, message string) {
	switch level {
	case "critical":
		c.Critical(message)
	case "warning":
		c.Warning(message)
	default:
		c.Ok(message)
	}
}

// evaluate formats the CPU usage and decides the level (ok|warning|critical)
// based on the idle percentage (usageStats[3]) against the thresholds.
func evaluate(usageStats []float64, warn, crit int) (string, string) {
	output, perf := formatOutput(usageStats, warn, crit)
	message := fmt.Sprintf("%s | %s", output, perf)

	switch {
	case usageStats[3] <= float64(100-crit):
		return "critical", message
	case usageStats[3] <= float64(100-warn):
		return "warning", message
	default:
		return "ok", message
	}
}

func cpuUsage(sleep int) ([]float64, error) {
	var totalDiff float64

	beforeStats, err := common.GetStats()
	if err != nil {
		return []float64{}, err
	}

	time.Sleep(time.Duration(sleep) * time.Second)

	afterStats, err := common.GetStats()
	if err != nil {
		return []float64{}, err
	}

	diffStats := make([]float64, len(beforeStats))
	for i := range beforeStats {
		diffStats[i] = afterStats[i] - beforeStats[i]
		totalDiff += diffStats[i]
	}

	usageStats := make([]float64, len(beforeStats))
	for i := range diffStats {
		usageStats[i] = 100.0 - (100.0 * (totalDiff - diffStats[i]) / totalDiff)
	}
	return usageStats, nil
}

func formatOutput(usageStats []float64, warn int, crit int) (string, string) {

	var other float64

	// This is based on
	// 0 - user
	// 1 - nice
	// 2 - system
	// 3 - idle
	// 4 - iowait
	// 5 - 9 (possible) other fields

	for i := range usageStats[5:] {
		other += usageStats[5+i]
	}
	output := fmt.Sprintf("user=%.2f%% system=%.2f%% iowait=%.2f%% other=%.2f%% idle=%.2f%%", usageStats[0]+usageStats[1], usageStats[2], usageStats[4], other, usageStats[3])
	perf := fmt.Sprintf("cpu_user=%.2f%%;%d;%d cpu_system=%.2f%%;%d;%d cpu_iowait=%.2f%%;%d;%d cpu_other=%.2f%%;%d;%d cpu_idle=%.2f%%", usageStats[0]+usageStats[1], warn, crit, usageStats[2], warn, crit, usageStats[4], warn, crit, other, warn, crit, usageStats[3])
	return output, perf
}
