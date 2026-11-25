#!/bin/sh
# Toren Linux Guided System Installer
# Sleek, interactive system installer with automatic drive detection,
# Btrfs/ext4 support, subvolume setup, and hybrid UEFI/BIOS bootloader.
set -eu

DUAL_BOOT=0
DRY_RUN=0
ENCRYPT_ARG=""
TARGET_DISK_ARG=""
CONFIG_ARG=""

while [ $# -gt 0 ]; do
	case "$1" in
		--dual-boot|--alongside)
			DUAL_BOOT=1
			shift
			;;
		--wipe|--clean)
			DUAL_BOOT=0
			shift
			;;
		--dry-run)
			DRY_RUN=1
			shift
			;;
		--encrypt)
			ENCRYPT_ARG=1
			shift
			;;
		--no-encrypt)
			ENCRYPT_ARG=0
			shift
			;;
		--config)
			if [ $# -lt 2 ]; then
				echo "Error: --config requires a file argument" >&2
				exit 1
			fi
			CONFIG_ARG="$2"
			shift 2
			;;
		--config=*)
			CONFIG_ARG="${1#*=}"
			shift
			;;
		-h|--help|help)
			echo "Usage: toren-installer [options] [TARGET_DISK]"
			echo "Options:"
			echo "  --config FILE          unattended answer file (YAML or key=value)"
			echo "  --alongside            install alongside existing OS in free space"
			echo "  --dual-boot            alias for --alongside"
			echo "  --wipe                 erase entire disk (clean installation)"
			echo "  --dry-run              print installation plan without making changes"
			echo "  --encrypt              enable LUKS2 root filesystem encryption"
			echo "  --no-encrypt           disable encryption"
			exit 0
			;;
		-*)
			echo "Unknown option: $1" >&2
			echo "Usage: toren-installer [--config FILE] [--alongside] [--dry-run] [--encrypt|--no-encrypt] [TARGET_DISK]" >&2
			exit 1
			;;
		*)
			if [ -z "$TARGET_DISK_ARG" ]; then
				TARGET_DISK_ARG="$1"
			fi
			shift
			;;
	esac
done

# Answer file parsing (unattended configuration)
CFG_DISK=""
CFG_MODE=""
CFG_FS=""
CFG_ENCRYPT=""
CFG_LUKS_PASS=""
CFG_LUKS_PASS_FILE=""
CFG_HOSTNAME=""
CFG_USER=""
CFG_PASS=""
CFG_PASS_HASH=""
CFG_ROOT_PASS=""
CFG_LOCK_ROOT=""
CFG_TIMEZONE=""
CFG_LOCALE=""
CFG_KEYMAP=""

get_answer_val() {
	k="$1"
	f="$2"
	[ -f "$f" ] || return 0
	val=$(grep -E "^[[:space:]]*${k}[[:space:]]*[:=]" "$f" 2>/dev/null | head -n1 | sed -E "s/^[[:space:]]*${k}[[:space:]]*[:=][[:space:]]*//; s/^[\"']//; s/[\"'][[:space:]]*$//" || true)
	printf '%s' "$val"
}

if [ -n "$CONFIG_ARG" ]; then
	if [ ! -f "$CONFIG_ARG" ]; then
		echo "Error: configuration answer file not found: $CONFIG_ARG" >&2
		exit 1
	fi
	CFG_DISK=$(get_answer_val "disk" "$CONFIG_ARG")
	CFG_MODE=$(get_answer_val "mode" "$CONFIG_ARG")
	CFG_FS=$(get_answer_val "filesystem" "$CONFIG_ARG")
	CFG_ENCRYPT=$(get_answer_val "encryption" "$CONFIG_ARG")
	CFG_LUKS_PASS=$(get_answer_val "luks_password" "$CONFIG_ARG")
	CFG_LUKS_PASS_FILE=$(get_answer_val "luks_passphrase_file" "$CONFIG_ARG")
	CFG_HOSTNAME=$(get_answer_val "hostname" "$CONFIG_ARG")
	CFG_USER=$(get_answer_val "user" "$CONFIG_ARG")
	[ -z "$CFG_USER" ] && CFG_USER=$(get_answer_val "username" "$CONFIG_ARG")
	CFG_PASS=$(get_answer_val "password" "$CONFIG_ARG")
	CFG_PASS_HASH=$(get_answer_val "password_hash" "$CONFIG_ARG")
	CFG_ROOT_PASS=$(get_answer_val "root_password" "$CONFIG_ARG")
	CFG_LOCK_ROOT=$(get_answer_val "lock_root" "$CONFIG_ARG")
	CFG_TIMEZONE=$(get_answer_val "timezone" "$CONFIG_ARG")
	CFG_LOCALE=$(get_answer_val "locale" "$CONFIG_ARG")
	CFG_KEYMAP=$(get_answer_val "keyboard" "$CONFIG_ARG")
	[ -z "$CFG_KEYMAP" ] && CFG_KEYMAP=$(get_answer_val "keymap" "$CONFIG_ARG")

	[ -n "$CFG_DISK" ] && TARGET_DISK_ARG="$CFG_DISK"
	case "$CFG_MODE" in
		alongside|dual-boot) DUAL_BOOT=1 ;;
		wipe|clean) DUAL_BOOT=0 ;;
	esac
	case "$CFG_ENCRYPT" in
		yes|true|1) ENCRYPT_ARG=1 ;;
		no|false|0) ENCRYPT_ARG=0 ;;
	esac
fi

BOLD=$(printf '\033[1m')
DIM=$(printf '\033[2m')
WHITE=$(printf '\033[1;37m')
GREEN=$(printf '\033[1;32m')
YELLOW=$(printf '\033[1;33m')
RED=$(printf '\033[1;31m')
RESET=$(printf '\033[0m')
CYAN=$(printf '\033[1;36m')

SAVED_STTY=""
if [ -t 0 ]; then
	SAVED_STTY=$(stty -g 2>/dev/null || true)
fi

DISKS_FILE=$(mktemp 2>/dev/null || echo "/tmp/toren-disks.$$")
TARGET_DIR="/mnt/toren-target"

cleanup_installer() {
	exit_code=$?
	# Always restore terminal settings and echo
	if [ -n "$SAVED_STTY" ]; then
		stty "$SAVED_STTY" 2>/dev/null || stty echo 2>/dev/null || true
	else
		stty echo 2>/dev/null || true
	fi
	rm -f "$DISKS_FILE" 2>/dev/null || true
	if [ -d "$TARGET_DIR" ]; then
		# Unmount in reverse order
		umount "$TARGET_DIR/boot/efi" 2>/dev/null || true
		umount "$TARGET_DIR/boot" 2>/dev/null || true
		umount "$TARGET_DIR/sys/firmware/efi/efivars" 2>/dev/null || true
		umount "$TARGET_DIR/sys" 2>/dev/null || true
		umount "$TARGET_DIR/proc" 2>/dev/null || true
		umount "$TARGET_DIR/dev/pts" 2>/dev/null || true
		umount "$TARGET_DIR/dev" 2>/dev/null || true
		umount "$TARGET_DIR/home" 2>/dev/null || true
		umount "$TARGET_DIR/.snapshots" 2>/dev/null || true
		umount "$TARGET_DIR" 2>/dev/null || true
		rmdir "$TARGET_DIR" 2>/dev/null || true
	fi
	if [ -b /dev/mapper/toren_root ]; then
		cryptsetup close toren_root 2>/dev/null || true
	fi
	if [ "$exit_code" -ne 0 ] && [ "$DRY_RUN" -eq 0 ]; then
		printf "\n${RED}${BOLD}Installation failed (exit code %s). Cleaned up mounts.${RESET}\n" "$exit_code"
		printf "${WHITE}Press Enter to close...${RESET}"
		read -r _ 2>/dev/null || true
	fi
}

on_signal() {
	trap - EXIT INT TERM
	if [ -n "$SAVED_STTY" ]; then
		stty "$SAVED_STTY" 2>/dev/null || stty echo 2>/dev/null || true
	else
		stty echo 2>/dev/null || true
	fi
	rm -f "$DISKS_FILE" 2>/dev/null || true
	if [ -d "$TARGET_DIR" ]; then
		umount "$TARGET_DIR/boot/efi" 2>/dev/null || true
		umount "$TARGET_DIR/boot" 2>/dev/null || true
		umount "$TARGET_DIR/sys/firmware/efi/efivars" 2>/dev/null || true
		umount "$TARGET_DIR/sys" 2>/dev/null || true
		umount "$TARGET_DIR/proc" 2>/dev/null || true
		umount "$TARGET_DIR/dev/pts" 2>/dev/null || true
		umount "$TARGET_DIR/dev" 2>/dev/null || true
		umount "$TARGET_DIR/home" 2>/dev/null || true
		umount "$TARGET_DIR/.snapshots" 2>/dev/null || true
		umount "$TARGET_DIR" 2>/dev/null || true
		rmdir "$TARGET_DIR" 2>/dev/null || true
	fi
	if [ -b /dev/mapper/toren_root ]; then
		cryptsetup close toren_root 2>/dev/null || true
	fi
	printf "\n"
	exit 130
}

trap on_signal INT TERM
trap cleanup_installer EXIT

clear 2>/dev/null || true

cat <<EOF
${WHITE}╭──────────────────────────────────────────────────────────────╮
│                         ${BOLD}TOREN LINUX${RESET}${WHITE}                          │
│                   Guided System Installer                    │
╰──────────────────────────────────────────────────────────────╯${RESET}

