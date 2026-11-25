package main

import (
	"fmt"
	"os/exec"
	"strings"
)

type snapshotEntry struct {
	ID     string `json:"id"`
	Date   string `json:"date"`
	Reason string `json:"reason"`
	Path   string `json:"path"`
}

func runUpdates(args []string) error {
	if len(args) == 0 || args[0] == "status" {
		return updatesStatus()
	}
	switch args[0] {
	case "check":
		return updatesCheck()
	case "channel":
		if len(args) < 2 {
			return getChannel()
		}
		if args[1] == "get" {
			return getChannel()
		}
		if args[1] == "set" {
			if len(args) < 3 {
				return fmt.Errorf("updates channel set needs <stable|unstable>")
			}
			return setChannel(args[2])
		}
		return fmt.Errorf("updates channel needs get|set")
	case "snapshots":
		return listSnapshots()
	default:
		return fmt.Errorf("unknown updates subcommand %q", args[0])
	}
}

func updatesStatus() error {
	channel := "unstable"
	out, err := exec.Command("toren", "channel", "get").Output()
	if err == nil {
		channel = strings.TrimSpace(string(out))
	}
	return printJSON(map[string]any{
		"channel": channel,
	})
}

func updatesCheck() error {
	cmd := exec.Command("toren", "update", "--check")
	out, _ := cmd.CombinedOutput()
	return printJSON(map[string]any{
		"output": strings.TrimSpace(string(out)),
	})
}

func getChannel() error {
	out, err := exec.Command("toren", "channel", "get").Output()
	if err != nil {
		return fmt.Errorf("toren channel get failed: %w", err)
	}
	return printJSON(map[string]string{
		"channel": strings.TrimSpace(string(out)),
	})
}

func setChannel(ch string) error {
	cmd := exec.Command("toren", "channel", "set", ch)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("toren channel set %s failed: %v: %s", ch, err, strings.TrimSpace(string(out)))
	}
	return printJSON(map[string]string{
		"status":  "set",
		"channel": ch,
	})
}

func listSnapshots() error {
	cmd := exec.Command("toren", "snapshot", "list")
	out, err := cmd.Output()
	var entries []snapshotEntry
	if err == nil {
		lines := strings.Split(string(out), "\n")
		for _, l := range lines {
			l = strings.TrimSpace(l)
			if l == "" || strings.HasPrefix(l, "ID") || strings.HasPrefix(l, "--") {
				continue
			}
			fields := strings.Fields(l)
			if len(fields) >= 3 {
				entries = append(entries, snapshotEntry{
					ID:     fields[0],
					Date:   fields[1],
					Reason: fields[2],
					Path:   strings.Join(fields[3:], " "),
				})
			}
		}
	}
	if entries == nil {
		entries = []snapshotEntry{}
	}
	return printJSON(entries)
}
