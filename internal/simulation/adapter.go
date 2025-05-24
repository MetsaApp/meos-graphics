package simulation

import (
	"sync"
	"time"

	"meos-graphics/internal/logger"
	"meos-graphics/internal/state"
)

type Adapter struct {
	state     *state.State
	generator *Generator
	connected bool
	mu        sync.RWMutex
	stopChan  chan struct{}
	ticker    *time.Ticker

	// Timing configuration
	duration     time.Duration
	phaseStart   time.Duration
	phaseRunning time.Duration
	phaseResults time.Duration
}

func NewAdapter(appState *state.State, duration, phaseStart, phaseRunning, phaseResults time.Duration) *Adapter {
	return &Adapter{
		state:        appState,
		generator:    NewGenerator(duration, phaseStart, phaseRunning, phaseResults),
		stopChan:     make(chan struct{}),
		duration:     duration,
		phaseStart:   phaseStart,
		phaseRunning: phaseRunning,
		phaseResults: phaseResults,
	}
}

func (a *Adapter) Connect() error {
	logger.InfoLogger.Println("Starting simulation mode")

	// Generate initial data
	baseTime := time.Now()
	event, controls, classes, clubs, competitors := a.generator.GenerateInitialData(baseTime)

	// Update state
	a.state.Lock()
	a.state.Event = &event
	a.state.Controls = controls
	a.state.Classes = classes
	a.state.Clubs = clubs
	a.state.Competitors = competitors
	a.state.Unlock()

	a.mu.Lock()
	a.connected = true
	// Recreate the stop channel in case it was closed before
	a.stopChan = make(chan struct{})
	a.mu.Unlock()

	logger.InfoLogger.Printf("Simulation initialized with %d classes and %d competitors",
		len(classes), len(competitors))

	return nil
}

func (a *Adapter) StartPolling() error {
	a.mu.RLock()
	if !a.connected {
		a.mu.RUnlock()
		return nil
	}
	a.mu.RUnlock()

	// Update every 100ms for smooth simulation
	a.ticker = time.NewTicker(100 * time.Millisecond)

	go func() {
		for {
			select {
			case <-a.stopChan:
				return
			case <-a.ticker.C:
				a.updateSimulation()
			}
		}
	}()

	logger.InfoLogger.Println("Started simulation updates")
	return nil
}

func (a *Adapter) Stop() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.connected {
		if a.ticker != nil {
			a.ticker.Stop()
		}
		close(a.stopChan)
		a.connected = false
	}
	return nil
}

func (a *Adapter) updateSimulation() {
	currentTime := time.Now()

	// Get updated competitors
	competitors := a.generator.UpdateSimulation(currentTime)

	// Get current state
	event := a.state.GetEvent()
	controls := a.state.GetControls()
	classes := a.state.GetClasses()
	clubs := a.state.GetClubs()

	// Update state atomically and notify listeners
	a.state.UpdateFromMeOS(event, controls, classes, clubs, competitors)

	// Log phase changes
	elapsed := currentTime.Sub(a.generator.startTime)

	if elapsed.Truncate(time.Minute) == elapsed {
		phase := "start list"
		phaseRunningEnd := a.phaseStart + a.phaseRunning
		if elapsed >= a.phaseStart && elapsed < phaseRunningEnd {
			phase = "running"
		} else if elapsed >= phaseRunningEnd && elapsed < a.duration {
			phase = "results"
		}

		logger.DebugLogger.Printf("Simulation at %v - phase: %s", elapsed.Round(time.Second), phase)
	}

	// Check for reset
	if elapsed >= a.duration {
		logger.InfoLogger.Println("Simulation cycle complete, restarting...")
		a.generator.resetSimulation()
	}
}

// GetSimulationStatus returns the current simulation phase and timing
func (a *Adapter) GetSimulationStatus() (phase string, nextPhaseIn time.Duration, isSimulation bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	
	if !a.connected {
		return "", 0, false
	}
	
	phase, nextPhaseIn = a.generator.GetCurrentPhase()
	return phase, nextPhaseIn, true
}
