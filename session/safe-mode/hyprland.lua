-- Toren Safe Mode Hyprland Configuration
-- Minimal fallback desktop for system recovery and repair.
-- No shell, no user configs, no external plugins.

hl.monitor({
    output = "",
    mode = "preferred",
    position = "auto",
    scale = 1,
})

local mainMod = "SUPER"

hl.bind(mainMod .. " + Return", hl.dsp.exec_cmd("ghostty"))
hl.bind(mainMod .. " + Q", hl.dsp.window.kill())
hl.bind(mainMod .. " + C", hl.dsp.window.kill())
hl.bind(mainMod .. " + M", hl.dsp.exit())

hl.config({
    general = {
        gaps_in = 4,
        gaps_out = 8,
        border_size = 2,
        col = {
            active_border = "rgba(33ccffee)",
            inactive_border = "rgba(595959aa)",
        },
    },
    decoration = {
        rounding = 4,
        blur = {
            enabled = false,
        },
        shadow = {
            enabled = false,
        },
    },
    animations = {
        enabled = false,
    },
    misc = {
        disable_hyprland_logo = true,
        disable_splash_rendering = true,
        disable_hyprland_guiutils_check = true,
        background_color = 0x11111b,
    },
})

hl.on("hyprland.start", function()
    hl.exec_cmd("ghostty")
end)
