-- Every key the keyboard has, in one place. session/binds.json is the same
-- list written for a person: `asura doctor` compares the two in both
-- directions, so a bind added here needs a line there and the other way round.
-- Shell surfaces are called as `asura ipc call asura <function> [argument]`;
-- `asura ipc` finds the running shell, so the key works whichever tree runs.
--
-- Flags: l = works over the lock screen, e = repeats while held, r = fires on
-- key RELEASE, m = mouse. `bindl` is the lock-screen flag and fires on the
-- press: a bind/bindl pair opens and closes a switcher in the same instant.
-- Anything that must act on release is `bindr`.

hl.config({
    binds = {
    -- pressing the key for the workspace you are already on returns to the
    -- previous one, so a single key toggles between two workspaces
    workspace_back_and_forth = true
    },
})

-- --- Desktop ---------------------------------------------------------------
hl.bind("SUPER + Space", hl.dsp.exec_cmd("asura ipc call asura togglePopover launcher"))
hl.bind("SUPER + ALT + Space", hl.dsp.exec_cmd("asura ipc call asura togglePopover menu"))
hl.bind("SUPER + Escape", hl.dsp.exec_cmd("asura ipc call asura togglePopover quick"))
hl.bind("SUPER + S", hl.dsp.exec_cmd("asura ipc call asura toggleSurface stash"))
hl.bind("SUPER + V", hl.dsp.exec_cmd("asura ipc call asura toggleSurface clipboard"))
hl.bind("SUPER + K", hl.dsp.exec_cmd("asura ipc call asura toggleCheatsheet"))
hl.bind("SUPER + L", hl.dsp.exec_cmd("asura-shell lock"))
-- Ctrl+Alt+Delete is the session screen — the full-screen power menu (lock,
-- logout, reboot, shutdown) the shell draws over the focused screen. The
-- chord is upstream's own for this surface, and it is also the one chord
-- every desktop shares for ending the session.
hl.bind("CTRL + ALT + Delete", hl.dsp.exec_cmd("asura ipc call asura session"))
hl.bind("SUPER + comma", hl.dsp.exec_cmd("asura-shell hub open"))
hl.bind("SUPER + grave", hl.dsp.exec_cmd("asura-shell voice"))
hl.bind("SUPER + W", hl.dsp.exec_cmd("asura-wallpaper wallpaper ui"))
hl.bind("SUPER + SHIFT + W", hl.dsp.exec_cmd("asura-wallpaper wallpaper random"))
hl.bind("SUPER + M", hl.dsp.exec_cmd("asura ipc call asura toggleSurface visualizer"))
hl.bind("SUPER + SHIFT + M", hl.dsp.exec_cmd("asura ipc call asura toggleSurface visualizer-overlay"))
hl.bind("SUPER + ALT + M", hl.dsp.exec_cmd("asura ipc call asura toggleSurface visualizer-place"))

-- Super+Tab is the workspace overview (an expo of live window previews); with
-- Alt it steps to the next desktop, inside the overview or opening it there.
hl.bind("SUPER + Tab", hl.dsp.exec_cmd("asura ipc call asura toggleOverview"))
hl.bind("SUPER + ALT + Tab", hl.dsp.exec_cmd("asura ipc call asura overviewDesktop 1"))

-- Alt+Tab is the most-recently-used window switcher: a press moves to the next
-- window and focuses it, Shift walks the list the other way, and the card stays
-- up 450 ms after the last press or release. The release is `bindr`. A release
-- bind needs its modifier still held, hence a pair for Shift; when Alt comes up
-- first the card ends itself after the dwell.
hl.bind("ALT + Tab", hl.dsp.exec_cmd("asura ipc call asura advanceSwitcher"))
hl.bind("ALT + SHIFT + Tab", hl.dsp.exec_cmd("asura ipc call asura reverseSwitcher"))
hl.bind("ALT + Tab", hl.dsp.exec_cmd("asura ipc call asura closeSwitcher"), { release = true })
hl.bind("ALT + SHIFT + Tab", hl.dsp.exec_cmd("asura ipc call asura closeSwitcher"), { release = true })

