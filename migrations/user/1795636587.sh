#!/bin/sh
echo "Blur, shadows and motion are set on the Window Manager page only; clear the old Performance switches"

# Settings had a second set of switches for these on the Performance page
# (disableBlur, disableShadows, reduceMotion in performance.json). Those rows
# are gone, so a switch left on would keep blur or shadows off with nothing in
# Settings to turn it back. Low power mode and the power profile still turn
# them down while they are on.
f="${XDG_CONFIG_HOME:-$HOME/.config}/toren/performance.json"
[ -f "$f" ] || exit 0
grep -Eq '"(disableBlur|disableShadows|reduceMotion)"' "$f" || exit 0
tmp="$f.toren-migrate"
sed -E 's/"(disableBlur|disableShadows|reduceMotion)"([[:space:]]*:[[:space:]]*)true/"\1"\2false/g' "$f" >"$tmp" && mv -f "$tmp" "$f"
exit 0
