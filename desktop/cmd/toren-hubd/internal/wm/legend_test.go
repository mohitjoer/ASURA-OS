package wm

import "testing"

// A digit family's number-pad copies (KP_<n> and the NumLock-off keysym) are the
// same shortcut: the family's row stands for them, so they are not listed again
// as the compositor's own binds.
func TestBuildBindRowsFoldsNumpadTwinsIntoFamily(t *testing.T) {
	cat := []CatalogBind{{
		ID: "workspace-movesilent", Category: "workspaces", Label: "Send the window to a workspace",
		Chord: "SUPER + SHIFT + {n}", Family: true,
	}}
	var live []liveBind
	for _, c := range []string{"SUPER + SHIFT + 1", "SUPER + SHIFT + KP_1", "SUPER + SHIFT + KP_End", "SUPER + F12"} {
		live = append(live, liveBind{chord: chordKey(c), display: c, dispatcher: "HL.Dispatcher(exec_cmd)", arg: "440"})
	}
	rows := buildBindRows(cat, Overrides{}, live)
	var extra []string
	for _, r := range rows[1:] {
		extra = append(extra, r.Chord)
	}
	if len(extra) != 1 || extra[0] != "SUPER + F12" {
		t.Fatalf("rows after the family = %v, want only SUPER + F12", extra)
	}
	if got := rows[1].Label; got != "Run a command" {
		t.Errorf("label of a Lua exec bind = %q, want %q", got, "Run a command")
	}
}

func TestNumpadTwin(t *testing.T) {
	for in, want := range map[string]string{
		"SUPER + 3":         "SUPER + KP_3",
		"SUPER + SHIFT + 0": "SUPER + SHIFT + KP_0",
		"SUPER + KP_3":      "",
		"SUPER + Q":         "",
	} {
		if got := numpadTwin(in); got != want {
			t.Errorf("numpadTwin(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestDescribeLuaDispatcher(t *testing.T) {
	for d, want := range map[string]string{
		"HL.Dispatcher(exec_cmd)":     "Run a command",
		"HL.Dispatcher(window.close)": "Close window",
		"HL.Dispatcher(layout.swap)":  "layout swap",
		"killactive":                  "Close window",
	} {
		if got := describeDispatcher(d, "440"); got != want {
			t.Errorf("describeDispatcher(%q) = %q, want %q", d, got, want)
		}
	}
}
