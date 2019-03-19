package main

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/portertech/sensu-plugins-go/lib/check"
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
	}

	output, perf := formatOutput(usage, warn, crit)
	switch {
	case usage[3] <= float64(100-crit):
		c.Critical(fmt.Sprintf("%s | %s", output, perf))
	case usage[3] <= float64(100-warn):
		c.Warning(fmt.Sprintf("%s | %s", output, perf))
	default:
		c.Ok(fmt.Sprintf("%s | %s", output, perf))
	}
}

func cpuUsage(sleep int) ([]float64, error) {
	var totalDiff float64

	beforeStats, err := getStats()
	if err != nil {
		return []float64{}, err
	}

	time.Sleep(time.Duration(sleep) * time.Second)

	afterStats, err := getStats()
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

func getStats() ([]float64, error) {
	contents, err := ioutil.ReadFile("/proc/stat")
	if err != nil {
		return []float64{}, err
	}

	line := strings.Split(string(contents), "\n")[0]
	stats := strings.Fields(line)[1:]

	result := make([]float64, len(stats))
	for i := range stats {
		result[i], err = strconv.ParseFloat(stats[i], 64)
		if err != nil {
			return result, err
		}
	}

	return result, nil
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
	perf := fmt.Sprintf("cpu_user=%.2f%%;%d;%d; cpu_system=%.2f%%;%d;%d; cpu_iowait=%.2f%%;%d;%d; cpu_other=%.2f%%;%d;%d; cpu_idle=%.2f%%;", usageStats[0]+usageStats[1], warn, crit, usageStats[2], warn, crit, usageStats[4], warn, crit, other, warn, crit, usageStats[3])
	return output, perf
}
