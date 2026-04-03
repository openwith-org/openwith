package config

import (
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
	// svelte without dot should get dot prepended
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
