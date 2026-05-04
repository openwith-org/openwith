// Package platform provides a cross-platform interface for managing
// default file associations. On macOS it uses LaunchServices plist
// manipulation; on Linux it uses xdg-mime.
package platform

import (
	"os/exec"
)

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

// CommandRunner defines an interface for executing external commands.
// This enables mocking in tests.
type CommandRunner interface {
	Run(cmd string, args ...string) error
	Output(cmd string, args ...string) ([]byte, error)
}

// RealCommandRunner executes commands using os/exec.
type RealCommandRunner struct{}

func (RealCommandRunner) Run(cmd string, args ...string) error {
	return exec.Command(cmd, args...).Run()
}

func (RealCommandRunner) Output(cmd string, args ...string) ([]byte, error) {
	return exec.Command(cmd, args...).Output()
}
