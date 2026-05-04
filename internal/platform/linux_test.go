//go:build linux

package platform

import (
	"path/filepath"
	"testing"
)

type mockRunner struct {
	outputs map[string][]byte
	errs   map[string]error
	runs   map[string]error
}

func (m *mockRunner) Run(cmd string, args ...string) error {
	key := cmd + " " + joinArgs(args)
	if m.runs != nil {
		return m.runs[key]
	}
	return nil
}

func (m *mockRunner) Output(cmd string, args ...string) ([]byte, error) {
	key := cmd + " " + joinArgs(args)
	if m.outputs != nil {
		return m.outputs[key], m.errs[key]
	}
	return nil, m.errs[key]
}

func joinArgs(args []string) string {
	result := ""
	for _, a := range args {
		if result != "" {
			result += " "
		}
		result += a
	}
	return result
}

func TestMimeTypeForExtKnown(t *testing.T) {
	cases := map[string]string{
		".json": "application/json",
		".py":   "text/x-python",
		".html": "text/html",
		".pdf":  "application/pdf",
		".png":  "image/png",
	}
	for ext, want := range cases {
		got := mimeTypeForExt(ext)
		if got != want {
			t.Errorf("mimeTypeForExt(%q) = %q, want %q", ext, got, want)
		}
	}
}

func TestMimeTypeForExtFallback(t *testing.T) {
	got := mimeTypeForExt(".unknownext")
	if got != "text/plain" {
		t.Errorf("mimeTypeForExt(.unknownext) = %q, want text/plain", got)
	}
}

func TestMimeTypeForExtAll(t *testing.T) {
	knownTypes := map[string]string{
		".json":  "application/json",
		".yaml":  "application/x-yaml",
		".yml":   "application/x-yaml",
		".xml":   "application/xml",
		".html":  "text/html",
		".css":   "text/css",
		".js":    "application/javascript",
		".ts":    "application/typescript",
		".tsx":   "application/typescript",
		".py":    "text/x-python",
		".rb":    "application/x-ruby",
		".go":    "text/x-go",
		".rs":    "text/x-rust",
		".c":     "text/x-csrc",
		".cpp":   "text/x-c++src",
		".h":     "text/x-chdr",
		".java":  "text/x-java",
		".swift": "text/x-swift",
		".sh":    "application/x-shellscript",
		".md":    "text/markdown",
		".txt":   "text/plain",
		".sql":   "application/sql",
		".csv":   "text/csv",
		".svg":   "image/svg+xml",
		".pdf":   "application/pdf",
		".png":   "image/png",
		".jpg":   "image/jpeg",
		".jpeg":  "image/jpeg",
		".gif":   "image/gif",
		".mp4":   "video/mp4",
		".mp3":   "audio/mpeg",
		".flac":  "audio/flac",
		".wav":   "audio/wav",
	}
	for ext, want := range knownTypes {
		got := mimeTypeForExt(ext)
		if got != want {
			t.Errorf("mimeTypeForExt(%q) = %q, want %q", ext, got, want)
		}
	}
}

func TestDetectAppSanitizesPath(t *testing.T) {
	backend := &LinuxBackend{
		DryRun: true,
		Runner: &mockRunner{
			outputs: map[string][]byte{
				"which code": []byte("/usr/bin/code"),
			},
			errs: map[string]error{
				"which code": nil,
			},
		},
	}

	result := backend.DetectApp("../etc/passwd")
	if result {
		t.Error("DetectApp should reject path traversal")
	}

	result = backend.DetectApp("code.desktop")
	if !result {
		t.Error("DetectApp should find code.desktop")
	}
}

func TestDetectAppNormalizesDotDesktop(t *testing.T) {
	backend := &LinuxBackend{
		DryRun: true,
		Runner: &mockRunner{
			outputs: map[string][]byte{
				"which firefox": []byte("/usr/bin/firefox"),
			},
			errs: map[string]error{
				"which firefox": nil,
			},
		},
	}

	result := backend.DetectApp("firefox.desktop")
	if !result {
		t.Error("DetectApp should find firefox when given firefox.desktop")
	}
}

func TestBackendWithMockRunner(t *testing.T) {
	backend := &LinuxBackend{
		DryRun: false,
		Runner: &mockRunner{
			outputs: map[string][]byte{
				"xdg-mime query default application/json": []byte("code.desktop"),
			},
			errs: map[string]error{
				"xdg-mime query default application/json": nil,
			},
		},
	}

	got := backend.GetDefault(".json")
	if got != "code" {
		t.Errorf("GetDefault(.json) = %q, want code", got)
	}
}

func TestBackendStripDesktopSuffix(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"code.desktop", "code"},
		{"firefox.desktop", "firefox"},
		{"org.example.App.desktop", "org.example.App"},
		{"no-suffix", "no-suffix"},
	}

	for _, tt := range tests {
		backend := &LinuxBackend{
			DryRun: true,
			Runner: &mockRunner{
				outputs: map[string][]byte{
					"xdg-mime query default text/plain": []byte(tt.input),
				},
				errs: map[string]error{
					"xdg-mime query default text/plain": nil,
				},
			},
		}
		got := backend.GetDefault(".txt")
		if got != tt.expected {
			t.Errorf("GetDefault(.txt) returned %q, want %q", got, tt.expected)
		}
	}
}

func TestBackendDryRunNoOp(t *testing.T) {
	backend := &LinuxBackend{
		DryRun: true,
		Runner: &mockRunner{
			runs: map[string]error{},
		},
	}

	err := backend.SetDefault(".json", "code.desktop")
	if err != nil {
		t.Errorf("SetDefault in dry-run mode returned error: %v", err)
	}

	err = backend.SetDefaultBrowser("firefox.desktop")
	if err != nil {
		t.Errorf("SetDefaultBrowser in dry-run mode returned error: %v", err)
	}
}

func BenchmarkMimeTypeForExt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = mimeTypeForExt(".json")
		_ = mimeTypeForExt(".py")
		_ = mimeTypeForExt(".html")
	}
}

var _ = filepath.ToSlash
