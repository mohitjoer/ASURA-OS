# Settings

- **Settings** (**Super + comma**, or the gear in the control centre) is the
  one place for everything: the desktop (bar, dock, look, clock, lock
  screen, screenshots, search, battery and notifications, on the Desktop
  page) and the system (displays, connections, input, keybinds, graphics and
  power, windows, updates).
- **Your desktop settings file** is `~/.config/illogical-impulse/config.json`;
  Settings writes it, and the desktop reloads it when it changes.
- **Hyprland** (gaps, borders, window rules, monitors, keyboard layout): put
  your changes in `~/.config/toren/hyprland/user.conf`; it is read after
  Toren's, so yours win. `hyprctl reload` applies them.
- **From a terminal**: `toren config get|set SCHEMA KEY [VALUE]`.

Never edit files under `/usr/share/toren`: every update replaces them. Your
files in your home folder survive every update.
