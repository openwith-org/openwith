package defaults

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"howett.net/plist"
)

// ValidBundleIDPattern matches valid reverse-DNS bundle identifiers.
var ValidBundleIDPattern = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

type Result struct {
	Extension string
	Success   bool
	Error     error
}

type CurrentDefault struct {
	Extension string
	AppName   string
}

type Client struct {
	DryRun bool
}

func New(dryRun bool) *Client {
	return &Client{DryRun: dryRun}
}

// ValidateBundleID checks that a bundle ID corresponds to an installed app.
func ValidateBundleID(bundleID string) bool {
	if !ValidBundleIDPattern.MatchString(bundleID) {
		return false
	}
	out, err := exec.Command("mdfind",
		"kMDItemCFBundleIdentifier == '"+bundleID+"'").Output()
	return err == nil && len(strings.TrimSpace(string(out))) > 0
}

// backupPlist creates a timestamped backup of the plist file.
func backupPlist(path string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get home dir: %w", err)
	}
	backupDir := filepath.Join(home, ".config", "openwith", "backups")
	if err := os.MkdirAll(backupDir, 0700); err != nil {
		return fmt.Errorf("create backup dir: %w", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read plist for backup: %w", err)
	}

	timestamp := time.Now().Format("20060102-150405")
	backupPath := filepath.Join(backupDir, fmt.Sprintf("launchservices-%s.plist", timestamp))

	// Reject symlinks to prevent symlink attacks
	if info, err := os.Lstat(backupPath); err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("backup path is a symlink: %s", backupPath)
		}
	}

	if err := os.WriteFile(backupPath, data, 0600); err != nil {
		return fmt.Errorf("write backup: %w", err)
	}

	// Keep only the 10 most recent backups
	pruneBackups(backupDir, 10)
	return nil
}

// pruneBackups keeps only the N most recent backup files.
func pruneBackups(dir string, keep int) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	var backups []os.DirEntry
	for _, e := range entries {
		if !e.IsDir() && strings.HasPrefix(e.Name(), "launchservices-") {
			backups = append(backups, e)
		}
	}
	if len(backups) <= keep {
		return
	}
	// DirEntry names are sorted lexically; timestamp format ensures chronological order
	for _, e := range backups[:len(backups)-keep] {
		os.Remove(filepath.Join(dir, e.Name()))
	}
}

// atomicWriteFile writes data to a temp file then renames it into place.
func atomicWriteFile(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".openwith-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("close temp file: %w", err)
	}
	if err := os.Chmod(tmpPath, perm); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("chmod temp file: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("rename temp to target: %w", err)
	}
	return nil
}

// Backup represents a saved plist backup.
type Backup struct {
	Name      string // filename
	Path      string // full path
	Timestamp string // human-readable timestamp
}

// ListBackups returns available backups, newest first.
func ListBackups() []Backup {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	backupDir := filepath.Join(home, ".config", "openwith", "backups")
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return nil
	}
	var backups []Backup
	for _, e := range entries {
		if e.IsDir() || !strings.HasPrefix(e.Name(), "launchservices-") {
			continue
		}
		// Parse timestamp from filename: launchservices-20060102-150405.plist
		name := strings.TrimSuffix(strings.TrimPrefix(e.Name(), "launchservices-"), ".plist")
		ts := name
		if t, err := time.Parse("20060102-150405", name); err == nil {
			ts = t.Format("2006-01-02 15:04:05")
		}
		backups = append(backups, Backup{
			Name:      e.Name(),
			Path:      filepath.Join(backupDir, e.Name()),
			Timestamp: ts,
		})
	}
	// Reverse to get newest first
	for i, j := 0, len(backups)-1; i < j; i, j = i+1, j-1 {
		backups[i], backups[j] = backups[j], backups[i]
	}
	return backups
}

// RestoreBackup restores a backup plist file, replacing the current one.
func (c *Client) RestoreBackup(backup Backup) error {
	if c.DryRun {
		return nil
	}

	// Reject symlinks to prevent reading attacker-controlled files
	if info, err := os.Lstat(backup.Path); err != nil {
		return fmt.Errorf("stat backup: %w", err)
	} else if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("backup file is a symlink: %s", backup.Path)
	}

	data, err := os.ReadFile(backup.Path)
	if err != nil {
		return fmt.Errorf("read backup: %w", err)
	}

	path, err := lsPlistPath()
	if err != nil {
		return err
	}
	if err := atomicWriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("restore backup: %w", err)
	}

	// Restart lsd to pick up restored settings
	exec.Command("killall", "lsd").Run()
	return nil
}

