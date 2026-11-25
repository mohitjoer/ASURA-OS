# Shell Plugins Guide

> **Toren's own shell only.** This guide applies when
> `gsettings get org.toren.shell desktop-shell` is `'toren'`. The default
> desktop is: see [`desktop.md`](desktop.md).

A shell plugin is a native widget extension running inside Quickshell. You author the logic and views;
Toren provides the design system tokens, placement framework, and settings persistence so every plugin
feels like a native element of the OS.

---

## The End-to-End Plugin Workflow

Follow this 4-step workflow to build and deploy plugins safely:

```
1. Scaffold:  toren plugin new <id> [--bar|--desktop|--popout]
2. Implement: Code UI with qs.PluginKit + Tokens/Ink
3. Audit:     toren plugin audit <path>
4. Install:   toren plugin install <path> [--bar] --yes
              toren plugin enable <id>
              toren-shell plugins reload
```

### 1. Scaffold

Run `toren plugin new <id>` (default target `--bar`, or `--desktop`, `--popout`).
This creates the plugin folder under `$(xdg-user-dir DOCUMENTS)/toren-plugins/<id>/` (or fallback `~/toren-plugins/<id>/`):

```
<plugin-id>/
  manifest.json        # Plugin metadata, dependencies, capabilities, settings schema
  service/Main.qml     # Headless background logic & process monitoring (no visual items)
  content/Widget.qml   # Visual UI widget rendered in the host surface
  content/Panel.qml    # Optional: flyout popup panel opened on widget click
  bin/                 # Optional helper scripts (shebang + executable bit required)
  README.md            # Purpose, capabilities, network hosts, and commands declared
  LICENSE              # License file
  assets/              # Preview captures (preview-widget.png)
```

### 2. Implement Using `qs.PluginKit`

All plugin UI must strictly use `qs.PluginKit` components and monochrome ink tokens:

```qml
import QtQuick
import Quickshell
import qs.PluginKit
import qs.PluginKit.Singletons

Item {
    id: root

    // Host injected properties:
    property string density: "glyph"     // or "full" in panel
    property real widthBudget: 220
    property real s: 1.0                // Scale multiplier
    property bool active: true
    property var pluginApi              // Standard API handle

    implicitWidth: contentRow.implicitWidth + Tokens.s2 * 2
    implicitHeight: 28 * s

    Row {
        id: contentRow
        anchors.centerIn: parent
        spacing: Tokens.s1

        Text {
            text: pluginApi.pluginSettings.label || "Toren"
            color: Tokens.ink
            font.family: Tokens.ui
            font.pixelSize: Tokens.fSmall
        }
    }

    MouseArea {
        anchors.fill: parent
        onClicked: {
            if (pluginApi.panelOpen !== undefined)
                pluginApi.togglePanel();
        }
    }
}
```

The injected `pluginApi` provides:

- `pluginApi.mainInstance`: Handle to your running `service/Main.qml` instance.
- `pluginApi.pluginSettings`: Resolved settings object declared in `manifest.json`.
- `pluginApi.saveSetting(key, value)`: Persists settings to the daemon.
- `pluginApi.stateDir`: Safe writable directory (`$XDG_STATE_HOME/toren/plugins/<id>`).
- `pluginApi.pluginDir`: Read-only path to installed plugin root (for resolving `bin/` scripts).
- `pluginApi.togglePanel()`, `openPanel()`, `closePanel()`: Controls bar popup panels.

### 3. Static Audit Verification (`toren plugin audit`)

Before installing, run static analysis:

```bash
toren plugin audit ~/toren-plugins/<id>
```

The audit is enforced by `desktop/cmd/toren/plugin_audit.go`. It inspects code without executing it.
Any blocking finding exits with non-zero and prevents installation.

### 4. Install, Enable, and Reload

Install the audited plugin, enable it, and notify the live shell:

```bash
# Install into ~/.local/share/toren/plugins/<id>
toren plugin install ~/toren-plugins/<id> --bar --yes

# Enable on the chosen surface
toren plugin enable <id>

# Reload plugins in the live Quickshell instance immediately
toren-shell plugins reload
```

---

## Static Audit Rules (R1–R11)

These 11 rules are strictly enforced by `toren plugin audit` and `toren plugin install`:

| Rule    | Name                   | Severity        | Enforced Invariant                                                                                                                                                                                                                                                                                                                                    |
| ------- | ---------------------- | --------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **R1**  | **Place**              | Blocking        | Authored in `~/toren-plugins/<id>/`. Never author directly inside `~/.local/share/toren/plugins/` (install destination) or `/usr/share/toren/` (package base).                                                                                                                                                                                        |
| **R2**  | **Shape**              | Blocking        | Manifest at root, `service/Main.qml` (headless), `content/Widget.qml` (view). No symlinks anywhere in tree. Helper scripts must be in `bin/` with shebang + exec permission.                                                                                                                                                                          |
| **R3**  | **Id**                 | Blocking        | Matches regex `^[a-z0-9][a-z0-9-]*$`. Must never collide with reserved built-in widget IDs (`clock`, `workspaces`, `cpu`, `gpu`, `memory`, `status`, etc.).                                                                                                                                                                                           |
| **R4**  | **Imports**            | Blocking        | Allowed imports: `QtQuick*`, `Quickshell*`, `qs.PluginKit`, `qs.PluginKit.Singletons`, and local relative files within the folder. **Never** import `shell.*`, internal `qs.*` modules, or use climbing relative imports (`../`).                                                                                                                     |
| **R5**  | **Settings**           | Blocking        | All settings declared in `manifest.json` under `metadata.settings`. Read via `pluginApi.pluginSettings`; write via `pluginApi.saveSetting(k, v)`. **Never** write to `shell.json` or `plugins.json` directly.                                                                                                                                         |
| **R6**  | **Commands**           | Blocking / Warn | External binaries executed must be in `bin/` or declared in `dependencies.commands` (or match the system allowlist: `toren`, `toren-shell`, `notify-send`, `xdg-open`, `wl-copy`, `cat`, `grep`, `jq`, etc.). **Zero privilege escalation** (`sudo`, `doas`, `su` blocked). `pkexec` allowed ONLY if exact command matches `capabilities.privileged`. |
| **R7**  | **Network**            | Blocking / Warn | Every contacted host must be declared in `capabilities.network`. **Pipe-to-shell is strictly forbidden** (`curl ... \| sh`, `wget ... \| bash`, `eval $(curl ...)` are blocking violations).                                                                                                                                                          |
| **R8**  | **Files**              | Blocking / Warn | Writable state restricted to `pluginApi.stateDir` (`$XDG_STATE_HOME/toren/plugins/<id>`), cache dir, or `/tmp`. Never access or touch `~/.ssh`, `/etc`, or shell startup scripts (`.bashrc`, `.zshrc`, `config.fish`).                                                                                                                                |
| **R9**  | **Shell Injection**    | Warning         | No dynamic `sh -c` / `bash -c` string concatenations (`"sh", "-c", "cmd " + arg`). Always pass safe argument arrays (`command: ["cmd", arg1, arg2]`).                                                                                                                                                                                                 |
| **R10** | **Secrets & Binaries** | Blocking        | Zero hardcoded API keys, private keys, or tokens (`ghp_`, `sk_`, private key headers). Zero compiled ELF/Mach-O binaries, `.so`, or `.dylib` files. Maximum plugin tree size: 20 MB.                                                                                                                                                                  |
| **R11** | **Honesty**            | Blocking        | Manifest must set `"official": false` for community plugins. Must specify valid `author` (`Name <email>`) and README documenting all network and system permissions.                                                                                                                                                                                  |

---

## Manifest Reference (`manifest.json`)

```json
{
  "id": "pomodoro",
  "name": "Pomodoro Timer",
  "version": "1.0.0",
  "author": "Developer <dev@example.com>",
  "official": false,
  "entryPoints": {
    "main": "service/Main.qml",
    "widget": "content/Widget.qml",
    "panel": "content/Panel.qml"
  },
  "dependencies": {
    "commands": ["paplay"]
  },
  "capabilities": {
    "network": ["api.example.com"],
    "privileged": []
  },
  "metadata": {
    "settings": {
      "duration": {
        "type": "int",
        "default": 25,
        "title": "Work Duration (minutes)"
      },
      "sound": {
        "type": "toggle",
        "default": true,
        "title": "Alert Sound"
      }
    }
  }
}
```

---

## Managing Plugins CLI

```bash
# List all installed plugins and their hosts
toren plugin list

# Run audit on a plugin directory
toren plugin audit ~/toren-plugins/pomodoro

# Install and enable on the bar
toren plugin install ~/toren-plugins/pomodoro --bar --yes

# Enable or disable
toren plugin enable pomodoro --bar
toren plugin disable pomodoro

# Uninstall
toren plugin remove pomodoro

# Live reload shell plugins without restarting Quickshell
toren-shell plugins reload
```
