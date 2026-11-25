#!/bin/sh
echo "Migrate user Hyprland configs to Lua format"

HYPR_DIR="${XDG_CONFIG_HOME:-$HOME/.config}/asura/hyprland"
[ -d "$HYPR_DIR" ] || exit 0

convert_file() {
    src="$1"
    dst="$2"
    bak="$1.bak"

    [ -f "$src" ] || return 0

    # Save .bak copy if not already backed up
    if [ ! -f "$bak" ]; then
        cp "$src" "$bak"
    fi

    # Convert to Lua using Python
    python3 - "$src" "$dst" <<'PYEOF'
import sys, re

src_path, dst_path = sys.argv[1], sys.argv[2]

def convert_line(line):
    raw = line.rstrip("\r\n")
    stripped = raw.strip()
    if not stripped:
        return ""
    if stripped.startswith("#"):
        return "-- " + stripped.lstrip("# ")

    # monitor = OUTPUT, MODE, POS, SCALE
    m_mon = re.match(r"^\s*monitor\s*=\s*(.*)$", stripped)
    if m_mon:
        parts = [p.strip() for p in m_mon.group(1).split(",")]
        if len(parts) >= 4:
            out, mode, pos, scale = parts[0], parts[1], parts[2], parts[3]
            return f'hl.monitor({{ output = "{out}", mode = "{mode}", position = "{pos}", scale = {scale} }})'
        elif len(parts) >= 1 and parts[0]:
            return f'hl.monitor({{ output = "{parts[0]}" }})'

    # env = KEY, VAL
    m_env = re.match(r"^\s*env\s*=\s*([^,]+),\s*(.*)$", stripped)
    if m_env:
        k = m_env.group(1).strip()
        v = m_env.group(2).strip()
        return f'hl.env("{k}", "{v}")'

    # exec-once = CMD
    m_eo = re.match(r"^\s*exec-once\s*=\s*(.*)$", stripped)
    if m_eo:
        cmd = m_eo.group(1).strip().replace('"', '\\"')
        return f'hl.on("hyprland.start", function() hl.exec_cmd("{cmd}") end)'

    # exec = CMD
    m_e = re.match(r"^\s*exec\s*=\s*(.*)$", stripped)
    if m_e:
        cmd = m_e.group(1).strip().replace('"', '\\"')
        return f'hl.exec_cmd("{cmd}")'

    # bind[flags] = MODS, KEY, DISP, ARGS
    m_bind = re.match(r"^\s*(bind[a-z]*)\s*=\s*(.*)$", stripped)
    if m_bind:
        dir_name = m_bind.group(1)
        parts = [p.strip() for p in m_bind.group(2).split(",")]
        if len(parts) >= 3:
            mods = parts[0]
            key = parts[1]
            disp = parts[2]
            args = ", ".join(parts[3:]) if len(parts) > 3 else ""
            chord_parts = []
            if mods:
                chord_parts.extend(mods.split())
            chord_parts.append(key)
            chord = " + ".join(chord_parts)

            opts = []
            if "l" in dir_name[4:]: opts.append("locked = true")
            if "e" in dir_name[4:]: opts.append("repeating = true")
            if "r" in dir_name[4:]: opts.append("release = true")
            if "m" in dir_name[4:]: opts.append("mouse = true")
            opt_str = f", {{ {', '.join(opts)} }}" if opts else ""

            if disp == "exec":
                cmd = args.replace('"', '\\"')
                return f'hl.bind("{chord}", hl.dsp.exec_cmd("{cmd}"){opt_str})'
            else:
                return f'hl.bind("{chord}", hl.dispatch("{disp}", "{args}"){opt_str})'

    return f"-- unconverted: {raw}"

try:
    with open(src_path, "r", encoding="utf-8", errors="replace") as f:
        lines = f.readlines()
    out = [convert_line(l) for l in lines]
    with open(dst_path, "w", encoding="utf-8") as f:
        f.write("\n".join(out) + "\n")
except Exception as e:
    sys.exit(1)
PYEOF
}

convert_file "$HYPR_DIR/user.conf" "$HYPR_DIR/user.lua"
convert_file "$HYPR_DIR/monitors-user.conf" "$HYPR_DIR/monitors-user.lua"

exit 0
