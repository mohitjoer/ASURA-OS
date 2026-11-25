#!/bin/sh
set -eu

REPO_ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
INSTALLER="$REPO_ROOT/installation/toren-installer.sh"

if [ ! -x "$INSTALLER" ]; then
	echo "toren-installer_test: installation/toren-installer.sh is not executable" >&2
	exit 1
fi

if ! sh -n "$INSTALLER"; then
	echo "toren-installer_test: shell syntax check failed on asura-installer.sh" >&2
	exit 1
fi

TMP_DIR=$(mktemp -d 2>/dev/null || echo "/tmp/asura-test.$$")
cleanup() {
	rm -rf "$TMP_DIR"
}
trap cleanup EXIT INT TERM

# Test 1: Dry run with encryption on (default or --encrypt)
OUT_ENCRYPT="$TMP_DIR/out_encrypt.txt"
PASS_FILE="$TMP_DIR/passphrase.txt"
printf "testpassphrase123\n" > "$PASS_FILE"

ASURA_LUKS_PASSPHRASE_FILE="$PASS_FILE" sh "$INSTALLER" --dry-run --encrypt /dev/sda <<'EOF' > "$OUT_ENCRYPT" 2>&1
1
test-host
test-user
userpass123
userpass123
Y
EOF

grep -Fq "LUKS2 (argon2id)" "$OUT_ENCRYPT" || {
	echo "toren-installer_test: missing LUKS2 in specification summary" >&2
	exit 1
}
grep -Fq "Format root partition with LUKS2 (argon2id, label ASURA_CRYPT)" "$OUT_ENCRYPT" || {
	echo "toren-installer_test: missing LUKS2 formatting step in dry-run summary" >&2
	exit 1
}
grep -Fq "Open LUKS container as /dev/mapper/asura_root" "$OUT_ENCRYPT" || {
	echo "toren-installer_test: missing /dev/mapper/asura_root step in dry-run summary" >&2
	exit 1
}
grep -Fq "Configure /etc/crypttab, dracut crypt module, fstab with persistent UUIDs, and GRUB rd.luks.uuid" "$OUT_ENCRYPT" || {
	echo "toren-installer_test: missing crypttab/dracut step in dry-run summary" >&2
	exit 1
}

# Test 2: Dry run with encryption disabled (--no-encrypt)
OUT_NO_ENCRYPT="$TMP_DIR/out_no_encrypt.txt"
sh "$INSTALLER" --dry-run --no-encrypt /dev/sda <<'EOF' > "$OUT_NO_ENCRYPT" 2>&1
1
test-host
test-user
userpass123
userpass123
Y
EOF

grep -Eq "Encryption:.*None" "$OUT_NO_ENCRYPT" || {
	echo "toren-installer_test: missing 'Encryption: None' in unencrypted specification summary" >&2
	exit 1
}
! sed -n '/DRY RUN SUMMARY:/,$p' "$OUT_NO_ENCRYPT" | grep -q "LUKS" || {
	echo "toren-installer_test: unexpected LUKS mention in unencrypted dry-run summary" >&2
	exit 1
}

# Test 3: Passphrase never leaks in dry-run output
if grep -q "testpassphrase123" "$OUT_ENCRYPT"; then
	echo "toren-installer_test: SECURITY LEAK: passphrase found in installer output!" >&2
	exit 1
fi

# Test 4: Unattended installation via --config answer file
CONFIG_FILE="$TMP_DIR/asura-answer.yaml"
cat >"$CONFIG_FILE" <<'EOF'
disk: /dev/sda
mode: wipe
filesystem: btrfs
encryption: yes
luks_password: secretunattendedpass123
hostname: unattended-host
username: unattended-user
password_hash: $6$rounds=4096$saltsalt$fakehashvalueforunattended123
timezone: UTC
locale: en_US.UTF-8
keyboard: us
EOF

OUT_CONFIG="$TMP_DIR/out_config.txt"
sh "$INSTALLER" --dry-run --config "$CONFIG_FILE" > "$OUT_CONFIG" 2>&1

grep -Fq "unattended-host" "$OUT_CONFIG" || {
	echo "toren-installer_test: missing unattended hostname in output" >&2
	exit 1
}
grep -Fq "unattended-user" "$OUT_CONFIG" || {
	echo "toren-installer_test: missing unattended user in output" >&2
	exit 1
}
if grep -q "secretunattendedpass123" "$OUT_CONFIG" || grep -q "fakehashvalueforunattended123" "$OUT_CONFIG"; then
	echo "toren-installer_test: SECURITY LEAK: unattended password or hash found in installer output!" >&2
	exit 1
fi

# Test 5: Alongside / dual-boot installation (--alongside / --dual-boot)
OUT_ALONGSIDE="$TMP_DIR/out_alongside.txt"
sh "$INSTALLER" --dry-run --alongside --no-encrypt /dev/sda <<'EOF' > "$OUT_ALONGSIDE" 2>&1
1
dual-host
dual-user
dualpass123
dualpass123
Y
EOF

grep -Fq "Install Alongside" "$OUT_ALONGSIDE" || {
	echo "toren-installer_test: missing 'Install Alongside' in specification summary" >&2
	exit 1
}
grep -Fq "preserve existing partition table" "$OUT_ALONGSIDE" || {
	echo "toren-installer_test: missing partition table preservation in alongside dry-run" >&2
	exit 1
}
grep -Fq "os-prober enabled" "$OUT_ALONGSIDE" || {
	echo "toren-installer_test: missing os-prober enablement in alongside dry-run" >&2
	exit 1
}
grep -Fq "without overwriting fallback EFI/BOOT/BOOTX64.EFI" "$OUT_ALONGSIDE" || {
	echo "toren-installer_test: missing fallback bootloader preservation in alongside dry-run" >&2
	exit 1
}
! sed -n '/DRY RUN SUMMARY:/,$p' "$OUT_ALONGSIDE" | grep -q "Wipe" || {
	echo "toren-installer_test: unexpected wipe in alongside dry-run summary" >&2
	exit 1
}

# The snapshot GRUB snippet is installed before the first grub-mkconfig, so
# the installed system's first grub.cfg already carries it.
snippet_line=$(grep -n '40_asura_snapshots' "$REPO_ROOT/installation/toren-installer.sh" | head -n1 | cut -d: -f1)
mkconfig_line=$(grep -n 'grub-mkconfig -o' "$REPO_ROOT/installation/toren-installer.sh" | head -n1 | cut -d: -f1)
if [ -z "$snippet_line" ] || [ -z "$mkconfig_line" ] || [ "$snippet_line" -gt "$mkconfig_line" ]; then
	echo "toren-installer_test: 40_asura_snapshots is not installed before grub-mkconfig" >&2
	exit 1
fi

echo "toren-installer_test: ok"
