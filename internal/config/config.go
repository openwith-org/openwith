package config

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/openwith-org/openwith/internal/extensions"
)

// Config holds user configuration loaded from disk.
type Config struct {
	ExtraExtensions []extensions.Extension
	EditorPriority  []string          // bundle IDs in preferred order
	EditorArgs      map[string]string // bundleID -> launch args (e.g., "--new-window")
	Theme           string            // "dark" or "light"
}

// maxConfigSize is the maximum allowed config file size (1MB).
const maxConfigSize = 1 << 20

// configPath returns the path to the config file.
func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "openwith", "config.txt"), nil
}

// Load reads the config file. Returns an empty config if the file doesn't exist.
// Format:
//
//	# Comments start with #
//	[extensions]
//	.vue = Web
//	.svelte = Web
//
//	[editor-priority]
//	com.microsoft.VSCode
//	dev.zed.Zed
func Load() Config {
	cfg := Config{
		EditorArgs: make(map[string]string),
	}

	path, err := configPath()
	if err != nil {
		return cfg
	}
	f, err := os.Open(path)
	if err != nil {
		return cfg
	}
	defer f.Close()

	// Reject excessively large config files
	if info, err := f.Stat(); err == nil && info.Size() > maxConfigSize {
		return cfg
	}

	return ParseConfig(f)
}

// ParseConfig parses config from a reader.
func ParseConfig(r io.Reader) Config {
	cfg := Config{
		EditorArgs: make(map[string]string),
	}

	section := ""
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.ToLower(strings.Trim(line, "[]"))
			continue
		}

		switch section {
		case "extensions":
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}
			ext := strings.TrimSpace(parts[0])
			cat := strings.TrimSpace(parts[1])
			if !strings.HasPrefix(ext, ".") {
				ext = "." + ext
			}
			cfg.ExtraExtensions = append(cfg.ExtraExtensions, extensions.Extension{
				Ext:      ext,
				Category: extensions.Category(cat),
			})

		case "editor-priority":
			cfg.EditorPriority = append(cfg.EditorPriority, line)

		case "editor-args":
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}
			bundleID := strings.TrimSpace(parts[0])
			args := strings.TrimSpace(parts[1])
			cfg.EditorArgs[bundleID] = args

		case "settings":
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			if key == "theme" {
				cfg.Theme = val
			}
		}
	}

	return cfg
}

// WriteDefault creates a default config file with comments explaining the format.
func WriteDefault() error {
	path, err := configPath()
	if err != nil {
		return err
	}
	if _, err := os.Stat(path); err == nil {
		return nil // already exists
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	content := `# openwith configuration
#
# [settings]
# theme = dark    # "dark" or "light"
#
# Add custom file extensions under [extensions].
# Format: .ext = Category
# Categories: Config, Docs, Scripts, Web, Mobile, Systems, DB, Data, DevOps, Other
#
# [extensions]
# .vue = Web
# .custom = Other
#
# Set preferred editor order under [editor-priority].
# Editors listed first will appear first when cycling with Tab.
# Use bundle IDs (shown in Bulk Mode).
#
# [editor-priority]
# com.microsoft.VSCode
# dev.zed.Zed
#
# Set per-editor launch arguments under [editor-args].
# Format: bundle.id = --flag1 --flag2
#
# [editor-args]
# com.microsoft.VSCode = --new-window
# com.sublimetext.4 = --new-window
`
	return os.WriteFile(path, []byte(content), 0600)
}
