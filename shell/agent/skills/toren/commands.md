# The `toren` command

Every command has its own `--help`; read it before using a verb you have not
used. Commands that change the system ask for `sudo` themselves. Say what you
will run, and why, before running one (rule 5 in [`SKILL.md`](SKILL.md)).

## The desktop

| Command | What it does |
| --- | --- |
| `toren layout [status]` | The current workspace's layout. |
| `toren layout set tile\|scroll\|master\|monocle\|float\|default [--workspace ID]` | Set one workspace's layout (`default` follows Settings → Window Manager → Layout). Also in the power menu and on the bar's layout mark. |
| `toren layout next\|prev` | Cycle the current workspace's layout. |
| `toren theme set [NAME]` | Switch to a named look (wallpaper, light or dark, colours); no NAME lists them. |
| `toren theme current` | The palette in use. |
| `toren brightness get\|set V\|up\|down [--device panel\|kbd]` | Screen and keyboard backlight. |
| `toren ipc call toren FUNCTION [ARGS]` | Call the desktop the way a keybinding does (see [`desktop.md`](desktop.md)). |
| `toren binds` | The keybinding catalogue as JSON (what the cheatsheet and the Keybinds page show). |
| `toren wm status` / `toren wm act ID [ARGS]` | The compositor seam: version, capabilities, one action. |
| `toren refresh` | Apply installed changes to the running session (services, Hyprland, the shell, the daemons). |
| `toren reload` | Restart the running shell and daemon. |
| `toren-shell hub open [SECTION]` | Open the settings window (what Super+, does), e.g. `toren-shell hub open keybinds`. |

## Power and hardware

| Command | What it does |
| --- | --- |
| `toren power status` | Power supply, governor and profile. |
| `toren power profile [balanced\|performance\|battery]` | Get or set the power profile. |
| `toren power threshold [PCT\|reset]` | The battery charge limit. Kept in `/etc/toren/charge-limit` and applied again whenever a battery appears (a udev rule), on any laptop whose battery supports it. Also Settings → Desktop → System → Charging. |
| `toren power sleep [s2idle\|deep]` | The suspend mode. |
| `toren gpu status\|switch MODE\|run CMD` | GPUs, hybrid/integrated/discrete mode, run one command on the discrete GPU. |
| `toren hardware ...` | What the machine is and which hardware fixes (quirks) apply. |
| `toren hibernate status\|setup\|remove` | Hibernation (Btrfs swapfile, resume, GRUB, dracut; snapshot first). |
| `toren fingerprint status\|enroll\|verify\|enable\|disable` | Fingerprint login. |
| `toren kernel list\|set NAME\|remove NAME` | Void kernel series (`linux`, `linux-lts`, `linux-mainline`). |
| `toren firmware [update]` | Firmware updates (fwupd). |

## Software

| Command | What it does |
| --- | --- |
| `toren pkg install\|remove\|search\|update NAME...` | Packages through XBPS; an application id (`org.mozilla.firefox`) goes to Flatpak. |
| `toren-store` | The app store (apps, themes, bar styles, plugins): [`store.md`](store.md). |
| `toren bundle list\|show NAME\|install NAME...` | Install a curated set (gaming, dev, editors, apps, creative, office); a snapshot is taken first. |
| `toren enable NAME` / `toren disable NAME` / `toren enable --list` | Optional services: bluetooth, docker, fingerprint, libvirt, printing, ssh, tailscale. Enabling installs, starts and adds the user to the group. |
| `toren hold PKG...` / `toren unhold PKG...` | Keep packages back from updates (no arguments lists held ones). |
| `toren role list\|gamer\|dev\|...` | Larger role setups (gaming stack, development toolchain). |
| `toren vm list\|windows\|launch NAME\|stop NAME` | Virtual machines. |
| `toren share send FILE DEVICE\|receive\|devices` | Send files over Tailscale (needs `toren enable tailscale`). |
| `toren clean` | Remove orphaned packages and the package cache. |
| `toren mirror fastest\|list\|set NAME\|reset` | Which Void mirror downloads come from. |

## Updates, state and repair

| Command | What it does |
| --- | --- |
| `toren update [--check\|--wait\|--force]` | Update Toren (and with sudo, the system): [`update.md`](update.md). |
| `toren channel get` / `sudo toren channel set stable\|unstable` | Which branch updates follow. |
| `toren version [-v\|--commit]` | Edition and the commit applied. |
| `toren status [--json]` | Revision, paths, drift, and whether the shell and daemon are up. |
| `toren doctor [--json\|--fix]` | Check everything: [`diagnostics.md`](diagnostics.md). |
| `toren migrate [--pending\|--list]` | The one-time fixes this machine has not run yet (`toren update` runs them). |
| `toren reinstall` | Re-apply the files of the commit already installed (repairs deleted or edited shipped files, updates nothing). |
| `toren reset PATH` | Drop one user edit and restore that file from the base. |
| `toren reset --desktop` | Reset the desktop's configuration to Toren's defaults (backed up first); `--restore BACKUP` puts a backup back. |
| `toren rollback` | The previous config revision. `sudo toren rollback --system SNAPSHOT` boots a pre-update snapshot. |
| `toren snapshot [list\|create]` | Btrfs system snapshots. |
| `toren recovery` | A text menu for a TTY with no desktop: doctor, reset, reinstall, snapshots, safe mode, logs. |
| `toren config list\|get\|set\|reset SCHEMA KEY [VALUE]` | Toren's own settings (gsettings schemas). |
| `toren hook list\|run EVENT` | The user's hook scripts in `~/.config/toren/hooks/<event>/` (pre-update, post-update, post-theme, post-login). |
| `toren plugin new\|audit\|install\|enable ...` | Toren shell plugins: [`plugins.md`](plugins.md). |
| `toren manual` | The user manual. |
