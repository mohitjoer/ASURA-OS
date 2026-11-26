# Toren Linux Diagnostics & Crash Triage Playbook

## 0. Start here

```bash
toren doctor --json     # every check as {"ok", "sections": [{"section", "ok", "lines"}]}
gsettings get org.toren.shell desktop-shell   # 'desktop' (default) or 'toren'
toren version -v        # which Toren commit is applied
```

Read the sections with `"ok": false` first; each line says what is wrong.
`toren doctor` exits non-zero when any check fails.

### The default desktop will not load or misbehaves

```bash
tail -n 60 ~/.local/state/toren/quickshell.log
sv status ~/.config/service/quickshell
```

- `Reload failed` / `Failed to load configuration` names a QML file and line:
  that file in `/usr/share/toren/shell-desktop/` is broken. Do not edit it in
  place (an update overwrites it); `toren update` brings the fixed version once
  it is on dev-unstable, and `toren rollback` returns to the previous one.
- After a fix, clear the crash guard and restart:
  `rm -f ~/.local/state/toren/desktop-quick-exits; sv restart ~/.config/service/quickshell`.
- A setting that "does nothing": check `~/.config/illogical-impulse/config.json`
  is valid JSON (`jq . ~/.config/illogical-impulse/config.json`); the desktop ignores
  a file it cannot parse.
- Apps that will not open from the dock or launcher: run the app's command in
  a terminal to see its error. The dock starts apps from the desktop process,
  so a broken session environment there (no `WAYLAND_DISPLAY`, no session bus)
  shows in `quickshell.log`.

### The settings window (Super+,) does not open

Super+, runs `toren-shell hub open`, which asks the `torend` service to start
`quickshell -p /usr/share/toren/shell/hub.qml`. The window's own output is not
kept, so a window that dies at once looks like nothing happened.

```bash
sv status ~/.config/service/torend ~/.config/service/dbus
toren-shell hub open; echo $?          # non-zero: torend is not answering
pgrep -af hub.qml                      # a stale window holding the lock?
quickshell -p /usr/share/toren/shell/hub.qml   # start it by hand: its errors print here
```

- `torend` down: `sv restart ~/.config/service/torend`, then read
  `~/.local/state/toren/torend.log`.
- A stale `hub.qml` process with no window holds the single-instance lock:
  `pkill -f hub.qml`.
- `wl_display ... invalid arguments for wl_surface.attach` in a virtual
  machine: software rendering is off for that process. `toren-graphics-env`
  turns it on for every service; an old install without it needs `toren update`.
- Opening it by hand works but Super+, does not: the difference is the
  service's environment; compare `cat /proc/$(pgrep -f 'toren-shell daemon')/environ | tr '\0' '\n'`
  with the terminal's.

The rest of this playbook (sections 1-3) covers runit services, logs, and
Toren's own shell.

---

This playbook provides actionable steps for diagnosing and resolving Quickshell errors,
daemon failures, QML binding issues, and broken user configurations on Toren Linux.

---

## 1. Checking Service & Crash Logs (Void Linux `runit`)

Toren Linux uses `runit` for process supervision, not systemd. `systemctl` and
`journalctl` exist only as thin bridges to runit and the service logs (see the
`systemd-bridge` skill); prefer `sv` and the log files.

### Inspecting User and System Services

Session services run under the user's `runsvdir` instance:

```bash
# Check status of all active user session services
sv status ~/.config/service/*

# The user services:
#   ~/.config/service/dbus             (the session bus; the others wait for it)
#   ~/.config/service/quickshell       (the desktop shell)
#   ~/.config/service/torend           (the desktop daemon; opens the settings window)
#   ~/.config/service/toren-userd      (applies themes and fonts to GTK, Qt and Hyprland)
#   ~/.config/service/toren-wallpaper  (wallpaper and colours)
#   ~/.config/service/toren-idle       (dim, lock, screen off: hypridle)
#   ~/.config/service/toren-updates    (update and reboot notices)
#   ~/.config/service/pipewire, wireplumber, portal, portal-hyprland

# Check system services (hardware, power, dbus, seatd)
sudo sv status /var/service/*

# Restart a failing service
sv restart ~/.config/service/quickshell
sv restart ~/.config/service/torend
```

### Inspecting Crash Dumps & Logs

System and kernel messages are kept by socklog under `/var/log/socklog/`.
The desktop's own services log to `~/.local/state/toren/`: `quickshell.log`,
`torend.log`, `toren-idle.log`, `toren-wallpaper.log`, `toren-updates.log`. Quickshell crash environments and state files are
written to `~/.local/state/toren/`:

