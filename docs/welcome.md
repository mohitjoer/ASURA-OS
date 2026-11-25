# Welcome, pin and reload cover

Three small Quickshell windows that run beside the shell. Each is an entry file at the
shell's root (`welcome.qml`, `pin.qml`, `reload-cover.qml`, code in the directory of the
same name) started by a wrapper in `session/`, the same shape as `hub.qml` +
`toren-hub`: a configuration's root is the directory of its entry file, and the `qs.*`
imports resolve only from the shell's root.

| Window | Started by | Compositor sees |
|---|---|---|
| Welcome tour | `toren-welcome [--if-first-run]` | window titled `Welcome to Toren` (float + centre rule in `fragments/W14.conf`) |
| Pinned shots | `toren-pin` (from `toren-shot`) | layer surface, namespace `toren-pin`, top layer |
| Reload cover | `toren-reload-cover begin\|finish\|fail` | layer surface, namespace `toren-reload-cover`, overlay layer |

**Welcome.** The compositor runs `toren-welcome --if-first-run` at every login; it exits
at once when `~/.local/state/asura/welcome-done` exists, and writes that file when the
tour closes cleanly. Delete the file to see the tour again, or run `toren-welcome`.
The interface-scale and bar-style choices go to the daemon (`toren-shell call
settings.patch`, which owns `shell.json`); the desktop-widget toggles write
`widgets.json`, which the desktop and the Hub's Widgets page also write.
The key list in step 2 is the default bindings, written into `StepBasics.qml`.

**Pin.** `toren-shot` copies a PNG to `$XDG_RUNTIME_DIR/asura/pins/pin-*.png`, touches
`.poke` there and runs `toren-pin`, which holds the single-instance lock. Drag to move,
scroll to resize, middle-click or Esc to close; the host exits with the last pin.

**Reload cover.** `toren-reload-cover begin` raises the curtain and returns once every
output is covered; the shell restarts; the new shell reads the token from
`$XDG_RUNTIME_DIR/toren-reload-cover.json` and runs `finish TOKEN` when it has drawn.
With no `finish` the cover shows "reload failed" after 16.5 s and opens. Whoever restarts
the shell only has to run `begin` first. `shell/reload-cover/reload_cover_test.sh` runs
the handshake against a stand-in for Quickshell.
