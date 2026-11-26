package wm

import (
	"os"
	"strings"
	"testing"
)

// A dotted option name nests in Lua: `col.active_border = ...` inside a table is
// a syntax error, and it stopped the whole of settings.lua from loading.
func TestSettingsLuaNestsDottedOptions(t *testing.T) {
	o := defaultOverrides()
	o.Appearance.ActiveBorder = "#ff0000"
	got := genSettingsLua(o, false)
	if strings.Contains(got, "col.active_border =") || strings.Contains(got, "col.inactive_border =") {
		t.Fatalf("settings.lua has a dotted key in a table:\n%s", got)
	}
	if !strings.Contains(got, "col = {") || !strings.Contains(got, `active_border = "rgb(ff0000)"`) {
		t.Errorf("border colour not nested under col:\n%s", got)
	}
}

// What settings.conf carries beyond options reaches settings.lua too.
func TestSettingsLuaCarriesRulesBindsAnimationsAndApps(t *testing.T) {
	o := defaultOverrides()
	o.Apps = map[string]string{"browser": "firefox"}
	o.WindowRules = []WindowRule{{Class: "mpv", Action: "size", Value: "800x600"}, {Class: "pavucontrol", Action: "float"}}
	o.LayerRules = []LayerRule{{Namespace: "x", Action: "abovelock"}}
	o.AppOverrides = []AppOverride{{Class: "kitty", Opacity: 0.9, Rounding: -1, BorderSize: -1, Blur: "off"}}
	o.Keybinds = []Keybind{{Keys: "SUPER + KP_1", Action: "exec", Value: "foot"}, {Keys: "SUPER + X", Action: "close", Release: true}}
	o.Unbinds = []string{"SUPER + Q"}
	o.Anim.Items = []AnimItem{{Leaf: "windowsIn", Enabled: true, Speed: 3, Bezier: "default", Style: "popin 80%"}, {Leaf: "fade"}}
	o.Autostart = []Autostart{{Command: "nm-applet"}}
	got := genSettingsLua(o, false)
	for _, want := range []string{
		`hl.env("BROWSER", "firefox")`,
		`hl.window_rule({ name = "toren-user-1", size = "800 600", match = { class = "mpv" } })`,
		`hl.window_rule({ name = "toren-user-2", float = true, match = { class = "pavucontrol" } })`,
		`hl.layer_rule({ name = "toren-layer-1", above_lock = 1, match = { namespace = "x" } })`,
		`hl.window_rule({ name = "toren-app-1", opacity = "0.9", no_blur = true, match = { class = "kitty" } })`,
		`pcall(hl.unbind, "SUPER + Q")`,
		`pcall(hl.bind, "SUPER + KP_1", hl.dsp.exec_cmd("foot"))`,
		`pcall(hl.bind, "SUPER + KP_End", hl.dsp.exec_cmd("foot"))`,
		`pcall(hl.bind, "SUPER + X", hl.dsp.window.close(), { release = true })`,
		`pcall(hl.animation, { leaf = "windowsIn", enabled = true, speed = 3, bezier = "default", style = "popin 80%" })`,
		`pcall(hl.animation, { leaf = "fade", enabled = false })`,
		`hl.exec_cmd("nm-applet")`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("settings.lua lacks %s\n%s", want, got)
		}
	}
}

// Hyprland 0.57 has no gestures:workspace_swipe options; the swipe is a gesture
// hyprland.lua adds unless settings.lua says otherwise.
func TestSettingsLuaWorkspaceSwipe(t *testing.T) {
	o := defaultOverrides()
	o.Input.WorkspaceSwipe = false
	o.Input.SwipeFingers = 4
	got := genSettingsLua(o, false)
	if strings.Contains(got, "workspace_swipe =") || strings.Contains(got, "workspace_swipe_fingers") {
		t.Errorf("settings.lua writes a removed option:\n%s", got)
	}
	for _, want := range []string{"TOREN_WORKSPACE_SWIPE = false", "TOREN_WORKSPACE_SWIPE_FINGERS = 4"} {
		if !strings.Contains(got, want) {
			t.Errorf("settings.lua lacks %s\n%s", want, got)
		}
	}
}

