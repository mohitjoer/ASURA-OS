# Installation, Dual-Boot & Adoption Guide

This document covers all three installation paths in Toren Linux, unattended deployments,
and converting existing Void Linux systems via `toren adopt`.

---

## 1. Installation Paths Overview

Toren Linux supports three distinct ways to get the desktop running:

| Path | Command | Use Case | Disk Changes |
|---|---|---|---|
| **Clean Install** | `toren install` | Dedicated drive or new machine | Erases disk, creates GPT/ESP, Btrfs subvolumes |
| **Install Alongside (Dual-Boot)** | `toren install --alongside` | Coexist with Windows or another Linux | Preserves partition table, uses unallocated free space, reuses ESP |
| **Unattended Install** | `toren install --config FILE` | Automated lab, VM, or CI deployment | Guided by answer file, passwords never logged in plaintext |
| **Adopt Existing Void** | `toren adopt` | Convert existing Void Linux glibc installation | In-place, reversible with `toren adopt --undo`, atomic backup |

---

## 2. Guided Installer (`toren-installer.sh`)

The guided system installer is located at `installation/toren-installer.sh` and launched via
`toren install` or the application launcher shortcut.

### Flag Reference
```sh
toren-installer [options] [TARGET_DISK]

Options:
  --alongside, --dual-boot   Install into contiguous free space alongside existing OS
  --wipe, --clean            Clean install: erase entire target drive (default)
  --dry-run                  Simulate partitioning, formatting and deployment without touching disk
  --config FILE              Unattended installation using documented answer file
  --encrypt                  Enable LUKS2 (argon2id) root filesystem encryption
  --no-encrypt               Disable encryption
  -h, --help                 Show help information
```

---

## 3. Dual-Boot Next to Windows / Other OS (`--alongside`)

Installing alongside another operating system safely coexists on the same drive without
erasing existing data or breaking Windows Boot Manager.

```sh
# Run installer in alongside mode
sudo asura install --alongside /dev/nvme0n1
```

### Safety Rules & Guardrails
1. **UEFI Mode Required**: The installer verifies `IS_UEFI=1`. Legacy BIOS / MBR alongside
   installation is strictly refused because partition table sharing risks bootloader corruption.
2. **GPT Table Required**: Target disk must use GPT. MBR partition tables are refused.
3. **BitLocker Detection**: Partitions are scanned using `blkid`. If Windows BitLocker encryption
   is active, the installer halts with instructions to suspend or disable BitLocker in Windows.
4. **Windows Fast Startup / Dirty NTFS Guard**: All NTFS partitions are checked. If Windows Fast
   Startup or hibernation left the volume dirty, the installer refuses to proceed to prevent
   filesystem corruption, guiding the user to perform a clean shutdown in Windows.
5. **ESP Preservation**: The existing EFI System Partition is discovered and preserved. It is
   **never** formatted with `mkfs.vfat`.
6. **Contiguous Free Space**: Scans for unallocated slices >= 20 GiB. Only the free slice is
   partitioned into `AsuraRoot` (and `ASURA_BOOT` if encrypted).
7. **GRUB OS-Prober**: Enables `GRUB_DISABLE_OS_PROBER=false` in `/etc/default/grub` so `grub-mkconfig`
   automatically discovers Windows Boot Manager and other installed kernels.
8. **Preserve Fallback Bootloader**: In dual-boot mode, `/boot/efi/EFI/BOOT/BOOTX64.EFI` is
   never overwritten, preserving vendor and Windows fallback boot options.

---

## 4. Unattended Installation (`--config FILE`)

Deploy Toren Linux non-interactively using an answer file. Passwords can be supplied as pre-hashed
shadow strings (`$6$...`) or plain strings. Passwords and passphrases are **never logged in cleartext**.

### Sample Answer File (`asura-answer.yaml` / properties)
```yaml
disk: /dev/sda
mode: alongside          # "alongside" (dual-boot) or "wipe" (clean)
filesystem: btrfs        # "btrfs" or "ext4"
encryption: yes          # "yes" or "no"
luks_password: SuperSecureLuksPass123!
hostname: asura-station
user: asura
password_hash: $6$rounds=4096$saltsalt$fakehashvalueforunattended123
timezone: UTC
locale: en_US.UTF-8
keyboard: us
```

### Executing Unattended Install
```sh
# Dry run to preview the execution plan
asura install --dry-run --config /path/to/asura-answer.yaml

# Execute unattended installation
sudo asura install --config /path/to/asura-answer.yaml
```

---

## 5. Converting an Existing Void Linux Install (`toren adopt`)

`toren adopt` installs the Toren desktop shell, portal graph, and services in-place onto an
existing Void Linux glibc x86_64 system without reinstalling, repartitioning, or touching the kernel.

### Commands
```sh
# 1. Preflight check (inspects Void, glibc, x86_64, disk space >= 5GB)
asura adopt --check

# 2. Execute adoption (atomic backup to /etc/asura/adopt/backup/<timestamp>/)
sudo asura adopt

# 3. Verify deployed session registrations and integrity
asura adopt --verify

# 4. Check for obsolete or conflicting display managers
asura adopt --prune

# 5. Full rollback (byte-for-byte undo using self-contained restore.sh)
sudo asura adopt --undo
```

### Adoption Guardrails
- **Distribution Check**: Strictly requires Void Linux (`/etc/os-release` `ID=void`).
- **Architecture Check**: Strictly requires `x86_64`.
- **C Library Check**: Strictly requires `glibc`. Musl installations are refused.
- **Reversibility**: Generates `/etc/asura/adopt/backup/<timestamp>/restore.sh` and `/etc/asura/adopt/state.json`.
  Running `toren adopt --undo` restores all replaced files byte-for-byte and unregisters Toren session components.
