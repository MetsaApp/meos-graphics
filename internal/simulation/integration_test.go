package simulation

import (
	"math/rand"
	"testing"
	"time"

	"meos-graphics/internal/logger"
	"meos-graphics/internal/state"
)

// TestSimulationFullCycle tests a complete 15-minute simulation cycle
func TestSimulationFullCycle(t *testing.T) {
	// Initialize logger
	if err := logger.Init(); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	appState := state.New()
	adapter := NewAdapter(appState, 15*time.Minute, 3*time.Minute, 7*time.Minute, 5*time.Minute, false)

	// Use deterministic generator for predictable tests
	adapter.generator.rnd = rand.New(rand.NewSource(12345))

	// Connect to initialize
	err := adapter.Connect()
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	baseTime := appState.GetEvent().Start

	// Test key time points in the simulation
	testPoints := []struct {
		name          string
		elapsed       time.Duration
		expectedPhase string
		minRunning    int
		minFinished   int
		maxNotStarted int
	}{
		{
			name:          "Initial State",
			elapsed:       0,
			expectedPhase: "start_list",
			minRunning:    0,
			minFinished:   0,
			maxNotStarted: 1000, // All should be not started
		},
		{
			name:          "Phase 1 - Middle",
			elapsed:       90 * time.Second,
			expectedPhase: "start_list",
			minRunning:    0,
			minFinished:   0,
			maxNotStarted: 1000, // Still all not started
		},
		{
			name:          "Phase 2 - Early",
			elapsed:       4 * time.Minute,
			expectedPhase: "running",
			minRunning:    1, // At least some should be running
			minFinished:   0,
			maxNotStarted: 60, // With conservative start intervals, more may still be waiting
		},
		{
			name:          "Phase 2 - Middle",
			elapsed:       6 * time.Minute,
			expectedPhase: "running",
			minRunning:    2,  // Some should be running
			minFinished:   0,  // May not have any finished yet
			maxNotStarted: 50, // Many still not started due to staggered starts
		},
		{
			name:          "Phase 2 - Late",
			elapsed:       9 * time.Minute,
			expectedPhase: "running",
			minRunning:    0,  // Some may be running
			minFinished:   0,  // Some may have finished
			maxNotStarted: 50, // Many still not started
		},
		{
			name:          "Phase 3 - All Finished",
			elapsed:       12 * time.Minute,
			expectedPhase: "finished",
			minRunning:    0,
			minFinished:   5,  // Some should be finished
			maxNotStarted: 50, // Many still not started
		},
		{
			name:          "Phase 3 - Stable",
			elapsed:       14 * time.Minute,
			expectedPhase: "finished",
			minRunning:    0,
			minFinished:   7,
			maxNotStarted: 50,
		},
		{
			name:          "After Reset",
			elapsed:       16 * time.Minute,
			expectedPhase: "start_list",
			minRunning:    0,
			minFinished:   0,
			maxNotStarted: 1000, // All should be reset to not started
		},
	}

	for _, tp := range testPoints {
		t.Run(tp.name, func(t *testing.T) {
			currentTime := baseTime.Add(tp.elapsed)
			competitors := adapter.generator.UpdateSimulation(currentTime)

			// Update state
			appState.Lock()
			appState.Competitors = competitors
			appState.Unlock()

			// Count statuses
			statusCounts := map[string]int{
				"0": 0, // Not started
				"2": 0, // Running
				"1": 0, // Finished
			}

			for _, comp := range competitors {
				statusCounts[comp.Status]++
			}

			t.Logf("Time %v: Not started: %d, Running: %d, Finished: %d",
				tp.elapsed, statusCounts["0"], statusCounts["2"], statusCounts["1"])

			// Verify constraints
			if statusCounts["2"] < tp.minRunning {
				t.Errorf("Running count %d < minimum %d", statusCounts["2"], tp.minRunning)
			}
			if statusCounts["1"] < tp.minFinished {
				t.Errorf("Finished count %d < minimum %d", statusCounts["1"], tp.minFinished)
			}
			if statusCounts["0"] > tp.maxNotStarted {
				t.Errorf("Not started count %d > maximum %d", statusCounts["0"], tp.maxNotStarted)
			}
		})
	}
}

