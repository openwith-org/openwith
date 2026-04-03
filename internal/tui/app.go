package tui

import (
	"maps"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/openwith-org/openwith/internal/config"
	"github.com/openwith-org/openwith/internal/defaults"
	"github.com/openwith-org/openwith/internal/editors"
	"github.com/openwith-org/openwith/internal/extensions"
)

type screen int

const (
	screenMenu screen = iota
	screenBulk
	screenPerExt
	screenConfirm
	screenApply
	screenRevert
	screenProfiles
	screenSaveProfile
	screenBrowser
)

// Messages
type NavigateMsg struct{ Screen screen }
type BulkSelectMsg struct{ Editor editors.Editor }
type EditorsDetectedMsg struct{ Editors []editors.Editor }
type DefaultsLoadedMsg struct{ Defaults map[string]string }

type applyBatchMsg struct {
	results []defaults.Result
}

type revertDoneMsg struct {
	err error
}

type profileLoadedMsg struct {
	changes map[string]string
}

type profileSavedMsg struct {
	err error
}

type transitionTickMsg struct{}

type App struct {
	current     screen
	menu        MenuModel
	bulk        BulkModel
	perExt      PerExtModel
	confirm     ConfirmModel
	apply       ApplyModel
	revert      RevertModel
	profiles    ProfilesModel
	saveProfile SaveProfileModel
	browser     BrowserModel

	editors          []editors.Editor
	defaults         map[string]string // ext -> current app name
	changes          map[string]string // ext -> chosen bundleID
	client           *defaults.Client
	cfg              config.Config
	dryRun           bool
	includeMedia     bool
	loading          bool
	width            int
	height           int
	transitionFrames int // remaining transition frames (0 = no transition)
}

func NewApp(dryRun bool, themeOverride string, includeMedia bool) *App {
	cfg := config.Load()

	// Apply theme: CLI flag takes precedence over config
	themeName := cfg.Theme
	if themeOverride != "" {
		themeName = themeOverride
	}
	if themeName != "" {
		ApplyTheme(ThemeByName(themeName))
	}

	// Merge custom extensions from config
	if len(cfg.ExtraExtensions) > 0 {
		extensions.Merge(cfg.ExtraExtensions)
	}

	// Enable media extensions if requested
	if includeMedia {
		extensions.EnableMediaExtensions()
	}

	// Create default config file if it doesn't exist
	_ = config.WriteDefault()

	return &App{
		current:      screenMenu,
		menu:         NewMenuModel(),
		client:       defaults.New(dryRun),
		cfg:          cfg,
		dryRun:       dryRun,
		includeMedia: includeMedia,
		defaults:     make(map[string]string),
		changes:      make(map[string]string),
		loading:      true,
	}
}

