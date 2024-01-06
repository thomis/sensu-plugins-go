package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	metrics := New("whatever")

	assert.NotNil(t, metrics)
}

func TestHostname(t *testing.T) {
	metrics := New("whatever")
	metrics = metrics.Hostname("a_hostname")

	assert.Equal(t, metrics.hostname, "a_hostname")
}

func TestScheme(t *testing.T) {
	metrics := New("whatever")
	metrics = metrics.Scheme("a_scheme")

	assert.Equal(t, metrics.scheme, "a_scheme")
}

func TestInit(t *testing.T) {
	metrics := New("whatever")
	metrics.Init()

	assert.NotNil(t, metrics.Option)
}

func TestPrint(t *testing.T) {
	metrics := New("whatever")
	metrics.Print(100)
}
