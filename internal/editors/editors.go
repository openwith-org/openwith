package editors

import (
	"os/exec"
	"strings"
	"sync"

	"github.com/openwith-org/openwith/internal/defaults"
)

type Editor struct {
	Name      string
	Key       string
	BundleID  string
	Installed bool
}

// AppKind categorizes what type of app this is.
type AppKind string

const (
	KindEditor AppKind = "editor"
	KindMedia  AppKind = "media"
)

var Registry = []Editor{
	{Name: "Antigravity", Key: "antigravity", BundleID: "com.google.antigravity"},
	{Name: "Windsurf", Key: "windsurf", BundleID: "com.exafunction.windsurf"},
	{Name: "VS Code", Key: "vscode", BundleID: "com.microsoft.VSCode"},
	{Name: "Cursor", Key: "cursor", BundleID: "com.todesktop.230313mzl4w4u92"},
	{Name: "Zed", Key: "zed", BundleID: "dev.zed.Zed"},
	{Name: "Sublime Text", Key: "sublime", BundleID: "com.sublimetext.4"},
	{Name: "Nova", Key: "nova", BundleID: "com.panic.Nova"},
	{Name: "BBEdit", Key: "bbedit", BundleID: "com.barebones.bbedit"},
	{Name: "TextMate", Key: "textmate", BundleID: "com.macromates.TextMate"},
	{Name: "Xcode", Key: "xcode", BundleID: "com.apple.dt.Xcode"},
	{Name: "IntelliJ IDEA", Key: "intellij", BundleID: "com.jetbrains.intellij"},
	{Name: "GoLand", Key: "goland", BundleID: "com.jetbrains.goland"},
	{Name: "PyCharm", Key: "pycharm", BundleID: "com.jetbrains.pycharm"},
	{Name: "WebStorm", Key: "webstorm", BundleID: "com.jetbrains.WebStorm"},
	{Name: "VimR", Key: "vimr", BundleID: "com.qvacua.VimR"},
	{Name: "Emacs", Key: "emacs", BundleID: "org.gnu.Emacs"},
	{Name: "Lapce", Key: "lapce", BundleID: "dev.lapce.lapce"},
	{Name: "Fleet", Key: "fleet", BundleID: "com.jetbrains.fleet"},
	{Name: "CotEditor", Key: "coteditor", BundleID: "com.coteditor.CotEditor"},
}

// MediaApps contains non-editor applications for media and document file types.
var MediaApps = []Editor{
	// Video players
	{Name: "VLC", Key: "vlc", BundleID: "org.videolan.vlc"},
	{Name: "IINA", Key: "iina", BundleID: "com.colliderli.iina"},
	{Name: "QuickTime Player", Key: "quicktime", BundleID: "com.apple.QuickTimePlayerX"},
	// Image viewers/editors
	{Name: "Preview", Key: "preview", BundleID: "com.apple.Preview"},
	{Name: "Pixelmator Pro", Key: "pixelmator", BundleID: "com.pixelmatorteam.pixelmator.x"},
	{Name: "Affinity Photo", Key: "affinity-photo", BundleID: "com.seriflabs.affinityphoto2"},
	{Name: "Adobe Photoshop", Key: "photoshop", BundleID: "com.adobe.Photoshop"},
	// PDF viewers
	{Name: "Skim", Key: "skim", BundleID: "net.sourceforge.skim-app.skim"},
	{Name: "PDF Expert", Key: "pdf-expert", BundleID: "com.readdle.PDFExpert-Mac"},
	// Music
	{Name: "Apple Music", Key: "music", BundleID: "com.apple.Music"},
	{Name: "Spotify", Key: "spotify", BundleID: "com.spotify.client"},
}

