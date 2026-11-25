// The ASURA "do everything" menu catalog.
// Each entry corresponds to an action in the single control menu tree.
// Data only, validated in a node test.

var CATEGORIES = [
  "All",
  "Apps",
  "Learn",
  "Style",
  "Setup",
  "Install / Remove",
  "Update",
  "Capture",
  "System",
];

var CATALOG = [
  // Apps
  {
    id: "apps-all",
    name: "Applications",
    category: "Apps",
    icon: "view-app-grid",
    desc: "Browse and launch desktop applications",
    mode: "all",
  },

  // Learn
  {
    id: "learn-cheatsheet",
    name: "Keybindings Cheatsheet",
    category: "Learn",
    icon: "input-keyboard",
    desc: "Interactive keyboard shortcuts legend",
    exec: ["toren", "ipc", "call", "toren", "toggleCheatsheet"],
  },
  {
    id: "learn-docs",
    name: "ASURA Documentation",
    category: "Learn",
    icon: "help-browser",
    desc: "Browse system guides and architecture notes",
    exec: ["toren-app", "browser", "https://torenos.dev/docs"],
  },

  // Style
  {
    id: "style-theme",
    name: "Theme & Palette",
    category: "Style",
    icon: "preferences-desktop-theme",
    desc: "Monochrome schemes, dark and light variants",
    exec: ["toren-shell", "hub", "open", "desktop"],
  },
  {
    id: "style-wallpaper",
    name: "Wallpaper Picker",
    category: "Style",
    icon: "preferences-desktop-wallpaper",
    desc: "Select or shuffle desktop wallpapers",
    exec: ["toren-wallpaper", "wallpaper", "ui"],
  },
  {
    id: "style-follow",
    name: "Toggle Wallpaper Colors",
    category: "Style",
    icon: "color-management",
    desc: "Match UI theme to active wallpaper",
    exec: ["toren-wallpaper", "wallpaper", "random"],
  },

  // Setup
  {
    id: "setup-wifi",
    name: "Wi-Fi & Network",
    category: "Setup",
    icon: "network-wireless",
    desc: "Manage wireless and wired connections",
    exec: ["toren-shell", "hub", "open", "connections"],
  },
  {
    id: "setup-bluetooth",
    name: "Bluetooth",
    category: "Setup",
    icon: "preferences-system-bluetooth",
    desc: "Pair and manage Bluetooth devices",
    exec: ["toren-shell", "hub", "open", "connections"],
  },
  {
    id: "setup-displays",
    name: "Displays & Monitors",
    category: "Setup",
    icon: "preferences-desktop-display",
    desc: "Resolution, refresh rate, and scaling",
    exec: ["toren-shell", "hub", "open", "displays"],
  },
  {
    id: "setup-audio",
    name: "Audio Settings",
    category: "Setup",
    icon: "audio-card",
    desc: "Volume, inputs, outputs, and equalizer",
    exec: ["toren-shell", "hub", "open", "audio"],
  },
  {
    id: "setup-power",
    name: "Power Profile",
    category: "Setup",
    icon: "battery",
    desc: "Performance, balanced, and power-saver profiles",
    exec: ["ghostty", "--title=ASURA Power", "-e", "toren", "power"],
  },
  {
    id: "setup-gpu",
    name: "GPU & Graphics",
    category: "Setup",
    icon: "video-display",
    desc: "Discrete GPU status and prime-run offload",
    exec: ["ghostty", "--title=ASURA GPU", "-e", "toren", "gpu"],
  },
  {
    id: "setup-input",
    name: "Keyboard & Touchpad",
    category: "Setup",
    icon: "input-touchpad",
    desc: "Key repeat rates, layout, and touchpad gestures",
    exec: ["toren-shell", "hub", "open", "input"],
  },

  // Install / Remove
  {
    id: "software-store",
    name: "Software Store",
    category: "Install / Remove",
    icon: "software-store",
    desc: "Discover and manage desktop applications",
    exec: ["toren-store"],
  },
  {
    id: "package-manager",
    name: "Package Manager (apm)",
    category: "Install / Remove",
    icon: "system-software-install",
    desc: "Search and install packages via apm",
    exec: ["ghostty", "--title=Package Manager (apm)", "-e", "apm"],
  },

  // Update
  {
    id: "update-toren",
    name: "ASURA Update",
    category: "Update",
    icon: "system-software-update",
    desc: "Update desktop components and shell with rollback",
    exec: [
      "ghostty",
      "--title=ASURA Update",
      "-e",
      "sh",
      "-c",
      "toren update; echo; echo 'Press Enter to close'; read _",
    ],
  },
  {
    id: "update-system",
    name: "Full System Update",
    category: "Update",
    icon: "software-update-available",
    desc: "Update base system packages via XBPS",
    exec: [
      "ghostty",
      "--title=System Update",
      "-e",
      "sh",
      "-c",
      "sudo xbps-install -Syu; toren update; echo; echo 'Press Enter to close'; read _",
    ],
  },
  {
    id: "update-rollback",
    name: "System Rollback",
    category: "Update",
    icon: "edit-undo",
    desc: "Re-materialize previous system snapshot",
    exec: [
      "ghostty",
      "--title=ASURA Rollback",
      "-e",
      "sh",
      "-c",
      "toren rollback; echo; echo 'Press Enter to close'; read _",
    ],
  },
  {
    id: "update-doctor",
    name: "System Diagnostics (Doctor)",
    category: "Update",
    icon: "dialog-information",
    desc: "Check configuration drift and health",
    exec: [
      "ghostty",
      "--title=ASURA Doctor",
      "-e",
      "sh",
      "-c",
      "toren doctor; echo; echo 'Press Enter to close'; read _",
    ],
  },

  // Capture
  {
    id: "capture-region",
    name: "Capture Region",
    category: "Capture",
    icon: "camera-photo",
    desc: "Interactive snipping tool and annotation",
    exec: ["toren-shot"],
  },
  {
    id: "capture-screen",
    name: "Capture Screen",
    category: "Capture",
    icon: "video-display",
    desc: "Quick full-screen capture",
    exec: ["toren", "ipc", "call", "toren", "toggleScreenshot"],
  },
  {
    id: "capture-record",
    name: "Screen Recording",
    category: "Capture",
    icon: "media-record",
    desc: "Record display or window video",
    exec: ["toren-cmd-screenrecord"],
  },

  // System
  {
    id: "system-lock",
    name: "Lock Screen",
    category: "System",
    icon: "system-lock-screen",
    desc: "Lock the desktop session",
    exec: ["toren-shell", "lock"],
  },
  {
    id: "system-suspend",
    name: "Suspend",
    category: "System",
    icon: "system-suspend",
    desc: "Suspend machine to RAM",
    exec: ["loginctl", "suspend"],
  },
  {
    id: "system-logout",
    name: "Log Out",
    category: "System",
    icon: "system-log-out",
    desc: "End current Wayland session",
    exec: ["hyprctl", "dispatch", "exit"],
  },
  {
    id: "system-reboot",
    name: "Reboot",
    category: "System",
    icon: "system-reboot",
    desc: "Restart the computer",
    exec: ["loginctl", "reboot"],
  },
  {
    id: "system-shutdown",
    name: "Shut Down",
    category: "System",
    icon: "system-shutdown",
    desc: "Power off the computer",
    exec: ["loginctl", "poweroff"],
  },
];

function commands() {
  var seen = {};
  return CATALOG.map(function (a) {
    return a.exec ? a.exec[0] : "";
  }).filter(function (c) {
    if (!c || seen[c]) return false;
    seen[c] = true;
    return true;
  });
}

function validate(catalog) {
  var problems = [];
  var seen = {};
  var cats = {};
  for (var i = 1; i < CATEGORIES.length; i++) cats[CATEGORIES[i]] = true;
  for (var j = 0; j < catalog.length; j++) {
    var a = catalog[j];
    if (!a.id) {
      problems.push("entry " + j + " missing id");
      continue;
    }
    if (seen[a.id]) problems.push("duplicate id: " + a.id);
    seen[a.id] = true;
    if (!a.name) problems.push(a.id + " missing name");
    if (!cats[a.category])
      problems.push(a.id + " has unknown category: " + a.category);
    if (!a.mode && (!Array.isArray(a.exec) || a.exec.length === 0))
      problems.push(a.id + " has empty exec and no mode");
  }
  return problems;
}

if (typeof module !== "undefined" && module.exports) {
  module.exports = { CATALOG, CATEGORIES, validate, commands };
}
