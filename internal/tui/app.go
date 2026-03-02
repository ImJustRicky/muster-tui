package tui

import (
	"github.com/ImJustRicky/muster-tui/internal/config"
	"github.com/ImJustRicky/muster-tui/internal/engine"
	"github.com/ImJustRicky/muster-tui/internal/registry"
	tea "github.com/charmbracelet/bubbletea"
)

type screen int

const (
	screenProjectPicker screen = iota
	screenDashboard
	screenDeploy
	screenLogViewer
)

// App is the root bubbletea model that manages screen routing.
type App struct {
	engine    *engine.Engine
	config    *config.DeployConfig
	screen    screen
	projects  ProjectsModel
	dashboard DashboardModel
	deploy    DeployModel
	logViewer LogViewerModel
	projectDir string
	width     int
	height    int
}

// NewApp creates the root application model starting at the dashboard.
func NewApp(eng *engine.Engine, cfg *config.DeployConfig) App {
	return App{
		engine:    eng,
		config:    cfg,
		screen:    screenDashboard,
		dashboard: NewDashboard(eng, cfg),
	}
}

// NewAppWithPicker creates the root model starting at the project picker.
func NewAppWithPicker(eng *engine.Engine, projects []registry.Project) App {
	return App{
		engine:   eng,
		screen:   screenProjectPicker,
		projects: NewProjects(projects),
	}
}

func (a App) Init() tea.Cmd {
	switch a.screen {
	case screenProjectPicker:
		return a.projects.Init()
	case screenDashboard:
		return a.dashboard.Init()
	}
	return nil
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		// Forward to active screen
		switch a.screen {
		case screenProjectPicker:
			m, cmd := a.projects.Update(msg)
			a.projects = m.(ProjectsModel)
			return a, cmd
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

	// Project picker → dashboard transition
	case ProjectSelectedMsg:
		a.projectDir = msg.Project.Path
		configPath := config.FindConfigIn(msg.Project.Path)
		cfg, err := config.LoadDeploy(configPath)
		if configPath == "" || err != nil {
			// Fall back — just go to dashboard with nil config
			a.screen = screenDashboard
			a.config = nil
			a.dashboard = NewDashboard(a.engine, nil)
			return a, a.dashboard.Init()
		}
		a.config = cfg
		a.screen = screenDashboard
		a.dashboard = NewDashboard(a.engine, cfg)
		return a, a.dashboard.Init()

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
	case screenProjectPicker:
		m, cmd := a.projects.Update(msg)
		a.projects = m.(ProjectsModel)
		return a, cmd
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
	case screenProjectPicker:
		return a.projects.View()
	case screenDashboard:
		return a.dashboard.View()
	case screenDeploy:
		return a.deploy.View()
	case screenLogViewer:
		return a.logViewer.View()
	}
	return ""
}
