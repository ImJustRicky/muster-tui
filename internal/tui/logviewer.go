package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// LogViewerMsg delivers a new log line to the viewer.
type LogViewerMsg struct {
	Line string
}

// CloseLogViewerMsg is sent when the user presses ctrl+o.
type CloseLogViewerMsg struct{}

// LogViewerModel is a full-screen scrollable log viewer.
type LogViewerModel struct {
	viewport   viewport.Model
	lines      []string
	title      string
	autoFollow bool
	ready      bool
}

// NewLogViewer creates a log viewer with the given title.
func NewLogViewer(title string) LogViewerModel {
	return LogViewerModel{
		title:      title,
		autoFollow: true,
		lines:      []string{},
	}
}

// Init returns nil; no initial command needed.
func (m LogViewerModel) Init() tea.Cmd {
	return nil
}

// Update handles window size changes, new log lines, and key input.
func (m LogViewerModel) Update(msg tea.Msg) (LogViewerModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		headerHeight := 1
		footerHeight := 1
		verticalMargin := headerHeight + footerHeight

		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-verticalMargin)
			m.viewport.SetContent(m.renderLines())
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMargin
		}

		if m.autoFollow {
			m.viewport.GotoBottom()
		}

	case LogViewerMsg:
		m.lines = append(m.lines, msg.Line)
		m.viewport.SetContent(m.renderLines())
		if m.autoFollow {
			m.viewport.GotoBottom()
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+o":
			return m, func() tea.Msg { return CloseLogViewerMsg{} }
		default:
			wasAtBottom := m.viewport.AtBottom()
			m.viewport, cmd = m.viewport.Update(msg)
			if !m.viewport.AtBottom() && wasAtBottom {
				m.autoFollow = false
			}
			if m.viewport.AtBottom() {
				m.autoFollow = true
			}
			return m, cmd
		}

	default:
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}

	return m, nil
}

// View renders the header, viewport, and footer.
func (m LogViewerModel) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	width := m.viewport.Width

	// Header
	titleText := m.title
	countText := fmt.Sprintf(" %d lines ", len(m.lines))
	gap := width - lipgloss.Width(titleText) - lipgloss.Width(countText) - 2 // 2 for padding
	if gap < 0 {
		gap = 0
	}
	header := StatusBarStyle.Width(width).Render(
		titleText + strings.Repeat(" ", gap) + countText,
	)

	// Footer
	helpText := "\u2191/\u2193 scroll \u2022 ctrl+o close"
	followIndicator := ""
	if m.autoFollow {
		followIndicator = " [following]"
	}
	footerGap := width - lipgloss.Width(helpText) - lipgloss.Width(followIndicator) - 2
	if footerGap < 0 {
		footerGap = 0
	}
	footer := StatusBarStyle.Width(width).Render(
		helpText + strings.Repeat(" ", footerGap) + followIndicator,
	)

	return header + "\n" + m.viewport.View() + "\n" + footer
}

// AppendLine adds a line and refreshes the viewport content.
func (m *LogViewerModel) AppendLine(line string) {
	m.lines = append(m.lines, line)
	if m.ready {
		m.viewport.SetContent(m.renderLines())
		if m.autoFollow {
			m.viewport.GotoBottom()
		}
	}
}

// SetSize updates the viewport dimensions.
func (m *LogViewerModel) SetSize(w, h int) {
	headerHeight := 1
	footerHeight := 1
	m.viewport.Width = w
	m.viewport.Height = h - headerHeight - footerHeight
	if !m.ready {
		m.viewport = viewport.New(w, h-headerHeight-footerHeight)
		m.ready = true
	}
	m.viewport.SetContent(m.renderLines())
}

// renderLines applies log coloring to all lines and joins them.
func (m LogViewerModel) renderLines() string {
	if len(m.lines) == 0 {
		return ""
	}

	rendered := make([]string, len(m.lines))
	for i, line := range m.lines {
		rendered[i] = colorizeLine(line)
	}
	return strings.Join(rendered, "\n")
}

// colorizeLine applies the appropriate style based on line content.
func colorizeLine(line string) string {
	lower := strings.ToLower(line)

	switch {
	case strings.HasPrefix(line, "==>") || strings.HasPrefix(line, "---") || strings.HasPrefix(line, "Step"):
		return LogStepStyle.Render(line)
	case strings.Contains(lower, "error"):
		return LogErrorStyle.Render(line)
	case strings.Contains(lower, "warn"):
		return LogWarnStyle.Render(line)
	case strings.Contains(lower, "success") || strings.Contains(line, "\u2713") || strings.Contains(lower, "ok"):
		return LogSuccessStyle.Render(line)
	default:
		return line
	}
}
