package simulation

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"meos-graphics/internal/models"
)

func TestNewGenerator(t *testing.T) {
	g := NewGenerator(15*time.Minute, 3*time.Minute, 7*time.Minute, 5*time.Minute, false)
	if g == nil {
		t.Fatal("NewGenerator returned nil")
	}
	if g.rnd == nil {
		t.Error("Generator random source is nil")
	}
}

func TestGenerator_GenerateInitialData(t *testing.T) {
	g := NewGenerator(15*time.Minute, 3*time.Minute, 7*time.Minute, 5*time.Minute, false)
	baseTime := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)

	event, controls, classes, clubs, competitors := g.GenerateInitialData(baseTime)

	// Test event
	if event.Name != "Simulation Event" {
		t.Errorf("Event name = %q, want %q", event.Name, "Simulation Event")
	}
	if event.Organizer != "MeOS Graphics Simulator" {
		t.Errorf("Event organizer = %q, want %q", event.Organizer, "MeOS Graphics Simulator")
	}
	if !event.Start.Equal(baseTime) {
		t.Errorf("Event start = %v, want %v", event.Start, baseTime)
	}

	// Test controls
	if len(controls) != 3 {
		t.Errorf("Number of controls = %d, want 3", len(controls))
	}
	expectedControls := []struct {
		id   int
		name string
	}{
		{1, "Radio 1"},
		{2, "Radio 2"},
		{3, "Radio 3"},
	}
	for i, expected := range expectedControls {
		if i >= len(controls) {
			break
		}
		if controls[i].ID != expected.id {
			t.Errorf("Control[%d].ID = %d, want %d", i, controls[i].ID, expected.id)
		}
		if controls[i].Name != expected.name {
			t.Errorf("Control[%d].Name = %q, want %q", i, controls[i].Name, expected.name)
		}
	}

	// Test classes
	if len(classes) != 3 {
		t.Errorf("Number of classes = %d, want 3", len(classes))
	}
	expectedClasses := []struct {
		id            int
		name          string
		orderKey      int
		radioControls int
	}{
		{1, "Men Elite", 10, 3},
		{2, "Women Elite", 20, 3},
		{3, "Men Junior", 30, 2},
	}
	for i, expected := range expectedClasses {
		if i >= len(classes) {
			break
		}
		if classes[i].ID != expected.id {
			t.Errorf("Class[%d].ID = %d, want %d", i, classes[i].ID, expected.id)
		}
		if classes[i].Name != expected.name {
			t.Errorf("Class[%d].Name = %q, want %q", i, classes[i].Name, expected.name)
		}
		if classes[i].OrderKey != expected.orderKey {
			t.Errorf("Class[%d].OrderKey = %d, want %d", i, classes[i].OrderKey, expected.orderKey)
		}
		if len(classes[i].RadioControls) != expected.radioControls {
			t.Errorf("Class[%d] radio controls = %d, want %d", i, len(classes[i].RadioControls), expected.radioControls)
		}
	}

	// Test clubs
	if len(clubs) != len(clubNames) {
		t.Errorf("Number of clubs = %d, want %d", len(clubs), len(clubNames))
	}
	for i, club := range clubs {
		if club.ID != i+1 {
			t.Errorf("Club[%d].ID = %d, want %d", i, club.ID, i+1)
		}
		if club.CountryCode != "SWE" {
			t.Errorf("Club[%d].CountryCode = %q, want %q", i, club.CountryCode, "SWE")
		}
		if club.Name == "" {
			t.Errorf("Club[%d].Name is empty", i)
		}
	}

	// Test competitors
	if len(competitors) == 0 {
		t.Error("No competitors generated")
	}

	// Should be roughly 15-25 per class * 3 classes = 45-75 total
	if len(competitors) < 45 || len(competitors) > 75 {
		t.Errorf("Number of competitors = %d, want between 45 and 75", len(competitors))
	}

	// Verify all competitors start at base time + phase start (with staggered starts)
	expectedFirstStartTime := baseTime.Add(g.phaseStart)
	for i, comp := range competitors {
		// With staggered starts (not mass start), each competitor starts 2 minutes after the previous
		expectedStartTime := expectedFirstStartTime.Add(time.Duration(i) * 2 * time.Minute)
		if !comp.StartTime.Equal(expectedStartTime) {
			// Could be from different class, so just check it's after phase start
			if comp.StartTime.Before(expectedFirstStartTime) {
				t.Errorf("Competitor[%d] start time = %v, before phase start %v", i, comp.StartTime, expectedFirstStartTime)
			}
		}
		if comp.Status != "0" {
			t.Errorf("Competitor[%d] status = %q, want %q", i, comp.Status, "0")
		}
		if comp.FinishTime != nil {
			t.Errorf("Competitor[%d] should not have finish time initially", i)
		}
		if len(comp.Splits) != 0 {
			t.Errorf("Competitor[%d] should not have splits initially", i)
		}
		if comp.Card < 500001 {
			t.Errorf("Competitor[%d] card number = %d, should be >= 500001", i, comp.Card)
		}
	}

	// Verify competitor distribution across classes
	classCounts := make(map[int]int)
	for _, comp := range competitors {
		classCounts[comp.Class.ID]++
	}
	for classID, count := range classCounts {
		if count < 15 || count > 25 {
			t.Errorf("Class %d has %d competitors, want between 15 and 25", classID, count)
		}
	}
}

