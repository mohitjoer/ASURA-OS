#!/bin/sh
echo "Laptop fixes keep their battery rule under their own name; the default rule is Toren's again"

# The ASUS and Framework laptop fixes wrote their charge limit into
# /etc/udev/rules.d/99-battery-threshold.rules, Toren's default file, and every
# update put the default back over it. They now use 99-asus-battery.rules and
# 99-framework-battery.rules; re-apply the matched fixes so the file appears.
command -v toren >/dev/null 2>&1 || exit 0
[ -f /var/lib/toren/hardware/applied.json ] || exit 0
grep -Eq '"(asus-laptop|asus-tuf-fx507zc4|framework-laptop)"' /var/lib/toren/hardware/applied.json || exit 0
toren hardware quirks apply >/dev/null 2>&1 || true
exit 0
