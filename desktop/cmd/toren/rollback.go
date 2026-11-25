package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"toren/desktop/internal/overlay"
)

// torenRollback reverts configuration overlay (--config) or restores Btrfs
// root snapshots (--system).
func torenRollback(args []string) int {
	set := newFlags("rollback", rollbackUsage)
	var (
		root       = set.String("root", "", "operate on a tree under this directory")
		configMode = set.Bool("config", false, "revert config overlay to previous revision")
		systemMode = set.Bool("system", false, "list or restore Btrfs root snapshots")
		listMode   = set.Bool("list", false, "list available revisions and snapshots")
		force      = set.Bool("force", false, "restore even when the snapshot has no modules for the installed kernel")
	)
	positional, err := parseFlags(set, args)
	if err != nil {
		return exitUsage
	}

	paths, err := overlayPaths(*root)
	if err != nil {
		fmt.Fprintln(os.Stderr, "toren rollback:", err)
		return exitFail
	}

	rollbackForce = *force
	if *listMode {
		return listRollbackTargets(paths)
	}

	if *systemMode {
		target := ""
		if len(positional) > 0 {
			target = positional[0]
		}
		return rollbackSystem(target)
	}

	// Default or --config mode
	if *configMode || len(positional) == 0 {
		return rollbackConfig(paths)
	}

	// Positional argument without explicit flag: check if it matches a system snapshot
	if isBtrfsSnapshot(positional[0]) {
		return rollbackSystem(positional[0])
	}
	return rollbackConfig(paths)
}

// rollbackConfig rolls back the ~/.config overlay to the previous revision.
func rollbackConfig(paths overlay.Paths) int {
	restored, err := overlay.Rollback(paths)
	if err != nil {
		fmt.Fprintln(os.Stderr, "toren rollback:", err)
		return exitFail
	}
	fmt.Printf("rolled back to revision %s (recorded %s)\n",
		restored.ID, restored.RecordedAt.Format("2006-01-02 15:04:05Z"))
	return exitOK
}

// rollbackSystem lists or restores a Btrfs root snapshot.
func rollbackSystem(target string) int {
	if target == "" {
		fmt.Println("Btrfs System Snapshots:")
		snaps, err := listBtrfsSnapshots()
		if err != nil {
			fmt.Fprintln(os.Stderr, "toren rollback:", err)
			return exitFail
		}
		if len(snaps) == 0 {
			fmt.Println("  no system snapshots found in /.snapshots")
			return exitOK
		}
		for _, s := range snaps {
			fmt.Println("  " + describeSnapshot("/.snapshots", s))
		}
		fmt.Println("\nTo restore a system snapshot, run:")
		fmt.Println("  sudo toren rollback --system <snapshot-name>")
		return exitOK
	}

	snapPath := target
	if !filepath.IsAbs(snapPath) {
		snapPath = filepath.Join("/.snapshots", target)
	}

	info, err := os.Stat(snapPath)
	if err != nil || !info.IsDir() {
		fmt.Fprintf(os.Stderr, "toren rollback: snapshot not found: %s\n", snapPath)
		return exitFail
	}

	name := filepath.Base(snapPath)
	fmt.Printf("==> Restoring system snapshot: %s\n", name)

	// The snapshot must carry modules for the kernel GRUB boots. /boot is its
	// own partition and is not rolled back, so a root from before a kernel
	// update would boot that kernel with no modules for it.
	if kernel := runningKernel(); kernel != "" && !rollbackForce {
		if _, err := os.Stat(filepath.Join(snapPath, "usr", "lib", "modules", kernel)); err != nil {
			fmt.Fprintf(os.Stderr, "toren rollback: %s has no modules for the installed kernel %s (/boot is not rolled back),\n"+
				"so the restored system would boot without them. Install a kernel the snapshot has, or pass --force.\n", name, kernel)
			return exitFail
		}
	}

	device, err := rootDevice()
	if err != nil {
		fmt.Fprintln(os.Stderr, "toren rollback:", err)
		return exitFail
	}
	top, err := os.MkdirTemp("", "toren-rollback-")
	if err != nil {
		fmt.Fprintln(os.Stderr, "toren rollback:", err)
		return exitFail
	}
	defer os.Remove(top)

	stamp := time.Now().UTC().Format("20060102T150405Z")
	steps, aside := systemRollbackPlan(device, top, name, stamp)
	for i, step := range steps {
		if err := runPrivileged(step...); err != nil {
			fmt.Fprintf(os.Stderr, "toren rollback: %s: %v\n", strings.Join(step, " "), err)
			if i == 4 {
				// @ was moved aside and no new @ was made: put it back, or
				// the next boot finds no root.
				fmt.Fprintln(os.Stderr, "toren rollback: putting the current root back as @")
				_ = runPrivileged("mv", filepath.Join(top, aside), filepath.Join(top, "@"))
			}
			if i >= 2 && i < 5 { // the top-level volume is mounted
				_ = runPrivileged("umount", top)
			}
			fmt.Fprintln(os.Stderr, "toren rollback: nothing was changed")
			return exitFail
		}
	}

	fmt.Println("==> Updating GRUB bootloader menu...")
	if err := runPrivileged("grub-mkconfig", "-o", "/boot/grub/grub.cfg"); err != nil {
		fmt.Fprintf(os.Stderr, "warning: grub-mkconfig failed: %v\n", err)
	}

	fmt.Printf("\nRollback prepared: @ is now a writable copy of %s.\n", name)
	fmt.Printf("The current root is kept as %s; it stays in use until you reboot.\n", aside)
	fmt.Println("Reboot to start the restored system. Once it works, the old root can be deleted:")
	fmt.Printf("  sudo mount -o subvolid=5 %s /mnt && sudo btrfs subvolume delete /mnt/%s && sudo umount /mnt\n", device, aside)
	return exitOK
}

