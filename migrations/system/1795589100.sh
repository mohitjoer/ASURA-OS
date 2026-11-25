#!/bin/sh
echo "Remove duplicate toren desktop entries and switch runit services to toren-*"

ROOT="${ROOT:-}"

for app in hub installer manual store vm; do
	if [ -f "$ROOT/usr/share/applications/toren-${app}.desktop" ]; then
		rm -f "$ROOT/usr/share/applications/toren-${app}.desktop"
	fi
done

if [ -f "$ROOT/usr/share/wayland-sessions/toren.desktop" ]; then
	rm -f "$ROOT/usr/share/wayland-sessions/toren.desktop"
fi
if [ -f "$ROOT/usr/share/wayland-sessions/toren-safe.desktop" ]; then
	rm -f "$ROOT/usr/share/wayland-sessions/toren-safe.desktop"
fi

for svc in sysd boot-guard maintenance; do
	if [ -d "$ROOT/etc/sv/toren-$svc" ]; then
		if [ -L "$ROOT/etc/runit/runsvdir/default/toren-$svc" ] || [ ! -e "$ROOT/etc/runit/runsvdir/default/toren-$svc" ]; then
			rm -f "$ROOT/etc/runit/runsvdir/default/toren-$svc"
			mkdir -p "$ROOT/etc/runit/runsvdir/default"
			ln -sf "/etc/sv/toren-$svc" "$ROOT/etc/runit/runsvdir/default/toren-$svc"
		fi
	fi
done

exit 0
