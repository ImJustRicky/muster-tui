package tui

import (
	"fmt"
	"strings"

	"github.com/ImJustRicky/muster-tui/internal/engine"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

type historyLoadedMsg struct {
	Events []engine.HistoryEvent
	Err    error
}

// HistoryModel shows deploy/rollback history in a scrollable viewport.
type HistoryModel struct {
	engine   *engine.Engine
	events   []engine.HistoryEvent
	viewport viewport.Model
	loading  bool
	showAll  bool
	err      error
	ready    bool
	width    int
	height   int
}

func NewHistory(eng *engine.Engine) HistoryModel {
	return HistoryModel{
		engine:  eng,
		loading: true,
	}
}

func (m HistoryModel) Init() tea.Cmd {
	return func() tea.Msg {
		events, err := m.engine.History(m.showAll, "")
		return historyLoadedMsg{Events: events, Err: err}
	}
}

func (m HistoryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		headerHeight := 3
		footerHeight := 2
		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-headerHeight-footerHeight)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - headerHeight - footerHeight
		}
		if len(m.events) > 0 {
			m.viewport.SetContent(m.renderEvents())
		}

	case historyLoadedMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
		} else {
			m.events = msg.Events
			if m.ready {
				m.viewport.SetContent(m.renderEvents())
				m.viewport.GotoTop()
			}
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "a":
			m.showAll = !m.showAll
			m.loading = true
			return m, m.Init()
		case "esc", "q":
			return m, func() tea.Msg { return BackToDashboardMsg{} }
		default:
			if m.ready {
				m.viewport, cmd = m.viewport.Update(msg)
				return m, cmd
			}
		}
	}

	return m, nil
}

func (m HistoryModel) View() string {
	var b strings.Builder

	// Header
	b.WriteString(HeaderStyle.Render("  muster  "))
	b.WriteString("  ")
	b.WriteString(TitleStyle.Render("History"))
	if m.showAll {
		b.WriteString("  " + SubtitleStyle.Render("(all)"))
	}
	b.WriteString("\n\n")

	if m.loading {
		b.WriteString(SubtitleStyle.Render("  Loading..."))
	} else if m.err != nil {
		b.WriteString(LogErrorStyle.Render(fmt.Sprintf("  Error: %s", m.err)))
	} else if !m.ready {
		b.WriteString(SubtitleStyle.Render("  Initializing..."))
	} else if len(m.events) == 0 {
		b.WriteString(SubtitleStyle.Render("  No history yet."))
	} else {
		b.WriteString(m.viewport.View())
	}

	b.WriteString("\n\n")

	hint := "a toggle all"
	if m.showAll {
		hint = "a toggle recent"
	}
	b.WriteString(SubtitleStyle.Render(fmt.Sprintf("  %s • ↑/↓ scroll • esc back", hint)))

	return b.String()
}

func (m HistoryModel) renderEvents() string {
	rows := make([][]string, len(m.events))
	for i, ev := range m.events {
		commit := ev.Commit
		if len(commit) > 7 {
			commit = commit[:7]
		}
		if commit == "" {
			commit = "-"
		}
		rows[i] = []string{ev.Timestamp, ev.Service, ev.Action, ev.Status, commit}
	}

	maxWidth := m.width - 4
	if maxWidth < 60 {
		maxWidth = 60
	}
	if maxWidth > 100 {
		maxWidth = 100
	}

	t := table.New().
		Headers("TIMESTAMP", "SERVICE", "ACTION", "STATUS", "COMMIT").
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

			// Color the status column
			if col == 3 && row >= 0 && row < len(rows) {
				val := rows[row][3]
				switch val {
				case "ok", "success":
					return base.Foreground(Green).Bold(true)
				case "failed":
					return base.Foreground(Red).Bold(true)
				default:
					return base.Foreground(Gray)
				}
			}

			// Timestamp and commit are dimmed
			if col == 0 || col == 4 {
				return base.Foreground(Gray).Faint(true)
			}

			return base
		})

	return lipgloss.NewStyle().Margin(0, 1).Render(t.Render())
}
