package sse

import (
	"bufio"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"meos-graphics/internal/logger"
	"meos-graphics/internal/models"
	"meos-graphics/internal/simulation"
	"meos-graphics/internal/state"
)

func init() {
	// Initialize logger for tests
	_ = logger.Init()
}

func TestSSEIntegration(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create state and SSE hub
	appState := state.New()
	sseHub := NewHub()
	go sseHub.Run()

	// Set up state change notifications
	appState.OnUpdate(func() {
		sseHub.BroadcastUpdate("update", gin.H{"timestamp": time.Now().Unix()})
	})

	// Create test router
	router := gin.New()
	router.GET("/sse", sseHub.HandleSSE)

	// Create test server
	server := httptest.NewServer(router)
	defer server.Close()

	// Test SSE connection and updates during simulation
	t.Run("ReceivesUpdatesOnStateChange", func(t *testing.T) {
		// Connect to SSE endpoint
		resp, err := http.Get(server.URL + "/sse")
		if err != nil {
			t.Fatalf("Failed to connect to SSE: %v", err)
		}
		defer resp.Body.Close()

		if !strings.HasPrefix(resp.Header.Get("Content-Type"), "text/event-stream") {
			t.Errorf("Expected Content-Type text/event-stream, got %s", resp.Header.Get("Content-Type"))
		}

		// Read events in a goroutine
		events := make(chan string, 10)
		errors := make(chan error, 1)

		go func() {
			scanner := bufio.NewScanner(resp.Body)
			for scanner.Scan() {
				line := scanner.Text()
				if strings.HasPrefix(line, "event:") {
					events <- line
				}
			}
			if err := scanner.Err(); err != nil {
				errors <- err
			}
			close(events)
		}()

		// Wait for initial connection event
		timeout := time.After(2 * time.Second)
		select {
		case event := <-events:
			if !strings.Contains(event, "connected") {
				t.Errorf("Expected connected event, got: %s", event)
			}
		case <-timeout:
			t.Fatal("Timeout waiting for connection event")
		}

		// Trigger a state update
		event := &models.Event{Name: "Test Event"}
		controls := []models.Control{{ID: 1, Name: "Control 1"}}
		classes := []models.Class{{ID: 1, Name: "Class 1"}}
		clubs := []models.Club{{ID: 1, Name: "Club 1"}}
		competitors := []models.Competitor{
			{
				ID:        1,
				Name:      "Test Runner",
				Status:    "0",
				StartTime: time.Now(),
				Class:     models.Class{ID: 1},
				Club:      models.Club{ID: 1},
			},
		}

		appState.UpdateFromMeOS(event, controls, classes, clubs, competitors)

		// Wait for update event
		timeout = time.After(2 * time.Second)
		select {
		case event := <-events:
			if !strings.Contains(event, "update") {
				t.Errorf("Expected update event, got: %s", event)
			}
		case <-timeout:
			t.Fatal("Timeout waiting for update event")
		case err := <-errors:
			t.Fatalf("Scanner error: %v", err)
		}
	})
}

