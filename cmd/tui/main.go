package main

import (
	"fmt"
	"os"

	"allaboutapps.dev/aw/go-starter/internal/config"
	"allaboutapps.dev/aw/go-starter/internal/tui"
)

func main() {
	// 1. Read configuration (FR1 - partial)
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

	// 2. Initialize and Run TUI (FR2)
	app := tui.NewApp(cfg)
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running application: %v\n", err)
		os.Exit(1)
	}
}
