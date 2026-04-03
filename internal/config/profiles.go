package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var validProfileName = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// Profile represents a saved set of file associations.
type Profile struct {
	Name    string
	Path    string
	Changes map[string]string // ext -> bundleID
}

func profilesDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home dir: %w", err)
	}
	return filepath.Join(home, ".config", "openwith", "profiles"), nil
}

// ListProfiles returns all saved profiles.
func ListProfiles() []Profile {
	dir, err := profilesDir()
	if err != nil {
		return nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var profiles []Profile
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".profile") {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".profile")
		p := Profile{
			Name: name,
			Path: filepath.Join(dir, e.Name()),
		}
		p.Changes = readProfileFile(p.Path)
		profiles = append(profiles, p)
	}
	return profiles
}

// SaveProfile saves a set of changes as a named profile.
func SaveProfile(name string, changes map[string]string) error {
	if !validProfileName.MatchString(name) {
		return fmt.Errorf("invalid profile name: %q (only alphanumeric, dash, underscore allowed)", name)
	}
	dir, err := profilesDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create profiles dir: %w", err)
	}

	path := filepath.Join(dir, name+".profile")

	exts := make([]string, 0, len(changes))
	for ext := range changes {
		exts = append(exts, ext)
	}
	sort.Strings(exts)

	var b strings.Builder
	b.WriteString("# openwith profile: " + name + "\n")
	b.WriteString("# Format: .ext = bundle.id\n\n")
	for _, ext := range exts {
		fmt.Fprintf(&b, "%s = %s\n", ext, changes[ext])
	}

	return os.WriteFile(path, []byte(b.String()), 0600)
}

// DeleteProfile removes a saved profile.
func DeleteProfile(name string) error {
	if !validProfileName.MatchString(name) {
		return fmt.Errorf("invalid profile name: %q", name)
	}
	dir, err := profilesDir()
	if err != nil {
		return err
	}
	path := filepath.Join(dir, name+".profile")
	return os.Remove(path)
}

// readProfileFile reads a profile file into a map.
func readProfileFile(path string) map[string]string {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	// Reject excessively large profile files
	if info, err := f.Stat(); err == nil && info.Size() > maxConfigSize {
		return nil
	}

	changes := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		ext := strings.TrimSpace(parts[0])
		bundleID := strings.TrimSpace(parts[1])
		changes[ext] = bundleID
	}
	return changes
}
