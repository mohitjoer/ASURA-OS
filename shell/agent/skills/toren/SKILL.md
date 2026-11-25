---
name: asura
description: >
  Help with a Toren Linux machine: Void Linux (glibc, runit) with Hyprland and
  the primary Quickshell desktop. Use for anything about this desktop or system:
  the bar, the dock, workspaces, the sidebars, the overview/launcher, settings,
  keybinds, windows and monitors in Hyprland, wallpaper and colours, updates
  (asura update), installing apps, themes and plugins (asura-store), services
  and logs (runit), and diagnosing problems (asura doctor). Read the system's
  state first; act through Toren's commands and the user's own files, never by
  editing shipped files.
---

# Toren Linux

Toren Linux is Void Linux x86_64 (glibc) with **runit** (not systemd), the
**Hyprland** compositor, and a **Quickshell** desktop. Packages come from
**XBPS** (`xbps-install`, `xbps-query`); there is no apt, no pacman, no
systemctl (see the `systemd-bridge` skill for the runit equivalents).

## Which desktop is running

The desktop shell is chosen by one setting:

```sh
gsettings get org.asura.shell desktop-shell     # 'desktop' (default) or 'asura'
```

- **`desktop`** (the default): Toren's primary Quickshell desktop, shipped at
  `/usr/share/asura/shell-desktop/`. Read [`desktop.md`](desktop.md).
- **`asura`**: Toren's own shell at `/usr/share/asura/shell/`. Only then do
  [`bar.md`](bar.md), [`gui.md`](gui.md) and [`plugins.md`](plugins.md) apply.

Check this before changing anything about the bar, dock or panels: the two
desktops have different files and settings.

## Rules

1. **Never edit shipped files.** `/usr/share/asura/` is replaced on every
   `asura update`. Change things through the user's files:
   - the desktop settings: `~/.config/illogical-impulse/config.json` (or its
     Settings app; it reloads the file on change),
   - Hyprland: `~/.config/asura/hyprland/user.conf` (sourced after Toren's
     config, so it wins),
   - Toren's settings: `asura config get|set SCHEMA KEY [VALUE]`.
2. **Keybindings are Toren's** (`/usr/share/asura/hyprland/binds.conf`) and
   reach the desktop through `asura ipc call asura <function>`. Do not change
   the user's shortcuts unless they ask.
3. **One writer per setting.** A setting has one owning tool; do not hand-edit
   the stores the daemons write (`shell.json`, `plugins.json`).
4. **Prefer the command over the file.** `asura`, `asura-store`,
   `asura-wallpaper` and `asura-agent` each have `--help`.
5. **Root changes need the user.** Anything under `/etc`, `/usr` or a system
   service (`sudo sv ...`, `sudo xbps-install ...`) is the user's call: say what
   you would run and why before running it.

## Guides

| Guide | Read it for |
|---|---|
| [`desktop.md`](desktop.md) | The default desktop: bar, dock, workspaces, sidebars, overview, settings, Hyprland look, keybind functions, and how Toren manages overlays. |
| [`update.md`](update.md) | `asura update`: how a machine follows dev-unstable, the CI gate, `--check`/`--wait`/`--force`, system packages, refreshing the desktop, rollback. |
| [`store.md`](store.md) | Installing and removing apps, themes, bar styles and plugins with `asura-store`. |
| [`diagnostics.md`](diagnostics.md) | Something is broken: `asura doctor` (and `--json`), logs, services, a desktop that will not load. |
| [`bar.md`](bar.md), [`gui.md`](gui.md), [`plugins.md`](plugins.md) | Only when `desktop-shell` is `asura`: Toren's own shell. |

## First look at a machine

```sh
asura doctor --json          # every check, machine-readable
asura version -v             # edition and the dev-unstable commit applied
asura update --check         # is a newer Toren available?
gsettings get org.asura.shell desktop-shell
tail -n 50 ~/.local/state/asura/quickshell.log
```
