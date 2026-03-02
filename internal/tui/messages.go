package tui

// Cross-screen navigation messages.
// Per-screen internal messages stay in their own files.

type ShowStatusMsg struct{}

type ShowHistoryMsg struct{}

type ShowDoctorMsg struct{}

type ShowSettingsMsg struct{}

type StartCleanupMsg struct{}

type ShowServicePickerMsg struct {
	Title  string
	Action string // "deploy", "logs", "rollback"
}

type ServicePickedMsg struct {
	ServiceKey  string
	ServiceName string
	Action      string
}

type ServicePickerCancelMsg struct{}

type StartDeployMsg struct {
	Services []string
}

type StartDryRunMsg struct {
	Services []string
}

type OpenLogViewerMsg struct {
	Lines []string
	Title string
}

type BackToDashboardMsg struct{}

type ShowSkillsMsg struct{}
