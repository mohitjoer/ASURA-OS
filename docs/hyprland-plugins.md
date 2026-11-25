# Hyprland Compositor Plugins

> **Status:** not available on Toren Linux yet. Void has no packaged Hyprland
> plugins, and Toren does not build them, so Settings has no Plugins page. This
> page describes the design for when they are added.

Hyprland compositor plugins are native shared libraries (`.so`) loaded directly into the
compositor process space. They provide deep compositor-level capabilities that shell
widgets cannot provide on their own: window title bars (`hyprbars`), acrylic / glass effects
(`hyprgrass`), window focus flashes (`hyprfocus`), dynamic cursors, and keyboard sounds.

This document describes how Hyprland plugins are built, supervised, and configured in Toren Linux.
For shell plugins written in QML / JavaScript that run inside Quickshell, see
[`docs/plugins.md`](plugins.md).

---

## Architecture & ABI Stability

Hyprland plugins are strictly ABI-locked to the exact build and git hash of the running
Hyprland compositor. When Hyprland updates via XBPS, any previously compiled plugin `.so`
built against a prior Hyprland version will fail to load until recompiled against the new headers.

Toren manages this lifecycle through:

1. **The Hub Backend (`toren-hubd`)**: `toren-hubd desktop plugins list` interrogates the
   installed plugins, checks their build hashes against the current compositor hash, and reports
   ABI status (`compatible`, `mismatch`, or `unbuilt`).
2. **Settings Hub GUI (`shell/hub/pages/PluginsPage.qml`)**: The Plugins page under
   _Settings > Compositor > Plugins_ displays status cards for every plugin on the system,
   providing one-click **Rebuild**, **Docs**, and **Remove** actions.
3. **Hyprland Plugin Manager (`hyprpm`)**: Upstream plugin toolchain integration used to
   clone repositories carrying a `hyprpm.toml`, build them against system headers, and stage
   them in `~/.local/share/hyprpm/`.

---

## Included & Curated Plugins

Toren Linux supports standard curated Hyprland plugins:

| Plugin            | Repository                      | Description                                            |
| ----------------- | ------------------------------- | ------------------------------------------------------ |
| `hyprbars`        | `hyprwm/hyprland-plugins`       | Client-side title bars with window controls and titles |
| `hyprfocus`       | `hyprfocus`                     | Window flash and animation on focus change             |
| `hyprtrails`      | `hyprwm/hyprland-plugins`       | Smooth trail visual effect trailing active windows     |
| `dynamic-cursors` | `VirtCode/hypr-dynamic-cursors` | Realistic cursor tilt, stretch, and momentum physics   |

---

## Configuration & Settings Flow

Compositor plugin settings ride Toren's hypr draft system:

1. When a plugin loads, its configuration keys are exported by the compositor.
2. In Settings Hub (_Plugins_), controls adjust properties (e.g. `plugin:hyprbars:bar_height`,
   `plugin:hyprbars:bar_color`).
3. On save, `toren-hubd` renders committed changes to `~/.config/toren/hyprland/settings.conf`.
4. Compositor reloading (`hyprctl reload`) activates the updated settings immediately without
   restarting the session.

---

## Adding Custom Plugins

To add a new Hyprland plugin:

1. In the Settings Hub, open **Plugins** and click **Add from Git**.
2. Enter the URL of a Git repository containing a valid `hyprpm.toml`.
3. Toren invokes the build chain, verifies header compatibility, and generates the `.so`.
4. Once built, the plugin appears in the plugins roster and can be enabled or configured immediately.
