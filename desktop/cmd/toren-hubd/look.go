package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type namedLook struct {
	Name      string `json:"name"`
	Title     string `json:"title"`
	Wallpaper string `json:"wallpaper"`
	Mode      string `json:"mode"`
	Scheme    string `json:"scheme"`
	Accent    string `json:"accent,omitempty"`
}

func runLook(args []string) error {
	if len(args) == 0 || args[0] == "list" {
		return listLooks()
	}
	switch args[0] {
	case "set":
		if len(args) < 2 {
			return fmt.Errorf("look set needs <name>")
		}
		return setLook(args[1])
	case "mode":
		if len(args) < 2 {
			return getMode()
		}
		return setMode(args[1])
	case "wallpapers":
		return listWallpapers()
	default:
		return fmt.Errorf("unknown look subcommand %q", args[0])
	}
}

func listLooks() error {
	candidates := []string{
		"/usr/share/asura/shell-desktop/toren/themes.json",
		filepath.Join(os.Getenv("HOME"), "code/asura/Toren-Linux/shell-desktop/toren/themes.json"),
	}
	var catPath string
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			catPath = c
			break
		}
	}
	var looks []namedLook
	if catPath != "" {
		if b, err := os.ReadFile(catPath); err == nil {
			var doc struct {
				Themes []namedLook `json:"themes"`
			}
			_ = json.Unmarshal(b, &doc)
			looks = doc.Themes
		}
	}
	if looks == nil {
		looks = []namedLook{}
	}
	current := currentThemeName()
	mode := currentMode()
	return printJSON(map[string]any{
		"looks":   looks,
		"current": current,
		"mode":    mode,
	})
}

func currentThemeName() string {
	stateFile := filepath.Join(os.Getenv("HOME"), ".local", "state", "asura", "theme")
	if b, err := os.ReadFile(stateFile); err == nil {
		return strings.TrimSpace(string(b))
	}
	return "default"
}

func currentMode() string {
	stateFile := filepath.Join(os.Getenv("HOME"), ".local", "state", "asura", "theme-mode")
	if b, err := os.ReadFile(stateFile); err == nil {
		return strings.TrimSpace(string(b))
	}
	return "dark"
}

func setLook(name string) error {
	cmd := exec.Command("asura", "theme", "set", name)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("asura theme set %s failed: %v: %s", name, err, strings.TrimSpace(string(out)))
	}
	return printJSON(map[string]any{
		"status": "applied",
		"theme":  name,
	})
}

func getMode() error {
	return printJSON(map[string]string{
		"mode": currentMode(),
	})
}

func setMode(mode string) error {
	mode = strings.ToLower(strings.TrimSpace(mode))
	if mode != "dark" && mode != "light" {
		return fmt.Errorf("mode must be 'dark' or 'light'")
	}
	stateDir := filepath.Join(os.Getenv("HOME"), ".local", "state", "asura")
	_ = os.MkdirAll(stateDir, 0o755)
	if err := os.WriteFile(filepath.Join(stateDir, "theme-mode"), []byte(mode+"\n"), 0o644); err != nil {
		return err
	}
	return printJSON(map[string]any{
		"status": "applied",
		"mode":   mode,
	})
}

func listWallpapers() error {
	var walls []string
	searchDirs := []string{
		"/usr/share/asura/backgrounds",
		filepath.Join(os.Getenv("HOME"), "Pictures", "Wallpapers"),
	}
	for _, dir := range searchDirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			ext := strings.ToLower(filepath.Ext(e.Name()))
			if ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".webp" {
				walls = append(walls, filepath.Join(dir, e.Name()))
			}
		}
	}
	if walls == nil {
		walls = []string{}
	}
	return printJSON(walls)
}
