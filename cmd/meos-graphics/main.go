package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"meos-graphics/internal/handlers"
	"meos-graphics/internal/logger"
	"meos-graphics/internal/meos"
	"meos-graphics/internal/middleware"
	"meos-graphics/internal/simulation"
	"meos-graphics/internal/state"
	"meos-graphics/internal/version"
)

func main() {
	// Parse command line flags
	simulationMode := flag.Bool("simulation", false, "Run in simulation mode")
	showVersion := flag.Bool("version", false, "Show version information")
	pollInterval := flag.Duration("poll-interval", 1*time.Second, "Poll interval for MeOS data updates (e.g., 200ms, 9s, 2m)")
	flag.Parse()

	if *showVersion {
		fmt.Printf("meos-graphics version %s\n", version.Version)
		os.Exit(0)
	}

	// Validate poll interval
	if *pollInterval < 100*time.Millisecond {
		fmt.Printf("Error: poll interval too small (minimum 100ms): %s\n", *pollInterval)
		os.Exit(1)
	}
	if *pollInterval > 1*time.Hour {
		fmt.Printf("Error: poll interval too large (maximum 1 hour): %s\n", *pollInterval)
		os.Exit(1)
	}

	if err := logger.Init(); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	logger.InfoLogger.Printf("Starting MeOS Graphics API Server v%s", version.Version)
	if *simulationMode {
		logger.InfoLogger.Println("Running in SIMULATION MODE")
	}

	// Initialize global state
	appState := state.New()

	// Create adapter based on mode
	var adapter interface {
		Connect() error
		StartPolling() error
		Stop() error
	}

	if *simulationMode {
		// Use simulation adapter
		adapter = simulation.NewAdapter(appState)
	} else {
		// Configure MeOS adapter
		config := meos.NewConfig()
		config.PollInterval = *pollInterval
		logger.InfoLogger.Printf("MeOS Configuration: %s:%d, Poll Interval: %s", config.Hostname, config.Port, config.PollInterval)

		if err := config.Validate(); err != nil {
			logger.ErrorLogger.Printf("Invalid configuration: %v", err)
			os.Exit(1)
		}

		adapter = meos.NewAdapter(config, appState)
	}

	// Connect adapter
	if err := adapter.Connect(); err != nil {
		logger.ErrorLogger.Printf("Failed to connect: %v", err)
		if !*simulationMode {
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
}
