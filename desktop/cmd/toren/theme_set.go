package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// `toren theme set NAME` switches the desktop to a named look: the
// wallpaper, light or dark, the colour scheme and, for some, a fixed accent.
// The desktop switcher (scripts/colors/switchwall.sh) applies it, and matugen
// recolours the shell, the terminal, btop and GTK apps from it in one step.
// The looks are listed in shell-desktop/toren/themes.json.

type namedTheme struct {
	Name      string `json:"name"`
	Title     string `json:"title"`
	Wallpaper string `json:"wallpaper"`
	Mode      string `json:"mode"`
	Scheme    string `json:"scheme"`
	Accent    string `json:"accent,omitempty"`
}

var (
	themeCatalog     = "/usr/share/toren/shell-desktop/toren/themes.json"
	themeWallpapers  = "/usr/share/toren/backgrounds"
	themeSwitcher    = "/usr/share/toren/shell-desktop/scripts/colors/switchwall.sh"
	themeDesktop     = currentDesktopShell
	themeRunSwitcher = func(args ...string) error {
		cmd := exec.Command(themeSwitcher, args...)
		cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
		return cmd.Run()
	}
	accentRe = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)
	schemes  = map[string]bool{
		"scheme-content": true, "scheme-expressive": true, "scheme-fidelity": true,
		"scheme-fruit-salad": true, "scheme-monochrome": true, "scheme-neutral": true,
		"scheme-rainbow": true, "scheme-tonal-spot": true,
	}
)

func loadNamedThemes(path string) ([]namedTheme, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc struct {
		Themes []namedTheme `json:"themes"`
	}
	if err := json.Unmarshal(b, &doc); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return doc.Themes, nil
}

// validateNamedTheme reports what is wrong with a catalogue entry, or nil.
func validateNamedTheme(t namedTheme) error {
	switch {
	case !regexp.MustCompile(`^[a-z0-9-]+$`).MatchString(t.Name):
		return fmt.Errorf("name %q must be lowercase letters, digits and dashes", t.Name)
	case t.Mode != "dark" && t.Mode != "light":
		return fmt.Errorf("%s: mode %q is not dark or light", t.Name, t.Mode)
	case !schemes[t.Scheme]:
		return fmt.Errorf("%s: scheme %q is not one matugen knows", t.Name, t.Scheme)
	case t.Accent != "" && !accentRe.MatchString(t.Accent):
		return fmt.Errorf("%s: accent %q is not #rrggbb", t.Name, t.Accent)
	case t.Wallpaper == "" || strings.Contains(t.Wallpaper, "/"):
		return fmt.Errorf("%s: wallpaper %q must be a file name in the backgrounds folder", t.Name, t.Wallpaper)
	}
	return nil
}

func themeStateFile() string {
	base := os.Getenv("XDG_STATE_HOME")
	if base == "" {
		base = filepath.Join(homeDir(), ".local", "state")
	}
	return filepath.Join(base, "toren", "theme")
}

func currentDesktopShell() string {
	out, err := exec.Command("gsettings", "get", "org.toren.shell", "desktop-shell").Output()
	if err != nil {
		return "desktop"
	}
	if v := strings.Trim(strings.TrimSpace(string(out)), "'"); v != "" {
		return v
	}
	return "desktop"
}

// themeSet is `toren theme set [NAME]`.
func themeSet(args []string) int {
	themes, err := loadNamedThemes(themeCatalog)
	if err != nil {
		fmt.Fprintln(os.Stderr, "toren theme set:", err)
		return exitFail
	}
	if len(args) == 0 {
		current := ""
		if b, err := os.ReadFile(themeStateFile()); err == nil {
			current = strings.TrimSpace(string(b))
		}
		fmt.Println("Looks for the desktop (toren theme set NAME):")
		for _, t := range themes {
			mark := " "
			if t.Name == current {
				mark = "*"
			}
			fmt.Printf("%s %-11s %-5s %s\n", mark, t.Name, t.Mode, t.Wallpaper)
		}
		return exitOK
	}
	if len(args) > 1 {
		fmt.Fprintln(os.Stderr, "usage: toren theme set [NAME]")
		return exitUsage
	}
	var chosen *namedTheme
	for i := range themes {
		if themes[i].Name == args[0] {
			chosen = &themes[i]
		}
	}
	if chosen == nil {
		var names []string
		for _, t := range themes {
			names = append(names, t.Name)
		}
		fmt.Fprintf(os.Stderr, "toren theme set: no look %q; one of: %s\n", args[0], strings.Join(names, ", "))
		return exitUsage
	}
	if err := validateNamedTheme(*chosen); err != nil {
		fmt.Fprintln(os.Stderr, "toren theme set:", err)
		return exitFail
	}
	if shell := themeDesktop(); shell != "desktop" {
		fmt.Fprintf(os.Stderr, "toren theme set: looks are for the desktop shell; this machine runs %q (toren theme generate IMAGE colours ASURA's shell)\n", shell)
		return exitFail
	}
	wall := filepath.Join(themeWallpapers, chosen.Wallpaper)
	if _, err := os.Stat(wall); err != nil {
		fmt.Fprintf(os.Stderr, "toren theme set: the wallpaper %s is not installed\n", wall)
		return exitFail
	}
	if err := themeRunSwitcher("--image", wall, "--mode", chosen.Mode, "--type", chosen.Scheme); err != nil {
		fmt.Fprintln(os.Stderr, "toren theme set: applying the wallpaper and colours:", err)
		return exitFail
	}
	// A fixed accent re-colours from it, keeping the wallpaper just set.
	if chosen.Accent != "" {
		if err := themeRunSwitcher("--noswitch", "--color", chosen.Accent, "--mode", chosen.Mode, "--type", chosen.Scheme); err != nil {
			fmt.Fprintln(os.Stderr, "toren theme set: applying the accent colour:", err)
			return exitFail
		}
	}
	_ = os.MkdirAll(filepath.Dir(themeStateFile()), 0o755)
	_ = os.WriteFile(themeStateFile(), []byte(chosen.Name+"\n"), 0o644)
	fmt.Printf("look: %s (%s, %s). The desktop, terminal, btop and GTK apps follow; open apps may need a restart.\n", chosen.Title, chosen.Mode, chosen.Wallpaper)
	runUserHooks("post-theme", map[string]string{
		"ASURA_THEME_NAME":      chosen.Name,
		"ASURA_THEME_WALLPAPER": chosen.Wallpaper,
		"ASURA_THEME_MODE":      chosen.Mode,
		"ASURA_THEME_SCHEME":    chosen.Scheme,
		"ASURA_THEME_ACCENT":    chosen.Accent,
	})
	return exitOK
}