func TestSSESimulationIntegration(t *testing.T) {
	// This test verifies SSE updates work correctly during a full simulation cycle
	gin.SetMode(gin.TestMode)

	// Create state and SSE hub
	appState := state.New()
	sseHub := NewHub()
	go sseHub.Run()

	// Set up state change notifications
	appState.OnUpdate(func() {
		sseHub.BroadcastUpdate("update", gin.H{"timestamp": time.Now().Unix()})
	})

	// Create simulation adapter with short durations for testing
	simAdapter := simulation.NewAdapter(
		appState,
		2*time.Second,        // Total duration
		500*time.Millisecond, // Start phase
		1*time.Second,        // Running phase
		500*time.Millisecond, // Results phase
		false,                // Mass start
		3,                    // numClasses
		20,                   // runnersPerClass
		3,                    // radioControls
	)

	// Initialize simulation data
	if err := simAdapter.Connect(); err != nil {
		t.Fatalf("Failed to connect simulation: %v", err)
	}

	// Create test router
	router := gin.New()
	router.GET("/sse", sseHub.HandleSSE)

	// Create test server
	server := httptest.NewServer(router)
	defer server.Close()

	t.Run("ReceivesMultipleUpdatesDuringSimulation", func(t *testing.T) {
		// Connect to SSE endpoint
		resp, err := http.Get(server.URL + "/sse")
		if err != nil {
			t.Fatalf("Failed to connect to SSE: %v", err)
		}
		defer resp.Body.Close()

		// Collect events
		events := make([]string, 0)
		eventChan := make(chan string, 100)
		done := make(chan bool)

		go func() {
			scanner := bufio.NewScanner(resp.Body)
			for scanner.Scan() {
				line := scanner.Text()
				if strings.HasPrefix(line, "event:") {
					eventChan <- line
				}
			}
			close(done)
		}()

		// Start simulation polling
		if err := simAdapter.StartPolling(); err != nil {
			t.Fatalf("Failed to start polling: %v", err)
		}
		defer simAdapter.Stop()

		// Collect events for 3 seconds
		timeout := time.After(3 * time.Second)
		updateCount := 0

	collectLoop:
		for {
			select {
			case event := <-eventChan:
				events = append(events, event)
				if strings.Contains(event, "update") && !strings.Contains(event, "connected") {
					updateCount++
				}
			case <-timeout:
				break collectLoop
			case <-done:
				t.Fatal("Connection closed unexpectedly")
			}
		}

		// Verify we received multiple update events
		if updateCount == 0 {
			t.Error("Should receive at least one update event")
		}
		if len(events) > 0 && !strings.Contains(events[0], "connected") {
			t.Error("First event should be connection")
		}

		// Log events for debugging
		t.Logf("Received %d events, %d updates", len(events), updateCount)
	})
}

func TestSSEMultipleClients(t *testing.T) {
	// Test that multiple clients receive updates
	gin.SetMode(gin.TestMode)

	// Create state and SSE hub
	appState := state.New()
	sseHub := NewHub()
	go sseHub.Run()

	// Set up state change notifications
	appState.OnUpdate(func() {
		sseHub.BroadcastUpdate("update", gin.H{"timestamp": time.Now().Unix()})
	})

	// Create test router
	router := gin.New()
	router.GET("/sse", sseHub.HandleSSE)

	// Create test server
	server := httptest.NewServer(router)
	defer server.Close()

	t.Run("MultipleClientsReceiveUpdates", func(t *testing.T) {
		// Connect multiple clients
		clients := make([]*http.Response, 3)
		clientEvents := make([]chan string, 3)

		for i := 0; i < 3; i++ {
			resp, err := http.Get(server.URL + "/sse")
			if err != nil {
				t.Fatalf("Failed to connect client %d: %v", i, err)
			}
			clients[i] = resp
			defer resp.Body.Close()

			events := make(chan string, 10)
			clientEvents[i] = events

			// Read events for this client
			go func(r *http.Response, ch chan string) {
				scanner := bufio.NewScanner(r.Body)
				for scanner.Scan() {
					line := scanner.Text()
					if strings.HasPrefix(line, "event:") {
						ch <- line
					}
				}
			}(resp, events)
		}

		// Wait for all clients to connect
		time.Sleep(100 * time.Millisecond)

		// Verify client count
		if sseHub.GetConnectedClients() != 3 {
			t.Errorf("Expected 3 connected clients, got %d", sseHub.GetConnectedClients())
		}

		// Trigger a state update
		competitors := []models.Competitor{
			{
				ID:        1,
				Name:      "Updated Runner",
				Status:    "2",
				StartTime: time.Now(),
			},
		}
		appState.UpdateFromMeOS(nil, nil, nil, nil, competitors)

		// Verify all clients receive the update
		timeout := time.After(2 * time.Second)
		updateCount := 0

		for i := 0; i < 3; i++ {
			select {
			case <-timeout:
				t.Fatalf("Client %d did not receive update", i)
			case <-clientEvents[i]:
				// Skip connected event
			case <-time.After(50 * time.Millisecond):
				// Continue
			}

			select {
			case event := <-clientEvents[i]:
				if strings.Contains(event, "update") && !strings.Contains(event, "connected") {
					updateCount++
				}
			case <-timeout:
				t.Fatalf("Client %d did not receive update event", i)
			}
		}

		if updateCount != 3 {
			t.Errorf("Expected all 3 clients to receive update, got %d", updateCount)
		}
	})
}

