package tui

import (
	"fmt"
	"sort"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type ConfirmModel struct {
	changes    map[string]string // ext -> bundleID
	defaults   map[string]string // ext -> current app name
	editors    map[string]string // bundleID -> editor name
	dryRun     bool
	offset     int
	viewHeight int
	width      int
}

func NewConfirmModel(changes map[string]string, defs map[string]string, editorNames map[string]string, dryRun bool, termHeight int) ConfirmModel {
	vh := termHeight - 10
	if vh < 5 {
		vh = 5
	}
	return ConfirmModel{
		changes:    changes,
		defaults:   defs,
		editors:    editorNames,
		dryRun:     dryRun,
		viewHeight: vh,
	}
}

func (m ConfirmModel) Update(msg tea.Msg) (ConfirmModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter", "y":
			return m, func() tea.Msg {
				return NavigateMsg{Screen: screenApply}
			}
		case "esc", "n":
			return m, func() tea.Msg {
				return NavigateMsg{Screen: screenMenu}
			}
		case "down", "j":
			maxOffset := len(m.changes) + 2 - m.viewHeight // +2 for header+separator
			if maxOffset < 0 {
				maxOffset = 0
			}
			if m.offset < maxOffset {
				m.offset++
			}
		case "up", "k":
			if m.offset > 0 {
				m.offset--
			}
		case "s":
			return m, func() tea.Msg {
				return NavigateMsg{Screen: screenSaveProfile}
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

func (m ConfirmModel) View() string {
	titleText := "Confirm Changes"
	if m.dryRun {
		titleText = BadgeStyle.Render(" DRY RUN ") + "  " + titleText
	}
	title := TitleStyle.Render(titleText)

	exts := make([]string, 0, len(m.changes))
	for ext := range m.changes {
		exts = append(exts, ext)
	}
	sort.Strings(exts)

	var rows []string
	header := fmt.Sprintf("  %-14s %-20s    %s", "Extension", "Current", "New")
	rows = append(rows, lipgloss.NewStyle().Bold(true).Foreground(White).Render(header))
	rows = append(rows, DimStyle.Render("  "+strings.Repeat("─", 56)))

	for _, ext := range exts {
		bundleID := m.changes[ext]
		current := m.defaults[ext]
		if current == "" {
			current = "unknown"
		}
		newName := m.editors[bundleID]
		if newName == "" {
			newName = bundleID
		}

		// Color-code: red strikethrough for old, green for new
		currentStyled := ErrorStyle.Render(current)
		newStyled := SuccessStyle.Render(newName)
		arrow := DimStyle.Render(" → ")
		if current == newName {
			// No actual change — dim both
			currentStyled = DimStyle.Render(current)
			newStyled = DimStyle.Render(newName)
			arrow = DimStyle.Render(" = ")
		}

		row := fmt.Sprintf("  %-14s %s%s%s",
			AccentStyle.Render(ext),
			currentStyled,
			arrow,
			newStyled,
		)
		rows = append(rows, row)
	}

	// Viewport
	visible := rows
	if m.offset > 0 && m.offset < len(rows) {
		visible = rows[m.offset:]
	}
	if len(visible) > m.viewHeight {
		visible = visible[:m.viewHeight]
	}

	scrollUp := m.offset > 0
	scrollDown := m.offset+m.viewHeight < len(rows)

	content := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Purple).
		Padding(1, 2).
		Render(strings.Join(visible, "\n"))

	summary := SubtitleStyle.Render(fmt.Sprintf("%d file type(s) will be updated", len(m.changes)))

	var scrollHint string
	if scrollUp || scrollDown {
		parts := []string{}
		if scrollUp {
			parts = append(parts, "▲ more above")
		}
		if scrollDown {
			parts = append(parts, "▼ more below")
		}
		scrollHint = DimStyle.Render("  " + strings.Join(parts, " • "))
	}

	help := HelpStyle.Render("↑/↓ scroll • enter confirm • s save as profile • esc cancel")

	var b strings.Builder
	b.WriteString(title + "\n" + summary + "\n\n" + content + "\n")
	if scrollHint != "" {
		b.WriteString(scrollHint + "\n")
	}
	b.WriteString("\n" + help)
	return b.String()
}
