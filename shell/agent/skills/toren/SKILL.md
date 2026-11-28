---
name: toren
description: >
  Help with a Toren Linux machine: Void Linux (glibc, runit) with Hyprland and
  the primary Quickshell desktop. Use for anything about this desktop or system:
  the bar, the dock, workspaces, the sidebars, the overview/launcher, settings,
  keybinds, windows and monitors in Hyprland, wallpaper and colours, updates
  (toren update), installing apps, themes and plugins (toren-store), services
  and logs (runit), and diagnosing problems (toren doctor). Read the system's
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
gsettings get org.toren.shell desktop-shell     # 'desktop' (default) or 'toren'
```

- **`desktop`** (the default): Toren's primary Quickshell desktop, shipped at
  `/usr/share/toren/shell-desktop/`. Read [`desktop.md`](desktop.md).
- **`toren`**: Toren's own shell at `/usr/share/toren/shell/`. Only then do
  [`bar.md`](bar.md), [`gui.md`](gui.md) and [`plugins.md`](plugins.md) apply.

Check this before changing anything about the bar, dock or panels: the two
desktops have different files and settings.

## Rules

1. **Never edit shipped files.** `/usr/share/toren/` is replaced on every
   `toren update`. Change things through the user's files:
   - the desktop settings: `~/.config/illogical-impulse/config.json` (or the
     settings window, Super+, → Desktop; the desktop reloads the file on change),
   - Hyprland: `~/.config/toren/hyprland/user.lua` (loaded after Toren's
     config, so it wins; `user.conf` only on a Hyprland older than 0.57),
   - Toren's settings: `toren config get|set SCHEMA KEY [VALUE]`.
2. **Keybindings are Toren's** (`/usr/share/toren/hyprland/binds.lua`) and
   reach the desktop through `toren ipc call toren <function>`. Do not change
   the user's shortcuts unless they ask; a rebind made in Settings → Keybinds
   is written to `~/.config/toren/hyprland/rebinds.lua`.
3. **One writer per setting.** A setting has one owning tool; do not hand-edit
   the stores the daemons write (`shell.json`, `plugins.json`).
4. **Prefer the command over the file.** `toren`, `toren-store`,
   `toren-wallpaper` and `toren-agent` each have `--help`.
5. **Root changes need the user.** Anything under `/etc`, `/usr` or a system
   service (`sudo sv ...`, `sudo xbps-install ...`) is the user's call: say what
   you would run and why before running it.
6. **Hyprland is 0.57 with a Lua config.** The old text commands fail:
   `hyprctl keyword ...` does nothing and `hyprctl dispatch workspace 3` is an
   error. Write a dispatch as Lua, `hyprctl dispatch 'hl.dsp.focus({ workspace = "3" })'`,
   and an option change as `hyprctl eval 'hl.config({ general = { gaps_in = 4 } })'`.
   Read state with `hyprctl -j clients|workspaces|monitors|binds` as before.

## Guides

| Guide                                                              | Read it for                                                                                                                                       |
| ------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------- |
| [`desktop.md`](desktop.md)                                         | The default desktop: bar, dock, workspaces, sidebars, overview, settings, Hyprland look, keybind functions, and how Toren manages overlays.       |
| [`commands.md`](commands.md)                                       | Every `toren` command in one table: layouts, power and charge limit, GPU, kernels, services, packages, repair. |
| [`update.md`](update.md)                                           | `toren update`: how a machine follows dev-unstable, the CI gate, `--check`/`--wait`/`--force`, system packages, refreshing the desktop, rollback. |
| [`store.md`](store.md)                                             | Installing and removing apps, themes, bar styles and plugins with `toren-store`.                                                                  |
| [`diagnostics.md`](diagnostics.md)                                 | Something is broken: `toren doctor` (and `--json`), logs, services, a desktop that will not load.                                                 |
| [`prowl.md`](prowl.md)                                             | `toren-prowl`: cited SQLite code search, secret masking, MCP wiring, and Zero-LLM smart routing.                                                  |
| [`bar.md`](bar.md), [`gui.md`](gui.md), [`plugins.md`](plugins.md) | Only when `desktop-shell` is `toren`: Toren's own shell.                                                                                          |

## First look at a machine

```sh
toren doctor --json          # every check, machine-readable
toren version -v             # edition and the dev-unstable commit applied
toren update --check         # is a newer Toren available?
gsettings get org.toren.shell desktop-shell
tail -n 50 ~/.local/state/toren/quickshell.log
```
