package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/openwith-org/openwith/internal/defaults"
	"github.com/openwith-org/openwith/internal/editors"
)

type browserSetMsg struct {
	err error
}

type BrowserModel struct {
	browsers       []editors.Editor
	cursor         int
	currentBrowser string
	client         *defaults.Client
	dryRun         bool
	done           bool
	err            error
	offset         int
	viewHeight     int
	width          int
}

func NewBrowserModel(client *defaults.Client, dryRun bool, termHeight int) BrowserModel {
	vh := termHeight - 10
	if vh < 5 {
		vh = 5
	}
	return BrowserModel{
		browsers:       editors.DetectInstalledBrowsers(),
		currentBrowser: client.GetDefaultBrowser(),
		client:         client,
		dryRun:         dryRun,
		viewHeight:     vh,
	}
}

func (m BrowserModel) Update(msg tea.Msg) (BrowserModel, tea.Cmd) {
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
			if m.cursor < len(m.browsers)-1 {
				m.cursor++
				m.ensureVisible()
			}
		case "enter":
			if len(m.browsers) > 0 {
				selected := m.browsers[m.cursor]
				client := m.client
				return m, func() tea.Msg {
					err := client.SetDefaultBrowser(selected.BundleID)
					return browserSetMsg{err: err}
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

func (m *BrowserModel) SetDone(err error) {
	m.done = true
	m.err = err
}

func (m *BrowserModel) ensureVisible() {
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+m.viewHeight {
		m.offset = m.cursor - m.viewHeight + 1
	}
}

func (m BrowserModel) View() string {
	titleText := "Default Browser"
	if m.dryRun {
		titleText = BadgeStyle.Render(" DRY RUN ") + "  " + titleText
	}
	title := TitleStyle.Render(titleText)

	if m.done {
		var result string
		if m.err != nil {
			result = ErrorStyle.Render(fmt.Sprintf("  Error: %s", m.err.Error()))
		} else {
			result = SuccessStyle.Render("  Default browser updated!")
		}
		help := HelpStyle.Render("enter/esc return to menu")
		return title + "\n\n" + result + "\n\n" + help
	}

	if len(m.browsers) == 0 {
		empty := SubtitleStyle.Render("No supported browsers found.")
		help := HelpStyle.Render("esc back to menu")
		return title + "\n\n" + empty + "\n\n" + help
	}

	subtitle := SubtitleStyle.Render(fmt.Sprintf("Current: %s", AccentStyle.Render(m.currentBrowser)))

	var lines []string
	for i, b := range m.browsers {
		cursor := "  "
		style := InactiveItemStyle
		if i == m.cursor {
			cursor = AccentStyle.Render("▸ ")
			style = ActiveItemStyle
		}
		label := style.Render(b.Name)
		extra := DimStyle.Render("  " + b.BundleID)
		lines = append(lines, fmt.Sprintf("%s%s%s", cursor, label, extra))
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

	help := HelpStyle.Render("↑/↓ navigate • enter select • esc back")

	return fmt.Sprintf("%s\n%s\n\n%s\n%s", title, subtitle, content, help)
}
