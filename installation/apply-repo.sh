#!/bin/sh
# Apply an ASURA OS checkout to a running system: everything the ISO build
# installs from the repository, installed the same way, so a change merged into
# dev-unstable reaches an installed machine through `toren update`.
#
# Usage: installation/apply-repo.sh [--no-packages] [--no-build] REPO [ROOT]
#   REPO           an TOREN-OS checkout (toren update passes the one it pulled)
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
# put_mirror SRC DEST-DIR: put, then delete the files under DEST-DIR that SRC
# no longer has, so a file removed from the repository (the desktop's old
# Settings app, say) does not stay on an updated system and keep running.
# Only regular files are removed (a symlink the ISO adds, like
# shell-desktop/asura, stays), then the directories left empty.
put_mirror() {
	put "$1" "$2"
	_src="$REPO/$1"
	_dst="$ROOT$2"
	[ -d "$_src" ] && [ -d "$_dst" ] || return 0
	(cd "$_dst" && find . -type f) | while IFS= read -r _f; do
		[ -e "$_src/$_f" ] || [ -L "$_src/$_f" ] || rm -f "$_dst/$_f"
	done
	find "$_dst" -mindepth 1 -depth -type d -empty -delete 2>/dev/null || true
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
put_mirror shell /usr/share/toren/shell
put_mirror shell-desktop /usr/share/toren/shell-desktop
rm -f "$ROOT/usr/share/toren/shell-desktop/toren/vendor.sh" \
	"$ROOT/usr/share/toren/shell-desktop/toren/manifest.txt"
put session/hyprland /usr/share/toren/hyprland
put session/safe-mode /usr/share/toren/safe-mode
put apps /usr/share/toren/apps
put_as apps/mimeapps/toren-mimeapps.list /usr/share/applications/toren-mimeapps.list
put shell/agent/skills /usr/share/toren/skills
put docs/manual /usr/share/doc/toren/manual
# One-time fixes for installed machines; toren update runs the new ones
# (toren migrate). Marked done only on a fresh install, never here.
put migrations /usr/share/toren/migrations
for f in "$ROOT"/usr/share/toren/migrations/*/*.sh; do
	[ -f "$f" ] && chmod 0755 "$f"
done
put desktop/catalog /usr/share/toren/store/catalog
put system/runit/user-services /usr/share/toren/runit/services
put system/hardware/models /usr/share/toren/hardware/models
put system/hardware/components /usr/share/toren/hardware/components

# ── The session ─────────────────────────────────────────────────────────────
log "session files and scripts"
for f in toren-session toren-session-safe toren-activate; do
	put_as "session/$f" "/usr/share/toren/session/$f"
	put_bin "session/$f" "/usr/bin/$f"
done
put_as session/toren.desktop /usr/share/toren/session/toren.desktop
put_as session/toren.desktop /usr/share/wayland-sessions/toren.desktop
put_as session/toren-safe.desktop /usr/share/toren/session/toren-safe.desktop
put_as session/toren-safe.desktop /usr/share/wayland-sessions/toren-safe.desktop
put_as session/binds.json /usr/share/toren/session/binds.json
put_as session/runit/toren-user-services /usr/share/toren/session/toren-user-services
put_as session/toren-graphics-env /usr/share/toren/session/toren-graphics-env
put_bin session/runit/toren-user-services /usr/bin/toren-user-services
put_as session/toren-store /usr/share/toren/session/toren-store
put_as session/toren-portals.conf /usr/share/xdg-desktop-portal/toren-portals.conf
put_as session/toren-portals.conf /usr/share/toren/session/toren-portals.conf
put session/units /usr/share/toren/units
put session/units /usr/share/toren/session/units
put session/units /usr/lib/systemd/user
put session/grub /usr/share/toren/session/grub
put session/sddm /usr/share/toren/session/sddm
for s in toren-hub toren-idle toren-installer-ui toren-pin toren-reload-cover toren-shot toren-vm toren-wallpaper-ui toren-welcome toren-greeter; do
	put_bin "session/$s" "/usr/bin/$s"
	put_as "session/$s" "/usr/share/toren/session/$s"
done
for f in "$REPO"/session/scripts/* "$REPO"/shell/scripts/*; do
	case "$f" in *.md) continue ;; esac
	[ -f "$f" ] || continue
	b=$(basename "$f")
	put_bin "${f#"$REPO"/}" "/usr/bin/$b"
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
for h in prime-run fan-profile systemctl journalctl toren-gpu with-mimalloc toren-charge-limit-apply; do
	put_bin "system/helpers/$h" "/usr/bin/$h"
done
put_bin installation/toren-installer.sh /usr/bin/toren-installer
put_bin desktop/cmd/toren-vm/toren-vmd /usr/bin/toren-vmd

# ── Settings schemas, branding, fonts, wallpapers ───────────────────────────
log "schemas, branding, fonts, wallpapers"
mkdir -p "$ROOT/usr/share/glib-2.0/schemas"
for f in "$REPO"/shell/org.toren.*.gschema.xml; do
	[ -f "$f" ] && cp -f "$f" "$ROOT/usr/share/glib-2.0/schemas/"
done
put public/branding /usr/share/toren/branding
put_as public/toren-logo.svg /usr/share/toren/branding/toren-logo.svg
put_as public/toren-logo-on-dark.svg /usr/share/toren/branding/toren-logo-on-dark.svg
put_as public/branding/toren-logo.txt /usr/share/toren/branding/fastfetch-ascii.txt
put public/fonts /usr/share/fonts/TTF
put public/wallpapers /usr/share/toren/backgrounds
put system/plymouth/toren /usr/share/plymouth/themes/toren
put_as shell/assets/toren-store.svg /usr/share/icons/hicolor/scalable/apps/toren-store.svg
put_as shell/assets/toren-installer.svg /usr/share/icons/hicolor/scalable/apps/toren-installer.svg
put_as shell/assets/toren-store.png /usr/share/icons/hicolor/512x512/apps/toren-store.png
put_as shell-desktop/assets/icons/toren-settings.svg /usr/share/icons/hicolor/scalable/apps/toren-settings.svg
put_as shell-desktop/assets/icons/toren-vm.svg /usr/share/icons/hicolor/scalable/apps/toren-vm.svg
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
put_as system/xbps.d/toren.conf /etc/xbps.d/toren.conf
# thermald starts only where it can work (system/runit/sv-conf/thermald/conf)
put_as system/runit/sv-conf/thermald/conf /etc/sv/thermald/conf
put_as system/xbps.d/toren.conf /etc/xbps.d/00-repository-main.conf
# Toren Linux repository signing key (installed only where missing is handled
# by the migration; here we always refresh from the checkout).
put system/xbps-keys /usr/share/toren/xbps/keys
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
		for f in "$ROOT/usr/share/toren/session"/toren-*; do
			[ -e "$f" ] || continue
			base=$(basename "$f")
			if [ -x "$f" ] && [ ! -e "$ROOT/usr/bin/$base" ]; then
				ln -sf "/usr/share/toren/session/$base" "$ROOT/usr/bin/$base"
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
	# its conf keeps it down on machines without Intel coretemp sensors
	if [ -d /etc/sv/thermald ] && [ ! -e /etc/runit/runsvdir/default/thermald ]; then
		ln -sf /etc/sv/thermald /etc/runit/runsvdir/default/thermald
	fi
fi

# ── Which commit this system now runs ───────────────────────────────────────
# toren update compares it with the branch: nothing new, nothing rebuilt.
if commit=$(git -c safe.directory='*' -C "$REPO" rev-parse HEAD 2>/dev/null); then
	printf '%s\n' "$commit" >"$ROOT/usr/share/toren/applied-commit"
	# And where it came from: toren update follows that checkout next time
	# instead of searching for one.
	printf '%s\n' "$REPO" >"$ROOT/usr/share/toren/applied-source"
fi

# ── Make the system notice ──────────────────────────────────────────────────
if [ "$ROOT" = "" ]; then
	command -v glib-compile-schemas >/dev/null 2>&1 && glib-compile-schemas /usr/share/glib-2.0/schemas
	command -v udevadm >/dev/null 2>&1 && udevadm control --reload-rules 2>/dev/null || true
	command -v fc-cache >/dev/null 2>&1 && fc-cache -f >/dev/null 2>&1 || true
else
	command -v glib-compile-schemas >/dev/null 2>&1 && glib-compile-schemas "$ROOT/usr/share/glib-2.0/schemas" || true
fi
log "done. toren update now refreshes the running desktop; or run: toren refresh"
