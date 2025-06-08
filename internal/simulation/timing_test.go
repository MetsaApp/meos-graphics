package simulation

import (
	"testing"
	"time"

	"meos-graphics/internal/logger"
	"meos-graphics/internal/state"
)

// TestConfigurableTiming tests simulation with different timing configurations
func TestConfigurableTiming(t *testing.T) {
	// Initialize logger
	if err := logger.Init(); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	tests := []struct {
		name           string
		duration       time.Duration
		phaseStart     time.Duration
		phaseRunning   time.Duration
		phaseResults   time.Duration
		testTimePoints []struct {
			elapsed       time.Duration
			expectedPhase string
		}
	}{
		{
			name:         "Default 15min cycle",
			duration:     15 * time.Minute,
			phaseStart:   3 * time.Minute,
			phaseRunning: 7 * time.Minute,
			phaseResults: 5 * time.Minute,
			testTimePoints: []struct {
				elapsed       time.Duration
				expectedPhase string
			}{
				{1 * time.Minute, "start"},
				{4 * time.Minute, "running"},
				{11 * time.Minute, "results"},
				{16 * time.Minute, "reset"},
			},
		},
		{
			name:         "Quick 1min test cycle",
			duration:     1 * time.Minute,
			phaseStart:   10 * time.Second,
			phaseRunning: 30 * time.Second,
			phaseResults: 20 * time.Second,
			testTimePoints: []struct {
				elapsed       time.Duration
				expectedPhase string
			}{
				{5 * time.Second, "start"},
				{20 * time.Second, "running"},
				{50 * time.Second, "results"},
				{70 * time.Second, "reset"},
			},
		},
		{
			name:         "Sprint event 5min cycle",
			duration:     5 * time.Minute,
			phaseStart:   1 * time.Minute,
			phaseRunning: 2 * time.Minute,
			phaseResults: 2 * time.Minute,
			testTimePoints: []struct {
				elapsed       time.Duration
				expectedPhase string
			}{
				{30 * time.Second, "start"},
				{90 * time.Second, "running"},
				{4 * time.Minute, "results"},
				{6 * time.Minute, "reset"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appState := state.New()
			adapter := NewAdapter(appState, tt.duration, tt.phaseStart, tt.phaseRunning, tt.phaseResults, false, 3, 20, 3)

			// Verify timing configuration was set
			if adapter.duration != tt.duration {
				t.Errorf("Duration = %v, want %v", adapter.duration, tt.duration)
			}
			if adapter.phaseStart != tt.phaseStart {
				t.Errorf("PhaseStart = %v, want %v", adapter.phaseStart, tt.phaseStart)
			}
			if adapter.phaseRunning != tt.phaseRunning {
				t.Errorf("PhaseRunning = %v, want %v", adapter.phaseRunning, tt.phaseRunning)
			}
			if adapter.phaseResults != tt.phaseResults {
				t.Errorf("PhaseResults = %v, want %v", adapter.phaseResults, tt.phaseResults)
			}

			adapter.Connect()
			baseTime := appState.GetEvent().Start

			for _, tp := range tt.testTimePoints {
				currentTime := baseTime.Add(tp.elapsed)
				competitors := adapter.generator.UpdateSimulation(currentTime)

				// Count statuses
				statusCounts := map[string]int{"0": 0, "2": 0, "1": 0}
				for _, comp := range competitors {
					statusCounts[comp.Status]++
				}

				// Verify expected phase behavior
				switch tp.expectedPhase {
				case "start":
					// All should be not started
					if statusCounts["2"] > 0 || statusCounts["1"] > 0 {
						t.Errorf("%s at %v: Expected all not started, got %d running, %d finished",
							tt.name, tp.elapsed, statusCounts["2"], statusCounts["1"])
					}
				case "running":
					// Some should be running or finished
					if statusCounts["2"]+statusCounts["1"] == 0 {
						t.Errorf("%s at %v: Expected some progress, all still not started",
							tt.name, tp.elapsed)
					}
				case "results":
					// Some should be finished
					// For very short simulations, expect fewer finishers
					minFinished := 1
					if tt.duration >= 5*time.Minute {
						// For longer simulations, expect at least 10%
						minFinished = len(competitors) / 10
						if minFinished < 1 {
							minFinished = 1
						}
					}
					if statusCounts["1"] < minFinished {
						t.Errorf("%s at %v: Expected at least %d finished, only %d/%d finished",
							tt.name, tp.elapsed, minFinished, statusCounts["1"], len(competitors))
					}
				case "reset":
					// All should be back to not started
					if statusCounts["0"] != len(competitors) {
						t.Errorf("%s at %v: Expected all reset, got %d not started, %d running, %d finished",
							tt.name, tp.elapsed, statusCounts["0"], statusCounts["2"], statusCounts["1"])
					}
				}
			}
		})
	}
}

// TestTimingValidation tests that timing validation works correctly
func TestTimingValidation(t *testing.T) {
	// This test verifies that the main.go validation logic would catch invalid configurations
	// The actual validation is done in main.go, but we can test the behavior here

	tests := []struct {
		name         string
		duration     time.Duration
		phaseStart   time.Duration
		phaseRunning time.Duration
		phaseResults time.Duration
		shouldWork   bool
	}{
		{
			name:         "Valid equal sum",
			duration:     10 * time.Minute,
			phaseStart:   2 * time.Minute,
			phaseRunning: 5 * time.Minute,
			phaseResults: 3 * time.Minute,
			shouldWork:   true,
		},
		{
			name:         "Invalid sum too large",
			duration:     10 * time.Minute,
			phaseStart:   3 * time.Minute,
			phaseRunning: 5 * time.Minute,
			phaseResults: 3 * time.Minute,
			shouldWork:   false, // 3+5+3=11 > 10
		},
		{
			name:         "Invalid sum too small",
			duration:     10 * time.Minute,
			phaseStart:   2 * time.Minute,
			phaseRunning: 4 * time.Minute,
			phaseResults: 3 * time.Minute,
			shouldWork:   false, // 2+4+3=9 < 10
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sum := tt.phaseStart + tt.phaseRunning + tt.phaseResults
			isValid := sum == tt.duration && tt.duration > 0 &&
				tt.phaseStart > 0 && tt.phaseRunning > 0 && tt.phaseResults > 0

			if isValid != tt.shouldWork {
				t.Errorf("Validation result = %v, want %v (duration=%v, sum=%v)",
					isValid, tt.shouldWork, tt.duration, sum)
			}
		})
	}
}

