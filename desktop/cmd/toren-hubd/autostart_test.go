package main

import (
	"path/filepath"
	"testing"

	"asura/desktop/cmd/toren-hubd/internal/wm"
)

func TestAutostartAddAndRemove(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))

	// Initially empty
	o := wm.LoadStore(desktopStorePath())
	if len(o.Autostart) != 0 {
		t.Fatalf("expected 0 autostart entries, got %d", len(o.Autostart))
	}

	// Add valid command
	if err := addAutostart("nm-applet --indicator"); err != nil {
		t.Fatalf("addAutostart failed: %v", err)
	}

	o = wm.LoadStore(desktopStorePath())
	if len(o.Autostart) != 1 {
		t.Fatalf("expected 1 autostart entry, got %d", len(o.Autostart))
	}
	if o.Autostart[0].Command != "nm-applet --indicator" {
		t.Errorf("unexpected autostart command: %q", o.Autostart[0].Command)
	}

	// Reject empty command
	if err := addAutostart("   "); err == nil {
		t.Fatal("expected error for empty autostart, got nil")
	}

	// Add another command
	if err := addAutostart("blueman-applet"); err != nil {
		t.Fatalf("addAutostart failed: %v", err)
	}
	o = wm.LoadStore(desktopStorePath())
	if len(o.Autostart) != 2 {
		t.Fatalf("expected 2 autostart entries, got %d", len(o.Autostart))
	}

	// Remove index 0
	if err := removeAutostart(0); err != nil {
		t.Fatalf("removeAutostart failed: %v", err)
	}
	o = wm.LoadStore(desktopStorePath())
	if len(o.Autostart) != 1 {
		t.Fatalf("expected 1 autostart entry after remove, got %d", len(o.Autostart))
	}
	if o.Autostart[0].Command != "blueman-applet" {
		t.Errorf("expected blueman-applet remaining, got: %+v", o.Autostart[0])
	}
}
