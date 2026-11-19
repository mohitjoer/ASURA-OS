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
put shell-end4 /usr/share/asura/shell-end4
rm -f "$ROOT/usr/share/asura/shell-end4/asura/vendor.sh" \
	"$ROOT/usr/share/asura/shell-end4/asura/hyprlang.py" \
	"$ROOT/usr/share/asura/shell-end4/asura/manifest.txt"
put session/hyprland /usr/share/asura/hyprland
put apps /usr/share/asura/apps
put_as apps/mimeapps/asura-mimeapps.list /usr/share/applications/asura-mimeapps.list
put shell/agent/skills /usr/share/asura/skills
put shell/sysmon /usr/share/asura/sysmon
put desktop/catalog /usr/share/asura/store/catalog
put system/runit/user-services /usr/share/asura/runit/services
put migrations /usr/share/asura/migrations
find "$ROOT/usr/share/asura/migrations" -type f -name '*.sh' -exec chmod 0755 {} + 2>/dev/null || true

# ── The session ─────────────────────────────────────────────────────────────
log "session files and scripts"
for f in asura-session asura-activate; do
	put_as "session/$f" "/usr/share/asura/session/$f"
	put_bin "session/$f" "/usr/bin/$f"
done
put_as session/asura.desktop /usr/share/asura/session/asura.desktop
put_as session/asura.desktop /usr/share/wayland-sessions/asura.desktop
put_as session/binds.json /usr/share/asura/session/binds.json
put_as session/runit/asura-user-services /usr/share/asura/session/asura-user-services
put_bin session/runit/asura-user-services /usr/bin/asura-user-services
put_as session/asura-store /usr/share/asura/session/asura-store
put_as session/asura-portals.conf /usr/share/xdg-desktop-portal/asura-portals.conf
put_as session/asura-portals.conf /usr/share/asura/session/asura-portals.conf
put session/units /usr/share/asura/units
put session/units /usr/share/asura/session/units
put session/units /usr/lib/systemd/user
put session/grub /usr/share/asura/session/grub
put session/sddm /usr/share/asura/session/sddm
for s in asura-hub asura-idle asura-pin asura-reload-cover asura-shot asura-vm asura-wallpaper-ui asura-welcome asura-greeter; do
	put_bin "session/$s" "/usr/bin/$s"
done
for f in "$REPO"/session/scripts/* "$REPO"/shell/scripts/*; do
	[ -f "$f" ] && put_bin "${f#"$REPO"/}" "/usr/bin/$(basename "$f")"
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
for h in prime-run fan-profile systemctl journalctl asura-gpu with-mimalloc; do
	put_bin "system/helpers/$h" "/usr/bin/$h"
done
put_bin installation/asura-installer.sh /usr/bin/asura-installer
put_bin desktop/cmd/asura-vm/asura-vmd /usr/bin/asura-vmd

# ── Settings schemas, branding, fonts, wallpapers ───────────────────────────
log "schemas, branding, fonts, wallpapers"
mkdir -p "$ROOT/usr/share/glib-2.0/schemas"
for f in "$REPO"/shell/org.asura.*.gschema.xml; do
	[ -f "$f" ] && cp -f "$f" "$ROOT/usr/share/glib-2.0/schemas/"
done
put public/branding /usr/share/asura/branding
put_as public/asura-logo.svg /usr/share/asura/branding/asura-logo.svg
put_as public/asura-logo-on-dark.svg /usr/share/asura/branding/asura-logo-on-dark.svg
put_as public/branding/asura-logo.txt /usr/share/asura/branding/fastfetch-ascii.txt
put public/fonts /usr/share/fonts/TTF
put public/wallpapers /usr/share/asura/backgrounds
put system/plymouth/asura /usr/share/plymouth/themes/asura
put_as shell/assets/asura-store.svg /usr/share/icons/hicolor/scalable/apps/asura-store.svg
put_as shell/assets/asura-installer.svg /usr/share/icons/hicolor/scalable/apps/asura-installer.svg
put_as shell/assets/asura-store.png /usr/share/icons/hicolor/512x512/apps/asura-store.png

# ── System configuration ────────────────────────────────────────────────────
log "system configuration"
for f in "$REPO"/system/udev/*.rules; do
	[ -f "$f" ] && put_as "${f#"$REPO"/}" "/etc/udev/rules.d/$(basename "$f")"
done
put system/modprobe.d /etc/modprobe.d
put system/modules-load.d /etc/modules-load.d
put system/tlp.d /etc/tlp.d
put system/sysctl.d /etc/sysctl.d
put_as system/xbps.d/asura.conf /etc/xbps.d/asura.conf
put_as system/xbps.d/asura.conf /etc/xbps.d/00-repository-main.conf
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
			go build -o "$out/" ./cmd/asura ./cmd/asura-sysd ./cmd/asura-userd ./cmd/asura-hubd \
				./cmd/asura-shell ./cmd/asura-wallpaper ./cmd/apm ./cmd/asura-store ./cmd/asura-sysmon
			go build -o "$out/asura-vm-mon" ./cmd/asura-vm/mon
			go build -o "$out/asura-vm-ssh" ./cmd/asura-vm/remote
		) && for b in "$out"/*; do
			# Replace by rename: a running daemon keeps its old binary until it restarts.
			install -m 0755 "$b" "$ROOT/usr/bin/.$(basename "$b").new"
			mv -f "$ROOT/usr/bin/.$(basename "$b").new" "$ROOT/usr/bin/$(basename "$b")"
		done || log "the Go build failed; the installed tools are unchanged"
		rm -rf "$out"
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

# ── Which commit this system now runs ───────────────────────────────────────
# asura update compares it with the branch: nothing new, nothing rebuilt.
if commit=$(git -c safe.directory='*' -C "$REPO" rev-parse HEAD 2>/dev/null); then
	printf '%s\n' "$commit" >"$ROOT/usr/share/asura/applied-commit"
	# And where it came from: asura update follows that checkout next time
	# instead of searching for one.
	printf '%s\n' "$REPO" >"$ROOT/usr/share/asura/applied-source"
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
