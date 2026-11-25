#!/bin/sh
echo "Add the Toren Linux repository signing key if it is missing"

KEY_FP="86:9a:a2:38:85:e8:65:cb:53:36:a7:ff:b9:08:a6:05"
KEY_FILE="/var/db/xbps/keys/${KEY_FP}.plist"

if [ -f "$KEY_FILE" ]; then
	exit 0
fi

mkdir -p /var/db/xbps/keys
for src in /usr/share/asura/xbps/keys /usr/share/asura/keys; do
	[ -d "$src" ] || continue
	for f in "$src"/*.plist; do
		[ -f "$f" ] || continue
		dst="/var/db/xbps/keys/$(basename "$f")"
		if [ ! -f "$dst" ]; then
			cp -f "$f" "$dst" 2>/dev/null || true
		fi
	done
done

exit 0