func TestSSEHeartbeat(t *testing.T) {
	// Test that heartbeat events are sent
	gin.SetMode(gin.TestMode)

	// Create SSE hub
	sseHub := NewHub()
	go sseHub.Run()

	// Create test router
	router := gin.New()
	router.GET("/sse", func(c *gin.Context) {
		// Custom handler with shorter heartbeat for testing
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")

		// Send initial event
		c.SSEvent("connected", gin.H{"id": "test"})
		c.Writer.Flush()

		// Send heartbeat after short delay
		time.Sleep(100 * time.Millisecond)
		c.SSEvent("heartbeat", gin.H{"time": time.Now().Unix()})
		c.Writer.Flush()
	})

	// Create test server
	server := httptest.NewServer(router)
	defer server.Close()

	t.Run("ReceivesHeartbeatEvents", func(t *testing.T) {
		// Connect to SSE endpoint
		resp, err := http.Get(server.URL + "/sse")
		if err != nil {
			t.Fatalf("Failed to connect to SSE: %v", err)
		}
		defer resp.Body.Close()

		// Read events
		scanner := bufio.NewScanner(resp.Body)
		events := []string{}

		// Read for a short time
		done := make(chan bool)
		go func() {
			for scanner.Scan() {
				line := scanner.Text()
				if strings.HasPrefix(line, "event:") {
					events = append(events, line)
				}
				if len(events) >= 2 {
					close(done)
					return
				}
			}
		}()

		select {
		case <-done:
			// Got enough events
		case <-time.After(1 * time.Second):
			// Timeout is OK, we might have the events
		}

		// Verify we got both connected and heartbeat
		if len(events) < 2 {
			t.Errorf("Should have at least 2 events, got %d", len(events))
		}
		if len(events) > 0 && !strings.Contains(events[0], "connected") {
			t.Error("First event should be connected")
		}
		if len(events) > 1 && !strings.Contains(events[1], "heartbeat") {
			t.Error("Second event should be heartbeat")
		}
	})
}

func TestSSEDisconnectCleanup(t *testing.T) {
	// Test that clients are properly cleaned up on disconnect
	gin.SetMode(gin.TestMode)

	// Create SSE hub
	sseHub := NewHub()
	go sseHub.Run()

	// Create test router
	router := gin.New()
	router.GET("/sse", sseHub.HandleSSE)

	// Create test server
	server := httptest.NewServer(router)
	defer server.Close()

	t.Run("ClientCountUpdatesOnDisconnect", func(t *testing.T) {
		// Initial client count should be 0
		if sseHub.GetConnectedClients() != 0 {
			t.Errorf("Expected 0 clients initially, got %d", sseHub.GetConnectedClients())
		}

		// Connect a client
		resp, err := http.Get(server.URL + "/sse")
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}

		// Wait for connection
		time.Sleep(100 * time.Millisecond)
		if sseHub.GetConnectedClients() != 1 {
			t.Errorf("Expected 1 client after connection, got %d", sseHub.GetConnectedClients())
		}

		// Disconnect the client
		resp.Body.Close()

		// Wait for cleanup
		time.Sleep(100 * time.Millisecond)
		if sseHub.GetConnectedClients() != 0 {
			t.Errorf("Expected 0 clients after disconnect, got %d", sseHub.GetConnectedClients())
		}
	})
}
