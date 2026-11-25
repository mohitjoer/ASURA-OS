package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// Keeping a machine current, beyond `toren update` itself:
//
//	toren channel set stable|unstable   which branch updates follow
//	toren firmware [check|update]       device firmware through fwupd
//	toren clean                         orphaned packages and the package cache
//	toren notify-updates                one check for the desktop notice (run by
//	                                    the toren-updates user service)
//
// and the "reboot needed" check every update ends with.

// channelFile holds the channel chosen with `toren channel set`. It is a file
// of its own because apply-repo.sh rewrites /etc/xbps.d/toren.conf on every
// update; nothing but `toren channel set` writes this one.
var channelFile = "/etc/toren/channel"

// channelBranches maps a channel to the branch updates follow. stable follows
// main, where releases are cut; unstable follows dev-unstable, which every
// merged change reaches first. Both only take commits that passed CI.
var channelBranches = map[string]string{
	"stable":   "main",
	"unstable": "dev-unstable",
}

// pinnedChannel is the channel in channelFile under root, or "".
func pinnedChannel(root string) string {
	b, err := os.ReadFile(filepath.Join(root, channelFile))
	if err != nil {
		return ""
	}
	ch := strings.TrimSpace(string(b))
	if _, ok := channelBranches[ch]; ok {
		return ch
	}
	return ""
}

// channelSet writes the channel. It is called from `toren channel set`.
func channelSet(root, ch string) int {
	branch, ok := channelBranches[ch]
	if !ok {
		fmt.Fprintf(os.Stderr, "toren channel set: %q is not a channel (stable or unstable)\n", ch)
		return exitUsage
	}
	path := filepath.Join(root, channelFile)
	write := mirrorWrite
	if root != "" {
		write = func(p, c string) error {
			if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
				return err
			}
			return os.WriteFile(p, []byte(c), 0o644)
		}
	}
	if err := write(path, ch+"\n"); err != nil {
		fmt.Fprintln(os.Stderr, "toren channel set:", err)
		return exitFail
	}
	fmt.Printf("channel: %s (updates follow %s; only commits that passed CI)\n", ch, branch)
	return exitOK
}

// ── Reboot needed ───────────────────────────────────────────────────────────

// Where kernels keep their modules, and how the running kernel is read;
// tests point these elsewhere.
var (
	modulesDir   = "/usr/lib/modules"
	bootedKernel = func() string {
		out, err := exec.Command("uname", "-r").Output()
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(out))
	}
)

// newestKernel is the newest kernel with modules installed, or "".
func newestKernel(dir string) string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	var kernels []string
	for _, e := range entries {
		if e.IsDir() && e.Name() != "" && e.Name()[0] >= '0' && e.Name()[0] <= '9' {
			kernels = append(kernels, e.Name())
		}
	}
	sort.Slice(kernels, func(i, j int) bool {
		return compareVersions(strings.ReplaceAll(kernels[i], "_", "."), strings.ReplaceAll(kernels[j], "_", ".")) < 0
	})
	if len(kernels) == 0 {
		return ""
	}
	return kernels[len(kernels)-1]
}

// rebootNeeded names the newer kernel when one is installed but not running.
func rebootNeeded() (string, bool) {
	running, newest := bootedKernel(), newestKernel(modulesDir)
	if running == "" || newest == "" || running == newest {
		return "", false
	}
	if compareVersions(strings.ReplaceAll(newest, "_", "."), strings.ReplaceAll(running, "_", ".")) <= 0 {
		return "", false
	}
	return newest, true
}

// reportRebootNeeded prints the notice at the end of an update.
func reportRebootNeeded() {
	if newest, ok := rebootNeeded(); ok {
		fmt.Printf("\nReboot to use the new kernel (%s; running %s).\n", newest, bootedKernel())
	}
}

// ── Firmware ────────────────────────────────────────────────────────────────

const firmwareUsage = `Usage: toren firmware [check | update]

Device firmware (laptop, SSD, dock, keyboard...) from the vendors' service, through fwupd.

  check    refresh the list and show available updates (the default)
  update   install them; some apply at the next reboot
`

func init() {
	register("firmware", "check or install device firmware updates (fwupd)", torenFirmware)
	register("clean", "remove orphaned packages and the downloaded package cache", torenClean)
	register("notify-updates", "check once for an update or a needed reboot, and notify the desktop", torenNotifyUpdates)
}

var fwupdRun = func(args ...string) error {
	cmd := exec.Command("fwupdmgr", args...)
	cmd.Stdout, cmd.Stderr, cmd.Stdin = os.Stdout, os.Stderr, os.Stdin
	return cmd.Run()
}

