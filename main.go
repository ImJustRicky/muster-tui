package main

import (
	"fmt"
	"os"

	"github.com/ImJustRicky/muster-tui/internal/auth"
	"github.com/ImJustRicky/muster-tui/internal/config"
	"github.com/ImJustRicky/muster-tui/internal/engine"
	"github.com/ImJustRicky/muster-tui/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

const version = "0.1.0"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-v":
			fmt.Printf("muster-tui v%s\n", version)
			return
		case "--help", "-h":
			fmt.Println("muster-tui — TUI frontend for muster")
			fmt.Println()
			fmt.Println("Usage: muster-tui [flags]")
			fmt.Println()
			fmt.Println("Flags:")
			fmt.Println("  -v, --version   Show version")
			fmt.Println("  -h, --help      Show this help")
			fmt.Println("  --set-token     Save an auth token for muster")
			return
		case "--set-token":
			if len(os.Args) < 3 {
				fmt.Fprintln(os.Stderr, "Usage: muster-tui --set-token <token>")
				os.Exit(1)
			}
			if err := auth.SaveToken(os.Args[2]); err != nil {
				fmt.Fprintf(os.Stderr, "Error saving token: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Token saved.")
			return
		}
	}

	// Load auth token
	token, err := auth.LoadToken()
	if err != nil || token == "" {
		fmt.Fprintln(os.Stderr, "No auth token configured.")
		fmt.Fprintln(os.Stderr, "Set MUSTER_TOKEN env var or run: muster-tui --set-token <token>")
		fmt.Fprintln(os.Stderr, "Create a token with: muster auth create <name> --scope admin")
		os.Exit(1)
	}

	// Find deploy.json
	configPath, err := config.FindConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "muster.json not found. Run this from a muster project directory.")
		os.Exit(1)
	}

	cfg, err := config.LoadDeploy(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading deploy.json: %v\n", err)
		os.Exit(1)
	}

	// Create engine
	eng, err := engine.NewEngine(token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Launch TUI
	app := tui.NewApp(eng, cfg)
	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
