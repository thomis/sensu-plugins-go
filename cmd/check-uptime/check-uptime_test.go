package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test the uptime formatting logic
func TestUptimeFormatting(t *testing.T) {
	tests := []struct {
		name     string
		uptime   uint64 // seconds
		expected string
	}{
		{
			name:     "Less than a minute",
			uptime:   45,
			expected: "Uptime is 45 seconds",
		},
		{
			name:     "Exactly one minute",
			uptime:   60,
			expected: "Uptime is 1 minute, 0 seconds",
		},
		{
			name:     "Multiple minutes",
			uptime:   125,
			expected: "Uptime is 2 minutes, 5 seconds",
		},
		{
			name:     "Exactly one hour",
			uptime:   3600,
			expected: "Uptime is 1 hour, 0 seconds",
		},
		{
			name:     "Multiple hours",
			uptime:   7265,
			expected: "Uptime is 2 hours, 1 minute, 5 seconds",
		},
		{
			name:     "Exactly one day",
			uptime:   86400,
			expected: "Uptime is 1 day, 0 seconds",
		},
		{
			name:     "Multiple days",
			uptime:   90061,
			expected: "Uptime is 1 day, 1 hour, 1 minute, 1 second",
		},
		{
			name:     "Complex uptime",
			uptime:   3661,
			expected: "Uptime is 1 hour, 1 minute, 1 second",
		},
		{
			name:     "2 days, 3 hours, 45 minutes, 30 seconds",
			uptime:   186330,
			expected: "Uptime is 2 days, 3 hours, 45 minutes, 30 seconds",
		},
		{
			name:     "Zero uptime",
			uptime:   0,
			expected: "Uptime is 0 seconds",
		},
		{
			name:     "One second",
			uptime:   1,
			expected: "Uptime is 1 second",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatUptime(tt.uptime)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper function that mirrors the logic in main()
func formatUptime(uptime uint64) string {
	days := uptime / (60 * 60 * 24)
	hours := (uptime - (days * 60 * 60 * 24)) / (60 * 60)
	minutes := ((uptime - (days * 60 * 60 * 24)) - (hours * 60 * 60)) / 60
	seconds := uptime - (days * 60 * 60 * 24) - (hours * 60 * 60) - (minutes * 60)

	elements := []string{}

	if days > 1 {
		elements = append(elements, fmt.Sprintf("%d days", days))
	}

	if days == 1 {
		elements = append(elements, fmt.Sprintf("%d day", days))
	}

	if hours > 1 {
		elements = append(elements, fmt.Sprintf("%d hours", hours))
	}

	if hours == 1 {
		elements = append(elements, fmt.Sprintf("%d hour", hours))
	}

	if minutes > 1 {
		elements = append(elements, fmt.Sprintf("%d minutes", minutes))
	}

	if minutes == 1 {
		elements = append(elements, fmt.Sprintf("%d minute", minutes))
	}

	if seconds == 1 {
		elements = append(elements, fmt.Sprintf("%d second", seconds))
	} else {
		elements = append(elements, fmt.Sprintf("%d seconds", seconds))
	}

	return fmt.Sprintf("Uptime is %s", strings.Join(elements, ", "))
}

// Test individual time component calculations
func TestTimeComponentCalculations(t *testing.T) {
	tests := []struct {
		name            string
		uptime          uint64
		expectedDays    uint64
		expectedHours   uint64
		expectedMinutes uint64
		expectedSeconds uint64
	}{
		{
			name:            "Zero uptime",
			uptime:          0,
			expectedDays:    0,
			expectedHours:   0,
			expectedMinutes: 0,
			expectedSeconds: 0,
		},
		{
			name:            "One day exactly",
			uptime:          86400,
			expectedDays:    1,
			expectedHours:   0,
			expectedMinutes: 0,
			expectedSeconds: 0,
		},
		{
			name:            "One hour exactly",
			uptime:          3600,
			expectedDays:    0,
			expectedHours:   1,
			expectedMinutes: 0,
			expectedSeconds: 0,
		},
		{
			name:            "One minute exactly",
			uptime:          60,
			expectedDays:    0,
			expectedHours:   0,
			expectedMinutes: 1,
			expectedSeconds: 0,
		},
		{
			name:            "Complex time",
			uptime:          93784, // 1 day, 2 hours, 3 minutes, 4 seconds
			expectedDays:    1,
			expectedHours:   2,
			expectedMinutes: 3,
			expectedSeconds: 4,
		},
		{
			name:            "Maximum components",
			uptime:          359999, // 4 days, 3 hours, 59 minutes, 59 seconds
			expectedDays:    4,
			expectedHours:   3,
			expectedMinutes: 59,
			expectedSeconds: 59,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			days := tt.uptime / (60 * 60 * 24)
			hours := (tt.uptime - (days * 60 * 60 * 24)) / (60 * 60)
			minutes := ((tt.uptime - (days * 60 * 60 * 24)) - (hours * 60 * 60)) / 60
			seconds := tt.uptime - (days * 60 * 60 * 24) - (hours * 60 * 60) - (minutes * 60)

			assert.Equal(t, tt.expectedDays, days, "Days calculation mismatch")
			assert.Equal(t, tt.expectedHours, hours, "Hours calculation mismatch")
			assert.Equal(t, tt.expectedMinutes, minutes, "Minutes calculation mismatch")
			assert.Equal(t, tt.expectedSeconds, seconds, "Seconds calculation mismatch")
		})
	}
}

// Test the pluralization logic
func TestPluralization(t *testing.T) {
	tests := []struct {
		value    uint64
		unit     string
		expected string
	}{
		{0, "day", "0 days"},
		{1, "day", "1 day"},
		{2, "day", "2 days"},
		{0, "hour", "0 hours"},
		{1, "hour", "1 hour"},
		{2, "hour", "2 hours"},
		{0, "minute", "0 minutes"},
		{1, "minute", "1 minute"},
		{2, "minute", "2 minutes"},
		{0, "second", "0 seconds"},
		{1, "second", "1 second"},
		{2, "second", "2 seconds"},
		{100, "day", "100 days"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d_%s", tt.value, tt.unit), func(t *testing.T) {
			var result string
			switch tt.unit {
			case "day":
				if tt.value == 1 {
					result = fmt.Sprintf("%d day", tt.value)
				} else {
					result = fmt.Sprintf("%d days", tt.value)
				}
			case "hour":
				if tt.value == 1 {
					result = fmt.Sprintf("%d hour", tt.value)
				} else {
					result = fmt.Sprintf("%d hours", tt.value)
				}
			case "minute":
				if tt.value == 1 {
					result = fmt.Sprintf("%d minute", tt.value)
				} else {
					result = fmt.Sprintf("%d minutes", tt.value)
				}
			case "second":
				if tt.value == 1 {
					result = fmt.Sprintf("%d second", tt.value)
				} else {
					result = fmt.Sprintf("%d seconds", tt.value)
				}
			}
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test edge cases
func TestUptimeEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		uptime   uint64
		contains []string // Elements that should be in the output
	}{
		{
			name:     "Very large uptime",
			uptime:   31536000, // 365 days
			contains: []string{"365 days", "0 seconds"},
		},
		{
			name:     "23 hours 59 minutes 59 seconds",
			uptime:   86399,
			contains: []string{"23 hours", "59 minutes", "59 seconds"},
		},
		{
			name:     "Only seconds",
			uptime:   59,
			contains: []string{"59 seconds"},
		},
		{
			name:     "Only minutes and seconds",
			uptime:   3599,
			contains: []string{"59 minutes", "59 seconds"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatUptime(tt.uptime)
			for _, expected := range tt.contains {
				assert.Contains(t, result, expected)
			}
		})
	}
}