func torenFirmware(args []string) int {
	mode := "check"
	if len(args) > 0 {
		mode = args[0]
	}
	if mode == "-h" || mode == "--help" {
		fmt.Print(firmwareUsage)
		return exitOK
	}
	if _, err := exec.LookPath("fwupdmgr"); err != nil {
		fmt.Fprintln(os.Stderr, "toren firmware: fwupd is not installed (sudo xbps-install -S fwupd)")
		return exitFail
	}
	switch mode {
	case "check":
		_ = fwupdRun("refresh", "--force")
		// fwupdmgr exits 2 when there is nothing to update.
		if err := fwupdRun("get-updates"); err != nil {
			if ee, ok := err.(*exec.ExitError); ok && ee.ExitCode() == 2 {
				fmt.Println("firmware: everything is up to date")
				return exitOK
			}
			return exitFail
		}
		return exitOK
	case "update":
		if err := fwupdRun("update"); err != nil {
			if ee, ok := err.(*exec.ExitError); ok && ee.ExitCode() == 2 {
				fmt.Println("firmware: nothing to update")
				return exitOK
			}
			return exitFail
		}
		return exitOK
	}
	fmt.Fprint(os.Stderr, firmwareUsage)
	return exitUsage
}

// ── Clean ───────────────────────────────────────────────────────────────────

const cleanUsage = `Usage: toren clean

Removes packages nothing needs any more (installed only as dependencies of
something since removed) and the downloaded package files in the cache. XBPS
lists what it will remove and asks before it does. Kernels are left alone.
`

func torenClean(args []string) int {
	if len(args) > 0 {
		fmt.Print(cleanUsage)
		if args[0] == "-h" || args[0] == "--help" {
			return exitOK
		}
		return exitUsage
	}
	if _, err := exec.LookPath("xbps-remove"); err != nil {
		fmt.Fprintln(os.Stderr, "toren clean: xbps-remove is not installed")
		return exitFail
	}
	fmt.Println("==> Packages nothing needs any more")
	if err := runPrivileged("xbps-remove", "-o"); err != nil {
		fmt.Fprintln(os.Stderr, "toren clean: removing orphans:", err)
		return exitFail
	}
	fmt.Println("==> The downloaded package cache")
	if err := runPrivileged("xbps-remove", "-O"); err != nil {
		fmt.Fprintln(os.Stderr, "toren clean: cleaning the cache:", err)
		return exitFail
	}
	return exitOK
}

// ── The desktop notice ──────────────────────────────────────────────────────

var (
	notifySend = func(title, body string) error {
		return exec.Command("notify-send", "-a", "ASURA", "-i", "system-software-update", title, body).Run()
	}
	updateCheckOutput = func() (string, int) {
		cmd := exec.Command(selfPath(), "update", "--check")
		out, err := cmd.CombinedOutput()
		code := 0
		if ee, ok := err.(*exec.ExitError); ok {
			code = ee.ExitCode()
		} else if err != nil {
			code = exitFail
		}
		return string(out), code
	}
)

// notifyStateDir remembers what was last announced, so a notice is shown once
// per new update or kernel, not every check.
func notifyStateDir() string {
	base := os.Getenv("XDG_STATE_HOME")
	if base == "" {
		base = filepath.Join(homeDir(), ".local", "state")
	}
	return filepath.Join(base, "toren", "notified")
}

// notifyOnce sends the notice unless key was the last one sent for kind.
func notifyOnce(kind, key, title, body string) {
	dir := notifyStateDir()
	path := filepath.Join(dir, kind)
	if b, err := os.ReadFile(path); err == nil && strings.TrimSpace(string(b)) == key {
		return
	}
	if err := notifySend(title, body); err != nil {
		return
	}
	_ = os.MkdirAll(dir, 0o755)
	_ = os.WriteFile(path, []byte(key+"\n"), 0o644)
}

func torenNotifyUpdates(args []string) int {
	out, code := updateCheckOutput()
	if code == exitNoMatch {
		line := ""
		for _, l := range strings.Split(out, "\n") {
			if strings.Contains(l, "update available") {
				line = strings.TrimSpace(l)
			}
		}
		if line == "" {
			line = "update available"
		}
		notifyOnce("update", line, "An ASURA update is available", "Run toren update in a terminal. ("+line+")")
	}
	if newest, ok := rebootNeeded(); ok {
		notifyOnce("reboot", newest, "Reboot to finish updating", "A new kernel ("+newest+") is installed; it is used after a reboot.")
	}
	return exitOK
}
