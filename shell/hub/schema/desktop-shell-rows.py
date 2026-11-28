#!/usr/bin/env python3
"""Regenerate shell/hub/schema/DesktopShellPage.js from the desktop's Config.qml.

Usage: python3 shell/hub/schema/desktop-shell-rows.py > shell/hub/schema/DesktopShellPage.js
"""
import json
import os
import re
import sys

ROOT = os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "..", "..")
CONFIG = os.path.join(ROOT, "shell-desktop", "modules", "common", "Config.qml")


def read_keys(path):
    """Every leaf option in Config.qml: dotted path -> type, default, note."""
    src = open(path).read().splitlines()
    stack=[]; out={}
    for line in src:
        s=line.strip()
        m=re.match(r'property JsonObject (\w+): JsonObject \{',s)
        ind=len(line)-len(line.lstrip())
        while stack and stack[-1][1]>=ind: stack.pop()
        if m:
            stack.append((m.group(1),ind)); continue
        m=re.match(r'property (bool|int|real|string|list<string>|var) (\w+): ("(?:[^"\\]|\\.)*"|[^/]+?)\s*(//\s*(.*))?$',s)
        if m and stack:
            dotted=".".join(n for n,_ in stack if n!="options")+"."+m.group(2)
            out[dotted]={"type":m.group(1),"def":m.group(3).strip(),"note":(m.group(5) or "").strip()}
    return out


keys = read_keys(CONFIG)
R = []
def row(tab, group, key, ctl, label, desc, **kw):
    if "store" in kw:
        # kept outside the desktop's config.json; the page reads and writes it
        r = {"tab": tab, "group": group, "key": key, "ctl": ctl, "label": label, "desc": desc}
        r.update(kw); R.append(r); return
    assert key in keys, key
    d = keys[key]["def"]
    t = keys[key]["type"]
    if t == "bool": dv = d == "true"
    elif t in ("int", "real"): dv = float(d) if "." in d else int(d)
    else:
        try:
            dv = json.loads(d)
        except ValueError:
            dv = ""  # an expression (e.g. a directory), not a literal
    r = {"tab": tab, "group": group, "key": key, "ctl": ctl, "label": label, "desc": desc, "def": dv}
    r.update(kw); R.append(r)

row("Bar & dock", "POSITION", "bar.bottom", "sw", "Bar at the bottom", "Moves the bar from the top edge to the bottom edge")
row("Bar & dock", "POSITION", "bar.autoHide.enable", "sw", "Hide the bar", "The bar slides away and comes back when the pointer reaches its edge")
row("Bar & dock", "SHAPE", "bar.cornerStyle", "seg", "Bar silhouette & corners", "Hug: the bar meets the screen corners. Float: a raised strip. Rect: a flat rectangle",
    opts=["Hug", "Float", "Rect"], map={"Hug": 0, "Float": 1, "Rect": 2})
row("Bar & dock", "SHAPE", "bar.floatStyleShadow", "sw", "Shadow", "A shadow under the bar when its shape is Float")
row("Bar & dock", "WEATHER", "bar.weather.enable", "sw", "Weather in the bar", "Shows the temperature and conditions next to the clock")
row("Bar & dock", "WEATHER", "bar.weather.enableGPS", "sw", "Find my location", "Uses the network location; turn off to name a city below")
row("Bar & dock", "WEATHER", "bar.weather.city", "text", "City", "Used when finding the location is off", eg="Berlin")
row("Bar & dock", "WEATHER", "bar.weather.useUSCS", "sw", "Fahrenheit", "US units instead of Celsius")

row("Bar & dock", "DOCK", "dock.enable", "sw", "Dock", "The app dock at the bottom of the screen")
row("Bar & dock", "DOCK", "dock.pinnedOnStartup", "sw", "Always show", "Keeps the dock on screen instead of hiding it while a window is focused")
row("Bar & dock", "DOCK", "dock.monochromeIcons", "sw", "Monochrome icons", "Tints the dock's icons to the colour scheme")
row("Bar & dock", "DOCK", "appearance.transparency.enable", "sw", "Transparency", "Lets the wallpaper show through panels and menus")

