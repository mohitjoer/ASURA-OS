package main

import (
	"reflect"
	"testing"

	"toren/desktop/internal/runner/runnertest"
)

func TestTorenCommands(t *testing.T) {
	config := `
# bind = SUPER, X, exec, toren-commented-out
bind = SUPER, Space, exec, toren ipc call toren togglePopover launcher
bind = SUPER, 1, exec, toren-workspace focus 1
bind = SUPER, 2, exec, toren-workspace focus 2
bindr = ALT, Tab, exec, [float] toren-app terminal
bind = SUPER, A, exec, hyprctl dispatch 'hl.dsp.window.float()'
bind = ,XF86AudioMute, exec, wpctl set-mute @DEFAULT_AUDIO_SINK@ toggle
bind = SUPER, Q, killactive,
exec-once = command -v toren-monitor >/dev/null 2>&1 && toren-monitor autoscale
exec-once = /usr/bin/toren-welcome --if-first-run
exec-once = /usr/share/toren/session/toren-activate
`
	want := []string{"toren", "toren-app", "toren-welcome", "toren-workspace"}
	if got := torenCommands(config); !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestRecommendedAbsent(t *testing.T) {
	fake := runnertest.New().
		Out("xbps-query -l",
			"ii hyprland-0.49.0_1 Dynamic tiling Wayland compositor\n"+
				"ii quickshell-0.0.1_1 Quickshell desktop shell\n"+
				"ii pipewire-1.0.0_1 PipeWire audio server\n"+
				"ii wireplumber-0.5.0_1 Session manager for PipeWire\n"+
				"ii seatd-0.8.0_1 Seat management daemon\n"+
				"ii NetworkManager-1.44.0_1 Network management daemon\n"+
				"ii ghostty-1.0.0_1 Fast native terminal\n")

	absent, total, err := recommendedAbsent(fake.Run)
	if err != nil {
		t.Fatal(err)
	}
	// xdg-desktop-portal-hyprland is the only missing one in this mock
	if want := []string{"xdg-desktop-portal-hyprland"}; !reflect.DeepEqual(absent, want) || total != 8 {
		t.Errorf("absent %v of %d, want %v of 8", absent, total, want)
	}
}

func TestParseSvStatus(t *testing.T) {
	cases := map[string]svState{
		"run: turnstiled: (pid 1328) 826s; run: log: (pid 1327) 826s\n":                svRunning,
		"down: elogind: 1s, normally up, want up; run: log: (pid 1324) 826s\n":         svFlapping,
		"down: tlp: 40s, normally up; run: log: (pid 9) 40s\n":                         svStopped,
		"fail: scx-lavd: unable to change to service directory: file does not exist\n": svMissing,
		"warning: x: unable to open supervise/ok: file does not exist\n":               svMissing,
		"": svUnknown,
	}
	for in, want := range cases {
		if got := parseSvStatus(in); got != want {
			t.Errorf("parseSvStatus(%q) = %d, want %d", in, got, want)
		}
	}
}
