package tui

import (
	"fmt"
	"strings"

	"github.com/ImJustRicky/muster-tui/internal/engine"
	tea "github.com/charmbracelet/bubbletea"
)

// Internal messages.

type deployEventMsg engine.DeployEvent

type deployDoneMsg struct{}

// DeployModel is the deploy progress view.
type DeployModel struct {
	engine         *engine.Engine
	services       []string
	dryRun         bool
	events         <-chan engine.DeployEvent
	currentService string
	currentName    string
	currentIndex   int
	totalServices  int
	logLines       []string
	allLogLines    []string
	status         string // "running", "success", "failed", "complete"
	width          int
	height         int
	err            error
}

func NewDeploy(eng *engine.Engine, services []string, dryRun bool) DeployModel {
	return DeployModel{
		engine:   eng,
		services: services,
		dryRun:   dryRun,
		status:   "running",
	}
}

func (m DeployModel) Init() tea.Cmd {
	return func() tea.Msg {
		ch, err := m.engine.Deploy(m.services, m.dryRun)
		if err != nil {
			return deployEventMsg(engine.DeployEvent{
				Event:  "error",
				Status: err.Error(),
			})
		}
		// Store the channel and read the first event
		ev, ok := <-ch
		if !ok {
			return deployDoneMsg{}
		}
		// We need to stash the channel — we'll do this via a closure trick:
		// Return a special message that carries the channel.
		return deployStartedMsg{Event: ev, Ch: ch}
	}
}

// deployStartedMsg carries the first event and the channel for subsequent reads.
type deployStartedMsg struct {
	Event engine.DeployEvent
	Ch    <-chan engine.DeployEvent
}

func (m DeployModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case deployStartedMsg:
		m.events = msg.Ch
		return m.processEvent(msg.Event)

	case deployEventMsg:
		return m.processEvent(engine.DeployEvent(msg))

	case deployDoneMsg:
		if m.status == "running" {
			m.status = "complete"
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+o":
			if len(m.allLogLines) > 0 {
				return m, func() tea.Msg {
					return OpenLogViewerMsg{Lines: m.allLogLines, Title: m.currentName}
				}
			}
		default:
			if m.status == "complete" || m.status == "failed" {
				return m, func() tea.Msg { return BackToDashboardMsg{} }
			}
		}
	}

	return m, nil
}

func (m DeployModel) processEvent(ev engine.DeployEvent) (tea.Model, tea.Cmd) {
	switch ev.Event {
	case "error":
		m.err = fmt.Errorf("%s", ev.Status)
		m.status = "failed"
		return m, nil

	case "start":
		m.currentService = ev.Service
		m.currentName = ev.Name
		m.currentIndex = ev.Index
		m.totalServices = ev.Total
		m.logLines = nil

	case "log":
		m.allLogLines = append(m.allLogLines, ev.Line)
		m.logLines = append(m.logLines, ev.Line)
		if len(m.logLines) > 6 {
			m.logLines = m.logLines[len(m.logLines)-6:]
		}

	case "health":
		line := fmt.Sprintf("[health] %s", ev.Status)
		m.allLogLines = append(m.allLogLines, line)
		m.logLines = append(m.logLines, line)
		if len(m.logLines) > 6 {
			m.logLines = m.logLines[len(m.logLines)-6:]
		}

	case "done":
		if ev.Status == "failed" {
			m.status = "failed"
		}

	case "dry_run":
		m.currentService = ev.Service
		m.currentName = ev.Name
		m.currentIndex = ev.Index
		m.totalServices = ev.Total
		line := fmt.Sprintf("[dry run] %s — hook: %s (%d lines)", ev.Name, ev.Hook, ev.HookLines)
		m.allLogLines = append(m.allLogLines, line)
		m.logLines = append(m.logLines, line)
		if len(m.logLines) > 6 {
			m.logLines = m.logLines[len(m.logLines)-6:]
		}

	case "complete":
		m.totalServices = ev.Total
		if m.status == "running" {
			m.status = "complete"
		}
	}

	if m.events != nil {
		return m, waitForDeployEvent(m.events)
	}
	return m, nil
}

func (m DeployModel) View() string {
	var b strings.Builder

	// Header
	if m.dryRun {
		b.WriteString(HeaderStyle.Render("  dry run  "))
	} else {
		b.WriteString(HeaderStyle.Render("  deploying  "))
	}
	b.WriteString("\n\n")

	// Error state
	if m.err != nil {
		b.WriteString(LogErrorStyle.Render(fmt.Sprintf("  Error: %s", m.err)))
		b.WriteString("\n\n")
		b.WriteString(SubtitleStyle.Render("  press any key to return"))
		return b.String()
	}

	// Progress bar
	barWidth := 30
	if m.width > 0 && m.width-10 < barWidth {
		barWidth = m.width - 10
	}
	if barWidth < 10 {
		barWidth = 10
	}
	b.WriteString("  " + renderProgressBar(m.currentIndex, m.totalServices, barWidth))
	b.WriteString("\n\n")

	// Current service
	if m.currentName != "" {
		label := m.currentName
		if m.currentService != "" && m.currentService != m.currentName {
			label = fmt.Sprintf("%s (%s)", m.currentName, m.currentService)
		}

		switch m.status {
		case "failed":
			b.WriteString("  " + ServiceUnhealthy.Render("● "+label) + LogErrorStyle.Render("  failed"))
		case "complete":
			b.WriteString("  " + ServiceHealthy.Render("● "+label) + LogSuccessStyle.Render("  done"))
		default:
			b.WriteString("  " + LogStepStyle.Render("● "+label))
		}
		b.WriteString("\n\n")
	}

	// Log box with last 6 lines
	if len(m.logLines) > 0 {
		logContent := strings.Join(m.logLines, "\n")
		boxWidth := m.width - 6
		if boxWidth < 20 {
			boxWidth = 20
		}
		if boxWidth > 70 {
			boxWidth = 70
		}
		box := BorderStyle.Width(boxWidth).Render(logContent)
		b.WriteString(box)
		b.WriteString("\n")
	}

	// Footer
	b.WriteString("\n")
	switch m.status {
	case "running":
		b.WriteString(SubtitleStyle.Render("  ctrl+o full log"))
	case "complete":
		b.WriteString(LogSuccessStyle.Render("  deploy complete ") + SubtitleStyle.Render("press any key"))
	case "failed":
		b.WriteString(LogErrorStyle.Render("  deploy failed ") + SubtitleStyle.Render("press any key"))
	}

	return b.String()
}

func renderProgressBar(current, total, width int) string {
	if total == 0 {
		return ""
	}
	filled := width * current / total
	if filled > width {
		filled = width
	}
	bar := strings.Repeat("#", filled) + strings.Repeat("-", width-filled)
	return fmt.Sprintf("[%s] %d/%d", bar, current, total)
}

func waitForDeployEvent(ch <-chan engine.DeployEvent) tea.Cmd {
	return func() tea.Msg {
		ev, ok := <-ch
		if !ok {
			return deployDoneMsg{}
		}
		return deployEventMsg(ev)
	}
}
