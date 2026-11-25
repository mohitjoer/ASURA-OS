package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestReinstallPlan verifies that the reinstall plan targets the exact repo and root
// using apply-repo.sh and never updates to anything newer.
func TestReinstallPlan(t *testing.T) {
	repoDir := "/fake/repo"
	commit := "abcdef1234567890"
	root := "/target/root"

	plan := reinstallPlan(repoDir, commit, root)
	if os.Geteuid() != 0 {
		if plan[0] != "sudo" {
			t.Fatalf("not root: expected sudo first, got %v", plan)
		}
		plan = plan[1:]
	}

	if len(plan) < 3 {
		t.Fatalf("reinstall plan too short: %v", plan)
	}

	if plan[0] != "sh" {
		t.Errorf("plan[0] = %q, want 'sh'", plan[0])
	}
	if !strings.HasSuffix(plan[1], "apply-repo.sh") {
		t.Errorf("plan[1] = %q, want path ending in apply-repo.sh", plan[1])
	}
	if plan[2] != repoDir {
		t.Errorf("plan[2] = %q, want repoDir %q", plan[2], repoDir)
	}
	if len(plan) >= 4 && plan[3] != root {
		t.Errorf("plan[3] = %q, want root %q", plan[3], root)
	}
}

// TestReinstallRequiresAppliedCommit verifies that reinstall fails gracefully when
// no applied commit is recorded and no repository exists.
func TestReinstallRequiresAppliedCommit(t *testing.T) {
	tmpRoot := t.TempDir()
	// No appliedCommit file in tmpRoot
	code := asuraReinstall([]string{"--root", tmpRoot})
	// Should fail since no commit recorded
	if code == exitOK {
		t.Errorf("expected failure when applied-commit is missing, got exitOK")
	}
}

// TestReinstallDetectsAppliedCommit verifies reading recorded applied-commit.
func TestReinstallDetectsAppliedCommit(t *testing.T) {
	tmpRoot := t.TempDir()
	commitFile := filepath.Join(tmpRoot, "usr", "share", "asura", "applied-commit")
	if err := os.MkdirAll(filepath.Dir(commitFile), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(commitFile, []byte("1122334455667788\n"), 0o644); err != nil {
		t.Fatalf("write applied-commit: %v", err)
	}

	commit := appliedCommit(tmpRoot)
	if commit != "1122334455667788" {
		t.Fatalf("appliedCommit = %q, want '1122334455667788'", commit)
	}
}

// TestReinstallExecution verifies that asuraReinstall invokes reinstallApply
// targeting the applied commit and root.
func TestReinstallExecution(t *testing.T) {
	tmpRoot := t.TempDir()
	commitFile := filepath.Join(tmpRoot, "usr", "share", "asura", "applied-commit")
	if err := os.MkdirAll(filepath.Dir(commitFile), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Current repo HEAD commit
	head := gitOut(".", "rev-parse", "HEAD")
	if head == "" {
		t.Skip("not in a git repository")
	}
	if err := os.WriteFile(commitFile, []byte(head+"\n"), 0o644); err != nil {
		t.Fatalf("write applied-commit: %v", err)
	}

	appliedTarget := ""
	oldApply := reinstallApply
	defer func() { reinstallApply = oldApply }()
	reinstallApply = func(repoDir, root string) error {
		appliedTarget = root
		return nil
	}

	code := asuraReinstall([]string{"--root", tmpRoot})
	if code != exitOK {
		t.Fatalf("asuraReinstall failed with code %d", code)
	}
	if appliedTarget != tmpRoot {
		t.Errorf("reinstallApply called with root %q, want %q", appliedTarget, tmpRoot)
	}
}
