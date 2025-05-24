package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"meos-graphics/internal/cmd"
	"meos-graphics/internal/handlers"
	"meos-graphics/internal/logger"
	"meos-graphics/internal/meos"
	"meos-graphics/internal/middleware"
	"meos-graphics/internal/service"
	"meos-graphics/internal/simulation"
	"meos-graphics/internal/sse"
	"meos-graphics/internal/state"
	"meos-graphics/internal/version"
	"meos-graphics/internal/web"

	_ "meos-graphics/docs" // Import generated swagger docs
)

// @title MeOS Graphics API
// @version 0.0.0 x-release-please-version
// @description REST API for accessing orienteering competition data from MeOS
// @termsOfService http://swagger.io/terms/

// @contact.name MeOS Graphics Support
// @contact.url https://github.com/MetsaApp/meos-graphics
// @contact.email support@example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8090
// @BasePath /
// @schemes http https

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

	// Set up SSE hub
	sseHub := sse.NewHub()
	go sseHub.Run()

	// Create service layer
	svc := service.New(appState)

	// Set up state change notifications
	appState.OnUpdate(func() {
		sseHub.BroadcastUpdate("update", gin.H{"timestamp": time.Now().Unix()})
	})

	// Set up HTTP server
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.Logger())

	// Load HTML templates
	router.SetHTMLTemplate(web.GetTemplates())

	// Create handlers
	h := handlers.New(appState)
	webHandler := web.New(svc)

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":         "ok",
			"meos_connected": true,
			"sse_clients":    sseHub.GetConnectedClients(),
		})
	})

	// API endpoints (REST)
	api := router.Group("/")
	api.GET("/classes", h.GetClasses)
	api.GET("/classes/:classId/startlist", h.GetStartList)
	api.GET("/classes/:classId/results", h.GetResults)
	api.GET("/classes/:classId/splits", h.GetSplits)

	// Web interface endpoints
	webGroup := router.Group("/web")
	webGroup.GET("/", webHandler.HomePage)
	webGroup.GET("/classes/:classId", webHandler.ClassPage)
	webGroup.GET("/classes/:classId/startlist", webHandler.StartListPartial)
	webGroup.GET("/classes/:classId/results", webHandler.ResultsPartial)
	webGroup.GET("/classes/:classId/splits", webHandler.SplitsPartial)

	// SSE endpoint
	router.GET("/sse", sseHub.HandleSSE)

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API documentation redirect
	router.GET("/docs", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})

	// Redirect root to web interface
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/web")
	})

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
