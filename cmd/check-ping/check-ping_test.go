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

func TestDialTimeout(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		port     int
		timeout  time.Duration
		wantErr  bool
		errCheck func(error) bool
	}{
		{
			name:    "Valid connection to localhost",
			host:    "localhost",
			port:    8021, // Port we saw listening
			timeout: 5 * time.Second,
			wantErr: false, // This may fail if service is not running
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