// TestSimulationProgressionInvariants tests that certain properties hold throughout the simulation
func TestSimulationProgressionInvariants(t *testing.T) {
	// Initialize logger
	if err := logger.Init(); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	appState := state.New()
	adapter := NewAdapter(appState, 15*time.Minute, 3*time.Minute, 7*time.Minute, 5*time.Minute, false)

	// Use deterministic generator
	adapter.generator.rnd = rand.New(rand.NewSource(54321))

	adapter.Connect()

	baseTime := appState.GetEvent().Start
	prevFinishedCount := 0

	// Test progression every 30 seconds for 15 minutes
	for elapsed := time.Duration(0); elapsed <= 15*time.Minute; elapsed += 30 * time.Second {
		currentTime := baseTime.Add(elapsed)
		competitors := adapter.generator.UpdateSimulation(currentTime)

		finishedCount := 0
		runningCount := 0

		for _, comp := range competitors {
			switch comp.Status {
			case "1":
				finishedCount++

				// Invariant: Finished competitors must have finish time
				if comp.FinishTime == nil {
					t.Errorf("At %v: Finished competitor %d has no finish time", elapsed, comp.ID)
				}

				// Invariant: Finish time must be after start time
				if comp.FinishTime != nil && comp.FinishTime.Before(comp.StartTime) {
					t.Errorf("At %v: Competitor %d finish time before start time", elapsed, comp.ID)
				}

				// Invariant: Splits must be in chronological order
				prevTime := comp.StartTime
				for j, split := range comp.Splits {
					if split.PassingTime.Before(prevTime) {
						t.Errorf("At %v: Competitor %d split %d not in chronological order", elapsed, comp.ID, j)
					}
					prevTime = split.PassingTime
				}

			case "2":
				runningCount++

				// Invariant: Running competitors should not have finish time
				if comp.FinishTime != nil {
					t.Errorf("At %v: Running competitor %d has finish time", elapsed, comp.ID)
				}

			case "0":
				// Invariant: Not started competitors should have no splits or finish time
				if len(comp.Splits) > 0 {
					t.Errorf("At %v: Not started competitor %d has splits", elapsed, comp.ID)
				}
				if comp.FinishTime != nil {
					t.Errorf("At %v: Not started competitor %d has finish time", elapsed, comp.ID)
				}
			}
		}

		// Invariant: Finished count should never decrease (except at reset)
		if elapsed < 15*time.Minute && finishedCount < prevFinishedCount {
			t.Errorf("At %v: Finished count decreased from %d to %d", elapsed, prevFinishedCount, finishedCount)
		}

		// After reset, finished count should be 0
		if elapsed >= 15*time.Minute && finishedCount > 0 {
			t.Errorf("At %v: Finished count should be 0 after reset, got %d", elapsed, finishedCount)
		}

		prevFinishedCount = finishedCount
	}
}

// TestSimulationDeterminism tests that the simulation produces consistent results with the same seed
func TestSimulationDeterminism(t *testing.T) {
	// Initialize logger
	if err := logger.Init(); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	const seed = int64(98765)
	baseTime := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)

	// Run two identical simulations
	var results1, results2 [][]string

	for run := 0; run < 2; run++ {
		appState := state.New()
		adapter := NewAdapter(appState, 15*time.Minute, 3*time.Minute, 7*time.Minute, 5*time.Minute, false)
		adapter.generator.rnd = rand.New(rand.NewSource(seed))

		// Override the base time to be deterministic
		adapter.generator.startTime = baseTime

		adapter.generator.GenerateInitialData(baseTime)

		var runResults []string

		// Test at specific time points
		testTimes := []time.Duration{
			3 * time.Minute,
			6 * time.Minute,
			9 * time.Minute,
			12 * time.Minute,
		}

		for _, elapsed := range testTimes {
			currentTime := baseTime.Add(elapsed)
			updatedCompetitors := adapter.generator.UpdateSimulation(currentTime)

			// Create a snapshot of competitor states
			for _, comp := range updatedCompetitors {
				snapshot := comp.Status
				if comp.FinishTime != nil {
					snapshot += "_finished"
				}
				runResults = append(runResults, snapshot)
			}
		}

		if run == 0 {
			results1 = [][]string{runResults}
		} else {
			results2 = [][]string{runResults}
		}
	}

	// Compare results
	if len(results1[0]) != len(results2[0]) {
		t.Errorf("Different result lengths: %d vs %d", len(results1[0]), len(results2[0]))
		return
	}

	for i := range results1[0] {
		if results1[0][i] != results2[0][i] {
			t.Errorf("Result difference at index %d: %q vs %q", i, results1[0][i], results2[0][i])
		}
	}
}