// GetDefaultBrowser reads the current default browser (http handler) from the plist.
func (c *Client) GetDefaultBrowser() string {
	db, err := readDB()
	if err != nil {
		return "unknown"
	}
	for _, h := range db.Handlers {
		if h.URLScheme == "https" && h.RoleAll != "" {
			return resolveAppName(h.RoleAll)
		}
	}
	for _, h := range db.Handlers {
		if h.URLScheme == "http" && h.RoleAll != "" {
			return resolveAppName(h.RoleAll)
		}
	}
	return "unknown"
}

// GetDefaultBrowserBundleID returns the bundle ID of the current default browser.
func (c *Client) GetDefaultBrowserBundleID() string {
	db, err := readDB()
	if err != nil {
		return ""
	}
	for _, h := range db.Handlers {
		if h.URLScheme == "https" && h.RoleAll != "" {
			return h.RoleAll
		}
	}
	for _, h := range db.Handlers {
		if h.URLScheme == "http" && h.RoleAll != "" {
			return h.RoleAll
		}
	}
	return ""
}

// SetDefaultBrowser sets the default browser for http and https URL schemes.
func (c *Client) SetDefaultBrowser(bundleID string) error {
	if c.DryRun {
		return nil
	}

	if !ValidateBundleID(bundleID) {
		return fmt.Errorf("app not found for bundle ID %s", bundleID)
	}

	path, err := lsPlistPath()
	if err != nil {
		return err
	}
	if err := backupPlist(path); err != nil {
		return fmt.Errorf("backup failed: %w", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read plist: %w", err)
	}

	var db lsDatabase
	if _, err := plist.Unmarshal(data, &db); err != nil {
		return fmt.Errorf("parse plist: %w", err)
	}

	now := time.Now().Unix()
	for _, scheme := range []string{"http", "https"} {
		upsertURLScheme(&db, scheme, bundleID, now)
	}

	out, err := plist.Marshal(&db, plist.BinaryFormat)
	if err != nil {
		return fmt.Errorf("marshal plist: %w", err)
	}

	if err := atomicWriteFile(path, out, 0600); err != nil {
		return fmt.Errorf("write plist: %w", err)
	}

	exec.Command("killall", "lsd").Run()
	return nil
}

func upsertURLScheme(db *lsDatabase, scheme, bundleID string, now int64) {
	for i := range db.Handlers {
		if db.Handlers[i].URLScheme == scheme {
			db.Handlers[i].RoleAll = bundleID
			db.Handlers[i].ModificationDate = now
			db.Handlers[i].PreferredVersions = map[string]string{"LSHandlerRoleAll": "-"}
			return
		}
	}
	db.Handlers = append(db.Handlers, lsHandler{
		URLScheme:         scheme,
		RoleAll:           bundleID,
		ModificationDate:  now,
		PreferredVersions: map[string]string{"LSHandlerRoleAll": "-"},
	})
}

// Static UTI mappings for extensions that have well-known UTIs.
// Extensions NOT in this map use the tag-based approach
// (LSHandlerContentTag + public.filename-extension).
var staticUTIs = map[string]string{
	".json":  "public.json",
	".yaml":  "public.yaml",
	".yml":   "public.yaml",
	".md":    "net.daringfireball.markdown",
	".txt":   "public.plain-text",
	".log":   "com.apple.log",
	".sh":    "public.shell-script",
	".bash":  "public.bash-script",
	".zsh":   "public.zsh-script",
	".py":    "public.python-script",
	".rb":    "public.ruby-script",
	".js":    "com.netscape.javascript-source",
	".tsx":   "com.microsoft.typescript",
	".css":   "public.css",
	".xml":   "public.xml",
	".svg":   "public.svg-image",
	".toml":  "public.toml",
	".ini":   "com.microsoft.ini",
	".swift": "public.swift-source",
	".java":  "com.sun.java-source",
	".c":     "public.c-source",
	".cpp":   "public.c-plus-plus-source",
	".h":     "public.c-header",
	".hpp":   "public.c-plus-plus-header",
	".sql":   "org.iso.sql",
	".html":  "public.html",
	".csv":   "public.comma-separated-values-text",
	// Media types
	".png":  "public.png",
	".jpg":  "public.jpeg",
	".jpeg": "public.jpeg",
	".gif":  "com.compuserve.gif",
	".tiff": "public.tiff",
	".bmp":  "com.microsoft.bmp",
	".ico":  "com.microsoft.ico",
	".heic": "public.heic",
	".psd":  "com.adobe.photoshop-image",
	".mp4":  "public.mpeg-4",
	".mov":  "com.apple.quicktime-movie",
	".avi":  "public.avi",
	".mp3":  "public.mp3",
	".flac": "org.xiph.flac",
	".wav":  "com.microsoft.waveform-audio",
	".aac":  "public.aac-audio",
	".m4a":  "com.apple.m4a-audio",
	".pdf":  "com.adobe.pdf",
	// .ts → public.mpeg-2-transport-stream (video), use tag-based
	// .cfg → public.toml (conflicts with .toml), use tag-based
}

type lsHandler struct {
	ContentType       string            `plist:"LSHandlerContentType,omitempty"`
	ContentTag        string            `plist:"LSHandlerContentTag,omitempty"`
	ContentTagClass   string            `plist:"LSHandlerContentTagClass,omitempty"`
	RoleAll           string            `plist:"LSHandlerRoleAll,omitempty"`
	ModificationDate  int64             `plist:"LSHandlerModificationDate"`
	PreferredVersions map[string]string `plist:"LSHandlerPreferredVersions,omitempty"`
	URLScheme         string            `plist:"LSHandlerURLScheme,omitempty"`
}

type lsDatabase struct {
	Handlers []lsHandler `plist:"LSHandlers"`
}

func lsPlistPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home dir: %w", err)
	}
	return filepath.Join(home, "Library", "Preferences",
		"com.apple.LaunchServices", "com.apple.launchservices.secure.plist"), nil
}

