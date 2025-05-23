package main

import (
	"flag"
	"fmt"
	"testing"
	"time"
)

func TestMeOSConfigFlags(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		expectedHost string
		expectedPort string
	}{
		{
			name:         "default values",
			args:         []string{},
			expectedHost: "localhost",
			expectedPort: "2009",
		},
		{
			name:         "custom host",
			args:         []string{"-meos-host", "192.168.1.100"},
			expectedHost: "192.168.1.100",
			expectedPort: "2009",
		},
		{
			name:         "custom port",
			args:         []string{"-meos-port", "8080"},
			expectedHost: "localhost",
			expectedPort: "8080",
		},
		{
			name:         "custom host and port",
			args:         []string{"-meos-host", "meos.example.com", "-meos-port", "3000"},
			expectedHost: "meos.example.com",
			expectedPort: "3000",
		},
		{
			name:         "port none",
			args:         []string{"-meos-port", "none"},
			expectedHost: "localhost",
			expectedPort: "none",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flags for each test
			flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)

			// Define flags as in main
			meosHost := flag.String("meos-host", "localhost", "MeOS server hostname or IP address")
			meosPort := flag.String("meos-port", "2009", "MeOS server port")

			// Parse test args
			err := flag.CommandLine.Parse(tt.args)
			if err != nil {
				t.Fatalf("Failed to parse flags: %v", err)
			}

			if *meosHost != tt.expectedHost {
				t.Errorf("meos-host = %v, want %v", *meosHost, tt.expectedHost)
			}
			if *meosPort != tt.expectedPort {
				t.Errorf("meos-port = %v, want %v", *meosPort, tt.expectedPort)
			}
		})
	}
}

func TestPollIntervalFlag(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected time.Duration
	}{
		{
			name:     "default value",
			args:     []string{},
			expected: 1 * time.Second,
		},
		{
			name:     "milliseconds format",
			args:     []string{"-poll-interval", "200ms"},
			expected: 200 * time.Millisecond,
		},
		{
			name:     "seconds format",
			args:     []string{"-poll-interval", "9s"},
			expected: 9 * time.Second,
		},
		{
			name:     "minutes format",
			args:     []string{"-poll-interval", "2m"},
			expected: 2 * time.Minute,
		},
		{
			name:     "combined format",
			args:     []string{"-poll-interval", "1m30s"},
			expected: 90 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flags for each test
			flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)

			// Define flags as in main
			pollInterval := flag.Duration("poll-interval", 1*time.Second, "Poll interval for MeOS data updates")

			// Parse test args
			err := flag.CommandLine.Parse(tt.args)
			if err != nil {
				t.Fatalf("Failed to parse flags: %v", err)
			}

			if *pollInterval != tt.expected {
				t.Errorf("poll-interval = %v, want %v", *pollInterval, tt.expected)
			}
		})
	}
}

func TestPollIntervalValidation(t *testing.T) {
	tests := []struct {
		name      string
		interval  time.Duration
		wantError bool
	}{
		{
			name:      "valid minimum",
			interval:  100 * time.Millisecond,
			wantError: false,
		},
		{
			name:      "valid maximum",
			interval:  1 * time.Hour,
			wantError: false,
		},
		{
			name:      "too small",
			interval:  50 * time.Millisecond,
			wantError: true,
		},
		{
			name:      "too large",
			interval:  2 * time.Hour,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test validation logic
			err := validatePollInterval(tt.interval)
			if (err != nil) != tt.wantError {
				t.Errorf("validatePollInterval(%v) error = %v, wantError %v", tt.interval, err, tt.wantError)
			}
		})
	}
}

// Helper function to match validation logic in main
func validatePollInterval(interval time.Duration) error {
	if interval < 100*time.Millisecond {
		return fmt.Errorf("poll interval too small (minimum 100ms): %s", interval)
	}
	if interval > 1*time.Hour {
		return fmt.Errorf("poll interval too large (maximum 1 hour): %s", interval)
	}
	return nil
}
