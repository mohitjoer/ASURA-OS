#!/bin/sh
# Tests for installation/apply-repo.sh.
#
# 1. Coverage: every repository path build-iso-void.sh installs from must be
#    installed by apply-repo.sh too, or `asura update` would leave that part of
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
# - ../asura-store, ../asura-sysmon: the ISO may build those two tools from
#   their standalone repositories; apply-repo.sh builds them from ASURA's own
#   copies (desktop/cmd/asura-store, asura-sysmon), which store-sync and
#   sysmon-sync keep byte-identical.
LIVE_ONLY="
public/branding/boot-splash.png
installation/binary-checksums.txt
../asura-store
../asura-sysmon
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
sh "$APPLY" --no-packages --no-build "$REPO_ROOT" "$T/root" >"$T/log" 2>&1 || { cat "$T/log"; exit 1; }
for f in \
	usr/share/asura/shell/shell.qml \
	usr/share/asura/shell-end4/shell.qml \
	usr/share/asura/shell-end4/asura/AsuraBridge.qml \
	usr/share/asura/shell-end4/matugen/config.toml \
	usr/share/asura/hyprland/binds.conf \
	usr/share/asura/runit/services/quickshell/run \
	usr/share/asura/session/binds.json \
	usr/share/glib-2.0/schemas/org.asura.shell.gschema.xml \
	usr/bin/asura-session \
	usr/bin/asura-installer \
	usr/share/asura/migrations/system/1795180000.sh \
	etc/nftables.conf; do
	[ -e "$T/root/$f" ] || { echo "apply-repo_test: $f was not installed" >&2; exit 1; }
done
# The end-4 build tooling stays out of the system.
[ ! -e "$T/root/usr/share/asura/shell-end4/asura/vendor.sh" ] || { echo "apply-repo_test: vendor.sh was installed" >&2; exit 1; }
[ -x "$T/root/usr/bin/asura-session" ] || { echo "apply-repo_test: asura-session is not executable" >&2; exit 1; }
[ -x "$T/root/usr/share/asura/migrations/system/1795180000.sh" ] || { echo "apply-repo_test: migration script is not executable" >&2; exit 1; }
echo "apply-repo_test: ok"