-- --- Apps ------------------------------------------------------------------
-- Two chords for the terminal: Return and T.
hl.bind("SUPER + T", hl.dsp.exec_cmd("asura-app terminal"))
hl.bind("SUPER + Return", hl.dsp.exec_cmd("asura-app terminal"))
hl.bind("SUPER + E", hl.dsp.exec_cmd("asura-app files"))
hl.bind("SUPER + B", hl.dsp.exec_cmd("asura-app browser"))
hl.bind("SUPER + N", hl.dsp.exec_cmd("asura ipc call asura togglePopover sidebar"))
hl.bind("SUPER + O", hl.dsp.exec_cmd("asura-app notes"))
hl.bind("CTRL + SHIFT + Escape", hl.dsp.exec_cmd("asura-app terminal -- btop"))
hl.bind("SUPER + SHIFT + Escape", hl.dsp.exec_cmd("asura-app terminal -- btop"))

-- --- Windows ---------------------------------------------------------------
hl.bind("SUPER + Q", hl.dsp.window.kill())
hl.bind("SUPER + F", hl.dsp.window.fullscreen(0))
hl.bind("SUPER + D", hl.dsp.window.fullscreen(1))
hl.bind("SUPER + C", hl.dsp.window.center())
hl.bind("SUPER + SHIFT + P", hl.dsp.window.pin())
hl.bind("SUPER + G", hl.dsp.group.toggle())
-- Float at 1000x660, centred; press again to tile back.
hl.bind("SUPER + A", hl.dsp.exec_cmd("hyprctl --batch \"dispatch togglefloating ; dispatch resizeactive exact 1000 660 ; dispatch centerwindow\""))
-- Split the current window's tile the other way (dwindle), and pseudo-tile a
-- floating window so it keeps its aspect ratio inside the tiled grid.
hl.bind("SUPER + J", hl.dsp.layout("togglesplit"))
hl.bind("SUPER + CTRL + P", hl.dsp.window.pseudo())
-- Resize mode (resize.conf). Entering a submap is silent, hence the toast.
hl.bind("SUPER + R", hl.dsp.exec_cmd("hyprctl dispatch submap resize && hyprctl notify -1 2200 0 \"Resize mode: arrows or hjkl resize. Esc exits\""))

hl.bind("SUPER + Left", hl.dsp.focus({ direction = "left" }))
hl.bind("SUPER + Right", hl.dsp.focus({ direction = "right" }))
hl.bind("SUPER + Up", hl.dsp.focus({ direction = "up" }))
hl.bind("SUPER + Down", hl.dsp.focus({ direction = "down" }))
hl.bind("SUPER + SHIFT + Left", hl.dsp.window.swap({ direction = "left" }))
hl.bind("SUPER + SHIFT + Right", hl.dsp.window.swap({ direction = "right" }))
hl.bind("SUPER + SHIFT + Up", hl.dsp.window.swap({ direction = "up" }))
hl.bind("SUPER + SHIFT + Down", hl.dsp.window.swap({ direction = "down" }))
hl.bind("SUPER + CTRL + Left", hl.dsp.window.resize({ x = -40, y = 0, relative = true }), { repeating = true })
hl.bind("SUPER + CTRL + Right", hl.dsp.window.resize({ x = 40, y = 0, relative = true }), { repeating = true })
hl.bind("SUPER + CTRL + Up", hl.dsp.window.resize({ x = 0, y = -40, relative = true }), { repeating = true })
hl.bind("SUPER + CTRL + Down", hl.dsp.window.resize({ x = 0, y = 40, relative = true }), { repeating = true })
hl.bind("SUPER + bracketleft", hl.dsp.window.move({ direction = "left" }))
hl.bind("SUPER + bracketright", hl.dsp.window.move({ direction = "right" }))

hl.bind("SUPER + mouse:272", hl.dsp.window.drag(), { mouse = true })
hl.bind("SUPER + mouse:273", hl.dsp.window.resize(), { mouse = true })