EOF

if [ "$DRY_RUN" -eq 1 ]; then
	printf "${YELLOW}${BOLD}● DRY RUN MODE ACTIVE${RESET} ${DIM}(No disks will be partitioned or modified)${RESET}\n\n"
fi

if [ "$(id -u)" -ne 0 ] && [ "$DRY_RUN" -eq 0 ]; then
	printf "${RED}toren-installer: must be run as root (e.g. sudo toren install)${RESET}\n" >&2
	exit 1
fi

# Pre-flight dependency check
REQUIRED_TOOLS="rsync parted mkfs.vfat mkfs.ext4 grub-install grub-mkconfig blkid"
MISSING_TOOLS=""
for tool in $REQUIRED_TOOLS; do
	if ! command -v "$tool" >/dev/null 2>&1; then
		MISSING_TOOLS="$MISSING_TOOLS $tool"
	fi
done
if [ -n "$MISSING_TOOLS" ] && [ "$DRY_RUN" -eq 0 ]; then
	printf "${RED}Pre-flight error: required installer tool(s) missing:${BOLD}%s${RESET}\n" "$MISSING_TOOLS" >&2
	exit 1
fi

# Pre-flight: Check UEFI firmware vs Legacy BIOS
if [ ! -d /sys/firmware/efi ]; then
	printf "${YELLOW}● Boot Mode:${RESET} ${BOLD}Legacy BIOS${RESET} ${DIM}(MBR/BIOS bootloader will be installed)${RESET}\n\n"
	IS_UEFI=0
else
	printf "${GREEN}● Boot Mode:${RESET} ${BOLD}UEFI (Verified)${RESET} ${DIM}(GRUB EFI 64-bit bootloader)${RESET}\n\n"
	IS_UEFI=1
fi

# ── STEP 1: Interactive Drive Selection ─────────────────────────────────────
printf "${WHITE}╭──────────────────────────────────────────────────────────────╮
│  ${BOLD}[ STEP 1/5 ]  Target Drive Selection${RESET}${WHITE}                        │
╰──────────────────────────────────────────────────────────────╯${RESET}\n"

rm -f "$DISKS_FILE"
touch "$DISKS_FILE"
idx=1

lsblk -d -n -o NAME,SIZE,MODEL 2>/dev/null | grep -v '^zram' | while read -r dname dsize dmodel; do
	[ -z "$dname" ] && continue
	case "$dname" in
		zram*|loop*|sr*|dm-*) continue ;;
	esac
	echo "/dev/$dname" >> "$DISKS_FILE"
	rec_tag=""
	[ "$idx" -eq 1 ] && rec_tag="${DIM}[Recommended]${RESET}"
	printf "  ${WHITE}[%d]${RESET} ${BOLD}/dev/%-10s${RESET} %-8s %-22s %s\n" "$idx" "$dname" "$dsize" "${dmodel:-Disk}" "$rec_tag"
	idx=$((idx + 1))
done

total_disks=$(wc -l < "$DISKS_FILE" 2>/dev/null | tr -d ' ' || echo 0)

if [ "$total_disks" -eq 0 ] && [ -z "$TARGET_DISK_ARG" ] && [ "$DRY_RUN" -eq 0 ]; then
	printf "\n${RED}Error: No suitable install drives found.${RESET}\n" >&2
	exit 1
fi

if [ -n "$TARGET_DISK_ARG" ]; then
	TARGET_DISK="$TARGET_DISK_ARG"
elif [ "$total_disks" -gt 0 ]; then
	printf "\n${WHITE}Select target drive [1-%d, or type path] (default: 1): ${RESET}" "$total_disks"
	read -r DRIVE_CHOICE || DRIVE_CHOICE="1"
	[ -z "$DRIVE_CHOICE" ] && DRIVE_CHOICE="1"

	case "$DRIVE_CHOICE" in
		[0-9]*)
			if [ "$DRIVE_CHOICE" -ge 1 ] 2>/dev/null && [ "$DRIVE_CHOICE" -le "$total_disks" ] 2>/dev/null; then
				TARGET_DISK=$(sed -n "${DRIVE_CHOICE}p" "$DISKS_FILE")
			else
				TARGET_DISK="$DRIVE_CHOICE"
			fi
			;;
		*)
			TARGET_DISK="$DRIVE_CHOICE"
			;;
	esac
else
	TARGET_DISK="/dev/sda"
fi

if [ ! -b "$TARGET_DISK" ] && [ "$DRY_RUN" -eq 0 ]; then
	printf "${RED}Error: %s is not a valid block device.${RESET}\n" "$TARGET_DISK" >&2
	exit 1
fi

printf "${GREEN}✓ Selected drive:${RESET} ${BOLD}%s${RESET}\n\n" "$TARGET_DISK"

# ── STEP 2: Filesystem Selection ───────────────────────────────────────────
printf "${WHITE}╭──────────────────────────────────────────────────────────────╮
│  ${BOLD}[ STEP 2/5 ]  Select Root Filesystem${RESET}${WHITE}                        │
╰──────────────────────────────────────────────────────────────╯${RESET}\n"
printf "  ${WHITE}[1]${RESET} ${BOLD}Btrfs${RESET} ${DIM}(Recommended: subvolumes @, @home, @snapshots, zstd:1 compression)${RESET}\n"
printf "  ${WHITE}[2]${RESET} ${BOLD}ext4${RESET}  ${DIM}(Standard Linux journaling filesystem)${RESET}\n\n"

if [ -n "$CFG_FS" ]; then
	case "$CFG_FS" in
		ext4) FS_CHOICE="2" ;;
		*) FS_CHOICE="1" ;;
	esac
	printf "Filesystem (unattended): %s\n" "$([ "$FS_CHOICE" = "2" ] && echo "ext4" || echo "btrfs")"
else
	printf "${WHITE}Choose filesystem [1/2] (default: 1): ${RESET}"
	read -r FS_CHOICE || FS_CHOICE="1"
	[ -z "$FS_CHOICE" ] && FS_CHOICE="1"
fi

# ── STEP 3: Disk Encryption (LUKS2) ─────────────────────────────────────────
printf "\n${WHITE}╭──────────────────────────────────────────────────────────────╮
│  ${BOLD}[ STEP 3/5 ]  Full-Disk Encryption (LUKS2)${RESET}${WHITE}                    │
╰──────────────────────────────────────────────────────────────╯${RESET}\n"

if [ "$ENCRYPT_ARG" = "1" ]; then
	ENCRYPT_DISK=1
elif [ "$ENCRYPT_ARG" = "0" ]; then
	ENCRYPT_DISK=0
else
	printf "Encrypt the disk? [Y/n] (default: Y): "
	read -r ENCRYPT_INPUT || ENCRYPT_INPUT="Y"
	[ -z "$ENCRYPT_INPUT" ] && ENCRYPT_INPUT="Y"
	case "$ENCRYPT_INPUT" in
		[nN]|[nN][oO]) ENCRYPT_DISK=0 ;;
		*) ENCRYPT_DISK=1 ;;
	esac
fi

LUKS_PASS=""
if [ "$ENCRYPT_DISK" -eq 1 ]; then
	if [ -n "${CFG_LUKS_PASS:-}" ]; then
		LUKS_PASS="$CFG_LUKS_PASS"
	elif [ -n "${CFG_LUKS_PASS_FILE:-}" ] && [ -f "$CFG_LUKS_PASS_FILE" ]; then
		LUKS_PASS=$(head -n 1 "$CFG_LUKS_PASS_FILE" | tr -d '\r\n')
	elif [ -n "${TOREN_LUKS_PASSPHRASE_FILE:-}" ] && [ -f "$TOREN_LUKS_PASSPHRASE_FILE" ]; then
		LUKS_PASS=$(head -n 1 "$TOREN_LUKS_PASSPHRASE_FILE" | tr -d '\r\n')
	elif [ -n "${ASURA_LUKS_PASSPHRASE_FILE:-}" ] && [ -f "$ASURA_LUKS_PASSPHRASE_FILE" ]; then
		LUKS_PASS=$(head -n 1 "$ASURA_LUKS_PASSPHRASE_FILE" | tr -d '\r\n')
	fi

	if [ -z "$LUKS_PASS" ]; then
		while true; do
			printf "Enter disk encryption passphrase (min 8 characters): "
			stty -echo 2>/dev/null || true
			read -r LUKS_PASS || LUKS_PASS=""
			stty echo 2>/dev/null || true
			printf "\n"
			if [ -z "$LUKS_PASS" ]; then
				if [ "$DRY_RUN" -eq 1 ]; then
					LUKS_PASS="dryrunpassphrase123"
					break
				fi
				printf "${RED}Passphrase cannot be empty. Please enter a valid passphrase.${RESET}\n"
				continue
			fi
			if [ "${#LUKS_PASS}" -lt 8 ]; then
				if [ "$DRY_RUN" -eq 1 ]; then
					LUKS_PASS="dryrunpassphrase123"
					break
				fi
				printf "${RED}Passphrase must be at least 8 characters. Please try again.${RESET}\n"
				continue
			fi
			printf "Confirm disk encryption passphrase: "
			stty -echo 2>/dev/null || true
			read -r LUKS_PASS_CONFIRM || LUKS_PASS_CONFIRM=""
			stty echo 2>/dev/null || true
			printf "\n"
			if [ "$LUKS_PASS" = "$LUKS_PASS_CONFIRM" ]; then
				break
			fi
			printf "${RED}Passphrases do not match. Please try again.${RESET}\n"
		done
	fi

	if ! command -v cryptsetup >/dev/null 2>&1 && [ "$DRY_RUN" -eq 0 ]; then
		printf "${RED}Pre-flight error: required tool 'cryptsetup' missing for disk encryption.${RESET}\n" >&2
		exit 1
	fi
	printf "${GREEN}✓ Disk encryption enabled (LUKS2 with argon2id).${RESET}\n"