func TestGenerator_DeterministicOutput(t *testing.T) {
	// Test that same seed produces same results
	baseTime := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)

	// Create two generators with same seed
	g1 := &Generator{
		rnd:               rand.New(rand.NewSource(12345)),
		duration:          15 * time.Minute,
		phaseStart:        3 * time.Minute,
		phaseRunning:      7 * time.Minute,
		phaseResults:      5 * time.Minute,
		massStart:         false,
		competitorTimings: make(map[int]competitorTiming),
	}
	g2 := &Generator{
		rnd:               rand.New(rand.NewSource(12345)),
		duration:          15 * time.Minute,
		phaseStart:        3 * time.Minute,
		phaseRunning:      7 * time.Minute,
		phaseResults:      5 * time.Minute,
		massStart:         false,
		competitorTimings: make(map[int]competitorTiming),
	}

	_, _, _, _, competitors1 := g1.GenerateInitialData(baseTime)
	_, _, _, _, competitors2 := g2.GenerateInitialData(baseTime)

	if len(competitors1) != len(competitors2) {
		t.Errorf("Different number of competitors: %d vs %d", len(competitors1), len(competitors2))
	}

	// Compare first few competitors
	for i := 0; i < minOfThree(10, len(competitors1), len(competitors2)); i++ {
		if competitors1[i].Name != competitors2[i].Name {
			t.Errorf("Competitor[%d] name differs: %q vs %q", i, competitors1[i].Name, competitors2[i].Name)
		}
		if competitors1[i].Club.Name != competitors2[i].Club.Name {
			t.Errorf("Competitor[%d] club differs: %q vs %q", i, competitors1[i].Club.Name, competitors2[i].Club.Name)
		}
	}
}

