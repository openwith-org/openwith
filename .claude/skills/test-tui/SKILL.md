---
name: test-tui
description: Thoroughly test the openwith TUI app using tmux to drive all screens, navigation, and user flows
disable-model-invocation: true
allowed-tools: Bash Read Grep
---

# TUI Integration Test via tmux

Build and interactively test the openwith TUI using tmux to send keystrokes and capture rendered output. This tests the real app in a real terminal — not unit tests.

## Setup

1. Build the app: `go build -o openwith .`
2. Kill any leftover tmux session: `tmux kill-session -t openwith-test 2>/dev/null`
3. Start a new tmux session running the app in dry-run mode:
   ```
   tmux new-session -d -s openwith-test -x 120 -y 40 './openwith --dry-run'
   ```
4. Wait 2 seconds for the app to load editors and render.

## Helper Pattern

For every test step, use this pattern:
```bash
# Send keys
tmux send-keys -t openwith-test <keys>
# Wait for render
sleep 0.5
# Capture output
tmux capture-pane -t openwith-test -p
```

Use `sleep 2` after actions that trigger async work (loading, applying changes).

## Test Plan

Run ALL of the following tests. After each capture, verify the expected content is present. Report PASS or FAIL for each test with the captured output.

### 1. Menu Screen
- **1a.** Capture initial screen. Verify: title "openwith", all 5 menu items visible (Bulk Mode, Per-Extension Mode, Default Browser, Load Profile, Revert to Backup), help text at bottom.
- **1b.** Press `j` (down). Capture. Verify: cursor moved to "Per-Extension Mode".
- **1c.** Press `k` (up). Capture. Verify: cursor back on "Bulk Mode".
- **1d.** Press `Down Down Down Down`. Capture. Verify: cursor on "Revert to Backup" (last item).
- **1e.** Press `Down`. Capture. Verify: cursor stays on last item (doesn't wrap or crash).

### 2. Bulk Mode
- **2a.** Navigate cursor to "Bulk Mode" (go back up), press `Enter`. Capture. Verify: title "Bulk Mode", list of editors with bundle IDs.
- **2b.** Press `j` then `k`. Capture. Verify: cursor moves and returns.
- **2c.** Press `Escape`. Capture. Verify: back at Menu screen.

### 3. Per-Extension Mode
- **3a.** Navigate to "Per-Extension Mode" and press `Enter`. Wait 1s. Capture. Verify: title "Per-Extension Mode", category headers (Config, Docs, Scripts, Web), extensions with current editors.
- **3b.** Press `Tab`. Capture. Verify: first extension changed editor, "1 change(s) pending" shown.
- **3c.** Press `Tab` again. Capture. Verify: editor cycled to next option.
- **3d.** Press `BTab` (Shift+Tab). Capture. Verify: editor cycled backward.
- **3e.** Press `j j j`. Capture. Verify: cursor moved down 3 positions.
- **3f.** Test search: press `/`. Capture. Verify: filter bar appears with prompt.
- **3g.** Type `py`. Capture. Verify: filtered to show .py extension, match count shown.
- **3h.** Press `Escape`. Capture. Verify: filter cleared, all extensions visible again.
- **3i.** Press `/`, type `json`, press `Enter`. Capture. Verify: filter applied and stays, showing json-related extensions.
- **3j.** Press `Escape`. Capture. Verify: filter cleared.
- **3k.** Make a change (Tab on an extension), then press `Enter` or `a`. Capture. Verify: navigated to Confirm screen.

### 4. Confirm Screen
- **4a.** Verify: title "Confirm Changes" with "DRY RUN" badge, table showing extensions with Current and New columns, change count.
- **4b.** Press `s`. Capture. Verify: navigated to Save Profile screen.

### 5. Save Profile Flow
- **5a.** Verify: title "Save Profile", input prompt visible.
- **5b.** Type `test-profile-1`. Capture. Verify: name appears in input field.
- **5c.** Press `Enter`. Wait 1s. Capture. Verify: success message 'Profile "test-profile-1" saved!' shown.
- **5d.** Press `Enter` to return. Capture. Verify: back at menu.

### 6. Load Profile
- **6a.** Navigate to "Load Profile" and press `Enter`. Capture. Verify: title "Profiles", "test-profile-1" appears in list with association count.
- **6b.** Press `Enter` to load it. Capture. Verify: navigated to Confirm screen with the profile's changes.
- **6c.** Press `Escape` to cancel. Capture. Verify: back at menu.

### 7. Delete Profile
- **7a.** Navigate to "Load Profile" and press `Enter`. Capture.
- **7b.** Press `d` to delete. Capture. Verify: profile removed from list (or "No profiles" message).
- **7c.** Press `Escape`. Capture. Verify: back at menu.

### 8. Default Browser
- **8a.** Navigate to "Default Browser" and press `Enter`. Wait 1s. Capture. Verify: title "Default Browser", current browser shown, list of browsers.
- **8b.** Press `j` to move cursor. Capture. Verify: cursor moved.
- **8c.** Press `Escape`. Capture. Verify: back at menu.

### 9. Revert Screen
- **9a.** Navigate to "Revert to Backup" and press `Enter`. Wait 1s. Capture. Verify: title "Revert to Backup", shows backup list or "No backups" message.
- **9b.** Press `Escape`. Capture. Verify: back at menu.

### 10. Apply Changes (Dry Run)
- **10a.** Go to Per-Extension mode, make a change, press Enter to confirm, then press Enter/y to apply.
- **10b.** Wait 2s. Capture. Verify: "Applying Changes" or "Changes Applied" title, progress/results shown with checkmarks.
- **10c.** When done, press `Enter`. Capture. Verify: back at menu.

### 11. Quit
- **11a.** Press `q`. Verify: tmux session ended or app exited.

## Cleanup

Always run at the end, even if tests fail:
```bash
tmux kill-session -t openwith-test 2>/dev/null
```

Also clean up any test profiles created:
```bash
rm -f ~/.config/openwith/profiles/test-profile-1.json
```

## Reporting

After all tests, print a summary table:

```
Test              Result
──────────────────────────
1a Menu render    PASS
1b Navigate down  FAIL - cursor didn't move
...
```

Include the captured output for any FAIL results so the issue can be debugged.

If $ARGUMENTS is provided, only run tests matching that prefix (e.g., `5` runs only section 5 Save Profile tests).
