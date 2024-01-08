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
	var value int
	fakeExit := func(code int) {
		value = code
	}

	check := New("something")
	check.ExitFn = fakeExit

	check.Ok("whatever")
	assert.Equal(t, 0, value)
}

func TestWarning(t *testing.T) {
	var value int
	fakeExit := func(code int) {
		value = code
	}

	check := New("something")
	check.ExitFn = fakeExit

	check.Warning("whatever")
	assert.Equal(t, 1, value)
}

func TestCritical(t *testing.T) {
	var value int
	fakeExit := func(code int) {
		value = code
	}

	check := New("something")
	check.ExitFn = fakeExit

	check.Critical("whatever")
	assert.Equal(t, 2, value)
}

func TestEror(t *testing.T) {
	var value int
	fakeExit := func(code int) {
		value = code
	}

	check := New("something")
	check.ExitFn = fakeExit

	check.Error(fmt.Errorf("whatever"))
	assert.Equal(t, 3, value)
}
