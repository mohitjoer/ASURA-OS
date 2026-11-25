#!/bin/sh
# Apply an ASURA OS checkout to a running system: everything the ISO build
# installs from the repository, installed the same way, so a change merged into
# dev-unstable reaches an installed machine through `asura update`.
#
# Usage: installation/apply-repo.sh [--no-packages] [--no-build] REPO [ROOT]
#   REPO           an ASURA-OS checkout (asura update passes the one it pulled)
#   ROOT           the system to apply to (default /); a directory for tests
#   --no-packages  do not install packages new in system/packages-void.yaml
#   --no-build     do not rebuild the Go tools
#
# Runs as root. Mirrors the repository-to-system copies in
# installation/build-iso-void.sh; installation/apply-repo_test.sh fails if the
# build copies something from the repository that this script does not, so
# the two cannot drift apart. Live-ISO-only steps (autologin, the installer's
# desktop icons in skel, mklive) are not here.
set -eu

PACKAGES=1
BUILD=1
while [ $# -gt 0 ]; do
	case "$1" in
		--no-packages) PACKAGES=0; shift ;;
		--no-build) BUILD=0; shift ;;
		-*) echo "apply-repo: unknown option $1" >&2; exit 2 ;;
		*) break ;;
	esac
done
REPO=${1:?usage: apply-repo.sh [--no-packages] [--no-build] REPO [ROOT]}
ROOT=${2:-/}
REPO=$(CDPATH='' cd -- "$REPO" && pwd)
ROOT=${ROOT%/}

log() { printf 'apply-repo: %s\n' "$*"; }

# copy SRC DEST-DIR: a file or a directory's contents, created as needed.
put() {
	_src="$REPO/$1"
	_dst="$ROOT$2"
	[ -e "$_src" ] || return 0
	mkdir -p "$_dst"
	if [ -d "$_src" ]; then
		cp -rf "$_src"/. "$_dst"/
	else
		cp -f "$_src" "$_dst"/
	fi
}
# copy SRC to an exact DEST path.
put_as() {
	[ -e "$REPO/$1" ] || return 0
	mkdir -p "$(dirname "$ROOT$2")"
	cp -f "$REPO/$1" "$ROOT$2"
}
# install SRC as an executable at DEST path.
put_bin() {
	[ -f "$REPO/$1" ] || return 0
	mkdir -p "$(dirname "$ROOT$2")"
	install -m 0755 "$REPO/$1" "$ROOT$2"
}

# ── The desktop ─────────────────────────────────────────────────────────────
log "desktop shells, Hyprland, apps"
put shell /usr/share/asura/shell
put shell-desktop /usr/share/asura/shell-desktop
rm -f "$ROOT/usr/share/asura/shell-desktop/toren/vendor.sh" \
	"$ROOT/usr/share/asura/shell-desktop/toren/manifest.txt"
