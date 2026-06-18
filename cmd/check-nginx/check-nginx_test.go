package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckProcessRunningSelf(t *testing.T) {
	f, err := os.CreateTemp("", "pid-*")
	assert.NoError(t, err)
	defer os.Remove(f.Name())
	_, err = f.WriteString(strconv.Itoa(os.Getpid()))
	assert.NoError(t, err)
	f.Close()

	running, err := checkProcessRunning(f.Name())
	assert.True(t, running)
	assert.NoError(t, err)
}

func TestCheckProcessRunningMissingFile(t *testing.T) {
	running, err := checkProcessRunning("/non/existent/nginx.pid")
	assert.False(t, running)
	assert.Error(t, err)
}

func TestCheckProcessRunningBadPid(t *testing.T) {
	f, err := os.CreateTemp("", "pid-*")
	assert.NoError(t, err)
	defer os.Remove(f.Name())
	_, err = f.WriteString("not-a-number")
	assert.NoError(t, err)
	f.Close()

	running, err := checkProcessRunning(f.Name())
	assert.False(t, running)
	assert.Error(t, err)
}

func TestNginxStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Active connections: 43 \nserver accepts handled requests\n7368 7368 10993\nReading: 0 Writing: 5 Waiting: 38\n")
	}))
	defer server.Close()

	connections, err := nginxStatus(server.URL, 5)
	assert.NoError(t, err)
	assert.Equal(t, int64(43), connections)
}

func TestNginxStatusNon200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	_, err := nginxStatus(server.URL, 5)
	assert.Error(t, err)
}
