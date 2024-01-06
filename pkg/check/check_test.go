package check

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	check := New("something")
	check.Init()

	assert.NotNil(t, check.Option)
}

func TestOk(t *testing.T) {
	check := New("something")

	go func() {
		defer func() {
			if r := recover(); r != nil {
				assert.Equal(t, fmt.Sprintf("%v", r), "unexpected call to os.Exit(0) during test")
			}
		}()
		check.Ok("whatever")
	}()
}