else
	printf "${DIM}Disk encryption disabled.${RESET}\n"
fi

# ── STEP 4: Identity & Configuration ───────────────────────────────────────
printf "\n${WHITE}╭──────────────────────────────────────────────────────────────╮
│  ${BOLD}[ STEP 4/5 ]  System Identity & User Setup${RESET}${WHITE}                  │
╰──────────────────────────────────────────────────────────────╯${RESET}\n"

if [ -n "$CFG_HOSTNAME" ]; then
	HOSTNAME_INPUT="$CFG_HOSTNAME"
	printf "Hostname (unattended): %s\n" "$HOSTNAME_INPUT"
else
	printf "Enter system hostname [default: toren-os]: "
	read -r HOSTNAME_INPUT || HOSTNAME_INPUT="toren-os"
	[ -z "$HOSTNAME_INPUT" ] && HOSTNAME_INPUT="toren-os"
fi

if [ -n "$CFG_USER" ]; then
	USERNAME_INPUT="$CFG_USER"
	printf "Primary user (unattended): %s\n" "$USERNAME_INPUT"
else
	printf "Enter primary username [default: toren]: "
	read -r USERNAME_INPUT || USERNAME_INPUT="toren"
	[ -z "$USERNAME_INPUT" ] && USERNAME_INPUT="toren"
fi

USER_PASS=""
USER_PASS_HASH=""

if [ -n "$CFG_PASS_HASH" ]; then
	USER_PASS_HASH="$CFG_PASS_HASH"
	printf "Password hash provided (unattended).\n"
elif [ -n "$CFG_PASS" ]; then
	USER_PASS="$CFG_PASS"
	printf "Password provided (unattended).\n"
else
	# Prompt for password for primary user
	while true; do
		printf "Enter password for %s: " "$USERNAME_INPUT"
		stty -echo 2>/dev/null || true
		read -r USER_PASS || USER_PASS=""
		stty echo 2>/dev/null || true
		printf "\n"
		if [ -z "$USER_PASS" ]; then
			if [ "$DRY_RUN" -eq 1 ]; then
				USER_PASS="dryrunpass"
				break
			fi
			printf "${RED}Password cannot be empty. Please enter a valid password.${RESET}\n"
			continue
		fi
		printf "Confirm password for %s: " "$USERNAME_INPUT"
		stty -echo 2>/dev/null || true
		read -r USER_PASS_CONFIRM || USER_PASS_CONFIRM=""
		stty echo 2>/dev/null || true
		printf "\n"
		if [ "$USER_PASS" = "$USER_PASS_CONFIRM" ]; then
			break
		fi
		printf "${RED}Passwords do not match. Please try again.${RESET}\n"
	done
fi

ROOT_PASS=""
if [ -n "$CFG_ROOT_PASS" ]; then
	ROOT_PASS="$CFG_ROOT_PASS"
elif [ "$CFG_LOCK_ROOT" = "yes" ] || [ "$CFG_LOCK_ROOT" = "true" ] || [ -n "$CONFIG_ARG" ]; then
	ROOT_PASS=""
else
	# Root account setup (lock or set password)
	printf "Lock root account (recommended for sudo setups)? [Y/n]: "
	read -r LOCK_ROOT_INPUT || LOCK_ROOT_INPUT="Y"
	[ -z "$LOCK_ROOT_INPUT" ] && LOCK_ROOT_INPUT="Y"

	case "$LOCK_ROOT_INPUT" in
		[nN]|[nN][oO])
			while true; do
				printf "Enter root password: "
				stty -echo 2>/dev/null || true
				read -r ROOT_PASS || ROOT_PASS=""
				stty echo 2>/dev/null || true
				printf "\n"
				if [ -z "$ROOT_PASS" ]; then
					if [ "$DRY_RUN" -eq 1 ]; then
						ROOT_PASS="dryrunroot"
						break
					fi
					printf "${RED}Root password cannot be empty.${RESET}\n"
					continue
				fi
				printf "Confirm root password: "
				stty -echo 2>/dev/null || true
				read -r ROOT_PASS_CONFIRM || ROOT_PASS_CONFIRM=""
				stty echo 2>/dev/null || true
				printf "\n"
				if [ "$ROOT_PASS" = "$ROOT_PASS_CONFIRM" ]; then
					break
				fi
				printf "${RED}Passwords do not match. Please try again.${RESET}\n"
			done
			;;
		*)
			ROOT_PASS=""
			;;
	esac
fi

# Alongside mode safety pre-checks and EFI/free-space discovery
check_alongside_safety() {
	disk="$1"

	# 1. Boot mode must be UEFI
	if [ "$IS_UEFI" -ne 1 ]; then
		printf "${RED}Error: Alongside / dual-boot installation requires UEFI mode.${RESET}\n" >&2
		printf "  Legacy BIOS (MBR) cannot safely share an existing disk without risking bootloader corruption.\n" >&2
		exit 1
	fi

	# 2. Check disk partition table
	ptable=$(parted -s "$disk" print 2>/dev/null | grep -i "^Partition Table:" | awk '{print $3}' || echo "")
	if [ "$ptable" = "msdos" ]; then
		printf "${RED}Error: Target disk %s has an MBR (msdos) partition table.${RESET}\n" "$disk" >&2
		printf "  Alongside installation requires GPT. Convert the disk to GPT or wipe the entire drive.\n" >&2
		exit 1
	fi

	# 3. Scan for BitLocker
	for part in $(lsblk -l -n -o NAME,TYPE "$disk" 2>/dev/null | grep 'part' | awk '{print $1}'); do
		pdev="/dev/$part"
		[ -b "$pdev" ] || continue
		ptype=$(blkid -s TYPE -o value "$pdev" 2>/dev/null || true)
		case "$ptype" in
			*BitLocker*|*bitlocker*)
				printf "${RED}Error: Windows BitLocker encryption detected on %s.${RESET}\n" "$pdev" >&2
				printf "  BitLocker locks partition structures and prevents safe alongside installation.\n" >&2
				printf "  Please boot into Windows, suspend or disable BitLocker, and re-run installer.\n" >&2
				exit 1
				;;
		esac
	done

	# 4. Scan NTFS partitions for Fast Startup / Hibernation
	for part in $(lsblk -l -n -o NAME,TYPE "$disk" 2>/dev/null | grep 'part' | awk '{print $1}'); do
		pdev="/dev/$part"
		[ -b "$pdev" ] || continue
		ptype=$(blkid -s TYPE -o value "$pdev" 2>/dev/null || true)
		if [ "$ptype" = "ntfs" ]; then
			ntfs_check=""
			if command -v ntfsfix >/dev/null 2>&1; then
				ntfs_check=$(ntfsfix -n "$pdev" 2>&1 || true)
			fi
			if echo "$ntfs_check" | grep -qiE "hibernated|dirty|fast startup"; then
				printf "${RED}Error: Windows Fast Startup or Hibernation is active on %s.${RESET}\n" "$pdev" >&2
				printf "  The NTFS filesystem was not cleanly unmounted. Modifying partitions may corrupt data.\n" >&2
				printf "  Please boot into Windows, disable Fast Startup in Power Options, shut down cleanly, and restart.\n" >&2
				exit 1
			fi
		fi
	done

	# 5. Discover existing ESP
	EXISTING_ESP=""
	for part in $(lsblk -l -n -o NAME,TYPE "$disk" 2>/dev/null | grep 'part' | awk '{print $1}'); do
		pdev="/dev/$part"
		[ -b "$pdev" ] || continue
		btype=$(blkid -s TYPE -o value "$pdev" 2>/dev/null || true)
		if [ "$btype" = "vfat" ]; then
			EXISTING_ESP="$pdev"
			break
		fi
	done

	if [ -z "$EXISTING_ESP" ] && [ "$DRY_RUN" -eq 0 ]; then
		printf "${RED}Error: No existing EFI System Partition (ESP) found on %s.${RESET}\n" "$disk" >&2
		printf "  Alongside installation requires an existing ESP to preserve existing OS boot entries.\n" >&2
		exit 1
	fi
	PART_EFI="${EXISTING_ESP:-${disk}p1}"
}

find_unallocated_slice() {
	disk="$1"
	parted -s "$disk" unit MiB print free 2>/dev/null | awk '
	/Free Space/ {
		start = $1; end = $2; sz = $3;
		sub(/MiB/, "", sz);
		sub(/MiB/, "", start);
		sub(/MiB/, "", end);
		if (sz+0 > max_sz) {
			max_sz = sz+0;
			best_start = start;
			best_end = end;
		}
	}
	END {
		if (max_sz >= 20480) {
			print best_start, best_end, max_sz;
		}
	}'
}

if [ "$DUAL_BOOT" -eq 1 ] && [ "$DRY_RUN" -eq 0 ]; then
	check_alongside_safety "$TARGET_DISK"
fi

