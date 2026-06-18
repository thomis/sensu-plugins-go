package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomis/sensu-plugins-go/pkg/check"
)

func TestEvaluate(t *testing.T) {
	// usageStats layout: user, nice, system, idle(3), iowait, ...
	tests := []struct {
		name          string
		usage         []float64
		warn          int
		crit          int
		expectedLevel string
	}{
		{"ok high idle", []float64{10, 0, 5, 80, 5}, 80, 90, "ok"},
		{"warning at threshold", []float64{40, 0, 20, 20, 20}, 80, 90, "warning"}, // idle 20 == 100-80
		{"critical at threshold", []float64{45, 0, 25, 10, 20}, 80, 90, "critical"},
		{"critical low idle", []float64{60, 0, 30, 5, 5}, 80, 90, "critical"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level, message := evaluate(tt.usage, tt.warn, tt.crit)
			assert.Equal(t, tt.expectedLevel, level)
			assert.Contains(t, message, "idle=")
			assert.Contains(t, message, "cpu_user=")
		})
	}
}

func TestReport(t *testing.T) {
	cases := map[string]int{"ok": 0, "warning": 1, "critical": 2, "other": 0}
	for level, expected := range cases {
		var got int
		c := check.New("CheckCPU")
		c.ExitFn = func(code int) { got = code }
		report(c, level, "msg")
		assert.Equal(t, expected, got, "level %q", level)
	}
}
