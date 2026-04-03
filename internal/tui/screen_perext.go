package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/openwith-org/openwith/internal/editors"
	"github.com/openwith-org/openwith/internal/extensions"
)

type perExtRow struct {
	ext      extensions.Extension
	isSep    bool
	sepLabel string
}

type PerExtModel struct {
	rows       []perExtRow
	cursor     int
	editors    []editors.Editor
	defaults   map[string]string // ext -> current app name
	changes    map[string]string // ext -> chosen bundle ID
	editorIdx  map[string]int    // ext -> index into editors slice
	offset     int
	viewHeight int
	width      int
	filter     string
	filtering  bool
	filtered   []int // indices into rows that match the filter
}

func NewPerExtModel(eds []editors.Editor, defs map[string]string, termHeight int) PerExtModel {
	groups := extensions.Grouped()
	var rows []perExtRow
	for _, g := range groups {
		rows = append(rows, perExtRow{isSep: true, sepLabel: string(g.Category)})
		for _, ext := range g.Extensions {
			rows = append(rows, perExtRow{ext: ext})
		}
	}

	vh := termHeight - 9 // title + subtitle + border + pending + help
	if vh < 5 {
		vh = 5
	}

	m := PerExtModel{
		rows:       rows,
		editors:    eds,
		defaults:   defs,
		changes:    make(map[string]string),
		editorIdx:  make(map[string]int),
		viewHeight: vh,
	}

	for i, r := range m.rows {
		if !r.isSep {
			m.cursor = i
			break
		}
	}

	return m
}

// buildFiltered rebuilds the filtered index list based on the current filter string.
func (m *PerExtModel) buildFiltered() {
	if m.filter == "" {
		m.filtered = nil
		return
	}
	query := strings.ToLower(m.filter)
	m.filtered = nil
	for i, r := range m.rows {
		if r.isSep {
			// Include separator if any extension in its category matches
			if m.categoryHasMatch(i, query) {
				m.filtered = append(m.filtered, i)
			}
			continue
		}
		if strings.Contains(strings.ToLower(r.ext.Ext), query) ||
			strings.Contains(strings.ToLower(string(r.ext.Category)), query) {
			m.filtered = append(m.filtered, i)
		}
	}
}

// categoryHasMatch checks if any extension after separator at index sepIdx matches the query.
func (m *PerExtModel) categoryHasMatch(sepIdx int, query string) bool {
	for j := sepIdx + 1; j < len(m.rows); j++ {
		if m.rows[j].isSep {
			break
		}
		if strings.Contains(strings.ToLower(m.rows[j].ext.Ext), query) ||
			strings.Contains(strings.ToLower(string(m.rows[j].ext.Category)), query) {
			return true
		}
	}
	return false
}

// isVisible returns whether a row index is visible given the current filter.
func (m *PerExtModel) isVisible(idx int) bool {
	if len(m.filtered) == 0 && m.filter == "" {
		return true
	}
	for _, fi := range m.filtered {
		if fi == idx {
			return true
		}
	}
	return false
}

func (m *PerExtModel) nextDataRow(dir int) int {
	pos := m.cursor
	for {
		pos += dir
		if pos < 0 || pos >= len(m.rows) {
			return m.cursor
		}
		if !m.rows[pos].isSep && m.isVisible(pos) {
			return pos
		}
	}
}

// ensureCursorOnVisible moves the cursor to the first visible data row if current isn't visible.
func (m *PerExtModel) ensureCursorOnVisible() {
	if m.isVisible(m.cursor) && !m.rows[m.cursor].isSep {
		return
	}
	// Find first visible data row
	for _, idx := range m.filtered {
		if !m.rows[idx].isSep {
			m.cursor = idx
			m.offset = 0
			m.ensureVisible()
			return
		}
	}
}

func (m PerExtModel) pendingCount() int {
	return len(m.changes)
}

func (m PerExtModel) Update(msg tea.Msg) (PerExtModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		key := msg.String()

		// Filter mode input handling
		if m.filtering {
			switch key {
			case "esc":
				m.filtering = false
				m.filter = ""
				m.filtered = nil
				return m, nil
			case "enter":
				m.filtering = false
				if len(m.filtered) > 0 {
					m.ensureCursorOnVisible()
				}
				return m, nil
			case "backspace":
				if len(m.filter) > 0 {
					m.filter = m.filter[:len(m.filter)-1]
					m.buildFiltered()
					if m.filter != "" {
						m.ensureCursorOnVisible()
					} else {
						m.filtered = nil
					}
				}
				return m, nil
			default:
				if len(key) == 1 && key >= " " {
					m.filter += key
					m.buildFiltered()
					m.ensureCursorOnVisible()
					return m, nil
				}
			}
			return m, nil
		}

		// Normal mode
		switch key {
		case "/":
			m.filtering = true
			m.filter = ""
			m.filtered = nil
			return m, nil
		case "up", "k":
			m.cursor = m.nextDataRow(-1)
			m.ensureVisible()
		case "down", "j":
			m.cursor = m.nextDataRow(1)
			m.ensureVisible()
		case "tab":
			if m.cursor < len(m.rows) && !m.rows[m.cursor].isSep && len(m.editors) > 0 {
				ext := m.rows[m.cursor].ext.Ext
				idx := m.editorIdx[ext]
				idx = (idx + 1) % (len(m.editors) + 1)
				m.editorIdx[ext] = idx
				if idx == 0 {
					delete(m.changes, ext)
				} else {
					m.changes[ext] = m.editors[idx-1].BundleID
				}
			}
		case "shift+tab":
			if m.cursor < len(m.rows) && !m.rows[m.cursor].isSep && len(m.editors) > 0 {
				ext := m.rows[m.cursor].ext.Ext
				idx := m.editorIdx[ext]
				idx = (idx - 1 + len(m.editors) + 1) % (len(m.editors) + 1)
				m.editorIdx[ext] = idx
				if idx == 0 {
					delete(m.changes, ext)
				} else {
					m.changes[ext] = m.editors[idx-1].BundleID
				}
			}
		case "enter", "a":
			if m.pendingCount() > 0 {
				return m, func() tea.Msg {
					return NavigateMsg{Screen: screenConfirm}
				}
			}
		case "esc":
			if m.filter != "" {
				m.filter = ""
				m.filtered = nil
				return m, nil
			}
			return m, func() tea.Msg {
				return NavigateMsg{Screen: screenMenu}
			}
		}
	case tea.WindowSizeMsg:
		m.viewHeight = msg.Height - 9
		if m.viewHeight < 5 {
			m.viewHeight = 5
		}
		m.width = msg.Width
	}
	return m, nil
}

