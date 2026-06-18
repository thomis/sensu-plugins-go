package main

import (
	"net"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckConnectionSuccess(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	assert.NoError(t, err)
	defer listener.Close()

	port := listener.Addr().(*net.TCPAddr).Port
	address, err := checkConnection("127.0.0.1", port, 5)
	assert.NoError(t, err)
	assert.Equal(t, "127.0.0.1:"+strconv.Itoa(port), address)
}

func TestCheckConnectionRefused(t *testing.T) {
	// Bind to a port, then close it so the port is free and connections refused.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	assert.NoError(t, err)
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	_, err = checkConnection("127.0.0.1", port, 2)
	assert.Error(t, err)
}
