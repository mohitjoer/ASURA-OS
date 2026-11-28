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
        "tab": "Bar & dock",
        "group": "DOCK",
        "key": "appearance.transparency.enable",
        "ctl": "sw",
        "label": "Transparency",
        "desc": "Lets the wallpaper show through panels and menus",
        "def": false
    },
    {
        "tab": "Bar & dock",
        "group": "SHAPE",
        "key": "bar.borderless",
        "ctl": "sw",
        "label": "Plain items",
        "desc": "Draws the bar's items without the boxes that group them",
        "def": false
    },
    {
        "tab": "Bar & dock",
        "group": "WORKSPACES",
        "key": "bar.workspaces.showAppIcons",
        "ctl": "sw",
        "label": "App icons",
        "desc": "Shows each workspace's main app instead of a dot",
        "def": true
    },
    {
        "tab": "Bar & dock",
        "group": "INDICATORS",
        "key": "bar.indicators.notifications.showUnreadCount",
        "ctl": "sw",
        "label": "Unread count",
        "desc": "Shows how many notifications are waiting",
        "def": false
    },
    {
        "tab": "Bar & dock",
        "group": "INDICATORS",
        "key": "bar.tooltips.clickToShow",
        "ctl": "sw",
        "label": "Tooltips on click",
        "desc": "Shows the bar's tooltips on click instead of on hover",
        "def": false
    },
    {
        "tab": "Bar & dock",
        "group": "TRAY",
        "key": "tray.monochromeIcons",
        "ctl": "sw",
        "label": "Monochrome tray icons",
        "desc": "Tints the tray's app icons to the colour scheme",
        "def": true
    },
    {
        "tab": "Bar & dock",
        "group": "WEATHER",
        "key": "bar.weather.fetchInterval",
        "ctl": "step",
        "label": "Weather refresh",
        "desc": "How often the weather is fetched",
        "def": 10,
        "unit": "min",
        "from": 5,
        "to": 120
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
        "tab": "Look",
        "group": "FONTS",
        "key": "appearance.fonts.main",
        "ctl": "text",
        "label": "Interface font",
        "desc": "The font of the bar, menus and panels",
        "def": "Google Sans Flex",
        "eg": "Google Sans Flex"
    },
    {
        "tab": "Look",
        "group": "FONTS",
        "key": "appearance.fonts.title",
        "ctl": "text",
        "label": "Title font",
        "desc": "Headings in panels and popups",
        "def": "Google Sans Flex",
        "eg": "Google Sans Flex"
    },
    {
        "tab": "Look",
        "group": "FONTS",
        "key": "appearance.fonts.monospace",
        "ctl": "text",
        "label": "Monospace font",
        "desc": "Code, numbers and the terminal-style parts",
        "def": "JetBrains Mono NF",
        "eg": "JetBrains Mono NF"
    },
    {
        "tab": "Look",
        "group": "WALLPAPER",
        "key": "toren.wallpaperPicker",
        "ctl": "action",
        "label": "Wallpaper",
        "desc": "Pick the wallpaper; the colours follow it",
        "store": "action",
        "actionLabel": "Choose",
        "run": [
            "toren-wallpaper",
            "wallpaper",
            "ui"
        ]
    },
    {
        "tab": "Look",
        "group": "WALLPAPER COLOURS",
        "key": "appearance.wallpaperTheming.enableAppsAndShell",
        "ctl": "sw",
        "label": "Desktop and GTK apps",
        "desc": "Take their colours from the wallpaper",
        "def": true
    },
    {
        "tab": "Look",
        "group": "WALLPAPER COLOURS",
        "key": "appearance.wallpaperTheming.enableQtApps",
        "ctl": "sw",
        "label": "Qt apps",
        "desc": "Take their colours from the wallpaper",
        "def": true
    },
    {
        "tab": "Look",
        "group": "WALLPAPER COLOURS",
        "key": "appearance.wallpaperTheming.enableTerminal",
        "ctl": "sw",
        "label": "Terminal",
        "desc": "Takes its colours from the wallpaper",
        "def": true
    },
    {
        "tab": "Look",
        "group": "BACKGROUND",
        "key": "background.parallax.enableWorkspace",
        "ctl": "sw",
        "label": "Parallax on workspaces",
        "desc": "The wallpaper slides a little as you change workspace",
        "def": false
    },
    {
        "tab": "Look",
        "group": "BACKGROUND",
        "key": "background.parallax.workspaceZoom",
        "ctl": "slid",
        "label": "Parallax zoom",
        "desc": "How much the wallpaper is enlarged to slide",
        "def": 1.07,
        "lo": 1.0,
        "hi": 1.5
    },
    {
        "tab": "Look",
        "group": "BACKGROUND",
        "key": "background.widgets.weather.enable",
        "ctl": "sw",
        "label": "Weather on the desktop",
        "desc": "A weather widget drawn on the wallpaper",
        "def": false
    },
    {
        "tab": "Look",
        "group": "OVERVIEW",
        "key": "overview.scale",
        "ctl": "slid",
        "label": "Overview size",
        "desc": "Each workspace's size in the overview, against the screen",
        "def": 0.18,
        "lo": 0.1,
        "hi": 0.4
    },
    {
        "tab": "Look",
        "group": "OVERVIEW",
        "key": "overview.centerIcons",
        "ctl": "sw",
        "label": "Centred app icons",
        "desc": "Draws each window's app icon in its middle",
        "def": true
    },
    {
        "tab": "Look",
        "group": "OVERVIEW",
        "key": "overview.orderRightLeft",
        "ctl": "sw",
        "label": "Right to left",
        "desc": "Orders the workspaces from right to left",
        "def": false
    },
    {
        "tab": "Look",
        "group": "OVERVIEW",
        "key": "overview.orderBottomUp",
        "ctl": "sw",
        "label": "Bottom to top",
        "desc": "Orders the rows from the bottom up",
        "def": false
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
        "tab": "Lock screen",
        "group": "LOCK SCREEN",
        "key": "lock.launchOnStartup",
        "ctl": "sw",
        "label": "Lock at login",
        "desc": "Starts the session locked",
        "def": false
    },
    {
        "tab": "Lock screen",
        "group": "SECURITY",
        "key": "lock.security.unlockKeyring",
        "ctl": "sw",
        "label": "Unlock the keyring",
        "desc": "Unlocking the screen also unlocks saved passwords",
        "def": true
    },
    {
        "tab": "Tools",
        "group": "SCREENSHOTS",
        "key": "screenSnip.savePath",
        "ctl": "text",
        "label": "Screenshot folder",
        "desc": "Where screenshots are saved; empty copies them to the clipboard only",
        "def": "",
        "eg": "~/Pictures/Screenshots"
    },
    {
        "tab": "Tools",
        "group": "SCREENSHOTS",
        "key": "screenRecord.savePath",
        "ctl": "text",
        "label": "Recording folder",
        "desc": "Where screen recordings are saved; empty uses Videos",
        "def": "",
        "eg": "~/Videos"
    },
    {
        "tab": "Tools",
        "group": "SCREENSHOTS",
        "key": "regionSelector.targetRegions.windows",
        "ctl": "sw",
        "label": "Snap to windows",
        "desc": "The region selector offers whole windows",
        "def": true
    },
    {
        "tab": "Tools",
        "group": "SCREENSHOTS",
        "key": "regionSelector.targetRegions.layers",
        "ctl": "sw",
        "label": "Snap to panels",
        "desc": "The region selector offers the bar and other panels",
        "def": false
    },
    {
        "tab": "Tools",
        "group": "SCREENSHOTS",
        "key": "regionSelector.targetRegions.content",
        "ctl": "sw",
        "label": "Snap to content",
        "desc": "The region selector offers areas inside windows",
        "def": true
    },
    {
        "tab": "Tools",
        "group": "SCREENSHOTS",
        "key": "regionSelector.rect.showAimLines",
        "ctl": "sw",
        "label": "Aim lines",
        "desc": "Guide lines across the screen while selecting",
        "def": true
    },
    {
        "tab": "Tools",
        "group": "SEARCH",
        "key": "search.engineBaseUrl",
        "ctl": "text",
        "label": "Web search",
        "desc": "The address a web search opens; the query is added at the end",
        "def": "https://www.google.com/search?q=",
        "eg": "https://duckduckgo.com/?q="
    },
    {
        "tab": "Tools",
        "group": "SEARCH",
        "key": "search.sloppy",
        "ctl": "sw",
        "label": "Forgiving search",
        "desc": "Finds apps even with typos (slower on many apps)",
        "def": false
    },
    {
        "tab": "Tools",
        "group": "SEARCH PREFIXES",
        "key": "search.prefix.app",
        "ctl": "text",
        "label": "Apps",
        "desc": "Type this first to search only apps",
        "def": ">",
        "eg": ">"
    },
    {
        "tab": "Tools",
        "group": "SEARCH PREFIXES",
        "key": "search.prefix.clipboard",
        "ctl": "text",
        "label": "Clipboard",
        "desc": "Type this first to search the clipboard",
        "def": ";",
        "eg": ";"
    },
    {
        "tab": "Tools",
        "group": "SEARCH PREFIXES",
        "key": "search.prefix.emojis",
        "ctl": "text",
        "label": "Emoji",
        "desc": "Type this first to search emoji",
        "def": ":",
        "eg": ":"
    },
    {
        "tab": "Tools",
        "group": "SEARCH PREFIXES",
        "key": "search.prefix.math",
        "ctl": "text",
        "label": "Calculator",
        "desc": "Type this first to calculate",
        "def": "=",
        "eg": "="
    },
    {
        "tab": "Tools",
        "group": "SEARCH PREFIXES",
        "key": "search.prefix.shellCommand",
        "ctl": "text",
        "label": "Command",
        "desc": "Type this first to run a shell command",
        "def": "$",
        "eg": "$"
    },
    {
        "tab": "Tools",
        "group": "SEARCH PREFIXES",
        "key": "search.prefix.webSearch",
        "ctl": "text",
        "label": "Web",
        "desc": "Type this first to search the web",
        "def": "?",
        "eg": "?"
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
    },
    {
        "tab": "System",
        "group": "NOTIFICATIONS",
        "key": "notifications.forceMonitor.enable",
        "ctl": "sw",
        "label": "One screen only",
        "desc": "Shows notifications on the screen named below",
        "def": false
    },
    {
        "tab": "System",
        "group": "NOTIFICATIONS",
        "key": "notifications.forceMonitor.name",
        "ctl": "text",
        "label": "Screen",
        "desc": "The screen's name, as Displays lists it",
        "def": "",
        "eg": "eDP-1"
    },
    {
        "tab": "System",
        "group": "CHARGING",
        "key": "toren.chargeLimit",
        "ctl": "step",
        "label": "Charge limit",
        "desc": "Stops charging at this level to keep the battery healthy; kept across reboots. Asks for your password",
        "store": "charge-limit",
        "unit": "%",
        "from": 50,
        "to": 100,
        "def": 80
    },
    {
        "tab": "System",
        "group": "WARNINGS",
        "key": "battery.full",
        "ctl": "step",
        "label": "Full battery notice",
        "desc": "Tells you when the battery reaches this level (101 = never)",
        "def": 101,
        "unit": "%",
        "from": 50,
        "to": 101
    },
    {
        "tab": "System",
        "group": "SUSPEND",
        "key": "battery.suspend",
        "ctl": "step",
        "label": "Suspend at",
        "desc": "The level at which the computer suspends",
        "def": 3,
        "unit": "%",
        "from": 1,
        "to": 15
    },
    {
        "tab": "System",
        "group": "VOLUME",
        "key": "audio.protection.maxAllowedIncrease",
        "ctl": "step",
        "label": "Largest jump",
        "desc": "The most the volume can rise in one step",
        "def": 10,
        "unit": "%",
        "from": 1,
        "to": 50
    },
    {
        "tab": "System",
        "group": "VOLUME",
        "key": "audio.protection.maxAllowed",
        "ctl": "step",
        "label": "Highest volume",
        "desc": "The volume never goes above this",
        "def": 99,
        "unit": "%",
        "from": 10,
        "to": 150
    },
    {
        "tab": "System",
        "group": "PERFORMANCE",
        "key": "resources.updateInterval",
        "ctl": "step",
        "label": "Resource refresh",
        "desc": "How often CPU and memory are read for the bar",
        "def": 3000,
        "unit": "ms",
        "from": 1000,
        "to": 10000
    }
];
