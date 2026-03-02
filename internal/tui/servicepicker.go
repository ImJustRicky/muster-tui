package tui

import (
	"fmt"
	"strings"

	"github.com/ImJustRicky/muster-tui/internal/config"
	tea "github.com/charmbracelet/bubbletea"
)

// ServicePickerModel lets the user pick a service from the deploy config.
type ServicePickerModel struct {
	services []servicePickerItem
	cursor   int
	title    string
	action   string
	width    int
	height   int
}

type servicePickerItem struct {
	Key  string
	Name string
}

// NewServicePicker creates a service picker from a deploy config.
func NewServicePicker(cfg *config.DeployConfig, title, action string) ServicePickerModel {
	var items []servicePickerItem

	if cfg == nil {
		return ServicePickerModel{title: title, action: action}
	}

	// Use deploy order if available
	order := cfg.DeployOrder
	if len(order) == 0 {
		for key := range cfg.Services {
			order = append(order, key)
		}
	}

	for _, key := range order {
		svc, ok := cfg.Services[key]
		if !ok {
			continue
		}
		name := svc.Name
		if name == "" {
			name = key
		}
		items = append(items, servicePickerItem{Key: key, Name: name})
	}

	return ServicePickerModel{
		services: items,
		title:    title,
		action:   action,
	}
}

func (m ServicePickerModel) Init() tea.Cmd { return nil }

func (m ServicePickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if m.cursor < len(m.services)-1 {
				m.cursor++
			}
		case "enter":
			if len(m.services) > 0 {
				item := m.services[m.cursor]
				return m, func() tea.Msg {
					return ServicePickedMsg{
						ServiceKey:  item.Key,
						ServiceName: item.Name,
						Action:      m.action,
					}
				}
			}
		case "esc", "q":
			return m, func() tea.Msg { return ServicePickerCancelMsg{} }
		}
	}
	return m, nil
}

func (m ServicePickerModel) View() string {
	var b strings.Builder

	b.WriteString(HeaderStyle.Render("  muster  "))
	b.WriteString("\n\n")
	b.WriteString(TitleStyle.Render(m.title))
	b.WriteString("\n\n")

	if len(m.services) == 0 {
		b.WriteString(SubtitleStyle.Render("  No services found."))
		b.WriteString("\n\n")
		b.WriteString(SubtitleStyle.Render("  press esc to go back"))
		return b.String()
	}

	for i, item := range m.services {
		if i == m.cursor {
			b.WriteString(MenuSelectedStyle.Render(fmt.Sprintf("  ▸ %s", item.Name)))
		} else {
			b.WriteString(MenuItemStyle.Render(fmt.Sprintf("    %s", item.Name)))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(SubtitleStyle.Render("  enter select • esc back"))

	return b.String()
}
