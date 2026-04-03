# openwith

A terminal UI for managing default file associations on macOS and Linux. Pick which editor opens `.py`, `.json`, `.tsx`, and 50+ other file types — in bulk or one at a time.

```
  openwith
  Manage default file associations on macOS

  ╭────────────────────────────────────────────╮
  │                                            │
  │  ▸ Bulk Mode                               │
  │      Set one editor for all file types     │
  │                                            │
  │    Per-Extension Mode                      │
  │      Choose a different editor per type    │
  │                                            │
  │    Default Browser                         │
  │      Choose which browser opens links      │
  │                                            │
  │    Load Profile                            │
  │      Apply a saved set of associations     │
  │                                            │
  │    Revert to Backup                        │
  │      Restore from a previous backup        │
  │                                            │
  ╰────────────────────────────────────────────╯

  ↑/↓ navigate • enter select • q quit
```

## Why

Every developer has opinions about which editor should open which file. macOS makes this tedious — right-click, "Open With", "Change All", repeat for every extension. Linux is similar — editing `mimeapps.list` or running `xdg-mime` for each type. `openwith` lets you set all your defaults in seconds from the terminal.

Nothing else occupies this niche. `duti` requires knowing bundle IDs. System Preferences handles one file at a time. `xdg-mime` requires knowing MIME types. This tool is terminal-native, developer-focused, and supports bulk changes.

## Install

### Homebrew

```bash
brew install openwith-org/tap/openwith
```

### From source

```bash
go install github.com/openwith-org/openwith@latest
```

### Build locally

```bash
git clone https://github.com/openwith-org/openwith.git
cd openwith
go build -o openwith .
```

## Usage

```bash
openwith                       # launch the interactive TUI
openwith --dry-run             # preview changes without writing anything
openwith --media               # include media file types (images, video, audio, PDFs)
openwith --theme light         # use the light color scheme
```

### Modes

**Bulk Mode** — pick one editor, apply it to all 54 file types at once.

**Per-Extension Mode** — navigate the extension list, press `Tab`/`Shift+Tab` to cycle through editors for each one. Press `/` to search and filter.

**Default Browser** — pick which browser handles `http://` and `https://` links. Detects Safari, Chrome, Firefox, Arc, Brave, Edge, and more.

**Load Profile** — apply a previously saved set of associations. Great for switching between "work" and "personal" configs.

**Revert to Backup** — every change creates a timestamped backup. Pick one to restore.

### Keyboard shortcuts

| Key | Where | Action |
|-----|-------|--------|
| `↑`/`↓` or `j`/`k` | Everywhere | Navigate |
| `Enter` | Menu, Bulk, Profiles | Select / confirm |
| `Tab` / `Shift+Tab` | Per-Extension | Cycle editor for current extension |
| `/` | Per-Extension | Search/filter extensions |
| `Esc` | Per-Extension | Clear filter or go back |
| `s` | Confirm screen | Save changes as a named profile |
| `r` | Apply screen | Retry failed extensions |
| `d` | Profiles screen | Delete selected profile |
| `q` | Menu | Quit |
| `Ctrl+C` | Everywhere | Quit |

### CLI flags

```
--dry-run              Preview changes without applying
--media                Include media file types (images, video, audio, PDFs)
--theme <dark|light>   Color scheme
--export <file>        Export current file associations to a file
--import <file>        Import and apply associations from a file
--completions <shell>  Generate shell completions (bash, zsh, fish)
--version              Print version and exit
```

### Import / Export

Share your file associations as a dotfile:

```bash
# Export current associations
openwith --export my-defaults.txt

# Import on another machine
openwith --import my-defaults.txt

# Print to stdout
openwith --export -
```

The export format is plain text:

```
.py = com.microsoft.VSCode  # VS Code
.go = dev.zed.Zed  # Zed
.rs = dev.zed.Zed  # Zed
```

### Shell completions

```bash
# Zsh (add to ~/.zshrc)
eval "$(openwith --completions zsh)"

# Bash (add to ~/.bashrc)
eval "$(openwith --completions bash)"

# Fish
openwith --completions fish | source
```

## Configuration

A config file is created automatically at `~/.config/openwith/config.txt` on first run.

```ini
# Color theme
[settings]
theme = dark

# Add custom file extensions
[extensions]
.custom = Other
.mdx = Web

# Preferred editor order (Tab cycles in this order)
[editor-priority]
com.microsoft.VSCode
dev.zed.Zed
com.sublimetext.4

# Per-editor launch arguments
[editor-args]
com.microsoft.VSCode = --new-window
com.sublimetext.4 = --new-window
```