# ── STEP 5: Review & Final Confirmation ────────────────────────────────────
FS_DISPLAY="Btrfs (subvolumes @, @home, @snapshots, compress=zstd:1)"
[ "$FS_CHOICE" = "2" ] && FS_DISPLAY="ext4 (defaults, noatime)"

BOOT_DISPLAY="GRUB EFI (UEFI 64-bit)"
[ "$IS_UEFI" -eq 0 ] && BOOT_DISPLAY="GRUB MBR (Legacy BIOS)"

ENCRYPT_DISPLAY="None"
[ "$ENCRYPT_DISK" -eq 1 ] && ENCRYPT_DISPLAY="LUKS2 (argon2id)"

ROOT_STATUS="Locked (sudo only via wheel group)"
[ -n "$ROOT_PASS" ] && ROOT_STATUS="Password set"

INSTALL_MODE_DISPLAY="Clean Install (Wipe entire disk)"
[ "$DUAL_BOOT" -eq 1 ] && INSTALL_MODE_DISPLAY="Install Alongside (Preserve existing OS & partition table)"

printf "\n${WHITE}╭──────────────────────────────────────────────────────────────╮
│                   INSTALLATION SPECIFICATION                 │
├──────────────────────────────────────────────────────────────┤${RESET}\n"
printf "${WHITE}│${RESET} Target Disk:    ${BOLD}%-44s${RESET} ${WHITE}│${RESET}\n" "$TARGET_DISK"
printf "${WHITE}│${RESET} Install Mode:   ${BOLD}%-44s${RESET} ${WHITE}│${RESET}\n" "$INSTALL_MODE_DISPLAY"
printf "${WHITE}│${RESET} Bootloader:     ${BOLD}%-44s${RESET} ${WHITE}│${RESET}\n" "$BOOT_DISPLAY"
printf "${WHITE}│${RESET} Encryption:     ${BOLD}%-44s${RESET} ${WHITE}│${RESET}\n" "$ENCRYPT_DISPLAY"
printf "${WHITE}│${RESET} Filesystem:     ${BOLD}%-44s${RESET} ${WHITE}│${RESET}\n" "$FS_DISPLAY"
printf "${WHITE}│${RESET} Hostname:       ${BOLD}%-44s${RESET} ${WHITE}│${RESET}\n" "$HOSTNAME_INPUT"
printf "${WHITE}│${RESET} User Account:   ${BOLD}%-44s${RESET} ${WHITE}│${RESET}\n" "$USERNAME_INPUT (wheel, toren-hw)"
printf "${WHITE}│${RESET} Root Account:   ${BOLD}%-44s${RESET} ${WHITE}│${RESET}\n" "$ROOT_STATUS"
printf "${WHITE}╰──────────────────────────────────────────────────────────────╯${RESET}\n\n"

if [ "$DRY_RUN" -eq 1 ]; then
	printf "${YELLOW}${BOLD}DRY RUN SUMMARY:${RESET}\n"
	if [ "$DUAL_BOOT" -eq 1 ]; then
		printf "  1. Install Alongside: preserve existing partition table and other operating systems\n"
		printf "  2. Pre-flight verification: UEFI verified, BitLocker inactive, NTFS clean, ESP preserved at %s\n" "${PART_EFI:-${TARGET_DISK}p1}"
		printf "  3. Allocate partition(s) into contiguous unallocated free space on %s\n" "$TARGET_DISK"
		if [ "$ENCRYPT_DISK" -eq 1 ]; then
			printf "     Format unencrypted /boot partition and root partition with LUKS2 (argon2id, label TOREN_CRYPT)\n"
			printf "     Open LUKS container as /dev/mapper/toren_root\n"
		fi
		printf "  4. Format root filesystem with %s (never wipe disk or existing partitions)\n" "$FS_DISPLAY"
		if [ "$FS_CHOICE" != "2" ]; then
			printf "     Create Btrfs subvolumes: @ -> /, @home -> /home, @snapshots -> /.snapshots\n"
		fi
		printf "  5. Deploy rootfs via rsync to /mnt/toren-target\n"
		printf "  6. Configure fstab with persistent UUIDs, preserve ESP without formatting\n"
		printf "  7. Set credentials: user '%s' password, root: %s\n" "$USERNAME_INPUT" "$ROOT_STATUS"
		printf "  8. Configure GRUB with os-prober enabled (GRUB_DISABLE_OS_PROBER=false) to detect existing OS\n"
		printf "  9. Install GRUB EFI to EFI/Toren without overwriting fallback EFI/BOOT/BOOTX64.EFI\n\n"
	elif [ "$ENCRYPT_DISK" -eq 1 ]; then
		if [ "$IS_UEFI" -eq 1 ]; then
			printf "  1. Partition %s (GPT: 512MB ESP, 1GB unencrypted /boot, LUKS2 root)\n" "$TARGET_DISK"
		else
			printf "  1. Partition %s (MBR: 1GB unencrypted /boot with boot flag, LUKS2 root)\n" "$TARGET_DISK"
		fi
		printf "  2. Format root partition with LUKS2 (argon2id, label TOREN_CRYPT)\n"
		printf "  3. Open LUKS container as /dev/mapper/toren_root\n"
		printf "  4. Format /boot with ext4 and /dev/mapper/toren_root with %s\n" "$FS_DISPLAY"
		if [ "$FS_CHOICE" != "2" ]; then
			printf "     Create Btrfs subvolumes: @ -> /, @home -> /home, @snapshots -> /.snapshots\n"
		fi
		printf "  5. Deploy rootfs via rsync to /mnt/toren-target\n"
		printf "  6. Configure /etc/crypttab, dracut crypt module, fstab with persistent UUIDs, and GRUB rd.luks.uuid\n"
		printf "  7. Set credentials: user '%s' password, root: %s\n" "$USERNAME_INPUT" "$ROOT_STATUS"
		printf "  8. Remove live-only configs (agetty autologin, NOPASSWD, live session files)\n"
		printf "  9. Run post-installation verification check (verifying /etc/crypttab and initramfs crypt support).\n\n"
	else
		printf "  1. Partition %s (GPT for UEFI with 512MB ESP, or MBR for BIOS)\n" "$TARGET_DISK"
		printf "  2. Format with %s\n" "$FS_DISPLAY"
		if [ "$FS_CHOICE" != "2" ]; then
			printf "     Create Btrfs subvolumes: @ -> /, @home -> /home, @snapshots -> /.snapshots\n"
		fi
		printf "  3. Deploy rootfs via rsync to /mnt/toren-target\n"
		printf "  4. Configure fstab with persistent UUIDs, regenerate initramfs, install GRUB\n"
		printf "  5. Set credentials: user '%s' password, root: %s\n" "$USERNAME_INPUT" "$ROOT_STATUS"
		printf "  6. Remove live-only configs (agetty autologin, NOPASSWD, live session files)\n"
		printf "  7. Run post-installation verification check.\n\n"
	fi
	printf "${GREEN}Dry-run completed successfully.${RESET}\n"
	exit 0
fi

if [ -z "$CONFIG_ARG" ]; then
	if [ "$DUAL_BOOT" -eq 1 ]; then
		printf "${YELLOW}${BOLD}⚠  Dual-Boot Confirmation: Installing alongside existing operating systems on %s.${RESET}\n" "$TARGET_DISK"
		printf "${WHITE}Type ${BOLD}'YES'${RESET}${WHITE} to partition free space and install Toren Linux: ${RESET}"
	else
		printf "${RED}${BOLD}⚠  CRITICAL WARNING: ALL EXISTING DATA ON %s WILL BE ERASED!${RESET}\n" "$TARGET_DISK"
		printf "${WHITE}Type ${BOLD}'YES'${RESET}${WHITE} to format disk and install Toren Linux: ${RESET}"
	fi
	read -r CONFIRM
	if [ "$CONFIRM" != "YES" ]; then
		printf "\n${DIM}Installation cancelled by user.${RESET}\n"
		exit 0
	fi
fi

# ── EXECUTION: Partitioning, Formatting, Deployment ────────────────────────
printf "\n${WHITE}${BOLD}[1/5] Partitioning target disk (%s)...${RESET}\n" "$TARGET_DISK"

if [ "$DUAL_BOOT" -eq 1 ]; then
	# Alongside mode: Never wipefs, never mklabel!
	slice_info=$(find_unallocated_slice "$TARGET_DISK" || echo "")
	slice_start=$(echo "$slice_info" | awk '{print $1}')
	slice_end=$(echo "$slice_info" | awk '{print $2}')
	if [ -z "$slice_start" ] || [ -z "$slice_end" ]; then
		echo "ERROR: Could not find contiguous unallocated free space (>= 20GB) on $TARGET_DISK" >&2
		exit 1
	fi

	if [ "$ENCRYPT_DISK" -eq 1 ]; then
		boot_end=$((slice_start + 1024))
		parted -s "$TARGET_DISK" mkpart "TOREN_BOOT" ext4 "${slice_start}MiB" "${boot_end}MiB"
		parted -s "$TARGET_DISK" mkpart "TorenRoot" "${boot_end}MiB" "${slice_end}MiB"
		# Find newly created partition names
		new_parts=$(lsblk -l -n -o NAME,TYPE "$TARGET_DISK" 2>/dev/null | grep 'part' | awk '{print $1}' | tail -n2)
		p_boot_name=$(echo "$new_parts" | head -n1)
		p_root_name=$(echo "$new_parts" | tail -n1)
		PART_BOOT="/dev/$p_boot_name"
		PART_ROOT_RAW="/dev/$p_root_name"
	else
		parted -s "$TARGET_DISK" mkpart "TorenRoot" "${slice_start}MiB" "${slice_end}MiB"
		new_part_name=$(lsblk -l -n -o NAME,TYPE "$TARGET_DISK" 2>/dev/null | grep 'part' | awk '{print $1}' | tail -n1)
		PART_ROOT_RAW="/dev/$new_part_name"
		PART_BOOT=""
	fi
