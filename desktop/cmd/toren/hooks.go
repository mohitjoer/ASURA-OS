package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const hookUsage = `Usage: toren hook run EVENT [KEY=VAL ...] | toren hook list [EVENT]

Runs user hook scripts from ~/.config/toren/hooks/<event>/ in name order.

Supported Events:
  pre-update      runs before toren update begins updating packages
  post-update     runs after toren update completes successfully
  post-theme      runs after toren theme applies a new wallpaper or colours
  post-login      runs upon graphical desktop session login

Environment Variables passed to hooks:
  ASURA_EVENT     the event name being fired
  ASURA_HOOK_DIR  the directory holding the executed hooks
  ASURA_USER      the desktop user name

Hooks are executed with a 30s timeout and logged to ~/.local/state/toren/hooks.log.
A failing hook is reported and never aborts Toren's internal workflows.
`

func init() {
	register("hook", "run or list user extension hooks", torenHook)
	helpText["hook"] = hookUsage
}

var validHookEvents = map[string]bool{
	"pre-update":  true,
	"post-update": true,
	"post-theme":  true,
	"post-login":  true,
}

func torenHook(args []string) int {
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, hookUsage)
		return exitUsage
	}
	switch args[0] {
	case "list":
		event := ""
		if len(args) > 1 {
			event = args[1]
		}
		return hookList(event)
	case "run":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "toren hook run: missing event name")
			fmt.Fprint(os.Stderr, hookUsage)
			return exitUsage
		}
		event := args[1]
		extraEnv := map[string]string{}
		for _, kv := range args[2:] {
			parts := strings.SplitN(kv, "=", 2)
			if len(parts) == 2 {
				extraEnv[parts[0]] = parts[1]
			}
		}
		runUserHooks(event, extraEnv)
		return exitOK
	default:
		fmt.Fprintf(os.Stderr, "toren hook: unknown subcommand %q\n\n", args[0])
		fmt.Fprint(os.Stderr, hookUsage)
		return exitUsage
	}
}

func hooksBaseDir() string {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		base = filepath.Join(homeDir(), ".config")
	}
	return filepath.Join(base, "toren", "hooks")
}

func hooksLogPath() string {
	base := os.Getenv("XDG_STATE_HOME")
	if base == "" {
		base = filepath.Join(homeDir(), ".local", "state")
	}
	return filepath.Join(base, "toren", "hooks.log")
}

func hookList(filterEvent string) int {
	base := hooksBaseDir()
	events := []string{"pre-update", "post-update", "post-theme", "post-login"}
	if filterEvent != "" {
		if !validHookEvents[filterEvent] {
			fmt.Fprintf(os.Stderr, "toren hook list: unknown event %q\n", filterEvent)
			return exitFail
		}
		events = []string{filterEvent}
	}

	foundAny := false
	for _, ev := range events {
		dir := filepath.Join(base, ev)
		entries, err := os.ReadDir(dir)
		if err != nil || len(entries) == 0 {
			continue
		}
		foundAny = true
		fmt.Printf("Event %q (%s):\n", ev, dir)
		for _, e := range entries {
			info, err := e.Info()
			if err != nil {
				continue
			}
			execMark := ""
			if info.Mode()&0o111 != 0 {
				execMark = " [executable]"
			}
			fmt.Printf("  - %s%s\n", e.Name(), execMark)
		}
	}
	if !foundAny {
		fmt.Println("No user hooks installed in " + base)
	}
	return exitOK
}

func runUserHooks(event string, extraEnv map[string]string) {
	if !validHookEvents[event] {
		fmt.Fprintf(os.Stderr, "toren hook: unsupported event %q\n", event)
		return
	}

	dir := filepath.Join(hooksBaseDir(), event)
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) == 0 {
		return
	}

	var scripts []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		path := filepath.Join(dir, e.Name())
		info, err := e.Info()
		if err == nil && info.Mode()&0o111 != 0 {
			scripts = append(scripts, path)
		}
	}
	sort.Strings(scripts)

	if len(scripts) == 0 {
		return
	}

	user := os.Getenv("USER")
	if sudoUser := os.Getenv("SUDO_USER"); sudoUser != "" {
		user = sudoUser
	}

	logFile := hooksLogPath()
	_ = os.MkdirAll(filepath.Dir(logFile), 0o755)

	for _, script := range scripts {
		name := filepath.Base(script)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

		var cmd *exec.Cmd
		if os.Geteuid() == 0 && os.Getenv("SUDO_USER") != "" {
			cmd = exec.CommandContext(ctx, "su", "-c", script, os.Getenv("SUDO_USER"))
		} else {
			cmd = exec.CommandContext(ctx, script)
		}

		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env,
			"ASURA_EVENT="+event,
			"ASURA_HOOK_DIR="+dir,
			"ASURA_USER="+user,
		)
		for k, v := range extraEnv {
			cmd.Env = append(cmd.Env, k+"="+v)
		}

		out, err := cmd.CombinedOutput()
		cancel()

		ts := time.Now().Format("2006-01-02 15:04:05")
		statusStr := "SUCCESS"
		if err != nil {
			statusStr = fmt.Sprintf("FAILED (%v)", err)
			fmt.Fprintf(os.Stderr, "toren hook [%s]: %s failed: %v\n", event, name, err)
		}

		logLine := fmt.Sprintf("[%s] [%s] %s -> %s\n", ts, event, name, statusStr)
		if len(out) > 0 {
			logLine += fmt.Sprintf("  Output: %s\n", strings.TrimSpace(string(out)))
		}

		if f, openErr := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644); openErr == nil {
			_, _ = f.WriteString(logLine)
			f.Close()
		}
	}
}
