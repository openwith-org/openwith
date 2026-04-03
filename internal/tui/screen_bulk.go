package tui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/openwith-org/openwith/internal/editors"
)

type BulkModel struct {
	editors    []editors.Editor
	cursor     int
	offset     int
	viewHeight int
	width      int
}

func NewBulkModel(eds []editors.Editor, termHeight int) BulkModel {
	vh := termHeight - 8 // title + subtitle + border + help
	if vh < 5 {
		vh = 5
	}
	return BulkModel{editors: eds, viewHeight: vh}
}

func (m *BulkModel) ensureVisible() {
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+m.viewHeight {
		m.offset = m.cursor - m.viewHeight + 1
	}
}

func (m BulkModel) Update(msg tea.Msg) (BulkModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				m.ensureVisible()
			}
		case "down", "j":
			if m.cursor < len(m.editors)-1 {
				m.cursor++
				m.ensureVisible()
			}
		case "enter":
			selected := m.editors[m.cursor]
			return m, func() tea.Msg {
				return BulkSelectMsg{Editor: selected}
			}
		case "esc":
			return m, func() tea.Msg {
				return NavigateMsg{Screen: screenMenu}
			}
		}
	case tea.WindowSizeMsg:
		m.viewHeight = msg.Height - 8
		if m.viewHeight < 5 {
			m.viewHeight = 5
		}
		m.width = msg.Width
	}
	return m, nil
}

func (m BulkModel) View() string {
	title := TitleStyle.Render("Bulk Mode")
	subtitle := SubtitleStyle.Render("Select an editor for all file types")

	var lines []string
	for i, ed := range m.editors {
		cursor := "  "
		style := InactiveItemStyle
		if i == m.cursor {
			cursor = AccentStyle.Render("▸ ")
			style = ActiveItemStyle
		}
		lines = append(lines, fmt.Sprintf("%s%s  %s",
			cursor,
			style.Render(ed.Name),
			DimStyle.Render(ed.BundleID),
		))
	}

	visible := lines
	if m.offset > 0 && m.offset < len(lines) {
		visible = lines[m.offset:]
	}
	if len(visible) > m.viewHeight {
		visible = visible[:m.viewHeight]
	}

	content := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Purple).
		Padding(1, 2).
		Render(joinLines(visible))

	help := HelpStyle.Render("↑/↓ navigate • enter select • esc back")

	return fmt.Sprintf("%s\n%s\n\n%s\n%s", title, subtitle, content, help)
}