ln -sfn toren "$ROOT/usr/share/asura/shell-desktop/asura"
put session/hyprland /usr/share/asura/hyprland
put session/safe-mode /usr/share/asura/safe-mode
put apps /usr/share/asura/apps
put_as apps/mimeapps/toren-mimeapps.list /usr/share/applications/toren-mimeapps.list
ln -sf toren-mimeapps.list "$ROOT/usr/share/applications/asura-mimeapps.list"
put shell/agent/skills /usr/share/asura/skills
put docs/manual /usr/share/doc/asura/manual
# One-time fixes for installed machines; asura update runs the new ones
# (asura migrate). Marked done only on a fresh install, never here.
put migrations /usr/share/asura/migrations
for f in "$ROOT"/usr/share/asura/migrations/*/*.sh; do
	[ -f "$f" ] && chmod 0755 "$f"
done
put desktop/catalog /usr/share/asura/store/catalog
put system/runit/user-services /usr/share/asura/runit/services
put system/hardware/models /usr/share/asura/hardware/models
put system/hardware/components /usr/share/asura/hardware/components

# ── The session ─────────────────────────────────────────────────────────────
log "session files and scripts"
for f in toren-session toren-session-safe toren-activate; do
	af=$(echo "$f" | sed 's/^toren/asura/')
	put_as "session/$f" "/usr/share/asura/session/$f"
	ln -sf "$f" "$ROOT/usr/share/asura/session/$af"
	put_bin "session/$f" "/usr/bin/$f"
	ln -sf "$f" "$ROOT/usr/bin/$af"
done
put_as session/toren.desktop /usr/share/asura/session/toren.desktop
put_as session/toren.desktop /usr/share/wayland-sessions/toren.desktop
put_as session/toren-safe.desktop /usr/share/asura/session/toren-safe.desktop
put_as session/toren-safe.desktop /usr/share/wayland-sessions/toren-safe.desktop
put_as session/binds.json /usr/share/asura/session/binds.json
put_as session/runit/toren-user-services /usr/share/asura/session/toren-user-services
ln -sf toren-user-services "$ROOT/usr/share/asura/session/asura-user-services"
put_bin session/runit/toren-user-services /usr/bin/toren-user-services
ln -sf toren-user-services "$ROOT/usr/bin/asura-user-services"
put_as session/toren-store /usr/share/asura/session/toren-store
ln -sf toren-store "$ROOT/usr/share/asura/session/asura-store"
put_as session/toren-portals.conf /usr/share/xdg-desktop-portal/asura-portals.conf
put_as session/toren-portals.conf /usr/share/asura/session/toren-portals.conf
ln -sf toren-portals.conf "$ROOT/usr/share/asura/session/asura-portals.conf"
put session/units /usr/share/asura/units
put session/units /usr/share/asura/session/units
put session/units /usr/lib/systemd/user
put session/grub /usr/share/asura/session/grub
put session/sddm /usr/share/asura/session/sddm
for s in toren-hub toren-idle toren-pin toren-reload-cover toren-shot toren-vm toren-wallpaper-ui toren-welcome toren-greeter; do
	as=$(echo "$s" | sed 's/^toren/asura/')
	put_bin "session/$s" "/usr/bin/$s"
	ln -sf "$s" "$ROOT/usr/bin/$as"
	put_as "session/$s" "/usr/share/asura/session/$s"
	ln -sf "$s" "$ROOT/usr/share/asura/session/$as"
done
for f in "$REPO"/session/scripts/* "$REPO"/shell/scripts/*; do
	case "$f" in *.md) continue ;; esac
	[ -f "$f" ] || continue
	b=$(basename "$f")
	put_bin "${f#"$REPO"/}" "/usr/bin/$b"
	case "$b" in toren-*) ln -sf "$b" "$ROOT/usr/bin/$(echo "$b" | sed 's/^toren/asura/')" ;; esac
done
put session/applications /usr/share/applications
put_as session/turnstile/turnstile.conf /etc/turnstile/turnstile.conf

# ── Helpers and the installer ───────────────────────────────────────────────
log "helpers"
for hdir in audio display input perf pkg; do
	for f in "$REPO/system/helpers/$hdir"/*; do
		[ -f "$f" ] || continue
		case "$f" in *.sh) continue ;; esac
		put_bin "${f#"$REPO"/}" "/usr/bin/$(basename "$f")"
	done
done
for h in prime-run fan-profile systemctl journalctl toren-gpu with-mimalloc; do
	put_bin "system/helpers/$h" "/usr/bin/$h"
done
ln -sf toren-gpu "$ROOT/usr/bin/asura-gpu"
put_bin installation/toren-installer.sh /usr/bin/toren-installer
ln -sf toren-installer "$ROOT/usr/bin/asura-installer"
put_bin desktop/cmd/toren-vm/toren-vmd /usr/bin/toren-vmd
ln -sf toren-vmd "$ROOT/usr/bin/asura-vmd"

# ── Settings schemas, branding, fonts, wallpapers ───────────────────────────
log "schemas, branding, fonts, wallpapers"
mkdir -p "$ROOT/usr/share/glib-2.0/schemas"
for f in "$REPO"/shell/org.asura.*.gschema.xml; do
	[ -f "$f" ] && cp -f "$f" "$ROOT/usr/share/glib-2.0/schemas/"
done
put public/branding /usr/share/asura/branding
put_as public/toren-logo.svg /usr/share/asura/branding/asura-logo.svg
put_as public/toren-logo-on-dark.svg /usr/share/asura/branding/asura-logo-on-dark.svg
put_as public/branding/toren-logo.txt /usr/share/asura/branding/fastfetch-ascii.txt
put public/fonts /usr/share/fonts/TTF
put public/wallpapers /usr/share/asura/backgrounds
put system/plymouth/toren /usr/share/plymouth/themes/asura
put_as shell/assets/toren-store.svg /usr/share/icons/hicolor/scalable/apps/asura-store.svg
put_as shell/assets/toren-installer.svg /usr/share/icons/hicolor/scalable/apps/asura-installer.svg
put_as shell/assets/toren-store.png /usr/share/icons/hicolor/512x512/apps/asura-store.png
put_as shell-desktop/assets/icons/asura-settings.svg /usr/share/icons/hicolor/scalable/apps/asura-settings.svg
put_as shell-desktop/assets/icons/asura-vm.svg /usr/share/icons/hicolor/scalable/apps/asura-vm.svg
put_as shell-desktop/assets/icons/preferences-system.svg /usr/share/icons/hicolor/scalable/apps/preferences-system.svg
put_as shell-desktop/assets/icons/preferences-system-network.svg /usr/share/icons/hicolor/scalable/apps/preferences-system-network.svg
put_as shell-desktop/assets/icons/computer.svg /usr/share/icons/hicolor/scalable/apps/computer.svg
put_as shell-desktop/assets/icons/utilities-terminal.svg /usr/share/icons/hicolor/scalable/apps/utilities-terminal.svg
put_as shell-desktop/assets/icons/utilities-system-monitor.svg /usr/share/icons/hicolor/scalable/apps/utilities-system-monitor.svg
put_as shell-desktop/assets/icons/opencode.svg /usr/share/icons/hicolor/scalable/apps/opencode.svg
gtk-update-icon-cache -qtf /usr/share/icons/hicolor 2>/dev/null || true

# ── System configuration ────────────────────────────────────────────────────
log "system configuration"
for f in "$REPO"/system/udev/*.rules; do
	[ -f "$f" ] && put_as "${f#"$REPO"/}" "/etc/udev/rules.d/$(basename "$f")"
done
put system/modprobe.d /etc/modprobe.d
put system/modules-load.d /etc/modules-load.d
put system/tlp.d /etc/tlp.d
put system/sysctl.d /etc/sysctl.d
put_as system/xbps.d/toren.conf /etc/xbps.d/asura.conf
put_as system/xbps.d/toren.conf /etc/xbps.d/00-repository-main.conf
# Toren Linux repository signing key (installed only where missing is handled
# by the migration; here we always refresh from the checkout).
put system/xbps-keys /usr/share/asura/xbps/keys
put system/xbps-keys /var/db/xbps/keys
put_as system/nftables/nftables.conf /etc/nftables.conf
put system/ananicy /etc/ananicy-cpp
for svc in "$REPO"/system/runit/services/*; do
	[ -d "$svc" ] || continue
	name=$(basename "$svc")
	put "system/runit/services/$name" "/etc/sv/$name"
	[ -f "$ROOT/etc/sv/$name/run" ] && chmod +x "$ROOT/etc/sv/$name/run"
done

# ── The Go tools ────────────────────────────────────────────────────────────
if [ "$BUILD" -eq 1 ]; then
	if command -v go >/dev/null 2>&1; then
		log "building the Go tools"
		out=$(mktemp -d)
		(
			cd "$REPO/desktop"
			export GOWORK=off GOFLAGS=-buildvcs=false CGO_ENABLED=0
			go build -o "$out/" ./cmd/toren ./cmd/toren-sysd ./cmd/toren-userd ./cmd/toren-hubd \
				./cmd/toren-shell ./cmd/toren-wallpaper ./cmd/apm ./cmd/toren-store ./cmd/toren-agent
			go build -o "$out/toren-vm-mon" ./cmd/toren-vm/mon
			go build -o "$out/toren-vm-ssh" ./cmd/toren-vm/remote
		) && for b in "$out"/*; do
			# Replace by rename: a running daemon keeps its old binary until it restarts.
			install -m 0755 "$b" "$ROOT/usr/bin/.$(basename "$b").new"
			mv -f "$ROOT/usr/bin/.$(basename "$b").new" "$ROOT/usr/bin/$(basename "$b")"
		done || log "the Go build failed; the installed tools are unchanged"
		rm -rf "$out"
		for tool in toren toren-store toren-shell toren-wallpaper toren-hubd toren-sysd toren-userd toren-agent toren-vm toren-vm-mon toren-vm-ssh toren-vmd; do
			asura_tool=$(echo "$tool" | sed 's/^toren/asura/')
			if [ -e "$ROOT/usr/bin/$tool" ]; then
				ln -sf "$tool" "$ROOT/usr/bin/$asura_tool"
			fi
		done
		for f in "$ROOT/usr/share/asura/session"/toren-*; do
			[ -e "$f" ] || continue
			base=$(basename "$f")
			asura_base=$(echo "$base" | sed 's/^toren/asura/')
			ln -sf "$base" "$ROOT/usr/share/asura/session/$asura_base"
			if [ -x "$f" ] && [ ! -e "$ROOT/usr/bin/$asura_base" ]; then
				ln -sf "/usr/share/asura/session/$base" "$ROOT/usr/bin/$asura_base"
			fi
		done
	else
		log "go is not installed; the Go tools are unchanged"
	fi
fi

# ── Packages new in the manifest ────────────────────────────────────────────
if [ "$PACKAGES" -eq 1 ] && [ "$ROOT" = "" ] && command -v xbps-query >/dev/null 2>&1; then
	missing=""
	for p in $(sed -n 's/^[[:space:]]\{4\}- \([A-Za-z0-9][A-Za-z0-9._+-]*\)[[:space:]]*$/\1/p' "$REPO/system/packages-void.yaml"); do
		xbps-query "$p" >/dev/null 2>&1 || missing="$missing $p"
	done
	if [ -n "$missing" ]; then
		log "installing packages new in the manifest:$missing"
		# shellcheck disable=SC2086
		xbps-install -Sy $missing || log "some packages could not be installed:$missing"
	fi
fi

# ── Beta system services ─────────────────────────────────────────────────────
if [ "$ROOT" = "" ]; then
	mkdir -p /etc/runit/runsvdir/default
	for svc in socklog-unix nanoklogd earlyoom toren-maintenance toren-sysd toren-boot-guard; do
		if [ -d "/etc/sv/$svc" ] && [ ! -e "/etc/runit/runsvdir/default/$svc" ]; then
			ln -sf "/etc/sv/$svc" "/etc/runit/runsvdir/default/$svc"
		fi
	done
	if grep -qi "GenuineIntel\|Intel" /proc/cpuinfo 2>/dev/null; then
		if [ -d /etc/sv/thermald ] && [ ! -e /etc/runit/runsvdir/default/thermald ]; then
			ln -sf /etc/sv/thermald /etc/runit/runsvdir/default/thermald
		fi
	fi
fi

# ── Which commit this system now runs ───────────────────────────────────────
# asura update compares it with the branch: nothing new, nothing rebuilt.
if commit=$(git -c safe.directory='*' -C "$REPO" rev-parse HEAD 2>/dev/null); then
	printf '%s\n' "$commit" >"$ROOT/usr/share/asura/applied-commit"
	# And where it came from: asura update follows that checkout next time
	# instead of searching for one.
	printf '%s\n' "$REPO" >"$ROOT/usr/share/asura/applied-source"
fi

# Transitional symlink for /usr/share/toren <-> /usr/share/asura
if [ -d "$ROOT/usr/share/asura" ] && [ ! -e "$ROOT/usr/share/toren" ]; then
	ln -sfn asura "$ROOT/usr/share/toren"
fi

# ── Make the system notice ──────────────────────────────────────────────────
if [ "$ROOT" = "" ]; then
	command -v glib-compile-schemas >/dev/null 2>&1 && glib-compile-schemas /usr/share/glib-2.0/schemas
	command -v udevadm >/dev/null 2>&1 && udevadm control --reload-rules 2>/dev/null || true
	command -v fc-cache >/dev/null 2>&1 && fc-cache -f >/dev/null 2>&1 || true
else
	command -v glib-compile-schemas >/dev/null 2>&1 && glib-compile-schemas "$ROOT/usr/share/glib-2.0/schemas" || true
fi
log "done. asura update now refreshes the running desktop; or run: asura refresh"
