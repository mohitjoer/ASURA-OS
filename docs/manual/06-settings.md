# Settings

- **The desktop's Settings app** (from the control centre, **Super + N**)
  changes the bar, the dock, the control centre and appearance.
- **Toren's settings** (**Super + comma**, the Hub) covers the system side:
  keybinds, power, graphics, wallpaper and more.
- **Your desktop settings file** is `~/.config/illogical-impulse/config.json`;
  the Settings app writes it, and the desktop reloads it when it changes.
- **Hyprland** (gaps, borders, window rules, monitors, keyboard layout): put
  your changes in `~/.config/asura/hyprland/user.conf`; it is read after
  Toren's, so yours win. `hyprctl reload` applies them.
- **From a terminal**: `toren config get|set SCHEMA KEY [VALUE]`.

Never edit files under `/usr/share/asura`: every update replaces them. Your
files in your home folder survive every update.