// Browsers contains common macOS web browsers.
var Browsers = []Editor{
	{Name: "Safari", Key: "safari", BundleID: "com.apple.Safari"},
	{Name: "Google Chrome", Key: "chrome", BundleID: "com.google.Chrome"},
	{Name: "Firefox", Key: "firefox", BundleID: "org.mozilla.firefox"},
	{Name: "Arc", Key: "arc", BundleID: "company.thebrowser.Browser"},
	{Name: "Brave", Key: "brave", BundleID: "com.brave.Browser"},
	{Name: "Microsoft Edge", Key: "edge", BundleID: "com.microsoft.edgemac"},
	{Name: "Vivaldi", Key: "vivaldi", BundleID: "com.vivaldi.Vivaldi"},
	{Name: "Opera", Key: "opera", BundleID: "com.operasoftware.Opera"},
	{Name: "Orion", Key: "orion", BundleID: "com.kagi.kagimacOS"},
	{Name: "Zen", Key: "zen", BundleID: "io.github.nicothin.zen-browser"},
}

// DetectInstalledBrowsers detects installed web browsers.
func DetectInstalledBrowsers() []Editor {
	return DetectInstalledFrom(Browsers)
}

// SortByPriority reorders editors so that bundle IDs in priority appear first, in order.
func SortByPriority(eds []Editor, priority []string) []Editor {
	if len(priority) == 0 {
		return eds
	}
	rank := make(map[string]int)
	for i, bid := range priority {
		rank[bid] = i + 1
	}

	prioritized := make([]Editor, 0, len(eds))
	rest := make([]Editor, 0, len(eds))
	for _, ed := range eds {
		if _, ok := rank[ed.BundleID]; ok {
			prioritized = append(prioritized, ed)
		} else {
			rest = append(rest, ed)
		}
	}

	// Sort prioritized by rank
	for i := 0; i < len(prioritized); i++ {
		for j := i + 1; j < len(prioritized); j++ {
			if rank[prioritized[i].BundleID] > rank[prioritized[j].BundleID] {
				prioritized[i], prioritized[j] = prioritized[j], prioritized[i]
			}
		}
	}

	return append(prioritized, rest...)
}

func DetectInstalled() []Editor {
	// Dynamically resolve version-specific bundle IDs before detection
	resolved := make([]Editor, len(Registry))
	copy(resolved, Registry)
	for i := range resolved {
		if resolved[i].Key == "sublime" {
			resolved[i].BundleID = resolveSublimeBundleID()
		}
	}
	return DetectInstalledFrom(resolved)
}

// resolveSublimeBundleID checks for installed Sublime Text versions newest-first.
func resolveSublimeBundleID() string {
	for _, v := range []string{"5", "4", "3"} {
		bid := "com.sublimetext." + v
		if !defaults.ValidBundleIDPattern.MatchString(bid) {
			continue
		}
		out, err := exec.Command("mdfind",
			"kMDItemCFBundleIdentifier == '"+bid+"'").Output()
		if err == nil && len(strings.TrimSpace(string(out))) > 0 {
			return bid
		}
	}
	return "com.sublimetext.4"
}

// DetectInstalledMedia detects installed media/document apps.
func DetectInstalledMedia() []Editor {
	return DetectInstalledFrom(MediaApps)
}

// DetectInstalledFrom checks which apps from a given registry are installed.
func DetectInstalledFrom(registry []Editor) []Editor {
	var mu sync.Mutex
	var wg sync.WaitGroup
	installed := make([]Editor, 0)

	sem := make(chan struct{}, 8)

	for _, ed := range registry {
		wg.Add(1)
		go func(e Editor) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			if !defaults.ValidBundleIDPattern.MatchString(e.BundleID) {
				return
			}
			out, err := exec.Command("mdfind",
				"kMDItemCFBundleIdentifier == '"+e.BundleID+"'").Output()
			if err == nil && len(strings.TrimSpace(string(out))) > 0 {
				e.Installed = true
				mu.Lock()
				installed = append(installed, e)
				mu.Unlock()
			}
		}(ed)
	}

	wg.Wait()
	return installed
}
