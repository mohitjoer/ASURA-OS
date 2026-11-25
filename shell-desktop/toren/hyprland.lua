-- The desktop's look, for Hyprland 0.57. An ASURA file, not vendored.

hl.config({
    general = {
        gaps_in = 4,
        gaps_out = 5,
        gaps_workspaces = 50,
        border_size = 2,
        col = {
            active_border = "rgba(0DB7D455)",
            inactive_border = "rgba(31313600)",
        },
    },
    decoration = {
        rounding = 0,
        active_opacity = 1.0,
        inactive_opacity = 1.0,
        dim_inactive = true,
        dim_strength = 0.05,
        dim_special = 0.2,
        blur = {
            enabled = true,
            xray = true,
            special = false,
            new_optimizations = true,
            size = 10,
            passes = 3,
            brightness = 1,
            noise = 0.05,
            contrast = 0.89,
            vibrancy = 0.5,
            vibrancy_darkness = 0.5,
            popups = false,
            popups_ignorealpha = 0.6,
            input_methods = true,
            input_methods_ignorealpha = 0.8,
        },
        shadow = {
            enabled = true,
            range = 20,
            offset = "0 2",
            render_power = 4,
            color = "rgba(00000020)",
        },
    },
    animations = {
        enabled = true,
    },
})

hl.curve("expressiveFastSpatial", { type = "bezier", points = { {0.42, 1.67}, {0.21, 0.90} } })
hl.curve("expressiveSlowSpatial", { type = "bezier", points = { {0.39, 1.29}, {0.35, 0.98} } })
hl.curve("expressiveDefaultSpatial", { type = "bezier", points = { {0.38, 1.21}, {0.22, 1.00} } })
hl.curve("emphasizedDecel", { type = "bezier", points = { {0.05, 0.7}, {0.1, 1} } })
hl.curve("emphasizedAccel", { type = "bezier", points = { {0.3, 0}, {0.8, 0.15} } })
hl.curve("standardDecel", { type = "bezier", points = { {0, 0}, {0, 1} } })
hl.curve("menu_decel", { type = "bezier", points = { {0.1, 1}, {0, 1} } })
hl.curve("menu_accel", { type = "bezier", points = { {0.52, 0.03}, {0.72, 0.08} } })
hl.curve("stall", { type = "bezier", points = { {1, -0.1}, {0.7, 0.85} } })

hl.animation({ leaf = "windowsIn", enabled = true, speed = 3, bezier = "emphasizedDecel", style = "popin 80%" })
hl.animation({ leaf = "fadeIn", enabled = true, speed = 3, bezier = "emphasizedDecel" })
hl.animation({ leaf = "windowsOut", enabled = true, speed = 2, bezier = "emphasizedDecel", style = "popin 90%" })
hl.animation({ leaf = "fadeOut", enabled = true, speed = 2, bezier = "emphasizedDecel" })
hl.animation({ leaf = "windowsMove", enabled = true, speed = 3, bezier = "emphasizedDecel", style = "slide" })
hl.animation({ leaf = "border", enabled = true, speed = 10, bezier = "emphasizedDecel" })
hl.animation({ leaf = "layersIn", enabled = true, speed = 2.7, bezier = "emphasizedDecel", style = "popin 93%" })
hl.animation({ leaf = "layersOut", enabled = true, speed = 2.4, bezier = "menu_accel", style = "popin 94%" })
hl.animation({ leaf = "fadeLayersIn", enabled = true, speed = 0.5, bezier = "menu_decel" })
hl.animation({ leaf = "fadeLayersOut", enabled = true, speed = 2.7, bezier = "stall" })
hl.animation({ leaf = "workspaces", enabled = true, speed = 7, bezier = "menu_decel", style = "slide" })
hl.animation({ leaf = "specialWorkspaceIn", enabled = true, speed = 2.8, bezier = "emphasizedDecel", style = "slidevert" })
hl.animation({ leaf = "specialWorkspaceOut", enabled = true, speed = 1.2, bezier = "emphasizedAccel", style = "slidevert" })
hl.animation({ leaf = "zoomFactor", enabled = true, speed = 3, bezier = "standardDecel" })

hl.layer_rule({ match = { namespace = ".*" }, xray = 1 })
hl.layer_rule({ match = { namespace = "quickshell:.*" }, blur_popups = true, blur = true, ignore_alpha = 0.79 })
hl.layer_rule({ match = { namespace = "quickshell:background" }, xray = 0, blur = false })
hl.layer_rule({ match = { namespace = "quickshell:bar" }, animation = "slide" })
hl.layer_rule({ match = { namespace = "quickshell:actionCenter" }, no_anim = true })
hl.layer_rule({ match = { namespace = "quickshell:cheatsheet" }, animation = "slide bottom" })
hl.layer_rule({ match = { namespace = "quickshell:dock" }, animation = "slide bottom" })
hl.layer_rule({ match = { namespace = "quickshell:screenCorners" }, animation = "popin 120%" })
hl.layer_rule({ match = { namespace = "quickshell:lockWindowPusher" }, no_anim = true })
hl.layer_rule({ match = { namespace = "quickshell:notificationPopup" }, animation = "fade" })
hl.layer_rule({ match = { namespace = "quickshell:overlay" }, no_anim = true, ignore_alpha = 1 })
hl.layer_rule({ match = { namespace = "quickshell:overview" }, no_anim = true })
hl.layer_rule({ match = { namespace = "quickshell:osk" }, animation = "slide bottom" })
hl.layer_rule({ match = { namespace = "quickshell:polkit" }, no_anim = true })
hl.layer_rule({ match = { namespace = "quickshell:popup" }, xray = 0, ignore_alpha = 1 })
hl.layer_rule({ match = { namespace = "quickshell:mediaControls" }, ignore_alpha = 1 })
hl.layer_rule({ match = { namespace = "quickshell:reloadPopup" }, animation = "slide" })
hl.layer_rule({ match = { namespace = "quickshell:regionSelector" }, no_anim = true })
hl.layer_rule({ match = { namespace = "quickshell:sidebarLeft" }, animation = "slide left" })
hl.layer_rule({ match = { namespace = "quickshell:sidebarRight" }, animation = "slide right" })
