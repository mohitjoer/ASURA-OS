package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestShippedLooksAreValid checks every look in the catalogue ASURA ships and
// that its wallpaper is one ASURA ships.
func TestShippedLooksAreValid(t *testing.T) {
	themes, err := loadNamedThemes("../../../shell-desktop/toren/themes.json")
	if err != nil {
		t.Fatal(err)
	}
	if len(themes) < 8 {
		t.Fatalf("only %d looks", len(themes))
	}
	seen := map[string]bool{}
	for _, th := range themes {
		if err := validateNamedTheme(th); err != nil {
			t.Error(err)
		}
		if seen[th.Name] {
			t.Errorf("%s is listed twice", th.Name)
		}
		seen[th.Name] = true
		if _, err := os.Stat(filepath.Join("../../../public/wallpapers", th.Wallpaper)); err != nil {
			t.Errorf("%s: wallpaper %s is not in public/wallpapers", th.Name, th.Wallpaper)
		}
	}
}

func themeSandbox(t *testing.T, desktop string) *[]string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", filepath.Join(dir, "state"))
	walls := filepath.Join(dir, "backgrounds")
	if err := os.MkdirAll(walls, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, w := range []string{"sakura-gate.jpg", "forest.png"} {
		if err := os.WriteFile(filepath.Join(walls, w), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	oldCat, oldWall, oldDesk, oldRun := themeCatalog, themeWallpapers, themeDesktop, themeRunSwitcher
	t.Cleanup(func() {
		themeCatalog, themeWallpapers, themeDesktop, themeRunSwitcher = oldCat, oldWall, oldDesk, oldRun
	})
	themeCatalog, themeWallpapers = "../../../shell-desktop/toren/themes.json", walls
	themeDesktop = func() string { return desktop }
	var calls []string
	themeRunSwitcher = func(args ...string) error { calls = append(calls, strings.Join(args, " ")); return nil }
	return &calls
}

func TestThemeSetRunsTheSwitcherAndRemembersTheLook(t *testing.T) {
	calls := themeSandbox(t, "desktop")
	if code := themeSet([]string{"forest"}); code != exitOK {
		t.Fatalf("code %d", code)
	}
	if len(*calls) != 1 || !strings.Contains((*calls)[0], "forest.png --mode dark --type scheme-content") {
		t.Fatalf("calls %v", *calls)
	}
	*calls = nil
	if code := themeSet([]string{"sakura"}); code != exitOK || len(*calls) != 2 || !strings.HasPrefix((*calls)[1], "--noswitch --color #f2a7c3") {
		t.Fatalf("accent look: %d %v", code, *calls)
	}
	if b, _ := os.ReadFile(themeStateFile()); strings.TrimSpace(string(b)) != "sakura" {
		t.Fatalf("remembered %q", b)
	}
}

func TestThemeSetRefusesOnASURAsOwnShellAndUnknownLooks(t *testing.T) {
	calls := themeSandbox(t, "toren")
	if code := themeSet([]string{"forest"}); code != exitFail || len(*calls) != 0 {
		t.Fatalf("on toren's shell: %d %v", code, *calls)
	}
	themeSandbox(t, "desktop")
	if code := themeSet([]string{"neon"}); code != exitUsage {
		t.Fatalf("unknown: %d", code)
	}
	if code := themeSet([]string{"kaiju"}); code != exitFail {
		t.Fatalf("a missing wallpaper must fail: %d", code)
	}
}