```bash
# List recent crash dumps and logs
ls -la ~/.local/state/toren/

# Inspect recent crash logs or core details
cat ~/.local/state/toren/quickshell.log 2>/dev/null
cat ~/.local/state/toren/toren-shell.log 2>/dev/null
```

### Reproducing QML Errors in an Interactive Test Harness

When the visual shell fails to start or crashes silently, run Quickshell manually in a terminal to capture QML binding warnings, missing module imports, or fatal syntax errors:

```bash
# Test the installed system shell
quickshell -p /usr/share/toren/shell

# Or test against a local repository checkout
quickshell -p shell/

# Test the Toren Hub independently
quickshell -p /usr/share/toren/shell/hub.qml
```

Look for:

- `ReferenceError: ... is not defined` -> Unimported module or missing singleton.
- `Cannot assign to read-only property ...` -> Mutating a read-only token or state binding.
- `Type ... unavailable` -> Broken relative component path or missing plugin dependency.

---

## 2. Static & Contract Verification

Before modifying QML or system code, run the built-in automated test suites:

### 1. Engine Load Verification (`shell/shell_test.sh`)

Verifies that both the desktop shell and the Hub compile and instantiate cleanly under the Quickshell Wayland engine:

```bash
sh shell/shell_test.sh
```

Expected output:

```
ok    — the shell loads with no errors
ok    — the Hub loads with no errors
```

### 2. QML Syntax & Linter Verification (`tests/contract/qml.sh`)

Runs `qmllint` across all QML files against the baseline:

```bash
sh tests/contract/qml.sh
```

Ensures zero new syntax-class errors or malformed property declarations.

### 3. Toren Doctor Diagnostic (`toren doctor`)

Validates system state, schema keys, keybindings, and portal compatibility:

```bash
# Human-readable report
toren doctor

# JSON report for programmatic inspection
toren doctor --json

# Automatically repair renamed keys in ~/.config/toren/shell.json
toren doctor --fix
```

What `toren doctor` inspects:

- **Portals**: Confirms `xdg-desktop-portal-hyprland` is active and `xdg-desktop-portal-gnome` is absent.
- **Keybinds**: Verifies the user's rebinds (`~/.config/toren/hyprland/rebinds.lua`) against the catalogue, `binds.json`.
- **Drift**: Pinpoints differences between `/usr/share/toren/` base templates and `~/.config/toren/user_edits/`.

---

## 3. Safe Recovery & Reloading

If a customized QML file or bad user edit breaks the shell, use the following recovery sequence:

### Step 1: Pinpoint the Broken File

Check for modified files in user overrides:

```bash
ls -la ~/.config/toren/user_edits/
```

### Step 2: Reset or Re-materialize

To revert a specific broken override back to package defaults:

```bash
# Drops the override from user_edits/ and restores the shipped base copy
toren reset <relative/path/to/file>

# `toren status` lists the edited files; reset one by the path it shows

# Everything the desktop reads, back to Toren's defaults (backed up first)
toren reset --desktop
```

Shipped files that were deleted or edited by hand come back with
`toren reinstall` (the installed commit, nothing newer).

To refresh all shipped templates while preserving user overrides:

```bash
toren materialize
```

### Step 3: Trigger Live Reload

Once fixed, reload the live components without rebooting:

```bash
# Signal user runit services (quickshell and torend)
toren reload

# Or trigger daemon-level shell surface reload
toren-shell reload

# Reload Hyprland compositor bindings and rules
hyprctl reload
```

---

## Common Crash Patterns & Solutions

| Symptom                             | Probable Cause                                                      | Fix                                                                                           |
| ----------------------------------- | ------------------------------------------------------------------- | --------------------------------------------------------------------------------------------- |
| Shell crashes immediately on launch | Missing QML import or syntax error in a modified file               | Run `quickshell -p shell/` to see the stack trace; run `sh tests/contract/qml.sh`.            |
| Black or missing bar widgets        | Non-existent widget ID in `shell.json` or invalid visibility key    | Run `toren-shell bar list --json` and `toren-shell bar defaults` to restore layout.           |
| Daemon fails to bind socket         | Stale `$XDG_RUNTIME_DIR/toren-shell.sock` or lockfile               | Verify service with `sv status ~/.config/service/torend`; clean up stale lock if daemon died. |
| Hub settings do not persist         | Multiple writers or manual editing of `shell.json`                  | Use `toren-shell` CLI or Hub UI to write settings; verify schema with `toren doctor`.         |
| Plugin fails to load on bar         | Failed static audit rule (e.g. internal import or forbidden binary) | Run `toren plugin audit <path>` and resolve blocking findings.                                |
