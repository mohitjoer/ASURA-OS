#!/bin/sh
# Build the Hyprland stack packages from Toren-Linux/void-packages (branch Toren).
#
# Usage: packaging/xbps/build-stack.sh [OUT_DIR] [VOID_PACKAGES_DIR]
#
# Clones Toren-Linux/void-packages at the Toren branch (or uses VOID_PACKAGES_DIR if provided),
# builds the Hyprland stack packages with xbps-src, and collects them in OUT_DIR.
set -eu

OUT_DIR=${1:-build/repo/current}
VOID_PACKAGES_DIR=${2:-}

mkdir -p "$OUT_DIR"
OUT_DIR=$(CDPATH='' cd -- "$OUT_DIR" && pwd)

STACK_PACKAGES="
hyprutils
hyprwayland-scanner
hyprlang
hyprcursor
hyprgraphics
aquamarine
hyprland-protocols
hyprwire
hyprtoolkit
hyprland
xdg-desktop-portal-hyprland
hypridle
hyprlock
hyprpicker
"

# If a local void-packages directory wasn't provided, clone Toren-Linux/void-packages
WORK_DIR=""
if [ -z "$VOID_PACKAGES_DIR" ] || [ ! -d "$VOID_PACKAGES_DIR" ]; then
	WORK_DIR=$(mktemp -d -t asura-void-packages-XXXXXX)
	trap 'rm -rf "$WORK_DIR"' EXIT INT TERM
	echo "build-stack: cloning Toren-Linux/void-packages at branch Toren..."
	git clone --depth 1 -b Toren https://github.com/Toren-Linux/void-packages.git "$WORK_DIR"
	VP_DIR="$WORK_DIR"
else
	VP_DIR="$VOID_PACKAGES_DIR"
fi

echo "build-stack: preparing to build stack in $VP_DIR..."
cd "$VP_DIR"
[ -x ./xbps-src ] || { echo "build-stack: $VP_DIR has no xbps-src" >&2; exit 1; }
for tool in bash git; do
	command -v "$tool" >/dev/null 2>&1 || { echo "build-stack: $tool is not installed (xbps-install -y bash sudo git)" >&2; exit 1; }
done
if [ "$(id -u)" = 0 ] && ! command -v sudo >/dev/null 2>&1; then
	echo "build-stack: sudo is not installed (xbps-install -y sudo)" >&2; exit 1
fi

# Run xbps-src the way Void's own CI does (common/travis/prepare.sh): as a
# non-root user in the xbuilder group, with the uchroot chroot style, in a
# privileged container. xbps-src refuses to build as root.
cat >>etc/conf <<'CONF'
XBPS_BUILD_ENVIRONMENT=void-packages-ci
XBPS_ALLOW_RESTRICTED=yes
XBPS_CHROOT_CMD=uchroot
CONF
if [ "$(id -u)" = 0 ]; then
	id builder >/dev/null 2>&1 || useradd -G xbuilder -M builder
	chown -R builder:builder "$VP_DIR"
	as_builder() { (cd "$VP_DIR" && sudo -Eu builder "$@"); }
else
	as_builder() { (cd "$VP_DIR" && "$@"); }
fi

# Run a step with its output in a log; on failure print the log and, in
# GitHub Actions, an annotation with its last lines (annotations show on the
# run page without signing in; the log does not).
logged() {
	name=$1; shift
	log="$VP_DIR/build-$name.log"
	if ! "$@" >"$log" 2>&1; then
		cat "$log"
		echo "build-stack: $name failed" >&2
		if [ -n "${GITHUB_ACTIONS:-}" ]; then
			tail -n 25 "$log" | sed 's/%/%25/g' | awk -v n="$name" '
				{ msg = msg (NR > 1 ? "%0A" : "") $0 }
				END { print ""; print "::error title=build-stack failed on " n "::" msg }'
		fi
		exit 1
	fi
	tail -n 3 "$log"
}

echo "build-stack: binary-bootstrap..."
logged binary-bootstrap as_builder ./xbps-src binary-bootstrap

jobs=$(nproc 2>/dev/null || echo 2)
for pkg in $STACK_PACKAGES; do
	[ -d "srcpkgs/$pkg" ] || { echo "build-stack: srcpkgs/$pkg is missing" >&2; exit 1; }
	echo "build-stack: building $pkg..."
	logged "$pkg" as_builder ./xbps-src -j"$jobs" pkg "$pkg"
done

# xbps-src puts packages under hostdir/binpkgs/<branch>/.
find hostdir/binpkgs -name "*.xbps" -exec cp -f {} "$OUT_DIR/" \;
n=$(find "$OUT_DIR" -maxdepth 1 -name "*.xbps" | wc -l)
[ "$n" -gt 0 ] || { echo "build-stack: no packages were built" >&2; exit 1; }
echo "build-stack: $n package(s) in $OUT_DIR"

# Index repository if xbps-rindex exists
if command -v xbps-rindex >/dev/null 2>&1; then
	echo "build-stack: indexing repository at $OUT_DIR..."
	for pkg in "$OUT_DIR"/*.xbps; do
		[ -e "$pkg" ] || continue
		xbps-rindex -a "$pkg"
	done

	# Retain up to 3 versions per package name
	if command -v python3 >/dev/null 2>&1; then
		python3 - "$OUT_DIR" <<'PY'
import os, sys, glob, re

out_dir = sys.argv[1]
xbps_files = glob.glob(os.path.join(out_dir, "*.xbps"))
pkgs = {}
for f in xbps_files:
    fname = os.path.basename(f)
    m = re.match(r"^(.+)-([0-9].+)\.([^.]+)\.xbps$", fname)
    if m:
        name = m.group(1)
        pkgs.setdefault(name, []).append((os.path.getmtime(f), f))

for name, files in pkgs.items():
    if len(files) > 3:
        files.sort(key=lambda x: x[0], reverse=True)
        for _, old_f in files[3:]:
            try:
                os.remove(old_f)
                sig_f = old_f + ".sig2"
                if os.path.exists(sig_f):
                    os.remove(sig_f)
                print(f"pruned old package version: {os.path.basename(old_f)}")
            except OSError:
                pass
PY
	fi

	xbps-rindex -c "$OUT_DIR" 2>/dev/null || true
	xbps-rindex -r "$OUT_DIR" 2>/dev/null || true

	# Sign repository and packages if key is set
	if [ -n "${XBPS_SIGNING_KEY:-}" ]; then
		echo "build-stack: signing repository and packages..."
		xbps-rindex --sign --signedby "Toren Linux" --privkey "$XBPS_SIGNING_KEY" "$OUT_DIR"
		for pkg in "$OUT_DIR"/*.xbps; do
			[ -e "$pkg" ] || continue
			xbps-rindex --sign-pkg --privkey "$XBPS_SIGNING_KEY" "$pkg"
		done
	else
		echo "build-stack: XBPS_SIGNING_KEY is not set; packages and repository are unsigned."
	fi
fi

echo "build-stack: complete."
