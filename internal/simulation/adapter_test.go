package simulation

import (
	"testing"
	"time"

	"meos-graphics/internal/logger"
	"meos-graphics/internal/state"
)

func TestNewAdapter(t *testing.T) {
	appState := state.New()
	adapter := NewAdapter(appState, 15*time.Minute, 3*time.Minute, 7*time.Minute, 5*time.Minute, false, 3, 20, 3)

	if adapter == nil {
		t.Fatal("NewAdapter() returned nil")
	}
	if adapter.state != appState {
		t.Error("Adapter state not set correctly")
	}
	if adapter.generator == nil {
		t.Error("Adapter generator is nil")
	}
	if adapter.connected {
		t.Error("Adapter should not be connected initially")
	}
}

func TestAdapter_Connect(t *testing.T) {
	// Initialize logger for tests
	if err := logger.Init(); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	appState := state.New()
	adapter := NewAdapter(appState, 15*time.Minute, 3*time.Minute, 7*time.Minute, 5*time.Minute, false, 3, 20, 3)

	err := adapter.Connect()
	if err != nil {
		t.Fatalf("Connect() failed: %v", err)
	}

	// Verify adapter is connected
	if !adapter.connected {
		t.Error("Adapter should be connected after Connect()")
	}

	// Verify state was populated
	event := appState.GetEvent()
	if event == nil {
		t.Fatal("Event should not be nil after Connect()")
	}
	if event.Name != "Simulation Event" {
		t.Errorf("Event name = %q, want %q", event.Name, "Simulation Event")
	}

	controls := appState.GetControls()
	if len(controls) != 3 {
		t.Errorf("Number of controls = %d, want 3", len(controls))
	}

	classes := appState.GetClasses()
	if len(classes) != 3 {
		t.Errorf("Number of classes = %d, want 3", len(classes))
	}

	clubs := appState.GetClubs()
	if len(clubs) != len(clubNames) {
		t.Errorf("Number of clubs = %d, want %d", len(clubs), len(clubNames))
	}

	competitors := appState.GetCompetitors()
	if len(competitors) == 0 {
		t.Error("No competitors generated")
	}

	// All competitors should be in "not started" state initially
	for i, comp := range competitors {
		if comp.Status != "0" {
			t.Errorf("Competitor[%d] status = %q, want %q", i, comp.Status, "0")
		}
	}
}

func TestAdapter_StartPolling_NotConnected(t *testing.T) {
	appState := state.New()
	adapter := NewAdapter(appState, 15*time.Minute, 3*time.Minute, 7*time.Minute, 5*time.Minute, false, 3, 20, 3)

	// Should not error when not connected
	err := adapter.StartPolling()
	if err != nil {
		t.Errorf("StartPolling() when not connected should not error, got: %v", err)
	}
}

func TestAdapter_StartPolling_Connected(t *testing.T) {
	// Initialize logger for tests
	if err := logger.Init(); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	appState := state.New()
	adapter := NewAdapter(appState, 15*time.Minute, 3*time.Minute, 7*time.Minute, 5*time.Minute, false, 3, 20, 3)

	// Connect first
	err := adapter.Connect()
	if err != nil {
		t.Fatalf("Connect() failed: %v", err)
	}

	// Start polling
	err = adapter.StartPolling()
	if err != nil {
		t.Errorf("StartPolling() failed: %v", err)
	}

	// Verify ticker was created
	if adapter.ticker == nil {
		t.Error("Ticker should be created after StartPolling()")
	}

	// Stop to clean up
	adapter.Stop()
}

func TestAdapter_Stop(t *testing.T) {
	// Initialize logger for tests
	if err := logger.Init(); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	appState := state.New()
	adapter := NewAdapter(appState, 15*time.Minute, 3*time.Minute, 7*time.Minute, 5*time.Minute, false, 3, 20, 3)

	// Connect and start polling
	adapter.Connect()
	adapter.StartPolling()

	// Verify it's running
	if !adapter.connected {
		t.Error("Adapter should be connected before stop")
	}

	// Stop
	err := adapter.Stop()
	if err != nil {
		t.Errorf("Stop() failed: %v", err)
	}

	// Verify it stopped
	if adapter.connected {
		t.Error("Adapter should not be connected after Stop()")
	}

	// Should be safe to call Stop() again
	err = adapter.Stop()
	if err != nil {
		t.Errorf("Second Stop() call should not error: %v", err)
	}
}

