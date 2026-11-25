package main

import (
	"fmt"
	"strconv"
	"strings"

	"asura/desktop/cmd/toren-hubd/internal/wm"
)

// autostart manages login-time user commands in desktop.json,
// which are rendered to settings.conf as `exec-once = ...`.
func runAutostart(args []string) error {
	if len(args) == 0 || args[0] == "list" {
		return listAutostart()
	}
	switch args[0] {
	case "add":
		if len(args) < 2 {
			return fmt.Errorf("autostart add needs <command>")
		}
		cmd := strings.Join(args[1:], " ")
		return addAutostart(cmd)
	case "remove", "rm":
		if len(args) < 2 {
			return fmt.Errorf("autostart remove needs <index>")
		}
		idx, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid autostart index %q: %w", args[1], err)
		}
		return removeAutostart(idx)
	default:
		return fmt.Errorf("unknown autostart subcommand %q", args[0])
	}
}

func listAutostart() error {
	o := wm.LoadStore(desktopStorePath())
	if o.Autostart == nil {
		return printJSON([]wm.Autostart{})
	}
	return printJSON(o.Autostart)
}

func addAutostart(cmd string) error {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return fmt.Errorf("autostart command cannot be empty")
	}
	return withDesktopLock(func() error {
		o := wm.LoadStore(desktopStorePath())
		o.Autostart = append(o.Autostart, wm.Autostart{Command: cmd})
		if err := setHyprSections(map[string]any{"autostart": o.Autostart}); err != nil {
			return err
		}
		if err := applyLive(); err != nil {
			return err
		}
		return printJSON(map[string]any{
			"status":  "added",
			"command": cmd,
		})
	})
}

func removeAutostart(index int) error {
	return withDesktopLock(func() error {
		o := wm.LoadStore(desktopStorePath())
		if index < 0 || index >= len(o.Autostart) {
			return fmt.Errorf("autostart index %d out of range (total %d)", index, len(o.Autostart))
		}
		removed := o.Autostart[index]
		o.Autostart = append(o.Autostart[:index], o.Autostart[index+1:]...)
		if err := setHyprSections(map[string]any{"autostart": o.Autostart}); err != nil {
			return err
		}
		if err := applyLive(); err != nil {
			return err
		}
		return printJSON(map[string]any{
			"status":  "removed",
			"index":   index,
			"removed": removed,
		})
	})
}
