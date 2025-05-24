package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"

	"meos-graphics/internal/cmd"
	"meos-graphics/internal/handlers"
	"meos-graphics/internal/logger"
	"meos-graphics/internal/meos"
	"meos-graphics/internal/middleware"
	"meos-graphics/internal/simulation"
	"meos-graphics/internal/state"
	"meos-graphics/internal/version"
)

func main() {
	rootCmd := cmd.NewRootCommand()
	rootCmd.RunE = run

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(_ *cobra.Command, _ []string) error {
	// Validate poll interval
	if cmd.PollInterval < 100*time.Millisecond {
		return fmt.Errorf("poll interval too small (minimum 100ms): %s", cmd.PollInterval)
	}
	if cmd.PollInterval > 1*time.Hour {
		return fmt.Errorf("poll interval too large (maximum 1 hour): %s", cmd.PollInterval)
	}

	// Check if simulation timing flags are used without simulation mode
	if !cmd.SimulationMode {
		// Check if any non-default simulation timing values are set
		defaultDuration := 15 * time.Minute
		defaultPhaseStart := 3 * time.Minute
		defaultPhaseRunning := 7 * time.Minute
		defaultPhaseResults := 5 * time.Minute

		if cmd.SimulationDuration != defaultDuration ||
			cmd.SimulationPhaseStart != defaultPhaseStart ||
			cmd.SimulationPhaseRunning != defaultPhaseRunning ||
			cmd.SimulationPhaseResults != defaultPhaseResults {
			return fmt.Errorf("simulation timing flags can only be used with --simulation mode")
		}
	}

	if err := logger.Init(); err != nil {
		return fmt.Errorf("failed to initialize logger: %v", err)
	}

	logger.InfoLogger.Printf("Starting MeOS Graphics API Server v%s", version.Version)
	if cmd.SimulationMode {
		logger.InfoLogger.Println("Running in SIMULATION MODE")

		// Validate simulation timing configuration
		if cmd.SimulationDuration <= 0 {
			return fmt.Errorf("simulation duration must be positive: %s", cmd.SimulationDuration)
		}

		// Validate each phase duration is positive
		if cmd.SimulationPhaseStart <= 0 {
			return fmt.Errorf("simulation-phase-start must be positive: %s", cmd.SimulationPhaseStart)
		}
		if cmd.SimulationPhaseRunning <= 0 {
			return fmt.Errorf("simulation-phase-running must be positive: %s", cmd.SimulationPhaseRunning)
		}
		if cmd.SimulationPhaseResults <= 0 {
			return fmt.Errorf("simulation-phase-results must be positive: %s", cmd.SimulationPhaseResults)
		}

		// Validate phase durations sum to total duration
		phaseSum := cmd.SimulationPhaseStart + cmd.SimulationPhaseRunning + cmd.SimulationPhaseResults
		if phaseSum != cmd.SimulationDuration {
			return fmt.Errorf("phase durations (%s + %s + %s = %s) must equal total duration (%s)",
				cmd.SimulationPhaseStart, cmd.SimulationPhaseRunning, cmd.SimulationPhaseResults,
				phaseSum, cmd.SimulationDuration)
		}

		logger.InfoLogger.Printf("Simulation timing: Total=%s, Start=%s, Running=%s, Results=%s",
			cmd.SimulationDuration, cmd.SimulationPhaseStart, cmd.SimulationPhaseRunning, cmd.SimulationPhaseResults)
	}

	// Initialize global state
	appState := state.New()

	// Create adapter based on mode
	var adapter interface {
		Connect() error
		StartPolling() error
		Stop() error
	}

	if cmd.SimulationMode {
		// Use simulation adapter with timing configuration
		adapter = simulation.NewAdapter(appState, cmd.SimulationDuration,
			cmd.SimulationPhaseStart, cmd.SimulationPhaseRunning, cmd.SimulationPhaseResults)
	} else {
		// Configure MeOS adapter
		config := meos.NewConfig()
		config.Hostname = cmd.MeosHost
		config.PortStr = cmd.MeosPort
		config.PollInterval = cmd.PollInterval

		// Log configuration based on port setting
		if config.PortStr == "none" {
			logger.InfoLogger.Printf("MeOS Configuration: %s (no port), Poll Interval: %s", config.Hostname, config.PollInterval)
		} else {
			logger.InfoLogger.Printf("MeOS Configuration: %s:%s, Poll Interval: %s", config.Hostname, config.PortStr, config.PollInterval)
		}

		if err := config.Validate(); err != nil {
			logger.ErrorLogger.Printf("Invalid configuration: %v", err)
			return err
		}

		adapter = meos.NewAdapter(config, appState)
	}

	// Connect adapter
	if err := adapter.Connect(); err != nil {
		logger.ErrorLogger.Printf("Failed to connect: %v", err)
		if !cmd.SimulationMode {
			logger.ErrorLogger.Println("Starting in offline mode - MeOS server not available")
		}
	} else {
		logger.InfoLogger.Println("Connected successfully")

		if err := adapter.StartPolling(); err != nil {
			logger.ErrorLogger.Printf("Failed to start polling: %v", err)
			logger.ErrorLogger.Println("Continuing without polling")
		} else {
			logger.InfoLogger.Println("Started polling for updates")
		}
	}

	// Set up HTTP server
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.Logger())

	// Create handlers
	h := handlers.New(appState)

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":         "ok",
			"meos_connected": true,
		})
	})

	// API endpoints
	router.GET("/classes", h.GetClasses)
	router.GET("/classes/:classId/startlist", h.GetStartList)
	router.GET("/classes/:classId/results", h.GetResults)
	router.GET("/classes/:classId/splits", h.GetSplits)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		logger.InfoLogger.Println("Graphics API server starting on :8090...")
		if err := router.Run(":8090"); err != nil {
			logger.ErrorLogger.Printf("Failed to start server: %v", err)
			os.Exit(1)
		}
	}()

	<-sigChan
	logger.InfoLogger.Println("Shutting down...")

	if err := adapter.Stop(); err != nil {
		logger.ErrorLogger.Printf("Error stopping adapter: %v", err)
	}

	logger.InfoLogger.Println("Shutdown complete")
	return nil
}
