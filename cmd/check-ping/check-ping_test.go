package main

import (
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Since the original code only has main(), we'll test the core functionality
// by testing net.DialTimeout behavior with different scenarios

// Helper function to find a port that is likely closed
func findClosedPort() int {
	// Try to bind to a random port, then close it immediately
	// This gives us a port that was recently available but is now closed
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		// Fallback to a high port number unlikely to be in use
		return 54321
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()
	// Give the OS a moment to fully release the port
	time.Sleep(10 * time.Millisecond)
	return port
}

func TestDialTimeout(t *testing.T) {
	// Create a test server to ensure we have a valid endpoint
	testServer, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	defer testServer.Close()

	// Get the port that was assigned
	testPort := testServer.Addr().(*net.TCPAddr).Port

	// Accept connections in the background
	go func() {
		for {
			conn, err := testServer.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()

	tests := []struct {
		name     string
		host     string
		port     int
		timeout  time.Duration
		wantErr  bool
		errCheck func(error) bool
	}{
		{
			name:    "Valid connection to test server",
			host:    "127.0.0.1",
			port:    testPort,
			timeout: 5 * time.Second,
			wantErr: false,
		},
		{
			name:    "Invalid port",
			host:    "localhost",
			port:    99999,
			timeout: 1 * time.Second,
			wantErr: true,
			errCheck: func(err error) bool {
				// Port out of range
				return err != nil
			},
		},
		{
			name:    "Connection refused on closed port",
			host:    "127.0.0.1",
			port:    findClosedPort(),
			timeout: 1 * time.Second,
			wantErr: true,
			errCheck: func(err error) bool {
				// Should get connection refused
				return err != nil
			},
		},
		{
			name:    "Non-existent host",
			host:    "non-existent-host-12345.invalid",
			port:    80,
			timeout: 1 * time.Second,
			wantErr: true,
			errCheck: func(err error) bool {
				// Should get a lookup error
				_, ok := err.(*net.OpError)
				return ok
			},
		},
		{
			name:    "Timeout test",
			host:    "192.0.2.1", // TEST-NET-1 (RFC 5737) - won't route
			port:    80,
			timeout: 100 * time.Millisecond,
			wantErr: true,
			errCheck: func(err error) bool {
				// Should timeout
				if netErr, ok := err.(net.Error); ok {
					return netErr.Timeout()
				}
				return false
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Using the same approach as in main()
			address := tt.host + ":" + strconv.Itoa(tt.port)

			conn, err := net.DialTimeout("tcp", address, tt.timeout)

			if (err != nil) != tt.wantErr {
				t.Errorf("DialTimeout() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errCheck != nil {
				assert.True(t, tt.errCheck(err), "Error check failed: %v", err)
			}

			if conn != nil {
				conn.Close()
			}
		})
	}
}

// Test to verify address formation
func TestAddressFormation(t *testing.T) {
	tests := []struct {
		host     string
		port     int
		expected string
	}{
		{"localhost", 22, "localhost:22"},
		{"127.0.0.1", 80, "127.0.0.1:80"},
		{"example.com", 443, "example.com:443"},
		{"192.168.1.1", 8080, "192.168.1.1:8080"},
		{"[::1]", 22, "[::1]:22"}, // IPv6
	}

	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			// Using the same approach as in main
			address := tt.host + ":" + strconv.Itoa(tt.port)
			assert.Equal(t, tt.expected, address)
		})
	}
}
