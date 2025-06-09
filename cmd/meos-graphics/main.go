package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
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

// @title meos-graphics
// @version 1.2.0 // x-release-please-version
// @description REST API for accessing orienteering competition data from MeOS
// @termsOfService http://swagger.io/terms/

// @contact.name @malpou
// @contact.url https://github.com/MetsaApp/meos-graphics
// @contact.email malthe@grundtvigsvej.dk

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

		logger.InfoLogger.Printf("Simulation timing: Total=%s, Start=%s, Running=%s, Results=%s, MassStart=%v",
			cmd.SimulationDuration, cmd.SimulationPhaseStart, cmd.SimulationPhaseRunning, cmd.SimulationPhaseResults,
			cmd.SimulationMassStart)
	}

	// Initialize global state
	appState := state.New()

	// Create adapter based on mode
	var adapter interface {
		Connect() error
		StartPolling() error
		Stop() error
	}

	var simulationAdapter *simulation.Adapter

	if cmd.SimulationMode {
		// Use simulation adapter with timing configuration
		simulationAdapter = simulation.NewAdapter(appState, cmd.SimulationDuration,
			cmd.SimulationPhaseStart, cmd.SimulationPhaseRunning, cmd.SimulationPhaseResults,
			cmd.SimulationMassStart)
		adapter = simulationAdapter
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

	// Serve static files from filesystem
	staticPath := getStaticPath()
	logger.InfoLogger.Printf("Serving static files from: %s", staticPath)

	// Verify the path exists and list contents for debugging
	if info, err := os.Stat(staticPath); err == nil && info.IsDir() {
		if entries, readErr := os.ReadDir(staticPath); readErr == nil {
			logger.DebugLogger.Printf("Static directory contents:")
			for _, entry := range entries {
				logger.DebugLogger.Printf("  - %s (dir: %v)", entry.Name(), entry.IsDir())
			}
		}
	} else {
		logger.ErrorLogger.Printf("Static path error: %v", err)
	}

	router.Static("/static", staticPath)

	// Create handlers
	h := handlers.New(appState)
	webHandler := web.New(svc, cmd.SimulationMode)

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		response := gin.H{
			"status":         "ok",
			"meos_connected": true,
			"sse_clients":    sseHub.GetConnectedClients(),
		}

		// Add simulation status if in simulation mode
		if simulationAdapter != nil {
			phase, nextPhaseIn, _ := simulationAdapter.GetSimulationStatus()
			response["simulation"] = gin.H{
				"phase":       phase,
				"nextPhaseIn": nextPhaseIn.Seconds(),
			}
		}

		c.JSON(200, response)
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

	// Simulation status endpoint (for web UI)
	router.GET("/simulation/status", func(c *gin.Context) {
		// Always return 200 since the web UI polls this endpoint
		// but only include simulation data when in simulation mode
		if simulationAdapter != nil {
			phase, nextPhaseIn, _ := simulationAdapter.GetSimulationStatus()
			c.JSON(200, gin.H{
				"enabled":     true,
				"phase":       phase,
				"nextPhaseIn": nextPhaseIn.Seconds(),
			})
		} else {
			c.JSON(200, gin.H{
				"enabled": false,
			})
		}
	})

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

	// Serve empty favicon to avoid 404 errors
	router.GET("/favicon.ico", func(c *gin.Context) {
		c.Data(http.StatusOK, "image/x-icon", []byte{})
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

// getStaticPath returns the path to the static files directory.
// It works correctly whether running with 'go run' from any directory
// or from a compiled binary.
func getStaticPath() string {
	// When using 'go run', we need to check relative to working directory first
	cwd, err := os.Getwd()
	if err == nil {
		possiblePaths := []string{
			filepath.Join(cwd, "web", "static"),             // Running from project root
			filepath.Join(cwd, "..", "..", "web", "static"), // Running from cmd/meos-graphics
			filepath.Join(cwd, "..", "web", "static"),       // Running from cmd/
		}

		for _, path := range possiblePaths {
			absPath, _ := filepath.Abs(path)
			if _, statErr := os.Stat(absPath); statErr == nil {
				return absPath
			}
		}
	}

	// Try to find the path relative to the source file location
	_, filename, _, ok := runtime.Caller(0)
	if ok {
		// Get the directory of this source file
		dir := filepath.Dir(filename)
		// Navigate from cmd/meos-graphics to web/static
		staticPath := filepath.Join(dir, "..", "..", "web", "static")
		if _, statErr := os.Stat(staticPath); statErr == nil {
			return staticPath
		}
	}

	// If source-based path doesn't work, try relative to the executable
	execPath, err := os.Executable()
	if err == nil {
		execDir := filepath.Dir(execPath)

		// Check common relative paths from executable location
		possiblePaths := []string{
			filepath.Join(execDir, "web", "static"),             // Binary in project root
			filepath.Join(execDir, "..", "..", "web", "static"), // Binary in bin/ or cmd/meos-graphics
			filepath.Join(execDir, "..", "web", "static"),       // Binary in bin/
		}

		for _, path := range possiblePaths {
			if _, statErr := os.Stat(path); statErr == nil {
				return path
			}
		}
	}

	// Default fallback
	return "./web/static"
}
