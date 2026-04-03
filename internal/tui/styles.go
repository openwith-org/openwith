package tui

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
)

func joinLines(lines []string) string {
	return strings.Join(lines, "\n")
}

// Theme defines a color palette.
type Theme struct {
	Primary  color.Color
	Accent   color.Color
	Success  color.Color
	Warning  color.Color
	Error    color.Color
	Text     color.Color
	TextDim  color.Color
	TextMute color.Color
	BG       color.Color
}

var DarkTheme = Theme{
	Primary:  lipgloss.Color("99"),  // purple
	Accent:   lipgloss.Color("86"),  // cyan
	Success:  lipgloss.Color("78"),  // green
	Warning:  lipgloss.Color("220"), // yellow
	Error:    lipgloss.Color("196"), // red
	Text:     lipgloss.Color("255"), // white
	TextDim:  lipgloss.Color("245"), // gray
	TextMute: lipgloss.Color("238"), // dark gray
	BG:       lipgloss.Color("0"),   // black
}

var LightTheme = Theme{
	Primary:  lipgloss.Color("55"),  // deep purple
	Accent:   lipgloss.Color("31"),  // dark cyan
	Success:  lipgloss.Color("28"),  // dark green
	Warning:  lipgloss.Color("172"), // dark yellow/orange
	Error:    lipgloss.Color("160"), // dark red
	Text:     lipgloss.Color("0"),   // black
	TextDim:  lipgloss.Color("242"), // mid gray
	TextMute: lipgloss.Color("248"), // light gray
	BG:       lipgloss.Color("255"), // white
}

// Active color variables — these are set by ApplyTheme.
var (
	Purple   color.Color = DarkTheme.Primary
	Cyan     color.Color = DarkTheme.Accent
	Green    color.Color = DarkTheme.Success
	Yellow   color.Color = DarkTheme.Warning
	Red      color.Color = DarkTheme.Error
	Gray     color.Color = DarkTheme.TextDim
	DarkGray color.Color = DarkTheme.TextMute
	White    color.Color = DarkTheme.Text
)

// Style variables — rebuilt by ApplyTheme.
var (
	TitleStyle    lipgloss.Style
	SubtitleStyle lipgloss.Style
	BoxStyle      lipgloss.Style

	ActiveItemStyle   lipgloss.Style
	InactiveItemStyle lipgloss.Style

	SuccessStyle  lipgloss.Style
	ErrorStyle    lipgloss.Style
	AccentStyle   lipgloss.Style
	DimStyle      lipgloss.Style
	WarnStyle     lipgloss.Style
	CategoryStyle lipgloss.Style
	HelpStyle     lipgloss.Style
	BadgeStyle    lipgloss.Style
)

func init() {
	ApplyTheme(DarkTheme)
}

// ApplyTheme sets all color variables and rebuilds styles from a theme.
func ApplyTheme(t Theme) {
	Purple = t.Primary
	Cyan = t.Accent
	Green = t.Success
	Yellow = t.Warning
	Red = t.Error
	Gray = t.TextDim
	DarkGray = t.TextMute
	White = t.Text

	TitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(Purple).
		Padding(1, 2)

	SubtitleStyle = lipgloss.NewStyle().
		Foreground(Gray).
		Padding(0, 2)

	BoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Purple).
		Padding(1, 2)

	ActiveItemStyle = lipgloss.NewStyle().
		Foreground(Cyan).
		Bold(true)

	InactiveItemStyle = lipgloss.NewStyle().
		Foreground(Gray)

	SuccessStyle = lipgloss.NewStyle().Foreground(Green)
	ErrorStyle = lipgloss.NewStyle().Foreground(Red)
	AccentStyle = lipgloss.NewStyle().Foreground(Cyan).Bold(true)
	DimStyle = lipgloss.NewStyle().Foreground(DarkGray)
	WarnStyle = lipgloss.NewStyle().Foreground(Yellow).Bold(true)

	CategoryStyle = lipgloss.NewStyle().
		Foreground(Purple).
		Bold(true).
		PaddingLeft(1)

	HelpStyle = lipgloss.NewStyle().
		Foreground(DarkGray).
		Padding(1, 2)

	BadgeStyle = lipgloss.NewStyle().
		Background(Yellow).
		Foreground(t.BG).
		Bold(true).
		Padding(0, 1)
}

// ThemeByName returns a theme by name ("dark" or "light").
func ThemeByName(name string) Theme {
	switch strings.ToLower(name) {
	case "light":
		return LightTheme
	default:
		return DarkTheme
	}
}
