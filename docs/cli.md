# The `toren` Control Tool (`toren`)

`toren` is the primary control CLI and system orchestration binary for Toren Linux.
It lives in `desktop/cmd/toren/` and is installed at `/usr/bin/asura`.

The command surface acts as the single source of truth for querying state, triggering updates,
managing hardware profiles, reconfiguring themes, applying settings, and running diagnostics.

---

## Command Reference

```text
asura <command> [flags]
```

### Core Operations & Lifecycle

| Command | Usage | Description |
|---|---|---|
| `update` | `toren update [--force]` | Pulls latest Toren system tree, updates packages via XBPS, and refreshes the desktop. |
| `rollback` | `toren rollback [--system]` | Re-materializes the previous config revision. With `--system`, stages Btrfs boot rollback. |
| `materialize` | `toren materialize` | Re-lays the baseline configuration and assets, replacing shipped defaults while preserving user overrides. |
| `reset` | `toren reset PATH \| --desktop \| --restore BACKUP` | Drops user edits for a path, or resets desktop configuration back to defaults with dated backup (`--desktop`). |
| `reinstall` | `asura reinstall` | Re-applies shipped Toren files for currently installed commit without updating newer code. |
| `recovery` | `asura recovery` | Interactive text menu on TTY (diagnostics, desktop reset, reinstall, snapshots, safe mode, logs). |
| `status` | `toren status` | Reports current git revision, system paths, and configuration drift. |
| `doctor` | `toren doctor` | Comprehensive health check: config drift, portals, runit user services, system services, and hardware power. |
| `version` | `asura version` | Prints `toren` version and compilation metadata. |

---

### Hardware, Power & Graphics

| Command | Usage | Description |
|---|---|---|
| `hardware` | `toren hardware report\|match\|lint-profiles\|capture [flags]` | Inspects hardware detection, matches machine profiles from `system/hardware/`, and captures sysfs fixtures. |
| `gpu` | `toren gpu [--json]` | Inspects integrated and discrete GPUs, reports driver status, and executes commands offloaded to discrete graphics. |
| `power` | `toren power [profile]` | Manages CPU governors (performance, powersave), EPP energy preferences, and hybrid GPU runtime power (D3cold). |
| `hibernate` | `toren hibernate status\|setup\|remove` | Opt-in Btrfs swapfile (@swap) creation, resume offset calculation, GRUB and dracut configuration for hibernation. |
| `fingerprint` | `toren fingerprint status\|enroll\|verify\|enable\|disable` | Opt-in fingerprint authentication via fprintd; safely inserts sufficient auth into sudo, hyprlock, and polkit. |
| `kernel` | `toren kernel list\|set\|remove` | Manage Void Linux kernel series (`linux`, `linux-lts`, `linux-mainline`), update GRUB default, and verify NVIDIA headers. |
| `vm` | `toren vm windows\|list\|launch\|stop` | Manage local VMs and run guided Windows 11 setup with TPM 2.0, UEFI, 4 cores, and SPICE integration. |
| `hook` | `toren hook run\|list` | Manage and trigger user extension hooks (`pre-update`, `post-update`, `post-theme`, `post-login`). |
| `share` | `toren share send\|receive\|devices` | Tailscale Taildrop file transfer: send files to peers, receive into `~/Downloads`, and notify desktop. |
| `enable` | `toren enable [service]` | Enables an optional service (`fingerprint`, `docker`, `bluetooth`, `printing`, `ssh`, `libvirt`, `tailscale`). |
| `disable` | `toren disable [service]` | Disables an optional service and safely rolls back its configuration. |
| `brightness` | `toren brightness [set\|get\|list] [value]` | Direct backlight control for screens and keyboard illumination without group permission friction. |

---

### Desktop, Shell & Compositor

| Command | Usage | Description |
|---|---|---|
| `config` | `toren config get\|set\|export [key] [val]` | Reads and modifies Toren desktop settings stored in GSettings / dconf. |
| `state` | `toren state` | Outputs live desktop state as JSON (active windows, workspaces, system tray). |
| `windows` | `toren windows` | Lists switchable windows sorted by most recently focused order. |
| `wm` | `toren wm status\|caps\|act` | Compositor integration seam querying capabilities and window management actions. |
| `ipc` | `toren ipc <verb> [args]` | Directly calls running Quickshell IPC surfaces. |
| `refresh` | `toren refresh` | Re-renders and applies desktop theme and wallpaper tokens to the active session. |
| `reload` | `toren reload` | Sends reload signals to Hyprland, `torend`, and Quickshell. |
| `menu` | `toren menu` | Opens the Toren application and action menu. |
| `apps` | `toren apps` | Lists indexed desktop applications as JSON for the launcher. |
| `icons` | `toren icons` | Dumps indexed icon themes and glyph mappings as JSON. |
| `binds` | `toren binds` | Outputs keybinding catalogue as JSON for cheat sheets and diagnostics. |

---

### Theming, Plugins & Packages

| Command | Usage | Description |
|---|---|---|
| `theme` | `toren theme generate\|render [image]` | Generates monochromatic and accent color palettes from an image using `matugen`. |
| `plugin` | `toren plugin new\|audit\|install\|enable\|remove` | Manages Quickshell UI plugins: scaffolds templates, audits safety rules, and manages extensions. |
| `pkg` | `toren pkg install\|remove\|search <name>` | Void Linux package operations wrapper over `xbps-*`. |
| `channel` | `toren channel [name]` | Manages the XBPS repository channel followed by the system. |
| `greeter` | `toren greeter enable\|disable\|test` | Configures and tests SDDM login greeter theme. |
| `install` | `toren install` | Launches the guided Toren Linux system installer GUI. |
| `role` | `toren role switch <role>` | Switches system roles (developer, security, gaming) via native container bridges. |
