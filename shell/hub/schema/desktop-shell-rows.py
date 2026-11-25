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
        m=re.match(r'property (bool|int|real|string|list<string>|var) (\w+): ([^/]+?)\s*(//\s*(.*))?$',s)
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
    else: dv = json.loads(d)
    r = {"tab": tab, "group": group, "key": key, "ctl": ctl, "label": label, "desc": desc, "def": dv}
    r.update(kw); R.append(r)

row("Bar & dock", "POSITION", "bar.bottom", "sw", "Bar at the bottom", "Moves the bar from the top edge to the bottom edge")
row("Bar & dock", "POSITION", "bar.autoHide.enable", "sw", "Hide the bar", "The bar slides away and comes back when the pointer reaches its edge")
row("Bar & dock", "SHAPE", "bar.cornerStyle", "seg", "Bar silhouette & corners", "Hug: the bar meets the screen corners. Float: a raised strip. Rect: a flat rectangle",
    opts=["Hug", "Float", "Rect"], map={"Hug": 0, "Float": 1, "Rect": 2})
row("Bar & dock", "SHAPE", "bar.showBackground", "sw", "Background", "Draws the bar's background; off leaves the items over the wallpaper")
row("Bar & dock", "SHAPE", "bar.floatStyleShadow", "sw", "Shadow", "A shadow under the bar when its shape is Float")
row("Bar & dock", "WEATHER", "bar.weather.enable", "sw", "Weather in the bar", "Shows the temperature and conditions next to the clock")
row("Bar & dock", "WEATHER", "bar.weather.enableGPS", "sw", "Find my location", "Uses the network location; turn off to name a city below")
row("Bar & dock", "WEATHER", "bar.weather.city", "text", "City", "Used when finding the location is off", eg="Berlin")
row("Bar & dock", "WEATHER", "bar.weather.useUSCS", "sw", "Fahrenheit", "US units instead of Celsius")

row("Bar & dock", "DOCK", "dock.enable", "sw", "Dock", "The app dock at the bottom of the screen")
row("Bar & dock", "DOCK", "dock.pinnedOnStartup", "sw", "Always show", "Keeps the dock on screen instead of hiding it while a window is focused")
row("Bar & dock", "DOCK", "dock.monochromeIcons", "sw", "Monochrome icons", "Tints the dock's icons to the colour scheme")

row("Look", "COLOURS", "appearance.palette.type", "seg", "Colour scheme", "How colours are taken from the wallpaper",
    opts=["Auto", "Tonal", "Neutral", "Vibrant", "Expressive", "Monochrome"],
    map={"Auto": "auto", "Tonal": "scheme-tonal-spot", "Neutral": "scheme-neutral", "Vibrant": "scheme-rainbow",
         "Expressive": "scheme-expressive", "Monochrome": "scheme-monochrome"})
row("Look", "COLOURS", "appearance.transparency.enable", "sw", "Transparency", "Lets the wallpaper show through panels and menus")
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
row("Clock", "DESKTOP CLOCK", "background.widgets.clock.enable", "sw", "Clock on the desktop", "A large clock drawn on the wallpaper")
row("Clock", "DESKTOP CLOCK", "background.widgets.clock.style", "seg", "Style", "Analogue dial or digits",
    opts=["Analogue", "Digital"], map={"Analogue": "cookie", "Digital": "digital"})

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

# tabs in this order; rows keep their order inside a tab
TABS = ["Bar & dock", "Look", "Clock", "Lock screen", "System"]
R.sort(key=lambda r: TABS.index(r["tab"]))

out = ".pragma library\n\n"
out += "// The default desktop's settings (~/.config/illogical-impulse/config.json),\n"
out += "// as the Hub shows them. `key` is the dotted path in that file, `def` the\n"
out += "// desktop's own default (modules/common/Config.qml), and `map`, where\n"
out += "// present, turns a shown choice into the value the file stores.\n"
out += "// Generated from the desktop's Config.qml; regenerate when it changes.\n\n"
out += "var rows = " + json.dumps(R, indent=4, ensure_ascii=False) + ";\n"
sys.stdout.write(out)