func readDB() (lsDatabase, error) {
	var db lsDatabase
	path, err := lsPlistPath()
	if err != nil {
		return db, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return db, err
	}
	_, err = plist.Unmarshal(data, &db)
	return db, err
}

// resolveAppName converts a bundle ID to an app name using mdfind + Spotlight.
func resolveAppName(bundleID string) string {
	if !ValidBundleIDPattern.MatchString(bundleID) {
		return bundleID
	}
	out, err := exec.Command("mdfind",
		"kMDItemCFBundleIdentifier == '"+bundleID+"'").Output()
	if err != nil || len(strings.TrimSpace(string(out))) == 0 {
		return bundleID
	}
	// First line is the app path, extract the .app name
	appPath := strings.SplitN(strings.TrimSpace(string(out)), "\n", 2)[0]
	base := filepath.Base(appPath)
	return strings.TrimSuffix(base, ".app")
}

// GetAllDefaults reads current file associations from the LS plist.
func (c *Client) GetAllDefaults(exts []string) []CurrentDefault {
	db, err := readDB()
	if err != nil {
		results := make([]CurrentDefault, len(exts))
		for i, ext := range exts {
			results[i] = CurrentDefault{Extension: ext, AppName: "unknown"}
		}
		return results
	}

	// Build lookup maps from the plist handlers
	utiToBundle := make(map[string]string)
	tagToBundle := make(map[string]string)
	for _, h := range db.Handlers {
		if h.ContentType != "" && h.RoleAll != "" {
			utiToBundle[h.ContentType] = h.RoleAll
		}
		if h.ContentTag != "" && h.ContentTagClass == "public.filename-extension" && h.RoleAll != "" {
			tagToBundle[h.ContentTag] = h.RoleAll
		}
	}

	// Resolve each extension to its bundle ID
	bundleByExt := make(map[string]string)
	uniqueBundles := make(map[string]bool)
	for _, ext := range exts {
		bare := strings.TrimPrefix(ext, ".")
		var bundleID string
		if uti, ok := staticUTIs[ext]; ok {
			bundleID = utiToBundle[uti]
		}
		if bundleID == "" {
			bundleID = tagToBundle[bare]
		}
		if bundleID != "" {
			bundleByExt[ext] = bundleID
			uniqueBundles[bundleID] = true
		}
	}

	// Resolve bundle IDs to app names concurrently
	nameByBundle := make(map[string]string)
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, 8)

	for bid := range uniqueBundles {
		wg.Add(1)
		go func(b string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			name := resolveAppName(b)
			mu.Lock()
			nameByBundle[b] = name
			mu.Unlock()
		}(bid)
	}
	wg.Wait()

	results := make([]CurrentDefault, len(exts))
	for i, ext := range exts {
		name := "unknown"
		if bid, ok := bundleByExt[ext]; ok {
			if n, ok := nameByBundle[bid]; ok {
				name = n
			}
		}
		results[i] = CurrentDefault{Extension: ext, AppName: name}
	}
	return results
}

// GetAllDefaultsRaw returns ext -> bundleID mapping (no app name resolution).
func (c *Client) GetAllDefaultsRaw(exts []string) map[string]string {
	db, err := readDB()
	if err != nil {
		return nil
	}

	utiToBundle := make(map[string]string)
	tagToBundle := make(map[string]string)
	for _, h := range db.Handlers {
		if h.ContentType != "" && h.RoleAll != "" {
			utiToBundle[h.ContentType] = h.RoleAll
		}
		if h.ContentTag != "" && h.ContentTagClass == "public.filename-extension" && h.RoleAll != "" {
			tagToBundle[h.ContentTag] = h.RoleAll
		}
	}

	result := make(map[string]string)
	for _, ext := range exts {
		bare := strings.TrimPrefix(ext, ".")
		var bundleID string
		if uti, ok := staticUTIs[ext]; ok {
			bundleID = utiToBundle[uti]
		}
		if bundleID == "" {
			bundleID = tagToBundle[bare]
		}
		if bundleID != "" {
			result[ext] = bundleID
		}
	}
	return result
}

