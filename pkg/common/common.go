package common

import (
	"os"
	"strconv"
	"strings"
)

func GetStats() ([]float64, error) {
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
