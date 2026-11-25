package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// TestResetDesktopBackupAndRestore verifies:
// 1. asura reset --desktop backs up user desktop configs to a dated backup folder
// 2. Original user configurations are cleared/reset
// 3. asura reset --restore BACKUP restores original user configs byte-for-byte
func TestResetDesktopBackupAndRestore(t *testing.T) {
	tmpHome := t.TempDir()

	// Create fake user desktop configurations
	fakeIIConfig := []byte(`{"theme": "custom-neon", "dark_mode": true, "volume": 85}`)
	fakeUserConf := []byte("# Custom user keybindings\nbind = SUPER, X, exec, custom-terminal\n")
	fakeUserEdits := []byte("custom user overlay edit content\n")
	fakeWallsState := []byte(`{"wallpaper": "neon-city.png", "blur": 15}`)

	iiPath := filepath.Join(tmpHome, ".config", "illogical-impulse", "config.json")
	userConfPath := filepath.Join(tmpHome, ".config", "asura", "hyprland", "user.conf")
	userEditsPath := filepath.Join(tmpHome, ".config", "asura", "user_edits", "shell.json")
	wallsStatePath := filepath.Join(tmpHome, ".local", "state", "asura", "asura-walls.json")

	mustWriteFile(t, iiPath, fakeIIConfig, 0o644)
	mustWriteFile(t, userConfPath, fakeUserConf, 0o644)
	mustWriteFile(t, userEditsPath, fakeUserEdits, 0o644)
	mustWriteFile(t, wallsStatePath, fakeWallsState, 0o644)

	// Step 1: Run `asura reset --desktop --root <tmpHome>`
	code := asuraReset([]string{"--desktop", "--root", tmpHome})
	if code != exitOK {
		t.Fatalf("asuraReset --desktop returned exit code %d, want %d", code, exitOK)
	}

	// Verify original files are gone/reset
	if _, err := os.Stat(iiPath); !os.IsNotExist(err) {
		t.Errorf("expected %s to be removed, but it exists", iiPath)
	}
	if _, err := os.Stat(userConfPath); !os.IsNotExist(err) {
		t.Errorf("expected %s to be removed, but it exists", userConfPath)
	}
	if _, err := os.Stat(userEditsPath); !os.IsNotExist(err) {
		t.Errorf("expected %s to be removed, but it exists", userEditsPath)
	}
	if _, err := os.Stat(wallsStatePath); !os.IsNotExist(err) {
		t.Errorf("expected %s to be removed, but it exists", wallsStatePath)
	}

	// Verify backup was created
	backupsBase := filepath.Join(tmpHome, ".local", "state", "asura", "backups")
	entries, err := os.ReadDir(backupsBase)
	if err != nil || len(entries) == 0 {
		t.Fatalf("expected backup directory in %s, got err: %v, count: %d", backupsBase, err, len(entries))
	}

	backupDir := filepath.Join(backupsBase, entries[0].Name())

	// Step 2: Run `asura reset --restore <backupDir> --root <tmpHome>`
	codeRestore := asuraReset([]string{"--restore", backupDir, "--root", tmpHome})
	if codeRestore != exitOK {
		t.Fatalf("asuraReset --restore returned exit code %d, want %d", codeRestore, exitOK)
	}

	// Step 3: Verify byte-for-byte restoration of all backed up files
	assertFileBytes(t, iiPath, fakeIIConfig)
	assertFileBytes(t, userConfPath, fakeUserConf)
	assertFileBytes(t, userEditsPath, fakeUserEdits)
	assertFileBytes(t, wallsStatePath, fakeWallsState)
}

func mustWriteFile(t *testing.T, path string, content []byte, mode os.FileMode) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, content, mode); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func assertFileBytes(t *testing.T, path string, want []byte) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading restored file %s: %v", path, err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("restored file %s content mismatch:\n got:  %q\n want: %q", path, string(got), string(want))
	}
}
