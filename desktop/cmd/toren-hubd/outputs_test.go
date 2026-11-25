package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"toren/desktop/cmd/toren-hubd/internal/wm"
)

// TestPublishGreeterPrimary: the Displays apply records the layout's main output
// where the login greeter can read it. The page's "Set as main" re-bases the
// layout onto the chosen output, so the origin entry wins; a layout that
// predates that convention falls back to the top-left-most enabled output, and
// a disabled output at the origin is not the main.
func TestPublishGreeterPrimary(t *testing.T) {
	file := filepath.Join(t.TempDir(), "greeter-primary")
	t.Setenv("ASURA_GREETER_PRIMARY_FILE", file)

	read := func() string {
		b, err := os.ReadFile(file)
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(b))
	}

	cases := []struct {
		name   string
		layout []wm.OutputLayout
		want   string
	}{
		{
			name: "origin is main",
			layout: []wm.OutputLayout{
				{Name: "HDMI-A-1", Enabled: true, X: 0, Y: 0},
				{Name: "eDP-1", Enabled: true, X: 1920, Y: 0},
			},
			want: "HDMI-A-1",
		},
		{
			name: "top-left without an origin",
			layout: []wm.OutputLayout{
				{Name: "DP-2", Enabled: true, X: 2560, Y: 0},
				{Name: "DP-1", Enabled: true, X: 1280, Y: 0},
			},
			want: "DP-1",
		},
		{
			name: "a disabled output at the origin is skipped",
			layout: []wm.OutputLayout{
				{Name: "DP-9", Enabled: false, X: 0, Y: 0},
				{Name: "eDP-1", Enabled: true, X: 0, Y: 0},
			},
			want: "eDP-1",
		},
		{
			name:   "nothing enabled publishes nothing",
			layout: []wm.OutputLayout{{Name: "DP-1", Enabled: false}},
			want:   "",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_ = os.Remove(file)
			publishGreeterPrimary(tc.layout)
			if got := read(); got != tc.want {
				t.Fatalf("published %q, want %q", got, tc.want)
			}
		})
	}
}

// TestOutputsUserConfAndRevert verifies:
// 1. apply-test writes ~/.config/toren/hyprland/monitors-user.conf
// 2. revert restores the previous file content byte-for-byte
func TestOutputsUserConfAndRevert(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	confPath := wm.MonitorsUserConfPath()
	initialContent := []byte("# Pre-existing user monitor config\nmonitor = eDP-1, 1920x1080@60, 0x0, 1.0\n")
	if err := os.MkdirAll(filepath.Dir(confPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(confPath, initialContent, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	newLayout := `[{"name":"eDP-1","enabled":true,"mode":"2560x1440@165","scale":1.25,"x":0,"y":0}]`

	// Test apply-test
	if err := applyOutputsTest(newLayout); err != nil {
		t.Fatalf("applyOutputsTest failed: %v", err)
	}

	// Verify new configuration was written
	written, err := os.ReadFile(confPath)
	if err != nil {
		t.Fatalf("read %s: %v", confPath, err)
	}
	if !strings.Contains(string(written), "2560x1440@165") {
		t.Errorf("monitors-user.conf missing new mode, got:\n%s", string(written))
	}

	// Test revert
	if err := revertOutputs(); err != nil {
		t.Fatalf("revertOutputs failed: %v", err)
	}

	// Verify byte-for-byte restore of original content
	restored, err := os.ReadFile(confPath)
	if err != nil {
		t.Fatalf("read restored %s: %v", confPath, err)
	}
	if string(restored) != string(initialContent) {
		t.Fatalf("reverted content mismatch:\n got:  %q\n want: %q", string(restored), string(initialContent))
	}
}
