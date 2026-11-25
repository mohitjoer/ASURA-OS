package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"asura/desktop/internal/runner/runnertest"
)

// The update's last step: services re-synced without touching runit's state or
// the user's `down`, the look written for the chosen shell, Hyprland reloaded,
// the shell restarted (by pkill when sv cannot), running daemons restarted.
func TestRefresh(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	rt := t.TempDir()
	t.Setenv("XDG_RUNTIME_DIR", rt)
	t.Setenv("HYPRLAND_INSTANCE_SIGNATURE", "sig1")
	t.Setenv("ASURA_DESKTOP_SHELL", "desktop")

	src := t.TempDir()
	write := func(p, s string, mode os.FileMode) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(s), mode); err != nil {
			t.Fatal(err)
		}
	}
	write(filepath.Join(src, "quickshell", "run"), "new run\n", 0o755)
	write(filepath.Join(src, "quickshell", "log", "run"), "new log\n", 0o755)
	write(filepath.Join(src, "asurad", "run"), "daemon\n", 0o755)
	write(filepath.Join(src, "asurad", "down"), "", 0o644)

	svc := filepath.Join(home, ".config", "service")
	write(filepath.Join(svc, "quickshell", "run"), "old run\n", 0o755)
	write(filepath.Join(svc, "quickshell", "supervise", "stat"), "run\n", 0o644)

	look := filepath.Join(t.TempDir(), "hyprland.conf")
	write(look, "# look\n", 0o644)
	oldSrc, oldLook := refreshServicesSrc, refreshShellLook
	refreshServicesSrc, refreshShellLook = src, look
	t.Cleanup(func() { refreshServicesSrc, refreshShellLook = oldSrc, oldLook })

	shellRestart := "sv restart " + filepath.Join(svc, "quickshell")
	fake := runnertest.New().
		Out("env HYPRLAND_INSTANCE_SIGNATURE=sig1 hyprctl reload", "ok").
		Fail(shellRestart, runnertest.Exit(1, "timeout: run: quickshell\n")).
		Out("pkill -x quickshell", "").
		Out("sv status "+filepath.Join(svc, "asurad"), "run: asurad\n").
		Out("sv restart "+filepath.Join(svc, "asurad"), "")

	var out, errOut bytes.Buffer
	if code := refreshMain(&out, &errOut, fake.Run); code != exitOK {
		t.Fatalf("exit %d; stderr %q; stdout %q", code, errOut.String(), out.String())
	}
	for _, w := range []string{"services:   2 synced", "compositor: reloaded", "shell:      restarted", "daemons:    restarted asurad"} {
		if !strings.Contains(out.String(), w) {
			t.Errorf("stdout lacks %q:\n%s", w, out.String())
		}
	}
	if b, _ := os.ReadFile(filepath.Join(svc, "quickshell", "run")); string(b) != "new run\n" {
		t.Errorf("run script not synced: %q", b)
	}
	if b, _ := os.ReadFile(filepath.Join(svc, "quickshell", "supervise", "stat")); string(b) != "run\n" {
		t.Errorf("runit state was touched: %q", b)
	}
	if _, err := os.Stat(filepath.Join(svc, "asurad", "down")); err == nil {
		t.Error("an installed `down` was copied over the user's choice")
	}
	b, _ := os.ReadFile(filepath.Join(home, ".config", "asura", "hyprland", "shell.conf"))
	if !strings.Contains(string(b), "source = "+look) {
		t.Errorf("shell.conf does not source the desktop look: %q", b)
	}

	// ASURA's own shell: the look file is empty of sources.
	t.Setenv("ASURA_DESKTOP_SHELL", "asura")
	if err := writeShellLook(runnertest.New().Run); err != nil {
		t.Fatal(err)
	}
	b, _ = os.ReadFile(filepath.Join(home, ".config", "asura", "hyprland", "shell.conf"))
	if strings.Contains(string(b), "source") {
		t.Errorf("ASURA's shell got the desktop look: %q", b)
	}
}

func TestSyncUserServicesNeedsAServiceDir(t *testing.T) {
	src := t.TempDir()
	if err := os.MkdirAll(filepath.Join(src, "quickshell"), 0o755); err != nil {
		t.Fatal(err)
	}
	n, err := syncUserServices(src, filepath.Join(t.TempDir(), "missing"))
	if err != nil || n != 0 {
		t.Fatalf("got %d, %v; want nothing done", n, err)
	}
}
