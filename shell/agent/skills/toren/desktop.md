# The default desktop (shell-desktop)

When `gsettings get org.toren.shell desktop-shell` is `'desktop'` (the default),
the desktop is Toren's Quickshell desktop, shipped at
`/usr/share/toren/shell-desktop/` and started by the runit user service
`~/.config/service/quickshell` (`quickshell -p /usr/share/toren/shell-desktop`).

## What the user sees (Toren's layout)

- **Bar** (top, 30 px, plain): workspaces on the left (only the ones that
  exist; the active one underlined, an empty one a dot), media while a player
  has a track, the clock and date in the centre, then the status icons (they
  open the control menu), CPU/RAM and the battery at the right end.
- **Dock**: an island at the bottom with the pinned and running apps. It
  auto-hides while a window is focused and comes back at the bottom edge.
- **Control menu** (right sidebar): quick toggles, volume and brightness
  sliders, notifications, calendar.
- **Overview / launcher**: search apps, run commands, calculate.
- **Power menu** (session screen): lock, log out, suspend, reboot, power off,
  and a row that sets the current workspace's layout.
- **Cheatsheet**: the keybindings, grouped (no other pages).
- There is no left sidebar (AI chat and anime browser are not loaded).

## Workspace layouts

Each workspace has its own layout: `tile` (dwindle), `scroll`, `master`,
`monocle` or `float`; a workspace with none follows the default (Settings →
Window Manager → Layout). Change it with `toren layout set NAME` (or
`next`/`prev`), from the power menu's layout row, or by clicking the layout mark
after the bar's workspaces. The choice is kept in
`~/.config/toren/layouts.json` and written to
`~/.config/toren/hyprland/layouts.lua`; float floats the workspace's windows
and switching away tiles back the ones it floated.

## The settings window

There is **one** settings window: Super+, (or the control menu's gear, or
`toren-shell hub open [SECTION]`, or `toren-hub`). The desktop's own upstream
Settings app is not shipped. A sidebar lists the sections with icons; groups
of several fold open and shut, and the open page's tabs are rows under it. The
**Desktop** section holds the desktop's options in six parts: Bar & dock, Look,
Clock, Lock screen, Tools and System (system includes the battery charge limit).
Other sections: Devices (Displays, Connections, Input, Graphics & Power),
Lockscreen, Window Manager (gaps, borders, animations, default layout),
Keys & apps (Keybinds, App Overrides, Window Rules), System (Performance,
Session), About. Changes preview live; Save writes them, Revert drops them.

## Changing it

**User settings** live in `~/.config/illogical-impulse/config.json`, which the
desktop watches and reloads. The settings window's Desktop section edits the
same file. Useful keys:

| Key                           | What it does                                                             |
| ----------------------------- | ------------------------------------------------------------------------ |
| `dock.enable`                 | the dock on or off                                                       |
| `dock.pinnedOnStartup`        | `true` keeps the dock shown (auto-hide off)                              |
| `dock.pinnedApps`             | desktop-entry ids, e.g. `["com.mitchellh.ghostty", "thunar", "firefox"]` |
| `dock.monochromeIcons`        | tint the dock's icons                                                    |
| `bar.borderless`              | no boxes behind the bar's groups                                         |
| `bar.workspaces.showAppIcons` | app icons in the workspace indicator                                     |
| `sidebar.quickSliders.enable` | volume/brightness sliders in the control menu                            |

Edit it with a JSON tool that keeps the rest of the file (e.g. `jq`), or point
the user to Settings (Super+,), Desktop page.

**Looks:** `toren theme set` lists named looks (wallpaper, light or dark,
colours); `toren theme set NAME` applies one, recolouring the shell, terminal,
btop and GTK apps.

**Colours** come from the wallpaper (matugen): `toren-wallpaper wallpaper ui`
opens the picker, `toren-wallpaper wallpaper random` picks one. The palette is
written to `~/.local/state/quickshell/user/generated/colors.json`.

**Hyprland** (gaps, borders, rules, monitors, input) is 0.57 with a Lua config:
Toren's is `/usr/share/toren/hyprland/hyprland.lua`, which then loads, from
`~/.config/toren/hyprland/`, `shell.lua`, `monitors-auto.lua`,
`monitors-user.lua`, `settings.lua` (written by the settings window),
`layouts.lua` (`toren layout`), `rebinds.lua`, `toren.lua`, `user.lua` and
`vm-perf.lua`, in that order; the last one wins. Put the user's own changes in
`user.lua` (Lua, e.g. `hl.config({ general = { gaps_out = 8 } })`). The
desktop's look (blur, animations, layer rules) is
`/usr/share/toren/shell-desktop/toren/hyprland.lua`, loaded through `shell.lua`
while desktop is the shell. Apply with `hyprctl reload` or `toren refresh`. The
`.conf` files are only for a Hyprland older than 0.57. Commands to a running
Hyprland are Lua too (rule 6 in [`SKILL.md`](SKILL.md)).

**In a virtual machine** every service that shows a window (the desktop, the
settings window, the wallpaper picker, the lock screen) uses software rendering
(`LIBGL_ALWAYS_SOFTWARE=1`, decided once in
`/usr/share/toren/session/toren-graphics-env`); a value the user sets wins.

## Keybind functions

Keybindings run `toren ipc call toren <function>`, answered by
`shell-desktop/toren/TorenBridge.qml`: `togglePopover <name>`, `toggleOverview`,
`overviewDesktop <delta>`, `advanceSwitcher`, `reverseSwitcher`,
`closeSwitcher`, `toggleCheatsheet`, `session`, `ocr`, `toggleScreenshot`,
`toggleSurface <name>`, `reloadSettings`. Test one by running the same command.

## How Toren changes the desktop (for code changes, not user requests)

`shell-desktop/` in the Toren repository is the upstream tree **byte for byte** at a
pinned commit; `sh shell-desktop/toren/vendor.sh verify` checks it and CI runs
it. A file Toren changes becomes an **overlay**: listed in `OVERLAYS` in
`shell-desktop/toren/vendor.sh` and removed from `shell-desktop/toren/manifest.txt`.
Editing a vendored file without that fails CI, and `toren update` then keeps
every machine on the last commit that passed. Every change must also load:
CI's `desktop-load` job loads the desktop with Quickshell and fails on any QML
error.

## When the desktop will not load

`tail -n 60 ~/.local/state/toren/quickshell.log`. A "Reload failed" usually
names one file and line ("Cannot assign to non-existent property", "Type X
unavailable"). If the desktop keeps crashing, the service waits 5 s between
restarts; after fixing, `rm -f ~/.local/state/toren/desktop-quick-exits` and
`sv restart ~/.config/service/quickshell`.