else
	# Clean install: wipe entire disk
	wipefs -a "$TARGET_DISK"

	if [ "$ENCRYPT_DISK" -eq 1 ]; then
		if [ "$IS_UEFI" -eq 1 ]; then
			parted -s "$TARGET_DISK" mklabel gpt
			parted -s "$TARGET_DISK" mkpart "ESP" fat32 1MiB 513MiB
			parted -s "$TARGET_DISK" set 1 esp on
			parted -s "$TARGET_DISK" mkpart "TOREN_BOOT" ext4 513MiB 1537MiB
			parted -s "$TARGET_DISK" mkpart "TorenRoot" 1537MiB 100%
		else
			parted -s "$TARGET_DISK" mklabel msdos
			parted -s "$TARGET_DISK" mkpart primary ext4 1MiB 1025MiB
			parted -s "$TARGET_DISK" set 1 boot on
			parted -s "$TARGET_DISK" mkpart primary 1025MiB 100%
		fi
	else
		if [ "$IS_UEFI" -eq 1 ]; then
			parted -s "$TARGET_DISK" mklabel gpt
			parted -s "$TARGET_DISK" mkpart "ESP" fat32 1MiB 513MiB
			parted -s "$TARGET_DISK" set 1 esp on
			parted -s "$TARGET_DISK" mkpart "TorenRoot" 513MiB 100%
		else
			parted -s "$TARGET_DISK" mklabel msdos
			parted -s "$TARGET_DISK" mkpart primary 1MiB 100%
			parted -s "$TARGET_DISK" set 1 boot on
		fi
	fi

	# Resolve partition names for clean install
	if [ "$ENCRYPT_DISK" -eq 1 ]; then
		if [ "$IS_UEFI" -eq 1 ]; then
			if [ -b "${TARGET_DISK}p1" ]; then
				PART_EFI="${TARGET_DISK}p1"
				PART_BOOT="${TARGET_DISK}p2"
				PART_ROOT_RAW="${TARGET_DISK}p3"
			else
				PART_EFI="${TARGET_DISK}1"
				PART_BOOT="${TARGET_DISK}2"
				PART_ROOT_RAW="${TARGET_DISK}3"
			fi
		else
			if [ -b "${TARGET_DISK}p1" ]; then
				PART_BOOT="${TARGET_DISK}p1"
				PART_ROOT_RAW="${TARGET_DISK}p2"
			else
				PART_BOOT="${TARGET_DISK}1"
				PART_ROOT_RAW="${TARGET_DISK}2"
			fi
			PART_EFI=""
		fi
	else
		if [ "$IS_UEFI" -eq 1 ]; then
			if [ -b "${TARGET_DISK}p1" ]; then
				PART_EFI="${TARGET_DISK}p1"
				PART_ROOT_RAW="${TARGET_DISK}p2"
			else
				PART_EFI="${TARGET_DISK}1"
				PART_ROOT_RAW="${TARGET_DISK}2"
			fi
		else
			if [ -b "${TARGET_DISK}p1" ]; then
				PART_ROOT_RAW="${TARGET_DISK}p1"
			else
				PART_ROOT_RAW="${TARGET_DISK}1"
			fi
			PART_EFI=""
		fi
		PART_BOOT=""
	fi
fi

sleep 1
printf "${GREEN}✓ Partition table generated.${RESET}\n"

printf "${WHITE}${BOLD}[2/5] Formatting filesystems...${RESET}\n"
if [ "$IS_UEFI" -eq 1 ]; then
	if [ "$DUAL_BOOT" -eq 1 ]; then
		printf "  Dual-boot mode: preserving existing ESP at %s\n" "$PART_EFI"
	else
		mkfs.vfat -F 32 -n "TOREN_ESP" "$PART_EFI"
	fi
fi

if [ "$ENCRYPT_DISK" -eq 1 ]; then
	printf "  Formatting separate /boot partition (%s) with ext4...\n" "$PART_BOOT"
	mkfs.ext4 -F -L "TOREN_BOOT" "$PART_BOOT"

	printf "  Encrypting root partition (%s) with LUKS2 (argon2id)...\n" "$PART_ROOT_RAW"
	printf '%s' "$LUKS_PASS" | cryptsetup luksFormat --type luks2 --pbkdf argon2id --label TOREN_CRYPT --batch-mode --key-file=- "$PART_ROOT_RAW"
	printf '%s' "$LUKS_PASS" | cryptsetup open --key-file=- "$PART_ROOT_RAW" toren_root
	PART_ROOT_FS="/dev/mapper/toren_root"
else
	PART_ROOT_FS="$PART_ROOT_RAW"
fi

mkdir -p "$TARGET_DIR"

if [ "$FS_CHOICE" != "2" ] && command -v mkfs.btrfs >/dev/null 2>&1; then
	printf "  Formatting root with Btrfs and creating subvolumes (@, @home, @snapshots)...\n"
	mkfs.btrfs -f -L "TOREN_ROOT" "$PART_ROOT_FS"
	FS_TYPE="btrfs"
	BTRFS_BASE_OPTS="defaults,noatime,compress=zstd:1"
	
	# Temporary mount to create subvolumes
	TMP_BTRFS="/mnt/toren-subvols.$$"
	mkdir -p "$TMP_BTRFS"
	mount -t btrfs "$PART_ROOT_FS" "$TMP_BTRFS"
	btrfs subvolume create "$TMP_BTRFS/@"
	btrfs subvolume create "$TMP_BTRFS/@home"
	btrfs subvolume create "$TMP_BTRFS/@snapshots"
	umount "$TMP_BTRFS"
	rmdir "$TMP_BTRFS"

	# Mount subvolumes into target tree
	mount -t btrfs -o "$BTRFS_BASE_OPTS,subvol=@" "$PART_ROOT_FS" "$TARGET_DIR"
	mkdir -p "$TARGET_DIR/home" "$TARGET_DIR/.snapshots"
	mount -t btrfs -o "$BTRFS_BASE_OPTS,subvol=@home" "$PART_ROOT_FS" "$TARGET_DIR/home"
	mount -t btrfs -o "$BTRFS_BASE_OPTS,subvol=@snapshots" "$PART_ROOT_FS" "$TARGET_DIR/.snapshots"
else
	printf "  Formatting root with ext4...\n"
	mkfs.ext4 -F -L "TOREN_ROOT" "$PART_ROOT_FS"
	FS_TYPE="ext4"
	mount -t ext4 -o defaults,noatime "$PART_ROOT_FS" "$TARGET_DIR"
fi

if [ "$ENCRYPT_DISK" -eq 1 ]; then
	mkdir -p "$TARGET_DIR/boot"
	mount "$PART_BOOT" "$TARGET_DIR/boot"
fi

if [ "$IS_UEFI" -eq 1 ]; then
	mkdir -p "$TARGET_DIR/boot/efi"
	mount "$PART_EFI" "$TARGET_DIR/boot/efi"
fi
printf "${GREEN}✓ Filesystems ready and mounted.${RESET}\n"

printf "${WHITE}${BOLD}[3/5] Deploying Toren Linux system image (rootfs)...${RESET}\n"
rsync -aAX --info=progress2 \
	--exclude='/dev/*' \
	--exclude='/proc/*' \
	--exclude='/sys/*' \
	--exclude='/tmp/*' \
	--exclude='/run/*' \
	--exclude='/mnt/*' \
	--exclude='/media/*' \
	--exclude='/lost+found' \
	--exclude='/var/cache/*' \
	/ "$TARGET_DIR/"
printf "${GREEN}✓ System image copied.${RESET}\n"

printf "${WHITE}${BOLD}[4/5] Configuring system services, fstab, and identity...${RESET}\n"
ROOT_UUID=$(blkid -s UUID -o value "$PART_ROOT_FS")
if [ -z "$ROOT_UUID" ]; then
	echo "ERROR: Failed to retrieve UUID for $PART_ROOT_FS" >&2
	exit 1
fi

if [ "$ENCRYPT_DISK" -eq 1 ]; then
	BOOT_UUID=$(blkid -s UUID -o value "$PART_BOOT")
	if [ -z "$BOOT_UUID" ]; then
		echo "ERROR: Failed to retrieve UUID for $PART_BOOT" >&2
		exit 1
	fi
	LUKS_UUID=$(blkid -s UUID -o value "$PART_ROOT_RAW")
	if [ -z "$LUKS_UUID" ]; then
		echo "ERROR: Failed to retrieve UUID for $PART_ROOT_RAW" >&2
		exit 1
	fi
	# Configure /etc/crypttab with UUID and discard support
	echo "toren_root UUID=$LUKS_UUID none luks,discard" > "$TARGET_DIR/etc/crypttab"
	chmod 0600 "$TARGET_DIR/etc/crypttab"

	# Configure dracut crypt module drop-in
	mkdir -p "$TARGET_DIR/etc/dracut.conf.d"
	cat >"$TARGET_DIR/etc/dracut.conf.d/40-toren-crypt.conf" <<'EOF'
