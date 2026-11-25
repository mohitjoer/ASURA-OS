package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

const greeterUsage = `Usage: toren greeter enable|disable|test

Toren's login screen is an SDDM theme. Nothing installs it for you: a login
screen that does not come up leaves nobody able to log in, so the one file that
turns it on is copied by you, with sudo, after this command has checked that
everything the screen needs is installed.

  enable    check what is installed and print the command that turns it on
  disable   print the command that turns it off
  test      show the screen on this desktop for a minute (Esc leaves sooner);
            logging in is switched off
`

const (
	greeterDropIn    = "/usr/share/toren/session/sddm/toren.conf"
	greeterInstalled = "/etc/sddm.conf.d/10-toren.conf"
	greeterTheme     = "/usr/share/toren/session/sddm/themes/toren"
	greeterWrapper   = "/usr/share/toren/session/toren-greeter"
	greeterBinary    = "/usr/bin/sddm-greeter-qt6"
)

var (
	// greeterRoot prefixes every path, so a test can lay out a machine.
	greeterRoot = ""
	// greeterTestTime is how long `test` shows the screen. Without a limit a
	// theme that ignored Esc would hold every screen until a reboot.
	greeterTestTime                  = 60 * time.Second
	greeterOut, greeterErr io.Writer = os.Stdout, os.Stderr
)

func init() {
	register("greeter", "turn the Toren login screen on or off, or try it", torenGreeter)
}

func torenGreeter(args []string) int {
	if len(args) == 1 {
		switch args[0] {
		case "enable":
			return greeterEnable()
		case "disable":
			return greeterDisable()
		case "test":
			return greeterTest()
		case "-h", "--help":
			fmt.Fprint(greeterOut, greeterUsage)
			return exitOK
		}
	}
	fmt.Fprint(greeterErr, greeterUsage)
	return exitUsage
}

// greeterMissing lists what the login screen needs that this machine lacks.
//
// Each item is fatal on its own: without the SDDM greeter or Weston, SDDM has
// nothing to start; without the wrapper it cannot start Weston; and a missing
// theme makes SDDM fall back to its built-in one without saying so.
func greeterMissing() []string {
	var missing []string
	need := func(path, fix string, executable bool) {
		info, err := os.Stat(greeterRoot + path)
		if err != nil || info.IsDir() || (executable && info.Mode()&0o111 == 0) {
			missing = append(missing, path+" ("+fix+")")
		}
	}
	need(greeterBinary, "sudo xbps-install -S sddm", true)
	need("/usr/bin/weston", "sudo xbps-install -S weston", true)
	need(greeterTheme+"/Main.qml", "reinstall the toren package", false)
	need(greeterWrapper, "reinstall the toren package", true)
	need(greeterDropIn, "reinstall the toren package", false)
	return missing
}

func greeterEnable() int {
	if missing := greeterMissing(); len(missing) > 0 {
		fmt.Fprintln(greeterErr, "toren greeter: not ready, so there is nothing to enable yet. Missing:")
		for _, m := range missing {
			fmt.Fprintln(greeterErr, "  "+m)
		}
		return exitFail
	}

	// The drop-in only matters to the display manager that reads it.
	manager, _ := os.ReadFile(greeterRoot + "/etc/X11/default-display-manager")
	if !strings.HasSuffix(strings.TrimSpace(string(manager)), "/sddm") {
		fmt.Fprintln(greeterOut, "Note: sddm is not this machine's display manager, so the file below does nothing until it is.")
		fmt.Fprintln(greeterOut)
	}
	fmt.Fprintf(greeterOut, "Ready. To turn the Toren login screen on, run:\n\n  sudo install -D -m 0644 %s %s\n\n", greeterDropIn, greeterInstalled)
	fmt.Fprintf(greeterOut, "If the login screen ever fails, switch to a text console with Ctrl+Alt+F3,\nlog in, and run:\n\n  sudo rm %s\n", greeterInstalled)
	return exitOK
}

func greeterDisable() int {
	if _, err := os.Stat(greeterRoot + greeterInstalled); err != nil {
		fmt.Fprintln(greeterOut, "The Toren login screen is not enabled; there is nothing to remove.")
		return exitOK
	}
	fmt.Fprintf(greeterOut, "To turn the Toren login screen off, run:\n\n  sudo rm %s\n", greeterInstalled)
	return exitOK
}

func greeterTest() int {
	if _, err := os.Stat(greeterRoot + greeterBinary); err != nil {
		fmt.Fprintln(greeterErr, "toren greeter: not available: sddm-greeter-qt6 is not installed (sudo xbps-install -S sddm)")
		return exitFail
	}
	if os.Getenv("WAYLAND_DISPLAY") == "" && os.Getenv("DISPLAY") == "" {
		fmt.Fprintln(greeterErr, "toren greeter: test needs a graphical session to draw in")
		return exitFail
	}

	ctx, cancel := context.WithTimeout(context.Background(), greeterTestTime)
	defer cancel()
	cmd := exec.CommandContext(ctx, greeterRoot+greeterBinary, "--test-mode", "--theme", greeterRoot+greeterTheme)
	cmd.Stdout, cmd.Stderr = greeterOut, greeterErr
	fmt.Fprintf(greeterOut, "Showing the login screen for %s. Esc leaves sooner; logging in is switched off.\n", greeterTestTime)

	err := cmd.Run()
	if ctx.Err() != nil {
		return exitOK
	}
	if err != nil {
		fmt.Fprintln(greeterErr, "toren greeter:", err)
		return exitFail
	}
	return exitOK
}
