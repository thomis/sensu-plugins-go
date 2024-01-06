package common

import (
	"github.com/stretchr/testify/assert"
	"runtime"
	"testing"
)

func TestGetStats(t *testing.T) {
	stats, err := GetStats()

	if runtime.GOOS == "linux" {
		assert.Nil(t, err)
	}

	if runtime.GOOS == "darwin" {
		assert.NotNil(t, err)
		assert.Equal(t, stats, []float64{})
	}

}
