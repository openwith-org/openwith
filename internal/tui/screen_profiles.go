package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/openwith-org/openwith/internal/config"
)

type ProfilesModel struct {
	profiles    []config.Profile
	cursor      int
	editorNames map[string]string // bundleID -> name
	changes     map[string]string // current pending changes (for save option)
	offset      int
	viewHeight  int
	width       int
}

func NewProfilesModel(currentChanges map[string]string, editorNames map[string]string, termHeight int) ProfilesModel {
	vh := termHeight - 10
	if vh < 5 {
		vh = 5
	}
	return ProfilesModel{
		profiles:    config.ListProfiles(),
		editorNames: editorNames,
		changes:     currentChanges,
		viewHeight:  vh,
	}
}

func (m ProfilesModel) Update(msg tea.Msg) (ProfilesModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				m.ensureVisible()
			}
		case "down", "j":
			if m.cursor < len(m.profiles)-1 {
				m.cursor++
				m.ensureVisible()
			}
		case "enter":
			if len(m.profiles) > 0 {
				changes := m.profiles[m.cursor].Changes
				return m, func() tea.Msg {
					return profileLoadedMsg{changes: changes}
				}
			}
		case "d":
			if len(m.profiles) > 0 {
				name := m.profiles[m.cursor].Name
				_ = config.DeleteProfile(name)
				m.profiles = config.ListProfiles()
				if m.cursor >= len(m.profiles) && m.cursor > 0 {
					m.cursor--
				}
			}
		case "s":
			return m, func() tea.Msg {
				return NavigateMsg{Screen: screenSaveProfile}
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

func (m *ProfilesModel) ensureVisible() {
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+m.viewHeight {
		m.offset = m.cursor - m.viewHeight + 1
	}
}

func (m ProfilesModel) View() string {
	title := TitleStyle.Render("Profiles")

	if len(m.profiles) == 0 {
		empty := SubtitleStyle.Render("No profiles saved yet.")
		hint := DimStyle.Render("  Use Per-Extension or Bulk mode, then save as a profile from the confirm screen (s).")
		help := HelpStyle.Render("s create new profile • esc back")
		return title + "\n\n" + empty + "\n" + hint + "\n\n" + help
	}

	subtitle := SubtitleStyle.Render("Select a profile to load")

	var lines []string
	for i, p := range m.profiles {
		cursor := "  "
		style := InactiveItemStyle
		if i == m.cursor {
			cursor = AccentStyle.Render("▸ ")
			style = ActiveItemStyle
		}
		count := len(p.Changes)
		label := fmt.Sprintf("%s  %s",
			style.Render(p.Name),
			DimStyle.Render(fmt.Sprintf("(%d associations)", count)),
		)
		lines = append(lines, cursor+label)
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

	help := HelpStyle.Render("↑/↓ navigate • enter load • s save new • d delete • esc back")

	return fmt.Sprintf("%s\n%s\n\n%s\n%s", title, subtitle, content, help)
}