// rollbackForce skips the kernel-modules check.
var rollbackForce bool

// systemRollbackPlan is the restore, as commands, for the layout the installer
// makes: top-level subvolumes @ (/, subvol=@ in fstab and GRUB), @home and
// @snapshots (/.snapshots). The current @ is renamed aside and a writable
// snapshot of @snapshots/NAME becomes @, so the next boot, which mounts @,
// starts the restored system; nothing in fstab or GRUB changes. It returns the
// steps and the name the current root is kept under.
func systemRollbackPlan(device, top, name, stamp string) ([][]string, string) {
	aside := "@pre-rollback-" + stamp
	return [][]string{
		{"mkdir", "-p", top},
		{"mount", "-o", "subvolid=5", device, top},
		{"test", "-d", filepath.Join(top, "@snapshots", name)},
		{"mv", filepath.Join(top, "@"), filepath.Join(top, aside)},
		{"btrfs", "subvolume", "snapshot", filepath.Join(top, "@snapshots", name), filepath.Join(top, "@")},
		{"umount", top},
	}, aside
}

// runPrivileged runs a command as root, through sudo when not root already.
func runPrivileged(args ...string) error {
	if os.Geteuid() != 0 {
		args = append([]string{"sudo"}, args...)
	}
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout, cmd.Stderr, cmd.Stdin = os.Stdout, os.Stderr, os.Stdin
	return cmd.Run()
}

// rootDevice is the block device / is mounted from ("/dev/sda2[/@]" -> "/dev/sda2").
func rootDevice() (string, error) {
	out, err := exec.Command("findmnt", "-no", "SOURCE", "/").Output()
	if err != nil {
		return "", fmt.Errorf("findmnt /: %w", err)
	}
	dev := strings.TrimSpace(string(out))
	if i := strings.Index(dev, "["); i > 0 {
		dev = dev[:i]
	}
	if !strings.HasPrefix(dev, "/dev/") {
		return "", fmt.Errorf("/ is mounted from %q, not a block device", dev)
	}
	return dev, nil
}

// runningKernel is the installed kernel's release, the one GRUB boots.
func runningKernel() string {
	out, err := exec.Command("uname", "-r").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// listRollbackTargets prints both config revisions and system snapshots.
func listRollbackTargets(paths overlay.Paths) int {
	fmt.Println("Configuration Overlay Revisions:")
	history, err := overlay.Revisions(paths)
	if err != nil || len(history) == 0 {
		fmt.Println("  no configuration revisions recorded")
	} else {
		for _, rev := range history {
			fmt.Printf("  • %s  (%s)\n", rev.ID, rev.RecordedAt.Format("2006-01-02 15:04:05Z"))
		}
	}

	fmt.Println("\nBtrfs System Snapshots:")
	snaps, err := listBtrfsSnapshots()
	if err != nil || len(snaps) == 0 {
		fmt.Println("  no system snapshots found")
	} else {
		for _, s := range snaps {
			fmt.Println("  " + describeSnapshot("/.snapshots", s))
		}
	}
	return exitOK
}

// isBtrfsSnapshot checks if the name matches an existing snapshot in /.snapshots.
func isBtrfsSnapshot(name string) bool {
	snapPath := name
	if !filepath.IsAbs(snapPath) {
		snapPath = filepath.Join("/.snapshots", name)
	}
	info, err := os.Stat(snapPath)
	return err == nil && info.IsDir()
}

// listBtrfsSnapshots returns available snapshot names in /.snapshots.
func listBtrfsSnapshots() ([]string, error) {
	return listSnapshotsIn("/.snapshots")
}

// snapshotSortKey orders snapshots by the timestamp in their names, so a
// pre-install one sorts among pre-update ones by time, not by prefix. Names
// with no timestamp sort after them (newest first means they come last).
func snapshotSortKey(name string) string {
	for _, prefix := range []string{snapshotPreUpdate, snapshotPreInstall, snapshotManual} {
		if stamp, ok := strings.CutPrefix(name, prefix); ok {
			return "1" + stamp
		}
	}
	return "0" + name
}

// listSnapshotsIn returns the snapshot names in dir, newest first: the
// automatic kinds by the time in their names, the rest by name.
func listSnapshotsIn(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var snaps []string
	for _, entry := range entries {
		if entry.IsDir() {
			snaps = append(snaps, entry.Name())
		}
	}
	sort.SliceStable(snaps, func(i, j int) bool { return snapshotSortKey(snaps[i]) > snapshotSortKey(snaps[j]) })
	return snaps, nil
}
