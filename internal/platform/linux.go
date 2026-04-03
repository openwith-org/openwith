//go:build linux

package platform

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// LinuxBackend implements Backend using xdg-mime and xdg-settings.
type LinuxBackend struct {
	DryRun bool
}

func NewLinuxBackend(dryRun bool) *LinuxBackend {
	return &LinuxBackend{DryRun: dryRun}
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
	out, err := exec.Command("xdg-mime", "query", "default", mime).Output()
	if err != nil {
		return "unknown"
	}
	desktop := strings.TrimSpace(string(out))
	// Strip .desktop suffix for display
	return strings.TrimSuffix(desktop, ".desktop")
}

func (b *LinuxBackend) GetDefaultBundleID(ext string) string {
	mime := mimeTypeForExt(ext)
	out, err := exec.Command("xdg-mime", "query", "default", mime).Output()
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
	return exec.Command("xdg-mime", "default", appID, mime).Run()
}

func (b *LinuxBackend) GetDefaultBrowser() string {
	out, err := exec.Command("xdg-settings", "get", "default-web-browser").Output()
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
	return exec.Command("xdg-settings", "set", "default-web-browser", appID).Run()
}

func (b *LinuxBackend) DetectApp(appID string) bool {
	// Sanitize appID to prevent path traversal
	appID = filepath.Base(appID)

	// On Linux, check if the .desktop file exists
	out, err := exec.Command("which", strings.TrimSuffix(appID, ".desktop")).Output()
	if err == nil && len(strings.TrimSpace(string(out))) > 0 {
		return true
	}
	// Also check common desktop file locations
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
