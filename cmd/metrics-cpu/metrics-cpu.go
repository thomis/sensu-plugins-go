package main

import (
	"time"

	"github.com/thomis/sensu-plugins-go/pkg/metrics"
	"github.com/thomis/sensu-plugins-go/pkg/common"
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

	beforeStats, err := common.GetStats()
	if err != nil {
		return usage, err
	}

	time.Sleep(time.Duration(sleep) * time.Second)

	afterStats, err := common.GetStats()
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
