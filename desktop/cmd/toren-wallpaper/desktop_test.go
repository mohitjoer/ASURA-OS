package main

import "testing"

func TestWallpaperKeysOpenDesktopInTheDesktopShell(t *testing.T) {
	savedShell, savedCall := desktopShell, callDesktopIPC
	defer func() { desktopShell, callDesktopIPC = savedShell, savedCall }()

	var calls []string
	callDesktopIPC = func(target, fn string) string {
		calls = append(calls, target+"."+fn)
		return "ok"
	}

	desktopShell = func() string { return "toren" }
	if isDesktopShell() {
		t.Fatal("the toren shell was taken for desktop")
	}

	desktopShell = func() string { return "desktop" }
	// isDesktopShell also needs the installed tree, which a test machine lacks.
	if isDesktopShell() {
		t.Skip("/usr/share/toren/shell-desktop exists here; skipping the absent-tree case")
	}
	if len(calls) != 0 {
		t.Fatalf("desktop was called while not installed: %v", calls)
	}
}

func TestApplyDesktopWallpaperCallsSwitchwall(t *testing.T) {
	savedShell, savedApply := desktopShell, applyDesktopWallpaper
	defer func() { desktopShell, applyDesktopWallpaper = savedShell, savedApply }()

	var appliedPath string
	applyDesktopWallpaper = func(path string) {
		appliedPath = path
	}

	desktopShell = func() string { return "desktop" }
	applyDesktopWallpaper("/path/to/test.png")
	if appliedPath != "/path/to/test.png" {
		t.Fatalf("expected /path/to/test.png, got %q", appliedPath)
	}
}
