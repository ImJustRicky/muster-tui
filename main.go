package main

import (
	"fmt"
	"os"

	"github.com/ImJustRicky/muster-tui/internal/auth"
	"github.com/ImJustRicky/muster-tui/internal/config"
	"github.com/ImJustRicky/muster-tui/internal/engine"
	"github.com/ImJustRicky/muster-tui/internal/registry"
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

	// Create engine
	eng, err := engine.NewEngine(token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Try to find config in current directory first
	configPath, configErr := config.FindConfig()

	if configErr == nil {
		// We're inside a muster project — go straight to dashboard
		cfg, err := config.LoadDeploy(configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}

		app := tui.NewApp(eng, cfg)
		p := tea.NewProgram(app, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Not in a project directory — check the registry for known projects
	pf, err := registry.Load()
	if err != nil || len(pf.Projects) == 0 {
		fmt.Fprintln(os.Stderr, "No muster.json found in current directory and no projects registered.")
		fmt.Fprintln(os.Stderr, "Run 'muster setup' in a project directory first.")
		os.Exit(1)
	}

	// Single project — go straight to it
	if len(pf.Projects) == 1 {
		p := pf.Projects[0]
		cfgPath := config.FindConfigIn(p.Path)
		if cfgPath == "" {
			fmt.Fprintf(os.Stderr, "Registered project %s no longer has a config file.\n", p.Path)
			fmt.Fprintln(os.Stderr, "Run 'muster projects --prune' to clean up.")
			os.Exit(1)
		}
		cfg, err := config.LoadDeploy(cfgPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}

		app := tui.NewApp(eng, cfg)
		prog := tea.NewProgram(app, tea.WithAltScreen())
		if _, err := prog.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Multiple projects — show project picker
	app := tui.NewAppWithPicker(eng, pf.Projects)
	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
