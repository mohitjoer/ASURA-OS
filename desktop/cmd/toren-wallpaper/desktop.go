package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// The desktop shell (org.asura.shell desktop-shell = desktop) draws its own
// background, wallpaper and palette, over this daemon's layer. When a wallpaper
// is applied, applyDesktopWallpaper forwards the image to the desktop's switchwall.sh
// so the background and Material 3 color palette stay synchronized.

// desktopShellDir is where the desktop shell is installed.
const desktopShellDir = "/usr/share/asura/shell-desktop"

// desktopShell reads the configured desktop: ASURA_DESKTOP_SHELL, then the
// GSettings key, then desktop, the schema's default.
var desktopShell = func() string {
	if v := strings.TrimSpace(os.Getenv("ASURA_DESKTOP_SHELL")); v != "" {
		return v
	}
	out, err := exec.Command("gsettings", "get", "org.asura.shell", "desktop-shell").Output()
	if err != nil {
		return "desktop"
	}
	return strings.Trim(strings.TrimSpace(string(out)), "'")
}

func isDesktopShell() bool {
	if desktopShell() != "desktop" {
		return false
	}
	_, err := os.Stat(desktopShellDir)
	return err == nil
}

// callDesktopIPC calls one of the desktop shell's IPC functions.
var callDesktopIPC = func(target, fn string) string {
	cmd := exec.Command("quickshell", "ipc", "--path", desktopShellDir, "--any-display", "call", target, fn)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "err wallpaper: the desktop did not answer: " + strings.TrimSpace(string(out))
	}
	return "ok"
}

// applyDesktopWallpaper runs the desktop's switchwall.sh script asynchronously so the
// desktop background and matugen color palette update to match the applied wallpaper.
var applyDesktopWallpaper = func(path string) {
	if !isDesktopShell() || path == "" {
		return
	}
	script := filepath.Join(desktopShellDir, "scripts", "colors", "switchwall.sh")
	if _, err := os.Stat(script); err != nil {
		return
	}
	cmd := exec.Command(script, "--image", path)
	_ = cmd.Start()
}
