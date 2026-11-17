# The ASURA Menu (`asura menu`)

The ASURA Menu is the single "do everything" command palette on **Super+Alt+Space**:
one unified menu tree that coordinates applications, system setup, styling, package
management, system updates, screen capture, and session power controls.

It is built directly into the existing Quickshell desktop launcher as a dedicated
`menu` mode, ensuring zero visual drift, full keyboard navigability, fuzzy search across all
categories and leaves, and strict compliance with ASURA's monochrome design system.

---

## Accessing the Menu

- **Keyboard Shortcut**: Press `SUPER ALT, Space` to toggle the menu open or closed.
- **Launcher Mode Key**: Open the standard launcher (`SUPER, Space`) and click the `MENU` mode button.
- **Command Line**: Run `asura menu` from any terminal or script to dispatch the menu toggle over IPC.

```sh
# Open or toggle the ASURA menu from the CLI
asura menu
```

---

## Menu Tree & Capabilities

Every leaf in the menu routes to a native tool or deep-links into the relevant Asura Hub settings page.
If a tool is not installed on the running system, the entry is visibly marked as unavailable rather
than failing silently.

| Category | Leaf Action | Target Command / Dispatch | Description |
|---|---|---|---|
| **Apps** | Applications | Mode `all` in launcher | Switch directly to the full desktop applications grid |
| **Learn** | Keybindings Cheatsheet | `asura ipc call asura toggleCheatsheet` | Interactive keyboard shortcuts reference legend |
| | ASURA Documentation | `asura-app browser https://asuraos.dev/docs` | System architecture records and user guides |
| **Style** | Theme & Palette | `asura-shell hub open desktop` | Monochromatic themes, light/dark variants, and accent tuning |
| | Wallpaper Picker | `asura-wallpaper wallpaper ui` | Interactive graphical wallpaper selector |
| | Toggle Wallpaper Colors | `asura-wallpaper wallpaper random` | Match desktop palette dynamically to active wallpaper |
| **Setup** | Wi-Fi & Network | `asura-shell hub open connections` | Manage wireless and wired network connections |
| | Bluetooth | `asura-shell hub open connections` | Scan, pair, and configure Bluetooth peripherals |
| | Displays & Monitors | `asura-shell hub open displays` | Multi-monitor layout, resolution, refresh rate, and scaling |
| | Audio Settings | `asura-shell hub open audio` | Sound sinks, volume levels, microphones, and equalizer |
| | Power Profile | `ghostty -e asura power` | Performance, balanced, and power-saver CPU/platform profiles |
| | GPU & Graphics | `ghostty -e asura gpu` | Dedicated GPU inspection, temperature, and prime offload |
| | Keyboard & Touchpad | `asura-shell hub open input` | Keyboard layouts, repeat delays, and touchpad gestures |
| **Install / Remove** | Software Store | `asura-store` | Visual showroom for desktop applications, themes, and plugins |
| | Package Manager | `ghostty -e apm` | Fast terminal-based package search and installation via `apm` |
| **Update** | ASURA Update | `ghostty -e sh -c "asura update; ..."` | Supervised system snapshot, overlay update, and rematerialize |
| | Full System Update | `ghostty -e sh -c "sudo xbps-install -Syu; asura update; ..."` | Complete base system update through XBPS package manager |
| | System Rollback | `ghostty -e sh -c "asura rollback; ..."` | Instantly roll back to previous snapshot revision |
| | System Diagnostics | `ghostty -e sh -c "asura doctor; ..."` | Run doctor health checks and report base-to-overlay drift |
| **Capture** | Capture Region | `asura-shot` | Interactive area snipping tool with annotation |
| | Capture Screen | `asura ipc call asura toggleScreenshot` | One-shot full-display screen capture |
| | Screen Recording | `asura-cmd-screenrecord` | Screen recording to video file |
| **System** | Lock Screen | `asura-shell lock` | Lock the session via Hyprlock |
| | Suspend | `loginctl suspend` | Suspend system to RAM |
| | Log Out | `hyprctl dispatch exit` | Terminate the current Wayland session |
| | Reboot | `loginctl reboot` | Gracefully restart the computer |
| | Shut Down | `loginctl poweroff` | Safely power off the computer |

---

## Implementation Architecture

1. **State Coordination (`ShellState.qml`)**:
   `ShellState` maintains `launcherOpen` and `launcherInitialMode`. Requesting `togglePopover("menu")` sets `launcherInitialMode = "menu"` and reveals the surface.
2. **IPC Integration (`shell/shell.qml`)**:
   The root IPC handler dispatches `togglePopover("menu")` into `toggleSurface("menu")`.
3. **Provider Layer (`shared/providers/menu/`)**:
   `Menu.qml` registers as provider `menu` with prefix `:menu`. It probes binary availability asynchronously on load and filters entries via fuzzy match against `name`, `category`, and `desc`.
4. **CLI Entry Point (`desktop/cmd/asura/menu.go`)**:
   The `asura menu` subcommand forwards calls directly to `asura ipc call asura togglePopover menu`.
5. **Internationalization (`shell/i18n/data/`)**:
   All category headers, leaf names, and descriptions are translated through the 35-language catalogue.
