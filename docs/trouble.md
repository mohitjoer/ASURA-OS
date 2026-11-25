# Toren Linux Troubleshooting and Recovery Guide

When a machine fails to boot graphically, displays a blank screen, or has broken desktop settings, Toren provides a complete set of safe, reversible recovery mechanisms designed to get you back into a working environment quickly.

---

## 1. Getting a TTY (Recovery of Last Resort)

If the graphical desktop or login greeter fails to start, crash loops, or hangs:

1. Press **`Ctrl + Alt + F2`** (or **`Ctrl + Alt + F3`** through **`F6`**).
2. The display switches immediately to a text virtual console (TTY).
3. Log in with your username and password.
4. From here, you have full terminal access and can run all Toren recovery tools.

To switch back to the graphical session when working:

- Press **`Ctrl + Alt + F1`** (or **`Ctrl + Alt + F7`** depending on the display manager).

---

## 2. Guided Recovery: `toren recovery`

On a TTY or in any terminal, run:

```sh
toren recovery
```

`toren recovery` displays a clear, guided text menu that requires no graphical environment:

```
============================================================
                Toren System Recovery
============================================================
1) Run Doctor (system diagnostics & configuration drift)
2) Reset Desktop (restore defaults, back up current files)
3) Reinstall Toren (re-apply files for installed commit)
4) System Snapshots (list & restore Btrfs root snapshots)
5) Safe-Mode Session (start minimal recovery desktop)
6) Show Logs (inspect recent session and shell error logs)
7) Exit
------------------------------------------------------------
```

Each option clearly describes the action it will perform and requests explicit confirmation (`[y/N]`) before executing.

---

## 3. Desktop Reset: `toren reset --desktop`

If your desktop layout, Hyprland binds, or Quickshell widgets are broken:

```sh
# Reset user desktop configuration back to packaged defaults
toren reset --desktop
```

### What it does:

- Creates a dated backup under `~/.local/state/toren/backups/desktop-YYYYMMDD-HHMMSS/`.
- Backs up and resets:
  - `~/.config/illogical-impulse/config.json` (desktop shell settings)
  - `~/.config/toren/hyprland/` (user Hyprland overrides and custom binds)
  - `~/.config/toren/user_edits/` (materialized overlay overrides)
  - `~/.local/state/toren/` (shell state and dynamic caches)
  - `~/.local/state/quickshell/` (quickshell runtime cache)
- Re-materializes pristine base defaults.
- **Never touches user personal documents or files in `~`.**

### Restoring a backup:

If you want your previous settings back byte-for-byte:

```sh
toren reset --restore ~/.local/state/toren/backups/desktop-YYYYMMDD-HHMMSS
```

---

## 4. Reinstall Shipped Files: `toren reinstall`

If core desktop files or scripts were accidentally modified or deleted from `/usr/share/toren/`, `/usr/bin/`, or `/etc/`:

```sh
toren reinstall
```

- Reads the exact dev-unstable commit currently applied (`/usr/share/toren/applied-commit`).
- Re-applies the repository files for that exact commit without pulling anything newer.
- Restores all shipped shell components, desktop entries, and scripts cleanly.

---

## 5. Toren (Safe Mode) Session

If you need a graphical desktop to troubleshoot but your custom Hyprland configuration or shell is crashing:

1. At the login screen (SDDM/GDM), select the session menu.
2. Select **`Toren Linux (safe mode)`**.
3. Log in.

### What Safe Mode provides:

- A rock-solid, minimal Hyprland compositor session.
- **Terminal hotkey**: Press **`Super + Return`** to launch a terminal.
- **Quit hotkey**: Press **`Super + M`** to exit the session.
- No desktop shell (no Quickshell or heavy daemons).
- No user plugins or custom configs sourced.

---

## 6. System Rollback via Btrfs Snapshots

If a recent system update or package installation broke hardware drivers or the kernel:

```sh
# List available system root snapshots
toren rollback --system

# Restore a specific snapshot (e.g., pre-update-20261122-140000)
sudo toren rollback --system pre-update-20261122-140000
```

---

## 7. Inspecting System & Session Logs

Session and shell errors are logged directly to user state:

```sh
# Session startup and compositor logs
cat ~/.local/state/toren/session.log

# Quickshell desktop errors
cat ~/.local/state/toren/quickshell.log

# Safe mode session logs
cat ~/.local/state/toren/session-safe.log
```