row("Look", "COLOURS", "appearance.palette.type", "seg", "Colour scheme", "How colours are taken from the wallpaper",
    opts=["Auto", "Tonal", "Neutral", "Vibrant", "Expressive", "Monochrome"],
    map={"Auto": "auto", "Tonal": "scheme-tonal-spot", "Neutral": "scheme-neutral", "Vibrant": "scheme-rainbow",
         "Expressive": "scheme-expressive", "Monochrome": "scheme-monochrome"})
row("Look", "SCREEN", "appearance.fakeScreenRounding", "seg", "Rounded screen corners", "Draws rounded corners over the screen's edges",
    opts=["Off", "Always", "Not fullscreen"], map={"Off": 0, "Always": 1, "Not fullscreen": 2})
row("Look", "APPS", "toren.gtkRoundedCorners", "sw", "Rounded app corners",
    "GTK and libadwaita apps (Files, settings dialogs) draw rounded window corners. Off keeps them square. Applies to apps opened after the change",
    store="gtk-corners", **{"def": False})
row("Look", "OVERVIEW", "overview.rows", "step", "Rows", "Workspaces per column in the overview", **{"from": 1, "to": 5})
row("Look", "OVERVIEW", "overview.columns", "step", "Columns", "Workspaces per row in the overview", **{"from": 1, "to": 10})

row("Clock", "TIME", "time.format", "seg", "Time format", "How the bar and lock screen show the time",
    opts=["24-hour", "12-hour"], map={"24-hour": "hh:mm", "12-hour": "h:mm ap"})
row("Clock", "TIME", "time.secondPrecision", "sw", "Seconds", "Shows seconds in the clock")

row("System", "NOTIFICATIONS", "notifications.timeout", "step", "Notification time", "How long a notification stays on screen", unit="ms", **{"from": 1000, "to": 30000})
row("System", "NOTIFICATIONS", "osd.timeout", "step", "Volume and brightness popup", "How long the volume or brightness popup stays", unit="ms", **{"from": 500, "to": 5000})

row("Lock screen", "LOCK SCREEN", "lock.centerClock", "sw", "Clock in the centre", "Puts the lock screen clock in the middle of the screen")
row("Lock screen", "LOCK SCREEN", "lock.blur.enable", "sw", "Blur the wallpaper", "Blurs the wallpaper behind the lock screen")
row("Lock screen", "LOCK SCREEN", "lock.showLockedText", "sw", "\"Locked\" label", "Shows a Locked label above the password field")
row("Lock screen", "SECURITY", "lock.security.requirePasswordToPower", "sw", "Password to shut down", "Asks for the password before shutting down or restarting from the lock screen")

row("System", "WARNINGS", "battery.low", "step", "Low battery warning", "Warns when the battery falls to this level", unit="%", **{"from": 5, "to": 50})
row("System", "WARNINGS", "battery.critical", "step", "Critical battery warning", "Warns again at this level", unit="%", **{"from": 1, "to": 20})
row("System", "WARNINGS", "sounds.battery", "sw", "Warning sound", "Plays a sound with the battery warnings")
row("System", "SUSPEND", "battery.automaticSuspend", "sw", "Suspend when empty", "Suspends the computer when the battery is nearly empty")

row("Look", "NIGHT LIGHT", "light.night.automatic", "sw", "On a schedule", "Warms the screen colours between the times below")
row("Look", "NIGHT LIGHT", "light.night.from", "text", "From", "24-hour time", eg="19:00")
row("Look", "NIGHT LIGHT", "light.night.to", "text", "To", "24-hour time", eg="06:30")
row("Look", "NIGHT LIGHT", "light.night.colorTemperature", "step", "Warmth", "Lower is warmer", unit="K", **{"from": 2500, "to": 6500})

row("System", "VOLUME", "audio.protection.enable", "sw", "Limit volume jumps", "Stops the volume from jumping up suddenly")

