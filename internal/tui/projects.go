package tui

import (
	"fmt"
	"strings"

	"github.com/ImJustRicky/muster-tui/internal/registry"
	tea "github.com/charmbracelet/bubbletea"
)

// ProjectSelectedMsg is sent when the user picks a project.
type ProjectSelectedMsg struct {
	Project registry.Project
}

// ProjectsModel shows a list of registered projects for selection.
type ProjectsModel struct {
	projects []registry.Project
	cursor   int
	width    int
	height   int
}

func NewProjects(projects []registry.Project) ProjectsModel {
	return ProjectsModel{
		projects: projects,
		cursor:   0,
	}
}

func (m ProjectsModel) Init() tea.Cmd {
	return nil
}

func (m ProjectsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.projects)-1 {
				m.cursor++
			}
		case "enter":
			if m.cursor < len(m.projects) {
				p := m.projects[m.cursor]
				return m, func() tea.Msg {
					return ProjectSelectedMsg{Project: p}
				}
			}
		case "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m ProjectsModel) View() string {
	var b strings.Builder

	header := HeaderStyle.Render("  muster  ")
	header += "  " + TitleStyle.Render("Select Project")
	b.WriteString(header)
	b.WriteString("\n\n")

	for i, p := range m.projects {
		name := p.Name
		if name == "" {
			name = p.Path
		}

		svcLabel := fmt.Sprintf("%d services", p.ServiceCount)
		if p.ServiceCount == 1 {
			svcLabel = "1 service"
		}

		if i == m.cursor {
			b.WriteString(MenuSelectedStyle.Render(fmt.Sprintf("  ▸ %s", name)))
			b.WriteString("\n")
			b.WriteString(SubtitleStyle.Render(fmt.Sprintf("      %s  •  %s", p.Path, svcLabel)))
		} else {
			b.WriteString(MenuItemStyle.Render(fmt.Sprintf("    %s", name)))
			b.WriteString("\n")
			b.WriteString(SubtitleStyle.Render(fmt.Sprintf("      %s  •  %s", p.Path, svcLabel)))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(SubtitleStyle.Render("  ↑/↓ navigate • enter select • q quit"))

	return b.String()
}
