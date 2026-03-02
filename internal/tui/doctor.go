package tui

import (
	"fmt"
	"strings"

	"github.com/ImJustRicky/muster-tui/internal/engine"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type doctorLoadedMsg struct {
	Result *engine.DoctorResult
	Err    error
}

// DoctorModel shows diagnostic check results.
type DoctorModel struct {
	engine  *engine.Engine
	result  *engine.DoctorResult
	loading bool
	err     error
	width   int
	height  int
}

func NewDoctor(eng *engine.Engine) DoctorModel {
	return DoctorModel{
		engine:  eng,
		loading: true,
	}
}

func (m DoctorModel) Init() tea.Cmd {
	return func() tea.Msg {
		result, err := m.engine.Doctor(false)
		return doctorLoadedMsg{Result: result, Err: err}
	}
}

func (m DoctorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case doctorLoadedMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
		} else {
			m.result = msg.Result
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "r":
			m.loading = true
			m.err = nil
			m.result = nil
			return m, m.Init()
		case "esc", "q":
			return m, func() tea.Msg { return BackToDashboardMsg{} }
		}
	}
	return m, nil
}

func (m DoctorModel) View() string {
	var b strings.Builder

	b.WriteString(HeaderStyle.Render("  muster  "))
	b.WriteString("  ")
	b.WriteString(TitleStyle.Render("Doctor"))
	b.WriteString("\n\n")

	if m.loading {
		b.WriteString(SubtitleStyle.Render("  Running diagnostics..."))
		b.WriteString("\n")
	} else if m.err != nil {
		b.WriteString(LogErrorStyle.Render(fmt.Sprintf("  Error: %s", m.err)))
		b.WriteString("\n")
	} else if m.result != nil {
		b.WriteString(m.renderChecks())
	}

	b.WriteString("\n")
	b.WriteString(SubtitleStyle.Render("  r re-run • esc back"))

	return b.String()
}

func (m DoctorModel) renderChecks() string {
	var b strings.Builder

	maxNameLen := 0
	for _, check := range m.result.Checks {
		if len(check.Name) > maxNameLen {
			maxNameLen = len(check.Name)
		}
	}

	boxWidth := maxNameLen + 30
	if boxWidth > m.width-8 {
		boxWidth = m.width - 8
	}
	if boxWidth < 40 {
		boxWidth = 40
	}

	var checks strings.Builder
	for _, check := range m.result.Checks {
		var icon string
		var nameStyle lipgloss.Style
		switch check.Status {
		case "pass":
			icon = LogSuccessStyle.Render("  ✓")
			nameStyle = lipgloss.NewStyle().Foreground(White)
		case "warn":
			icon = LogWarnStyle.Render("  !")
			nameStyle = lipgloss.NewStyle().Foreground(Yellow)
		case "fail":
			icon = LogErrorStyle.Render("  ✗")
			nameStyle = lipgloss.NewStyle().Foreground(Red)
		default:
			icon = SubtitleStyle.Render("  ?")
			nameStyle = lipgloss.NewStyle().Foreground(Gray)
		}

		line := fmt.Sprintf("%s  %s", icon, nameStyle.Render(check.Name))
		if check.Detail != "" {
			line += SubtitleStyle.Render("  " + check.Detail)
		}
		checks.WriteString(line + "\n")
	}

	box := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(Gold).
		Padding(1, 1).
		Width(boxWidth).
		Render(checks.String())

	b.WriteString(lipgloss.NewStyle().Margin(0, 2).Render(box))
	b.WriteString("\n\n")

	// Summary bar
	summary := lipgloss.NewStyle().Margin(0, 2)
	parts := []string{
		LogSuccessStyle.Render(fmt.Sprintf("  ✓ %d passed", m.result.Pass)),
	}
	if m.result.Warnings > 0 {
		parts = append(parts, LogWarnStyle.Render(fmt.Sprintf("  ! %d warnings", m.result.Warnings)))
	}
	if m.result.Failures > 0 {
		parts = append(parts, LogErrorStyle.Render(fmt.Sprintf("  ✗ %d failures", m.result.Failures)))
	}
	b.WriteString(summary.Render(strings.Join(parts, "    ")))

	return b.String()
}