-- --- Monitors (the l/r/u/d selector picks the monitor in that direction) ----
hl.bind("SUPER + ALT + Left", hl.dsp.focus({ monitor = "l" }))
hl.bind("SUPER + ALT + Right", hl.dsp.focus({ monitor = "r" }))
hl.bind("SUPER + ALT + Up", hl.dsp.focus({ monitor = "u" }))
hl.bind("SUPER + ALT + Down", hl.dsp.focus({ monitor = "d" }))
hl.bind("SUPER + ALT + SHIFT + Left", hl.dsp.window.move({ monitor = "l" }))
hl.bind("SUPER + ALT + SHIFT + Right", hl.dsp.window.move({ monitor = "r" }))
hl.bind("SUPER + ALT + SHIFT + Up", hl.dsp.window.move({ monitor = "u" }))
hl.bind("SUPER + ALT + SHIFT + Down", hl.dsp.window.move({ monitor = "d" }))
hl.bind("SUPER + CTRL + ALT + Left", hl.dsp.workspace.move({ monitor = "l" }))
hl.bind("SUPER + CTRL + ALT + Right", hl.dsp.workspace.move({ monitor = "r" }))
hl.bind("SUPER + CTRL + ALT + Up", hl.dsp.workspace.move({ monitor = "u" }))
hl.bind("SUPER + CTRL + ALT + Down", hl.dsp.workspace.move({ monitor = "d" }))
hl.bind("SUPER + P", hl.dsp.exec_cmd("asura-monitor toggle"))

-- --- Workspaces ------------------------------------------------------------
-- Super+N is the Nth workspace of the CURRENT DESKTOP (a desktop is a block of
-- 10 workspace ids; see asura-workspace), so on desktop 2 Super+3 is ws13.
-- Super+Alt+N sends the window there and follows, Super+Shift+N sends it
-- without following. The number pad does the same, bound twice: on KP_<n> (the
-- keysym with NumLock on) and on the navigation keysym the same key sends with
-- NumLock off.
hl.bind("SUPER + H", hl.dsp.exec_cmd("asura-workspace hide"))
hl.bind("SUPER + ALT + H", hl.dsp.exec_cmd("asura-workspace scratch"))
hl.bind("SUPER + mouse_up", hl.dsp.focus({ workspace = "r-1" }))
hl.bind("SUPER + mouse_down", hl.dsp.focus({ workspace = "r+1" }))
hl.bind("SUPER + Prior", hl.dsp.focus({ workspace = "r-1" }))
hl.bind("SUPER + Next", hl.dsp.focus({ workspace = "r+1" }))
hl.bind("SUPER + SHIFT + Prior", hl.dsp.window.move({ workspace = "r-1" }))
hl.bind("SUPER + SHIFT + Next", hl.dsp.window.move({ workspace = "r+1" }))

hl.bind("SUPER + 1", hl.dsp.exec_cmd("asura-workspace focus 1"))
hl.bind("SUPER + 2", hl.dsp.exec_cmd("asura-workspace focus 2"))
hl.bind("SUPER + 3", hl.dsp.exec_cmd("asura-workspace focus 3"))
hl.bind("SUPER + 4", hl.dsp.exec_cmd("asura-workspace focus 4"))
hl.bind("SUPER + 5", hl.dsp.exec_cmd("asura-workspace focus 5"))
hl.bind("SUPER + 6", hl.dsp.exec_cmd("asura-workspace focus 6"))
hl.bind("SUPER + 7", hl.dsp.exec_cmd("asura-workspace focus 7"))
hl.bind("SUPER + 8", hl.dsp.exec_cmd("asura-workspace focus 8"))
hl.bind("SUPER + 9", hl.dsp.exec_cmd("asura-workspace focus 9"))
hl.bind("SUPER + 0", hl.dsp.exec_cmd("asura-workspace focus 10"))
hl.bind("SUPER + KP_1", hl.dsp.exec_cmd("asura-workspace focus 1"))
hl.bind("SUPER + KP_2", hl.dsp.exec_cmd("asura-workspace focus 2"))
hl.bind("SUPER + KP_3", hl.dsp.exec_cmd("asura-workspace focus 3"))
hl.bind("SUPER + KP_4", hl.dsp.exec_cmd("asura-workspace focus 4"))
hl.bind("SUPER + KP_5", hl.dsp.exec_cmd("asura-workspace focus 5"))
hl.bind("SUPER + KP_6", hl.dsp.exec_cmd("asura-workspace focus 6"))
hl.bind("SUPER + KP_7", hl.dsp.exec_cmd("asura-workspace focus 7"))
hl.bind("SUPER + KP_8", hl.dsp.exec_cmd("asura-workspace focus 8"))
hl.bind("SUPER + KP_9", hl.dsp.exec_cmd("asura-workspace focus 9"))
hl.bind("SUPER + KP_0", hl.dsp.exec_cmd("asura-workspace focus 10"))
hl.bind("SUPER + KP_End", hl.dsp.exec_cmd("asura-workspace focus 1"))
hl.bind("SUPER + KP_Down", hl.dsp.exec_cmd("asura-workspace focus 2"))
hl.bind("SUPER + KP_Next", hl.dsp.exec_cmd("asura-workspace focus 3"))
hl.bind("SUPER + KP_Left", hl.dsp.exec_cmd("asura-workspace focus 4"))
hl.bind("SUPER + KP_Begin", hl.dsp.exec_cmd("asura-workspace focus 5"))
hl.bind("SUPER + KP_Right", hl.dsp.exec_cmd("asura-workspace focus 6"))
hl.bind("SUPER + KP_Home", hl.dsp.exec_cmd("asura-workspace focus 7"))
hl.bind("SUPER + KP_Up", hl.dsp.exec_cmd("asura-workspace focus 8"))
hl.bind("SUPER + KP_Prior", hl.dsp.exec_cmd("asura-workspace focus 9"))
hl.bind("SUPER + KP_Insert", hl.dsp.exec_cmd("asura-workspace focus 10"))

