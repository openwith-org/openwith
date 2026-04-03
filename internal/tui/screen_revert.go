package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/openwith-org/openwith/internal/defaults"
)

type RevertModel struct {
	backups    []defaults.Backup
	cursor     int
	client     *defaults.Client
	dryRun     bool
	done       bool
	err        error
	offset     int
	viewHeight int
	width      int
}

func NewRevertModel(client *defaults.Client, dryRun bool, termHeight int) RevertModel {
	vh := termHeight - 10
	if vh < 5 {
		vh = 5
	}
	return RevertModel{
		backups:    defaults.ListBackups(),
		client:     client,
		dryRun:     dryRun,
		viewHeight: vh,
	}
}

func (m *RevertModel) SetDone(err error) {
	m.done = true
	m.err = err
}

func (m RevertModel) Update(msg tea.Msg) (RevertModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if m.done {
			switch msg.String() {
			case "enter", "esc", "q":
				return m, func() tea.Msg {
					return NavigateMsg{Screen: screenMenu}
				}
			}
			return m, nil
		}

		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				m.ensureVisible()
			}
		case "down", "j":
			if m.cursor < len(m.backups)-1 {
				m.cursor++
				m.ensureVisible()
			}
		case "enter":
			if len(m.backups) > 0 {
				backup := m.backups[m.cursor]
				client := m.client
				return m, func() tea.Msg {
					err := client.RestoreBackup(backup)
					return revertDoneMsg{err: err}
				}
			}
		case "esc":
			return m, func() tea.Msg {
				return NavigateMsg{Screen: screenMenu}
			}
		}
	case tea.WindowSizeMsg:
		m.viewHeight = msg.Height - 10
		if m.viewHeight < 5 {
			m.viewHeight = 5
		}
		m.width = msg.Width
	}
	return m, nil
}

func (m *RevertModel) ensureVisible() {
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+m.viewHeight {
		m.offset = m.cursor - m.viewHeight + 1
	}
}

func (m RevertModel) View() string {
	titleText := "Revert to Backup"
	if m.dryRun {
		titleText = BadgeStyle.Render(" DRY RUN ") + "  " + titleText
	}
	title := TitleStyle.Render(titleText)

	if m.done {
		var result string
		if m.err != nil {
			result = ErrorStyle.Render(fmt.Sprintf("  Restore failed: %s", m.err.Error()))
		} else {
			result = SuccessStyle.Render("  Backup restored successfully!")
		}
		help := HelpStyle.Render("enter/esc return to menu")
		return title + "\n\n" + result + "\n\n" + help
	}

	if len(m.backups) == 0 {
		empty := SubtitleStyle.Render("No backups available yet. Backups are created automatically when changes are applied.")
		help := HelpStyle.Render("esc back to menu")
		return title + "\n\n" + empty + "\n\n" + help
	}

	subtitle := SubtitleStyle.Render("Select a backup to restore")

	var lines []string
	for i, b := range m.backups {
		cursor := "  "
		style := InactiveItemStyle
		if i == m.cursor {
			cursor = AccentStyle.Render("▸ ")
			style = ActiveItemStyle
		}
		label := style.Render(b.Timestamp)
		if i == 0 {
			label += DimStyle.Render("  (most recent)")
		}
		lines = append(lines, fmt.Sprintf("%s%s", cursor, label))
	}

	// Viewport
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
		Render(strings.Join(visible, "\n"))

	help := HelpStyle.Render("↑/↓ navigate • enter restore • esc back")

	return fmt.Sprintf("%s\n%s\n\n%s\n%s", title, subtitle, content, help)
}
