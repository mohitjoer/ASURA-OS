package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

var flatpakAppIDRe = regexp.MustCompile(`^[A-Za-z0-9_-]+(\.[A-Za-z0-9_-]+){2,}$`)

var seedAppsFlatpak = []struct{ id, title, summary string }{
	{"com.discordapp.Discord", "Discord (Flatpak)", "Voice, video, and text chat for communities"},
	{"com.spotify.Client", "Spotify (Flatpak)", "Music streaming desktop client"},
	{"md.obsidian.Obsidian", "Obsidian (Flatpak)", "Markdown knowledge base and notes"},
	{"com.bitwarden.desktop", "Bitwarden (Flatpak)", "Open-source password manager"},
	{"org.signal.Signal", "Signal (Flatpak)", "Private encrypted messenger"},
	{"com.visualstudio.code", "VS Code (Flatpak)", "Visual Studio Code editor"},
}

func isFlatpakID(id string) bool {
	return flatpakAppIDRe.MatchString(id)
}

func flatpakInstalled(ctx context.Context) map[string]string {
	res := map[string]string{}
	if !have("flatpak") {
		return res
	}
	out, err := exec.CommandContext(ctx, "flatpak", "list", "--app", "--columns=application,version").Output()
	if err != nil {
		return res
	}
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		fields := strings.Split(strings.TrimSpace(scanner.Text()), "\t")
		if len(fields) == 0 || fields[0] == "" {
			continue
		}
		ver := "installed"
		if len(fields) > 1 && strings.TrimSpace(fields[1]) != "" {
			ver = strings.TrimSpace(fields[1])
		}
		res[strings.TrimSpace(fields[0])] = ver
	}
	return res
}

func flatpakSearch(ctx context.Context, query string, installed map[string]string) []Item {
	if !have("flatpak") {
		return nil
	}
	out, err := exec.CommandContext(ctx, "flatpak", "search", "--columns=application,name,description,version", query).Output()
	if err != nil {
		return nil
	}
	var items []Item
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), "\t")
		if len(fields) < 2 {
			continue
		}
		appID := strings.TrimSpace(fields[0])
		if !isFlatpakID(appID) {
			continue
		}
		name := strings.TrimSpace(fields[1])
		desc := ""
		ver := ""
		if len(fields) > 2 {
			desc = strings.TrimSpace(fields[2])
		}
		if len(fields) > 3 {
			ver = strings.TrimSpace(fields[3])
		}
		instVer := installed[appID]
		items = append(items, Item{
			ID:               appID,
			Category:         "apps",
			Name:             name + " (Flatpak)",
			Summary:          desc,
			Version:          ver,
			Installed:        instVer != "",
			InstalledVersion: instVer,
		})
	}
	return items
}

func flatpakInstall(ctx context.Context, appID string) error {
	if !have("flatpak") {
		return fmt.Errorf("flatpak is not installed")
	}
	emitStage("AUTHORIZING", 10)
	_ = exec.CommandContext(ctx, "flatpak", "remote-add", "--user", "--if-not-exists", "flathub", "https://dl.flathub.org/repo/flathub.flatpakrepo").Run()
	emitStage("INSTALLING", 40)
	out, err := exec.CommandContext(ctx, "flatpak", "install", "--user", "-y", "--noninteractive", "flathub", appID).CombinedOutput()
	if err != nil {
		if msg := strings.TrimSpace(string(out)); msg != "" {
			return fmt.Errorf("flatpak install failed: %s", msg)
		}
		return fmt.Errorf("flatpak install failed: %w", err)
	}
	emitStage("INSTALLING", 100)
	return nil
}

func flatpakRemove(ctx context.Context, appID string) error {
	if !have("flatpak") {
		return fmt.Errorf("flatpak is not installed")
	}
	emitStage("REMOVING", 40)
	out, err := exec.CommandContext(ctx, "flatpak", "uninstall", "--user", "-y", "--noninteractive", appID).CombinedOutput()
	if err != nil {
		if msg := strings.TrimSpace(string(out)); msg != "" {
			return fmt.Errorf("flatpak remove failed: %s", msg)
		}
		return fmt.Errorf("flatpak remove failed: %w", err)
	}
	emitStage("REMOVING", 100)
	return nil
}