add_dracutmodules+=" crypt "
install_items+=" /etc/crypttab "
EOF
fi

if [ "$IS_UEFI" -eq 1 ]; then
	EFI_UUID=$(blkid -s UUID -o value "$PART_EFI")
	if [ -z "$EFI_UUID" ]; then
		echo "ERROR: Failed to retrieve UUID for $PART_EFI" >&2
		exit 1
	fi
	if [ "$FS_TYPE" = "btrfs" ]; then
		cat >"$TARGET_DIR/etc/fstab" <<EOF
# /etc/fstab: static file system information for Toren Linux
UUID=$ROOT_UUID / btrfs defaults,noatime,compress=zstd:1,subvol=@ 0 0
UUID=$ROOT_UUID /home btrfs defaults,noatime,compress=zstd:1,subvol=@home 0 0
UUID=$ROOT_UUID /.snapshots btrfs defaults,noatime,compress=zstd:1,subvol=@snapshots 0 0
EOF
	else
		cat >"$TARGET_DIR/etc/fstab" <<EOF
# /etc/fstab: static file system information for Toren Linux
UUID=$ROOT_UUID / ext4 defaults,noatime 0 1
EOF
	fi
	if [ "$ENCRYPT_DISK" -eq 1 ]; then
		cat >>"$TARGET_DIR/etc/fstab" <<EOF
UUID=$BOOT_UUID /boot ext4 defaults,noatime 0 2
EOF
	fi
	cat >>"$TARGET_DIR/etc/fstab" <<EOF
UUID=$EFI_UUID /boot/efi vfat defaults,noatime,umask=0077 0 2
tmpfs /tmp tmpfs defaults,nosuid,nodev 0 0
EOF
else
	if [ "$FS_TYPE" = "btrfs" ]; then
		cat >"$TARGET_DIR/etc/fstab" <<EOF
# /etc/fstab: static file system information for Toren Linux
UUID=$ROOT_UUID / btrfs defaults,noatime,compress=zstd:1,subvol=@ 0 0
UUID=$ROOT_UUID /home btrfs defaults,noatime,compress=zstd:1,subvol=@home 0 0
UUID=$ROOT_UUID /.snapshots btrfs defaults,noatime,compress=zstd:1,subvol=@snapshots 0 0
EOF
	else
		cat >"$TARGET_DIR/etc/fstab" <<EOF
# /etc/fstab: static file system information for Toren Linux
UUID=$ROOT_UUID / ext4 defaults,noatime 0 1
EOF
	fi
	if [ "$ENCRYPT_DISK" -eq 1 ]; then
		cat >>"$TARGET_DIR/etc/fstab" <<EOF
UUID=$BOOT_UUID /boot ext4 defaults,noatime 0 2
EOF
	fi
	cat >>"$TARGET_DIR/etc/fstab" <<EOF
tmpfs /tmp tmpfs defaults,nosuid,nodev 0 0
EOF
fi

# Set hostname and OS identity
printf '%s\n' "$HOSTNAME_INPUT" > "$TARGET_DIR/etc/hostname"

cat >"$TARGET_DIR/etc/os-release" <<'EOF'
NAME="Toren Linux"
PRETTY_NAME="Toren Linux Beta 001"
ID=toren
ID_LIKE="void linux"
VERSION_ID="0.1"
VERSION_CODENAME=beta-001
HOME_URL="https://github.com/mohitjoer/Toren-Linux"
SUPPORT_URL="https://github.com/mohitjoer/Toren-Linux/issues"
BUG_REPORT_URL="https://github.com/mohitjoer/Toren-Linux/issues"
EOF

# Purge live installer artifacts and markers
rm -f "$TARGET_DIR/etc/toren-live" "$TARGET_DIR/etc/toren-live"
rm -f "$TARGET_DIR/usr/bin/toren-installer"
rm -f "$TARGET_DIR/usr/bin/void-installer"
rm -f "$TARGET_DIR/usr/share/applications/toren-installer.desktop" "$TARGET_DIR/usr/share/applications/toren-installer.desktop"
rm -f "$TARGET_DIR/home"/*/Desktop/toren-installer.desktop "$TARGET_DIR/home"/*/Desktop/toren-installer.desktop 2>/dev/null || true
rm -f "$TARGET_DIR/etc/skel/Desktop/toren-installer.desktop" "$TARGET_DIR/etc/skel/Desktop/toren-installer.desktop" 2>/dev/null || true

# Ensure dedicated hardware and docker groups exist on target
chroot "$TARGET_DIR" getent group toren-hw >/dev/null 2>&1 || chroot "$TARGET_DIR" groupadd -r toren-hw || true
chroot "$TARGET_DIR" getent group toren-hw >/dev/null 2>&1 || chroot "$TARGET_DIR" groupadd -r toren-hw
chroot "$TARGET_DIR" getent group docker >/dev/null 2>&1 || chroot "$TARGET_DIR" groupadd -r docker || true

# Record live user hash from running environment to verify against later
LIVE_SHADOW_HASH=$(grep -E '^(toren|toren):' /etc/shadow 2>/dev/null | head -n 1 | cut -d: -f2 || true)

# Completely purge live user account, credentials, and live home directory from target
for live_user in toren toren; do
	if chroot "$TARGET_DIR" id -u "$live_user" >/dev/null 2>&1 && [ "$live_user" != "$USERNAME_INPUT" ]; then
		chroot "$TARGET_DIR" userdel -r "$live_user" >/dev/null 2>&1 || true
	fi
	if [ "$live_user" != "$USERNAME_INPUT" ]; then
		rm -rf "$TARGET_DIR/home/$live_user"
	fi
done

# User account creation on target
if ! chroot "$TARGET_DIR" id -u "$USERNAME_INPUT" >/dev/null 2>&1; then
	chroot "$TARGET_DIR" useradd -m -s /bin/zsh -G wheel,_seatd,video,audio,input,storage,network,toren-hw,docker "$USERNAME_INPUT"
else
	chroot "$TARGET_DIR" usermod -aG wheel,_seatd,video,audio,input,storage,network,toren-hw,docker "$USERNAME_INPUT"
fi

# Materialize toren configuration for newly created user
chroot "$TARGET_DIR" su - "$USERNAME_INPUT" -c "toren materialize" 2>/dev/null || true
if [ -f "$TARGET_DIR/etc/skel/.zshrc" ] && [ ! -f "$TARGET_DIR/home/$USERNAME_INPUT/.zshrc" ]; then
	cp "$TARGET_DIR/etc/skel/.zshrc" "$TARGET_DIR/home/$USERNAME_INPUT/.zshrc" 2>/dev/null || true
	chroot "$TARGET_DIR" chown -R "$USERNAME_INPUT:$USERNAME_INPUT" "/home/$USERNAME_INPUT" 2>/dev/null || true
fi

# Set password for primary user with explicit SHA512 encryption or pre-computed hash
if [ -n "${USER_PASS_HASH:-}" ]; then
	chroot "$TARGET_DIR" usermod -p "$USER_PASS_HASH" "$USERNAME_INPUT"
else
	echo "$USERNAME_INPUT:$USER_PASS" | chroot "$TARGET_DIR" chpasswd -c SHA512
fi

# Localization and keymap configuration
if [ -n "${CFG_TIMEZONE:-}" ] && [ -f "$TARGET_DIR/usr/share/zoneinfo/$CFG_TIMEZONE" ]; then
	ln -sf "/usr/share/zoneinfo/$CFG_TIMEZONE" "$TARGET_DIR/etc/localtime"
	sed -i "s/^#\?TIMEZONE=.*/TIMEZONE=\"$CFG_TIMEZONE\"/" "$TARGET_DIR/etc/rc.conf" 2>/dev/null || true
fi
if [ -n "${CFG_KEYMAP:-}" ]; then
	sed -i "s/^#\?KEYMAP=.*/KEYMAP=\"$CFG_KEYMAP\"/" "$TARGET_DIR/etc/rc.conf" 2>/dev/null || true
fi
if [ -n "${CFG_LOCALE:-}" ]; then
	echo "LANG=\"$CFG_LOCALE\"" > "$TARGET_DIR/etc/locale.conf"
fi

# Configure root password or lock root account
if [ -n "$ROOT_PASS" ]; then
	echo "root:$ROOT_PASS" | chroot "$TARGET_DIR" chpasswd -c SHA512
else
	chroot "$TARGET_DIR" usermod -p '!' root
fi

# Clean up any shadow/passwd backup files that might contain the live hash
rm -f "$TARGET_DIR/etc/shadow-" "$TARGET_DIR/etc/passwd-" "$TARGET_DIR/etc/group-" "$TARGET_DIR/etc/gshadow-"

