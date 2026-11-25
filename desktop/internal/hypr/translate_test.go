package hypr

import (
	"testing"
)

func TestDispatchToLuaMappings(t *testing.T) {
	tests := []struct {
		name     string
		disp     string
		args     []string
		expected string
	}{
		{
			name:     "workspace number",
			disp:     "workspace",
			args:     []string{"3"},
			expected: `hl.dsp.focus({ workspace = "3" })`,
		},
		{
			name:     "workspace relative forward",
			disp:     "workspace",
			args:     []string{"r+1"},
			expected: `hl.dsp.focus({ workspace = "r+1" })`,
		},
		{
			name:     "workspace relative back",
			disp:     "workspace",
			args:     []string{"r-1"},
			expected: `hl.dsp.focus({ workspace = "r-1" })`,
		},
		{
			name:     "workspace relative existing",
			disp:     "workspace",
			args:     []string{"e+1"},
			expected: `hl.dsp.focus({ workspace = "e+1" })`,
		},
		{
			name:     "workspace named",
			disp:     "workspace",
			args:     []string{"name:foo"},
			expected: `hl.dsp.focus({ workspace = "name:foo" })`,
		},
		{
			name:     "workspace special",
			disp:     "workspace",
			args:     []string{"special:x"},
			expected: `hl.dsp.focus({ workspace = "special:x" })`,
		},
		{
			name:     "movetoworkspace single arg",
			disp:     "movetoworkspace",
			args:     []string{"3"},
			expected: `hl.dsp.window.move({ workspace = "3" })`,
		},
		{
			name:     "movetoworkspace with window",
			disp:     "movetoworkspace",
			args:     []string{"3,address:0xdeadbeef"},
			expected: `hl.dsp.window.move({ workspace = "3", window = "address:0xdeadbeef" })`,
		},
		{
			name:     "movetoworkspacesilent without window",
			disp:     "movetoworkspacesilent",
			args:     []string{"5"},
			expected: `hl.dsp.window.move({ workspace = "5", follow = false })`,
		},
		{
			name:     "movetoworkspacesilent with window",
			disp:     "movetoworkspacesilent",
			args:     []string{"5,address:0xdeadbeef"},
			expected: `hl.dsp.window.move({ workspace = "5", follow = false, window = "address:0xdeadbeef" })`,
		},
		{
			name:     "focuswindow address",
			disp:     "focuswindow",
			args:     []string{"address:0xdeadbeef"},
			expected: `hl.dsp.focus({ window = "address:0xdeadbeef" })`,
		},
		{
			name:     "closewindow address",
			disp:     "closewindow",
			args:     []string{"address:0xdeadbeef"},
			expected: `hl.dsp.window.close({ window = "address:0xdeadbeef" })`,
		},
		{
			name:     "closewindow without args",
			disp:     "closewindow",
			args:     nil,
			expected: `hl.dsp.window.close()`,
		},
		{
			name:     "killactive",
			disp:     "killactive",
			args:     nil,
			expected: `hl.dsp.window.close()`,
		},
		{
			name:     "pin address",
			disp:     "pin",
			args:     []string{"address:0xdeadbeef"},
			expected: `hl.dsp.window.pin({ window = "address:0xdeadbeef" })`,
		},
		{
			name:     "togglefloating no args",
			disp:     "togglefloating",
			args:     nil,
			expected: `hl.dsp.window.float()`,
		},
		{
			name:     "togglefloating enable",
			disp:     "togglefloating",
			args:     []string{"enable"},
			expected: `hl.dsp.window.float({ action = "enable" })`,
		},
		{
			name:     "togglefloating address",
			disp:     "togglefloating",
			args:     []string{"address:0xdeadbeef"},
			expected: `hl.dsp.window.float({ window = "address:0xdeadbeef" })`,
		},
		{
			name:     "movewindowpixel exact with address",
			disp:     "movewindowpixel",
			args:     []string{"exact 100 200,address:0xdeadbeef"},
			expected: `hl.dsp.window.move({ x = 100, y = 200, window = "address:0xdeadbeef" })`,
		},
		{
			name:     "resizeactive exact",
			disp:     "resizeactive",
			args:     []string{"exact 1000 660"},
			expected: `hl.dsp.window.resize({ x = 1000, y = 660 })`,
		},
		{
			name:     "resizeactive relative",
			disp:     "resizeactive",
			args:     []string{"10 -10"},
			expected: `hl.dsp.window.resize({ x = 10, y = -10, relative = true })`,
		},
		{
			name:     "centerwindow",
			disp:     "centerwindow",
			args:     nil,
			expected: `hl.dsp.window.center()`,
		},
		{
			name:     "bringactivetotop",
			disp:     "bringactivetotop",
			args:     nil,
			expected: `hl.dsp.window.bring_to_top()`,
		},
		{
			name:     "cyclenext",
			disp:     "cyclenext",
			args:     nil,
			expected: `hl.dsp.window.cycle_next()`,
		},
		{
			name:     "cyclenext prev",
			disp:     "cyclenext",
			args:     []string{"prev"},
			expected: `hl.dsp.window.cycle_next({ next = false })`,
		},
		{
			name:     "exec command",
			disp:     "exec",
			args:     []string{"kitty"},
			expected: `hl.dsp.exec_cmd("kitty")`,
		},
		{
			name:     "global name",
			disp:     "global",
			args:     []string{"quickshell:toggle"},
			expected: `hl.dsp.global("quickshell:toggle")`,
		},
		{
			name:     "submap name",
			disp:     "submap",
			args:     []string{"resize"},
			expected: `hl.dsp.submap("resize")`,
		},
		{
			name:     "submap reset",
			disp:     "submap",
			args:     []string{"reset"},
			expected: `hl.dsp.submap("reset")`,
		},
		{
			name:     "dpms on",
			disp:     "dpms",
			args:     []string{"on"},
			expected: `hl.dsp.dpms({ action = "on" })`,
		},
		{
			name:     "dpms off",
			disp:     "dpms",
			args:     []string{"off"},
			expected: `hl.dsp.dpms({ action = "off" })`,
		},
		{
			name:     "dpms off with monitor",
			disp:     "dpms",
			args:     []string{"off", "eDP-2"},
			expected: `hl.dsp.dpms({ action = "off", monitor = "eDP-2" })`,
		},
		{
			name:     "fullscreen toggle",
			disp:     "fullscreen",
			args:     []string{"0"},
			expected: `hl.dsp.window.fullscreen()`,
		},
		{
			name:     "fullscreen maximized",
			disp:     "fullscreen",
			args:     []string{"1"},
			expected: `hl.dsp.window.fullscreen({ mode = "1" })`,
		},
		{
			name:     "exit",
			disp:     "exit",
			args:     nil,
			expected: `hl.dsp.exit()`,
		},
		{
			name:     "togglespecialworkspace with name",
			disp:     "togglespecialworkspace",
			args:     []string{"sharebar"},
			expected: `hl.dsp.workspace.toggle_special("sharebar")`,
		},
		{
			name:     "togglespecialworkspace default",
			disp:     "togglespecialworkspace",
			args:     nil,
			expected: `hl.dsp.workspace.toggle_special()`,
		},
		{
			name:     "moveworkspacetomonitor",
			disp:     "moveworkspacetomonitor",
			args:     []string{"9", "eDP-2"},
			expected: `hl.dsp.workspace.move({ workspace = "9", monitor = "eDP-2" })`,
		},
		{
			name:     "escaping quotes in titles",
			disp:     "focuswindow",
			args:     []string{`title:Alice's "Notes"`},
			expected: `hl.dsp.focus({ window = "title:Alice's \"Notes\"" })`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DispatchToLua(tt.disp, tt.args...)
			if got != tt.expected {
				t.Errorf("DispatchToLua(%q, %v)\n got:  %s\n want: %s", tt.disp, tt.args, got, tt.expected)
			}
		})
	}
}

