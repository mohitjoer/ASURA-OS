package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunUserHooks(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	t.Setenv("XDG_STATE_HOME", filepath.Join(home, ".local", "state"))

	hookDir := filepath.Join(home, ".config", "toren", "hooks", "post-theme")
	_ = os.MkdirAll(hookDir, 0o755)

	markerFile := filepath.Join(home, "marker.txt")
	script := "#!/bin/sh\necho \"EVENT=$ASURA_EVENT THEME=$ASURA_THEME_NAME USER=$ASURA_USER\" >> \"" + markerFile + "\"\n"
	scriptPath := filepath.Join(hookDir, "01-test-hook.sh")
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	// Add failing script to test resilience
	failingScript := "#!/bin/sh\nexit 42\n"
	if err := os.WriteFile(filepath.Join(hookDir, "02-failing.sh"), []byte(failingScript), 0o755); err != nil {
		t.Fatal(err)
	}

	runUserHooks("post-theme", map[string]string{
		"ASURA_THEME_NAME": "cyberpunk",
	})

	b, err := os.ReadFile(markerFile)
	if err != nil {
		t.Fatalf("hook did not write marker file: %v", err)
	}
	out := string(b)
	if !strings.Contains(out, "EVENT=post-theme") || !strings.Contains(out, "THEME=cyberpunk") {
		t.Errorf("unexpected marker content: %s", out)
	}

	// Verify log file was written
	logB, err := os.ReadFile(hooksLogPath())
	if err != nil {
		t.Fatalf("hooks.log not created: %v", err)
	}
	logContent := string(logB)
	if !strings.Contains(logContent, "01-test-hook.sh -> SUCCESS") {
		t.Errorf("hooks.log missing success entry: %s", logContent)
	}
	if !strings.Contains(logContent, "02-failing.sh -> FAILED") {
		t.Errorf("hooks.log missing failed entry: %s", logContent)
	}
}

func TestHookCLI(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))

	// List with no hooks
	if code := torenHook([]string{"list"}); code != exitOK {
		t.Errorf("toren hook list exited %d", code)
	}

	// Run with missing event
	if code := torenHook([]string{"run"}); code != exitUsage {
		t.Errorf("toren hook run with no event exited %d", code)
	}

	// Run with valid event
	if code := torenHook([]string{"run", "post-login"}); code != exitOK {
		t.Errorf("toren hook run post-login exited %d", code)
	}
}
