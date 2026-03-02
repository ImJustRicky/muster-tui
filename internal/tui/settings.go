package tui

import (
	"fmt"
	"strings"

	"github.com/ImJustRicky/muster-tui/internal/config"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type settingDef struct {
	key     string
	label   string
	desc    string
	options []string // for string cycling; nil = numeric
	min     int      // for numeric
	max     int
}

var settingDefs = []settingDef{
	{
		key:     "tui_mode",
		label:   "TUI Mode",
		desc:    "Which interface to launch with `muster`",
		options: []string{"go", "bash"},
	},
	{
		key:     "color_mode",
		label:   "Color Mode",
		desc:    "Terminal color output",
		options: []string{"auto", "always", "never"},
	},
	{
		key:     "log_color_mode",
		label:   "Log Color Mode",
		desc:    "Deploy log coloring",
		options: []string{"auto", "raw", "none"},
	},
	{
		key:     "default_stack",
		label:   "Default Stack",
		desc:    "Default stack type for new projects",
		options: []string{"bare", "docker", "compose", "k8s", "dev"},
	},
	{
		key:     "log_retention_days",
		label:   "Log Retention",
		desc:    "Days to keep deploy logs",
		min:     1,
		max:     90,
	},
	{
		key:     "default_health_timeout",
		label:   "Health Timeout",
		desc:    "Seconds to wait for health checks",
		min:     1,
		max:     120,
	},
	{
		key:     "update_check",
		label:   "Update Check",
		desc:    "Check for updates on launch",
		options: []string{"on", "off"},
	},
}

// SettingsModel is an interactive settings editor.
type SettingsModel struct {
	settings *config.GlobalSettings
	cursor   int
	saved    bool
	err      error
	width    int
	height   int
}

func NewSettings() SettingsModel {
	settings, err := config.LoadGlobal()
	return SettingsModel{
		settings: settings,
		err:      err,
	}
}

func (m SettingsModel) Init() tea.Cmd { return nil }

func (m SettingsModel) getValue(def settingDef) string {
	if m.settings == nil {
		return ""
	}
	switch def.key {
	case "color_mode":
		return m.settings.ColorMode
	case "log_color_mode":
		v := m.settings.LogColorMode
		if v == "" {
			return "auto"
		}
		return v
	case "log_retention_days":
		return fmt.Sprintf("%d", m.settings.LogRetentionDays)
	case "default_stack":
		return m.settings.DefaultStack
	case "default_health_timeout":
		return fmt.Sprintf("%d", m.settings.DefaultHealthTimeout)
	case "update_check":
		return m.settings.UpdateCheck
	case "tui_mode":
		v := m.settings.TUIMode
		if v == "" {
			return "go"
		}
		return v
	}
	return ""
}

func (m *SettingsModel) cycleValue(def settingDef, delta int) {
	if m.settings == nil {
		return
	}
	if len(def.options) > 0 {
		current := m.getValue(def)
		idx := 0
		for i, opt := range def.options {
			if opt == current {
				idx = i
				break
			}
		}
		idx += delta
		if idx < 0 {
			idx = len(def.options) - 1
		}
		if idx >= len(def.options) {
			idx = 0
		}
		m.setStringValue(def.key, def.options[idx])
	} else {
		// Numeric
		current := m.getIntValue(def.key)
		current += delta
		if current < def.min {
			current = def.min
		}
		if current > def.max {
			current = def.max
		}
		m.setIntValue(def.key, current)
	}
	m.save()
}

func (m SettingsModel) getIntValue(key string) int {
	switch key {
	case "log_retention_days":
		return m.settings.LogRetentionDays
	case "default_health_timeout":
		return m.settings.DefaultHealthTimeout
	}
	return 0
}

func (m *SettingsModel) setIntValue(key string, val int) {
	switch key {
	case "log_retention_days":
		m.settings.LogRetentionDays = val
	case "default_health_timeout":
		m.settings.DefaultHealthTimeout = val
	}
}

func (m *SettingsModel) setStringValue(key, val string) {
	switch key {
	case "color_mode":
		m.settings.ColorMode = val
	case "log_color_mode":
		m.settings.LogColorMode = val
	case "default_stack":
		m.settings.DefaultStack = val
	case "update_check":
		m.settings.UpdateCheck = val
	case "tui_mode":
		m.settings.TUIMode = val
	}
}

func (m *SettingsModel) save() {
	if m.settings == nil {
		return
	}
	if err := config.SaveGlobal(m.settings); err != nil {
		m.err = err
		m.saved = false
	} else {
		m.err = nil
		m.saved = true
	}
}

func (m SettingsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			return m, func() tea.Msg { return BackToDashboardMsg{} }
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				m.saved = false
			}
		case "down", "j":
			if m.cursor < len(settingDefs)-1 {
				m.cursor++
				m.saved = false
			}
		case "enter", " ", "right", "l", "tab":
			m.cycleValue(settingDefs[m.cursor], 1)
		case "left", "h", "shift+tab":
			m.cycleValue(settingDefs[m.cursor], -1)
		}
	}
	return m, nil
}

