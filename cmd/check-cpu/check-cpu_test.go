package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test CPU usage calculation logic
func TestCpuUsageCalculation(t *testing.T) {
	tests := []struct {
		name          string
		beforeStats   []float64
		afterStats    []float64
		expectedUsage []float64
	}{
		{
			name: "Simple CPU stats",
			// user, nice, system, idle, iowait
			beforeStats: []float64{1000, 100, 500, 8000, 400},
			afterStats:  []float64{1100, 100, 600, 8100, 500},
			// Diff: 100, 0, 100, 100, 100 = total 400
			// Usage calculation: 100 - 100*(400-diff[i])/400
			// user: 100 - 100*(400-100)/400 = 100 - 75 = 25%
			// nice: 100 - 100*(400-0)/400 = 100 - 100 = 0%
			// system: 100 - 100*(400-100)/400 = 100 - 75 = 25%
			// idle: 100 - 100*(400-100)/400 = 100 - 75 = 25%
			// iowait: 100 - 100*(400-100)/400 = 100 - 75 = 25%
			expectedUsage: []float64{25, 0, 25, 25, 25},
		},
		{
			name:        "High idle CPU",
			beforeStats: []float64{1000, 0, 500, 8500, 0},
			afterStats:  []float64{1010, 0, 510, 8970, 10},
			// Diff: 10, 0, 10, 470, 10 = total 500
			// Usage calculation: 100 - 100*(500-diff[i])/500
			// user: 100 - 100*(500-10)/500 = 100 - 98 = 2%
			// nice: 100 - 100*(500-0)/500 = 100 - 100 = 0%
			// system: 100 - 100*(500-10)/500 = 100 - 98 = 2%
			// idle: 100 - 100*(500-470)/500 = 100 - 6 = 94%
			// iowait: 100 - 100*(500-10)/500 = 100 - 98 = 2%
			expectedUsage: []float64{2, 0, 2, 94, 2},
		},
		{
			name:        "All fields changing",
			beforeStats: []float64{1000, 100, 500, 7000, 400, 100, 100, 100, 100},
			afterStats:  []float64{1200, 150, 600, 7800, 500, 150, 150, 150, 150},
			// Diff: 200, 50, 100, 800, 100, 50, 50, 50, 50 = total 1450
			// Usage calculation: 100 - 100*(1450-diff[i])/1450
			// user: 100 - 100*(1450-200)/1450 = 100 - 86.21 = 13.79%
			// nice: 100 - 100*(1450-50)/1450 = 100 - 96.55 = 3.45%
			// system: 100 - 100*(1450-100)/1450 = 100 - 93.10 = 6.90%
			// idle: 100 - 100*(1450-800)/1450 = 100 - 44.83 = 55.17%
			// iowait: 100 - 100*(1450-100)/1450 = 100 - 93.10 = 6.90%
			expectedUsage: []float64{13.79, 3.45, 6.90, 55.17, 6.90, 3.45, 3.45, 3.45, 3.45},
		},
		{
			name:        "No change in stats",
			beforeStats: []float64{1000, 100, 500, 8000, 400},
			afterStats:  []float64{1000, 100, 500, 8000, 400},
			// No diff, should handle division by zero
			expectedUsage: []float64{100, 100, 100, 100, 100}, // or could be NaN handling
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usage := calculateCpuUsage(tt.beforeStats, tt.afterStats)

			assert.Equal(t, len(tt.beforeStats), len(usage))

			for i, expected := range tt.expectedUsage {
				assert.InDelta(t, expected, usage[i], 0.01, "Mismatch at index %d", i)
			}
		})
	}
}

// Helper function that mimics cpuUsage calculation
func calculateCpuUsage(beforeStats, afterStats []float64) []float64 {
	var totalDiff float64

	diffStats := make([]float64, len(beforeStats))
	for i := range beforeStats {
		diffStats[i] = afterStats[i] - beforeStats[i]
		totalDiff += diffStats[i]
	}

	usageStats := make([]float64, len(beforeStats))
	if totalDiff == 0 {
		// Handle division by zero
		for i := range usageStats {
			usageStats[i] = 100.0
		}
	} else {
		for i := range diffStats {
			usageStats[i] = 100.0 - (100.0 * (totalDiff - diffStats[i]) / totalDiff)
		}
	}
	return usageStats
}

