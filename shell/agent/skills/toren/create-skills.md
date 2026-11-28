# Self-Creating Apps, Shell Plugins, and Themes

Use these workflows when the user asks in natural language to create a new desktop app, Quickshell plugin, or theme on Toren Linux.

## 1. Creating a New Shell Plugin

1. Scaffold under `~/toren-plugins/<id>/` with:
   - `manifest.json` (`id`, `name`, `version`, `entry` QML file, `slot` such as `bar` or `sidebar`)
   - QML component (using `qs.modules.common` tokens: `Appearance`, `StyledText`, `MaterialSymbol`, etc.)
2. Audit and install using Toren's plugin pipeline:
   ```sh
   toren plugin audit ~/toren-plugins/<id>
   toren plugin install ~/toren-plugins/<id> --bar --yes
   toren plugin enable <id> --bar
   ```

## 2. Creating a New Theme

1. Define a theme JSON (colors, surface tokens, accent, font/radius overrides, optional wallpaper path) or generate palette tokens.
2. Apply and register via `toren` / `toren-hubd` / `toren-store` (`~/.config/toren/themes/` or `~/.config/illogical-impulse/config.json`).

## 3. Creating a New Desktop App

1. Build the app binary or script into `~/.local/bin/<app-id>`.
2. Add a desktop entry at `~/.local/share/applications/<app-id>.desktop` (`Exec=`, `Icon=`, `Type=Application`, `Categories=`).
3. Refresh the launcher/desktop database (`update-desktop-database ~/.local/share/applications` if available) so it appears immediately in the overview and dock.
