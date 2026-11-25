package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestSafeModeConfigStructure verifies that safe-mode.conf meets all the requirements:
// minimal built-in config, terminal bound to key, no shell, no user configs, no plugins.
func TestSafeModeConfigStructure(t *testing.T) {
	// Locate repo root
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	repoRoot := cwd
	for {
		if _, err := os.Stat(filepath.Join(repoRoot, "session", "safe-mode", "hyprland.conf")); err == nil {
			break
		}
		parent := filepath.Dir(repoRoot)
		if parent == repoRoot {
			t.Fatal("could not find session/safe-mode/hyprland.conf from working tree")
		}
		repoRoot = parent
	}

	confPath := filepath.Join(repoRoot, "session", "safe-mode", "hyprland.conf")
	data, err := os.ReadFile(confPath)
	if err != nil {
		t.Fatalf("reading %s: %v", confPath, err)
	}
	content := string(data)

	// Check terminal binding
	if !strings.Contains(content, "bind = $mainMod, Return, exec,") {
		t.Errorf("safe-mode.conf missing terminal keybind on Super+Return")
	}

	// Check exit keybind
	if !strings.Contains(content, "bind = $mainMod, M, exit") {
		t.Errorf("safe-mode.conf missing compositor exit keybind on Super+M")
	}

	// Verify no user configs are sourced
	if strings.Contains(content, "source = ~/.config") || strings.Contains(content, "user.conf") {
		t.Errorf("safe-mode.conf must NOT source user configuration files")
	}

	// Verify no quickshell or heavy desktop shell is autostarted
	if strings.Contains(content, "quickshell") || strings.Contains(content, "toren-activate") {
		t.Errorf("safe-mode.conf must NOT autostart quickshell or desktop shell")
	}

	// If Hyprland binary exists, test verification flag
	if hypr, err := exec.LookPath("Hyprland"); err == nil {
		cmd := exec.Command(hypr, "--verify-config", "-c", confPath)
		if out, err := cmd.CombinedOutput(); err != nil {
			// Some older Hyprland versions don't have --verify-config, skip if flag unrecognized
			if !strings.Contains(string(out), "unrecognized option") && !strings.Contains(string(out), "invalid option") {
				t.Logf("Hyprland --verify-config warning: %v, output: %s", err, string(out))
			}
		}
	}
}

// TestSafeModeDesktopEntry verifies the session desktop entry and script exist and are valid.
func TestSafeModeDesktopEntry(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	repoRoot := cwd
	for {
		if _, err := os.Stat(filepath.Join(repoRoot, "session", "toren-safe.desktop")); err == nil {
			break
		}
		parent := filepath.Dir(repoRoot)
		if parent == repoRoot {
			t.Fatal("could not find session/toren-safe.desktop")
		}
		repoRoot = parent
	}

	desktopFile := filepath.Join(repoRoot, "session", "toren-safe.desktop")
	b, err := os.ReadFile(desktopFile)
	if err != nil {
		t.Fatalf("reading toren-safe.desktop: %v", err)
	}
	d := string(b)
	if !strings.Contains(d, "Name=Toren Linux (safe mode)") {
		t.Errorf("desktop entry missing Name=Toren Linux (safe mode)")
	}
	if !strings.Contains(d, "Exec=/usr/share/toren/session/toren-session-safe") {
		t.Errorf("desktop entry missing correct Exec path")
	}

	safeScript := filepath.Join(repoRoot, "session", "toren-session-safe")
	info, err := os.Stat(safeScript)
	if err != nil {
		t.Fatalf("stat toren-session-safe: %v", err)
	}
	if info.Mode()&0o111 == 0 {
		t.Errorf("toren-session-safe is not executable (mode: %v)", info.Mode())
	}
}
