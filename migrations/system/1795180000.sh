#!/bin/sh
echo "Register the ISO's prebuilt Hyprland packages under their names in the XBPS database"

# Images built before the fix registered Hyprland, its libraries, the portal,
# hypridle and hyprlock under "name-version_rev" keys, so xbps did not know
# they were installed and every update tried to fetch them again (and failed
# on the third-party repository's old sonames). Re-key them by name. An entry
# is moved only when its key is its own pkgver and its name is not taken.
pkgdb=/var/db/xbps/pkgdb-0.38.plist
[ -f "$pkgdb" ] || exit 0
command -v python3 >/dev/null 2>&1 || { echo "python3 is not installed; skipping"; exit 0; }
python3 -I - "$pkgdb" <<'PY'
import os, plistlib, sys
path = sys.argv[1]
with open(path, "rb") as f:
    db = plistlib.load(f)
fixed = []
for key in list(db):
    value = db[key]
    if key.startswith("_") or not isinstance(value, dict):
        continue
    name = value.get("pkgname")
    if name and key != name and key == value.get("pkgver") and name not in db:
        db[name] = db.pop(key)
        fixed.append(name)
if fixed:
    with open(path + ".toren-new", "wb") as f:
        plistlib.dump(db, f)
    os.replace(path + ".toren-new", path)
    print("registered under their names: " + " ".join(fixed))
PY
