package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSendWithStatusCodes(t *testing.T) {
	tests := []struct {
		name             string
		serverStatus     int
		expectedCode     int
		expectedStatus   string
		expectedContains string
	}{
		{
			name:             "Expected 200 OK",
			serverStatus:     http.StatusOK,
			expectedCode:     http.StatusOK,
			expectedStatus:   "OK",
			expectedContains: "Status code [200]",
		},
		{
			name:             "Expected 201 Created",
			serverStatus:     http.StatusCreated,
			expectedCode:     http.StatusCreated,
			expectedStatus:   "OK",
			expectedContains: "Status code [201]",
		},
		{
			name:             "Unexpected 404 when expecting 200",
			serverStatus:     http.StatusNotFound,
			expectedCode:     http.StatusOK,
			expectedStatus:   "CRITICAL",
			expectedContains: "Status code [404]",
		},
		{
			name:             "Unexpected 500 when expecting 200",
			serverStatus:     http.StatusInternalServerError,
			expectedCode:     http.StatusOK,
			expectedStatus:   "CRITICAL",
			expectedContains: "Status code [500]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)
				json.NewEncoder(w).Encode(map[string]string{"status": "test"})
			}))
			defer server.Close()

			req := &request{
				url:     server.URL,
				timeout: 5 * time.Second,
				method:  "GET",
				code:    tt.expectedCode,
			}

			status, response, err := send(req)

			assert.Nil(t, err)
			assert.Equal(t, tt.expectedStatus, status)
			assert.Contains(t, response, tt.expectedContains)
		})
	}
}

