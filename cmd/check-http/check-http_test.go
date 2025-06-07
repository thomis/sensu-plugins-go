package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStatusCode(t *testing.T) {
	tests := []struct {
		name           string
		serverStatus   int
		expectedStatus int
		expectedError  bool
		serverDelay    time.Duration
		timeout        int
	}{
		{
			name:           "HTTP 200 OK",
			serverStatus:   http.StatusOK,
			expectedStatus: http.StatusOK,
			expectedError:  false,
			timeout:        5,
		},
		{
			name:           "HTTP 301 Redirect",
			serverStatus:   http.StatusMovedPermanently,
			expectedStatus: http.StatusMovedPermanently,
			expectedError:  false,
			timeout:        5,
		},
		{
			name:           "HTTP 404 Not Found",
			serverStatus:   http.StatusNotFound,
			expectedStatus: http.StatusNotFound,
			expectedError:  false,
			timeout:        5,
		},
		{
			name:           "HTTP 500 Internal Server Error",
			serverStatus:   http.StatusInternalServerError,
			expectedStatus: http.StatusInternalServerError,
			expectedError:  false,
			timeout:        5,
		},
		{
			name:           "Request Timeout",
			serverStatus:   http.StatusOK,
			expectedStatus: 0,
			expectedError:  true,
			serverDelay:    2 * time.Second,
			timeout:        1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.serverDelay > 0 {
					time.Sleep(tt.serverDelay)
				}
				w.WriteHeader(tt.serverStatus)
			}))
			defer server.Close()

			input := input{
				Url:      server.URL,
				Timeout:  tt.timeout,
				Insecure: false,
			}

			status, err := statusCode(input)

			if tt.expectedError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.expectedStatus, status)
			}
		})
	}
}

func TestStatusCodeWithAuthentication(t *testing.T) {
	tests := []struct {
		name           string
		username       string
		password       string
		serverUsername string
		serverPassword string
		expectedStatus int
		expectedError  bool
	}{
		{
			name:           "Valid Basic Authentication",
			username:       "testuser",
			password:       "testpass",
			serverUsername: "testuser",
			serverPassword: "testpass",
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:           "Invalid Basic Authentication",
			username:       "wronguser",
			password:       "wrongpass",
			serverUsername: "testuser",
			serverPassword: "testpass",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  false,
		},
		{
			name:           "No Authentication Provided",
			username:       "",
			password:       "",
			serverUsername: "testuser",
			serverPassword: "testpass",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				username, password, ok := r.BasicAuth()
				if !ok || username != tt.serverUsername || password != tt.serverPassword {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			input := input{
				Url:      server.URL,
				Timeout:  5,
				Insecure: false,
				Username: tt.username,
				Password: tt.password,
			}

			status, err := statusCode(input)

			if tt.expectedError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.expectedStatus, status)
			}
		})
	}
}

func TestStatusCodeWithInvalidURL(t *testing.T) {
	tests := []struct {
		name          string
		url           string
		expectedError bool
	}{
		{
			name:          "Invalid URL",
			url:           "not-a-valid-url",
			expectedError: true,
		},
		{
			name:          "Empty URL",
			url:           "",
			expectedError: true,
		},
		{
			name:          "Non-existent host",
			url:           "http://non-existent-host-12345.com",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := input{
				Url:      tt.url,
				Timeout:  2,
				Insecure: false,
			}

			status, err := statusCode(input)

			if tt.expectedError {
				assert.NotNil(t, err)
				assert.Equal(t, 0, status)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestStatusCodeWithHTTPS(t *testing.T) {
	// Test with TLS server
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "HTTPS Test")
	}))
	defer server.Close()

	tests := []struct {
		name          string
		insecure      bool
		expectedError bool
	}{
		{
			name:          "HTTPS with insecure flag",
			insecure:      true,
			expectedError: false,
		},
		{
			name:          "HTTPS without insecure flag (self-signed cert)",
			insecure:      false,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := input{
				Url:      server.URL,
				Timeout:  5,
				Insecure: tt.insecure,
			}

			status, err := statusCode(input)

			if tt.expectedError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, http.StatusOK, status)
			}
		})
	}
}