# Moved from the desktop's own Settings app (shell-desktop/settings.qml), so
# the Hub is the only settings panel.
row("Bar & dock", "SHAPE", "bar.borderless", "sw", "Plain items", "Draws the bar's items without the boxes that group them")
row("Bar & dock", "WORKSPACES", "bar.workspaces.showAppIcons", "sw", "App icons", "Shows each workspace's main app instead of a dot")
row("Bar & dock", "INDICATORS", "bar.indicators.notifications.showUnreadCount", "sw", "Unread count", "Shows how many notifications are waiting")
row("Bar & dock", "INDICATORS", "bar.tooltips.clickToShow", "sw", "Tooltips on click", "Shows the bar's tooltips on click instead of on hover")
row("Bar & dock", "TRAY", "tray.monochromeIcons", "sw", "Monochrome tray icons", "Tints the tray's app icons to the colour scheme")
row("Bar & dock", "WEATHER", "bar.weather.fetchInterval", "step", "Weather refresh", "How often the weather is fetched", unit="min", **{"from": 5, "to": 120})

row("Look", "FONTS", "appearance.fonts.main", "text", "Interface font", "The font of the bar, menus and panels", eg="Google Sans Flex")
row("Look", "FONTS", "appearance.fonts.title", "text", "Title font", "Headings in panels and popups", eg="Google Sans Flex")
row("Look", "FONTS", "appearance.fonts.monospace", "text", "Monospace font", "Code, numbers and the terminal-style parts", eg="JetBrains Mono NF")
row("Look", "WALLPAPER", "toren.wallpaperPicker", "action", "Wallpaper", "Pick the wallpaper; the colours follow it",
    store="action", actionLabel="Choose", run=["toren-wallpaper", "wallpaper", "ui"])
row("Look", "WALLPAPER COLOURS", "appearance.wallpaperTheming.enableAppsAndShell", "sw", "Desktop and GTK apps", "Take their colours from the wallpaper")
row("Look", "WALLPAPER COLOURS", "appearance.wallpaperTheming.enableQtApps", "sw", "Qt apps", "Take their colours from the wallpaper")
row("Look", "WALLPAPER COLOURS", "appearance.wallpaperTheming.enableTerminal", "sw", "Terminal", "Takes its colours from the wallpaper")
row("Look", "BACKGROUND", "background.parallax.enableWorkspace", "sw", "Parallax on workspaces", "The wallpaper slides a little as you change workspace")
row("Look", "BACKGROUND", "background.parallax.workspaceZoom", "slid", "Parallax zoom", "How much the wallpaper is enlarged to slide", **{"lo": 1.0, "hi": 1.5})
row("Look", "BACKGROUND", "background.widgets.weather.enable", "sw", "Weather on the desktop", "A weather widget drawn on the wallpaper")
row("Look", "OVERVIEW", "overview.scale", "slid", "Overview size", "Each workspace's size in the overview, against the screen", **{"lo": 0.1, "hi": 0.4})
row("Look", "OVERVIEW", "overview.centerIcons", "sw", "Centred app icons", "Draws each window's app icon in its middle")
row("Look", "OVERVIEW", "overview.orderRightLeft", "sw", "Right to left", "Orders the workspaces from right to left")
row("Look", "OVERVIEW", "overview.orderBottomUp", "sw", "Bottom to top", "Orders the rows from the bottom up")

