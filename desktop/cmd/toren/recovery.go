package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func init() {
	register("recovery", "guided recovery menu on TTY (diagnostics, reset, snapshots, safe-mode)", torenRecovery)
}

// torenRecovery launches the interactive recovery menu for TTY.
func torenRecovery(args []string) int {
	return runRecoveryMenu(defaultRecoveryRunner())
}

type recoveryRunner struct {
	in        io.Reader
	out       io.Writer
	doctor    func() int
	reset     func() int
	reinstall func() int
	rollback  func() int
	safeMode  func() int
	showLogs  func() int
}

func defaultRecoveryRunner() *recoveryRunner {
	return &recoveryRunner{
		in:  os.Stdin,
		out: os.Stdout,
		doctor: func() int {
			return torenDoctor([]string{})
		},
		reset: func() int {
			return torenReset([]string{"--desktop"})
		},
		reinstall: func() int {
			return torenReinstall([]string{})
		},
		rollback: func() int {
			return torenRollback([]string{"--system"})
		},
		safeMode: func() int {
			fmt.Println("Starting safe-mode session via /usr/share/toren/session/toren-session-safe...")
			safeBin := "/usr/share/toren/session/toren-session-safe"
			if _, err := os.Stat(safeBin); err != nil {
				safeBin = "session/toren-session-safe"
			}
			cmd := exec.Command(safeBin)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "failed to run safe-mode session: %v\n", err)
				return exitFail
			}
			return exitOK
		},
		showLogs: func() int {
			printRecentLogs(os.Stdout)
			return exitOK
		},
	}
}

func printRecentLogs(out io.Writer) {
	home := os.Getenv("HOME")
	stateDir := filepath.Join(home, ".local", "state", "toren")
	logFiles := []string{
		filepath.Join(stateDir, "session.log"),
		filepath.Join(stateDir, "quickshell.log"),
		filepath.Join(stateDir, "session-safe.log"),
	}
	found := false
	for _, f := range logFiles {
		if data, err := os.ReadFile(f); err == nil && len(data) > 0 {
			found = true
			fmt.Fprintf(out, "\n=== %s ===\n", filepath.Base(f))
			lines := strings.Split(string(data), "\n")
			start := 0
			if len(lines) > 25 {
				start = len(lines) - 25
			}
			for _, line := range lines[start:] {
				fmt.Fprintln(out, line)
			}
		}
	}
	if !found {
		fmt.Fprintln(out, "No recent ASURA log files found in ~/.local/state/toren/")
	}
}

func confirm(reader *bufio.Reader, out io.Writer, prompt string) bool {
	fmt.Fprint(out, prompt)
	line, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	ans := strings.ToLower(strings.TrimSpace(line))
	return ans == "y" || ans == "yes"
}

func runRecoveryMenu(r *recoveryRunner) int {
	reader := bufio.NewReader(r.in)
	for {
		fmt.Fprintln(r.out, "")
		fmt.Fprintln(r.out, "============================================================")
		fmt.Fprintln(r.out, "                ASURA System Recovery                       ")
		fmt.Fprintln(r.out, "============================================================")
		fmt.Fprintln(r.out, "1) Run Doctor (system diagnostics & configuration drift)")
		fmt.Fprintln(r.out, "2) Reset Desktop (restore defaults, back up current files)")
		fmt.Fprintln(r.out, "3) Reinstall ASURA (re-apply files for installed commit)")
		fmt.Fprintln(r.out, "4) System Snapshots (list & restore Btrfs root snapshots)")
		fmt.Fprintln(r.out, "5) Safe-Mode Session (start minimal recovery desktop)")
		fmt.Fprintln(r.out, "6) Show Logs (inspect recent session and shell error logs)")
		fmt.Fprintln(r.out, "7) Exit")
		fmt.Fprintln(r.out, "------------------------------------------------------------")
		fmt.Fprint(r.out, "Select an option [1-7]: ")

		line, err := reader.ReadString('\n')
		if err != nil {
			return exitOK
		}
		choice := strings.TrimSpace(line)
		switch choice {
		case "1":
			fmt.Fprintln(r.out, "\n[Option 1: System Doctor]")
			fmt.Fprintln(r.out, "This runs 'toren doctor' to inspect service states, missing tools,")
			fmt.Fprintln(r.out, "and any configuration drift between packaged base and user settings.")
			if confirm(reader, r.out, "Run system doctor now? [y/N]: ") {
				r.doctor()
			}
		case "2":
			fmt.Fprintln(r.out, "\n[Option 2: Reset Desktop Configuration]")
			fmt.Fprintln(r.out, "This moves your current desktop configuration (desktop config.json,")
			fmt.Fprintln(r.out, "Hyprland overrides, shell state) into a dated backup folder under")
			fmt.Fprintln(r.out, "~/.local/state/toren/backups/ and restores clean ASURA defaults.")
			fmt.Fprintln(r.out, "Nothing outside your desktop configuration is touched.")
			if confirm(reader, r.out, "Reset desktop configuration to defaults? [y/N]: ") {
				r.reset()
			}
		case "3":
			fmt.Fprintln(r.out, "\n[Option 3: Reinstall Shipped ASURA Files]")
			fmt.Fprintln(r.out, "This re-applies repository files for the exact commit currently installed")
			fmt.Fprintln(r.out, "to repair deleted or damaged system files, without updating newer code.")
			if confirm(reader, r.out, "Reinstall ASURA files now? [y/N]: ") {
				r.reinstall()
			}
		case "4":
			fmt.Fprintln(r.out, "\n[Option 4: System Snapshots]")
			fmt.Fprintln(r.out, "This lists available Btrfs root snapshots in /.snapshots and allows you")
			fmt.Fprintln(r.out, "to stage a boot rollback to a known-good earlier snapshot.")
			if confirm(reader, r.out, "Inspect and manage snapshots? [y/N]: ") {
				r.rollback()
			}
		case "5":
			fmt.Fprintln(r.out, "\n[Option 5: Safe-Mode Session]")
			fmt.Fprintln(r.out, "This starts a minimal Hyprland recovery session with a terminal window,")
			fmt.Fprintln(r.out, "no desktop shell, and no user plugins, allowing manual repair.")
			if confirm(reader, r.out, "Launch safe-mode session now? [y/N]: ") {
				r.safeMode()
			}
		case "6":
			fmt.Fprintln(r.out, "\n[Option 6: View System & Session Logs]")
			fmt.Fprintln(r.out, "This reads and displays the most recent 25 lines from session.log,")
			fmt.Fprintln(r.out, "quickshell.log, and recovery logs.")
			if confirm(reader, r.out, "Display recent logs? [y/N]: ") {
				r.showLogs()
			}
		case "7", "q", "exit", "quit":
			fmt.Fprintln(r.out, "Exiting recovery menu.")
			return exitOK
		default:
			fmt.Fprintln(r.out, "Invalid choice. Please select 1 through 7.")
		}
	}
}
