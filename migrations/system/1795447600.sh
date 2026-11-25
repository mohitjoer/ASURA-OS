#!/bin/sh
echo "Add ASURA package repository to /etc/xbps.d and remove third-party Hyprland repository"

# Update /etc/xbps.d to place ASURA's repository first and remove the old third-party
# Hyprland repository.
CONF_DIR="/etc/xbps.d"
mkdir -p "$CONF_DIR"

# Clean out third-party repository references from any conf files in /etc/xbps.d
for conf in "$CONF_DIR"/*.conf; do
	[ -f "$conf" ] || continue
	if grep -Fq 'hyprland-void' "$conf"; then
		sed -i '/hyprland-void/d' "$conf"
	fi
done

# Ensure ASURA repository line is present in /etc/xbps.d/asura.conf
ASURA_CONF="$CONF_DIR/asura.conf"
REPO_LINE="repository=https://mohitjoer.github.io/ASURA-OS/current"

if [ -f "$ASURA_CONF" ]; then
	if ! grep -Fq "$REPO_LINE" "$ASURA_CONF"; then
		# Prepend the ASURA repo line
		tmpfile=$(mktemp)
		{
			echo "$REPO_LINE"
			cat "$ASURA_CONF"
		} > "$tmpfile" && mv -f "$tmpfile" "$ASURA_CONF"
	fi
else
	cat >"$ASURA_CONF" <<EOF
# Asura OS Package Repositories
channel=stable

$REPO_LINE
repository=https://repo-fastly.voidlinux.org/current
repository=https://repo-fastly.voidlinux.org/current/nonfree
repository=https://repo-fastly.voidlinux.org/current/multilib
repository=https://repo-fastly.voidlinux.org/current/multilib/nonfree
EOF
fi

# Ensure 00-repository-main.conf points to asura.conf if needed
if [ ! -e "$CONF_DIR/00-repository-main.conf" ]; then
	ln -sf asura.conf "$CONF_DIR/00-repository-main.conf" 2>/dev/null || true
fi

# Remove third-party key if present
rm -f /var/db/xbps/keys/30:93:47:9e:1f:55:c5:c5:07:b1:9a:e6:8b:10:5b:4b.plist 2>/dev/null || true

# Install ASURA public signing key if present in the package or tree
for keydir in /usr/share/asura/xbps/keys /usr/share/asura/keys; do
	if [ -d "$keydir" ]; then
		mkdir -p /var/db/xbps/keys
		cp -f "$keydir"/*.plist /var/db/xbps/keys/ 2>/dev/null || true
	fi
done

exit 0
