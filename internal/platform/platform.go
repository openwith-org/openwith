// Package platform provides a cross-platform interface for managing
// default file associations. On macOS it uses LaunchServices plist
// manipulation; on Linux it uses xdg-mime.
package platform

// Backend defines the interface for platform-specific file association operations.
type Backend interface {
	// GetDefault returns the current default app name for an extension.
	GetDefault(ext string) string
	// GetDefaultBundleID returns the current default app identifier for an extension.
	GetDefaultBundleID(ext string) string
	// SetDefault sets the default app for an extension.
	SetDefault(ext string, appID string) error
	// GetDefaultBrowser returns the current default browser name.
	GetDefaultBrowser() string
	// SetDefaultBrowser sets the default browser.
	SetDefaultBrowser(appID string) error
	// DetectApps returns a list of installed application identifiers.
	DetectApp(appID string) bool
}
