package tui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type menuItem struct {
	title       string
	description string
	screen      screen
}

type MenuModel struct {
	items  []menuItem
	cursor int
}

func NewMenuModel() MenuModel {
	return MenuModel{
		items: []menuItem{
			{
				title:       "Bulk Mode",
				description: "Set one editor for all file types at once",
				screen:      screenBulk,
			},
			{
				title:       "Per-Extension Mode",
				description: "Choose a different editor for each file type",
				screen:      screenPerExt,
			},
			{
				title:       "Default Browser",
				description: "Choose which browser opens links",
				screen:      screenBrowser,
			},
			{
				title:       "Load Profile",
				description: "Apply a saved set of file associations",
				screen:      screenProfiles,
			},
			{
				title:       "Revert to Backup",
				description: "Restore file associations from a previous backup",
				screen:      screenRevert,
			},
		},
	}
}

func (m MenuModel) Update(msg tea.Msg) (MenuModel, tea.Cmd) {
	if msg, ok := msg.(tea.KeyPressMsg); ok {
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case "enter":
			return m, func() tea.Msg {
				return NavigateMsg{Screen: m.items[m.cursor].screen}
			}
		}
	}
	return m, nil
}

func (m MenuModel) View() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(Purple).
		Padding(1, 0).
		Render("  openwith")

	subtitle := SubtitleStyle.Render("Manage default file associations on macOS")

	var items string
	for i, item := range m.items {
		cursor := "  "
		style := InactiveItemStyle
		descStyle := DimStyle
		if i == m.cursor {
			cursor = AccentStyle.Render("▸ ")
			style = ActiveItemStyle
			descStyle = lipgloss.NewStyle().Foreground(Gray)
		}
		items += fmt.Sprintf("%s%s\n%s%s\n\n",
			cursor,
			style.Render(item.title),
			"    ",
			descStyle.Render(item.description),
		)
	}

	content := BoxStyle.Render(items)

	help := HelpStyle.Render("↑/↓ navigate • enter select • q quit")

	return fmt.Sprintf("%s\n%s\n\n%s\n%s", title, subtitle, content, help)
}