func TestSendWithHTTPMethods(t *testing.T) {
	tests := []struct {
		name         string
		method       string
		body         string
		expectedBody string
	}{
		{
			name:         "GET request",
			method:       "GET",
			body:         "",
			expectedBody: "",
		},
		{
			name:         "POST request with JSON body",
			method:       "POST",
			body:         `{"key": "value"}`,
			expectedBody: `{"key": "value"}`,
		},
		{
			name:         "PUT request with JSON body",
			method:       "PUT",
			body:         `{"id": 1, "name": "test"}`,
			expectedBody: `{"id": 1, "name": "test"}`,
		},
		{
			name:         "DELETE request",
			method:       "DELETE",
			body:         "",
			expectedBody: "",
		},
		{
			name:         "PATCH request with JSON body",
			method:       "PATCH",
			body:         `{"update": "field"}`,
			expectedBody: `{"update": "field"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.method, r.Method)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				if tt.body != "" {
					body, _ := io.ReadAll(r.Body)
					assert.Equal(t, tt.expectedBody, string(body))
				}

				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]string{"method": r.Method})
			}))
			defer server.Close()

			req := &request{
				url:     server.URL,
				timeout: 5 * time.Second,
				method:  tt.method,
				body:    tt.body,
				code:    http.StatusOK,
			}

			status, response, err := send(req)

			assert.Nil(t, err)
			assert.Equal(t, "OK", status)
			assert.Contains(t, response, "Status code [200]")
		})
	}
}

func TestSendWithPatternMatching(t *testing.T) {
	tests := []struct {
		name           string
		responseBody   map[string]interface{}
		pattern        string
		expectedStatus string
		expectedError  bool
	}{
		{
			name:           "Pattern matches JSON response",
			responseBody:   map[string]interface{}{"status": "healthy", "version": "1.2.3"},
			pattern:        `"status":\s*"healthy"`,
			expectedStatus: "OK",
			expectedError:  false,
		},
		{
			name:           "Pattern does not match JSON response",
			responseBody:   map[string]interface{}{"status": "unhealthy", "version": "1.2.3"},
			pattern:        `"status":\s*"healthy"`,
			expectedStatus: "CRITICAL",
			expectedError:  false,
		},
		{
			name:           "Complex pattern matching",
			responseBody:   map[string]interface{}{"users": []int{1, 2, 3}, "total": 3},
			pattern:        `"total":\s*3`,
			expectedStatus: "OK",
			expectedError:  false,
		},
		{
			name:           "Invalid regex pattern",
			responseBody:   map[string]interface{}{"status": "ok"},
			pattern:        `[`,
			expectedStatus: "",
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(tt.responseBody)
			}))
			defer server.Close()

			req := &request{
				url:     server.URL,
				timeout: 5 * time.Second,
				method:  "GET",
				code:    http.StatusOK,
				pattern: tt.pattern,
			}

			status, response, err := send(req)

			if tt.expectedError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.expectedStatus, status)
				if tt.expectedStatus == "CRITICAL" {
					assert.Contains(t, response, "pattern")
					assert.Contains(t, response, "doesn't match")
				}
			}
		})
	}
}

func TestSendWithAuthentication(t *testing.T) {
	tests := []struct {
		name           string
		username       string
		password       string
		serverUsername string
		serverPassword string
		expectedStatus string
	}{
		{
			name:           "Valid Basic Authentication",
			username:       "testuser",
			password:       "testpass",
			serverUsername: "testuser",
			serverPassword: "testpass",
			expectedStatus: "OK",
		},
		{
			name:           "Invalid Basic Authentication",
			username:       "wronguser",
			password:       "wrongpass",
			serverUsername: "testuser",
			serverPassword: "testpass",
			expectedStatus: "CRITICAL",
		},
		{
			name:           "No Authentication Provided",
			username:       "",
			password:       "",
			serverUsername: "testuser",
			serverPassword: "testpass",
			expectedStatus: "CRITICAL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				username, password, ok := r.BasicAuth()
				if !ok || username != tt.serverUsername || password != tt.serverPassword {
					w.WriteHeader(http.StatusUnauthorized)
					json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
					return
				}
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]string{"status": "authorized"})
			}))
			defer server.Close()

			req := &request{
				url:      server.URL,
				timeout:  5 * time.Second,
				method:   "GET",
				code:     http.StatusOK,
				username: tt.username,
				password: tt.password,
			}

			status, response, err := send(req)

			assert.Nil(t, err)
			assert.Equal(t, tt.expectedStatus, status)
			if tt.expectedStatus == "CRITICAL" {
				assert.Contains(t, response, "Status code [401]")
			}
		})
	}
}

func TestSendWithTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	req := &request{
		url:     server.URL,
		timeout: 1 * time.Second,
		method:  "GET",
		code:    http.StatusOK,
	}

	status, _, err := send(req)

	assert.NotNil(t, err)
	assert.Equal(t, "CRITICAL", status)
	assert.Contains(t, err.Error(), "deadline exceeded")
}

func TestSendWithInvalidURL(t *testing.T) {
	tests := []struct {
		name          string
		url           string
		expectedError string
	}{
		{
			name:          "Invalid URL scheme",
			url:           "not-a-valid-url",
			expectedError: "unsupported protocol scheme",
		},
		{
			name:          "Empty URL",
			url:           "",
			expectedError: "unsupported protocol scheme",
		},
		{
			name:          "Non-existent host",
			url:           "http://non-existent-host-12345.com",
			expectedError: "no such host",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &request{
				url:     tt.url,
				timeout: 2 * time.Second,
				method:  "GET",
				code:    http.StatusOK,
			}

			status, _, err := send(req)

			assert.NotNil(t, err)
			assert.Equal(t, "CRITICAL", status)
			assert.Contains(t, strings.ToLower(err.Error()), strings.ToLower(tt.expectedError))
		})
	}
}

func TestSendWithHTTPS(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"secure": "true"})
	}))
	defer server.Close()

	tests := []struct {
		name           string
		insecure       bool
		expectedStatus string
		expectedError  bool
	}{
		{
			name:           "HTTPS with insecure flag",
			insecure:       true,
			expectedStatus: "OK",
			expectedError:  false,
		},
		{
			name:           "HTTPS without insecure flag (self-signed cert)",
			insecure:       false,
			expectedStatus: "CRITICAL",
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &request{
				url:      server.URL,
				timeout:  5 * time.Second,
				method:   "GET",
				code:     http.StatusOK,
				insecure: tt.insecure,
			}

			status, response, err := send(req)

			if tt.expectedError {
				assert.NotNil(t, err)
				assert.Equal(t, "CRITICAL", status)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.expectedStatus, status)
				assert.Contains(t, response, "Status code [200]")
			}
		})
	}
}

func TestSendWithProxy(t *testing.T) {
	// Create a proxy server
	proxyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simple proxy that just returns success
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"via": "proxy"})
	}))
	defer proxyServer.Close()

	// Create target server
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"target": "reached"})
	}))
	defer targetServer.Close()

	tests := []struct {
		name           string
		proxyURL       string
		noProxy        bool
		expectedStatus string
		expectedError  bool
	}{
		{
			name:           "With valid proxy URL",
			proxyURL:       proxyServer.URL,
			noProxy:        false,
			expectedStatus: "OK",
			expectedError:  false,
		},
		{
			name:           "With noProxy flag set",
			proxyURL:       "",
			noProxy:        true,
			expectedStatus: "OK",
			expectedError:  false,
		},
		{
			name:           "With invalid proxy URL",
			proxyURL:       "not-a-valid-url",
			noProxy:        false,
			expectedStatus: "",
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &request{
				url:      targetServer.URL,
				timeout:  5 * time.Second,
				method:   "GET",
				code:     http.StatusOK,
				proxyURL: tt.proxyURL,
				noProxy:  tt.noProxy,
			}

			status, response, err := send(req)

			if tt.expectedError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.expectedStatus, status)
				assert.Contains(t, response, "Status code [200]")
			}
		})
	}
}

func TestSendResponseTime(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add a small delay to ensure measurable response time
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "{}")
	}))
	defer server.Close()

	req := &request{
		url:     server.URL,
		timeout: 5 * time.Second,
		method:  "GET",
		code:    http.StatusOK,
	}

	status, response, err := send(req)

	assert.Nil(t, err)
	assert.Equal(t, "OK", status)
	assert.Contains(t, response, "took [")
	assert.Contains(t, response, " ms]")
}