### Profiles

Profiles are saved in `~/.config/openwith/profiles/`. You can create them from the confirm screen by pressing `s`, or load/delete them from the "Load Profile" menu option.

### Backups

Every time changes are applied, a timestamped backup of the LaunchServices plist is saved to `~/.config/openwith/backups/`. The 10 most recent backups are kept.

## Supported editors

The app auto-detects which of these are installed via Spotlight:

| Editor | Bundle ID |
|--------|-----------|
| Antigravity | `com.google.antigravity` |
| Windsurf | `com.exafunction.windsurf` |
| VS Code | `com.microsoft.VSCode` |
| Cursor | `com.todesktop.230313mzl4w4u92` |
| Zed | `dev.zed.Zed` |
| Sublime Text | `com.sublimetext.*` (auto-detected) |
| Nova | `com.panic.Nova` |
| BBEdit | `com.barebones.bbedit` |
| TextMate | `com.macromates.TextMate` |
| Xcode | `com.apple.dt.Xcode` |
| IntelliJ IDEA | `com.jetbrains.intellij` |
| GoLand | `com.jetbrains.goland` |
| PyCharm | `com.jetbrains.pycharm` |
| WebStorm | `com.jetbrains.WebStorm` |
| VimR | `com.qvacua.VimR` |
| Emacs | `org.gnu.Emacs` |
| Lapce | `dev.lapce.lapce` |
| Fleet | `com.jetbrains.fleet` |
| CotEditor | `com.coteditor.CotEditor` |

With `--media`, media apps are also detected: VLC, IINA, QuickTime, Preview, Pixelmator Pro, Affinity Photo, Adobe Photoshop, Skim, PDF Expert, Apple Music, Spotify.

## Supported file types

### Code (54 types, always included)

**Config:** `.json` `.jsonc` `.yaml` `.yml` `.toml` `.ini` `.cfg` `.env` `.nix`
**Docs:** `.md` `.txt` `.log`
**Scripts:** `.sh` `.bash` `.zsh` `.fish` `.py` `.rb` `.rs` `.go` `.lua` `.zig` `.nim`
**Web:** `.js` `.jsx` `.ts` `.tsx` `.html` `.css` `.scss` `.vue` `.svelte` `.astro`
**Mobile:** `.dart` `.swift` `.kt` `.java`
**Systems:** `.c` `.cpp` `.h` `.hpp` `.proto`
**DB:** `.sql` `.graphql` `.prisma`
**Data:** `.csv` `.xml` `.svg`
**DevOps:** `.dockerfile` `.tf` `.hcl`
**Other:** `.lock` `.gitignore` `.editorconfig`

### Media (25 types, with `--media`)

**Images:** `.png` `.jpg` `.jpeg` `.gif` `.webp` `.tiff` `.bmp` `.ico` `.heic` `.psd`
**Video:** `.mp4` `.mkv` `.avi` `.mov` `.wmv` `.webm` `.flv`
**Audio:** `.mp3` `.flac` `.wav` `.aac` `.ogg` `.m4a` `.wma`
**PDFs:** `.pdf`

Add more via the config file's `[extensions]` section.

## How it works

The app uses a platform-specific backend selected at build time via build tags.

### macOS

Directly reads and writes the LaunchServices plist at:

```
~/Library/Preferences/com.apple.LaunchServices/com.apple.launchservices.secure.plist
```

It maps file extensions to UTIs (Uniform Type Identifiers), updates the handler entries in the plist, writes it back atomically (temp file + rename), and restarts `lsd` to reload the changes. Editors are detected via Spotlight (`mdfind`).

### Linux

Uses `xdg-mime` to query and set default applications per MIME type, and `xdg-settings` for the default browser. Apps are detected by checking for `.desktop` files in standard locations (`/usr/share/applications`, `~/.local/share/applications`).

### Safety

- A timestamped backup is created before every write (macOS)
- Writes are atomic (temp file + `os.Rename`) to prevent corruption (macOS)
- App identifiers are validated before writing (via `mdfind` on macOS, `.desktop` file lookup on Linux)
- If `lsd` fails to restart, a warning is shown (changes still saved, active after log out/in) (macOS)
- Dry-run mode lets you preview everything without applying changes

## Requirements

- **macOS**: Spotlight must be enabled for editor detection
- **Linux**: `xdg-mime` and `xdg-settings` must be available (typically pre-installed on most desktop distros)
- Go 1.21+ (to build from source)

## License

MIT
