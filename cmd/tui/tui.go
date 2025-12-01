package tui

import (
	"fmt"
	"os"

	"github.com/pmaojo/goploy/internal/config"
	"github.com/pmaojo/goploy/internal/tui"
	"github.com/spf13/cobra"
)

// New creates and returns the `tui` subcommand.
// It initializes the Terminal User Interface application by parsing the configuration
// and running the TUI loop.
func New() *cobra.Command {
	return &cobra.Command{
		Use:   "tui",
		Short: "Starts the Terminal User Interface",
		Long: `Starts the interactive Terminal User Interface (TUI).

Requires a goploy.yaml configuration file in the current directory.`,
		Run: func(_ *cobra.Command, _ []string) {
			runTui()
		},
	}
}

func runTui() {
	// 1. Read configuration
	// We assume goploy.yaml is in the current directory for now
	data, err := os.ReadFile("goploy.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading goploy.yaml: %v\n", err)
		os.Exit(1)
	}

	cfg, err := config.ParseGoployConfig(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing goploy.yaml: %v\n", err)
		os.Exit(1)
	}

	// 2. Initialize and Run TUI
	app := tui.NewApp(cfg)
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running application: %v\n", err)
		os.Exit(1)
	}
}
