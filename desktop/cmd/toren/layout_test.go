package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"toren/desktop/internal/runner/runnertest"
)

func TestRenderLayoutLua(t *testing.T) {
	st := layoutState{Workspaces: map[string]string{"10": "master", "2": "scroll", "5": "float", "1": "tile"}}
	got := renderLayoutLua(st)
	for _, want := range []string{
		`hl.workspace_rule({ workspace = "1", layout = "dwindle" })`,
		`hl.workspace_rule({ workspace = "2", layout = "scrolling" })`,
		`hl.window_rule({ name = "toren-layout-float-5", match = { workspace = "5" }, float = true })`,
		`hl.workspace_rule({ workspace = "10", layout = "master" })`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("layouts.lua lacks %s:\n%s", want, got)
		}
	}
	if strings.Index(got, `workspace = "2"`) > strings.Index(got, `workspace = "10"`) {
		t.Errorf("workspaces are not in numeric order:\n%s", got)
	}
}

func TestNextLayoutCycles(t *testing.T) {
	if got := nextLayout("tile", 1); got != "scroll" {
		t.Errorf("after tile: %s", got)
	}
	if got := nextLayout("float", 1); got != "tile" {
		t.Errorf("after float: %s", got)
	}
	if got := nextLayout("tile", -1); got != "float" {
		t.Errorf("before tile: %s", got)
	}
}

func TestLayoutSetFloatAndBack(t *testing.T) {
	cfg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", cfg)
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())

	clients := `[{"address":"0xa","floating":false,"workspace":{"id":3}},` +
		`{"address":"0xb","floating":true,"workspace":{"id":3}},` +
		`{"address":"0xc","floating":false,"workspace":{"id":4}}]`
	fake := runnertest.New().
		Out("hyprctl -j activeworkspace", `{"id":3,"name":"3"}`).
		Out("hyprctl reload", "ok").
		Out("hyprctl -r eval return true", "ok").
		Out("hyprctl -j clients", clients).
		Out(`hyprctl dispatch hl.dsp.window.float({ action = "enable", window = "address:0xa" })`, "ok").
		Out(`notify-send -a Toren -r 47213 -t 1500 -i view-grid-symbolic Workspace 3: Floating layout`, "")

	var out, errOut bytes.Buffer
	if code := layoutMain([]string{"set", "float"}, &out, &errOut, fake.Run); code != exitOK {
		t.Fatalf("set float = %d: %s", code, errOut.String())
	}
	// only the tiled window on workspace 3 was floated
	if !fake.Ran(`hyprctl dispatch hl.dsp.window.float({ action = "enable", window = "address:0xa" })`) {
		t.Error("the tiled window was not floated")
	}
	for _, c := range fake.Calls() {
		if strings.Contains(c, "0xb") || strings.Contains(c, "0xc") {
			t.Errorf("touched a window it should not: %s", c)
		}
	}
	lua, _ := os.ReadFile(layoutLuaPath())
	if !strings.Contains(string(lua), `"toren-layout-float-3"`) {
		t.Errorf("layouts.lua has no float rule for workspace 3:\n%s", lua)
	}

	// back to tiling: only the window toren floated is tiled again
	clientsFloated := `[{"address":"0xa","floating":true,"workspace":{"id":3}},` +
		`{"address":"0xb","floating":true,"workspace":{"id":3}}]`
	fake = runnertest.New().
		Out("hyprctl -j activeworkspace", `{"id":3,"name":"3"}`).
		Out("hyprctl reload", "ok").
		Out("hyprctl -r eval return true", "ok").
		Out("hyprctl -j clients", clientsFloated).
		Out(`hyprctl dispatch hl.dsp.window.float({ action = "disable", window = "address:0xa" })`, "ok").
		Out(`notify-send -a Toren -r 47213 -t 1500 -i view-grid-symbolic Workspace 3: Tiling layout`, "")
	out.Reset()
	if code := layoutMain([]string{"set", "tile"}, &out, &errOut, fake.Run); code != exitOK {
		t.Fatalf("set tile = %d: %s", code, errOut.String())
	}
	if !fake.Ran(`hyprctl dispatch hl.dsp.window.float({ action = "disable", window = "address:0xa" })`) {
		t.Error("the window toren floated was not tiled again")
	}
	if fake.Ran(`hyprctl dispatch hl.dsp.window.float({ action = "disable", window = "address:0xb" })`) {
		t.Error("a window the app floated itself was tiled")
	}
	if st := loadLayoutState(layoutStatePath()); st.Workspaces["3"] != "tile" {
		t.Errorf("state = %v", st.Workspaces)
	}
}

func TestLayoutRejectsUnknown(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	fake := runnertest.New().Out("hyprctl -j activeworkspace", `{"id":1,"name":"1"}`)
	var out, errOut bytes.Buffer
	if code := layoutMain([]string{"set", "spiral"}, &out, &errOut, fake.Run); code != exitUsage {
		t.Errorf("set spiral = %d, want %d", code, exitUsage)
	}
}