func (m *PerExtModel) ensureVisible() {
	if m.cursor < m.offset {
		// Scroll up — include category separator above if present
		m.offset = m.cursor
		if m.offset > 0 && m.rows[m.offset-1].isSep {
			m.offset--
		}
	}
	if m.cursor >= m.offset+m.viewHeight {
		m.offset = m.cursor - m.viewHeight + 1
	}
}

func (m PerExtModel) editorNameForExt(ext string) string {
	if bundleID, ok := m.changes[ext]; ok {
		for _, ed := range m.editors {
			if ed.BundleID == bundleID {
				return ed.Name
			}
		}
	}
	return ""
}

func (m PerExtModel) View() string {
	title := TitleStyle.Render("Per-Extension Mode")

	subtitleText := "Tab to cycle editors • / to search • Enter to apply"
	if m.filtering {
		subtitleText = "Type to filter • Enter to confirm • Esc to cancel"
	}
	subtitle := SubtitleStyle.Render(subtitleText)

	// Filter bar
	filterBar := ""
	if m.filtering || m.filter != "" {
		filterIcon := AccentStyle.Render("🔍 ")
		filterText := m.filter
		if m.filtering {
			filterText += "▏" // cursor
		}
		matchCount := 0
		if m.filter != "" {
			for _, idx := range m.filtered {
				if !m.rows[idx].isSep {
					matchCount++
				}
			}
			filterBar = fmt.Sprintf("  %s%s  %s",
				filterIcon,
				AccentStyle.Render(filterText),
				DimStyle.Render(fmt.Sprintf("(%d matches)", matchCount)),
			)
		} else {
			filterBar = fmt.Sprintf("  %s%s", filterIcon, AccentStyle.Render(filterText))
		}
	}

	var lines []string
	for i, row := range m.rows {
		if !m.isVisible(i) {
			continue
		}

		if row.isSep {
			lines = append(lines, CategoryStyle.Render("── "+row.sepLabel+" ──"))
			continue
		}

		ext := row.ext.Ext
		current := m.defaults[ext]
		if current == "" {
			current = "unknown"
		}
		newEditor := m.editorNameForExt(ext)

		isSelected := i == m.cursor
		cursor := "  "
		extStyle := InactiveItemStyle
		if isSelected {
			cursor = AccentStyle.Render("▸ ")
			extStyle = ActiveItemStyle
		}

		newCol := DimStyle.Render("---")
		if newEditor != "" {
			newCol = AccentStyle.Render(newEditor)
		}

		line := fmt.Sprintf("%s%-14s %s%-20s%s %s",
			cursor,
			extStyle.Render(ext),
			DimStyle.Render(""),
			DimStyle.Render(current),
			DimStyle.Render(" → "),
			newCol,
		)
		lines = append(lines, line)
	}

	// Viewport
	visible := lines
	if m.offset > 0 && m.offset < len(lines) {
		visible = lines[m.offset:]
	}
	if len(visible) > m.viewHeight {
		visible = visible[:m.viewHeight]
	}

	// Scroll indicators
	scrollUp := m.offset > 0
	scrollDown := m.offset+m.viewHeight < len(lines)
	topIndicator := ""
	botIndicator := ""
	if scrollUp {
		topIndicator = DimStyle.Render("  ▲ more above")
	}
	if scrollDown {
		botIndicator = DimStyle.Render("  ▼ more below")
	}

	content := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Purple).
		Padding(0, 1).
		Render(strings.Join(visible, "\n"))

	pending := ""
	if n := m.pendingCount(); n > 0 {
		pending = WarnStyle.Render(fmt.Sprintf("  %d change(s) pending", n))
	}

	helpText := "↑/↓ navigate • tab/shift+tab cycle editor • / search • enter apply • esc back"
	if m.filter != "" && !m.filtering {
		helpText = "↑/↓ navigate • tab/shift+tab cycle editor • / search • esc clear filter • enter apply"
	}
	help := HelpStyle.Render(helpText)

	var b strings.Builder
	b.WriteString(title + "\n" + subtitle + "\n")
	if filterBar != "" {
		b.WriteString(filterBar + "\n")
	}
	if topIndicator != "" {
		b.WriteString(topIndicator + "\n")
	}
	b.WriteString("\n" + content + "\n")
	if botIndicator != "" {
		b.WriteString(botIndicator + "\n")
	}
	if pending != "" {
		b.WriteString(pending + "\n")
	}
	b.WriteString(help)
	return b.String()
}
