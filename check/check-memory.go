package main

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/portertech/sensu-plugins-go/lib/check"
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
	}

	usage := 100.0 - (100.0 * memAvailable / memTotal)

	switch {
	case usage >= float64(crit):
		c.Critical(fmt.Sprintf("%.2f%% MemTotal:%.2fMB MemAvailable:%.2fMB | mem_usage=%.2f%%;%d;%d mem_available=%.2fMB", usage, memTotal, memAvailable, usage, warn, crit, memAvailable))
	case usage >= float64(warn):
		c.Warning(fmt.Sprintf("%.2f%% MemTotal:%.2fMB MemAvailable:%.2fMB | mem_usage=%.2f%%;%d;%d mem_available=%.2fMB", usage, memTotal, memAvailable, usage, warn, crit, memAvailable))
	default:
		c.Ok(fmt.Sprintf("%.2f%% MemTotal:%.2fMB MemAvailable:%.2fMB | mem_usage=%.2f%%;%d;%d mem_available=%.2fMB", usage, memTotal, memAvailable, usage, warn, crit, memAvailable))
	}
}

func memoryUsage() (float64, float64, error) {
	var (
		memTotal     float64
		memAvailable float64
	)

	contents, err := ioutil.ReadFile("/proc/meminfo")
	if err != nil {
		return 0.0, 0.0, err
	}

	lines := strings.Split(string(contents), "\n")

	for i := range lines[:len(lines)-1] {
		stats := strings.Fields(lines[i])
		switch {
		case stats[0] == "MemTotal:":
			memTotal, err = strconv.ParseFloat(stats[1], 64)
			if err != nil {
				return 0.0, 0.0, err
			}
		case stats[0] == "MemAvailable:":
			memAvailable, err = strconv.ParseFloat(stats[1], 64)
			if err != nil {
				return 0.0, 0.0, err
			}
		}
		if memTotal > 0.0 && memAvailable > 0.0 {
			break
		}
	}

	return memTotal / 1024.0, memAvailable / 1024.0, nil
}
