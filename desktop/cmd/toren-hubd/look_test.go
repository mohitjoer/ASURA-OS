package main

import (
	"testing"
)

func TestLookMode(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	// Default mode
	if m := currentMode(); m != "dark" {
		t.Errorf("expected default mode 'dark', got %q", m)
	}

	// Set mode to light
	if err := setMode("light"); err != nil {
		t.Fatalf("setMode failed: %v", err)
	}
	if m := currentMode(); m != "light" {
		t.Errorf("expected mode 'light', got %q", m)
	}

	// Reject invalid mode
	if err := setMode("neon"); err == nil {
		t.Fatal("expected error for invalid mode 'neon', got nil")
	}

	// List looks should not crash even if themes file not installed
	if err := listLooks(); err != nil {
		t.Fatalf("listLooks failed: %v", err)
	}
}
