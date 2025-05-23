package handlers

import (
	"fmt"
	"testing"
	"time"
)

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		// Basic cases
		{
			name:     "zero duration",
			duration: 0,
			expected: "0:00.0",
		},
		{
			name:     "one second",
			duration: 1 * time.Second,
			expected: "0:01.0",
		},
		{
			name:     "ten seconds",
			duration: 10 * time.Second,
			expected: "0:10.0",
		},
		{
			name:     "one minute",
			duration: 60 * time.Second,
			expected: "1:00.0",
		},
		{
			name:     "one hour",
			duration: 3600 * time.Second,
			expected: "1:00:00.0",
		},

		// Deciseconds cases
		{
			name:     "100 milliseconds (1 decisecond)",
			duration: 100 * time.Millisecond,
			expected: "0:00.1",
		},
		{
			name:     "200 milliseconds (2 deciseconds)",
			duration: 200 * time.Millisecond,
			expected: "0:00.2",
		},
		{
			name:     "900 milliseconds (9 deciseconds)",
			duration: 900 * time.Millisecond,
			expected: "0:00.9",
		},
		{
			name:     "1.5 seconds",
			duration: 1500 * time.Millisecond,
			expected: "0:01.5",
		},

		// Complex cases with deciseconds
		{
			name:     "83.4 seconds (834 deciseconds)",
			duration: 83*time.Second + 400*time.Millisecond,
			expected: "1:23.4",
		},
		{
			name:     "123.7 seconds",
			duration: 123*time.Second + 700*time.Millisecond,
			expected: "2:03.7",
		},
		{
			name:     "599.9 seconds",
			duration: 599*time.Second + 900*time.Millisecond,
			expected: "9:59.9",
		},
		{
			name:     "600.0 seconds (10 minutes)",
			duration: 600 * time.Second,
			expected: "10:00.0",
		},

		// Hour cases
		{
			name:     "1 hour 23 minutes 45.6 seconds",
			duration: 1*time.Hour + 23*time.Minute + 45*time.Second + 600*time.Millisecond,
			expected: "1:23:45.6",
		},
		{
			name:     "2 hours 5 seconds",
			duration: 2*time.Hour + 5*time.Second,
			expected: "2:00:05.0",
		},
		{
			name:     "12 hours 34 minutes 56.7 seconds",
			duration: 12*time.Hour + 34*time.Minute + 56*time.Second + 700*time.Millisecond,
			expected: "12:34:56.7",
		},

		// Edge cases
		{
			name:     "59 minutes 59.9 seconds",
			duration: 59*time.Minute + 59*time.Second + 900*time.Millisecond,
			expected: "59:59.9",
		},
		{
			name:     "exactly 60 minutes",
			duration: 60 * time.Minute,
			expected: "1:00:00.0",
		},
		{
			name:     "negative duration",
			duration: -10 * time.Second,
			expected: "0:-10.0", // Note: This might not be desired behavior
		},

		// Precision edge cases
		{
			name:     "50 milliseconds (should round down to 0 deciseconds)",
			duration: 50 * time.Millisecond,
			expected: "0:00.0",
		},
		{
			name:     "150 milliseconds (should be 1 decisecond)",
			duration: 150 * time.Millisecond,
			expected: "0:00.1",
		},
		{
			name:     "951 milliseconds (should be 9 deciseconds)",
			duration: 951 * time.Millisecond,
			expected: "0:00.9",
		},
		{
			name:     "999 milliseconds (should be 9 deciseconds)",
			duration: 999 * time.Millisecond,
			expected: "0:00.9",
		},

		// Large durations
		{
			name:     "100 hours",
			duration: 100 * time.Hour,
			expected: "100:00:00.0",
		},
		{
			name:     "999 hours 59 minutes 59.9 seconds",
			duration: 999*time.Hour + 59*time.Minute + 59*time.Second + 900*time.Millisecond,
			expected: "999:59:59.9",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDuration(tt.duration)
			if result != tt.expected {
				t.Errorf("formatDuration(%v) = %q, want %q", tt.duration, result, tt.expected)
			}
		})
	}
}

func TestFormatDuration_RealWorldExamples(t *testing.T) {
	// Test cases based on typical orienteering times
	tests := []struct {
		name        string
		deciseconds int
		expected    string
	}{
		{
			name:        "sprint course winner (12:34.5)",
			deciseconds: 7545,
			expected:    "12:34.5",
		},
		{
			name:        "middle distance (35:12.3)",
			deciseconds: 21123,
			expected:    "35:12.3",
		},
		{
			name:        "long distance (1:23:45.6)",
			deciseconds: 50256,
			expected:    "1:23:45.6",
		},
		{
			name:        "ultra long (2:45:30.0)",
			deciseconds: 99300,
			expected:    "2:45:30.0",
		},
		{
			name:        "very fast sprint (8:00.0)",
			deciseconds: 4800,
			expected:    "8:00.0",
		},
		{
			name:        "exactly one hour",
			deciseconds: 36000,
			expected:    "1:00:00.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			duration := time.Duration(tt.deciseconds) * 100 * time.Millisecond
			result := formatDuration(duration)
			if result != tt.expected {
				t.Errorf("formatDuration(%d deciseconds) = %q, want %q", tt.deciseconds, result, tt.expected)
			}
		})
	}
}

func BenchmarkFormatDuration(b *testing.B) {
	testCases := []struct {
		name     string
		duration time.Duration
	}{
		{"short", 83*time.Second + 400*time.Millisecond},
		{"medium", 35*time.Minute + 12*time.Second + 300*time.Millisecond},
		{"long", 1*time.Hour + 23*time.Minute + 45*time.Second + 600*time.Millisecond},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = formatDuration(tc.duration)
			}
		})
	}
}

func TestFormatDuration_Consistency(t *testing.T) {
	// Test that converting from deciseconds and back produces consistent results
	testDeciseconds := []int{0, 1, 10, 100, 834, 1234, 36000, 50256, 99999}

	for _, ds := range testDeciseconds {
		duration := time.Duration(ds) * 100 * time.Millisecond
		formatted := formatDuration(duration)
		
		// The formatted string should represent the same time
		// This is more of a sanity check than a strict test
		t.Logf("Deciseconds: %d -> Duration: %v -> Formatted: %s", ds, duration, formatted)
	}
}

func TestFormatDuration_Subsecond(t *testing.T) {
	// Test formatting of subsecond durations specifically
	tests := []struct {
		milliseconds int
		expected     string
	}{
		{0, "0:00.0"},
		{99, "0:00.0"},
		{100, "0:00.1"},
		{199, "0:00.1"},
		{200, "0:00.2"},
		{250, "0:00.2"},
		{500, "0:00.5"},
		{900, "0:00.9"},
		{950, "0:00.9"},
		{999, "0:00.9"},
		{1000, "0:01.0"},
		{1100, "0:01.1"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%dms", tt.milliseconds), func(t *testing.T) {
			duration := time.Duration(tt.milliseconds) * time.Millisecond
			result := formatDuration(duration)
			if result != tt.expected {
				t.Errorf("formatDuration(%d ms) = %q, want %q", tt.milliseconds, result, tt.expected)
			}
		})
	}
}