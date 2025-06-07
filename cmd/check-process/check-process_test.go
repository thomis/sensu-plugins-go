package main

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCountProcessRegexpCompile(t *testing.T) {
	tests := []struct {
		name        string
		pattern     string
		wantErr     bool
		errContains string
	}{
		{
			name:    "Valid simple pattern",
			pattern: "nginx",
			wantErr: false,
		},
		{
			name:    "Valid regex pattern",
			pattern: "nginx.*worker",
			wantErr: false,
		},
		{
			name:    "Valid pattern with anchors",
			pattern: "^/usr/bin/python3$",
			wantErr: false,
		},
		{
			name:        "Invalid regex pattern",
			pattern:     "[invalid",
			wantErr:     true,
			errContains: "error parsing regexp",
		},
		{
			name:        "Invalid regex with unmatched parenthesis",
			pattern:     "(unclosed",
			wantErr:     true,
			errContains: "error parsing regexp",
		},
		{
			name:    "Empty pattern",
			pattern: "",
			wantErr: false, // Empty pattern is valid regex
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := regexp.Compile(tt.pattern)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestProcessMatching(t *testing.T) {
	// Get current process info for testing
	currentPid := os.Getpid()

	tests := []struct {
		name     string
		pattern  string
		cmdLines []struct {
			pid     int32
			cmdLine string
		}
		expected int
	}{
		{
			name:    "Match single process",
			pattern: "nginx",
			cmdLines: []struct {
				pid     int32
				cmdLine string
			}{
				{pid: 1234, cmdLine: "/usr/sbin/nginx -g daemon off;"},
				{pid: int32(currentPid), cmdLine: "check-process"}, // Current process
			},
			expected: 1, // Should exclude current process
		},
		{
			name:    "Match multiple processes",
			pattern: "python",
			cmdLines: []struct {
				pid     int32
				cmdLine string
			}{
				{pid: 1234, cmdLine: "/usr/bin/python3 script.py"},
				{pid: 5678, cmdLine: "/usr/bin/python manage.py runserver"},
				{pid: 9012, cmdLine: "python3 -m http.server"},
				{pid: int32(currentPid), cmdLine: "check-process"},
			},
			expected: 3,
		},
		{
			name:    "No matches",
			pattern: "non-existent-process",
			cmdLines: []struct {
				pid     int32
				cmdLine string
			}{
				{pid: 1234, cmdLine: "/usr/sbin/nginx"},
				{pid: 5678, cmdLine: "/usr/bin/apache2"},
			},
			expected: 0,
		},
		{
			name:    "Regex pattern matching",
			pattern: "nginx.*worker",
			cmdLines: []struct {
				pid     int32
				cmdLine string
			}{
				{pid: 1234, cmdLine: "nginx: master process"},
				{pid: 5678, cmdLine: "nginx: worker process"},
				{pid: 9012, cmdLine: "nginx: worker process"},
			},
			expected: 2,
		},
		{
			name:    "Case sensitive matching",
			pattern: "NGINX",
			cmdLines: []struct {
				pid     int32
				cmdLine string
			}{
				{pid: 1234, cmdLine: "nginx: master process"},
				{pid: 5678, cmdLine: "NGINX: worker process"},
			},
			expected: 1,
		},
		{
			name:    "Pattern with special characters",
			pattern: `\[kworker/.*\]`,
			cmdLines: []struct {
				pid     int32
				cmdLine string
			}{
				{pid: 1234, cmdLine: "[kworker/0:1]"},
				{pid: 5678, cmdLine: "[kworker/1:2]"},
				{pid: 9012, cmdLine: "kworker/2:0"}, // No brackets
			},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re, err := regexp.Compile(tt.pattern)
			assert.NoError(t, err)

			count := 0
			for _, proc := range tt.cmdLines {
				// Skip current process
				if proc.pid == int32(currentPid) {
					continue
				}
				if re.Match([]byte(proc.cmdLine)) {
					count++
				}
			}

			assert.Equal(t, tt.expected, count)
		})
	}
}

// Test the output formatting logic
func TestProcessOutputFormatting(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		count    int
		critical bool
		expected string
	}{
		{
			name:     "No process found",
			pattern:  "nginx",
			count:    0,
			critical: true,
			expected: "Unable to find process [nginx]",
		},
		{
			name:     "Single process found",
			pattern:  "sshd",
			count:    1,
			critical: false,
			expected: "Process [sshd]: 1 occurence(s)",
		},
		{
			name:     "Multiple processes found",
			pattern:  "python.*",
			count:    5,
			critical: false,
			expected: "Process [python.*]: 5 occurence(s)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output string
			if tt.count == 0 {
				output = fmt.Sprintf("Unable to find process [%s]", tt.pattern)
			} else {
				output = fmt.Sprintf("Process [%s]: %d occurence(s)", tt.pattern, tt.count)
			}
			assert.Equal(t, tt.expected, output)
		})
	}
}

// Test helper to simulate process command line matching
func TestSimulatedCountProcess(t *testing.T) {
	// This simulates the countProcess function logic without actual process access
	simulateCountProcess := func(pattern string, processes []struct {
		pid     int32
		cmdLine string
	}, currentPid int) (int, error) {
		count := 0

		re, err := regexp.Compile(pattern)
		if err != nil {
			return count, err
		}

		for _, process := range processes {
			if int32(currentPid) == process.pid {
				continue
			}
			if re.Match([]byte(process.cmdLine)) {
				fmt.Printf(" - (%d) %s\n", process.pid, process.cmdLine)
				count += 1
			}
		}

		return count, nil
	}

	tests := []struct {
		name      string
		pattern   string
		processes []struct {
			pid     int32
			cmdLine string
		}
		expected int
		wantErr  bool
	}{
		{
			name:    "Count nginx processes",
			pattern: "nginx",
			processes: []struct {
				pid     int32
				cmdLine string
			}{
				{pid: 1000, cmdLine: "/usr/sbin/nginx -g daemon off;"},
				{pid: 1001, cmdLine: "nginx: worker process"},
				{pid: 1002, cmdLine: "nginx: worker process"},
				{pid: 9999, cmdLine: "check-process -p nginx"}, // Simulated current process
			},
			expected: 3,
			wantErr:  false,
		},
		{
			name:    "Invalid regex pattern",
			pattern: "[invalid",
			processes: []struct {
				pid     int32
				cmdLine string
			}{
				{pid: 1000, cmdLine: "some process"},
			},
			expected: 0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, err := simulateCountProcess(tt.pattern, tt.processes, 9999)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, count)
			}
		})
	}
}