func (m SettingsModel) View() string {
	var b strings.Builder

	// Header
	b.WriteString(HeaderStyle.Render("  muster  "))
	b.WriteString("  ")
	b.WriteString(TitleStyle.Render("Settings"))
	b.WriteString("\n\n")

	if m.err != nil {
		b.WriteString(LogErrorStyle.Render(fmt.Sprintf("  Error: %s", m.err)))
		b.WriteString("\n\n")
	}

	if m.settings == nil {
		b.WriteString(SubtitleStyle.Render("  No settings available."))
		return b.String()
	}

	// Calculate widths
	boxWidth := m.width - 8
	if boxWidth < 50 {
		boxWidth = 50
	}
	if boxWidth > 90 {
		boxWidth = 90
	}

	labelWidth := 22
	valueWidth := 18

	// Render setting rows
	var rows strings.Builder
	for i, def := range settingDefs {
		current := m.getValue(def)
		isSelected := i == m.cursor

		// Label
		var labelStyle lipgloss.Style
		if isSelected {
			labelStyle = lipgloss.NewStyle().Foreground(Gold).Bold(true)
		} else {
			labelStyle = lipgloss.NewStyle().Foreground(White)
		}

		// Value with selector
		var valueStr string
		if len(def.options) > 0 {
			valueStr = m.renderOptionSelector(def, current, isSelected)
		} else {
			valueStr = m.renderNumericSelector(def, current, isSelected)
		}

		// Description
		descStyle := lipgloss.NewStyle().Foreground(Gray).Faint(true)

		// Cursor
		cursor := "   "
		if isSelected {
			cursor = lipgloss.NewStyle().Foreground(Gold).Bold(true).Render(" ▸ ")
		}

		label := labelStyle.Render(fmt.Sprintf("%-*s", labelWidth, def.label))
		value := fmt.Sprintf("%-*s", valueWidth, valueStr)

		row := cursor + label + "  " + value + "  " + descStyle.Render(def.desc)
		rows.WriteString(row + "\n")

		// Add spacing between rows
		if i < len(settingDefs)-1 {
			rows.WriteString("\n")
		}
	}

	box := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(Gold).
		Padding(1, 2).
		Width(boxWidth).
		Render(rows.String())

	b.WriteString(lipgloss.NewStyle().Margin(0, 3).Render(box))
	b.WriteString("\n\n")

	// Saved indicator
	if m.saved {
		b.WriteString(lipgloss.NewStyle().Margin(0, 3).Render(
			LogSuccessStyle.Render("  ✓ Saved"),
		))
		b.WriteString("\n\n")
	}

	// Footer
	footer := SubtitleStyle.Render("  ↑/↓ navigate • ←/→ change value • esc back")
	b.WriteString(lipgloss.NewStyle().Margin(0, 3).Render(footer))

	return b.String()
}

func (m SettingsModel) renderOptionSelector(def settingDef, current string, selected bool) string {
	var parts []string
	for _, opt := range def.options {
		if opt == current {
			style := lipgloss.NewStyle().Foreground(Green).Bold(true)
			if selected {
				style = style.Background(lipgloss.Color("#2a3a2a"))
			}
			parts = append(parts, style.Render(" "+opt+" "))
		} else {
			style := lipgloss.NewStyle().Foreground(Gray).Faint(true)
			parts = append(parts, style.Render(" "+opt+" "))
		}
	}
	return strings.Join(parts, SubtitleStyle.Render("·"))
}

func (m SettingsModel) renderNumericSelector(def settingDef, current string, selected bool) string {
	var style lipgloss.Style
	if selected {
		style = lipgloss.NewStyle().Foreground(Green).Bold(true).Background(lipgloss.Color("#2a3a2a"))
	} else {
		style = lipgloss.NewStyle().Foreground(Green).Bold(true)
	}

	arrows := SubtitleStyle.Render("◂ ") + style.Render(current) + SubtitleStyle.Render(" ▸")
	if !selected {
		arrows = "  " + style.Render(current) + "  "
	}
	return arrows
}
