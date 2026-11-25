pragma Singleton
import QtQuick
import Quickshell.Hyprland

QtObject {
    id: root

    function escapeLua(str) {
        if (str === null || str === undefined) return '""';
        return '"' + String(str).replace(/\\/g, '\\\\').replace(/"/g, '\\"').replace(/\n/g, '\\n') + '"';
    }

    function toLua(legacy) {
        if (!legacy) return "";
        const s = String(legacy).trim();
        if (s.startsWith("hl.")) return s;

        // Parse command and argument
        const firstSpace = s.indexOf(' ');
        const cmd = firstSpace === -1 ? s : s.slice(0, firstSpace);
        const args = firstSpace === -1 ? "" : s.slice(firstSpace + 1).trim();

        switch (cmd) {
        case "workspace":
            return `hl.dsp.focus({ workspace = ${escapeLua(args)} })`;

        case "movetoworkspace": {
            // movetoworkspace X[,address:A]
            const commaIdx = args.indexOf(',');
            if (commaIdx !== -1) {
                const ws = args.slice(0, commaIdx).trim();
                const win = args.slice(commaIdx + 1).trim();
                return `hl.dsp.window.move({ workspace = ${escapeLua(ws)}, window = ${escapeLua(win)} })`;
            }
            return `hl.dsp.window.move({ workspace = ${escapeLua(args)} })`;
        }

        case "movetoworkspacesilent": {
            // movetoworkspacesilent X[,address:A]
            const commaIdx = args.indexOf(',');
            if (commaIdx !== -1) {
                const ws = args.slice(0, commaIdx).trim();
                const win = args.slice(commaIdx + 1).trim();
                return `hl.dsp.window.move({ workspace = ${escapeLua(ws)}, follow = false, window = ${escapeLua(win)} })`;
            }
            return `hl.dsp.window.move({ workspace = ${escapeLua(args)}, follow = false })`;
        }

        case "focuswindow":
            return `hl.dsp.focus({ window = ${escapeLua(args)} })`;

        case "closewindow":
            return `hl.dsp.window.close({ window = ${escapeLua(args)} })`;

        case "killactive":
            return `hl.dsp.window.close()`;

        case "pin":
            if (args) {
                return `hl.dsp.window.pin({ window = ${escapeLua(args)} })`;
            }
            return `hl.dsp.window.pin()`;

        case "togglefloating":
            if (args) {
                return `hl.dsp.window.float({ window = ${escapeLua(args)} })`;
            }
            return `hl.dsp.window.float()`;

        case "movewindowpixel": {
            // movewindowpixel exact X Y,address:A
            let rest = args;
            if (rest.startsWith("exact ")) {
                rest = rest.slice(6).trim();
            }
            const commaIdx = rest.indexOf(',');
            if (commaIdx !== -1) {
                const coords = rest.slice(0, commaIdx).trim().split(/\s+/);
                const win = rest.slice(commaIdx + 1).trim();
                const x = parseFloat(coords[0]) || 0;
                const y = parseFloat(coords[1]) || 0;
                return `hl.dsp.window.move({ x = ${x}, y = ${y}, window = ${escapeLua(win)} })`;
            }
            const coords = rest.trim().split(/\s+/);
            const x = parseFloat(coords[0]) || 0;
            const y = parseFloat(coords[1]) || 0;
            return `hl.dsp.window.move({ x = ${x}, y = ${y} })`;
        }

        case "resizeactive": {
            // resizeactive exact W H  or  resizeactive W H
            let rest = args;
            let relative = true;
            if (rest.startsWith("exact ")) {
                rest = rest.slice(6).trim();
                relative = false;
            }
            const parts = rest.split(/\s+/);
            const w = parseFloat(parts[0]) || 0;
            const h = parseFloat(parts[1]) || 0;
            return `hl.dsp.window.resize({ x = ${w}, y = ${h}${relative ? ", relative = true" : ""} })`;
        }

        case "centerwindow":
            return `hl.dsp.window.center()`;

        case "bringactivetotop":
            return `hl.dsp.window.bring_to_top()`;

        case "cyclenext":
            if (args === "prev") {
                return `hl.dsp.window.cycle_next({ next = false })`;
            }
            return `hl.dsp.window.cycle_next()`;

        case "global":
            return `hl.dsp.global(${escapeLua(args)})`;

        case "submap":
            return `hl.dsp.submap(${escapeLua(args)})`;

        case "dpms": {
            const parts = args.split(/\s+/);
            const action = parts[0];
            if (parts.length > 1) {
                return `hl.dsp.dpms({ action = ${escapeLua(action)}, monitor = ${escapeLua(parts[1])} })`;
            }
            return `hl.dsp.dpms({ action = ${escapeLua(action)} })`;
        }

        case "exec": {
            let cmdStr = args;
            if (cmdStr.includes("hyprctl keyword cursor:no_warps")) {
                cmdStr = cmdStr.replace(/hyprctl keyword cursor:no_warps\s+(true|false)/g, function(_, val) {
                    return `hyprctl eval 'hl.config({ cursor = { no_warps = ${val} } })'`;
                });
            }
            return `hl.dsp.exec_cmd(${escapeLua(cmdStr)})`;
        }

        default:
            return s;
        }
    }

    function run(legacy) {
        const lua = toLua(legacy);
        return Hyprland.dispatch(lua);
    }
}
