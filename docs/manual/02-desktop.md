# The desktop

- **The bar** (top): workspaces on the left, media and the clock in the
  middle, status icons, CPU, memory and battery on the right. Click the status
  icons for the control centre.
- **The dock** (bottom): pinned and running apps. It hides while you work in
  a window; move the pointer to the bottom edge to bring it back.
- **The control centre** (**Super + N**): quick toggles (Wi-Fi, Bluetooth, do
  not disturb...), volume and brightness, notifications, calendar.
- **The launcher** (**Super + Space**): apps, commands and calculations.
- **The overview** (**Super + Tab**): every workspace and window at once.

## Shortcuts

| Keys | Does |
|---|---|
| Super + Space | launcher |
| Super + Tab | overview |
| Alt + Tab | switch windows |
| Super + N | control centre |
| Super + K | all shortcuts |
| Super + V | clipboard history |
| Super + Enter, Super + T | terminal |
| Super + E | files |
| Super + B | browser |
| Super + Q | close the window |
| Super + F / Super + D | fullscreen / maximise |
| Super + A | float the window, centred |
| Super + arrows | move focus |
| Super + Shift + arrows | swap windows |
| Super + Ctrl + arrows | resize |
| Super + R | resize mode (arrows, Esc to leave) |
| Super + 1 … 0 | go to workspace 1 … 10 |
| Super + Alt + 1 … 0 | move the window to a workspace and follow it |
| Super + Shift + 1 … 0 | move the window to a workspace, stay here |
| Super + mouse wheel | previous / next workspace |
| Super + drag | move (left button) or resize (right button) a window |
| Print, Super + Shift + S | screenshot of an area |
| Shift + Print | screenshot of the monitor |
| Super + Shift + T | copy text from the screen (OCR) |
| Super + Shift + C | colour picker |
| Super + W / Super + Shift + W | choose / random wallpaper |
| Super + L | lock |
| Ctrl + Alt + Delete | log out, restart, shut down |
| Ctrl + Shift + Escape | system monitor (btop) |
| Super + Shift + A | restart audio |
| Super + comma | Toren settings (the Hub) |

## Layouts per workspace

Each workspace can arrange its windows its own way, and changing one does not
change the others:

| Layout | What it does |
|---|---|
| Tiling | each new window splits the space (the default) |
| Scrolling | windows sit side by side in a row wider than the screen; scroll through them |
| Master | one large window, the others stacked beside it |
| Monocle | one window fills the screen at a time |
| Floating | windows float freely and overlap |

Pick one in the power menu (Ctrl+Alt+Delete, "Layout of workspace N"), or click
the layout mark in the bar to step to the next one. From a terminal:

```sh
toren layout                 # the current workspace's layout
toren layout set scroll      # tile, scroll, master, monocle, float, default
toren layout next            # step to the next one
```

A workspace set to Default follows the tiling layout chosen in Settings ->
Window Manager -> Layout. Choices are kept across logins
(`~/.config/toren/layouts.json`).
