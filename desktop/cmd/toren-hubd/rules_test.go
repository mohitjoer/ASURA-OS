package main

import (
	"path/filepath"
	"testing"

	"asura/desktop/cmd/toren-hubd/internal/wm"
)

func TestRulesAddAndRemove(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))

	// Initially empty
	o := wm.LoadStore(desktopStorePath())
	if len(o.WindowRules) != 0 {
		t.Fatalf("expected 0 rules, got %d", len(o.WindowRules))
	}

	// Add valid rule
	if err := addRule("Spotify", "float", ""); err != nil {
		t.Fatalf("addRule failed: %v", err)
	}

	o = wm.LoadStore(desktopStorePath())
	if len(o.WindowRules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(o.WindowRules))
	}
	if o.WindowRules[0].Class != "Spotify" || o.WindowRules[0].Action != "float" {
		t.Errorf("unexpected rule: %+v", o.WindowRules[0])
	}

	// Add another rule with value
	if err := addRule("Gimp", "opacity", "0.9"); err != nil {
		t.Fatalf("addRule failed: %v", err)
	}
	o = wm.LoadStore(desktopStorePath())
	if len(o.WindowRules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(o.WindowRules))
	}

	// Reject invalid action
	if err := addRule("BadApp", "invalidaction", ""); err == nil {
		t.Fatal("expected error for invalid action, got nil")
	}

	// Reject empty class
	if err := addRule("", "float", ""); err == nil {
		t.Fatal("expected error for empty class, got nil")
	}

	// Remove index 0
	if err := removeRule(0); err != nil {
		t.Fatalf("removeRule failed: %v", err)
	}
	o = wm.LoadStore(desktopStorePath())
	if len(o.WindowRules) != 1 {
		t.Fatalf("expected 1 rule after remove, got %d", len(o.WindowRules))
	}
	if o.WindowRules[0].Class != "Gimp" {
		t.Errorf("expected Gimp rule remaining, got: %+v", o.WindowRules[0])
	}

	// Remove out of bounds
	if err := removeRule(99); err == nil {
		t.Fatal("expected error for out of bounds remove, got nil")
	}
}
