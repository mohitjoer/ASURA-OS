package main

import (
	"fmt"
	"regexp"
	"strings"

	"toren/desktop/cmd/toren-hubd/internal/wm"
)

var validEnvKeyRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

var riskyEnvVars = map[string]string{
	"LD_PRELOAD":      "Modifying LD_PRELOAD injects libraries into every process and can break login or system utilities.",
	"LD_LIBRARY_PATH": "Modifying LD_LIBRARY_PATH can cause ABI incompatibilities and crash graphical sessions.",
	"PATH":            "Altering PATH can hide core system binaries and prevent the desktop shell from launching.",
	"PYTHONPATH":      "Modifying PYTHONPATH can break system python scripts including theme tools.",
	"SHELL":           "Setting an invalid SHELL can prevent terminal emulators and subshells from opening.",
	"SUDO_COMMAND":    "Overriding sudo environment variables can alter privilege escalation behavior.",
	"IFS":             "Changing Internal Field Separator breaks standard shell script parsing.",
}

type envInfo struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	Risky   bool   `json:"risky"`
	Warning string `json:"warning,omitempty"`
}

func checkRiskyEnv(key string) (bool, string) {
	upper := strings.ToUpper(strings.TrimSpace(key))
	if warn, ok := riskyEnvVars[upper]; ok {
		return true, warn
	}
	return false, ""
}

func runEnv(args []string) error {
	if len(args) == 0 || args[0] == "list" {
		return listEnv()
	}
	switch args[0] {
	case "check":
		if len(args) < 2 {
			return fmt.Errorf("env check needs <key>")
		}
		risky, warn := checkRiskyEnv(args[1])
		return printJSON(map[string]any{
			"key":     args[1],
			"risky":   risky,
			"warning": warn,
		})
	case "set":
		if len(args) < 3 {
			return fmt.Errorf("env set needs <key> <value>")
		}
		return setEnvVar(args[1], strings.Join(args[2:], " "))
	case "remove", "rm":
		if len(args) < 2 {
			return fmt.Errorf("env remove needs <key>")
		}
		return removeEnvVar(args[1])
	default:
		return fmt.Errorf("unknown env subcommand %q", args[0])
	}
}

func listEnv() error {
	o := wm.LoadStore(desktopStorePath())
	var out []envInfo
	for _, e := range o.Env {
		risky, warn := checkRiskyEnv(e.Key)
		out = append(out, envInfo{
			Key:     e.Key,
			Value:   e.Value,
			Risky:   risky,
			Warning: warn,
		})
	}
	if out == nil {
		out = []envInfo{}
	}
	return printJSON(out)
}

func setEnvVar(key, val string) error {
	key = strings.TrimSpace(key)
	if !validEnvKeyRe.MatchString(key) {
		return fmt.Errorf("invalid environment variable name %q (must be alphanumeric/underscore starting with letter/underscore)", key)
	}
	return withDesktopLock(func() error {
		o := wm.LoadStore(desktopStorePath())
		found := false
		for i, e := range o.Env {
			if e.Key == key {
				o.Env[i].Value = val
				found = true
				break
			}
		}
		if !found {
			o.Env = append(o.Env, wm.EnvVar{Key: key, Value: val})
		}
		if err := setHyprSections(map[string]any{"env": o.Env}); err != nil {
			return err
		}
		if err := applyLive(); err != nil {
			return err
		}
		risky, warn := checkRiskyEnv(key)
		return printJSON(map[string]any{
			"status":  "saved",
			"key":     key,
			"value":   val,
			"risky":   risky,
			"warning": warn,
		})
	})
}

func removeEnvVar(key string) error {
	key = strings.TrimSpace(key)
	return withDesktopLock(func() error {
		o := wm.LoadStore(desktopStorePath())
		idx := -1
		for i, e := range o.Env {
			if e.Key == key {
				idx = i
				break
			}
		}
		if idx == -1 {
			return fmt.Errorf("environment variable %q not found", key)
		}
		o.Env = append(o.Env[:idx], o.Env[idx+1:]...)
		if err := setHyprSections(map[string]any{"env": o.Env}); err != nil {
			return err
		}
		if err := applyLive(); err != nil {
			return err
		}
		return printJSON(map[string]any{
			"status": "removed",
			"key":    key,
		})
	})
}
