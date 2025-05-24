package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"meos-graphics/internal/cmd"
)

func main() {
	rootCmd := cmd.NewRootCommand()

	// Create docs directory
	docsDir := "docs"
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating docs directory: %v\n", err)
		os.Exit(1)
	}

	// Generate custom documentation
	customContent := generateDocumentation(rootCmd)

	// Write to CLI_FLAGS.md
	outputPath := filepath.Join(docsDir, "CLI_FLAGS.md")
	if err := os.WriteFile(outputPath, []byte(customContent), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing CLI_FLAGS.md: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Documentation generated: %s\n", outputPath)
}

func generateDocumentation(cmd *cobra.Command) string {
	var sb strings.Builder

	sb.WriteString("# MeOS Graphics CLI Flags\n\n")
	sb.WriteString(cmd.Long)
	sb.WriteString("\n\n")

	sb.WriteString("## Usage\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString("meos-graphics [flags]\n")
	sb.WriteString("```\n\n")

	sb.WriteString("## Available Flags\n\n")

	// Get the help output
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	_ = cmd.Usage() // Ignore error as Usage() always returns nil
	helpOutput := buf.String()

	// Parse flags from help output
	lines := strings.Split(helpOutput, "\n")
	inFlags := false
	for _, line := range lines {
		if strings.Contains(line, "Flags:") {
			inFlags = true
			continue
		}
		if inFlags && strings.TrimSpace(line) == "" {
			inFlags = false
			continue
		}
		if inFlags && strings.HasPrefix(line, "  ") {
			// Parse flag line
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "-") {
				parts := strings.SplitN(line, "   ", 2)
				if len(parts) >= 2 {
					flagDef := strings.TrimSpace(parts[0])
					description := strings.TrimSpace(parts[1])

					// Skip help and version flags
					if strings.Contains(flagDef, "help") || strings.Contains(flagDef, "version") {
						continue
					}

					// Extract flag name
					flagName := ""
					flagType := ""
					fields := strings.Fields(flagDef)
					for _, field := range fields {
						if strings.HasPrefix(field, "--") {
							flagName = strings.TrimPrefix(field, "--")
						} else if field != "-h," && !strings.HasPrefix(field, "-") {
							flagType = field
						}
					}

					// Extract default value
					defaultValue := ""
					if strings.Contains(description, "(default ") {
						start := strings.Index(description, "(default ")
						end := strings.Index(description[start:], ")")
						if end > 0 {
							defaultValue = description[start+9 : start+end]
							description = strings.TrimSpace(description[:start])
						}
					}

					if flagName != "" {
						sb.WriteString(fmt.Sprintf("### --%s\n\n", flagName))
						if flagType != "" {
							sb.WriteString(fmt.Sprintf("- **Type**: %s\n", flagType))
						}
						if defaultValue != "" {
							sb.WriteString(fmt.Sprintf("- **Default**: %s\n", defaultValue))
						}
						sb.WriteString(fmt.Sprintf("- **Description**: %s\n\n", description))
					}
				}
			}
		}
	}

	// Add examples
	sb.WriteString("## Examples\n\n")
	sb.WriteString("### Run in simulation mode\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString("meos-graphics --simulation\n")
	sb.WriteString("```\n\n")

	sb.WriteString("### Connect to custom MeOS server\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString("meos-graphics --meos-host=10.0.0.5 --meos-port=3000\n")
	sb.WriteString("```\n\n")

	sb.WriteString("### Use faster poll interval\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString("meos-graphics --poll-interval=200ms\n")
	sb.WriteString("```\n\n")

	sb.WriteString("### Connect to MeOS server without specifying port\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString("meos-graphics --meos-host=meos.example.com --meos-port=none\n")
	sb.WriteString("```\n\n")

	sb.WriteString("### Show version information\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString("meos-graphics --version\n")
	sb.WriteString("```\n\n")

	sb.WriteString("## Environment Variables\n\n")
	sb.WriteString("Currently, no environment variables are supported. All configuration is done through command-line flags.\n\n")

	sb.WriteString("## Notes\n\n")
	sb.WriteString("- The `--poll-interval` flag accepts Go duration strings (e.g., \"200ms\", \"1s\", \"2m\", \"1h\")\n")
	sb.WriteString("- When using `--meos-port=none`, the port is omitted from the MeOS server URL\n")
	sb.WriteString("- In simulation mode, the application generates test data without connecting to a real MeOS server\n")

	return sb.String()
}
