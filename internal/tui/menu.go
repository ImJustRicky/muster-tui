package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// MenuItem represents a single menu option.
type MenuItem struct {
	Label       string
	Description string
	Value       string
}

// MenuModel is a single-select menu component.
type MenuModel struct {
	Items  []MenuItem
	Cursor int
	Title  string
}

// MenuSelectMsg is sent when the user presses enter on a menu item.
type MenuSelectMsg struct {
	Value string
	Index int
}

// NewMenu creates a menu with the given title and items.
func NewMenu(title string, items []MenuItem) MenuModel {
	return MenuModel{
		Title:  title,
		Items:  items,
		Cursor: 0,
	}
}

// Selected returns the currently highlighted menu item.
func (m MenuModel) Selected() MenuItem {
	if m.Cursor >= 0 && m.Cursor < len(m.Items) {
		return m.Items[m.Cursor]
	}
	return MenuItem{}
}

// Update handles key messages for navigation and selection.
func (m MenuModel) Update(msg tea.Msg) (MenuModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "j":
			if m.Cursor < len(m.Items)-1 {
				m.Cursor++
			}
		case "enter":
			item := m.Selected()
			return m, func() tea.Msg {
				return MenuSelectMsg{Value: item.Value, Index: m.Cursor}
			}
		}
	}
	return m, nil
}

// View renders the menu.
func (m MenuModel) View() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render(m.Title))
	b.WriteString("\n\n")

	for i, item := range m.Items {
		if i == m.Cursor {
			b.WriteString(MenuSelectedStyle.Render(fmt.Sprintf("  ▸ %s", item.Label)))
		} else {
			b.WriteString(MenuItemStyle.Render(fmt.Sprintf("    %s", item.Label)))
		}
		b.WriteString("\n")

		if item.Description != "" {
			if i == m.Cursor {
				b.WriteString(SubtitleStyle.Render(fmt.Sprintf("      %s", item.Description)))
			} else {
				b.WriteString(SubtitleStyle.Render(fmt.Sprintf("      %s", item.Description)))
			}
			b.WriteString("\n")
		}
	}

	return b.String()
}