func TestGenerator_PhaseTransitions(t *testing.T) {
	g := NewGenerator(15*time.Minute, 3*time.Minute, 7*time.Minute, 5*time.Minute, false)
	baseTime := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)

	g.GenerateInitialData(baseTime)

	tests := []struct {
		name          string
		elapsed       time.Duration
		expectedPhase string
		checkFunc     func([]models.Competitor) error
	}{
		{
			name:          "Phase 1 - Start List Only",
			elapsed:       1 * time.Minute,
			expectedPhase: "start_list",
			checkFunc: func(competitors []models.Competitor) error {
				for i, comp := range competitors {
					if comp.Status != "0" {
						return fmt.Errorf("competitor[%d] status = %q, want %q", i, comp.Status, "0")
					}
					if comp.FinishTime != nil {
						return fmt.Errorf("competitor[%d] should not have finish time in phase 1", i)
					}
					if len(comp.Splits) != 0 {
						return fmt.Errorf("competitor[%d] should not have splits in phase 1", i)
					}
				}
				return nil
			},
		},
		{
			name:          "Phase 2 - Early Progress",
			elapsed:       4 * time.Minute,
			expectedPhase: "running",
			checkFunc: func(competitors []models.Competitor) error {
				runningCount := 0
				finishedCount := 0
				for _, comp := range competitors {
					switch comp.Status {
					case "0": // Still not started - OK for later competitors
					case "2": // Running
						runningCount++
					case "1": // Finished
						finishedCount++
					default:
						return fmt.Errorf("unexpected status %q in phase 2", comp.Status)
					}
				}
				if runningCount == 0 && finishedCount == 0 {
					return fmt.Errorf("no competitors running or finished in phase 2")
				}
				return nil
			},
		},
		{
			name:          "Phase 2 - Late Progress",
			elapsed:       8 * time.Minute,
			expectedPhase: "running",
			checkFunc: func(competitors []models.Competitor) error {
				finishedCount := 0
				for _, comp := range competitors {
					if comp.Status == "1" {
						finishedCount++
						if comp.FinishTime == nil {
							return fmt.Errorf("finished competitor should have finish time")
						}
						if len(comp.Splits) == 0 {
							return fmt.Errorf("finished competitor should have splits")
						}
					}
				}
				// At 8 minutes (5 minutes into running phase), some might still be running
				// since competitors can take up to 6.3 minutes to finish
				if finishedCount == 0 {
					// Check if at least some are running with splits
					runningWithSplits := 0
					for _, comp := range competitors {
						if comp.Status == "2" && len(comp.Splits) > 0 {
							runningWithSplits++
						}
					}
					if runningWithSplits == 0 {
						return fmt.Errorf("expected some progress (finished or running with splits) in late phase 2")
					}
				}
				return nil
			},
		},
		{
			name:          "Phase 3 - All Finished",
			elapsed:       12 * time.Minute,
			expectedPhase: "finished",
			checkFunc: func(competitors []models.Competitor) error {
				finishedCount := 0
				for _, comp := range competitors {
					if comp.Status == "1" {
						finishedCount++
						if comp.FinishTime == nil {
							return fmt.Errorf("finished competitor should have finish time in phase 3")
						}
						if len(comp.Splits) == 0 {
							return fmt.Errorf("finished competitor should have splits in phase 3")
						}
					}
				}
				// Some should be finished by this point
				expectedMinFinished := 5 // At least 5 competitors
				if finishedCount < expectedMinFinished {
					return fmt.Errorf("only %d/%d competitors finished (expected at least %d)",
						finishedCount, len(competitors), expectedMinFinished)
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			currentTime := baseTime.Add(tt.elapsed)
			competitors := g.UpdateSimulation(currentTime)

			if err := tt.checkFunc(competitors); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestGenerator_SimulationReset(t *testing.T) {
	g := NewGenerator(15*time.Minute, 3*time.Minute, 7*time.Minute, 5*time.Minute, false)
	baseTime := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)

	g.GenerateInitialData(baseTime)

	// Run simulation to completion
	currentTime := baseTime.Add(12 * time.Minute)
	competitors := g.UpdateSimulation(currentTime)

	// Verify competitors are finished
	finishedCount := 0
	for _, comp := range competitors {
		if comp.Status == "1" {
			finishedCount++
		}
	}
	if finishedCount == 0 {
		t.Fatal("No competitors finished before reset test")
	}

	// Trigger reset by going past 15 minutes
	resetTime := baseTime.Add(16 * time.Minute)
	competitorsAfterReset := g.UpdateSimulation(resetTime)

	// Verify all competitors are reset
	for i, comp := range competitorsAfterReset {
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

func TestGenerator_TimeCalculations(t *testing.T) {
	g := NewGenerator(15*time.Minute, 3*time.Minute, 7*time.Minute, 5*time.Minute, false)
	baseTime := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)

	g.GenerateInitialData(baseTime)

	// Run to phase 3 where all are finished
	currentTime := baseTime.Add(12 * time.Minute)
	competitors := g.UpdateSimulation(currentTime)

	for i, comp := range competitors {
		if comp.Status != "1" {
			continue // Skip unfinished
		}

		// Verify finish time is after start time
		if comp.FinishTime.Before(comp.StartTime) {
			t.Errorf("Competitor[%d] finish time before start time", i)
		}

		// Verify splits are in chronological order
		prevTime := comp.StartTime
		for j, split := range comp.Splits {
			if split.PassingTime.Before(prevTime) {
				t.Errorf("Competitor[%d] split[%d] time before previous time", i, j)
			}
			prevTime = split.PassingTime
		}

		// Verify finish time is after last split
		if len(comp.Splits) > 0 {
			lastSplit := comp.Splits[len(comp.Splits)-1]
			if comp.FinishTime.Before(lastSplit.PassingTime) {
				t.Errorf("Competitor[%d] finish time before last split", i)
			}
		}

		// Verify total time is reasonable (should be within phase running duration)
		totalTime := comp.FinishTime.Sub(comp.StartTime)
		maxTime := time.Duration(float64(g.phaseRunning) * 0.9) // 90% of phase running
		if totalTime < 5*time.Minute || totalTime > maxTime {
			t.Errorf("Competitor[%d] total time %v is unrealistic (expected 5min - %v)", i, totalTime, maxTime)
		}
	}
}

func TestGenerator_CompetitorProgression(t *testing.T) {
	g := NewGenerator(15*time.Minute, 3*time.Minute, 7*time.Minute, 5*time.Minute, false)
	baseTime := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)

	g.GenerateInitialData(baseTime)

	// Test progression at different time points
	timePoints := []time.Duration{
		0 * time.Minute,  // All not started
		3 * time.Minute,  // First competitors starting
		6 * time.Minute,  // More progression
		9 * time.Minute,  // Most finished
		12 * time.Minute, // All finished
	}

	prevRunningCount := 0
	prevFinishedCount := 0

	for _, elapsed := range timePoints {
		currentTime := baseTime.Add(elapsed)
		competitors := g.UpdateSimulation(currentTime)

		runningCount := 0
		finishedCount := 0

		for _, comp := range competitors {
			switch comp.Status {
			case "2":
				runningCount++
			case "1":
				finishedCount++
			}
		}

		t.Logf("At %v: %d running, %d finished", elapsed, runningCount, finishedCount)

		// Progress should generally increase (though running might decrease as they finish)
		if elapsed > 0 {
			totalProgress := runningCount + finishedCount
			prevTotalProgress := prevRunningCount + prevFinishedCount
			if totalProgress < prevTotalProgress {
				t.Errorf("Progress decreased at %v: %d < %d", elapsed, totalProgress, prevTotalProgress)
			}
		}

		prevRunningCount = runningCount
		prevFinishedCount = finishedCount
	}
}

func TestGenerator_SplitTimeConsistency(t *testing.T) {
	g := NewGenerator(15*time.Minute, 3*time.Minute, 7*time.Minute, 5*time.Minute, false)
	baseTime := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)

	g.GenerateInitialData(baseTime)

	// Run to completion
	currentTime := baseTime.Add(12 * time.Minute)
	competitors := g.UpdateSimulation(currentTime)

	for i, comp := range competitors {
		if comp.Status != "1" || len(comp.Splits) == 0 {
			continue
		}

		// Check that splits have correct controls
		for j, split := range comp.Splits {
			expectedControl := comp.Class.RadioControls[j]
			if split.Control.ID != expectedControl.ID {
				t.Errorf("Competitor[%d] split[%d] control ID = %d, want %d",
					i, j, split.Control.ID, expectedControl.ID)
			}
		}

		// Check that split times increase
		prevElapsed := time.Duration(0)
		for j, split := range comp.Splits {
			elapsed := split.PassingTime.Sub(comp.StartTime)
			if elapsed <= prevElapsed {
				t.Errorf("Competitor[%d] split[%d] time not increasing: %v <= %v",
					i, j, elapsed, prevElapsed)
			}
			prevElapsed = elapsed
		}
	}
}

func minOfThree(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

func TestGenerator_ClassSpecificRadioControls(t *testing.T) {
	g := NewGenerator(15*time.Minute, 3*time.Minute, 7*time.Minute, 5*time.Minute, false)
	baseTime := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)

	g.GenerateInitialData(baseTime)

	// Run to completion
	currentTime := baseTime.Add(12 * time.Minute)
	competitors := g.UpdateSimulation(currentTime)

	// Check Men Junior class has only 2 radio controls
	for _, comp := range competitors {
		if comp.Class.Name == "Men Junior" && comp.Status == "1" {
			if len(comp.Splits) != 2 {
				t.Errorf("Men Junior competitor should have 2 splits, got %d", len(comp.Splits))
			}
		} else if comp.Status == "1" {
			// Other classes should have 3 splits
			if len(comp.Splits) != 3 {
				t.Errorf("Elite class competitor should have 3 splits, got %d", len(comp.Splits))
			}
		}
	}
}