func (a *App) Init() tea.Cmd {
	includeMedia := a.includeMedia
	return tea.Batch(
		func() tea.Msg {
			eds := editors.DetectInstalled()
			if includeMedia {
				mediaEds := editors.DetectInstalledMedia()
				eds = append(eds, mediaEds...)
			}
			return EditorsDetectedMsg{Editors: eds}
		},
		func() tea.Msg {
			exts := extensions.AllExts()
			defs := a.client.GetAllDefaults(exts)
			m := make(map[string]string)
			for _, d := range defs {
				m[d.Extension] = d.AppName
			}
			return DefaultsLoadedMsg{Defaults: m}
		},
	)
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" || (msg.String() == "q" && a.current == screenMenu) {
			return a, tea.Quit
		}

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height

	case EditorsDetectedMsg:
		a.editors = editors.SortByPriority(msg.Editors, a.cfg.EditorPriority)
		a.bulk = NewBulkModel(a.editors, a.height)
		if len(a.defaults) > 0 {
			a.loading = false
			a.perExt = NewPerExtModel(a.editors, a.defaults, a.height)
		}
		return a, nil

	case DefaultsLoadedMsg:
		a.defaults = msg.Defaults
		if len(a.editors) > 0 {
			a.loading = false
			a.perExt = NewPerExtModel(a.editors, a.defaults, a.height)
		}
		return a, nil

	case NavigateMsg:
		a.current = msg.Screen
		var extraCmd tea.Cmd
		switch msg.Screen {
		case screenConfirm:
			editorNames := make(map[string]string)
			for _, ed := range a.editors {
				editorNames[ed.BundleID] = ed.Name
			}
			a.confirm = NewConfirmModel(a.changes, a.defaults, editorNames, a.dryRun, a.height)
		case screenApply:
			a.apply = NewApplyModel(len(a.changes), a.dryRun, a.height)
			extraCmd = a.applyChanges()
		case screenMenu:
			a.changes = make(map[string]string)
			a.perExt = NewPerExtModel(a.editors, a.defaults, a.height)
		case screenRevert:
			a.revert = NewRevertModel(a.client, a.dryRun, a.height)
		case screenProfiles:
			editorNames := make(map[string]string)
			for _, ed := range a.editors {
				editorNames[ed.BundleID] = ed.Name
			}
			a.profiles = NewProfilesModel(a.changes, editorNames, a.height)
		case screenSaveProfile:
			a.saveProfile = NewSaveProfileModel(a.height)
		case screenBrowser:
			a.browser = NewBrowserModel(a.client, a.dryRun, a.height)
		}
		transCmd := a.startTransition()
		if extraCmd != nil {
			return a, tea.Batch(transCmd, extraCmd)
		}
		return a, transCmd

	case BulkSelectMsg:
		a.changes = make(map[string]string)
		for _, ext := range extensions.All {
			a.changes[ext.Ext] = msg.Editor.BundleID
		}
		a.current = screenConfirm
		editorNames := make(map[string]string)
		for _, ed := range a.editors {
			editorNames[ed.BundleID] = ed.Name
		}
		a.confirm = NewConfirmModel(a.changes, a.defaults, editorNames, a.dryRun, a.height)
		return a, a.startTransition()

	case applyBatchMsg:
		for _, r := range msg.results {
			a.apply.AddResult(r)
		}
		return a, nil

	case revertDoneMsg:
		a.revert.SetDone(msg.err)
		return a, nil

	case profileLoadedMsg:
		a.changes = msg.changes
		a.current = screenConfirm
		editorNames := make(map[string]string)
		for _, ed := range a.editors {
			editorNames[ed.BundleID] = ed.Name
		}
		a.confirm = NewConfirmModel(a.changes, a.defaults, editorNames, a.dryRun, a.height)
		return a, a.startTransition()

	case profileSavedMsg:
		a.saveProfile.SetDone(msg.err)
		return a, nil

	case browserSetMsg:
		a.browser.SetDone(msg.err)
		return a, nil

	case retryMsg:
		// Rebuild changes map with only failed extensions
		failedExts := a.apply.FailedExtensions()
		retryChanges := make(map[string]string)
		for _, ext := range failedExts {
			if bid, ok := a.changes[ext]; ok {
				retryChanges[ext] = bid
			}
		}
		if len(retryChanges) > 0 {
			a.changes = retryChanges
			a.apply = NewApplyModel(len(retryChanges), a.dryRun, a.height)
			return a, a.applyChanges()
		}
		return a, nil

	case transitionTickMsg:
		if a.transitionFrames > 0 {
			a.transitionFrames--
			if a.transitionFrames > 0 {
				return a, tea.Tick(30*time.Millisecond, func(time.Time) tea.Msg {
					return transitionTickMsg{}
				})
			}
		}
		return a, nil
	}

	// Delegate to active screen
	var cmd tea.Cmd
	switch a.current {
	case screenMenu:
		a.menu, cmd = a.menu.Update(msg)
	case screenBulk:
		a.bulk, cmd = a.bulk.Update(msg)
	case screenPerExt:
		a.perExt, cmd = a.perExt.Update(msg)
		a.changes = maps.Clone(a.perExt.changes)
	case screenConfirm:
		a.confirm, cmd = a.confirm.Update(msg)
	case screenApply:
		a.apply, cmd = a.apply.Update(msg)
	case screenRevert:
		a.revert, cmd = a.revert.Update(msg)
	case screenProfiles:
		a.profiles, cmd = a.profiles.Update(msg)
	case screenSaveProfile:
		a.saveProfile, cmd = a.saveProfile.Update(msg, a.changes)
	case screenBrowser:
		a.browser, cmd = a.browser.Update(msg)
	}

	return a, cmd
}

const transitionDuration = 3 // number of frames

func (a *App) startTransition() tea.Cmd {
	a.transitionFrames = transitionDuration
	return tea.Tick(30*time.Millisecond, func(time.Time) tea.Msg {
		return transitionTickMsg{}
	})
}

func (a *App) view(content string) tea.View {
	// Apply fade-in during transitions
	if a.transitionFrames > 0 {
		opacity := float64(transitionDuration-a.transitionFrames) / float64(transitionDuration)
		content = applyFade(content, opacity)
	}
	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

// applyFade dims content by replacing foreground with a muted color at low opacity.
func applyFade(content string, opacity float64) string {
	if opacity >= 1.0 {
		return content
	}
	// Simple fade: at low opacity show blank lines, at mid opacity show dimmed text
	if opacity < 0.3 {
		lines := strings.Split(content, "\n")
		blank := make([]string, len(lines))
		for i := range blank {
			blank[i] = ""
		}
		return strings.Join(blank, "\n")
	}
	// At mid opacity, render everything in dim
	return lipgloss.NewStyle().Foreground(DarkGray).Render(content)
}

func (a *App) View() tea.View {
	if a.loading {
		return a.view(TitleStyle.Render("openwith") + "\n" +
			SubtitleStyle.Render("Detecting installed editors...") + "\n")
	}

	if len(a.editors) == 0 {
		return a.view(TitleStyle.Render("openwith") + "\n" +
			ErrorStyle.Render("  No supported editors found on this system.") + "\n" +
			HelpStyle.Render("q quit") + "\n")
	}

	switch a.current {
	case screenMenu:
		return a.view(a.menu.View())
	case screenBulk:
		return a.view(a.bulk.View())
	case screenPerExt:
		return a.view(a.perExt.View())
	case screenConfirm:
		return a.view(a.confirm.View())
	case screenApply:
		return a.view(a.apply.View())
	case screenRevert:
		return a.view(a.revert.View())
	case screenProfiles:
		return a.view(a.profiles.View())
	case screenSaveProfile:
		return a.view(a.saveProfile.View())
	case screenBrowser:
		return a.view(a.browser.View())
	}
	return a.view("")
}

func (a *App) applyChanges() tea.Cmd {
	changes := make(map[string]string)
	for k, v := range a.changes {
		changes[k] = v
	}
	client := a.client

	return func() tea.Msg {
		results := make(chan defaults.Result)
		go client.ApplyChanges(changes, results)

		var allResults []defaults.Result
		for r := range results {
			allResults = append(allResults, r)
		}
		return applyBatchMsg{results: allResults}
	}
}