func TestAdapter_UpdateSimulation(t *testing.T) {
	// Initialize logger for tests
	if err := logger.Init(); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	appState := state.New()
	adapter := NewAdapter(appState, 15*time.Minute, 3*time.Minute, 7*time.Minute, 5*time.Minute, false, 3, 20, 3)

	// Connect to initialize data
	adapter.Connect()

	// Get initial state
	initialCompetitors := appState.GetCompetitors()
	allNotStarted := true
	for _, comp := range initialCompetitors {
		if comp.Status != "0" {
			allNotStarted = false
			break
		}
	}
	if !allNotStarted {
		t.Error("All competitors should be 'not started' initially")
	}

	// Manually trigger an update (simulating time passage)
	adapter.updateSimulation()

	// State should be updated in the background
	// Note: In a real scenario, we'd mock time or use dependency injection
	// For now, we just verify the update function doesn't crash
}

func TestAdapter_SimulationCycle(t *testing.T) {
	// Initialize logger for tests
	if err := logger.Init(); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	appState := state.New()
	adapter := NewAdapter(appState, 15*time.Minute, 3*time.Minute, 7*time.Minute, 5*time.Minute, false, 3, 20, 3)

	// Connect to initialize
	adapter.Connect()

	// Get a baseline time before any updates
	event := appState.GetEvent()
	if event == nil {
		t.Fatal("Event should not be nil")
	}

	// Simulate the passage of time by directly calling the generator
	// This tests the integration between adapter and generator
	baseTime := event.Start

	// Test different phases
	phases := []struct {
		name    string
		elapsed time.Duration
	}{
		{"start", 0},
		{"early_running", 4 * time.Minute},
		{"late_running", 8 * time.Minute},
		{"all_finished", 12 * time.Minute},
		{"reset", 16 * time.Minute},
	}

	for _, phase := range phases {
		t.Run(phase.name, func(t *testing.T) {
			currentTime := baseTime.Add(phase.elapsed)
			competitors := adapter.generator.UpdateSimulation(currentTime)

			// Update the state with new competitors
			appState.Lock()
			appState.Competitors = competitors
			appState.Unlock()

			// Verify state consistency
			updatedCompetitors := appState.GetCompetitors()
			if len(updatedCompetitors) != len(competitors) {
				t.Errorf("State update failed: got %d competitors, want %d",
					len(updatedCompetitors), len(competitors))
			}

			// Log phase status for debugging
			statusCounts := make(map[string]int)
			for _, comp := range updatedCompetitors {
				statusCounts[comp.Status]++
			}
			t.Logf("Phase %s (%v): Status counts: %v", phase.name, phase.elapsed, statusCounts)
		})
	}
}

func TestAdapter_ConcurrentAccess(t *testing.T) {
	// Initialize logger for tests
	if err := logger.Init(); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	appState := state.New()
	adapter := NewAdapter(appState, 15*time.Minute, 3*time.Minute, 7*time.Minute, 5*time.Minute, false, 3, 20, 3)

	// Connect and start
	adapter.Connect()
	adapter.StartPolling()
	defer adapter.Stop()

	// Simulate concurrent access to state while simulation is running
	done := make(chan bool)

	// Reader goroutine
	go func() {
		for i := 0; i < 100; i++ {
			competitors := appState.GetCompetitors()
			if len(competitors) == 0 {
				t.Errorf("No competitors found during concurrent read %d", i)
			}
			time.Sleep(time.Millisecond)
		}
		done <- true
	}()

	// Another reader goroutine
	go func() {
		for i := 0; i < 100; i++ {
			event := appState.GetEvent()
			if event == nil {
				t.Errorf("Event is nil during concurrent read %d", i)
			}
			time.Sleep(time.Millisecond)
		}
		done <- true
	}()

	// Wait for readers to complete
	<-done
	<-done

	// If we get here without deadlock or panic, the test passes
}

