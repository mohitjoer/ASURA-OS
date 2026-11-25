#!/bin/sh
echo "Migrate legacy state files and profile formats to current names"

# 1. State files: rename toren-ryowalls.json to toren-walls.json
for dir in "${XDG_STATE_HOME:-$HOME/.local/state}" "${XDG_STATE_HOME:-$HOME/.local/state}/toren"; do
    old="$dir/toren-ryowalls.json"
    new="$dir/toren-walls.json"
    if [ -f "$old" ]; then
        if [ ! -f "$new" ]; then
            mv "$old" "$new"
        else
            rm -f "$old"
        fi
    fi
done

# 2. Decor art directory: Pictures/ryodecors -> Pictures/decors
old_decors="$HOME/Pictures/ryodecors"
new_decors="$HOME/Pictures/decors"
if [ -d "$old_decors" ]; then
    if [ ! -e "$new_decors" ]; then
        mv "$old_decors" "$new_decors"
    else
        cp -rn "$old_decors"/* "$new_decors"/ 2>/dev/null || true
        rm -rf "$old_decors"
    fi
fi

# 3. Profile files: convert *.ryoprofile to *.torenprofile with "torenprofile" key
command -v python3 >/dev/null 2>&1 && python3 -I - <<'PY'
import os, json

home = os.environ.get("HOME", "")
if not home or not os.path.isdir(home):
    raise SystemExit(0)

matches = []
for root, dirs, files in os.walk(home):
    rel = os.path.relpath(root, home)
    depth = 0 if rel == "." else len(rel.split(os.sep))
    if depth > 3:
        dirs.clear()
        continue
    for f in files:
        if f.endswith(".ryoprofile"):
            matches.append(os.path.join(root, f))

for real in matches:
    try:
        with open(real, "r", encoding="utf-8") as f:
            data = json.load(f)
        if isinstance(data, dict):
            if "ryoprofile" in data:
                data["torenprofile"] = data.pop("ryoprofile")
            new_path = real[:-11] + ".torenprofile"
            with open(new_path, "w", encoding="utf-8") as f:
                json.dump(data, f)
            os.remove(real)
    except Exception:
        pass
PY

exit 0
