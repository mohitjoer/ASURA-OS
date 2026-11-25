package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"asura/desktop/internal/runner"
)

// `asura refresh` brings the running desktop up to what is installed, without
// logging out. `asura update` runs it after applying a pulled checkout:
//
//  1. the user's runit services are re-synced from the installed ones, so a
//     changed run script (the shell's, a daemon's) is the one that restarts;
//  2. the compositor look for the chosen desktop shell is rewritten (the same
//     file asura-session writes at login) and Hyprland reloads its config;
//  3. the desktop shell and ASURA's user daemons restart on the new code;
//  4. a notification says it happened.
//
// Audio (pipewire, wireplumber) and the portals are left running: restarting
// them would cut sound and break open file dialogs for nothing.

const refreshUsage = `Usage: asura refresh

Applies installed desktop changes to the running session: re-syncs the user
services, reloads Hyprland, restarts the desktop shell and ASURA's daemons.
Run as root (sudo asura update), it refreshes the session of the user who ran
sudo.
`

// The installed user services, and the ones refresh restarts when running.
var (
	refreshServicesSrc = "/usr/share/asura/runit/services"
	refreshShellLook   = "/usr/share/asura/shell-desktop/toren/hyprland.conf"
	refreshDaemons     = []string{"asurad", "asura-userd", "asura-hubd", "asura-wallpaper", "asura-idle"}
)

func init() {
	register("refresh", "apply installed desktop changes to the running session", asuraRefresh)
}

func asuraRefresh(args []string) int {
	if len(args) != 0 {
		if args[0] == "-h" || args[0] == "--help" {
			fmt.Print(refreshUsage)
			return exitOK
		}
		fmt.Fprint(os.Stderr, refreshUsage)
		return exitUsage
	}
	if os.Geteuid() == 0 {
		return refreshAsSudoUser(os.Stdout, os.Stderr)
	}
	return refreshMain(os.Stdout, os.Stderr, runner.Exec(20*time.Second))
}

// refreshAsSudoUser: root has no desktop session of its own. Under sudo the
// session belongs to SUDO_USER, so refresh runs again as that user, in that
// user's runtime directory.
func refreshAsSudoUser(out, errOut io.Writer) int {
	name := os.Getenv("SUDO_USER")
	if name == "" || name == "root" {
		fmt.Fprintln(out, "refresh: run as root with no SUDO_USER; no session to refresh")
		return exitOK
	}
	u, err := user.Lookup(name)
	if err != nil {
		fmt.Fprintf(errOut, "refresh: %v\n", err)
		return exitFail
	}
	self, err := os.Executable()
	if err != nil {
		self = "asura"
	}
	cmd := exec.Command("sudo", "-u", name, "env",
		"HOME="+u.HomeDir,
		"XDG_RUNTIME_DIR=/run/user/"+u.Uid,
		"DBUS_SESSION_BUS_ADDRESS=unix:path=/run/user/"+u.Uid+"/bus",
		self, "refresh")
	cmd.Stdout, cmd.Stderr = out, errOut
	if err := cmd.Run(); err != nil {
		return exitFail
	}
	return exitOK
}

// refreshMain takes the runner so a test never reaches the live session.
func refreshMain(out, errOut io.Writer, run runner.Func) int {
	code := exitOK
	svcDir := userServiceDir()

	// 1. Services.
	if n, err := syncUserServices(refreshServicesSrc, svcDir); err != nil {
		fmt.Fprintf(errOut, "services:   not synced: %v\n", err)
		code = exitFail
	} else {
		fmt.Fprintf(out, "services:   %d synced\n", n)
	}

	// 2. Compositor.
	if err := writeShellLook(run); err != nil {
		fmt.Fprintf(errOut, "look:       not written: %v\n", err)
	}
	env := hyprlandEnv()
	if env == "" {
		fmt.Fprintln(out, "compositor: not running")
	} else if _, err := run("env", env, "hyprctl", "reload"); err != nil {
		fmt.Fprintf(errOut, "compositor: not reloaded: %s\n", runner.Reason(err))
		code = exitFail
	} else {
		fmt.Fprintln(out, "compositor: reloaded")
	}

	// 3. The shell. sv restart, and when the session has no runsvdir (or sv
	// times out), end the process: the session's supervisor starts it again.
	// The shell's run script pauses before restarting the desktop when it exits
	// within 15 s three times in a row. A restart asked for here is not a
	// crash: clear that count first, so a refresh is never delayed.
	clearShellCrashCount()
	shellPath := filepath.Join(svcDir, reloadShellService)
	if _, err := run("sv", "restart", shellPath); err == nil {
		fmt.Fprintln(out, "shell:      restarted")
	} else if _, kerr := run("pkill", "-x", "quickshell"); kerr == nil {
		fmt.Fprintln(out, "shell:      restarted")
	} else {
		fmt.Fprintf(errOut, "shell:      not restarted: %s\n", runner.Reason(err))
		code = exitFail
	}

	// Daemons that are running restart on their new binaries.
	var restarted []string
	for _, d := range refreshDaemons {
		p := filepath.Join(svcDir, d)
		if _, err := run("sv", "status", p); err != nil {
			continue
		}
		if _, err := run("sv", "restart", p); err != nil {
			fmt.Fprintf(errOut, "daemon:     %s not restarted: %s\n", d, runner.Reason(err))
			code = exitFail
			continue
		}
		restarted = append(restarted, d)
	}
	if len(restarted) > 0 {
		fmt.Fprintf(out, "daemons:    restarted %s\n", strings.Join(restarted, ", "))
	}

	// 4. Tell the user; no notification daemon is not an error.
	_, _ = run("notify-send", "-a", "ASURA", "-i", "system-software-update", "ASURA updated", "The desktop was refreshed with the latest changes.")
	return code
}

