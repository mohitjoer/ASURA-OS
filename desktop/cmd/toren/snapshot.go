package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// System snapshots: read-only Btrfs snapshots of / in /.snapshots, named by
// why they were taken and when. `toren update` takes one before a system
// update, toren-pkg-add before installing or removing packages, and a person
// can take one by hand. `toren rollback --system` lists and restores them.

const (
	snapshotPreUpdate  = "pre-update-"
	snapshotPreInstall = "pre-install-"
	snapshotManual     = "manual-"
	snapshotStamp      = "20060102-150405"
)

// snapshotReasons maps `toren snapshot create --reason` to a name prefix.
var snapshotReasons = map[string]string{
	"update":  snapshotPreUpdate,
	"install": snapshotPreInstall,
	"manual":  snapshotManual,
}

const snapshotUsage = `Usage: toren snapshot [list | create [--reason update|install|manual]]

System snapshots: read-only Btrfs snapshots of / in /.snapshots.

  list                  every snapshot, newest first, with its date, why it
                        was taken, and the ASURA commit it holds (the default)
  create                take one now (as root, through sudo); --reason names
                        it (default manual). The automatic kinds keep their
                        newest five; manual ones are never removed.

Restore one with: sudo toren rollback --system NAME
`

func init() {
	register("snapshot", "take or list system (Btrfs) snapshots", torenSnapshot)
}

func torenSnapshot(args []string) int {
	if len(args) == 0 || args[0] == "list" {
		return printSnapshots("/.snapshots")
	}
	switch args[0] {
	case "-h", "--help", "help":
		fmt.Print(snapshotUsage)
		return exitOK
	case "create":
		reason := "manual"
		rest := args[1:]
		for len(rest) > 0 {
			switch {
			case rest[0] == "--reason" && len(rest) > 1:
				reason, rest = rest[1], rest[2:]
			case strings.HasPrefix(rest[0], "--reason="):
				reason, rest = strings.TrimPrefix(rest[0], "--reason="), rest[1:]
			default:
				fmt.Fprint(os.Stderr, snapshotUsage)
				return exitUsage
			}
		}
		prefix, ok := snapshotReasons[reason]
		if !ok {
			fmt.Fprintf(os.Stderr, "toren snapshot: unknown reason %q (update, install or manual)\n", reason)
			return exitUsage
		}
		path, err := createBtrfsSnapshotAs(prefix)
		if err != nil {
			fmt.Fprintln(os.Stderr, "toren snapshot:", err)
			return exitFail
		}
		if path == "" {
			fmt.Println("no snapshot: / is not Btrfs, or btrfs-progs is not installed")
			return exitOK
		}
		fmt.Printf("system snapshot: %s\n", path)
		if prefix != snapshotManual {
			pruneBtrfsSnapshots(filepath.Dir(path), keepUpdateSnapshots)
		}
		return exitOK
	}
	fmt.Fprint(os.Stderr, snapshotUsage)
	return exitUsage
}

// printSnapshots lists dir's snapshots, newest first, one described per line.
func printSnapshots(dir string) int {
	snaps, err := listSnapshotsIn(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "toren snapshot:", err)
		return exitFail
	}
	if len(snaps) == 0 {
		fmt.Printf("no system snapshots in %s\n", dir)
		return exitOK
	}
	for _, name := range snaps {
		fmt.Println("  " + describeSnapshot(dir, name))
	}
	return exitOK
}

// describeSnapshot is one line for a snapshot: its name, when and why it was
// taken, and the ASURA commit it holds when that was recorded.
func describeSnapshot(dir, name string) string {
	when, why := "", "taken by hand or by another tool"
	for prefix, reason := range map[string]string{
		snapshotPreUpdate:  "before a system update",
		snapshotPreInstall: "before installing or removing packages",
		snapshotManual:     "taken by hand",
	} {
		if stamp, ok := strings.CutPrefix(name, prefix); ok {
			why = reason
			if t, err := time.ParseInLocation(snapshotStamp, stamp, time.Local); err == nil {
				when = t.Format("2006-01-02 15:04")
			}
		}
	}
	if when == "" {
		if info, err := os.Stat(filepath.Join(dir, name)); err == nil {
			when = info.ModTime().Format("2006-01-02 15:04")
		}
	}
	line := fmt.Sprintf("%-32s %-16s %s", name, when, why)
	if commit := appliedCommit(filepath.Join(dir, name)); commit != "" {
		if len(commit) > 8 {
			commit = commit[:8]
		}
		line += " (ASURA " + commit + ")"
	}
	return line
}
