#!/bin/sh
# Check that the live rootfs inside a built ISO has its key binaries and that
# their shared libraries resolve. Used by the `iso` job in ci.yml and by
# release.yml (through .github/actions/build-iso).
#
# Usage: installation/ci/verify-iso-rootfs.sh ISO
# Needs root (sudo) for loop mounts.
set -eu

iso=${1:?usage: verify-iso-rootfs.sh ISO}
WORK=$(mktemp -d)
MNT="$WORK/iso"
SQUASH_MNT="$WORK/squash"
ROOTFS_MNT="$WORK/rootfs"
mkdir -p "$MNT" "$SQUASH_MNT" "$ROOTFS_MNT"

cleanup() {
	for m in "$ROOTFS_MNT" "$SQUASH_MNT" "$MNT"; do
		if mountpoint -q "$m"; then sudo umount "$m" || true; fi
	done
	rm -rf "$WORK"
}
trap cleanup EXIT

sudo mount -o loop,ro "$iso" "$MNT"
sudo mount -o loop,ro "$MNT/LiveOS/squashfs.img" "$SQUASH_MNT"

TARGET_ROOT="$SQUASH_MNT"
if [ -f "$SQUASH_MNT/LiveOS/rootfs.img" ]; then
	sudo mount -o loop,ro "$SQUASH_MNT/LiveOS/rootfs.img" "$ROOTFS_MNT"
	TARGET_ROOT="$ROOTFS_MNT"
fi

echo "== Checking Hyprland, Quickshell, and ASURA binaries in rootfs ($TARGET_ROOT) =="
failed=0
check_binary() {
	bin="$1"
	echo "Checking $bin..."
	if [ ! -f "$TARGET_ROOT$bin" ]; then
		echo "::error::Binary $bin is missing from Live rootfs!"
		failed=$((failed + 1))
		return
	fi
	ldso="$(find "$TARGET_ROOT/usr/lib"* "$TARGET_ROOT/lib"* -name 'ld-linux*.so*' 2>/dev/null | head -n 1 || true)"
	if [ -n "$ldso" ] && [ -x "$ldso" ]; then
		missing=$(LD_LIBRARY_PATH="$TARGET_ROOT/usr/lib:$TARGET_ROOT/lib" "$ldso" --list "$TARGET_ROOT$bin" 2>&1 | grep "not found" || true)
	else
		missing=$(LD_LIBRARY_PATH="$TARGET_ROOT/usr/lib:$TARGET_ROOT/lib" ldd "$TARGET_ROOT$bin" 2>&1 | grep "not found" || true)
	fi
	if [ -n "$missing" ]; then
		echo "::error::Binary $bin has missing shared libraries:"
		echo "$missing"
		failed=$((failed + 1))
	else
		echo "  OK: all shared libraries resolved for $bin"
	fi
}

check_binary "/usr/bin/Hyprland"
check_binary "/usr/bin/asura"
check_binary "/usr/bin/asura-shell"
check_binary "/usr/bin/asura-sysd"
if [ -f "$TARGET_ROOT/usr/bin/quickshell" ]; then
	check_binary "/usr/bin/quickshell"
fi

# Plymouth label plugin
if [ ! -f "$TARGET_ROOT/usr/lib/plymouth/label.so" ]; then
	echo "::error::Plymouth label.so is missing from /usr/lib/plymouth!"
	failed=$((failed + 1))
else
	echo "  OK: /usr/lib/plymouth/label.so present"
fi

# No foreign libsystemd requirement in libseat
if [ -f "$TARGET_ROOT/usr/lib/libseat.so.1" ]; then
	if readelf -d "$TARGET_ROOT/usr/lib/libseat.so.1" 2>/dev/null | grep -q "libsystemd.so.0"; then
		echo "::error::/usr/lib/libseat.so.1 is dynamically linked against libsystemd.so.0!"
		failed=$((failed + 1))
	else
		echo "  OK: /usr/lib/libseat.so.1 is native (no libsystemd requirement)"
	fi
fi

if [ "$failed" -gt 0 ]; then
	echo "Rootfs library validation failed with $failed error(s)!"
	exit 1
fi
echo "Rootfs library and binary validation passed cleanly!"
