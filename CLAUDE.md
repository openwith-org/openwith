# openwith

macOS and Linux TUI for managing default file associations — pick which editor opens `.py`, `.json`, `.tsx`, etc.

## Build & Run

```bash
go build -o openwith .
./openwith                     # interactive mode
./openwith --dry-run           # preview changes without applying
./openwith --media             # include media file types
./openwith --theme light       # use light color scheme
./openwith --export config.txt # export current associations
./openwith --import config.txt # import associations from file
./openwith --version           # print version
./openwith --completions zsh   # generate shell completions
```

## Architecture

**Cross-platform (macOS + Linux)** via build-tag-gated backends in `internal/platform/`.

- **macOS**: Directly manipulates the LaunchServices plist at `~/Library/Preferences/com.apple.LaunchServices/com.apple.launchservices.secure.plist`. Requires Spotlight (`mdfind`) for editor detection. Does NOT use `duti` at runtime.
- **Linux**: Uses `xdg-mime` for MIME type associations and `xdg-settings` for the default browser. Detects apps via `.desktop` files in standard locations.

### Screens (Bubble Tea v2)

`internal/tui/app.go` — main app model with 9 screens:
1. **Menu** → choose mode
2. **Bulk** → one editor for all file types
3. **Per-Extension** → cycle editors per extension with Tab/Shift+Tab, `/` to search
4. **Confirm** → review changes with color-coded diff before applying (s to save profile)
5. **Apply** → progress and results with error messages
6. **Default Browser** → pick which browser opens links
7. **Load Profile** → load/delete saved profiles
8. **Save Profile** → save current changes as a named profile
9. **Revert** → restore from a previous backup

### Internal Packages

- `internal/tui/` — all TUI screens and styles (Bubble Tea + Lipgloss v2)
- `internal/defaults/` — macOS plist read/write, UTI mappings, change application logic, backup/restore, browser defaults
- `internal/editors/` — editor/browser/media app registries, detection, priority sorting
- `internal/extensions/` — file extension list organized by category, merge support, media extensions
- `internal/config/` — user config file loading, profiles (`~/.config/openwith/`)
- `internal/platform/` — cross-platform backend (macOS plist, Linux xdg-mime)

### Key Files

- `internal/editors/editors.go` — editor registry (add new editors here)
- `internal/extensions/extensions.go` — extension list and categories (add new file types here)
- `internal/defaults/defaults.go` — plist manipulation, UTI mappings, `ApplyChanges()`

## Dependencies

- `charm.land/bubbletea/v2` — terminal UI framework
- `charm.land/lipgloss/v2` — terminal styling
- `howett.net/plist` — macOS plist parsing

## Safety Features

- Plist backup before every write on macOS (timestamped, stored in `~/.config/openwith/backups/`, auto-prunes to 10)
- Atomic file writes (temp file + rename) to prevent corruption (macOS)
- App ID validation before writing (via `mdfind` on macOS, `.desktop` file lookup on Linux)
- `killall lsd` errors surfaced as warnings in UI (macOS)

## Known Limitations

- 19 built-in editors + 11 media apps + 10 browsers — expandable via config
- 54 code extensions + 25 media extensions — expandable via config file
- macOS: Spotlight must be enabled for editor detection
- macOS: No file locking on plist writes (TOCTOU still possible)
- Linux: Requires `xdg-mime` and `xdg-settings`
