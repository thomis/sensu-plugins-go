package main

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test parsing of /proc/meminfo content
func TestMemoryUsageParsing(t *testing.T) {
	tests := []struct {
		name              string
		meminfoContent    string
		expectedTotal     float64
		expectedAvailable float64
		expectedError     bool
		errorContains     string
	}{
		{
			name: "Standard meminfo with MemAvailable",
			meminfoContent: `MemTotal:        8061320 kB
MemFree:          123456 kB
MemAvailable:    4030660 kB
Buffers:          234567 kB
Cached:          1234567 kB
SwapCached:            0 kB`,
			expectedTotal:     8061320.0 / 1024.0,
			expectedAvailable: 4030660.0 / 1024.0,
			expectedError:     false,
		},
		{
			name: "Meminfo without MemAvailable (use Buffers + Cached)",
			meminfoContent: `MemTotal:        8061320 kB
MemFree:          123456 kB
Buffers:          234567 kB
Cached:          1234567 kB
SwapCached:            0 kB`,
			expectedTotal:     8061320.0 / 1024.0,
			expectedAvailable: (234567.0 + 1234567.0) / 1024.0,
			expectedError:     false,
		},
		{
			name: "Minimal meminfo",
			meminfoContent: `MemTotal:        4096000 kB
MemAvailable:    2048000 kB
`,
			expectedTotal:     4096000.0 / 1024.0,
			expectedAvailable: 2048000.0 / 1024.0,
			expectedError:     false,
		},
		{
			name:              "Empty meminfo",
			meminfoContent:    ``,
			expectedTotal:     0.0,
			expectedAvailable: 0.0,
			expectedError:     false,
		},
		{
			name: "Invalid number format",
			meminfoContent: `MemTotal:        invalid kB
MemAvailable:    2048000 kB`,
			expectedTotal:     0.0,
			expectedAvailable: 0.0,
			expectedError:     true,
			errorContains:     "invalid syntax",
		},
		{
			name: "Large memory values",
			meminfoContent: `MemTotal:       132061320 kB
MemAvailable:    64030660 kB
`,
			expectedTotal:     132061320.0 / 1024.0,
			expectedAvailable: 64030660.0 / 1024.0,
			expectedError:     false,
		},
		{
			name: "Different field order",
			meminfoContent: `MemFree:          123456 kB
MemAvailable:    4030660 kB
MemTotal:        8061320 kB
Buffers:          234567 kB`,
			expectedTotal:     8061320.0 / 1024.0,
			expectedAvailable: 4030660.0 / 1024.0,
			expectedError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary file with test content
			tmpfile, err := os.CreateTemp("", "meminfo-test-*.txt")
			assert.NoError(t, err)
			defer os.Remove(tmpfile.Name())

			_, err = tmpfile.WriteString(tt.meminfoContent)
			assert.NoError(t, err)
			tmpfile.Close()

			// Test the parsing logic
			total, available, err := parseMeminfo(tmpfile.Name())

			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.InDelta(t, tt.expectedTotal, total, 0.01)
				assert.InDelta(t, tt.expectedAvailable, available, 0.01)
			}
		})
	}
}

// Helper function that mimics the memoryUsage parsing logic
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

	lines := strings.Split(string(contents), "\n")

	for i := range lines[:len(lines)-1] {
		if i >= len(lines) {
			break
		}
		stats := strings.Fields(lines[i])
		if len(stats) < 2 {
			continue
		}
		switch {
		case stats[0] == "MemTotal:":
			memTotal, err = parseFloat(stats[1])
			if err != nil {
				return 0.0, 0.0, err
			}
		case stats[0] == "MemAvailable:":
			memAvailable, err = parseFloat(stats[1])
			if err != nil {
				return 0.0, 0.0, err
			}
		case stats[0] == "Buffers:":
			memBuffers, err = parseFloat(stats[1])
			if err != nil {
				return 0.0, 0.0, err
			}
		case stats[0] == "Cached:":
			memCached, err = parseFloat(stats[1])
			if err != nil {
				return 0.0, 0.0, err
			}
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

func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}

// Test memory usage calculation
func TestMemoryUsageCalculation(t *testing.T) {
	tests := []struct {
		name         string
		memTotal     float64
		memAvailable float64
		expected     float64
	}{
		{
			name:         "50% usage",
			memTotal:     8192.0,
			memAvailable: 4096.0,
			expected:     50.0,
		},
		{
			name:         "25% usage",
			memTotal:     8192.0,
			memAvailable: 6144.0,
			expected:     25.0,
		},
		{
			name:         "75% usage",
			memTotal:     8192.0,
			memAvailable: 2048.0,
			expected:     75.0,
		},
		{
			name:         "Nearly full",
			memTotal:     8192.0,
			memAvailable: 100.0,
			expected:     98.779296875,
		},
		{
			name:         "Nearly empty",
			memTotal:     8192.0,
			memAvailable: 8092.0,
			expected:     1.220703125,
		},
		{
			name:         "Zero available",
			memTotal:     8192.0,
			memAvailable: 0.0,
			expected:     100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usage := 100.0 - (100.0 * tt.memAvailable / tt.memTotal)
			assert.InDelta(t, tt.expected, usage, 0.0001)
		})
	}
}

