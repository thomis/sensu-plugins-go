package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatUptime(t *testing.T) {
	tests := []struct {
		name     string
		uptime   uint64
		expected string
	}{
		{"Zero", 0, "0 seconds"},
		{"One second", 1, "1 second"},
		{"Less than a minute", 45, "45 seconds"},
		{"Exactly one minute", 60, "1 minute, 0 seconds"},
		{"Multiple minutes", 125, "2 minutes, 5 seconds"},
		{"Exactly one hour", 3600, "1 hour, 0 seconds"},
		{"Multiple hours", 7265, "2 hours, 1 minute, 5 seconds"},
		{"Exactly one day", 86400, "1 day, 0 seconds"},
		{"One of each", 90061, "1 day, 1 hour, 1 minute, 1 second"},
		{"Hour minute second", 3661, "1 hour, 1 minute, 1 second"},
		{"Multiple days", 186330, "2 days, 3 hours, 45 minutes, 30 seconds"},
		{"23h59m59s", 86399, "23 hours, 59 minutes, 59 seconds"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, formatUptime(tt.uptime))
		})
	}
}
