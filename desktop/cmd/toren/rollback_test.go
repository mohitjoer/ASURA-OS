package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRollbackUsage(t *testing.T) {
	// Running with invalid flag returns exitUsage
	if code := torenRollback([]string{"--invalid-flag"}); code != exitUsage {
		t.Errorf("torenRollback(--invalid-flag) = %d, want %d", code, exitUsage)
	}
}

func TestRollbackListEmpty(t *testing.T) {
	temp := t.TempDir()
	code := torenRollback([]string{"--root", temp, "--list"})
	if code != exitOK {
		t.Errorf("torenRollback(--list) = %d, want %d", code, exitOK)
	}
}

func TestRollbackSystemList(t *testing.T) {
	code := torenRollback([]string{"--system"})
	if code != exitOK {
		t.Errorf("torenRollback(--system) = %d, want %d", code, exitOK)
	}
}

func TestRollbackNonExistentSystemSnapshot(t *testing.T) {
	code := torenRollback([]string{"--system", "non-existent-snapshot-12345"})
	if code != exitFail {
		t.Errorf("torenRollback(--system non-existent) = %d, want %d", code, exitFail)
	}
}

func TestIsBtrfsSnapshotHelper(t *testing.T) {
	temp := t.TempDir()
	snap := filepath.Join(temp, "snap1")
	if isBtrfsSnapshot(snap) {
		t.Errorf("isBtrfsSnapshot before mkdir reported true")
	}
	if err := os.Mkdir(snap, 0755); err != nil {
		t.Fatal(err)
	}
	if !isBtrfsSnapshot(snap) {
		t.Errorf("isBtrfsSnapshot after mkdir reported false")
	}
}

// The restore for the installer's layout: @ aside, a writable snapshot of
// @snapshots/NAME as @, on the top-level volume mounted by subvolid=5. Nothing
// is done inside the running root (the old code snapshotted into /@_restored_*,
// nested in @, and never used it).
func TestSystemRollbackPlan(t *testing.T) {
	steps, aside := systemRollbackPlan("/dev/sda2", "/tmp/top", "pre-update-1", "20261118T000000Z")
	if aside != "@pre-rollback-20261118T000000Z" {
		t.Fatalf("aside = %q", aside)
	}
	want := []string{
		"mkdir -p /tmp/top",
		"mount -o subvolid=5 /dev/sda2 /tmp/top",
		"test -d /tmp/top/@snapshots/pre-update-1",
		"mv /tmp/top/@ /tmp/top/@pre-rollback-20261118T000000Z",
		"btrfs subvolume snapshot /tmp/top/@snapshots/pre-update-1 /tmp/top/@",
		"umount /tmp/top",
	}
	if len(steps) != len(want) {
		t.Fatalf("%d steps, want %d: %v", len(steps), len(want), steps)
	}
	for i, step := range steps {
		if got := strings.Join(step, " "); got != want[i] {
			t.Errorf("step %d = %q, want %q", i, got, want[i])
		}
	}
	for _, step := range steps {
		if strings.Contains(strings.Join(step, " "), "-r ") {
			t.Error("the new @ must be writable: no read-only snapshot")
		}
	}
}