// The live preview goes to `hyprctl eval` as Lua: there is no keyword command.
func TestKeywordsToLua(t *testing.T) {
	got := keywordsToLua([]string{
		"keyword general:gaps_in 9",
		"keyword general:col.active_border rgb(ff0000) rgb(3a3a3a) 45deg",
		"keyword animation windowsIn,1,2,default,popin 78%",
		"keyword gestures:workspace_swipe true",
		"dispatch nothing",
	})
	for _, want := range []string{
		"hl.config({ general = { gaps_in = 9 } })",
		`hl.config({ general = { col = { active_border = { colors = { "rgb(ff0000)", "rgb(3a3a3a)" }, angle = 45 } } } })`,
		`pcall(hl.animation, { leaf = "windowsIn", enabled = true, speed = 2, bezier = "default", style = "popin 78%" })`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("preview lacks %s\n%s", want, got)
		}
	}
	if strings.Contains(got, "workspace_swipe") || strings.Contains(got, "nothing") {
		t.Errorf("preview carries what it should drop:\n%s", got)
	}
}

func TestMonitorSpecLua(t *testing.T) {
	if got := monitorSpecLua("DP-1, disable"); got != "hl.monitor({ output = \"DP-1\", disabled = true })\n" {
		t.Errorf("disabled output: %q", got)
	}
	got := monitorSpecLua("eDP-1, 1920x1080@60, 0x0, 1.25, transform, 1, vrr, 1, mirror, DP-2")
	for _, want := range []string{`output = "eDP-1"`, `scale = "1.25"`, "transform = 1,", "vrr = 1,", `mirror = "DP-2"`} {
		if !strings.Contains(got, want) {
			t.Errorf("monitor lacks %s:\n%s", want, got)
		}
	}
}

// A rebind takes the shipped action from binds.lua as written, options and
// multi-line functions included.
func TestParseLuaBinds(t *testing.T) {
	p := t.TempDir() + "/binds.lua"
	src := `-- comment, hl.bind("NOT + THIS", x)
hl.bind("SUPER + Q", hl.dsp.window.kill())
hl.bind("ALT + Tab", hl.dsp.exec_cmd("toren ipc call toren closeSwitcher, now"), { release = true })
hl.bind("SUPER + A", function()
    hl.dispatch(hl.dsp.window.float())
    hl.dispatch(hl.dsp.window.resize({ x = 1000, y = 660 }))
end)
`
	if err := os.WriteFile(p, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	got := parseLuaBinds(p)
	if len(got) != 3 {
		t.Fatalf("parsed %d chords, want 3: %v", len(got), got)
	}
	if b := got[chordKey("SUPER + Q")]; len(b) != 1 || b[0].expr != "hl.dsp.window.kill()" {
		t.Errorf("SUPER + Q = %+v", b)
	}
	if b := got[chordKey("ALT + Tab")]; len(b) != 1 || b[0].opts != "{ release = true }" || !strings.Contains(b[0].expr, "closeSwitcher, now") {
		t.Errorf("ALT + Tab = %+v", b)
	}
	if b := got[chordKey("SUPER + A")]; len(b) != 1 || !strings.HasPrefix(b[0].expr, "function()") || !strings.HasSuffix(b[0].expr, "end") {
		t.Errorf("SUPER + A = %+v", b)
	}
	var sb strings.Builder
	writeLuaBind(&sb, "SUPER + CTRL + Tab", got[chordKey("ALT + Tab")][0])
	if want := `hl.bind("SUPER + CTRL + Tab", hl.dsp.exec_cmd("toren ipc call toren closeSwitcher, now"), { release = true })`; strings.TrimSpace(sb.String()) != want {
		t.Errorf("rebind = %s", sb.String())
	}
}

// Without binds.lua, a hyprlang bind becomes a real dispatcher, never a call
// that runs once at load (hl.dispatch), and keeps its flags.
func TestConfBindLua(t *testing.T) {
	got := confBindLua(confBind{flags: "r", rest: "togglefloating,"})
	if strings.Contains(got.expr, "hl.dispatch(") || got.opts != "{ release = true }" {
		t.Errorf("confBindLua = %+v", got)
	}
}
