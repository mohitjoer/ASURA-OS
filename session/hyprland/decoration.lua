hl.config({
    general = {
        gaps_in = 6,
        gaps_out = 12,
        border_size = 1,
        layout = "dwindle",
        resize_on_border = true,
        allow_tearing = true,
        col = {
            active_border = { colors = { "rgba(f2f2f2ee)", "rgba(8a8a8a66)" }, angle = 45 },
            inactive_border = "rgba(37373788)",
        },
    },
    dwindle = {
        preserve_split = true,
        force_split = 2,
        smart_split = false,
        smart_resizing = true,
    },
    master = {
        new_status = "master",
        new_on_top = true,
        mfact = 0.55,
        orientation = "left",
    },
    decoration = {
        rounding = 12,
        rounding_power = 2,
        active_opacity = 0.98,
        inactive_opacity = 0.92,
        shadow = {
            enabled = true,
            range = 28,
            render_power = 3,
            color = "rgba(00000077)",
            color_inactive = "rgba(00000044)",
        },
        blur = {
            enabled = true,
            size = 8,
            passes = 3,
            vibrancy = 0.22,
            noise = 0.015,
            new_optimizations = true,
            ignore_opacity = true,
        },
    },
})

hl.exec_cmd("toren-perf-apply")
hl.exec_cmd([[n=$(grep -oE '"screenShader" *: *"[a-z]+"' "$HOME/.config/toren/shell.json" 2>/dev/null | grep -oE '[a-z]+"$' | tr -d '"'); [ -n "$n" ] && hyprctl keyword decoration:screen_shader "/usr/share/toren/hyprland/shaders/$n.glsl" >/dev/null]])
