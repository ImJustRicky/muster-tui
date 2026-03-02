package tui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/ImJustRicky/muster-tui/internal/config"
	"github.com/ImJustRicky/muster-tui/internal/engine"
	"github.com/ImJustRicky/muster-tui/internal/skills"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Internal messages.

type dashboardRefreshMsg struct {
	Services map[string]engine.ServiceStatus
	Err      error
}

type dashboardTickMsg struct{}

// DashboardModel is the main dashboard view with split-panel layout.
type DashboardModel struct {
	engine      *engine.Engine
	config      *config.DeployConfig
	services    map[string]engine.ServiceStatus
	skills      []skills.Skill
	menu        MenuModel
	width       int
	height      int
	loading     bool
	err         error
	lastRefresh time.Time
	projectDir  string
}

func NewDashboard(eng *engine.Engine, cfg *config.DeployConfig) DashboardModel {
	items := []MenuItem{
		{Label: "Deploy All", Description: "Deploy all services", Value: "deploy_all"},
		{Label: "Deploy Service", Description: "Pick a service to deploy", Value: "deploy_select"},
		{Label: "Status", Description: "Service health details", Value: "status"},
		{Label: "Logs", Description: "Stream service logs", Value: "logs"},
		{Label: "Rollback", Description: "Revert a service", Value: "rollback"},
		{Label: "History", Description: "Deploy/rollback history", Value: "history"},
		{Label: "Doctor", Description: "Run diagnostics", Value: "doctor"},
		{Label: "Skills", Description: "Manage addon skills", Value: "skills"},
		{Label: "Cleanup", Description: "Clean stuck processes", Value: "cleanup"},
		{Label: "Settings", Description: "View and edit settings", Value: "settings"},
		{Label: "Quit", Value: "quit"},
	}

	// Discover project dir for skill scanning
	projectDir := ""
	if cfg != nil {
		// Use cwd — the engine runs commands from here
		if dir, err := findProjectDir(); err == nil {
			projectDir = dir
		}
	}

	return DashboardModel{
		engine:     eng,
		config:     cfg,
		menu:       NewMenu("", items),
		loading:    true,
		skills:     skills.ScanAll(projectDir),
		projectDir: projectDir,
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
			return m, func() tea.Msg {
				return ShowServicePickerMsg{Title: "Deploy which service?", Action: "deploy"}
			}
		case "status":
			return m, func() tea.Msg { return ShowStatusMsg{} }
		case "history":
			return m, func() tea.Msg { return ShowHistoryMsg{} }
		case "doctor":
			return m, func() tea.Msg { return ShowDoctorMsg{} }
		case "logs":
			return m, func() tea.Msg {
				return ShowServicePickerMsg{Title: "Stream logs for", Action: "logs"}
			}
		case "rollback":
			return m, func() tea.Msg {
				return ShowServicePickerMsg{Title: "Rollback which service?", Action: "rollback"}
			}
		case "cleanup":
			return m, func() tea.Msg { return StartCleanupMsg{} }
		case "settings":
			return m, func() tea.Msg { return ShowSettingsMsg{} }
		case "skills":
			return m, func() tea.Msg { return ShowSkillsMsg{} }
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
	if m.width == 0 {
		return "  Loading..."
	}

	// ── Full-width header bar ──
	header := m.renderHeader()

	// ── Left panel: services + skills ──
	leftContent := m.renderLeftPanel()

	// ── Right panel: action menu ──
	rightContent := m.renderRightPanel()

	// Panel sizing
	totalWidth := m.width
	if totalWidth > 120 {
		totalWidth = 120
	}
	leftWidth := totalWidth * 2 / 5
	rightWidth := totalWidth - leftWidth - 3 // -3 for gap
	if leftWidth < 25 {
		leftWidth = 25
	}
	if rightWidth < 25 {
		rightWidth = 25
	}

	panelHeight := m.height - 5 // header + footer + padding
	if panelHeight < 10 {
		panelHeight = 10
	}

	leftBox := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(Gold).
		Width(leftWidth).
		Height(panelHeight).
		Padding(1, 2).
		Render(leftContent)

	rightBox := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#444444")).
		Width(rightWidth).
		Height(panelHeight).
		Padding(1, 2).
		Render(rightContent)

	panels := lipgloss.JoinHorizontal(lipgloss.Top, leftBox, " ", rightBox)

	// ── Full-width footer bar ──
	footer := m.renderFooter(totalWidth + 2)

	// Stack everything with margin
	body := lipgloss.NewStyle().Margin(0, 1).Render(
		lipgloss.JoinVertical(lipgloss.Left, header, "", panels, footer),
	)

	return body
}

func (m DashboardModel) renderHeader() string {
	badge := HeaderStyle.Render("  muster  ")
	project := ""
	if m.config != nil && m.config.Project != "" {
		project = "  " + TitleStyle.Render(m.config.Project)
	}

	left := badge + project

	// Version on right
	ver := ""
	if v, err := m.engine.Version(); err == nil {
		ver = SubtitleStyle.Render(v)
	}

	totalWidth := m.width - 4
	if totalWidth > 120 {
		totalWidth = 120
	}
	gap := totalWidth - lipgloss.Width(left) - lipgloss.Width(ver)
	if gap < 1 {
		gap = 1
	}

	return left + strings.Repeat(" ", gap) + ver
}

func (m DashboardModel) renderLeftPanel() string {
	var b strings.Builder

	// Services section
	b.WriteString(TitleStyle.Render("Services"))
	b.WriteString("\n\n")

	if m.loading && m.services == nil {
		b.WriteString(SubtitleStyle.Render("Loading..."))
		b.WriteString("\n")
	} else if m.err != nil {
		b.WriteString(LogErrorStyle.Render(fmt.Sprintf("Error: %s", m.err)))
		b.WriteString("\n")
	} else {
		b.WriteString(m.renderServices())
	}

	// Skills section
	b.WriteString("\n")
	b.WriteString(TitleStyle.Render("Skills"))
	b.WriteString("\n\n")

	if len(m.skills) == 0 {
		b.WriteString(SubtitleStyle.Render("No skills installed"))
		b.WriteString("\n")
	} else {
		for _, s := range m.skills {
			var dot string
			var nameStyle lipgloss.Style
			if s.Enabled {
				dot = ServiceHealthy.Render("●")
				nameStyle = lipgloss.NewStyle().Foreground(White)
			} else {
				dot = ServiceDisabled.Render("○")
				nameStyle = lipgloss.NewStyle().Foreground(Gray)
			}

			line := fmt.Sprintf("%s %s", dot, nameStyle.Render(s.Name))

			// Show hooks or status
			if s.Enabled && len(s.Hooks) > 0 {
				hookStr := truncateHooks(s.Hooks)
				line += "  " + SubtitleStyle.Render(hookStr)
			} else if !s.Enabled {
				line += "  " + SubtitleStyle.Render("disabled")
			} else {
				line += "  " + SubtitleStyle.Render("manual")
			}

			// Scope tag for global skills
			if s.Scope == "global" {
				line += "  " + lipgloss.NewStyle().Foreground(lipgloss.Color("#555555")).Render("global")
			}

			b.WriteString(line)
			b.WriteString("\n")
		}
	}

	return b.String()
}

func (m DashboardModel) renderRightPanel() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("Actions"))
	b.WriteString("\n\n")

	for i, item := range m.menu.Items {
		if i == m.menu.Cursor {
			b.WriteString(MenuSelectedStyle.Render(fmt.Sprintf("▸ %s", item.Label)))
			if item.Description != "" {
				b.WriteString("\n")
				b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render(
					fmt.Sprintf("  %s", item.Description),
				))
			}
		} else {
			b.WriteString(MenuItemStyle.Render(fmt.Sprintf("  %s", item.Label)))
		}
		b.WriteString("\n")
	}

	return b.String()
}