hl.bind("SUPER + ALT + 1", hl.dsp.exec_cmd("asura-workspace move 1"))
hl.bind("SUPER + ALT + 2", hl.dsp.exec_cmd("asura-workspace move 2"))
hl.bind("SUPER + ALT + 3", hl.dsp.exec_cmd("asura-workspace move 3"))
hl.bind("SUPER + ALT + 4", hl.dsp.exec_cmd("asura-workspace move 4"))
hl.bind("SUPER + ALT + 5", hl.dsp.exec_cmd("asura-workspace move 5"))
hl.bind("SUPER + ALT + 6", hl.dsp.exec_cmd("asura-workspace move 6"))
hl.bind("SUPER + ALT + 7", hl.dsp.exec_cmd("asura-workspace move 7"))
hl.bind("SUPER + ALT + 8", hl.dsp.exec_cmd("asura-workspace move 8"))
hl.bind("SUPER + ALT + 9", hl.dsp.exec_cmd("asura-workspace move 9"))
hl.bind("SUPER + ALT + 0", hl.dsp.exec_cmd("asura-workspace move 10"))
hl.bind("SUPER + ALT + KP_1", hl.dsp.exec_cmd("asura-workspace move 1"))
hl.bind("SUPER + ALT + KP_2", hl.dsp.exec_cmd("asura-workspace move 2"))
hl.bind("SUPER + ALT + KP_3", hl.dsp.exec_cmd("asura-workspace move 3"))
hl.bind("SUPER + ALT + KP_4", hl.dsp.exec_cmd("asura-workspace move 4"))
hl.bind("SUPER + ALT + KP_5", hl.dsp.exec_cmd("asura-workspace move 5"))
hl.bind("SUPER + ALT + KP_6", hl.dsp.exec_cmd("asura-workspace move 6"))
hl.bind("SUPER + ALT + KP_7", hl.dsp.exec_cmd("asura-workspace move 7"))
hl.bind("SUPER + ALT + KP_8", hl.dsp.exec_cmd("asura-workspace move 8"))
hl.bind("SUPER + ALT + KP_9", hl.dsp.exec_cmd("asura-workspace move 9"))
hl.bind("SUPER + ALT + KP_0", hl.dsp.exec_cmd("asura-workspace move 10"))
hl.bind("SUPER + ALT + KP_End", hl.dsp.exec_cmd("asura-workspace move 1"))
hl.bind("SUPER + ALT + KP_Down", hl.dsp.exec_cmd("asura-workspace move 2"))
hl.bind("SUPER + ALT + KP_Next", hl.dsp.exec_cmd("asura-workspace move 3"))
hl.bind("SUPER + ALT + KP_Left", hl.dsp.exec_cmd("asura-workspace move 4"))
hl.bind("SUPER + ALT + KP_Begin", hl.dsp.exec_cmd("asura-workspace move 5"))
hl.bind("SUPER + ALT + KP_Right", hl.dsp.exec_cmd("asura-workspace move 6"))
hl.bind("SUPER + ALT + KP_Home", hl.dsp.exec_cmd("asura-workspace move 7"))
hl.bind("SUPER + ALT + KP_Up", hl.dsp.exec_cmd("asura-workspace move 8"))
hl.bind("SUPER + ALT + KP_Prior", hl.dsp.exec_cmd("asura-workspace move 9"))
hl.bind("SUPER + ALT + KP_Insert", hl.dsp.exec_cmd("asura-workspace move 10"))

