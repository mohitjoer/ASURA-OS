package main

import (
	"fmt"

	"asura/desktop/cmd/toren-hubd/internal/wm"
)

// The keybind legend the Hub keybinds page and the Super+K cheatsheet draw.
// ASURA's own shortcuts are shown read-only (project rule: ASURA shortcuts cannot be changed).
// User-added custom bindings in ~/.config/asura/hyprland/ are shown and editable.

type bind struct {
	wm.BindRow
	Desc  string `json:"desc"`
	Combo string `json:"combo"`
}

type category struct {
	Name  string `json:"name"`
	Binds []bind `json:"binds"`
}

type legend struct {
	Categories []category `json:"categories"`
}

func runKeybinds(args []string) error {
	if len(args) == 0 || args[0] == "list" {
		return printJSON(keybinds())
	}
	switch args[0] {
	case "add":
		if len(args) < 3 {
			return fmt.Errorf("keybinds add needs <chord> <command> [label]")
		}
		label := args[1]
		if len(args) >= 4 {
			label = args[3]
		}
		return addCustomKeybind(args[1], args[2], label)
	case "remove", "rm":
		if len(args) < 2 {
			return fmt.Errorf("keybinds remove needs <chord>")
		}
		return removeCustomKeybind(args[1])
	default:
		return fmt.Errorf("unknown keybinds subcommand %q", args[0])
	}
}

// keybinds groups the provider's effective legend into sections.
// Shipped binds are marked Locked=true, Rebindable=false.
// User custom binds are added under "Custom Shortcuts".
func keybinds() legend {
	l := legend{Categories: []category{}}
	rows, err := wm.Open().Binds()
	if err != nil {
		return l
	}

	at := map[string]int{}
	for _, r := range rows {
		i, ok := at[r.Category]
		if !ok {
			i = len(l.Categories)
			at[r.Category] = i
			l.Categories = append(l.Categories, category{Name: r.Category, Binds: []bind{}})
		}
		l.Categories[i].Binds = append(l.Categories[i].Binds, bind{
			BindRow: r,
			Desc:    r.Label,
			Combo:   r.Default,
		})
	}

	// Add user-added custom binds from desktop.json
	o := wm.LoadStore(desktopStorePath())
	if len(o.Keybinds) > 0 {
		var customBinds []bind
		for _, k := range o.Keybinds {
			chord := wm.NormChord(k.Keys)
			customBinds = append(customBinds, bind{
				BindRow: wm.BindRow{
					ID:         "custom-" + chord,
					Category:   "Custom Shortcuts",
					Label:      k.Value,
					Keys:       wm.DisplayKeys(chord),
					Default:    chord,
					Chord:      chord,
					Kind:       wm.BindCustom,
					Rebindable: true,
					Locked:     false,
				},
				Desc:  k.Value,
				Combo: chord,
			})
		}
		if len(customBinds) > 0 {
			l.Categories = append(l.Categories, category{
				Name:  "Custom Shortcuts",
				Binds: customBinds,
			})
		}
	}

	return l
}

func addCustomKeybind(keys, command, label string) error {
	return withDesktopLock(func() error {
		o := wm.LoadStore(desktopStorePath())
		chord := wm.NormChord(keys)
		found := false
		for i := range o.Keybinds {
			if wm.NormChord(o.Keybinds[i].Keys) == chord {
				o.Keybinds[i].Action = "exec"
				o.Keybinds[i].Value = command
				found = true
				break
			}
		}
		if !found {
			o.Keybinds = append(o.Keybinds, wm.Keybind{
				Keys:   chord,
				Action: "exec",
				Value:  command,
			})
		}
		if err := setHyprSections(map[string]any{"keybinds": o.Keybinds}); err != nil {
			return err
		}
		if err := applyLive(); err != nil {
			return err
		}
		return printJSON(map[string]string{
			"status":  "added",
			"chord":   chord,
			"command": command,
		})
	})
}

func removeCustomKeybind(keys string) error {
	return withDesktopLock(func() error {
		o := wm.LoadStore(desktopStorePath())
		chord := wm.NormChord(keys)
		var kept []wm.Keybind
		for _, k := range o.Keybinds {
			if wm.NormChord(k.Keys) != chord {
				kept = append(kept, k)
			}
		}
		o.Keybinds = kept
		if err := setHyprSections(map[string]any{"keybinds": o.Keybinds}); err != nil {
			return err
		}
		if err := applyLive(); err != nil {
			return err
		}
		return printJSON(map[string]string{
			"status": "removed",
			"chord":  chord,
		})
	})
}
