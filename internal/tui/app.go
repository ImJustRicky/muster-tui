package tui

import (
	"github.com/ImJustRicky/muster-tui/internal/config"
	"github.com/ImJustRicky/muster-tui/internal/engine"
	tea "github.com/charmbracelet/bubbletea"
)

type screen int

const (
	screenDashboard screen = iota
	screenDeploy
	screenLogViewer
)

// App is the root bubbletea model that manages screen routing.
type App struct {
	engine    *engine.Engine
	config    *config.DeployConfig
	screen    screen
	dashboard DashboardModel
	deploy    DeployModel
	logViewer LogViewerModel
	width     int
	height    int
}

// NewApp creates the root application model.
func NewApp(eng *engine.Engine, cfg *config.DeployConfig) App {
	return App{
		engine:    eng,
		config:    cfg,
		screen:    screenDashboard,
		dashboard: NewDashboard(eng, cfg),
	}
}

func (a App) Init() tea.Cmd {
	return a.dashboard.Init()
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		// Forward to active screen
		switch a.screen {
		case screenDashboard:
			m, cmd := a.dashboard.Update(msg)
			a.dashboard = m.(DashboardModel)
			return a, cmd
		case screenDeploy:
			m, cmd := a.deploy.Update(msg)
			a.deploy = m.(DeployModel)
			return a, cmd
		case screenLogViewer:
			lv, cmd := a.logViewer.Update(msg)
			a.logViewer = lv
			return a, cmd
		}

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return a, tea.Quit
		}

	// Navigation messages
	case StartDeployMsg:
		a.screen = screenDeploy
		a.deploy = NewDeploy(a.engine, msg.Services, false)
		return a, a.deploy.Init()

	case StartDryRunMsg:
		a.screen = screenDeploy
		a.deploy = NewDeploy(a.engine, msg.Services, true)
		return a, a.deploy.Init()

	case OpenLogViewerMsg:
		a.screen = screenLogViewer
		a.logViewer = NewLogViewer(msg.Title)
		a.logViewer.SetSize(a.width, a.height)
		for _, line := range msg.Lines {
			a.logViewer.AppendLine(line)
		}
		return a, nil

	case CloseLogViewerMsg:
		a.screen = screenDeploy
		return a, nil

	case BackToDashboardMsg:
		a.screen = screenDashboard
		a.dashboard = NewDashboard(a.engine, a.config)
		return a, a.dashboard.Init()
	}

	// Route to active screen
	switch a.screen {
	case screenDashboard:
		m, cmd := a.dashboard.Update(msg)
		a.dashboard = m.(DashboardModel)
		return a, cmd
	case screenDeploy:
		m, cmd := a.deploy.Update(msg)
		a.deploy = m.(DeployModel)
		return a, cmd
	case screenLogViewer:
		lv, cmd := a.logViewer.Update(msg)
		a.logViewer = lv
		return a, cmd
	}

	return a, nil
}

func (a App) View() string {
	switch a.screen {
	case screenDashboard:
		return a.dashboard.View()
	case screenDeploy:
		return a.deploy.View()
	case screenLogViewer:
		return a.logViewer.View()
	}
	return ""
}
