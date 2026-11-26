pragma Singleton
pragma ComponentBehavior: Bound

import qs.modules.common
import qs.modules.common.functions
import QtQuick
import Quickshell
import Quickshell.Io
import Quickshell.Hyprland

/**
 * Toren overlay (shell-desktop/toren/vendor.sh OVERLAYS): the cheatsheet's
 * keybindings come from Toren's legend (`toren-hubd keybinds`), the list the
 * settings window's Keybinds page shows: every shortcut with its name and group,
 * the user's rebinds and own shortcuts included. Toren's binds.lua gives
 * Hyprland's binds no descriptions, so `hyprctl binds` alone left the sheet empty.
 * Each row is shaped like a `hyprctl binds -j` entry (modmask, key,
 * "Group: what it does") so the cheatsheet draws it unchanged; a digit family
 * ("Super + 1..0") is one row the sheet shows as <Number>.
 */
Singleton {
    id: root
    property var keybinds: []
    property var keybindCategories: []

    readonly property var modBits: ({ "SUPER": 64, "CTRL": 4, "ALT": 8, "SHIFT": 1 })

    function groupName(id) {
        const s = String(id || "");
        if (s === "Hyprland") return "Other";
        return s.length ? s.charAt(0).toUpperCase() + s.slice(1) : "Other";
    }

    function rowFor(bind, group) {
        const chord = String(bind.chord || bind.combo || "");
        if (!chord) return null;
        let modmask = 0, key = "";
        for (const raw of chord.split("+")) {
            const tok = raw.trim();
            if (root.modBits[tok.toUpperCase()] !== undefined) modmask |= root.modBits[tok.toUpperCase()];
            else key = tok;
        }
        if (!key) return null;
        if (key === "{n}") key = "1";            // a digit family: the sheet shows <Number>
        key = key.replace(/^KP_/, "Num ");
        const label = String(bind.label || bind.desc || "").replace(/:/g, " -");
        if (!label) return null;
        return { modmask: modmask, key: key, description: group + ": " + label };
    }

    Connections {
        target: Hyprland

        function onRawEvent(event) {
            if (event.name == "configreloaded") {
                getKeybinds.running = true
            }
        }
    }

    Process {
        id: getKeybinds
        running: true
        command: ["toren-hubd", "keybinds"]

        stdout: StdioCollector {
            onStreamFinished: {
                try {
                    const legend = JSON.parse(text)
                    const rows = []
                    const groups = []
                    const seen = {}     // one row per chord: a custom shortcut is listed twice in the legend
                    // the compositor's own, unnamed binds last, so a named row wins a chord
                    const cats = (legend.categories || []).slice().sort((a, b) => (a.name === "Hyprland") - (b.name === "Hyprland"))
                    for (const cat of cats) {
                        const group = root.groupName(cat.name)
                        for (const b of (cat.binds || [])) {
                            const row = root.rowFor(b, group)
                            if (!row) continue
                            const id = row.modmask + "+" + row.key
                            if (seen[id]) continue
                            seen[id] = true
                            rows.push(row)
                            if (!groups.includes(group)) groups.push(group)
                        }
                    }
                    root.keybinds = rows
                    root.keybindCategories = groups
                } catch (e) {
                    console.error("[CheatsheetKeybinds] Error parsing keybinds:", e)
                }
            }
        }
    }
}
