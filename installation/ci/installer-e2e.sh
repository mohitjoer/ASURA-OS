#!/bin/sh
# End-to-end installer test: install ASURA from a built ISO onto a disk image,
# then check the result.
#
# Usage: installation/ci/installer-e2e.sh ISO MODE OUTDIR
#   MODE plain  Btrfs, no encryption. The installed disk is then booted in
#               QEMU (BIOS), and the test passes when runit reaches stage 2.
#   MODE luks   Btrfs on LUKS2. The installed disk is unlocked with the
#               passphrase and its crypttab, fstab and GRUB command line are
#               checked. (Booting it would need the passphrase typed on the
#               serial console; that is a follow-up.)
#
# The installer that runs is this checkout's installation/asura-installer.sh,
# inside the ISO's own live rootfs (read-only, with a tmpfs overlay on top),
# so it copies exactly the system the ISO would install. Firmware is forced to
# BIOS by hiding the host's /sys/firmware/efi: a UEFI install inside CI would
# try to write the host's NVRAM.
#
# Needs root (sudo), loop devices, qemu-system-x86_64 (plain mode) and,
# ideally, /dev/kvm. Writes OUTDIR/install.log, OUTDIR/serial.log (plain).
set -eu

ISO=${1:?usage: installer-e2e.sh ISO MODE OUTDIR}
MODE=${2:?usage: installer-e2e.sh ISO MODE OUTDIR}
OUT=${3:?usage: installer-e2e.sh ISO MODE OUTDIR}
case "$MODE" in plain|luks) ;; *) echo "installer-e2e: MODE must be plain or luks" >&2; exit 2 ;; esac

REPO_ROOT=$(CDPATH='' cd -- "$(dirname -- "$0")/../.." && pwd)
mkdir -p "$OUT"
OUT=$(CDPATH='' cd -- "$OUT" && pwd)
WORK=$(mktemp -d)
ISO_MNT="$WORK/iso"
SQUASH_MNT="$WORK/squash"
ROOTFS_MNT="$WORK/rootfs"
OVL="$WORK/overlay"
CHROOT="$WORK/chroot"
DISK="$WORK/disk.img"
TARGET_MNT="$WORK/target"
LOOP=""
PASSPHRASE="asura-ci-passphrase"
USER_NAME="ciuser"
USER_PASS="ci-password-123"

log() { printf 'installer-e2e: %s\n' "$*"; }

cleanup() {
	set +e
	for m in "$TARGET_MNT/boot" "$TARGET_MNT" \
		"$CHROOT/sys/firmware" "$CHROOT/run" "$CHROOT/sys" "$CHROOT/proc" \
		"$CHROOT/dev/pts" "$CHROOT/dev" "$CHROOT" "$OVL" \
		"$ROOTFS_MNT" "$SQUASH_MNT" "$ISO_MNT"; do
		mountpoint -q "$m" 2>/dev/null && sudo umount -l "$m"
	done
	[ -e /dev/mapper/asura_ci ] && sudo cryptsetup close asura_ci
	[ -e /dev/mapper/asura_root ] && sudo cryptsetup close asura_root
	[ -n "$LOOP" ] && sudo losetup -d "$LOOP"
	[ -f "$WORK/qemu.pid" ] && sudo kill "$(cat "$WORK/qemu.pid")" 2>/dev/null
	sudo rm -rf "$WORK"
}
trap cleanup EXIT

# ── 1. The ISO's live root, writable through an overlay ──────────────────────
mkdir -p "$ISO_MNT" "$SQUASH_MNT" "$ROOTFS_MNT" "$OVL" "$CHROOT" "$TARGET_MNT"
sudo mount -o loop,ro "$ISO" "$ISO_MNT"
sudo mount -o loop,ro "$ISO_MNT/LiveOS/squashfs.img" "$SQUASH_MNT"
LOWER="$SQUASH_MNT"
if [ -f "$SQUASH_MNT/LiveOS/rootfs.img" ]; then
	sudo mount -o loop,ro "$SQUASH_MNT/LiveOS/rootfs.img" "$ROOTFS_MNT"
	LOWER="$ROOTFS_MNT"
fi
sudo mount -t tmpfs -o size=4G tmpfs "$OVL"
sudo mkdir -p "$OVL/upper" "$OVL/work"
sudo mount -t overlay overlay -o "lowerdir=$LOWER,upperdir=$OVL/upper,workdir=$OVL/work" "$CHROOT"

sudo mount --bind /dev "$CHROOT/dev"
sudo mount --bind /dev/pts "$CHROOT/dev/pts"
sudo mount -t proc proc "$CHROOT/proc"
sudo mount --bind /sys "$CHROOT/sys"
# Force BIOS: the installer decides by /sys/firmware/efi.
sudo mount -t tmpfs tmpfs "$CHROOT/sys/firmware"
sudo mount -t tmpfs tmpfs "$CHROOT/run"

# This checkout's installer, not the one baked into the ISO.
sudo install -m 0755 "$REPO_ROOT/installation/asura-installer.sh" "$CHROOT/tmp/asura-installer.sh"
printf '%s\n' "$PASSPHRASE" | sudo tee "$CHROOT/tmp/luks-passphrase" >/dev/null

# ── 2. The target disk ──────────────────────────────────────────────────────
truncate -s 12G "$DISK"
LOOP=$(sudo losetup -fP --show "$DISK")
log "target disk $LOOP ($MODE)"

if [ "$MODE" = luks ]; then
	ENC_FLAG=--encrypt
	ENC_ENV="ASURA_LUKS_PASSPHRASE_FILE=/tmp/luks-passphrase"
