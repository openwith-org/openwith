//go:build linux

package platform

import (
	"os"
	"path/filepath"
	"strings"
)

// LinuxBackend implements Backend using xdg-mime and xdg-settings.
type LinuxBackend struct {
	DryRun bool
	Runner CommandRunner
}

func NewLinuxBackend(dryRun bool) *LinuxBackend {
	return &LinuxBackend{DryRun: dryRun, Runner: RealCommandRunner{}}
}

func (b *LinuxBackend) run() CommandRunner {
	if b.Runner != nil {
		return b.Runner
	}
	return RealCommandRunner{}
}

// mimeTypeForExt returns the MIME type for a file extension using xdg-mime.
func mimeTypeForExt(ext string) string {
	// Common mappings — xdg-mime query filetype needs an actual file,
	// so we use a static map for known types.
	mimeTypes := map[string]string{
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
	if mt, ok := mimeTypes[ext]; ok {
		return mt
	}
	// Fallback: text/plain for unknown extensions
	return "text/plain"
}

func (b *LinuxBackend) GetDefault(ext string) string {
	mime := mimeTypeForExt(ext)
	runner := b.run()
	out, err := runner.Output("xdg-mime", "query", "default", mime)
	if err != nil {
		return "unknown"
	}
	desktop := strings.TrimSpace(string(out))
	return strings.TrimSuffix(desktop, ".desktop")
}

func (b *LinuxBackend) GetDefaultBundleID(ext string) string {
	mime := mimeTypeForExt(ext)
	runner := b.run()
	out, err := runner.Output("xdg-mime", "query", "default", mime)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func (b *LinuxBackend) SetDefault(ext string, appID string) error {
	if b.DryRun {
		return nil
	}
	mime := mimeTypeForExt(ext)
	runner := b.run()
	return runner.Run("xdg-mime", "default", appID, mime)
}

func (b *LinuxBackend) GetDefaultBrowser() string {
	runner := b.run()
	out, err := runner.Output("xdg-settings", "get", "default-web-browser")
	if err != nil {
		return "unknown"
	}
	desktop := strings.TrimSpace(string(out))
	return strings.TrimSuffix(desktop, ".desktop")
}

func (b *LinuxBackend) SetDefaultBrowser(appID string) error {
	if b.DryRun {
		return nil
	}
	runner := b.run()
	return runner.Run("xdg-settings", "set", "default-web-browser", appID)
}

func (b *LinuxBackend) DetectApp(appID string) bool {
	appID = filepath.Base(appID)

	runner := b.run()
	out, err := runner.Output("which", strings.TrimSuffix(appID, ".desktop"))
	if err == nil && len(strings.TrimSpace(string(out))) > 0 {
		return true
	}

	paths := []string{
		filepath.Join("/usr/share/applications", appID),
		filepath.Join("/usr/local/share/applications", appID),
	}
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, ".local/share/applications", appID))
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return true
		}
	}
	return false
}
