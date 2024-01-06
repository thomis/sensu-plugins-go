package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiskUsageNoPath(t *testing.T) {
	results, err := diskUsage("")

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
