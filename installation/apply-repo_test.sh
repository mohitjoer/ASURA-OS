#!/bin/sh
# Tests for installation/apply-repo.sh.
#
# 1. Coverage: every repository path build-iso-void.sh installs from must be
#    installed by apply-repo.sh too, or `toren update` would leave that part of
#    an installed system behind. Paths only the live ISO needs are listed in
#    LIVE_ONLY with the reason.
# 2. A real run into a temporary root puts the key files where the ISO does.
set -eu

REPO_ROOT=$(CDPATH='' cd -- "$(dirname -- "$0")/.." && pwd)
APPLY="$REPO_ROOT/installation/apply-repo.sh"
BUILD="$REPO_ROOT/installation/build-iso-void.sh"

sh -n "$APPLY"

# Sources apply-repo.sh does not need to name:
# - mklive's boot splash data and build tooling: the live ISO's.
# - ../toren-store: the ISO may build that tool from
#   its standalone repository; apply-repo.sh builds it from ASURA's own
#   copy (desktop/cmd/toren-store), which store-sync keeps byte-identical.
LIVE_ONLY="
public/branding/boot-splash.png
installation/binary-checksums.txt
../toren-store
"

fail=0
for src in $(grep -oE '"\$REPO_ROOT/[A-Za-z0-9._/-]+' "$BUILD" | sed 's#^"\$REPO_ROOT/##' | sort -u); do
	case " $(echo $LIVE_ONLY) " in *" $src "*) continue ;; esac
	# Covered when apply-repo.sh names the path, or a directory above it.
	p=$src
	covered=0
	while [ -n "$p" ] && [ "$p" != "." ]; do
		if grep -qF "$p" "$APPLY"; then covered=1; break; fi
		p=$(dirname "$p")
	done
	if [ "$covered" -ne 1 ]; then
		echo "apply-repo_test: build-iso-void.sh installs $src, apply-repo.sh does not" >&2
		fail=1
	fi
done
[ "$fail" -eq 0 ] || exit 1

T=$(mktemp -d)
trap 'rm -rf "$T"' EXIT
# A file left by an older Toren (removed from the repository since) is
# deleted on update; a symlink the ISO adds stays.
mkdir -p "$T/root/usr/share/toren/shell-desktop/modules/settings"
echo old >"$T/root/usr/share/toren/shell-desktop/settings.qml"
echo old >"$T/root/usr/share/toren/shell-desktop/modules/settings/BarConfig.qml"
ln -s toren "$T/root/usr/share/toren/shell-desktop/asura"
sh "$APPLY" --no-packages --no-build "$REPO_ROOT" "$T/root" >"$T/log" 2>&1 || { cat "$T/log"; exit 1; }
for f in \
	usr/share/toren/shell/shell.qml \
	usr/share/toren/shell-desktop/shell.qml \
	usr/share/toren/shell-desktop/toren/TorenBridge.qml \
	usr/share/toren/shell-desktop/matugen/config.toml \
	usr/share/toren/hyprland/binds.conf \
	usr/share/toren/runit/services/quickshell/run \
	usr/share/toren/session/binds.json \
	usr/share/toren/hardware/models/generic-laptop.yaml \
	usr/share/glib-2.0/schemas/org.toren.shell.gschema.xml \
	usr/bin/toren-session \
	usr/bin/toren-installer \
	etc/nftables.conf; do
	[ -e "$T/root/$f" ] || { echo "apply-repo_test: $f was not installed" >&2; exit 1; }
done
for f in shell-desktop/settings.qml shell-desktop/modules/settings; do
	[ ! -e "$T/root/usr/share/toren/$f" ] || { echo "apply-repo_test: the removed $f is still installed" >&2; exit 1; }
done
[ -L "$T/root/usr/share/toren/shell-desktop/asura" ] || { echo "apply-repo_test: the ISO's asura symlink was removed" >&2; exit 1; }
# The desktop build tooling stays out of the system.
[ ! -e "$T/root/usr/share/toren/shell-desktop/toren/vendor.sh" ] || { echo "apply-repo_test: vendor.sh was installed" >&2; exit 1; }
[ -x "$T/root/usr/bin/toren-session" ] || { echo "apply-repo_test: toren-session is not executable" >&2; exit 1; }
echo "apply-repo_test: ok"
