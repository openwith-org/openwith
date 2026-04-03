package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/openwith-org/openwith/internal/config"
)

type SaveProfileModel struct {
	name string
	done bool
	err  error
}

func NewSaveProfileModel(termHeight int) SaveProfileModel {
	return SaveProfileModel{}
}

func (m *SaveProfileModel) SetDone(err error) {
	m.done = true
	m.err = err
}

func (m SaveProfileModel) Update(msg tea.Msg, changes map[string]string) (SaveProfileModel, tea.Cmd) {
	if msg, ok := msg.(tea.KeyPressMsg); ok {
		if m.done {
			switch msg.String() {
			case "enter", "esc":
				return m, func() tea.Msg {
					return NavigateMsg{Screen: screenMenu}
				}
			}
			return m, nil
		}

		key := msg.String()
		switch key {
		case "enter":
			name := strings.TrimSpace(m.name)
			if name == "" {
				return m, nil
			}
			if len(changes) == 0 {
				m.done = true
				m.err = fmt.Errorf("no changes to save — configure associations first")
				return m, nil
			}
			changeCopy := make(map[string]string)
			for k, v := range changes {
				changeCopy[k] = v
			}
			profileName := name
			return m, func() tea.Msg {
				err := config.SaveProfile(profileName, changeCopy)
				return profileSavedMsg{err: err}
			}
		case "esc":
			return m, func() tea.Msg {
				return NavigateMsg{Screen: screenProfiles}
			}
		case "backspace":
			if len(m.name) > 0 {
				m.name = m.name[:len(m.name)-1]
			}
		default:
			if len(key) == 1 && key >= " " {
				// Only allow safe filename characters
				c := key[0]
				if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
					(c >= '0' && c <= '9') || c == '-' || c == '_' {
					m.name += key
				}
			}
		}
	}
	return m, nil
}

func (m SaveProfileModel) View() string {
	title := TitleStyle.Render("Save Profile")

	if m.done {
		var result string
		if m.err != nil {
			result = ErrorStyle.Render(fmt.Sprintf("  Error: %s", m.err.Error()))
		} else {
			result = SuccessStyle.Render(fmt.Sprintf("  Profile \"%s\" saved!", m.name))
		}
		help := HelpStyle.Render("enter/esc return to menu")
		return title + "\n\n" + result + "\n\n" + help
	}

	prompt := SubtitleStyle.Render("Enter a name for this profile:")
	input := fmt.Sprintf("  %s%s",
		AccentStyle.Render(m.name),
		AccentStyle.Render("▏"),
	)
	hint := DimStyle.Render("  (letters, numbers, dashes, underscores)")
	help := HelpStyle.Render("enter save • esc cancel")

	return title + "\n\n" + prompt + "\n\n" + input + "\n" + hint + "\n\n" + help
}
