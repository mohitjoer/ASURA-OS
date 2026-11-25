package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsTorenRepo(t *testing.T) {
	tmp := t.TempDir()
	if isTorenRepo(tmp) {
		t.Fatal("expected empty directory to not be Toren repo")
	}

	if err := os.Mkdir(filepath.Join(tmp, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if isTorenRepo(tmp) {
		t.Fatal("expected .git without shell/hub to not be Toren repo")
	}

	if err := os.MkdirAll(filepath.Join(tmp, "shell", "hub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if !isTorenRepo(tmp) {
		t.Fatal("expected directory with .git and shell/hub to be Toren repo")
	}
}

func TestFindOrCloneRepoFromEnv(t *testing.T) {
	tmp := t.TempDir()
	if err := os.Mkdir(filepath.Join(tmp, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(tmp, "shell", "hub"), 0o755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("ASURA_SRC_DIR", tmp)
	repo, err := findOrCloneRepo("dev-unstable")
	if err != nil {
		t.Fatalf("findOrCloneRepo failed: %v", err)
	}
	if repo != tmp {
		t.Fatalf("got %s, want %s", repo, tmp)
	}
}

func TestTorenUpdateDryRunUnprivileged(t *testing.T) {
	tmp := t.TempDir()
	// Set up fixture paths
	configDir := filepath.Join(tmp, "home", ".config", "toren")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	baseDir := filepath.Join(tmp, "usr", "share", "toren")
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		t.Fatal(err)
	}

	code := torenUpdate([]string{"--root", tmp, "--dry-run", "--no-snapshot"})
	if code != exitOK {
		t.Fatalf("toren update unprivileged dry-run returned %d, want %d", code, exitOK)
	}
}

func TestApplyRepoArgs(t *testing.T) {
	args := applyRepoArgs("/src/toren", "/")
	want := filepath.Join("/src/toren", "installation", "apply-repo.sh")
	if os.Geteuid() != 0 {
		if args[0] != "sudo" {
			t.Fatalf("not root: want sudo first, got %v", args)
		}
		args = args[1:]
	}
	if len(args) != 3 || args[0] != "sh" || args[1] != want || args[2] != "/src/toren" {
		t.Fatalf("root /: got %v", args)
	}
	args = applyRepoArgs("/src/toren", "/mnt/target")
	if args[len(args)-1] != "/mnt/target" {
		t.Fatalf("a --root tree must be passed on: got %v", args)
	}
}

func TestApplyRepoRefusesOldCheckout(t *testing.T) {
	if err := applyRepo(t.TempDir(), "/"); err == nil {
		t.Fatal("a checkout without installation/apply-repo.sh must be an error, not a silent no-op")
	}
}

func TestParseBrokenShlibs(t *testing.T) {
	out := "[*] Updating repository ...\n" +
		"aquamarine-0.8.0_2: broken, unresolvable shlib `libdisplay-info.so.2'\n" +
		"hyprgraphics-0.1.3_2: broken, unresolvable shlib `libjxl.so.0.11'\n" +
		"hyprgraphics-0.1.3_2: broken, unresolvable shlib `libjxl_threads.so.0.11'\n" +
		"ERROR: Transaction aborted due to unresolved shlibs.\n"
	pkgs, libs := parseBrokenShlibs(out)
	if strings.Join(pkgs, " ") != "aquamarine hyprgraphics" {
		t.Errorf("pkgs = %v", pkgs)
	}
	if strings.Join(libs, " ") != "libdisplay-info.so.2 libjxl.so.0.11 libjxl_threads.so.0.11" {
		t.Errorf("libs = %v", libs)
	}
}

func TestOldUpdateSnapshots(t *testing.T) {
	names := []string{"pre-update-20261101-100000", "manual-backup", "pre-update-20261103-100000",
		"pre-update-20261102-100000", "pre-update-20261105-100000", "pre-update-20261104-100000"}
	got := oldUpdateSnapshots(names, 3)
	if strings.Join(got, " ") != "pre-update-20261101-100000 pre-update-20261102-100000" {
		t.Errorf("old = %v", got)
	}
	if got := oldUpdateSnapshots(names, 10); len(got) != 0 {
		t.Errorf("nothing beyond the limit, got %v", got)
	}
}
