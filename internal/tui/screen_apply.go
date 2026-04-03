package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/openwith-org/openwith/internal/defaults"
)

type retryMsg struct{}

type ApplyModel struct {
	results     []defaults.Result
	total       int
	done        bool
	dryRun      bool
	offset      int
	viewHeight  int
	width       int
	hasFailures bool
}

func NewApplyModel(total int, dryRun bool, termHeight int) ApplyModel {
	vh := termHeight - 10
	if vh < 5 {
		vh = 5
	}
	return ApplyModel{
		total:      total,
		dryRun:     dryRun,
		viewHeight: vh,
	}
}

func (m *ApplyModel) AddResult(r defaults.Result) {
	m.results = append(m.results, r)
	if !r.Success {
		m.hasFailures = true
	}
	if len(m.results) >= m.total {
		m.done = true
	}
	// Auto-scroll to bottom as results come in
	if len(m.results) > m.viewHeight {
		m.offset = len(m.results) - m.viewHeight
	}
}

// FailedExtensions returns the extensions that failed.
func (m ApplyModel) FailedExtensions() []string {
	var failed []string
	for _, r := range m.results {
		if !r.Success {
			failed = append(failed, r.Extension)
		}
	}
	return failed
}

func (m ApplyModel) Update(msg tea.Msg) (ApplyModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if m.done {
			switch msg.String() {
			case "q", "esc", "enter":
				return m, func() tea.Msg {
					return NavigateMsg{Screen: screenMenu}
				}
			case "r":
				if m.hasFailures {
					return m, func() tea.Msg {
						return retryMsg{}
					}
				}
			}
		}
		switch msg.String() {
		case "down", "j":
			maxOffset := len(m.results) - m.viewHeight
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

func (m ApplyModel) View() string {
	titleText := "Applying Changes"
	if m.done {
		titleText = "Changes Applied"
	}
	if m.dryRun {
		titleText = BadgeStyle.Render(" DRY RUN ") + "  " + titleText
	}
	title := TitleStyle.Render(titleText)

	var lines []string
	for _, r := range m.results {
		switch {
		case !r.Success:
			icon := ErrorStyle.Render("✗")
			errMsg := ""
			if r.Error != nil {
				errMsg = ErrorStyle.Render(" — " + r.Error.Error())
			}
			lines = append(lines, fmt.Sprintf("  %s %s%s", icon, r.Extension, errMsg))
		case r.Error != nil:
			// Success with warning (e.g., lsd restart failed)
			icon := WarnStyle.Render("⚠")
			lines = append(lines, fmt.Sprintf("  %s %s  %s", icon, r.Extension,
				WarnStyle.Render(r.Error.Error())))
		default:
			icon := SuccessStyle.Render("✓")
			lines = append(lines, fmt.Sprintf("  %s %s", icon, r.Extension))
		}
	}

	if !m.done {
		remaining := m.total - len(m.results)
		lines = append(lines, DimStyle.Render(fmt.Sprintf("  ... %d remaining", remaining)))
	}

	// Viewport
	visible := lines
	if m.offset > 0 && m.offset < len(lines) {
		visible = lines[m.offset:]
	}
	if len(visible) > m.viewHeight {
		visible = visible[:m.viewHeight]
	}

	scrollUp := m.offset > 0
	scrollDown := m.offset+m.viewHeight < len(lines)

	content := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Purple).
		Padding(1, 2).
		Render(strings.Join(visible, "\n"))

	var summary string
	if m.done {
		successes := 0
		warnings := 0
		failures := 0
		for _, r := range m.results {
			switch {
			case r.Success && r.Error != nil:
				warnings++
			case r.Success:
				successes++
			default:
				failures++
			}
		}
		switch {
		case failures == 0 && warnings == 0:
			summary = SuccessStyle.Render(fmt.Sprintf("  All %d file types updated successfully!", successes))
		case failures == 0:
			summary = WarnStyle.Render(fmt.Sprintf("  %d updated with %d warning(s)", successes+warnings, warnings))
		default:
			summary = ErrorStyle.Render(fmt.Sprintf("  %d succeeded, %d failed", successes+warnings, failures))
		}
	} else {
		summary = SubtitleStyle.Render(fmt.Sprintf("  %d / %d", len(m.results), m.total))
	}

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

	help := ""
	if m.done {
		if m.hasFailures {
			help = HelpStyle.Render("r retry failed • enter/esc return to menu • q quit")
		} else {
			help = HelpStyle.Render("enter/esc return to menu • q quit")
		}
	}

	var b strings.Builder
	b.WriteString(title + "\n" + summary + "\n\n" + content + "\n")
	if scrollHint != "" {
		b.WriteString(scrollHint + "\n")
	}
	if help != "" {
		b.WriteString(help)
	}
	return b.String()
}