# Purge live-only configurations, live sudoers, polkit rules, and autologin from installed system
rm -f "$TARGET_DIR/etc/sudoers.d/99-toren-live" "$TARGET_DIR/etc/sudoers.d/99-void-live" "$TARGET_DIR/etc/sudoers.d/"*live*
rm -f "$TARGET_DIR/etc/polkit-1/rules.d/toren-live.rules"
rm -f "$TARGET_DIR/etc/default/live.conf"
rm -f "$TARGET_DIR/etc/sv/agetty-tty1/conf"
rm -f "$TARGET_DIR/etc/profile.d/toren-autologin.sh"
rm -f "$TARGET_DIR/home/$USERNAME_INPUT/.hushlogin" 2>/dev/null || true
rm -f "$TARGET_DIR/etc/skel/.hushlogin" 2>/dev/null || true
rm -rf "$TARGET_DIR/tmp"/* "$TARGET_DIR/var/tmp"/*

# Replace live MOTD, issue, and issue.net with installed system banners
cat >"$TARGET_DIR/etc/motd" <<'MOTD'
       #                    #         Toren Linux Beta 001
     %##                    ##%       https://torenos.dev
    ####   %%%%%%%%%%%%%%   ####
  ######% %%%%%%%%%%%%%%%% %######    Robust. Minimal. Kinetic.
%###########% %%%%%%%%  ###########%
 %###########          ###########%
     ########  @@@@@@  ########
   %% ########%      %######## %%%
   %%  %%##################%%% %%%
   %% ##     %########%     ## %%%
   %% ####%%%%%      %%%%%#### %%
       %####################%
      %%  %##############%  %%
       ### ############## %##
       %##  ############  ###
        ###  %        %  ###
         %################%
           %############%
             ##########
MOTD

cat >"$TARGET_DIR/etc/issue" <<'ISSUE'
       #                    #         Toren Linux Beta 001 (\m)
     %##                    ##%       Kernel \r on an \m (\l)
    ####   %%%%%%%%%%%%%%   ####
  ######% %%%%%%%%%%%%%%%% %######    Robust. Minimal. Kinetic.
%###########% %%%%%%%%  ###########%  https://torenos.dev
 %###########          ###########%
     ########  @@@@@@  ########
   %% ########%      %######## %%%
   %%  %%##################%%% %%%
   %% ##     %########%     ## %%%
   %% ####%%%%%      %%%%%#### %%
       %####################%
      %%  %##############%  %%
       ### ############## %##
       %##  ############  ###
        ###  %        %  ###
         %################%
           %############%
             ##########
ISSUE
cp "$TARGET_DIR/etc/motd" "$TARGET_DIR/etc/issue.net"

# Strict sudoers configuration: require password for wheel group (NO NOPASSWD)
mkdir -p "$TARGET_DIR/etc/sudoers.d"
chmod 0750 "$TARGET_DIR/etc/sudoers.d"
echo '%wheel ALL=(ALL:ALL) ALL' > "$TARGET_DIR/etc/sudoers.d/wheel"
chmod 0440 "$TARGET_DIR/etc/sudoers.d/wheel"
if [ -f "$TARGET_DIR/etc/sudoers" ]; then
	sed -i '/NOPASSWD/d' "$TARGET_DIR/etc/sudoers"
fi

# The session dismisses the boot splash before the compositor claims KMS, and
# plymouthd runs as root, so the wheel group gets exactly these commands without
# a password. Every other sudo call still asks for one. The drop-in name sorts
# after "wheel" on purpose: sudo applies the LAST matching entry, so a file
# sorted before wheel would be overridden by wheel's password rule.
cat >"$TARGET_DIR/etc/sudoers.d/zz-toren-session" <<'SUDOERS'
# Toren's session: the boot splash is dismissed by toren-session before the
# compositor claims KMS. plymouthd runs as root, so these exact commands are
# the only passwordless privileges the session needs. Both spellings are
# granted because sudo resolves a bare name through secure_path, where
# /usr/sbin precedes /usr/bin. The file name sorts after "wheel" because
# sudo applies the last matching sudoers entry.
Cmnd_Alias TOREN_SESSION_PLYMOUTH = /usr/bin/plymouth --quit, /usr/sbin/plymouth --quit, /usr/bin/killall -9 plymouthd, /usr/sbin/killall -9 plymouthd, /usr/bin/pkill -9 -x plymouthd, /usr/sbin/pkill -9 -x plymouthd
%wheel ALL=(root) NOPASSWD: TOREN_SESSION_PLYMOUTH
SUDOERS
chmod 0440 "$TARGET_DIR/etc/sudoers.d/zz-toren-session"

# Enable core system runit services
mkdir -p "$TARGET_DIR/etc/runit/runsvdir/default"
for svc in dbus seatd elogind udevd NetworkManager acpid turnstiled tlp zramen ananicy-cpp chronyd nftables cupsd toren-sysd toren-boot-guard; do
	if [ -d "$TARGET_DIR/etc/sv/$svc" ]; then
		ln -sf "/etc/sv/$svc" "$TARGET_DIR/etc/runit/runsvdir/default/$svc"
	fi
done

# Ensure wpa_supplicant does not run as a standalone service (managed by NetworkManager)
rm -f "$TARGET_DIR/etc/runit/runsvdir/default/wpa_supplicant" "$TARGET_DIR/var/service/wpa_supplicant"

# The quickest Void mirror for this machine (toren mirror; updates keep it).
# Best effort: offline, or no mirror answering, keeps ASURA's default.
if [ -x "$TARGET_DIR/usr/bin/toren" ]; then
	echo "Choosing the quickest package mirror..."
	timeout 60 chroot "$TARGET_DIR" /usr/bin/toren mirror fastest >/dev/null 2>&1 || true
fi

# Layer 3 Tuning: Configure zramen zstd compression
mkdir -p "$TARGET_DIR/etc/sv/zramen"
cat >"$TARGET_DIR/etc/sv/zramen/conf" <<'ZRAMCONF'
export ZRAM_COMP_ALGORITHM=zstd
export ZRAM_SIZE=50
export ZRAM_PRIORITY=32767
ZRAMCONF

# Layer 3 Tuning: Avoid global ld.so.preload (opt-in runner /usr/bin/with-mimalloc provided)
rm -f "$TARGET_DIR/etc/ld.so.preload"

# Configure GRUB bootloader identity
mkdir -p "$TARGET_DIR/etc/default"
cat >"$TARGET_DIR/etc/default/grub" <<'GRUBDEF'
GRUB_DISTRIBUTOR="Toren Linux"
GRUB_DEFAULT=saved
GRUB_SAVEDEFAULT=true
GRUB_TIMEOUT=3
GRUB_CMDLINE_LINUX_DEFAULT="quiet splash loglevel=3 rd.udev.log_level=3 udev.log_level=3 rd.plymouth=1 plymouth.enable=1 vt.global_cursor_default=0 fbcon=nodefer console=ttyS0,115200 console=tty0"
GRUB_DISABLE_OS_PROBER=false
GRUBDEF

if [ "$ENCRYPT_DISK" -eq 1 ]; then
	cat >>"$TARGET_DIR/etc/default/grub" <<EOF
GRUB_CMDLINE_LINUX="rd.luks.uuid=$LUKS_UUID rd.luks.allow-discards"
EOF
fi

# Ensure Plymouth theme and dracut configuration exist on installed target
mkdir -p "$TARGET_DIR/etc/plymouth" "$TARGET_DIR/etc/dracut.conf.d"
cat >"$TARGET_DIR/etc/plymouth/plymouthd.conf" <<'EOF'
[Daemon]
Theme=toren
ShowDelay=0
DeviceTimeout=8
EOF

cat >"$TARGET_DIR/etc/dracut.conf.d/50-plymouth.conf" <<'EOF'
add_dracutmodules+=" plymouth "
add_drivers+=" vboxvideo vmwgfx bochs virtio_gpu qxl simpledrm "
install_optional_items+=" /usr/lib/plymouth/script.so /usr/lib/plymouth/label.so /usr/lib/plymouth/label-pango.so /usr/lib/plymouth/renderers/drm.so /usr/lib/plymouth/renderers/frame-buffer.so /etc/fonts/fonts.conf /etc/fonts/local.conf /usr/share/fonts/TTF/DejaVuSans.ttf /usr/share/fonts/TTF/DejaVuSans-Bold.ttf /usr/share/fonts/TTF/DejaVuSansMono.ttf "
EOF

if [ -d "$TARGET_DIR/usr/share/plymouth/themes/toren" ]; then
	ln -sf /usr/share/plymouth/themes/toren/toren.plymouth "$TARGET_DIR/usr/share/plymouth/themes/default.plymouth" 2>/dev/null || true
fi
chroot "$TARGET_DIR" plymouth-set-default-theme toren 2>/dev/null || true

# Bind mounts for chroot maintenance and hardware detection
mount --bind /dev "$TARGET_DIR/dev"
mount --bind /proc "$TARGET_DIR/proc"
mount --bind /sys "$TARGET_DIR/sys"

# Apply hardware quirks across matched layers if this machine matches known profiles
if [ -x "$TARGET_DIR/usr/bin/toren" ]; then
	chroot "$TARGET_DIR" /usr/bin/toren hardware quirks apply >/dev/null 2>&1 || true
fi

printf "${GREEN}✓ Host, services, and accounts configured.${RESET}\n"

printf "${WHITE}${BOLD}[4.5/5] Regenerating initramfs for installed kernel...${RESET}\n"
KERNEL_PKG=$(chroot "$TARGET_DIR" xbps-query -l 2>/dev/null | grep -E '^ii linux[0-9]' | awk '{print $2}' | cut -d- -f1 | head -n1 || echo "linux")
[ -z "$KERNEL_PKG" ] && KERNEL_PKG="linux"

if chroot "$TARGET_DIR" xbps-reconfigure -f "$KERNEL_PKG" >/dev/null 2>&1; then
	printf "  Regenerated initramfs via xbps-reconfigure.\n"
elif chroot "$TARGET_DIR" dracut --force >/dev/null 2>&1; then
	printf "  Regenerated initramfs via dracut.\n"
fi

# Verify kernel and initramfs exist
KERNEL_IMG=$(ls "$TARGET_DIR/boot"/vmlinuz* 2>/dev/null | head -n1 || true)
INITRD_IMG=$(ls "$TARGET_DIR/boot"/initramfs* 2>/dev/null | head -n1 || true)
if [ -z "$KERNEL_IMG" ] || [ -z "$INITRD_IMG" ]; then
	echo "ERROR: Kernel image or initramfs not found in $TARGET_DIR/boot after regeneration!" >&2
	exit 1
fi
printf "${GREEN}✓ Kernel and initramfs verified in /boot.${RESET}\n"

printf "${WHITE}${BOLD}[5/5] Installing bootloader (%s)...${RESET}\n" "$BOOT_DISPLAY"

# The Toren snapshot GRUB snippet goes into /etc/grub.d before grub-mkconfig
# runs, so the first grub.cfg already has it.
if [ -d "$TARGET_DIR/usr/share/toren/session/grub" ]; then
	mkdir -p "$TARGET_DIR/etc/grub.d"
	cp "$TARGET_DIR/usr/share/toren/session/grub"/* "$TARGET_DIR/etc/grub.d/" 2>/dev/null || true
	chmod +x "$TARGET_DIR/etc/grub.d"/40_toren_snapshots "$TARGET_DIR/etc/grub.d"/40_toren_snapshots 2>/dev/null || true
fi

if [ "$IS_UEFI" -eq 1 ]; then
	if [ -d /sys/firmware/efi/efivars ]; then
		mkdir -p "$TARGET_DIR/sys/firmware/efi/efivars"
		mount -t efivarfs efivarfs "$TARGET_DIR/sys/firmware/efi/efivars" 2>/dev/null || true
	fi
	chroot "$TARGET_DIR" grub-install --target=x86_64-efi --efi-directory=/boot/efi --bootloader-id=Toren --recheck
	mkdir -p "$TARGET_DIR/boot/efi/EFI/BOOT"
	if [ "$DUAL_BOOT" -eq 0 ]; then
		if [ -f "$TARGET_DIR/boot/efi/EFI/Toren/grubx64.efi" ]; then
			cp -f "$TARGET_DIR/boot/efi/EFI/Toren/grubx64.efi" "$TARGET_DIR/boot/efi/EFI/BOOT/BOOTX64.EFI"
		elif [ -f "$TARGET_DIR/boot/efi/EFI/toren/grubx64.efi" ]; then
			cp -f "$TARGET_DIR/boot/efi/EFI/toren/grubx64.efi" "$TARGET_DIR/boot/efi/EFI/BOOT/BOOTX64.EFI"
		fi
	else
		printf "  Dual-boot mode: preserving existing fallback bootloader at EFI/BOOT/BOOTX64.EFI\n"
	fi
	chroot "$TARGET_DIR" grub-mkconfig -o /boot/grub/grub.cfg
else
	chroot "$TARGET_DIR" grub-install --target=i386-pc --recheck "$TARGET_DISK"
	chroot "$TARGET_DIR" grub-mkconfig -o /boot/grub/grub.cfg
fi


printf "${GREEN}✓ Bootloader successfully installed.${RESET}\n"

# ── POST-INSTALLATION VERIFICATION ─────────────────────────────────────────
printf "\n${WHITE}╭──────────────────────────────────────────────────────────────╮
│                POST-INSTALLATION VERIFICATION                │
├──────────────────────────────────────────────────────────────┤${RESET}\n"

V_FSTAB="FAIL"
if grep -q "UUID=" "$TARGET_DIR/etc/fstab" && ! grep -q "UUID= " "$TARGET_DIR/etc/fstab"; then
	V_FSTAB="PASS"
fi

V_GRUB="FAIL"
if [ -s "$TARGET_DIR/boot/grub/grub.cfg" ]; then
	V_GRUB="PASS"
fi

V_KERNEL="FAIL"
if [ -n "$KERNEL_IMG" ] && [ -n "$INITRD_IMG" ]; then
	V_KERNEL="PASS"
fi

V_USER="FAIL"
if chroot "$TARGET_DIR" id -u "$USERNAME_INPUT" >/dev/null 2>&1; then
	TARGET_USER_HASH=$(grep "^${USERNAME_INPUT}:" "$TARGET_DIR/etc/shadow" 2>/dev/null | cut -d: -f2 || true)
	if [ -n "$TARGET_USER_HASH" ] && \
	   [ "$TARGET_USER_HASH" != "!" ] && \
	   [ "$TARGET_USER_HASH" != "*" ] && \
	   [ "$(printf '%s' "$TARGET_USER_HASH" | cut -c1)" != "!" ] && \
	   { [ -z "$LIVE_SHADOW_HASH" ] || [ "$TARGET_USER_HASH" != "$LIVE_SHADOW_HASH" ]; }; then
		V_USER="PASS"
	fi
fi

V_SERVICES="FAIL"
if [ -L "$TARGET_DIR/etc/runit/runsvdir/default/dbus" ] && [ -L "$TARGET_DIR/etc/runit/runsvdir/default/seatd" ]; then
	V_SERVICES="PASS"
fi

V_CRYPT="N/A"
if [ "$ENCRYPT_DISK" -eq 1 ]; then
	V_CRYPT="FAIL"
	if [ -s "$TARGET_DIR/etc/crypttab" ] && grep -q "$LUKS_UUID" "$TARGET_DIR/etc/crypttab"; then
		if lsinitrd "$INITRD_IMG" 2>/dev/null | grep -qE "crypt|cryptsetup"; then
			V_CRYPT="PASS"
		elif [ -f "$TARGET_DIR/etc/dracut.conf.d/40-toren-crypt.conf" ]; then
			V_CRYPT="PASS"
		fi
	fi
fi

printf "${WHITE}│${RESET} [1] /etc/fstab UUIDs:       ${BOLD}%-33s${RESET} ${WHITE}│${RESET}\n" "$V_FSTAB"
printf "${WHITE}│${RESET} [2] /boot/grub/grub.cfg:    ${BOLD}%-33s${RESET} ${WHITE}│${RESET}\n" "$V_GRUB"
printf "${WHITE}│${RESET} [3] Kernel & Initramfs:     ${BOLD}%-33s${RESET} ${WHITE}│${RESET}\n" "$V_KERNEL"
printf "${WHITE}│${RESET} [4] Primary User Created:   ${BOLD}%-33s${RESET} ${WHITE}│${RESET}\n" "$V_USER ($USERNAME_INPUT)"
printf "${WHITE}│${RESET} [5] Runit Services Linked:  ${BOLD}%-33s${RESET} ${WHITE}│${RESET}\n" "$V_SERVICES"
if [ "$ENCRYPT_DISK" -eq 1 ]; then
	printf "${WHITE}│${RESET} [6] LUKS & Crypttab:        ${BOLD}%-33s${RESET} ${WHITE}│${RESET}\n" "$V_CRYPT"
fi
printf "${WHITE}╰──────────────────────────────────────────────────────────────╯${RESET}\n\n"

if [ "$V_FSTAB" != "PASS" ] || [ "$V_GRUB" != "PASS" ] || [ "$V_KERNEL" != "PASS" ] || [ "$V_USER" != "PASS" ] || [ "$V_SERVICES" != "PASS" ] || { [ "$ENCRYPT_DISK" -eq 1 ] && [ "$V_CRYPT" != "PASS" ]; }; then
	printf "${RED}${BOLD}Verification failed. Aborting installation.${RESET}\n" >&2
	exit 1
fi

printf "${GREEN}${BOLD}All post-installation verification checks passed.${RESET}\n"

# Unmount cleanly
if [ "$IS_UEFI" -eq 1 ]; then
	umount "$TARGET_DIR/sys/firmware/efi/efivars" 2>/dev/null || true
	umount "$TARGET_DIR/boot/efi" 2>/dev/null || true
fi
if [ "$ENCRYPT_DISK" -eq 1 ]; then
	umount "$TARGET_DIR/boot" 2>/dev/null || true
fi
umount "$TARGET_DIR/sys"
umount "$TARGET_DIR/proc"
umount "$TARGET_DIR/dev"

if [ "$FS_TYPE" = "btrfs" ]; then
	umount "$TARGET_DIR/home"
	umount "$TARGET_DIR/.snapshots"
fi
umount "$TARGET_DIR"

if [ "$ENCRYPT_DISK" -eq 1 ]; then
	cryptsetup close toren_root 2>/dev/null || true
fi

cat <<EOF

${GREEN}╭──────────────────────────────────────────────────────────────╮
│            ✓ TOREN LINUX Successfully Installed!             │
│                                                              │
│   Toren Linux has been deployed to your storage drive.       │
│   The installer has been automatically deleted.              │
│   You can now safely reboot into your new installation.      │
╰──────────────────────────────────────────────────────────────╯${RESET}

EOF

printf "Would you like to reboot into Toren Linux now? [Y/n]: "
read -r REBOOT_CHOICE || REBOOT_CHOICE="y"
[ -z "$REBOOT_CHOICE" ] && REBOOT_CHOICE="y"

case "$REBOOT_CHOICE" in
	[yY]|[yY][eE][sS])
		printf "\n${WHITE}Rebooting system...${RESET}\n"
		reboot
		;;
	*)
		printf "\n${DIM}Returning to live desktop environment. You may reboot anytime.${RESET}\n\n"
		printf "${WHITE}Press Enter to exit...${RESET}"
		read -r _ || true
		;;
esac