// Test formatOutput function
func TestFormatOutput(t *testing.T) {
	tests := []struct {
		name           string
		usageStats     []float64
		warn           int
		crit           int
		expectedOutput string
		expectedPerf   string
	}{
		{
			name:           "Standard 5-field CPU stats",
			usageStats:     []float64{10.50, 2.50, 5.00, 80.00, 2.00},
			warn:           80,
			crit:           90,
			expectedOutput: "user=13.00% system=5.00% iowait=2.00% other=0.00% idle=80.00%",
			expectedPerf:   "cpu_user=13.00%;80;90 cpu_system=5.00%;80;90 cpu_iowait=2.00%;80;90 cpu_other=0.00%;80;90 cpu_idle=80.00%",
		},
		{
			name:           "CPU stats with additional fields",
			usageStats:     []float64{10.00, 0.00, 5.00, 75.00, 2.00, 3.00, 2.00, 1.00, 2.00},
			warn:           80,
			crit:           90,
			expectedOutput: "user=10.00% system=5.00% iowait=2.00% other=8.00% idle=75.00%",
			expectedPerf:   "cpu_user=10.00%;80;90 cpu_system=5.00%;80;90 cpu_iowait=2.00%;80;90 cpu_other=8.00%;80;90 cpu_idle=75.00%",
		},
		{
			name:           "High CPU usage",
			usageStats:     []float64{40.00, 10.00, 30.00, 15.00, 5.00},
			warn:           80,
			crit:           90,
			expectedOutput: "user=50.00% system=30.00% iowait=5.00% other=0.00% idle=15.00%",
			expectedPerf:   "cpu_user=50.00%;80;90 cpu_system=30.00%;80;90 cpu_iowait=5.00%;80;90 cpu_other=0.00%;80;90 cpu_idle=15.00%",
		},
		{
			name:           "Zero values",
			usageStats:     []float64{0.00, 0.00, 0.00, 100.00, 0.00},
			warn:           80,
			crit:           90,
			expectedOutput: "user=0.00% system=0.00% iowait=0.00% other=0.00% idle=100.00%",
			expectedPerf:   "cpu_user=0.00%;80;90 cpu_system=0.00%;80;90 cpu_iowait=0.00%;80;90 cpu_other=0.00%;80;90 cpu_idle=100.00%",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, perf := formatOutput(tt.usageStats, tt.warn, tt.crit)
			assert.Equal(t, tt.expectedOutput, output)
			assert.Equal(t, tt.expectedPerf, perf)
		})
	}
}

// Test threshold evaluation based on idle CPU
func TestThresholdEvaluation(t *testing.T) {
	tests := []struct {
		name     string
		idleCpu  float64
		warn     int
		crit     int
		expected string // "ok", "warning", or "critical"
	}{
		{
			name:     "High idle - OK",
			idleCpu:  80.0,
			warn:     80,
			crit:     90,
			expected: "ok",
		},
		{
			name:     "Exactly at warning threshold",
			idleCpu:  20.0, // 100-80 = 20
			warn:     80,
			crit:     90,
			expected: "warning",
		},
		{
			name:     "Between warning and critical",
			idleCpu:  15.0,
			warn:     80,
			crit:     90,
			expected: "warning",
		},
		{
			name:     "Exactly at critical threshold",
			idleCpu:  10.0, // 100-90 = 10
			warn:     80,
			crit:     90,
			expected: "critical",
		},
		{
			name:     "Below critical threshold",
			idleCpu:  5.0,
			warn:     80,
			crit:     90,
			expected: "critical",
		},
		{
			name:     "Edge case - just above warning",
			idleCpu:  20.01,
			warn:     80,
			crit:     90,
			expected: "ok",
		},
		{
			name:     "Edge case - just below warning",
			idleCpu:  19.99,
			warn:     80,
			crit:     90,
			expected: "warning",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var status string
			switch {
			case tt.idleCpu <= float64(100-tt.crit):
				status = "critical"
			case tt.idleCpu <= float64(100-tt.warn):
				status = "warning"
			default:
				status = "ok"
			}
			assert.Equal(t, tt.expected, status)
		})
	}
}