else
	ENC_FLAG=--no-encrypt
	ENC_ENV=""
fi

# The answers, in the order the installer asks: filesystem (1 = Btrfs),
# hostname, user, password twice, lock root (Y), the YES that erases the disk,
# no reboot, and the final Enter.
log "running the installer"
set +e
# The log is written by this (unprivileged) shell on purpose: it lands in OUTDIR.
# shellcheck disable=SC2024,SC2086
printf '1\nasura-ci\n%s\n%s\n%s\nY\nYES\nn\n\n' "$USER_NAME" "$USER_PASS" "$USER_PASS" |
	sudo chroot "$CHROOT" /usr/bin/env -i PATH=/usr/bin:/usr/sbin:/bin:/sbin TERM=dumb HOME=/root $ENC_ENV \
		sh /tmp/asura-installer.sh "$ENC_FLAG" "$LOOP" >"$OUT/install.log" 2>&1
rc=$?
set -e
tail -n 40 "$OUT/install.log"
if [ "$rc" -ne 0 ]; then
	echo "::error::the installer exited $rc (full log in install.log)"
	exit 1
fi
grep -q "All post-installation verification checks passed" "$OUT/install.log" || {
	echo "::error::the installer finished without passing its own verification"
	exit 1
}
log "installer finished and verified itself"

# The installer unmounts what it mounted; close its LUKS mapping too if it is
# still open, then re-attach the image so the partitions are read afresh.
[ -e /dev/mapper/asura_root ] && sudo cryptsetup close asura_root
sudo losetup -d "$LOOP"
LOOP=$(sudo losetup -fP --show "$DISK")
sudo udevadm settle 2>/dev/null || true
sleep 1

if [ "$MODE" = luks ]; then
	# ── 3a. Encrypted: unlock with the passphrase and check the wiring ──────
	ROOT_PART="${LOOP}p2"
	sudo cryptsetup isLuks "$ROOT_PART" || { echo "::error::$ROOT_PART is not LUKS"; exit 1; }
	printf '%s' "$PASSPHRASE" | sudo cryptsetup open --key-file=- "$ROOT_PART" asura_ci
	LUKS_UUID=$(sudo cryptsetup luksUUID "$ROOT_PART")
	sudo mount -o subvol=@ /dev/mapper/asura_ci "$TARGET_MNT"
	sudo mount "${LOOP}p1" "$TARGET_MNT/boot"
	sudo grep -q "UUID=$LUKS_UUID" "$TARGET_MNT/etc/crypttab" || { echo "::error::crypttab does not name the LUKS partition"; exit 1; }
	sudo grep -q "rd.luks.uuid=$LUKS_UUID" "$TARGET_MNT/etc/default/grub" || { echo "::error::GRUB_CMDLINE_LINUX lacks rd.luks.uuid"; exit 1; }
	sudo grep -q "rd.luks.uuid=$LUKS_UUID" "$TARGET_MNT/boot/grub/grub.cfg" || { echo "::error::grub.cfg lacks rd.luks.uuid"; exit 1; }
	sudo grep -qE '^UUID=[^ ]+ +/ +btrfs' "$TARGET_MNT/etc/fstab" || { echo "::error::fstab has no Btrfs root"; exit 1; }
	sudo grep -qE ' /boot +ext4' "$TARGET_MNT/etc/fstab" || { echo "::error::fstab has no /boot"; exit 1; }
	log "LUKS install: unlocks with the passphrase; crypttab, fstab and GRUB agree ($LUKS_UUID)"
	exit 0
fi

# ── 3b. Plain: boot the installed disk ──────────────────────────────────────
# The installed command line is quiet and ends on console=tty0. For the test,
# make the serial port the console and drop the quiet, in the test's copy only.
ROOT_PART="${LOOP}p1"
sudo mount -o subvol=@ "$ROOT_PART" "$TARGET_MNT"
GRUB_CFG="$TARGET_MNT/boot/grub/grub.cfg"
sudo test -s "$GRUB_CFG" || { echo "::error::no grub.cfg on the installed disk"; exit 1; }
sudo sed -i -E '/^\s*linux\s/{s/ (quiet|splash|loglevel=[0-9]+|rd\.plymouth=1|plymouth\.enable=1)//g; s/ console=tty0//g; s/$/ console=ttyS0,115200/}' "$GRUB_CFG"
sudo umount "$TARGET_MNT"
sudo losetup -d "$LOOP"
LOOP=""

ACCEL="-cpu qemu64"
[ -w /dev/kvm ] && ACCEL="-enable-kvm -cpu host"
log "booting the installed disk ($ACCEL)"
# shellcheck disable=SC2086
sudo qemu-system-x86_64 $ACCEL -m 2048 -smp 2 \
	-drive file="$DISK",format=raw,if=virtio \
	-display none -serial file:"$OUT/serial.log" \
	-daemonize -pidfile "$WORK/qemu.pid"

booted=0
for i in $(seq 1 120); do
	sleep 2
	if sudo grep -qa "runit: enter stage: /etc/runit/2" "$OUT/serial.log" 2>/dev/null; then
		log "installed system reached runit stage 2 after ~$((i * 2))s"
		booted=1
		break
	fi
	if sudo grep -qaE "Kernel panic|dracut-initqueue.*timeout|Could not boot|emergency shell" "$OUT/serial.log" 2>/dev/null; then
		break
	fi
done
sudo chmod a+r "$OUT/serial.log" 2>/dev/null || true
echo "== serial (last 60 lines) =="
tail -n 60 "$OUT/serial.log" 2>/dev/null || true
if [ "$booted" -ne 1 ]; then
	echo "::error::the installed system did not reach runit stage 2"
	exit 1
fi
