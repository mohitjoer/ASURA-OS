package main

import (
	"fmt"
	"strconv"
	"strings"

	"toren/desktop/cmd/toren-hubd/internal/wm"
)

// rules is the backend for the Window Rules page in Settings Hub.
// It manages custom window rules (float, size, workspace, opacity, etc.)
// written to user configuration and applied live to Hyprland.

func runRules(args []string) error {
	if len(args) == 0 || args[0] == "list" {
		return listRules()
	}
	switch args[0] {
	case "add":
		if len(args) < 3 {
			return fmt.Errorf("rules add needs <class> <action> [value]")
		}
		val := ""
		if len(args) >= 4 {
			val = args[3]
		}
		return addRule(args[1], args[2], val)
	case "remove", "rm":
		if len(args) < 2 {
			return fmt.Errorf("rules remove needs <index>")
		}
		idx, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid rule index %q: %w", args[1], err)
		}
		return removeRule(idx)
	default:
		return fmt.Errorf("unknown rules subcommand %q", args[0])
	}
}

func listRules() error {
	o := wm.LoadStore(desktopStorePath())
	if o.WindowRules == nil {
		return printJSON([]wm.WindowRule{})
	}
	return printJSON(o.WindowRules)
}

func addRule(class, action, value string) error {
	class = strings.TrimSpace(class)
	action = strings.ToLower(strings.TrimSpace(action))
	value = strings.TrimSpace(value)

	if class == "" {
		return fmt.Errorf("window rule requires an app class")
	}

	validActions := map[string]bool{
		"float": true, "tile": true, "size": true, "workspace": true,
		"opacity": true, "pin": true, "fullscreen": true, "center": true,
		"noblur": true, "blur": true, "noborder": true, "noshadow": true,
		"norounding": true, "nodim": true, "noanim": true, "opaque": true,
		"xray": true, "idleinhibit": true,
	}
	if !validActions[action] {
		return fmt.Errorf("unsupported window rule action %q", action)
	}

	return withDesktopLock(func() error {
		o := wm.LoadStore(desktopStorePath())
		o.WindowRules = append(o.WindowRules, wm.WindowRule{
			Class:  class,
			Action: action,
			Value:  value,
		})
		if err := setHyprSections(map[string]any{"windowRules": o.WindowRules}); err != nil {
			return err
		}
		if err := applyLive(); err != nil {
			return err
		}
		if err := wm.CheckConfigErrors(); err != nil {
			// Roll back if Hyprland reports a config error
			o.WindowRules = o.WindowRules[:len(o.WindowRules)-1]
			_ = setHyprSections(map[string]any{"windowRules": o.WindowRules})
			_ = applyLive()
			return fmt.Errorf("window rule rejected by Hyprland: %w", err)
		}
		return printJSON(map[string]any{
			"status": "added",
			"rule": wm.WindowRule{
				Class:  class,
				Action: action,
				Value:  value,
			},
		})
	})
}

func removeRule(index int) error {
	return withDesktopLock(func() error {
		o := wm.LoadStore(desktopStorePath())
		if index < 0 || index >= len(o.WindowRules) {
			return fmt.Errorf("rule index %d out of range (total %d)", index, len(o.WindowRules))
		}
		removed := o.WindowRules[index]
		o.WindowRules = append(o.WindowRules[:index], o.WindowRules[index+1:]...)
		if err := setHyprSections(map[string]any{"windowRules": o.WindowRules}); err != nil {
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