hl.bind("SUPER + SHIFT + 1", hl.dsp.exec_cmd("asura-workspace movesilent 1"))
hl.bind("SUPER + SHIFT + 2", hl.dsp.exec_cmd("asura-workspace movesilent 2"))
hl.bind("SUPER + SHIFT + 3", hl.dsp.exec_cmd("asura-workspace movesilent 3"))
hl.bind("SUPER + SHIFT + 4", hl.dsp.exec_cmd("asura-workspace movesilent 4"))
hl.bind("SUPER + SHIFT + 5", hl.dsp.exec_cmd("asura-workspace movesilent 5"))
hl.bind("SUPER + SHIFT + 6", hl.dsp.exec_cmd("asura-workspace movesilent 6"))
hl.bind("SUPER + SHIFT + 7", hl.dsp.exec_cmd("asura-workspace movesilent 7"))
hl.bind("SUPER + SHIFT + 8", hl.dsp.exec_cmd("asura-workspace movesilent 8"))
hl.bind("SUPER + SHIFT + 9", hl.dsp.exec_cmd("asura-workspace movesilent 9"))
hl.bind("SUPER + SHIFT + 0", hl.dsp.exec_cmd("asura-workspace movesilent 10"))
hl.bind("SUPER + SHIFT + KP_1", hl.dsp.exec_cmd("asura-workspace movesilent 1"))
hl.bind("SUPER + SHIFT + KP_2", hl.dsp.exec_cmd("asura-workspace movesilent 2"))
hl.bind("SUPER + SHIFT + KP_3", hl.dsp.exec_cmd("asura-workspace movesilent 3"))
hl.bind("SUPER + SHIFT + KP_4", hl.dsp.exec_cmd("asura-workspace movesilent 4"))
hl.bind("SUPER + SHIFT + KP_5", hl.dsp.exec_cmd("asura-workspace movesilent 5"))
hl.bind("SUPER + SHIFT + KP_6", hl.dsp.exec_cmd("asura-workspace movesilent 6"))
hl.bind("SUPER + SHIFT + KP_7", hl.dsp.exec_cmd("asura-workspace movesilent 7"))
hl.bind("SUPER + SHIFT + KP_8", hl.dsp.exec_cmd("asura-workspace movesilent 8"))
hl.bind("SUPER + SHIFT + KP_9", hl.dsp.exec_cmd("asura-workspace movesilent 9"))
hl.bind("SUPER + SHIFT + KP_0", hl.dsp.exec_cmd("asura-workspace movesilent 10"))
hl.bind("SUPER + SHIFT + KP_End", hl.dsp.exec_cmd("asura-workspace movesilent 1"))
hl.bind("SUPER + SHIFT + KP_Down", hl.dsp.exec_cmd("asura-workspace movesilent 2"))
hl.bind("SUPER + SHIFT + KP_Next", hl.dsp.exec_cmd("asura-workspace movesilent 3"))
hl.bind("SUPER + SHIFT + KP_Left", hl.dsp.exec_cmd("asura-workspace movesilent 4"))
hl.bind("SUPER + SHIFT + KP_Begin", hl.dsp.exec_cmd("asura-workspace movesilent 5"))
hl.bind("SUPER + SHIFT + KP_Right", hl.dsp.exec_cmd("asura-workspace movesilent 6"))
hl.bind("SUPER + SHIFT + KP_Home", hl.dsp.exec_cmd("asura-workspace movesilent 7"))
hl.bind("SUPER + SHIFT + KP_Up", hl.dsp.exec_cmd("asura-workspace movesilent 8"))
hl.bind("SUPER + SHIFT + KP_Prior", hl.dsp.exec_cmd("asura-workspace movesilent 9"))
hl.bind("SUPER + SHIFT + KP_Insert", hl.dsp.exec_cmd("asura-workspace movesilent 10"))

