package main

import (
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/thomis/sensu-plugins-go/pkg/check"
)

func TestParseThresholds(t *testing.T) {
	critMin, critMax, warnMin, warnMax, err := parseThresholds("10:50", "5:30")
	assert.NoError(t, err)
	assert.Equal(t, int64(10), critMin)
	assert.Equal(t, int64(50), critMax)
	assert.Equal(t, int64(5), warnMin)
	assert.Equal(t, int64(30), warnMax)

	// max is optional
	critMin, critMax, warnMin, warnMax, err = parseThresholds("10", "5")
	assert.NoError(t, err)
	assert.Equal(t, int64(10), critMin)
	assert.Equal(t, int64(0), critMax)
	assert.Equal(t, int64(5), warnMin)
	assert.Equal(t, int64(0), warnMax)
}

func TestParseThresholdsErrors(t *testing.T) {
	cases := []struct{ crit, warn string }{
		{"abc", "5"},    // invalid crit min
		{"10:abc", "5"}, // invalid crit max
		{"5", "abc"},    // invalid warn min
		{"5", "10:abc"}, // invalid warn max
		{"50:10", "5"},  // crit min > max
		{"5", "50:10"},  // warn min > max
	}
	for _, c := range cases {
		_, _, _, _, err := parseThresholds(c.crit, c.warn)
		assert.Error(t, err, "crit=%s warn=%s", c.crit, c.warn)
	}
}

func TestExecProcessCount(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	mock.ExpectQuery("PROCESSLIST").WillReturnRows(
		sqlmock.NewRows([]string{"count"}).AddRow(42))

	count, err := execProcessCount(db)
	assert.NoError(t, err)
	assert.Equal(t, int64(42), count)
}

func TestExecProcessCountError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	mock.ExpectQuery("PROCESSLIST").WillReturnError(fmt.Errorf("access denied"))

	_, err = execProcessCount(db)
	assert.Error(t, err)
}

func newTestSession() (*session, *int) {
	code := new(int)
	s := &session{Check: check.New("test")}
	s.Check.ExitFn = func(c int) { *code = c }
	return s, code
}

func TestReport(t *testing.T) {
	tests := []struct {
		name         string
		critMin      int64
		critMax      int64
		warnMin      int64
		warnMax      int64
		processCount int64
		expected     int
	}{
		{"critical exceed max", 0, 50, 0, 30, 100, 2},
		{"critical below min", 10, 0, 0, 0, 5, 2},
		{"warning exceed max", 0, 100, 0, 30, 40, 1},
		{"warning below min", 0, 0, 10, 0, 8, 1},
		{"ok", 0, 100, 0, 50, 20, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, code := newTestSession()
			s.CritMin, s.CritMax = tt.critMin, tt.critMax
			s.WarnMin, s.WarnMax = tt.warnMin, tt.warnMax
			s.ProcessCount = tt.processCount
			s.report()
			assert.Equal(t, tt.expected, *code)
		})
	}
}
