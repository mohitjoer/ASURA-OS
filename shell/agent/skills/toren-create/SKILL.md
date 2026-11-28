---
name: toren-create
description: >
  Self-create new desktop apps, Quickshell shell plugins, and themes on Toren
  Linux directly from natural-language prompts.
---

# Creating Apps, Shell Plugins, and Themes on Toren Linux

Follow [`../toren/create-skills.md`](../toren/create-skills.md) when a user asks to build a new app, shell plugin, or theme from a natural-language description:

- **Shell Plugins**: Generate `manifest.json` + QML under `~/toren-plugins/<id>/`, run `toren plugin audit`, then `toren plugin install` and `toren plugin enable`.
- **Themes**: Generate theme color tokens / wallpaper metadata and apply through Toren's theme/config commands.
- **Desktop Apps**: Install executable under `~/.local/bin/<app>` and create `~/.local/share/applications/<app>.desktop` so it is immediately launchable from the desktop overview and dock.
