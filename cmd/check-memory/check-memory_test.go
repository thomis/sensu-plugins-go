package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomis/sensu-plugins-go/pkg/check"
)

func TestParseMeminfo(t *testing.T) {
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
			meminfoContent: "MemTotal:        8061320 kB\nMemFree:          123456 kB\n" +
				"MemAvailable:    4030660 kB\nBuffers:          234567 kB\nCached:          1234567 kB\n",
			expectedTotal:     8061320.0 / 1024.0,
			expectedAvailable: 4030660.0 / 1024.0,
		},
		{
			name: "Without MemAvailable (Buffers + Cached)",
			meminfoContent: "MemTotal:        8061320 kB\nMemFree:          123456 kB\n" +
				"Buffers:          234567 kB\nCached:          1234567 kB\n",
			expectedTotal:     8061320.0 / 1024.0,
			expectedAvailable: (234567.0 + 1234567.0) / 1024.0,
		},
		{
			name:              "Empty meminfo",
			meminfoContent:    "",
			expectedTotal:     0.0,
			expectedAvailable: 0.0,
		},
		{
			name:           "Invalid number format",
			meminfoContent: "MemTotal:        invalid kB\nMemAvailable:    2048000 kB\n",
			expectedError:  true,
			errorContains:  "invalid syntax",
		},
		{
			name: "Different field order",
			meminfoContent: "MemFree:          123456 kB\nMemAvailable:    4030660 kB\n" +
				"MemTotal:        8061320 kB\nBuffers:          234567 kB\n",
			expectedTotal:     8061320.0 / 1024.0,
			expectedAvailable: 4030660.0 / 1024.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpfile, err := os.CreateTemp("", "meminfo-*.txt")
			assert.NoError(t, err)
			defer os.Remove(tmpfile.Name())
			_, err = tmpfile.WriteString(tt.meminfoContent)
			assert.NoError(t, err)
			tmpfile.Close()

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

func TestParseMeminfoMissingFile(t *testing.T) {
	_, _, err := parseMeminfo("/non/existent/file")
	assert.Error(t, err)
}

func TestEvaluate(t *testing.T) {
	tests := []struct {
		name          string
		memTotal      float64
		memAvailable  float64
		warn          int
		crit          int
		expectedLevel string
	}{
		{"ok below warning", 8192, 4096, 80, 90, "ok"},       // 50%
		{"warning at threshold", 100, 20, 80, 90, "warning"}, // 80%
		{"warning between", 100, 15, 80, 90, "warning"},      // 85%
		{"critical at threshold", 100, 10, 80, 90, "critical"},
		{"critical above", 100, 2, 80, 90, "critical"}, // 98%
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level, message := evaluate(tt.memTotal, tt.memAvailable, tt.warn, tt.crit)
			assert.Equal(t, tt.expectedLevel, level)
			assert.Contains(t, message, "MemTotal:")
			assert.Contains(t, message, "mem_usage=")
		})
	}
}

func TestReport(t *testing.T) {
	cases := map[string]int{"ok": 0, "warning": 1, "critical": 2, "other": 0}
	for level, expected := range cases {
		var got int
		c := check.New("CheckMemory")
		c.ExitFn = func(code int) { got = code }
		report(c, level, "msg")
		assert.Equal(t, expected, got, "level %q", level)
	}
}