// syncUserServices copies the installed user services over the user's copies:
// every file but runit's own state (supervise/) and the user's `down` choice.
// A service directory the user does not have yet is added, and runsvdir starts
// it. Nothing is done when the user has no service directory at all.
func syncUserServices(src, dst string) (int, error) {
	if _, err := os.Stat(dst); err != nil {
		return 0, nil
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	n := 0
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		from, to := filepath.Join(src, e.Name()), filepath.Join(dst, e.Name())
		err := filepath.WalkDir(from, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			rel, _ := filepath.Rel(from, path)
			if d.IsDir() {
				if d.Name() == "supervise" {
					return filepath.SkipDir
				}
				return os.MkdirAll(filepath.Join(to, rel), 0o755)
			}
			if d.Name() == "down" || !d.Type().IsRegular() {
				return nil
			}
			return copyFileAtomic(path, filepath.Join(to, rel))
		})
		if err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}

func copyFileAtomic(from, to string) error {
	data, err := os.ReadFile(from)
	if err != nil {
		return err
	}
	info, err := os.Stat(from)
	if err != nil {
		return err
	}
	tmp := to + ".asura-new"
	if err := os.WriteFile(tmp, data, info.Mode().Perm()); err != nil {
		return err
	}
	if err := os.Chmod(tmp, info.Mode().Perm()); err != nil {
		return err
	}
	return os.Rename(tmp, to)
}

// writeShellLook writes ~/.config/asura/hyprland/shell.conf the way
// asura-session does at login: the desktop's compositor look while it is the
// desktop shell, nothing for ASURA's shell.
func writeShellLook(run runner.Func) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dir := filepath.Join(home, ".config", "asura", "hyprland")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	shell := os.Getenv("ASURA_DESKTOP_SHELL")
	if shell == "" {
		if got, err := run("gsettings", "get", "org.asura.shell", "desktop-shell"); err == nil {
			shell = strings.Trim(strings.TrimSpace(got), "'")
		}
	}
	text := "# Written by asura-session at login; do not edit.\n"
	if (shell == "" || shell == "desktop") && fileExists(refreshShellLook) {
		text += "source = " + refreshShellLook + "\n"
	}
	return os.WriteFile(filepath.Join(dir, "shell.conf"), []byte(text), 0o644)
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

// hyprlandEnv names the running Hyprland instance for hyprctl, which needs
// HYPRLAND_INSTANCE_SIGNATURE; under sudo or from a TTY it is not in the
// environment, so take the newest instance with a live socket.
func hyprlandEnv() string {
	if sig := os.Getenv("HYPRLAND_INSTANCE_SIGNATURE"); sig != "" {
		return "HYPRLAND_INSTANCE_SIGNATURE=" + sig
	}
	rt := os.Getenv("XDG_RUNTIME_DIR")
	if rt == "" {
		rt = "/run/user/" + strconv.Itoa(os.Getuid())
	}
	dirs, _ := filepath.Glob(filepath.Join(rt, "hypr", "*"))
	var best string
	var bestTime time.Time
	for _, d := range dirs {
		info, err := os.Stat(filepath.Join(d, ".socket.sock"))
		if err != nil {
			continue
		}
		if best == "" || info.ModTime().After(bestTime) {
			best, bestTime = filepath.Base(d), info.ModTime()
		}
	}
	if best == "" {
		return ""
	}
	return "HYPRLAND_INSTANCE_SIGNATURE=" + best
}

// clearShellCrashCount removes the quick-exit count the quickshell run
// script keeps (system/runit/user-services/quickshell/run).
func clearShellCrashCount() {
	dir := os.Getenv("XDG_STATE_HOME")
	if dir == "" {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, ".local", "state")
	}
	for _, f := range []string{"desktop-quick-exits", "desktop-last-start"} {
		_ = os.Remove(filepath.Join(dir, "asura", f))
	}
}