row("Tools", "SCREENSHOTS", "screenSnip.savePath", "text", "Screenshot folder", "Where screenshots are saved; empty copies them to the clipboard only", eg="~/Pictures/Screenshots")
row("Tools", "SCREENSHOTS", "screenRecord.savePath", "text", "Recording folder", "Where screen recordings are saved; empty uses Videos", eg="~/Videos")
row("Tools", "SCREENSHOTS", "regionSelector.targetRegions.windows", "sw", "Snap to windows", "The region selector offers whole windows")
row("Tools", "SCREENSHOTS", "regionSelector.targetRegions.layers", "sw", "Snap to panels", "The region selector offers the bar and other panels")
row("Tools", "SCREENSHOTS", "regionSelector.targetRegions.content", "sw", "Snap to content", "The region selector offers areas inside windows")
row("Tools", "SCREENSHOTS", "regionSelector.rect.showAimLines", "sw", "Aim lines", "Guide lines across the screen while selecting")
row("Tools", "SEARCH", "search.engineBaseUrl", "text", "Web search", "The address a web search opens; the query is added at the end", eg="https://duckduckgo.com/?q=")
row("Tools", "SEARCH", "search.sloppy", "sw", "Forgiving search", "Finds apps even with typos (slower on many apps)")
row("Tools", "SEARCH PREFIXES", "search.prefix.app", "text", "Apps", "Type this first to search only apps", eg=">")
row("Tools", "SEARCH PREFIXES", "search.prefix.clipboard", "text", "Clipboard", "Type this first to search the clipboard", eg=";")
row("Tools", "SEARCH PREFIXES", "search.prefix.emojis", "text", "Emoji", "Type this first to search emoji", eg=":")
row("Tools", "SEARCH PREFIXES", "search.prefix.math", "text", "Calculator", "Type this first to calculate", eg="=")
row("Tools", "SEARCH PREFIXES", "search.prefix.shellCommand", "text", "Command", "Type this first to run a shell command", eg="$")
row("Tools", "SEARCH PREFIXES", "search.prefix.webSearch", "text", "Web", "Type this first to search the web", eg="?")

row("Lock screen", "LOCK SCREEN", "lock.launchOnStartup", "sw", "Lock at login", "Starts the session locked")
row("Lock screen", "SECURITY", "lock.security.unlockKeyring", "sw", "Unlock the keyring", "Unlocking the screen also unlocks saved passwords")

row("System", "NOTIFICATIONS", "notifications.forceMonitor.enable", "sw", "One screen only", "Shows notifications on the screen named below")
row("System", "NOTIFICATIONS", "notifications.forceMonitor.name", "text", "Screen", "The screen's name, as Displays lists it", eg="eDP-1")
row("System", "CHARGING", "toren.chargeLimit", "step", "Charge limit",
    "Stops charging at this level to keep the battery healthy; kept across reboots. Asks for your password",
    store="charge-limit", unit="%", **{"from": 50, "to": 100, "def": 80})
row("System", "WARNINGS", "battery.full", "step", "Full battery notice", "Tells you when the battery reaches this level (101 = never)", unit="%", **{"from": 50, "to": 101})
row("System", "SUSPEND", "battery.suspend", "step", "Suspend at", "The level at which the computer suspends", unit="%", **{"from": 1, "to": 15})
row("System", "VOLUME", "audio.protection.maxAllowedIncrease", "step", "Largest jump", "The most the volume can rise in one step", unit="%", **{"from": 1, "to": 50})
row("System", "VOLUME", "audio.protection.maxAllowed", "step", "Highest volume", "The volume never goes above this", unit="%", **{"from": 10, "to": 150})
row("System", "PERFORMANCE", "resources.updateInterval", "step", "Resource refresh", "How often CPU and memory are read for the bar", unit="ms", **{"from": 1000, "to": 10000})

# tabs in this order; rows keep their order inside a tab
TABS = ["Bar & dock", "Look", "Clock", "Lock screen", "Tools", "System"]
R.sort(key=lambda r: TABS.index(r["tab"]))

out = ".pragma library\n\n"
out += "// The default desktop's settings (~/.config/illogical-impulse/config.json),\n"
out += "// as the Hub shows them. `key` is the dotted path in that file, `def` the\n"
out += "// desktop's own default (modules/common/Config.qml), and `map`, where\n"
out += "// present, turns a shown choice into the value the file stores.\n"
out += "// Generated from the desktop's Config.qml; regenerate when it changes.\n\n"
out += "var rows = " + json.dumps(R, indent=4, ensure_ascii=False) + ";\n"
sys.stdout.write(out)