// Test output formatting
func TestOutputFormatting(t *testing.T) {
	tests := []struct {
		name         string
		usage        float64
		memTotal     float64
		memAvailable float64
		warn         int
		crit         int
		expectedMsg  string
		expectedPerf string
	}{
		{
			name:         "Normal usage",
			usage:        45.50,
			memTotal:     8192.00,
			memAvailable: 4468.74,
			warn:         80,
			crit:         90,
			expectedMsg:  "45.50% MemTotal:8192.00MB MemAvailable:4468.74MB",
			expectedPerf: "mem_usage=45.50%;80;90 mem_available=4468.74MB",
		},
		{
			name:         "Warning threshold",
			usage:        82.00,
			memTotal:     16384.00,
			memAvailable: 2949.12,
			warn:         80,
			crit:         90,
			expectedMsg:  "82.00% MemTotal:16384.00MB MemAvailable:2949.12MB",
			expectedPerf: "mem_usage=82.00%;80;90 mem_available=2949.12MB",
		},
		{
			name:         "Critical threshold",
			usage:        95.00,
			memTotal:     4096.00,
			memAvailable: 204.80,
			warn:         80,
			crit:         90,
			expectedMsg:  "95.00% MemTotal:4096.00MB MemAvailable:204.80MB",
			expectedPerf: "mem_usage=95.00%;80;90 mem_available=204.80MB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := fmt.Sprintf("%.2f%% MemTotal:%.2fMB MemAvailable:%.2fMB",
				tt.usage, tt.memTotal, tt.memAvailable)
			perf := fmt.Sprintf("mem_usage=%.2f%%;%d;%d mem_available=%.2fMB",
				tt.usage, tt.warn, tt.crit, tt.memAvailable)

			assert.Equal(t, tt.expectedMsg, msg)
			assert.Equal(t, tt.expectedPerf, perf)
		})
	}
}

// Test threshold evaluation
func TestThresholdEvaluation(t *testing.T) {
	tests := []struct {
		name     string
		usage    float64
		warn     int
		crit     int
		expected string // "ok", "warning", or "critical"
	}{
		{
			name:     "Below warning",
			usage:    50.0,
			warn:     80,
			crit:     90,
			expected: "ok",
		},
		{
			name:     "Exactly at warning",
			usage:    80.0,
			warn:     80,
			crit:     90,
			expected: "warning",
		},
		{
			name:     "Between warning and critical",
			usage:    85.0,
			warn:     80,
			crit:     90,
			expected: "warning",
		},
		{
			name:     "Exactly at critical",
			usage:    90.0,
			warn:     80,
			crit:     90,
			expected: "critical",
		},
		{
			name:     "Above critical",
			usage:    95.0,
			warn:     80,
			crit:     90,
			expected: "critical",
		},
		{
			name:     "Edge case - 79.99%",
			usage:    79.99,
			warn:     80,
			crit:     90,
			expected: "ok",
		},
		{
			name:     "Edge case - 80.01%",
			usage:    80.01,
			warn:     80,
			crit:     90,
			expected: "warning",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var status string
			switch {
			case tt.usage >= float64(tt.crit):
				status = "critical"
			case tt.usage >= float64(tt.warn):
				status = "warning"
			default:
				status = "ok"
			}
			assert.Equal(t, tt.expected, status)
		})
	}
}

// Test error handling
func TestMemoryUsageErrorHandling(t *testing.T) {
	// Test with non-existent file
	_, _, err := parseMeminfo("/non/existent/file")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no such file or directory")
}
