package tui

import (
	"fmt"
	"strings"

	"github.com/ImJustRicky/muster-tui/internal/config"
	"github.com/ImJustRicky/muster-tui/internal/engine"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

type statusLoadedMsg struct {
	Services map[string]engine.ServiceStatus
	Err      error
}

// StatusModel shows service health in a table.
type StatusModel struct {
	engine   *engine.Engine
	config   *config.DeployConfig
	services map[string]engine.ServiceStatus
	loading  bool
	err      error
	width    int
	height   int
}

func NewStatus(eng *engine.Engine, cfg *config.DeployConfig) StatusModel {
	return StatusModel{
		engine:  eng,
		config:  cfg,
		loading: true,
	}
}

func (m StatusModel) Init() tea.Cmd {
	return func() tea.Msg {
		result, err := m.engine.Status()
		if err != nil {
			return statusLoadedMsg{Err: err}
		}
		return statusLoadedMsg{Services: result.Services}
	}
}

func (m StatusModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case statusLoadedMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
		} else {
			m.services = msg.Services
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "r":
			m.loading = true
			m.err = nil
			return m, m.Init()
		case "esc", "q":
			return m, func() tea.Msg { return BackToDashboardMsg{} }
		}
	}
	return m, nil
}

func (m StatusModel) View() string {
	var b strings.Builder

	b.WriteString(HeaderStyle.Render("  muster  "))
	b.WriteString("  ")
	b.WriteString(TitleStyle.Render("Status"))
	b.WriteString("\n\n")

	if m.loading {
		b.WriteString(SubtitleStyle.Render("  Loading..."))
		b.WriteString("\n")
	} else if m.err != nil {
		b.WriteString(LogErrorStyle.Render(fmt.Sprintf("  Error: %s", m.err)))
		b.WriteString("\n")
	} else {
		b.WriteString(m.renderStatusTable())
	}

	b.WriteString("\n")
	b.WriteString(SubtitleStyle.Render("  r refresh • esc back"))

	return b.String()
}

func (m StatusModel) renderStatusTable() string {
	var order []string
	if m.config != nil {
		order = m.config.DeployOrder
	}
	if len(order) == 0 {
		for key := range m.services {
			order = append(order, key)
		}
	}

	rows := make([][]string, 0, len(order))
	for _, key := range order {
		svc, ok := m.services[key]
		if !ok {
			continue
		}

		name := svc.Name
		if name == "" {
			name = key
		}

		var dot string
		switch svc.Status {
		case "healthy":
			dot = "●"
		case "unhealthy", "failed":
			dot = "●"
		default:
			dot = "○"
		}

		status := svc.Status
		if status == "" {
			status = "unknown"
		}

		healthType := svc.HealthType
		if healthType == "" {
			healthType = "-"
		}

		detail := svc.Detail
		if detail == "" {
			detail = "-"
		}

		rows = append(rows, []string{dot + " " + name, status, healthType, detail})
	}

	if len(rows) == 0 {
		return SubtitleStyle.Render("  No services found.")
	}

	maxWidth := m.width - 6
	if maxWidth < 50 {
		maxWidth = 50
	}
	if maxWidth > 90 {
		maxWidth = 90
	}

	t := table.New().
		Headers("SERVICE", "STATUS", "TYPE", "DETAIL").
		Rows(rows...).
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(Gold)).
		Width(maxWidth).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return lipgloss.NewStyle().
					Bold(true).
					Foreground(Gold).
					Padding(0, 1)
			}
			base := lipgloss.NewStyle().Padding(0, 1)

			// Color the status column based on value
			if col == 1 && row >= 0 && row < len(rows) {
				val := rows[row][1]
				switch val {
				case "healthy":
					return base.Foreground(Green).Bold(true)
				case "unhealthy", "failed":
					return base.Foreground(Red).Bold(true)
				default:
					return base.Foreground(Gray)
				}
			}

			// Color the service dot+name
			if col == 0 && row >= 0 && row < len(rows) {
				val := rows[row][1] // check status for color
				switch val {
				case "healthy":
					return base.Foreground(Green)
				case "unhealthy", "failed":
					return base.Foreground(Red)
				default:
					return base.Foreground(Gray)
				}
			}

			if col == 3 {
				return base.Foreground(Gray).Faint(true)
			}

			return base
		})

	return lipgloss.NewStyle().Margin(0, 2).Render(t.Render())
}
