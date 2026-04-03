package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/openwith-org/openwith/internal/defaults"
	"github.com/openwith-org/openwith/internal/extensions"
	"github.com/openwith-org/openwith/internal/tui"
)

var version = "dev"

func main() {
	dryRun := flag.Bool("dry-run", false, "Show what would change without applying")
	showVersion := flag.Bool("version", false, "Print version and exit")
	completions := flag.String("completions", "", "Generate shell completions (bash, zsh, fish)")
	theme := flag.String("theme", "", "Color theme: dark (default) or light")
	includeMedia := flag.Bool("media", false, "Include media file types (images, video, audio, PDFs)")
	exportFile := flag.String("export", "", "Export current file associations to a file")
	importFile := flag.String("import", "", "Import and apply file associations from a file")
	flag.Parse()

	if *showVersion {
		fmt.Println("openwith", version)
		return
	}

	if *completions != "" {
		printCompletions(*completions)
		return
	}

	if *exportFile != "" {
		if err := runExport(*exportFile); err != nil {
			fmt.Fprintf(os.Stderr, "Export error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if *importFile != "" {
		if err := runImport(*importFile, *dryRun); err != nil {
			fmt.Fprintf(os.Stderr, "Import error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	app := tui.NewApp(*dryRun, *theme, *includeMedia)
	p := tea.NewProgram(app)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printCompletions(shell string) {
	switch shell {
	case "bash":
		fmt.Print(`_openwith() {
    local cur="${COMP_WORDS[COMP_CWORD]}"
    COMPREPLY=($(compgen -W "--dry-run --version --completions --theme --help" -- "$cur"))
}
complete -F _openwith openwith
`)
	case "zsh":
		fmt.Print(`#compdef openwith

_openwith() {
    _arguments \
        '--dry-run[Show what would change without applying]' \
        '--version[Print version and exit]' \
        '--completions[Generate shell completions (bash, zsh, fish)]:shell:(bash zsh fish)' \
        '--theme[Color theme]:theme:(dark light)' \
        '--help[Show help]'
}

_openwith "$@"
`)
	case "fish":
		fmt.Print(`complete -c openwith -l dry-run -d "Show what would change without applying"
complete -c openwith -l version -d "Print version and exit"
complete -c openwith -l completions -d "Generate shell completions" -x -a "bash zsh fish"
complete -c openwith -l theme -d "Color theme" -x -a "dark light"
complete -c openwith -l help -d "Show help"
`)
	default:
		fmt.Fprintf(os.Stderr, "Unknown shell: %s (supported: bash, zsh, fish)\n", shell)
		os.Exit(1)
	}
}

func runExport(path string) error {
	client := defaults.New(false)
	exts := extensions.AllExts()
	defs := client.GetAllDefaults(exts)

	// Build ext -> bundleID mapping from the plist
	// GetAllDefaults resolves to app names, but we need bundle IDs for import.
	// Re-read the raw data for bundle IDs.
	rawDefaults := client.GetAllDefaultsRaw(exts)

	var b strings.Builder
	b.WriteString("# openwith export\n")
	b.WriteString("# Format: .ext = bundle.id  # App Name\n\n")

	sortedExts := make([]string, 0, len(rawDefaults))
	for ext := range rawDefaults {
		sortedExts = append(sortedExts, ext)
	}
	sort.Strings(sortedExts)

	nameByExt := make(map[string]string)
	for _, d := range defs {
		nameByExt[d.Extension] = d.AppName
	}

	for _, ext := range sortedExts {
		bundleID := rawDefaults[ext]
		name := nameByExt[ext]
		if name == "" || name == "unknown" {
			name = bundleID
		}
		fmt.Fprintf(&b, "%s = %s  # %s\n", ext, bundleID, name)
	}

	if path == "-" {
		fmt.Print(b.String())
		return nil
	}
	return os.WriteFile(path, []byte(b.String()), 0600)
}

// parseImportLines parses import file lines from a reader into ext -> bundleID pairs.
func parseImportLines(r io.Reader) (map[string]string, error) {
	changes := make(map[string]string)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Strip inline comments
		if idx := strings.Index(line, "#"); idx > 0 {
			line = strings.TrimSpace(line[:idx])
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		ext := strings.TrimSpace(parts[0])
		bundleID := strings.TrimSpace(parts[1])
		if ext != "" && bundleID != "" {
			changes[ext] = bundleID
		}
	}
	if len(changes) == 0 {
		return nil, fmt.Errorf("no valid associations found")
	}
	return changes, nil
}

func runImport(path string, dryRun bool) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open import file: %w", err)
	}
	defer f.Close()

	changes, err := parseImportLines(f)
	if err != nil {
		return fmt.Errorf("%w in %s", err, path)
	}

	fmt.Printf("Importing %d file associations", len(changes))
	if dryRun {
		fmt.Println(" (dry run)")
	} else {
		fmt.Println()
	}

	client := defaults.New(dryRun)
	results := make(chan defaults.Result)
	go client.ApplyChanges(changes, results)

	successes, failures := 0, 0
	for r := range results {
		if r.Success {
			successes++
			if r.Error != nil {
				fmt.Printf("  ⚠ %s — %s\n", r.Extension, r.Error)
			} else {
				fmt.Printf("  ✓ %s\n", r.Extension)
			}
		} else {
			failures++
			errMsg := ""
			if r.Error != nil {
				errMsg = r.Error.Error()
			}
			fmt.Printf("  ✗ %s — %s\n", r.Extension, errMsg)
		}
	}

	if failures > 0 {
		return fmt.Errorf("%d of %d failed", failures, successes+failures)
	}
	fmt.Printf("All %d associations imported successfully!\n", successes)
	return nil
}
