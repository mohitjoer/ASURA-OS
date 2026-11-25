package shell

import (
	"path/filepath"
	"testing"
)

func TestBindsParityBetweenHyprlangAndLua(t *testing.T) {
	root := "../../.."
	confPath := filepath.Join(root, "session/hyprland/hyprland.conf")
	luaPath := filepath.Join(root, "session/hyprland/hyprland.lua")
	catPath := filepath.Join(root, "session/binds.json")

	cat, err := LoadCatalogue(catPath)
	if err != nil {
		t.Fatalf("LoadCatalogue: %v", err)
	}

	confBody, err := readCompositorConfig(confPath)
	if err != nil {
		t.Fatalf("readCompositorConfig(conf): %v", err)
	}
	luaBody, err := readCompositorConfig(luaPath)
	if err != nil {
		t.Fatalf("readCompositorConfig(lua): %v", err)
	}

	confBinds := parseBinds(string(confBody))
	luaBinds := parseBinds(string(luaBody))

	if len(confBinds) == 0 {
		t.Fatal("confBinds is empty")
	}
	if len(luaBinds) == 0 {
		t.Fatal("luaBinds is empty")
	}

	if len(confBinds) != len(luaBinds) {
		t.Fatalf("bind count mismatch: conf has %d, lua has %d", len(confBinds), len(luaBinds))
	}

	confMap := make(map[string]compositorBind)
	for _, b := range confBinds {
		key := b.Chord
		if b.isRelease() {
			key += "@release"
		}
		confMap[key] = b
	}

	for _, lb := range luaBinds {
		key := lb.Chord
		if lb.isRelease() {
			key += "@release"
		}
		cb, ok := confMap[key]
		if !ok {
			t.Errorf("lua binds chord %q (release=%v) not found in conf", lb.Chord, lb.isRelease())
			continue
		}
		lAct, lArg := lb.action()
		cAct, cArg := cb.action()
		if lAct != cAct || lArg != cArg {
			t.Errorf("chord %q action mismatch: lua has (%q, %q), conf has (%q, %q)",
				lb.Chord, lAct, lArg, cAct, cArg)
		}
	}

	findingsConf := compareBinds(cat, catPath, confPath, string(confBody))
	if len(findingsConf) != 0 {
		for _, f := range findingsConf {
			t.Errorf("conf finding: %s: %s", f.Path, f.Detail)
		}
	}

	findingsLua := compareBinds(cat, catPath, luaPath, string(luaBody))
	if len(findingsLua) != 0 {
		for _, f := range findingsLua {
			t.Errorf("lua finding: %s: %s", f.Path, f.Detail)
		}
	}
}
