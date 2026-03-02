package tui

import (
	"time"

	"github.com/ImJustRicky/muster-tui/internal/config"
	"github.com/ImJustRicky/muster-tui/internal/engine"
	"github.com/ImJustRicky/muster-tui/internal/registry"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type screen int

const (
	screenProjectPicker screen = iota
	screenDashboard
	screenDeploy
	screenLogViewer
	screenStatus
	screenHistory
	screenDoctor
	screenServicePicker
	screenRawOutput
	screenSettings
)

// ctrlCResetMsg clears the Ctrl+C warning after timeout.
type ctrlCResetMsg struct{}

// App is the root bubbletea model that manages screen routing.
type App struct {
	engine        *engine.Engine
	config        *config.DeployConfig
	screen        screen
	projects      ProjectsModel
	dashboard     DashboardModel
	deploy        DeployModel
	logViewer     LogViewerModel
	status        StatusModel
	history       HistoryModel
	doctor        DoctorModel
	servicePicker ServicePickerModel
	rawOutput     RawOutputModel
	settings      SettingsModel
	projectDir    string
	width         int
	height        int
	ctrlCPressed  bool // true after first Ctrl+C
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
		return a.forwardToScreen(msg)

	case ctrlCResetMsg:
		a.ctrlCPressed = false
		return a, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			if a.ctrlCPressed {
				return a, tea.Quit
			}
			a.ctrlCPressed = true
			return a, tea.Tick(5*time.Second, func(time.Time) tea.Msg {
				return ctrlCResetMsg{}
			})
		}
		// Any other key clears the Ctrl+C warning
		if a.ctrlCPressed {
			a.ctrlCPressed = false
		}

	// Project picker → dashboard transition
	case ProjectSelectedMsg:
		a.projectDir = msg.Project.Path
		configPath := config.FindConfigIn(msg.Project.Path)
		cfg, err := config.LoadDeploy(configPath)
		if configPath == "" || err != nil {
			a.screen = screenDashboard
			a.config = nil
			a.dashboard = NewDashboard(a.engine, nil)
			return a, a.dashboard.Init()
		}
		a.config = cfg
		a.screen = screenDashboard
		a.dashboard = NewDashboard(a.engine, cfg)
		return a, a.dashboard.Init()

	// Deploy
	case StartDeployMsg:
		a.screen = screenDeploy
		a.deploy = NewDeploy(a.engine, msg.Services, false)
		return a, a.deploy.Init()

	case StartDryRunMsg:
		a.screen = screenDeploy
		a.deploy = NewDeploy(a.engine, msg.Services, true)
		return a, a.deploy.Init()

	// Log viewer (from deploy)
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

	// Status
	case ShowStatusMsg:
		a.screen = screenStatus
		a.status = NewStatus(a.engine, a.config)
		return a, a.status.Init()

	// History
	case ShowHistoryMsg:
		a.screen = screenHistory
		a.history = NewHistory(a.engine)
		return a, a.history.Init()

	// Doctor
	case ShowDoctorMsg:
		a.screen = screenDoctor
		a.doctor = NewDoctor(a.engine)
		return a, a.doctor.Init()

	// Service picker (for deploy service, logs, rollback)
	case ShowServicePickerMsg:
		a.screen = screenServicePicker
		a.servicePicker = NewServicePicker(a.config, msg.Title, msg.Action)
		return a, a.servicePicker.Init()

	case ServicePickedMsg:
		switch msg.Action {
		case "deploy":
			a.screen = screenDeploy
			a.deploy = NewDeploy(a.engine, []string{msg.ServiceKey}, false)
			return a, a.deploy.Init()
		case "logs":
			a.screen = screenRawOutput
			a.rawOutput = NewRawOutput(a.engine, "Logs: "+msg.ServiceName, []string{"logs", msg.ServiceKey})
			return a, a.rawOutput.Init()
		case "rollback":
			a.screen = screenRawOutput
			a.rawOutput = NewRawOutput(a.engine, "Rollback: "+msg.ServiceName, []string{"rollback", msg.ServiceKey})
			return a, a.rawOutput.Init()
		}

	case ServicePickerCancelMsg:
		a.screen = screenDashboard
		return a, nil

	// Skills (raw output)
	case ShowSkillsMsg:
		a.screen = screenRawOutput
		a.rawOutput = NewRawOutput(a.engine, "Skills", []string{"skill", "list"})
		return a, a.rawOutput.Init()

	// Cleanup (raw output)
	case StartCleanupMsg:
		a.screen = screenRawOutput
		a.rawOutput = NewRawOutput(a.engine, "Cleanup", []string{"cleanup"})
		return a, a.rawOutput.Init()

	// Settings
	case ShowSettingsMsg:
		a.screen = screenSettings
		a.settings = NewSettings()
		return a, a.settings.Init()

	// Back to dashboard
	case BackToDashboardMsg:
		a.screen = screenDashboard
		a.dashboard = NewDashboard(a.engine, a.config)
		a.dashboard.width = a.width
		a.dashboard.height = a.height
		return a, a.dashboard.Init()
	}

	// Route to active screen
	return a.forwardToScreen(msg)
}

func (a App) forwardToScreen(msg tea.Msg) (tea.Model, tea.Cmd) {
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
	case screenStatus:
		m, cmd := a.status.Update(msg)
		a.status = m.(StatusModel)
		return a, cmd
	case screenHistory:
		m, cmd := a.history.Update(msg)
		a.history = m.(HistoryModel)
		return a, cmd
	case screenDoctor:
		m, cmd := a.doctor.Update(msg)
		a.doctor = m.(DoctorModel)
		return a, cmd
	case screenServicePicker:
		m, cmd := a.servicePicker.Update(msg)
		a.servicePicker = m.(ServicePickerModel)
		return a, cmd
	case screenRawOutput:
		m, cmd := a.rawOutput.Update(msg)
		a.rawOutput = m.(RawOutputModel)
		return a, cmd
	case screenSettings:
		m, cmd := a.settings.Update(msg)
		a.settings = m.(SettingsModel)
		return a, cmd
	}
	return a, nil
}

func (a App) View() string {
	view := ""
	switch a.screen {
	case screenProjectPicker:
		view = a.projects.View()
	case screenDashboard:
		view = a.dashboard.View()
	case screenDeploy:
		view = a.deploy.View()
	case screenLogViewer:
		view = a.logViewer.View()
	case screenStatus:
		view = a.status.View()
	case screenHistory:
		view = a.history.View()
	case screenDoctor:
		view = a.doctor.View()
	case screenServicePicker:
		view = a.servicePicker.View()
	case screenRawOutput:
		view = a.rawOutput.View()
	case screenSettings:
		view = a.settings.View()
	}

	// Overlay Ctrl+C warning
	if a.ctrlCPressed {
		warning := lipgloss.NewStyle().
			Foreground(Yellow).
			Bold(true).
			Render("  ! Press Ctrl+C again to quit")
		view += "\n" + warning
	}

	return view
}
