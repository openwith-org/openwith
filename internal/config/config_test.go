package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/openwith-org/openwith/internal/extensions"
)

func TestParseConfigEmpty(t *testing.T) {
	cfg := ParseConfig(strings.NewReader(""))
	if len(cfg.ExtraExtensions) != 0 || len(cfg.EditorPriority) != 0 {
		t.Error("empty input should produce empty config")
	}
	if cfg.EditorArgs == nil {
		t.Error("EditorArgs should be initialized")
	}
}

func TestParseConfigCommentsOnly(t *testing.T) {
	input := "# this is a comment\n# another comment\n"
	cfg := ParseConfig(strings.NewReader(input))
	if len(cfg.ExtraExtensions) != 0 {
		t.Error("comments-only input should produce empty config")
	}
}

func TestParseConfigExtensions(t *testing.T) {
	input := "[extensions]\n.vue = Web\nsvelte = Mobile\n"
	cfg := ParseConfig(strings.NewReader(input))
	if len(cfg.ExtraExtensions) != 2 {
		t.Fatalf("expected 2 extensions, got %d", len(cfg.ExtraExtensions))
	}
	if cfg.ExtraExtensions[0].Ext != ".vue" || cfg.ExtraExtensions[0].Category != extensions.Web {
		t.Errorf("first ext = %+v", cfg.ExtraExtensions[0])
	}
	if cfg.ExtraExtensions[1].Ext != ".svelte" {
		t.Errorf("expected .svelte, got %q", cfg.ExtraExtensions[1].Ext)
	}
}

func TestParseConfigEditorPriority(t *testing.T) {
	input := "[editor-priority]\ncom.microsoft.VSCode\ndev.zed.Zed\n"
	cfg := ParseConfig(strings.NewReader(input))
	if len(cfg.EditorPriority) != 2 {
		t.Fatalf("expected 2 priorities, got %d", len(cfg.EditorPriority))
	}
	if cfg.EditorPriority[0] != "com.microsoft.VSCode" {
		t.Errorf("first priority = %q", cfg.EditorPriority[0])
	}
}

func TestParseConfigEditorArgs(t *testing.T) {
	input := "[editor-args]\ncom.microsoft.VSCode = --new-window\n"
	cfg := ParseConfig(strings.NewReader(input))
	if cfg.EditorArgs["com.microsoft.VSCode"] != "--new-window" {
		t.Errorf("editor args = %v", cfg.EditorArgs)
	}
}

func TestParseConfigSettings(t *testing.T) {
	input := "[settings]\ntheme = light\n"
	cfg := ParseConfig(strings.NewReader(input))
	if cfg.Theme != "light" {
		t.Errorf("theme = %q, want light", cfg.Theme)
	}
}

func TestParseConfigMixedSections(t *testing.T) {
	input := `# Config file
[extensions]
.vue = Web

[editor-priority]
com.microsoft.VSCode

[settings]
theme = dark

[editor-args]
com.microsoft.VSCode = --new-window
`
	cfg := ParseConfig(strings.NewReader(input))
	if len(cfg.ExtraExtensions) != 1 {
		t.Errorf("expected 1 extension, got %d", len(cfg.ExtraExtensions))
	}
	if len(cfg.EditorPriority) != 1 {
		t.Errorf("expected 1 priority, got %d", len(cfg.EditorPriority))
	}
	if cfg.Theme != "dark" {
		t.Errorf("theme = %q", cfg.Theme)
	}
	if cfg.EditorArgs["com.microsoft.VSCode"] != "--new-window" {
		t.Errorf("editor args = %v", cfg.EditorArgs)
	}
}

func TestParseConfigMalformedLines(t *testing.T) {
	input := "[extensions]\nno-equals-sign\n.vue = Web\n"
	cfg := ParseConfig(strings.NewReader(input))
	if len(cfg.ExtraExtensions) != 1 {
		t.Errorf("expected 1 extension (malformed skipped), got %d", len(cfg.ExtraExtensions))
	}
}

func TestReadProfileFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.profile")
	os.WriteFile(path, []byte(`# comment
.json = com.app1
.md = com.app2

# another comment
.py = com.app3
`), 0600)

	changes := readProfileFile(path)
	if len(changes) != 3 {
		t.Fatalf("expected 3 changes, got %d", len(changes))
	}
	if changes[".json"] != "com.app1" {
		t.Errorf(".json = %q", changes[".json"])
	}
	if changes[".md"] != "com.app2" {
		t.Errorf(".md = %q", changes[".md"])
	}
	if changes[".py"] != "com.app3" {
		t.Errorf(".py = %q", changes[".py"])
	}
}

func TestReadProfileFileEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.profile")
	os.WriteFile(path, []byte(""), 0600)

	changes := readProfileFile(path)
	if len(changes) != 0 {
		t.Errorf("expected empty map, got %d entries", len(changes))
	}
}

func TestReadProfileFileCommentsOnly(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "comments.profile")
	os.WriteFile(path, []byte("# only comments\n# another comment\n"), 0600)

	changes := readProfileFile(path)
	if len(changes) != 0 {
		t.Errorf("expected empty map, got %d entries", len(changes))
	}
}

func TestReadProfileFileMalformed(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.profile")
	os.WriteFile(path, []byte("no equals\njson = app1\njson = app2\n"), 0600)

	changes := readProfileFile(path)
	if len(changes) != 1 {
		t.Errorf("expected 1 change (last wins), got %d", len(changes))
	}
}

func TestValidProfileNamePattern(t *testing.T) {
	valid := []string{"abc", "ABC", "123", "a-b", "a_b", "a-b_1"}
	invalid := []string{"invalid name", "invalid!", "invalid?", "invalid/"}

	for _, v := range valid {
		if !validProfileName.MatchString(v) {
			t.Errorf("expected %q to be valid", v)
		}
	}
	for _, v := range invalid {
		if validProfileName.MatchString(v) {
			t.Errorf("expected %q to be invalid", v)
		}
	}
}