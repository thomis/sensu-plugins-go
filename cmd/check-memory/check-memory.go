package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/thomis/sensu-plugins-go/pkg/check"
)

func main() {
	var (
		warn int
		crit int
	)

	c := check.New("CheckMemory")
	c.Option.IntVarP(&warn, "warn", "w", 80, "Warning (>=) threshold level")
	c.Option.IntVarP(&crit, "crit", "c", 90, "Critical (>=) threshold level")
	c.Init()

	memTotal, memAvailable, err := memoryUsage()
	if err != nil {
		c.Error(err)
		return
	}

	level, message := evaluate(memTotal, memAvailable, warn, crit)
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

// evaluate computes the memory usage percentage and returns the level
// (ok|warning|critical) together with the formatted output and perfdata.
func evaluate(memTotal, memAvailable float64, warn, crit int) (string, string) {
	usage := 100.0 - (100.0 * memAvailable / memTotal)

	message := fmt.Sprintf("%.2f%% MemTotal:%.2fMB MemAvailable:%.2fMB | mem_usage=%.2f%%;%d;%d mem_available=%.2fMB",
		usage, memTotal, memAvailable, usage, warn, crit, memAvailable)

	switch {
	case usage >= float64(crit):
		return "critical", message
	case usage >= float64(warn):
		return "warning", message
	default:
		return "ok", message
	}
}

func memoryUsage() (float64, float64, error) {
	return parseMeminfo("/proc/meminfo")
}

// parseMeminfo reads a meminfo-formatted file and returns total and available
// memory in MB. Systems without MemAvailable fall back to Buffers + Cached.
func parseMeminfo(filepath string) (float64, float64, error) {
	var (
		memTotal     float64
		memAvailable float64
		memBuffers   float64
		memCached    float64
	)

	contents, err := os.ReadFile(filepath)
	if err != nil {
		return 0.0, 0.0, err
	}

	for _, line := range strings.Split(string(contents), "\n") {
		stats := strings.Fields(line)
		if len(stats) < 2 {
			continue
		}

		switch stats[0] {
		case "MemTotal:":
			memTotal, err = strconv.ParseFloat(stats[1], 64)
		case "MemAvailable:":
			memAvailable, err = strconv.ParseFloat(stats[1], 64)
		case "Buffers:":
			memBuffers, err = strconv.ParseFloat(stats[1], 64)
		case "Cached:":
			memCached, err = strconv.ParseFloat(stats[1], 64)
		}
		if err != nil {
			return 0.0, 0.0, err
		}

		if memTotal > 0.0 && memAvailable > 0.0 {
			break
		}
	}

	// Deal with systems that don't provide MemAvailable
	if memAvailable == 0 && (memBuffers != 0 || memCached != 0) {
		memAvailable = memBuffers + memCached
	}

	return memTotal / 1024.0, memAvailable / 1024.0, nil
}
