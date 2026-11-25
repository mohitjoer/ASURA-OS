package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAdoptPreflightRejections(t *testing.T) {
	tmpDir := t.TempDir()

	// 1. Missing /etc/os-release -> rejected
	rep := checkAdoptPreflight(tmpDir)
	if rep.Passed {
		t.Errorf("expected preflight to fail on missing /etc/os-release")
	}

	// 2. Non-Void distribution -> rejected
	etcDir := filepath.Join(tmpDir, "etc")
	_ = os.MkdirAll(etcDir, 0755)
	_ = os.WriteFile(filepath.Join(etcDir, "os-release"), []byte("ID=ubuntu\nNAME=Ubuntu\n"), 0644)
	rep = checkAdoptPreflight(tmpDir)
	if rep.Passed || !strings.Contains(rep.Error, "strictly requires Void Linux") {
		t.Errorf("expected rejection for non-Void OS, got %v", rep.Error)
	}

	// 3. Void Linux with musl -> rejected
	_ = os.WriteFile(filepath.Join(etcDir, "os-release"), []byte("ID=void\nNAME=\"Void Linux\"\n"), 0644)
	libDir := filepath.Join(tmpDir, "lib")
	_ = os.MkdirAll(libDir, 0755)
	_ = os.WriteFile(filepath.Join(libDir, "ld-musl-x86_64.so.1"), []byte("musl"), 0755)
	rep = checkAdoptPreflight(tmpDir)
	if rep.Passed || !strings.Contains(rep.Error, "musl detected") {
		t.Errorf("expected rejection for musl, got %v", rep.Error)
	}

	// 4. Void Linux with glibc -> passes
	_ = os.Remove(filepath.Join(libDir, "ld-musl-x86_64.so.1"))
	rep = checkAdoptPreflight(tmpDir)
	if !rep.Passed {
		t.Errorf("expected preflight to pass for Void glibc, got error: %v", rep.Error)
	}
}

func TestAdoptLifecycleExecuteVerifyUndo(t *testing.T) {
	tmpRoot := t.TempDir()

	// Setup fake Void glibc root
	etcDir := filepath.Join(tmpRoot, "etc")
	_ = os.MkdirAll(etcDir, 0755)
	_ = os.WriteFile(filepath.Join(etcDir, "os-release"), []byte("ID=void\nNAME=\"Void Linux\"\n"), 0644)

	// Existing sddm.conf and environment to back up
	origSDDM := "[Theme]\nCurrent=breeze\n"
	origEnv := "LANG=en_US.UTF-8\n"
	_ = os.WriteFile(filepath.Join(etcDir, "sddm.conf"), []byte(origSDDM), 0644)
	_ = os.WriteFile(filepath.Join(etcDir, "environment"), []byte(origEnv), 0644)

	// 1. Execute adoption
	rc := runAdoptExecute(tmpRoot)
	if rc != exitOK {
		t.Fatalf("runAdoptExecute failed with code %d", rc)
	}

	// Verify state.json exists
	statePath := filepath.Join(tmpRoot, "etc", "toren", "adopt", "state.json")
	if _, err := os.Stat(statePath); err != nil {
		t.Fatalf("expected state.json at %s, got error: %v", statePath, err)
	}

	// Verify created registrations exist
	sessionPath := filepath.Join(tmpRoot, "usr", "share", "wayland-sessions", "toren.desktop")
	if _, err := os.Stat(sessionPath); err != nil {
		t.Fatalf("expected toren.desktop at %s", sessionPath)
	}
	portalPath := filepath.Join(tmpRoot, "usr", "share", "xdg-desktop-portal", "toren-portals.conf")
	if _, err := os.Stat(portalPath); err != nil {
		t.Fatalf("expected toren-portals.conf at %s", portalPath)
	}

	// 2. Verify check passes
	vrc := runAdoptVerify(tmpRoot, false)
	if vrc != exitOK {
		t.Fatalf("runAdoptVerify failed with code %d", vrc)
	}

	// 3. Undo adoption
	urc := runAdoptUndo(tmpRoot)
	if urc != exitOK {
		t.Fatalf("runAdoptUndo failed with code %d", urc)
	}

	// Verify toren.desktop and toren-portals.conf removed
	if _, err := os.Stat(sessionPath); !os.IsNotExist(err) {
		t.Fatalf("expected toren.desktop to be removed after undo")
	}
	if _, err := os.Stat(portalPath); !os.IsNotExist(err) {
		t.Fatalf("expected toren-portals.conf to be removed after undo")
	}

	// Verify original files restored byte-for-byte
	sddmData, _ := os.ReadFile(filepath.Join(etcDir, "sddm.conf"))
	if string(sddmData) != origSDDM {
		t.Fatalf("sddm.conf restore mismatch: want %q, got %q", origSDDM, string(sddmData))
	}
	envData, _ := os.ReadFile(filepath.Join(etcDir, "environment"))
	if string(envData) != origEnv {
		t.Fatalf("environment restore mismatch: want %q, got %q", origEnv, string(envData))
	}

	// Verify state.json removed
	if _, err := os.Stat(statePath); !os.IsNotExist(err) {
		t.Fatalf("expected state.json to be removed after undo")
	}
}