// Test CPU stats array handling
func TestCpuStatsArrayHandling(t *testing.T) {
	tests := []struct {
		name           string
		stats          []float64
		expectedUser   float64
		expectedSystem float64
		expectedIdle   float64
		expectedIowait float64
		expectedOther  float64
	}{
		{
			name:           "5-field stats (basic)",
			stats:          []float64{10.0, 2.0, 5.0, 80.0, 3.0},
			expectedUser:   12.0, // user + nice
			expectedSystem: 5.0,
			expectedIdle:   80.0,
			expectedIowait: 3.0,
			expectedOther:  0.0,
		},
		{
			name:           "9-field stats (extended)",
			stats:          []float64{10.0, 2.0, 5.0, 70.0, 3.0, 2.0, 3.0, 2.0, 3.0},
			expectedUser:   12.0, // user + nice
			expectedSystem: 5.0,
			expectedIdle:   70.0,
			expectedIowait: 3.0,
			expectedOther:  10.0, // sum of fields 5-8
		},
		{
			name:           "10-field stats",
			stats:          []float64{8.0, 1.0, 4.0, 75.0, 2.0, 1.0, 2.0, 3.0, 2.0, 2.0},
			expectedUser:   9.0, // user + nice
			expectedSystem: 4.0,
			expectedIdle:   75.0,
			expectedIowait: 2.0,
			expectedOther:  10.0, // sum of fields 5-9
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var other float64
			for i := range tt.stats[5:] {
				other += tt.stats[5+i]
			}

			user := tt.stats[0] + tt.stats[1]
			system := tt.stats[2]
			idle := tt.stats[3]
			iowait := tt.stats[4]

			assert.Equal(t, tt.expectedUser, user)
			assert.Equal(t, tt.expectedSystem, system)
			assert.Equal(t, tt.expectedIdle, idle)
			assert.Equal(t, tt.expectedIowait, iowait)
			assert.Equal(t, tt.expectedOther, other)
		})
	}
}

// Test edge cases
func TestCpuUsageEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		beforeStats []float64
		afterStats  []float64
		shouldPanic bool
	}{
		{
			name:        "Empty stats arrays",
			beforeStats: []float64{},
			afterStats:  []float64{},
			shouldPanic: false,
		},
		{
			name:        "Mismatched array lengths",
			beforeStats: []float64{1, 2, 3, 4, 5},
			afterStats:  []float64{1, 2, 3},
			shouldPanic: true,
		},
		{
			name:        "Negative differences",
			beforeStats: []float64{100, 50, 200, 7000, 100},
			afterStats:  []float64{90, 40, 190, 7100, 110}, // Some values decreased
			shouldPanic: false,                             // Should handle gracefully
		},
		{
			name:        "Very large values",
			beforeStats: []float64{1e10, 1e9, 1e10, 1e11, 1e9},
			afterStats:  []float64{1e10 + 1000, 1e9 + 100, 1e10 + 1000, 1e11 + 10000, 1e9 + 1000},
			shouldPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				assert.Panics(t, func() {
					_ = calculateCpuUsage(tt.beforeStats, tt.afterStats)
				})
			} else {
				assert.NotPanics(t, func() {
					usage := calculateCpuUsage(tt.beforeStats, tt.afterStats)
					if len(tt.beforeStats) > 0 {
						assert.Equal(t, len(tt.beforeStats), len(usage))
					}
				})
			}
		})
	}
}

// Test formatting precision
func TestFormattingPrecision(t *testing.T) {
	tests := []struct {
		value    float64
		expected string
	}{
		{10.0, "10.00"},
		{10.5, "10.50"},
		{10.555, "10.55"}, // Banker's rounding
		{10.554, "10.55"},
		{0.0, "0.00"},
		{99.999, "100.00"},
		{100.0, "100.00"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%.2f", tt.value), func(t *testing.T) {
			result := fmt.Sprintf("%.2f", tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}
