package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomis/sensu-plugins-go/pkg/check"
)

func TestCountProcessInvalidPattern(t *testing.T) {
	_, err := countProcess("[invalid(")
	assert.Error(t, err)
}

func TestCountProcessNoMatch(t *testing.T) {
	// A pattern that is extremely unlikely to match any running process. This
	// still exercises the full scan (process list, pid skip, cmdline, match).
	count, err := countProcess("zzz_unlikely_process_name_xyz_123")
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestDescribe(t *testing.T) {
	level, message := describe("nginx", 0)
	assert.Equal(t, "critical", level)
	assert.Contains(t, message, "Unable to find process [nginx]")

	level, message = describe("nginx", 3)
	assert.Equal(t, "ok", level)
	assert.Contains(t, message, "3 occurence(s)")
}

func TestReport(t *testing.T) {
	cases := map[string]int{"ok": 0, "critical": 2, "other": 0}
	for level, expected := range cases {
		var got int
		c := check.New("CheckProcess")
		c.ExitFn = func(code int) { got = code }
		report(c, level, "msg")
		assert.Equal(t, expected, got, "level %q", level)
	}
}