func TestKeywordToLuaMappings(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		val      string
		expected string
	}{
		{
			name:     "misc:disable_autoreload true",
			key:      "misc:disable_autoreload",
			val:      "true",
			expected: `hl.config({ misc = { disable_autoreload = true } })`,
		},
		{
			name:     "misc:disable_autoreload false",
			key:      "misc:disable_autoreload",
			val:      "false",
			expected: `hl.config({ misc = { disable_autoreload = false } })`,
		},
		{
			name:     "general:gaps_in number",
			key:      "general:gaps_in",
			val:      "5",
			expected: `hl.config({ general = { gaps_in = 5 } })`,
		},
		{
			name:     "decoration:rounding number",
			key:      "decoration:rounding",
			val:      "10",
			expected: `hl.config({ decoration = { rounding = 10 } })`,
		},
		{
			name:     "input:follow_mouse number",
			key:      "input:follow_mouse",
			val:      "2",
			expected: `hl.config({ input = { follow_mouse = 2 } })`,
		},
		{
			name:     "decoration:screen_shader path",
			key:      "decoration:screen_shader",
			val:      "/path/to/shader.glsl",
			expected: `hl.config({ decoration = { screen_shader = "/path/to/shader.glsl" } })`,
		},
		{
			name:     "general:col.active_border with dot in key",
			key:      "general:col.active_border",
			val:      "rgb(f3701e)",
			expected: `hl.config({ general = { ["col.active_border"] = "rgb(f3701e)" } })`,
		},
		{
			name:     "workspace layout rule",
			key:      "workspace",
			val:      "9,layout:master",
			expected: `hl.workspace_rule({ workspace = "9", layout = "master" })`,
		},
		{
			name:     "monitor disable",
			key:      "monitor",
			val:      "DP-1,disable",
			expected: `hl.monitor({ output = "DP-1", disabled = true })`,
		},
		{
			name:     "monitor mode position scale",
			key:      "monitor",
			val:      "DP-1,2560x1440@143.91,1920x0,1.25",
			expected: `hl.monitor({ output = "DP-1", mode = "2560x1440@143.91", position = "1920x0", scale = "1.25" })`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := KeywordToLua(tt.key, tt.val)
			if got != tt.expected {
				t.Errorf("KeywordToLua(%q, %q)\n got:  %s\n want: %s", tt.key, tt.val, got, tt.expected)
			}
		})
	}
}