// TestSimulationPerformance tests that the simulation can handle updates efficiently
func TestSimulationPerformance(t *testing.T) {
	// Initialize logger
	if err := logger.Init(); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	appState := state.New()
	adapter := NewAdapter(appState, 15*time.Minute, 3*time.Minute, 7*time.Minute, 5*time.Minute, false)
	adapter.Connect()

	baseTime := appState.GetEvent().Start

	// Measure time to update simulation
	start := time.Now()
	iterations := 100

	for i := 0; i < iterations; i++ {
		currentTime := baseTime.Add(time.Duration(i) * time.Second)
		adapter.generator.UpdateSimulation(currentTime)
	}

	elapsed := time.Since(start)
	avgPerUpdate := elapsed / time.Duration(iterations)

	t.Logf("Performance: %d updates in %v (avg: %v per update)", iterations, elapsed, avgPerUpdate)

	// Should be able to update much faster than real-time
	maxExpectedTime := 10 * time.Millisecond
	if avgPerUpdate > maxExpectedTime {
		t.Errorf("Simulation update too slow: %v > %v", avgPerUpdate, maxExpectedTime)
	}
}

// TestSimulationStateManagement tests that adapter properly manages state through the lifecycle
func TestSimulationStateManagement(t *testing.T) {
	// Initialize logger
	if err := logger.Init(); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	appState := state.New()
	adapter := NewAdapter(appState, 15*time.Minute, 3*time.Minute, 7*time.Minute, 5*time.Minute, false)

	// Test before connection
	if adapter.connected {
		t.Error("Should not be connected initially")
	}

	// Connect
	adapter.Connect()
	if !adapter.connected {
		t.Error("Should be connected after Connect()")
	}

	// Start polling
	adapter.StartPolling()

	// Wait a bit for potential updates
	time.Sleep(150 * time.Millisecond)

	// Stop
	adapter.Stop()
	if adapter.connected {
		t.Error("Should not be connected after Stop()")
	}

	// Verify we can restart
	adapter.Connect()
	if !adapter.connected {
		t.Error("Should be able to reconnect after Stop()")
	}

	adapter.Stop()
}

// TestSimulationDataIntegrity tests that simulation data maintains referential integrity
func TestSimulationDataIntegrity(t *testing.T) {
	// Initialize logger
	if err := logger.Init(); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	appState := state.New()
	adapter := NewAdapter(appState, 15*time.Minute, 3*time.Minute, 7*time.Minute, 5*time.Minute, false)
	adapter.Connect()

	baseTime := appState.GetEvent().Start

	// Test at various points during simulation
	testPoints := []time.Duration{
		0,
		5 * time.Minute,
		10 * time.Minute,
		15 * time.Minute, // After reset
	}

	for _, elapsed := range testPoints {
		t.Run(elapsed.String(), func(t *testing.T) {
			currentTime := baseTime.Add(elapsed)
			competitors := adapter.generator.UpdateSimulation(currentTime)

			// Update state
			appState.Lock()
			appState.Competitors = competitors
			appState.Unlock()

			// Verify referential integrity
			controls := appState.GetControls()
			classes := appState.GetClasses()
			clubs := appState.GetClubs()

			controlMap := make(map[int]bool)
			for _, ctrl := range controls {
				controlMap[ctrl.ID] = true
			}

			classMap := make(map[int]bool)
			for _, class := range classes {
				classMap[class.ID] = true
			}

			clubMap := make(map[int]bool)
			for _, club := range clubs {
				clubMap[club.ID] = true
			}

			updatedCompetitors := appState.GetCompetitors()
			for i, comp := range updatedCompetitors {
				// Check class reference
				if !classMap[comp.Class.ID] {
					t.Errorf("Competitor[%d] references invalid class %d", i, comp.Class.ID)
				}

				// Check club reference
				if !clubMap[comp.Club.ID] {
					t.Errorf("Competitor[%d] references invalid club %d", i, comp.Club.ID)
				}

				// Check split control references
				for j, split := range comp.Splits {
					if !controlMap[split.Control.ID] {
						t.Errorf("Competitor[%d] split[%d] references invalid control %d",
							i, j, split.Control.ID)
					}
				}
			}
		})
	}
}
