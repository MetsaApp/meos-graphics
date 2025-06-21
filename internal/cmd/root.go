package cmd

import (
	"time"

	"github.com/spf13/cobra"
	"meos-graphics/internal/version"
)

var (
	SimulationMode bool
	PollInterval   time.Duration
	MeosHost       string
	MeosPort       string
	SwaggerHost    string
	Language       string

	// Simulation timing configuration
	SimulationDuration     time.Duration
	SimulationPhaseStart   time.Duration
	SimulationPhaseRunning time.Duration
	SimulationPhaseResults time.Duration
	SimulationMassStart    bool

	// Simulation content configuration
	SimulationNumClasses      int
	SimulationRunnersPerClass int
	SimulationRadioControls   int
)

// NewRootCommand creates and returns the root cobra command
func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "meos-graphics",
		Short: "MeOS Graphics API Server",
		Long: `MeOS Graphics API Server connects to MeOS (orienteering event software) 
and provides competition data for graphics displays.

The server can run in two modes:
- Normal mode: Connects to a real MeOS server
- Simulation mode: Generates test data for development`,
		Version: version.Version,
	}

	rootCmd.Flags().BoolVar(&SimulationMode, "simulation", false, "Run in simulation mode")
	rootCmd.Flags().DurationVar(&PollInterval, "poll-interval", 1*time.Second, "Poll interval for MeOS data updates (e.g., 200ms, 9s, 2m)")
	rootCmd.Flags().StringVar(&MeosHost, "meos-host", "localhost", "MeOS server hostname or IP address")
	rootCmd.Flags().StringVar(&MeosPort, "meos-port", "2009", "MeOS server port (use 'none' to omit port from URL)")
	rootCmd.Flags().StringVar(&SwaggerHost, "swagger-host", "localhost:8090", "Hostname for Swagger documentation API calls")
	rootCmd.Flags().StringVarP(&Language, "language", "l", "en", "Language for status display (en=English, da=Danish)")

	// Simulation timing flags
	rootCmd.Flags().DurationVar(&SimulationDuration, "simulation-duration", 15*time.Minute, "Total simulation cycle duration (only with --simulation)")
	rootCmd.Flags().DurationVar(&SimulationPhaseStart, "simulation-phase-start", 3*time.Minute, "Duration of start list phase (only with --simulation)")
	rootCmd.Flags().DurationVar(&SimulationPhaseRunning, "simulation-phase-running", 7*time.Minute, "Duration of running phase (only with --simulation)")
	rootCmd.Flags().DurationVar(&SimulationPhaseResults, "simulation-phase-results", 5*time.Minute, "Duration of results phase (only with --simulation)")
	rootCmd.Flags().BoolVar(&SimulationMassStart, "simulation-mass-start", false, "Use mass start instead of staggered starts (only with --simulation)")

	// Simulation content flags
	rootCmd.Flags().IntVar(&SimulationNumClasses, "simulation-classes", 3, "Number of competition classes to generate (only with --simulation)")
	rootCmd.Flags().IntVar(&SimulationRunnersPerClass, "simulation-runners", 20, "Number of competitors per class (only with --simulation)")
	rootCmd.Flags().IntVar(&SimulationRadioControls, "simulation-controls", 3, "Number of radio controls per class (only with --simulation)")

	return rootCmd
}