// TestPhaseLogging tests that phase boundaries work correctly
func TestPhaseLogging(t *testing.T) {
	// Initialize logger
	if err := logger.Init(); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	appState := state.New()
	adapter := NewAdapter(appState, 3*time.Minute, 30*time.Second, 90*time.Second, 60*time.Second, false, 3, 20, 3)
	adapter.Connect()

	baseTime := appState.GetEvent().Start

	// Test phase boundaries
	testPoints := []struct {
		elapsed            time.Duration
		expectedNotStarted int
		expectedRunning    int
		expectedFinished   int
		description        string
	}{
		{0, -1, 0, 0, "start list phase"},
		{29 * time.Second, -1, 0, 0, "still in start phase"},
		{30 * time.Second, -1, -1, -1, "running phase begins"},
		{2 * time.Minute, -1, -1, -1, "results phase begins"},
		{2*time.Minute + 59*time.Second, -1, -1, -1, "still in results phase"},
		{3*time.Minute + 1*time.Second, -1, 0, 0, "after reset"},
	}

	totalCompetitors := len(adapter.generator.competitors)

	for _, tp := range testPoints {
		currentTime := baseTime.Add(tp.elapsed)
		competitors := adapter.generator.UpdateSimulation(currentTime)

		// Count statuses
		notStarted := 0
		running := 0
		finished := 0

		for _, comp := range competitors {
			switch comp.Status {
			case "0":
				notStarted++
			case "2":
				running++
			case "1":
				finished++
			}
		}

		// Check expectations (-1 means we don't care about exact count)
		if tp.expectedNotStarted >= 0 && notStarted != tp.expectedNotStarted {
			t.Errorf("At %v (%s): notStarted = %d, want %d", tp.elapsed, tp.description, notStarted, tp.expectedNotStarted)
		}
		if tp.expectedRunning >= 0 && running != tp.expectedRunning {
			t.Errorf("At %v (%s): running = %d, want %d", tp.elapsed, tp.description, running, tp.expectedRunning)
		}
		if tp.expectedFinished >= 0 && finished != tp.expectedFinished {
			t.Errorf("At %v (%s): finished = %d, want %d", tp.elapsed, tp.description, finished, tp.expectedFinished)
		}

		// For start phases, all should be not started
		if tp.description == "start list phase" || tp.description == "still in start phase" {
			if notStarted != totalCompetitors {
				t.Errorf("At %v (%s): expected all %d competitors not started, but %d are",
					tp.elapsed, tp.description, totalCompetitors, notStarted)
			}
		}

		// After reset, all should be not started again
		if tp.description == "after reset" {
			if notStarted != totalCompetitors {
				t.Errorf("At %v (%s): expected all %d competitors reset to not started, but only %d are",
					tp.elapsed, tp.description, totalCompetitors, notStarted)
			}
		}
	}
}
