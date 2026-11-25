-- Asura's compositor configuration for Hyprland 0.57 (Lua).
--
-- Installed at /usr/share/asura/hyprland/hyprland.lua and loaded by the
-- session with `Hyprland -c`. Relative sources resolve against this file's
-- directory. The order is: environment, monitors, input, look, binds,
-- rules, autostart, then the files the user and the Hub own, so the last word
-- is always the user's.

local configDir = debug.getinfo(1, "S").source:match("@?(.*)/") or "/usr/share/asura/hyprland"
local function load_part(name)
    local p = configDir .. "/" .. name .. ".lua"
    local f = io.open(p, "r")
    if f then
        f:close()
        dofile(p)
    else
        pcall(require, name)
    end
end

load_part("env")
load_part("monitors")
load_part("input")
load_part("misc")
load_part("decoration")
load_part("animations")
load_part("binds")
load_part("resize")
load_part("lid")
load_part("window-rules")
load_part("layer-rules")
load_part("autostart")

-- The session's anchor. The compositor is up by the time this runs, which is
-- the first moment the display exists, and asura-activate puts the
-- compositor's environment into the user manager before starting the target that
-- pulls in the portal dispatcher and the shell.
hl.on("hyprland.start", function()
    hl.exec_cmd("/usr/share/asura/session/asura-activate")
end)

-- State that belongs to the user and to the tools, in this order; the last
-- file wins. They live in ~/.config/asura/hyprland/.
local home = os.getenv("HOME") or ""
local userDir = home .. "/.config/asura/hyprland/"
local function safe_source(file)
    local p = userDir .. file
    local f = io.open(p, "r")
    if f then
        f:close()
        pcall(dofile, p)
    end
end

safe_source("shell.lua")
safe_source("monitors-auto.lua")
safe_source("monitors-user.lua")
safe_source("settings.lua")
safe_source("rebinds.lua")
safe_source("asura.lua")
safe_source("user.lua")
safe_source("vm-perf.lua")
