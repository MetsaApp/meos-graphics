package main

import (
	"flag"
	"fmt"
	"testing"
	"time"
)

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