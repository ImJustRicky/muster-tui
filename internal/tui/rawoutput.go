package tui

import (
	"strings"

	"github.com/ImJustRicky/muster-tui/internal/engine"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type rawOutputLineMsg struct {
	Line string
}

type rawOutputDoneMsg struct{}

type rawOutputErrorMsg struct {
	Err error
}

// RawOutputModel streams raw command output in a scrollable viewport.
type RawOutputModel struct {
	engine     *engine.Engine
	title      string
	args       []string
	ch         <-chan string
	viewport   viewport.Model
	lines      []string
	done       bool
	err        error
	autoFollow bool
	ready      bool
	width      int
	height     int
}

func NewRawOutput(eng *engine.Engine, title string, args []string) RawOutputModel {
	return RawOutputModel{
		engine:     eng,
		title:      title,
		args:       args,
		autoFollow: true,
	}
}

func (m RawOutputModel) Init() tea.Cmd {
	return func() tea.Msg {
		ch, err := m.engine.RunRaw(m.args)
		if err != nil {
			return rawOutputErrorMsg{Err: err}
		}
		// Read first line
		line, ok := <-ch
		if !ok {
			return rawOutputDoneMsg{}
		}
		return rawOutputStartedMsg{Line: line, Ch: ch}
	}
}

type rawOutputStartedMsg struct {
	Line string
	Ch   <-chan string
}

func (m RawOutputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		headerHeight := 1
		footerHeight := 1
		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-headerHeight-footerHeight)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - headerHeight - footerHeight
		}
		m.viewport.SetContent(m.renderLines())
		if m.autoFollow {
			m.viewport.GotoBottom()
		}

	case rawOutputErrorMsg:
		m.err = msg.Err
		m.done = true

	case rawOutputStartedMsg:
		m.ch = msg.Ch
		m.lines = append(m.lines, msg.Line)
		if m.ready {
			m.viewport.SetContent(m.renderLines())
			if m.autoFollow {
				m.viewport.GotoBottom()
			}
		}
		return m, waitForRawLine(m.ch)

	case rawOutputLineMsg:
		m.lines = append(m.lines, msg.Line)
		if m.ready {
			m.viewport.SetContent(m.renderLines())
			if m.autoFollow {
				m.viewport.GotoBottom()
			}
		}
		if m.ch != nil {
			return m, waitForRawLine(m.ch)
		}

	case rawOutputDoneMsg:
		m.done = true

	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			if m.done {
				return m, func() tea.Msg { return BackToDashboardMsg{} }
			}
		default:
			if m.ready {
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
		}
	}

	return m, nil
}

func (m RawOutputModel) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	width := m.viewport.Width

	// Header
	titleText := m.title
	var statusText string
	if m.done {
		if m.err != nil {
			statusText = LogErrorStyle.Render(" error ")
		} else {
			statusText = LogSuccessStyle.Render(" done ")
		}
	} else {
		statusText = LogStepStyle.Render(" running... ")
	}
	gap := width - lipgloss.Width(titleText) - lipgloss.Width(statusText) - 2
	if gap < 0 {
		gap = 0
	}
	header := StatusBarStyle.Width(width).Render(
		titleText + strings.Repeat(" ", gap) + statusText,
	)

	// Footer
	var helpText string
	if m.done {
		helpText = "esc back"
	} else {
		helpText = "↑/↓ scroll"
	}
	followIndicator := ""
	if m.autoFollow && !m.done {
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

func (m RawOutputModel) renderLines() string {
	if len(m.lines) == 0 {
		return ""
	}
	rendered := make([]string, len(m.lines))
	for i, line := range m.lines {
		rendered[i] = colorizeLine(line)
	}
	return strings.Join(rendered, "\n")
}

// waitForRawLine reads the next line from a raw output channel.
// The channel reference is captured via closure.
var waitForRawLine = func(ch <-chan string) tea.Cmd {
	return func() tea.Msg {
		line, ok := <-ch
		if !ok {
			return rawOutputDoneMsg{}
		}
		return rawOutputLineMsg{Line: line}
	}
}