func (m DashboardModel) renderServices() string {
	var b strings.Builder

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

		line := fmt.Sprintf("%s %s", dot, nameStyle.Render(name))
		if svc.Status != "" {
			line += "  " + SubtitleStyle.Render(svc.Status)
		}

		b.WriteString(line)
		b.WriteString("\n")
	}

	return b.String()
}

func (m DashboardModel) renderFooter(width int) string {
	left := SubtitleStyle.Render("↑/↓ navigate • enter select • r refresh • q quit")

	right := ""
	if !m.lastRefresh.IsZero() {
		right = SubtitleStyle.Render(fmt.Sprintf("last refresh %s", m.lastRefresh.Format("15:04:05")))
	}

	gap := width - lipgloss.Width(left) - lipgloss.Width(right) - 2
	if gap < 1 {
		gap = 1
	}

	return StatusBarStyle.Width(width).Render(left + strings.Repeat(" ", gap) + right)
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

// truncateHooks formats hooks list for display.
func truncateHooks(hooks []string) string {
	if len(hooks) == 0 {
		return ""
	}
	// Shorten hook names
	short := make([]string, len(hooks))
	for i, h := range hooks {
		h = strings.TrimPrefix(h, "pre-")
		h = strings.TrimPrefix(h, "post-")
		short[i] = h
	}
	// Deduplicate
	seen := map[string]bool{}
	var unique []string
	for _, s := range short {
		if !seen[s] {
			seen[s] = true
			unique = append(unique, s)
		}
	}
	return strings.Join(unique, ", ")
}

// findProjectDir returns the project root (directory containing the config file).
func findProjectDir() (string, error) {
	cfgPath, err := config.FindConfig()
	if err != nil {
		return "", err
	}
	return filepath.Dir(cfgPath), nil
}
