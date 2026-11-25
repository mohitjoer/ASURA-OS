#!/bin/sh
# Build ASURA's XBPS package and repository.
#
# Usage: packaging/xbps/build-repo.sh OUT_DIR STAGE_DIR
#
# STAGE_DIR is a tree installation/apply-repo.sh has already filled (with the
# Go tools built): the package holds exactly what an update installs. This
# script, which needs xbps-create and xbps-rindex (a Void container in CI),
# turns that tree into the `toren` package, indexes OUT_DIR, and signs both
# when XBPS_SIGNING_KEY names a private key (PEM) file.
#
# The version is 0.1.<commits on the branch>_1, so every newer commit is a
# newer package and xbps upgrades to it; ASURA_PKG_VERSION overrides it.
set -eu

OUT_DIR=${1:?usage: build-repo.sh OUT_DIR STAGE_DIR}
STAGE=${2:?usage: build-repo.sh OUT_DIR STAGE_DIR}
STAGE=$(CDPATH='' cd -- "$STAGE" && pwd)
REPO_ROOT=$(CDPATH='' cd -- "$(dirname -- "$0")/../.." && pwd)
ARCH=x86_64

if [ -n "${ASURA_PKG_VERSION:-}" ]; then
	VERSION=$ASURA_PKG_VERSION
else
	count=$(git -C "$REPO_ROOT" rev-list --count HEAD 2>/dev/null || echo 0)
	VERSION="0.1.$count"
fi
PKGVER="toren-${VERSION}_1"

# The package must hold what makes a machine ASURA: refuse a stage that is
# missing any of these (an apply-repo run that failed part way).
for f in \
	usr/bin/toren \
	usr/bin/toren-session \
	usr/share/toren/session/toren-activate \
	usr/share/toren/shell-desktop/shell.qml \
	usr/share/toren/hyprland/hyprland.lua \
	usr/share/toren/hyprland/hyprland.conf \
	usr/share/toren/migrations; do
	if [ ! -e "$STAGE/$f" ]; then
		echo "build-repo: $f is not in $STAGE; run installation/apply-repo.sh into it first" >&2
		exit 1
	fi
done
for f in usr/bin/toren usr/bin/toren-session usr/share/toren/session/toren-activate; do
	[ -x "$STAGE/$f" ] || { echo "build-repo: $f is not executable" >&2; exit 1; }
done

# Files another Void package owns cannot be in this one (xbps refuses the
# install); apply-repo.sh keeps installing ASURA's copies of them.
for f in etc/nftables.conf etc/turnstile/turnstile.conf; do
	rm -f "$STAGE/$f"
done

# Files under /etc are the machine's: xbps keeps a locally edited one and
# installs the new copy beside it, instead of overwriting the edit. Service
# scripts (/etc/sv) are ASURA's code, not configuration, and always update.
conf_files=$(cd "$STAGE" && find etc -type f ! -path 'etc/sv/*' 2>/dev/null | sed 's#^#/#' | sort | tr '\n' ' ')

mkdir -p "$OUT_DIR"
OUT_DIR=$(CDPATH='' cd -- "$OUT_DIR" && pwd)
echo "build-repo: creating $PKGVER"
(
	cd "$OUT_DIR"
	xbps-create \
		-A "$ARCH" \
		-n "$PKGVER" \
		-s "Toren Linux: the desktop, its tools, and its system configuration" \
		-m "mohitjoer <mohitjoe.r@gmail.com>" \
		-l "GPL-3.0-only" \
		-H "https://github.com/mohitjoer/Toren-Linux" \
		-D "hyprland>=0.57 hyprland-guiutils>=0 quickshell>=0 seatd>=0 turnstile>=0" \
		-F "$conf_files" \
		"$STAGE"
)

echo "build-repo: indexing $OUT_DIR"
xbps-rindex -a "$OUT_DIR/$PKGVER.$ARCH.xbps"

# Retain up to 3 versions per package name, then clean obsolete index entries
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
            except OSError:
                pass
PY
fi

xbps-rindex -c "$OUT_DIR" 2>/dev/null || true
# A rolling repository: drop packages the index no longer names.
xbps-rindex -r "$OUT_DIR"

if [ -n "${XBPS_SIGNING_KEY:-}" ]; then
	echo "build-repo: signing the repository and the package"
	xbps-rindex --sign --signedby "Toren Linux" --privkey "$XBPS_SIGNING_KEY" "$OUT_DIR"
	xbps-rindex --sign-pkg --privkey "$XBPS_SIGNING_KEY" "$OUT_DIR/$PKGVER.$ARCH.xbps"
else
	echo "build-repo: XBPS_SIGNING_KEY is not set; the repository is unsigned (not publishable)"
fi

ls -l "$OUT_DIR"
