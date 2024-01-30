package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiskUsageNoPath(t *testing.T) {
	results, err := diskUsage("")

	// Exp. macOS 12.x.x => df: -l and -T are mutually exclusive. We might need to exclude from testing
	// Exp. macOS 14.x.x => Works fine
	assert.Nil(t, err)
	assert.GreaterOrEqual(t, len(results), 0)
}

func TestDiskUsage(t *testing.T) {
	results, err := diskUsage("unknown")

	assert.NotNil(t, err)
	assert.Equal(t, "exit status 1", err.Error())
	assert.GreaterOrEqual(t, len(results), 0)
}

func TestContains(t *testing.T) {
	assert.True(t, Contains([]string{"a", "b", "c"}, "b"))
	assert.False(t, Contains([]string{"a", "b", "c"}, "d"))
}

func TestAdjPercent(t *testing.T) {
	value := adjPercent(2048, 80, 100, 60)
	assert.Equal(t, float64(100), value)
}

func TestParseExcludesWithEmptyStrings(t *testing.T) {
	session := session{
		Input: input{
			MountExclude:  "",
			FstypeExclude: ""}}

	(&session).parseExcludes()
	assert.Equal(t, 1, len(session.MountExcludes))
	assert.Equal(t, 1, len(session.FstypeExcludes))
	assert.Equal(t, "", session.MountExcludes[0])
	assert.Equal(t, "", session.FstypeExcludes[0])
}

func TestCaluculateFCritAndFWarnVariant1(t *testing.T) {
	session := session{
		Input: input{
			Minimum: 1,
			Crit:    90,
			Magic:   1,
			Normal:  80,
		},
		FSize: 1,
	}

	(&session).caluculateFCritAndFWarn()
	(&session).caluculateFCritAndFWarn()
	assert.Equal(t, 90.0, session.FCrit)
	assert.Equal(t, 0.0, session.FWarn)
}

func TestCaluculateFCritAndFWarnVariant2(t *testing.T) {
	session := session{
		Input: input{
			Minimum: 1,
			Crit:    90,
			Magic:   1,
			Normal:  80,
		},
		FSize: 1024 * 1024,
	}

	(&session).caluculateFCritAndFWarn()
	assert.Equal(t, 90.0, session.FCrit)
	assert.Equal(t, 0.0, session.FWarn)
}
