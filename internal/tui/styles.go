package tui

import "github.com/charmbracelet/lipgloss"

// Theme colors
var (
	Gold     = lipgloss.Color("#d7af00")
	Red      = lipgloss.Color("#ff5f5f")
	Green    = lipgloss.Color("#87d787")
	Yellow   = lipgloss.Color("#ffd75f")
	Gray     = lipgloss.Color("#6c6c6c")
	White    = lipgloss.Color("#ffffff")
	Black    = lipgloss.Color("#000000")
	DarkGray = lipgloss.Color("#3a3a3a")
)

// Layout styles
var (
	HeaderStyle = lipgloss.NewStyle().
			Background(Gold).
			Foreground(Black).
			Bold(true).
			Padding(0, 1)

	TitleStyle = lipgloss.NewStyle().
			Foreground(Gold).
			Bold(true)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(Gray).
			Faint(true)
)

// Service health indicators
var (
	ServiceHealthy   = lipgloss.NewStyle().Foreground(Green)
	ServiceUnhealthy = lipgloss.NewStyle().Foreground(Red)
	ServiceDisabled  = lipgloss.NewStyle().Foreground(Gray)
)

// Menu styles
var (
	MenuItemStyle = lipgloss.NewStyle()

	MenuSelectedStyle = lipgloss.NewStyle().
				Foreground(Gold).
				Bold(true)
)

// Status bar
var StatusBarStyle = lipgloss.NewStyle().
	Background(DarkGray).
	Foreground(White).
	Padding(0, 1)

// Log line styles
var (
	LogErrorStyle = lipgloss.NewStyle().Foreground(Red)
	LogWarnStyle  = lipgloss.NewStyle().Foreground(Yellow)
	LogSuccessStyle = lipgloss.NewStyle().Foreground(Green)
	LogStepStyle  = lipgloss.NewStyle().Foreground(Gold)
)

// Box / border style
var BorderStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(Gold)

// Progress bar styles
var (
	ProgressFullStyle  = lipgloss.NewStyle().Background(Gold)
	ProgressEmptyStyle = lipgloss.NewStyle().Background(DarkGray)
)
