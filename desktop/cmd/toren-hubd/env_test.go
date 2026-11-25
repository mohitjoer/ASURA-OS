package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"toren/desktop/cmd/toren-hubd/internal/wm"
)

func TestEnvSetRemoveAndRisky(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))

	// Check risky detection
	risky, warn := checkRiskyEnv("LD_PRELOAD")
	if !risky || warn == "" {
		t.Errorf("expected LD_PRELOAD to be risky with warning, got risky=%v, warn=%q", risky, warn)
	}
	risky, _ = checkRiskyEnv("PATH")
	if !risky {
		t.Error("expected PATH to be risky")
	}
	risky, _ = checkRiskyEnv("MOZ_ENABLE_WAYLAND")
	if risky {
		t.Error("expected MOZ_ENABLE_WAYLAND not to be risky")
	}

	// Set valid env var
	if err := setEnvVar("MOZ_ENABLE_WAYLAND", "1"); err != nil {
		t.Fatalf("setEnvVar failed: %v", err)
	}

	o := wm.LoadStore(desktopStorePath())
	if len(o.Env) != 1 {
		t.Fatalf("expected 1 env var in store, got %d", len(o.Env))
	}
	if o.Env[0].Key != "MOZ_ENABLE_WAYLAND" || o.Env[0].Value != "1" {
		t.Errorf("unexpected env var: %+v", o.Env[0])
	}

	// Verify ~/.config/toren/session/toren-session.env is written
	envFile := wm.SessionEnvPath()
	b, err := os.ReadFile(envFile)
	if err != nil {
		t.Fatalf("failed to read session env file: %v", err)
	}
	if !strings.Contains(string(b), `MOZ_ENABLE_WAYLAND="1"`) {
		t.Errorf("session env missing MOZ_ENABLE_WAYLAND: %s", string(b))
	}

	// Set risky env var
	if err := setEnvVar("LD_PRELOAD", "/opt/libtest.so"); err != nil {
		t.Fatalf("setEnvVar failed on risky var: %v", err)
	}

	// Reject invalid env key name
	if err := setEnvVar("BAD-KEY", "val"); err == nil {
		t.Fatal("expected error for BAD-KEY, got nil")
	}

	// Remove env var
	if err := removeEnvVar("MOZ_ENABLE_WAYLAND"); err != nil {
		t.Fatalf("removeEnvVar failed: %v", err)
	}
	o = wm.LoadStore(desktopStorePath())
	if len(o.Env) != 1 {
		t.Fatalf("expected 1 env var after remove, got %d", len(o.Env))
	}
	if o.Env[0].Key != "LD_PRELOAD" {
		t.Errorf("expected LD_PRELOAD remaining, got: %+v", o.Env[0])
	}
}
