package meos

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"meos-graphics/internal/logger"
	"meos-graphics/internal/state"
)

func TestAdapter_PollInterval(t *testing.T) {
	// Initialize logger for tests
	if err := logger.Init(); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	tests := []struct {
		name               string
		pollInterval       time.Duration
		testDuration       time.Duration
		expectedMinCalls   int
		expectedMaxCalls   int
	}{
		{
			name:               "100ms interval",
			pollInterval:       100 * time.Millisecond,
			testDuration:       550 * time.Millisecond,
			expectedMinCalls:   4, // Initial call + 4 polls
			expectedMaxCalls:   6, // Allow some timing variance
		},
		{
			name:               "500ms interval",
			pollInterval:       500 * time.Millisecond,
			testDuration:       1100 * time.Millisecond,
			expectedMinCalls:   2, // Initial call + 2 polls
			expectedMaxCalls:   3, // Allow some timing variance
		},
		{
			name:               "1s interval",
			pollInterval:       1 * time.Second,
			testDuration:       2100 * time.Millisecond,
			expectedMinCalls:   2, // Initial call + 2 polls
			expectedMaxCalls:   3, // Allow some timing variance
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var callCount int32

			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				atomic.AddInt32(&callCount, 1)
				w.Header().Set("Content-Type", "application/xml")
				w.WriteHeader(http.StatusOK)
				count := atomic.LoadInt32(&callCount)
				// Return different nextdifference each time to trigger updates
				response := `<?xml version="1.0" encoding="UTF-8"?>
<MOPComplete nextdifference="` + string(rune('a'+count)) + `">
    <competition date="2024-01-01" organizer="Test Organizer" zerotime="10:00:00">Test Competition</competition>
</MOPComplete>`
				w.Write([]byte(response))
			}))
			defer server.Close()

			// Parse test server URL
			// server.URL is like "http://127.0.0.1:12345"
			serverURL := server.URL
			// Remove "http://"
			if len(serverURL) > 7 && serverURL[:7] == "http://" {
				serverURL = serverURL[7:]
			}
			// Split host and port
			colonIndex := -1
			for i := len(serverURL) - 1; i >= 0; i-- {
				if serverURL[i] == ':' {
					colonIndex = i
					break
				}
			}
			host := "127.0.0.1"
			port := 80
			if colonIndex > 0 {
				host = serverURL[:colonIndex]
				portStr := serverURL[colonIndex+1:]
				port = 0
				for _, c := range portStr {
					if c >= '0' && c <= '9' {
						port = port*10 + int(c-'0')
					}
				}
			}

			// Create adapter with test config
			config := &Config{
				Hostname:     host,
				Port:         port,
				PollInterval: tt.pollInterval,
				HTTPS:        false,
			}

			appState := state.New()
			adapter := NewAdapter(config, appState)

			// Connect and start polling
			err := adapter.Connect()
			if err != nil {
				t.Fatalf("Failed to connect: %v", err)
			}

			err = adapter.StartPolling()
			if err != nil {
				t.Fatalf("Failed to start polling: %v", err)
			}

			// Wait for test duration
			time.Sleep(tt.testDuration)

			// Stop adapter
			err = adapter.Stop()
			if err != nil {
				t.Errorf("Failed to stop adapter: %v", err)
			}

			// Check call count
			finalCount := int(atomic.LoadInt32(&callCount))
			if finalCount < tt.expectedMinCalls || finalCount > tt.expectedMaxCalls {
				t.Errorf("Poll count = %d, want between %d and %d", finalCount, tt.expectedMinCalls, tt.expectedMaxCalls)
			}
		})
	}
}

func TestAdapter_Connect(t *testing.T) {
	// Initialize logger for tests
	if err := logger.Init(); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		response := `<?xml version="1.0" encoding="UTF-8"?>
<MOPComplete nextdifference="abc123">
    <competition date="2024-01-01" organizer="Test Organizer" zerotime="10:00:00">Test Competition</competition>
</MOPComplete>`
		w.Write([]byte(response))
	}))
	defer server.Close()

	// Parse test server URL
	// server.URL is like "http://127.0.0.1:12345"
	serverURL := server.URL
	// Remove "http://"
	if len(serverURL) > 7 && serverURL[:7] == "http://" {
		serverURL = serverURL[7:]
	}
	// Split host and port
	colonIndex := -1
	for i := len(serverURL) - 1; i >= 0; i-- {
		if serverURL[i] == ':' {
			colonIndex = i
			break
		}
	}
	host := "127.0.0.1"
	port := 80
	if colonIndex > 0 {
		host = serverURL[:colonIndex]
		portStr := serverURL[colonIndex+1:]
		port = 0
		for _, c := range portStr {
			if c >= '0' && c <= '9' {
				port = port*10 + int(c-'0')
			}
		}
	}

	// Create adapter
	config := &Config{
		Hostname:     host,
		Port:         port,
		PollInterval: 1 * time.Second,
		HTTPS:        false,
	}

	appState := state.New()
	adapter := NewAdapter(config, appState)

	// Test connection
	err := adapter.Connect()
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Verify adapter is connected
	if !adapter.connected {
		t.Error("Adapter should be connected after successful Connect()")
	}

	// Verify state was updated
	event := appState.GetEvent()
	if event == nil {
		t.Fatal("Event should not be nil after connection")
	}
	if event.Name != "Test Competition" {
		t.Errorf("Event name = %q, want %q", event.Name, "Test Competition")
	}
}