func (c *Client) ApplyChanges(changes map[string]string, results chan<- Result) {
	defer close(results)

	if c.DryRun {
		for ext := range changes {
			results <- Result{Extension: ext, Success: true}
		}
		return
	}

	// Validate bundle IDs before making any changes
	// Only need to check unique bundle IDs once
	validated := make(map[string]bool)
	for ext, bundleID := range changes {
		if checked, ok := validated[bundleID]; ok {
			if !checked {
				results <- Result{Extension: ext, Success: false,
					Error: fmt.Errorf("app not found for bundle ID %s", bundleID)}
				return
			}
			continue
		}
		valid := ValidateBundleID(bundleID)
		validated[bundleID] = valid
		if !valid {
			results <- Result{Extension: ext, Success: false,
				Error: fmt.Errorf("app not found for bundle ID %s", bundleID)}
			return
		}
	}

	path, err := lsPlistPath()
	if err != nil {
		for ext := range changes {
			results <- Result{Extension: ext, Success: false, Error: err}
		}
		return
	}

	// Create backup before modifying
	if err := backupPlist(path); err != nil {
		for ext := range changes {
			results <- Result{Extension: ext, Success: false,
				Error: fmt.Errorf("backup failed: %w", err)}
		}
		return
	}

	data, err := os.ReadFile(path)
	if err != nil {
		for ext := range changes {
			results <- Result{Extension: ext, Success: false,
				Error: fmt.Errorf("read plist: %w", err)}
		}
		return
	}

	var db lsDatabase
	if _, err := plist.Unmarshal(data, &db); err != nil {
		for ext := range changes {
			results <- Result{Extension: ext, Success: false,
				Error: fmt.Errorf("parse plist: %w", err)}
		}
		return
	}

	now := time.Now().Unix()

	for ext, bundleID := range changes {
		bare := strings.TrimPrefix(ext, ".")
		if uti, ok := staticUTIs[ext]; ok {
			upsertByUTI(&db, uti, bundleID, now)
		} else {
			upsertByTag(&db, bare, bundleID, now)
		}
	}

	out, err := plist.Marshal(&db, plist.BinaryFormat)
	if err != nil {
		for ext := range changes {
			results <- Result{Extension: ext, Success: false,
				Error: fmt.Errorf("marshal plist: %w", err)}
		}
		return
	}

	// Atomic write: temp file + rename
	if err := atomicWriteFile(path, out, 0600); err != nil {
		for ext := range changes {
			results <- Result{Extension: ext, Success: false,
				Error: fmt.Errorf("write plist: %w", err)}
		}
		return
	}

	// Restart lsd to pick up the changes
	if err := exec.Command("killall", "lsd").Run(); err != nil {
		// Changes were written successfully but lsd didn't restart
		for ext := range changes {
			results <- Result{Extension: ext, Success: true,
				Error: fmt.Errorf("changes saved but lsd restart failed (log out/in to activate)")}
		}
		return
	}

	for ext := range changes {
		results <- Result{Extension: ext, Success: true}
	}
}

func upsertByUTI(db *lsDatabase, uti, bundleID string, now int64) {
	for i := range db.Handlers {
		if db.Handlers[i].ContentType == uti {
			db.Handlers[i].RoleAll = bundleID
			db.Handlers[i].ModificationDate = now
			db.Handlers[i].PreferredVersions = map[string]string{"LSHandlerRoleAll": "-"}
			return
		}
	}
	db.Handlers = append(db.Handlers, lsHandler{
		ContentType:       uti,
		RoleAll:           bundleID,
		ModificationDate:  now,
		PreferredVersions: map[string]string{"LSHandlerRoleAll": "-"},
	})
}

func upsertByTag(db *lsDatabase, tag, bundleID string, now int64) {
	for i := range db.Handlers {
		if db.Handlers[i].ContentTag == tag &&
			db.Handlers[i].ContentTagClass == "public.filename-extension" {
			db.Handlers[i].RoleAll = bundleID
			db.Handlers[i].ModificationDate = now
			db.Handlers[i].PreferredVersions = map[string]string{"LSHandlerRoleAll": "-"}
			return
		}
	}
	db.Handlers = append(db.Handlers, lsHandler{
		ContentTag:        tag,
		ContentTagClass:   "public.filename-extension",
		RoleAll:           bundleID,
		ModificationDate:  now,
		PreferredVersions: map[string]string{"LSHandlerRoleAll": "-"},
	})
}
