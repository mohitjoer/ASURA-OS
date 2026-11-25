.pragma library

// The default desktop's settings (~/.config/illogical-impulse/config.json),
// as the Hub shows them. `key` is the dotted path in that file, `def` the
// desktop's own default (modules/common/Config.qml), and `map`, where
// present, turns a shown choice into the value the file stores.
// Generated from the desktop's Config.qml; regenerate when it changes.

var rows = [
    {
        "tab": "Bar & dock",
        "group": "POSITION",
        "key": "bar.bottom",
        "ctl": "sw",
        "label": "Bar at the bottom",
        "desc": "Moves the bar from the top edge to the bottom edge",
        "def": false
    },
    {
        "tab": "Bar & dock",
        "group": "POSITION",
        "key": "bar.autoHide.enable",
        "ctl": "sw",
        "label": "Hide the bar",
        "desc": "The bar slides away and comes back when the pointer reaches its edge",
        "def": false
    },
    {
        "tab": "Bar & dock",
        "group": "SHAPE",
        "key": "bar.cornerStyle",
        "ctl": "seg",
        "label": "Bar silhouette & corners",
        "desc": "Hug: the bar meets the screen corners. Float: a raised strip. Rect: a flat rectangle",
        "def": 0,
        "opts": [
            "Hug",
            "Float",
            "Rect"
        ],
        "map": {
            "Hug": 0,
            "Float": 1,
            "Rect": 2
        }
    },
    {
        "tab": "Bar & dock",
        "group": "SHAPE",
        "key": "bar.showBackground",
        "ctl": "sw",
        "label": "Background",
        "desc": "Draws the bar's background; off leaves the items over the wallpaper",
        "def": true
    },
    {
        "tab": "Bar & dock",
        "group": "SHAPE",
        "key": "bar.floatStyleShadow",
        "ctl": "sw",
        "label": "Shadow",
        "desc": "A shadow under the bar when its shape is Float",
        "def": true
    },
    {
        "tab": "Bar & dock",
        "group": "WEATHER",
        "key": "bar.weather.enable",
        "ctl": "sw",
        "label": "Weather in the bar",
        "desc": "Shows the temperature and conditions next to the clock",
        "def": false
    },
    {
        "tab": "Bar & dock",
        "group": "WEATHER",
        "key": "bar.weather.enableGPS",
        "ctl": "sw",
        "label": "Find my location",
        "desc": "Uses the network location; turn off to name a city below",
        "def": true
    },
    {
        "tab": "Bar & dock",
        "group": "WEATHER",
        "key": "bar.weather.city",
        "ctl": "text",
        "label": "City",
        "desc": "Used when finding the location is off",
        "def": "",
        "eg": "Berlin"
    },
    {
        "tab": "Bar & dock",
        "group": "WEATHER",
        "key": "bar.weather.useUSCS",
        "ctl": "sw",
        "label": "Fahrenheit",
        "desc": "US units instead of Celsius",
        "def": false
    },
    {
        "tab": "Bar & dock",
        "group": "DOCK",
        "key": "dock.enable",
        "ctl": "sw",
        "label": "Dock",
        "desc": "The app dock at the bottom of the screen",
        "def": false
    },
    {
        "tab": "Bar & dock",
        "group": "DOCK",
        "key": "dock.pinnedOnStartup",
        "ctl": "sw",
        "label": "Always show",
        "desc": "Keeps the dock on screen instead of hiding it while a window is focused",
        "def": false
    },
    {
        "tab": "Bar & dock",
        "group": "DOCK",
        "key": "dock.monochromeIcons",
        "ctl": "sw",
        "label": "Monochrome icons",
        "desc": "Tints the dock's icons to the colour scheme",
        "def": true
    },
    {
        "tab": "Look",
        "group": "COLOURS",
        "key": "appearance.palette.type",
        "ctl": "seg",
        "label": "Colour scheme",
        "desc": "How colours are taken from the wallpaper",
        "def": "auto",
        "opts": [
            "Auto",
            "Tonal",
            "Neutral",
            "Vibrant",
            "Expressive",
            "Monochrome"
        ],
        "map": {
            "Auto": "auto",
            "Tonal": "scheme-tonal-spot",
            "Neutral": "scheme-neutral",
            "Vibrant": "scheme-rainbow",
            "Expressive": "scheme-expressive",
            "Monochrome": "scheme-monochrome"
        }
    },
    {
        "tab": "Look",
        "group": "COLOURS",
        "key": "appearance.transparency.enable",
        "ctl": "sw",
        "label": "Transparency",
        "desc": "Lets the wallpaper show through panels and menus",
        "def": false
    },
    {
        "tab": "Look",
        "group": "SCREEN",
        "key": "appearance.fakeScreenRounding",
        "ctl": "seg",
        "label": "Rounded screen corners",
        "desc": "Draws rounded corners over the screen's edges",
        "def": 2,
        "opts": [
            "Off",
            "Always",
            "Not fullscreen"
        ],
        "map": {
            "Off": 0,
            "Always": 1,
            "Not fullscreen": 2
        }
    },
    {
        "tab": "Look",
        "group": "APPS",
        "key": "toren.gtkRoundedCorners",
        "ctl": "sw",
        "label": "Rounded app corners",
        "desc": "GTK and libadwaita apps (Files, settings dialogs) draw rounded window corners. Off keeps them square. Applies to apps opened after the change",
        "store": "gtk-corners",
        "def": false
    },
    {
        "tab": "Look",
        "group": "OVERVIEW",
        "key": "overview.rows",
        "ctl": "step",
        "label": "Rows",
        "desc": "Workspaces per column in the overview",
        "def": 2,
        "from": 1,
        "to": 5
    },
    {
        "tab": "Look",
        "group": "OVERVIEW",
        "key": "overview.columns",
        "ctl": "step",
        "label": "Columns",
        "desc": "Workspaces per row in the overview",
        "def": 5,
        "from": 1,
        "to": 10
    },
    {
        "tab": "Look",
        "group": "NIGHT LIGHT",
        "key": "light.night.automatic",
        "ctl": "sw",
        "label": "On a schedule",
        "desc": "Warms the screen colours between the times below",
        "def": true
    },
    {
        "tab": "Look",
        "group": "NIGHT LIGHT",
        "key": "light.night.from",
        "ctl": "text",
        "label": "From",
        "desc": "24-hour time",
        "def": "19:00",
        "eg": "19:00"
    },
    {
        "tab": "Look",
        "group": "NIGHT LIGHT",
        "key": "light.night.to",
        "ctl": "text",
        "label": "To",
        "desc": "24-hour time",
        "def": "06:30",
        "eg": "06:30"
    },
    {
        "tab": "Look",
        "group": "NIGHT LIGHT",
        "key": "light.night.colorTemperature",
        "ctl": "step",
        "label": "Warmth",
        "desc": "Lower is warmer",
        "def": 5000,
        "unit": "K",
        "from": 2500,
        "to": 6500
    },
    {
        "tab": "Clock",
        "group": "TIME",
        "key": "time.format",
        "ctl": "seg",
        "label": "Time format",
        "desc": "How the bar and lock screen show the time",
        "def": "hh:mm",
        "opts": [
            "24-hour",
            "12-hour"
        ],
        "map": {
            "24-hour": "hh:mm",
            "12-hour": "h:mm ap"
        }
    },
    {
        "tab": "Clock",
        "group": "TIME",
        "key": "time.secondPrecision",
        "ctl": "sw",
        "label": "Seconds",
        "desc": "Shows seconds in the clock",
        "def": false
    },
    {
        "tab": "Clock",
        "group": "DESKTOP CLOCK",
        "key": "background.widgets.clock.enable",
        "ctl": "sw",
        "label": "Clock on the desktop",
        "desc": "A large clock drawn on the wallpaper",
        "def": false
    },
    {
        "tab": "Clock",
        "group": "DESKTOP CLOCK",
        "key": "background.widgets.clock.style",
        "ctl": "seg",
        "label": "Style",
        "desc": "Analogue dial or digits",
        "def": "cookie",
        "opts": [
            "Analogue",
            "Digital"
        ],
        "map": {
            "Analogue": "cookie",
            "Digital": "digital"
        }
    },
    {
        "tab": "Lock screen",
        "group": "LOCK SCREEN",
        "key": "lock.centerClock",
        "ctl": "sw",
        "label": "Clock in the centre",
        "desc": "Puts the lock screen clock in the middle of the screen",
        "def": true
    },
    {
        "tab": "Lock screen",
        "group": "LOCK SCREEN",
        "key": "lock.blur.enable",
        "ctl": "sw",
        "label": "Blur the wallpaper",
        "desc": "Blurs the wallpaper behind the lock screen",
        "def": true
    },
    {
        "tab": "Lock screen",
        "group": "LOCK SCREEN",
        "key": "lock.showLockedText",
        "ctl": "sw",
        "label": "\"Locked\" label",
        "desc": "Shows a Locked label above the password field",
        "def": true
    },
    {
        "tab": "Lock screen",
        "group": "SECURITY",
        "key": "lock.security.requirePasswordToPower",
        "ctl": "sw",
        "label": "Password to shut down",
        "desc": "Asks for the password before shutting down or restarting from the lock screen",
        "def": false
    },
    {
        "tab": "System",
        "group": "NOTIFICATIONS",
        "key": "notifications.timeout",
        "ctl": "step",
        "label": "Notification time",
        "desc": "How long a notification stays on screen",
        "def": 7000,
        "unit": "ms",
        "from": 1000,
        "to": 30000
    },
    {
        "tab": "System",
        "group": "NOTIFICATIONS",
        "key": "osd.timeout",
        "ctl": "step",
        "label": "Volume and brightness popup",
        "desc": "How long the volume or brightness popup stays",
        "def": 1000,
        "unit": "ms",
        "from": 500,
        "to": 5000
    },
    {
        "tab": "System",
        "group": "WARNINGS",
        "key": "battery.low",
        "ctl": "step",
        "label": "Low battery warning",
        "desc": "Warns when the battery falls to this level",
        "def": 20,
        "unit": "%",
        "from": 5,
        "to": 50
    },
    {
        "tab": "System",
        "group": "WARNINGS",
        "key": "battery.critical",
        "ctl": "step",
        "label": "Critical battery warning",
        "desc": "Warns again at this level",
        "def": 5,
        "unit": "%",
        "from": 1,
        "to": 20
    },
    {
        "tab": "System",
        "group": "WARNINGS",
        "key": "sounds.battery",
        "ctl": "sw",
        "label": "Warning sound",
        "desc": "Plays a sound with the battery warnings",
        "def": false
    },
    {
        "tab": "System",
        "group": "SUSPEND",
        "key": "battery.automaticSuspend",
        "ctl": "sw",
        "label": "Suspend when empty",
        "desc": "Suspends the computer when the battery is nearly empty",
        "def": true
    },
    {
        "tab": "System",
        "group": "VOLUME",
        "key": "audio.protection.enable",
        "ctl": "sw",
        "label": "Limit volume jumps",
        "desc": "Stops the volume from jumping up suddenly",
        "def": false
    }
];