func TestAdapter_ResetBehavior(t *testing.T) {
	// Initialize logger for tests
	if err := logger.Init(); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	appState := state.New()
	adapter := NewAdapter(appState, 15*time.Minute, 3*time.Minute, 7*time.Minute, 5*time.Minute, false, 3, 20, 3)

	adapter.Connect()

	// Manually advance generator to a finished state
	event := appState.GetEvent()
	finishedTime := event.Start.Add(14 * time.Minute) // Just before reset at 15 minutes
	competitors := adapter.generator.UpdateSimulation(finishedTime)

	// Verify some are finished
	finishedCount := 0
	for _, comp := range competitors {
		if comp.Status == "1" {
			finishedCount++
		}
	}
	if finishedCount == 0 {
		t.Fatal("No competitors finished before reset test")
	}

	// Trigger reset
	resetTime := event.Start.Add(16 * time.Minute)
	resetCompetitors := adapter.generator.UpdateSimulation(resetTime)

	// Verify all are reset
	for i, comp := range resetCompetitors {
		if comp.Status != "0" {
			t.Errorf("After reset: competitor[%d] status = %q, want %q", i, comp.Status, "0")
		}
		if comp.FinishTime != nil {
			t.Errorf("After reset: competitor[%d] should not have finish time", i)
		}
		if len(comp.Splits) != 0 {
			t.Errorf("After reset: competitor[%d] should not have splits", i)
		}
	}
}

func TestAdapter_StateConsistency(t *testing.T) {
	// Initialize logger for tests
	if err := logger.Init(); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	appState := state.New()
	adapter := NewAdapter(appState, 15*time.Minute, 3*time.Minute, 7*time.Minute, 5*time.Minute, false, 3, 20, 3)

	adapter.Connect()

	// Verify all entities have proper references
	competitors := appState.GetCompetitors()
	classes := appState.GetClasses()
	clubs := appState.GetClubs()
	controls := appState.GetControls()

	// Create lookup maps for validation
	classMap := make(map[int]bool)
	for _, class := range classes {
		classMap[class.ID] = true
	}

	clubMap := make(map[int]bool)
	for _, club := range clubs {
		clubMap[club.ID] = true
	}

	controlMap := make(map[int]bool)
	for _, control := range controls {
		controlMap[control.ID] = true
	}

	// Verify competitor references
	for i, comp := range competitors {
		// Check class reference
		if !classMap[comp.Class.ID] {
			t.Errorf("Competitor[%d] references non-existent class ID %d", i, comp.Class.ID)
		}

		// Check club reference
		if !clubMap[comp.Club.ID] {
			t.Errorf("Competitor[%d] references non-existent club ID %d", i, comp.Club.ID)
		}

		// Check control references in splits (after simulation runs)
		for j, split := range comp.Splits {
			if !controlMap[split.Control.ID] {
				t.Errorf("Competitor[%d] split[%d] references non-existent control ID %d",
					i, j, split.Control.ID)
			}
		}
	}
}

func BenchmarkAdapter_UpdateSimulation(b *testing.B) {
	// Initialize logger
	logger.Init()

	appState := state.New()
	adapter := NewAdapter(appState, 15*time.Minute, 3*time.Minute, 7*time.Minute, 5*time.Minute, false, 3, 20, 3)
	adapter.Connect()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		adapter.updateSimulation()
	}
}

func BenchmarkAdapter_StateRead(b *testing.B) {
	// Initialize logger
	logger.Init()

	appState := state.New()
	adapter := NewAdapter(appState, 15*time.Minute, 3*time.Minute, 7*time.Minute, 5*time.Minute, false, 3, 20, 3)
	adapter.Connect()

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			competitors := appState.GetCompetitors()
			_ = len(competitors) // Use the result to prevent optimization
		}
	})
}
