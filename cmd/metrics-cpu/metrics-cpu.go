package main

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/thomis/sensu-plugins-go/pkg/metrics"
)

func main() {
	var sleep int

	m := metrics.New("cpu.usage")
	m.Option.IntVarP(&sleep, "sleep", "s", 1, "SLEEP")
	m.Init()

	usage, err := cpuUsage(sleep)
	if err == nil {
		m.Print(usage)
	}
}

func cpuUsage(sleep int) (float64, error) {
	var usage, totalDiff float64

	beforeStats, err := getStats()
	if err != nil {
		return usage, err
	}

	time.Sleep(time.Duration(sleep) * time.Second)

	afterStats, err := getStats()
	if err != nil {
		return usage, err
	}

	diffStats := make([]float64, len(beforeStats))
	for i := range beforeStats {
		diffStats[i] = afterStats[i] - beforeStats[i]
		totalDiff += diffStats[i]
	}

	usage = 100.0 * (totalDiff - diffStats[3]) / totalDiff
	return usage, nil
}

func getStats() ([]float64, error) {
	contents, err := os.ReadFile("/proc/stat")
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
