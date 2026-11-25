package main

import (
	"reflect"
	"testing"

	"asura/desktop/internal/runner/runnertest"
)

func TestAsuraCommands(t *testing.T) {
	config := `
# bind = SUPER, X, exec, asura-commented-out
bind = SUPER, Space, exec, asura ipc call asura togglePopover launcher
bind = SUPER, 1, exec, asura-workspace focus 1
bind = SUPER, 2, exec, asura-workspace focus 2
bindr = ALT, Tab, exec, [float] asura-app terminal
bind = SUPER, A, exec, hyprctl dispatch togglefloating
bind = ,XF86AudioMute, exec, wpctl set-mute @DEFAULT_AUDIO_SINK@ toggle
bind = SUPER, Q, killactive,
exec-once = command -v asura-monitor >/dev/null 2>&1 && asura-monitor autoscale
exec-once = /usr/bin/asura-welcome --if-first-run
exec-once = /usr/share/asura/session/asura-activate
`
	want := []string{"asura", "asura-app", "asura-welcome", "asura-workspace"}
	if got := asuraCommands(config); !reflect.DeepEqual(got, want) {
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
