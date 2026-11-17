#!/bin/sh
set -eu

REPO_ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
INSTALLER="$REPO_ROOT/installation/asura-installer.sh"

if [ ! -x "$INSTALLER" ]; then
	echo "asura-installer_test: installation/asura-installer.sh is not executable" >&2
	exit 1
fi

if ! sh -n "$INSTALLER"; then
	echo "asura-installer_test: shell syntax check failed on asura-installer.sh" >&2
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
	echo "asura-installer_test: missing LUKS2 in specification summary" >&2
	exit 1
}
grep -Fq "Format root partition with LUKS2 (argon2id, label ASURA_CRYPT)" "$OUT_ENCRYPT" || {
	echo "asura-installer_test: missing LUKS2 formatting step in dry-run summary" >&2
	exit 1
}
grep -Fq "Open LUKS container as /dev/mapper/asura_root" "$OUT_ENCRYPT" || {
	echo "asura-installer_test: missing /dev/mapper/asura_root step in dry-run summary" >&2
	exit 1
}
grep -Fq "Configure /etc/crypttab, dracut crypt module, fstab with persistent UUIDs, and GRUB rd.luks.uuid" "$OUT_ENCRYPT" || {
	echo "asura-installer_test: missing crypttab/dracut step in dry-run summary" >&2
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
	echo "asura-installer_test: missing 'Encryption: None' in unencrypted specification summary" >&2
	exit 1
}
! sed -n '/DRY RUN SUMMARY:/,$p' "$OUT_NO_ENCRYPT" | grep -q "LUKS" || {
	echo "asura-installer_test: unexpected LUKS mention in unencrypted dry-run summary" >&2
	exit 1
}

# Test 3: Passphrase never leaks in dry-run output
if grep -q "testpassphrase123" "$OUT_ENCRYPT"; then
	echo "asura-installer_test: SECURITY LEAK: passphrase found in installer output!" >&2
	exit 1
fi

echo "asura-installer_test: ok"
