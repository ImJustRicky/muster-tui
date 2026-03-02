package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/ImJustRicky/muster-tui/internal/config"
	"github.com/ImJustRicky/muster-tui/internal/engine"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Navigation messages emitted by the dashboard.

type StartDeployMsg struct {
	Services []string
}

type ShowHistoryMsg struct{}

type ShowDoctorMsg struct{}

// Internal messages.

type dashboardRefreshMsg struct {
	Services map[string]engine.ServiceStatus
	Err      error
}

type dashboardTickMsg struct{}

// DashboardModel is the main dashboard view.
type DashboardModel struct {
	engine      *engine.Engine
	config      *config.DeployConfig
	services    map[string]engine.ServiceStatus
	menu        MenuModel
	width       int
	height      int
	loading     bool
	err         error
	lastRefresh time.Time
}

func NewDashboard(eng *engine.Engine, cfg *config.DeployConfig) DashboardModel {
	items := []MenuItem{
		{Label: "Deploy All", Value: "deploy_all"},
		{Label: "Deploy Service...", Value: "deploy_select"},
		{Label: "History", Value: "history"},
		{Label: "Doctor", Value: "doctor"},
		{Label: "Quit", Value: "quit"},
	}

	return DashboardModel{
		engine:  eng,
		config:  cfg,
		menu:    NewMenu("Actions", items),
		loading: true,
	}
}

func (m DashboardModel) Init() tea.Cmd {
	return tea.Batch(refreshStatus(m.engine), dashboardTick())
}

func (m DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case dashboardRefreshMsg:
		m.loading = false
		m.lastRefresh = time.Now()
		if msg.Err != nil {
			m.err = msg.Err
		} else {
			m.err = nil
			m.services = msg.Services
		}

	case dashboardTickMsg:
		return m, tea.Batch(refreshStatus(m.engine), dashboardTick())

	case MenuSelectMsg:
		switch msg.Value {
		case "deploy_all":
			return m, func() tea.Msg { return StartDeployMsg{} }
		case "deploy_select":
			// For now emit deploy all; service picker can be added later
			return m, func() tea.Msg { return StartDeployMsg{} }
		case "history":
			return m, func() tea.Msg { return ShowHistoryMsg{} }
		case "doctor":
			return m, func() tea.Msg { return ShowDoctorMsg{} }
		case "quit":
			return m, tea.Quit
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "r":
			m.loading = true
			return m, refreshStatus(m.engine)
		default:
			var cmd tea.Cmd
			m.menu, cmd = m.menu.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func (m DashboardModel) View() string {
	var b strings.Builder

	// Header bar
	header := HeaderStyle.Render("  muster  ")
	if m.config != nil && m.config.Project != "" {
		header += "  " + TitleStyle.Render(m.config.Project)
	}
	b.WriteString(header)
	b.WriteString("\n\n")

	// Services panel
	if m.loading && m.services == nil {
		b.WriteString(SubtitleStyle.Render("  Loading services..."))
		b.WriteString("\n")
	} else if m.err != nil {
		b.WriteString(LogErrorStyle.Render(fmt.Sprintf("  Error: %s", m.err)))
		b.WriteString("\n")
	} else {
		b.WriteString(m.renderServices())
	}

	b.WriteString("\n")

	// Action menu
	b.WriteString(m.menu.View())

	// Footer with last refresh time
	if !m.lastRefresh.IsZero() {
		footer := SubtitleStyle.Render(fmt.Sprintf("  last refresh: %s", m.lastRefresh.Format("15:04:05")))
		b.WriteString("\n")
		b.WriteString(footer)
	}

	return b.String()
}

func (m DashboardModel) renderServices() string {
	var b strings.Builder

	// Use deploy order if available, otherwise iterate map
	order := m.config.DeployOrder
	if len(order) == 0 {
		for key := range m.services {
			order = append(order, key)
		}
	}

	for _, key := range order {
		svc, ok := m.services[key]
		if !ok {
			continue
		}

		var dot string
		var nameStyle lipgloss.Style

		switch svc.Status {
		case "healthy":
			dot = ServiceHealthy.Render("●")
			nameStyle = ServiceHealthy
		case "unhealthy", "failed":
			dot = ServiceUnhealthy.Render("●")
			nameStyle = ServiceUnhealthy
		default:
			dot = ServiceDisabled.Render("○")
			nameStyle = ServiceDisabled
		}

		name := svc.Name
		if name == "" {
			name = key
		}

		line := fmt.Sprintf("  %s %s", dot, nameStyle.Render(name))
		if svc.Status != "" {
			line += SubtitleStyle.Render(fmt.Sprintf("  %s", svc.Status))
		}

		b.WriteString(line)
		b.WriteString("\n")
	}

	return b.String()
}

func refreshStatus(eng *engine.Engine) tea.Cmd {
	return func() tea.Msg {
		result, err := eng.Status()
		if err != nil {
			return dashboardRefreshMsg{Err: err}
		}
		return dashboardRefreshMsg{Services: result.Services}
	}
}

func dashboardTick() tea.Cmd {
	return tea.Tick(20*time.Second, func(time.Time) tea.Msg {
		return dashboardTickMsg{}
	})
}