-- --- Tools -----------------------------------------------------------------
-- Print and Super+Shift+S open the screenshot editor on a region, or on the
-- window under the pointer; Shift+Print starts on a whole monitor. Both the
-- print key and its XF86 twin are bound: keyboards differ on which the Fn row
-- emits.
hl.bind("Print", hl.dsp.exec_cmd("asura-shot"))
hl.bind("SUPER + SHIFT + S", hl.dsp.exec_cmd("asura-shot"))
hl.bind("SHIFT + Print", hl.dsp.exec_cmd("env ASURA_SHOT_MODE=monitor asura-shot"))
hl.bind("SUPER + SHIFT + C", hl.dsp.exec_cmd("asura-cmd-color-picker"))
-- Super+Shift+T draws a rectangle or a circle and copies the text inside it
-- to the clipboard: the shell's region selector reads the region with
-- tesseract-ocr. The colour picker keeps its chord next door.
hl.bind("SUPER + SHIFT + T", hl.dsp.exec_cmd("asura ipc call asura ocr"))

-- --- Media and volume (locked: they work over the lock screen) ---------------
-- asura-volume honours the volume panel's boost toggle: off caps at 100%, on 150%.
hl.bind("XF86AudioRaiseVolume", hl.dsp.exec_cmd("asura-volume up"), { locked = true, repeating = true })
hl.bind("XF86AudioLowerVolume", hl.dsp.exec_cmd("asura-volume down"), { locked = true, repeating = true })
hl.bind("XF86AudioMute", hl.dsp.exec_cmd("wpctl set-mute @DEFAULT_AUDIO_SINK@ toggle"), { locked = true })
hl.bind("XF86AudioPlay", hl.dsp.exec_cmd("playerctl play-pause"), { locked = true })
hl.bind("XF86AudioPause", hl.dsp.exec_cmd("playerctl play-pause"), { locked = true })
hl.bind("XF86AudioNext", hl.dsp.exec_cmd("playerctl next"), { locked = true })
hl.bind("XF86AudioPrev", hl.dsp.exec_cmd("playerctl previous"), { locked = true })
hl.bind("SUPER + SHIFT + A", hl.dsp.exec_cmd("asura-restart-audio"))
hl.bind("SUPER + SHIFT + V", hl.dsp.exec_cmd("asura-summon asura-vm flock -n -o /tmp/asura-vm.lock asura-vm"))

-- --- Hardware (W18) --------------------------------------------------------
-- Brightness goes through asura-cmd-brightness -> `asura brightness`, which
-- asks systemd-logind to write the value, so no group membership is needed.
hl.bind("XF86MonBrightnessUp", hl.dsp.exec_cmd("asura-cmd-brightness +5"), { locked = true, repeating = true })
hl.bind("XF86MonBrightnessDown", hl.dsp.exec_cmd("asura-cmd-brightness -5"), { locked = true, repeating = true })
-- Keyboard backlight keys, for laptops whose driver reports them as key events
-- (the ASUS driver here steps the LED itself and never does).
hl.bind("XF86KbdBrightnessUp", hl.dsp.exec_cmd("asura brightness up --device kbd"), { locked = true, repeating = true })
hl.bind("XF86KbdBrightnessDown", hl.dsp.exec_cmd("asura brightness down --device kbd"), { locked = true, repeating = true })
-- The Fn touchpad key. Hyprland has no read-back for a device's state, so
-- asura-touchpad keeps the intent in the runtime directory.
hl.bind("XF86TouchpadToggle", hl.dsp.exec_cmd("asura-touchpad toggle"), { locked = true })
hl.bind("XF86TouchpadOn", hl.dsp.exec_cmd("asura-touchpad on"), { locked = true })
hl.bind("XF86TouchpadOff", hl.dsp.exec_cmd("asura-touchpad off"), { locked = true })

-- Thermal and fan profiles (Fn+F5 on ASUS TUF/ROG and laptops with fan switch keys)
hl.bind("XF86Launch4", hl.dsp.exec_cmd("fan-profile"), { locked = true })
