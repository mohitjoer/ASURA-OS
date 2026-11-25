package main

import (
	"bytes"
	"strings"
	"testing"

	"asura/desktop/internal/runner/runnertest"
)

func TestHoldList(t *testing.T) {
	withNoSudo(t)
	fake := runnertest.New().Out("xbps-query -H", "hyprland\nhyprutils\n")

	var out, errOut bytes.Buffer
	code := holdMain(nil, &out, &errOut, fake.Run)
	if code != exitOK {
		t.Fatalf("holdMain(nil) returned code %d, want %d", code, exitOK)
	}

	if !strings.Contains(out.String(), "hyprland") {
		t.Errorf("expected hyprland in output, got: %s", out.String())
	}
	if !fake.Ran("xbps-query -H") {
		t.Errorf("expected xbps-query -H to run, calls: %v", fake.Calls())
	}
}

func TestHoldListEmpty(t *testing.T) {
	withNoSudo(t)
	fake := runnertest.New().Out("xbps-query -H", "")

	var out, errOut bytes.Buffer
	code := holdMain(nil, &out, &errOut, fake.Run)
	if code != exitOK {
		t.Fatalf("holdMain(nil) returned code %d, want %d", code, exitOK)
	}

	if !strings.Contains(out.String(), "No packages are currently held.") {
		t.Errorf("expected empty message, got: %s", out.String())
	}
}

func TestHoldPackages(t *testing.T) {
	withNoSudo(t)
	fake := runnertest.New().Out("xbps-pkgdb -m hold hyprland hyprutils", "")

	var out, errOut bytes.Buffer
	code := holdMain([]string{"hyprland", "hyprutils"}, &out, &errOut, fake.Run)
	if code != exitOK {
		t.Fatalf("holdMain returned code %d, want %d", code, exitOK)
	}

	if !fake.Ran("xbps-pkgdb -m hold hyprland hyprutils") {
		t.Errorf("expected xbps-pkgdb to run, calls: %v", fake.Calls())
	}
	if !strings.Contains(out.String(), "Held package(s): hyprland hyprutils") {
		t.Errorf("expected confirmation message, got: %s", out.String())
	}
}

func TestUnholdUsageWithoutArgs(t *testing.T) {
	withNoSudo(t)
	fake := runnertest.New()

	var out, errOut bytes.Buffer
	code := unholdMain(nil, &out, &errOut, fake.Run)
	if code != exitUsage {
		t.Fatalf("unholdMain(nil) returned code %d, want %d", code, exitUsage)
	}
}

func TestUnholdPackages(t *testing.T) {
	withNoSudo(t)
	fake := runnertest.New().Out("xbps-pkgdb -m unhold hyprland", "")

	var out, errOut bytes.Buffer
	code := unholdMain([]string{"hyprland"}, &out, &errOut, fake.Run)
	if code != exitOK {
		t.Fatalf("unholdMain returned code %d, want %d", code, exitOK)
	}

	if !fake.Ran("xbps-pkgdb -m unhold hyprland") {
		t.Errorf("expected xbps-pkgdb unhold to run, calls: %v", fake.Calls())
	}
	if !strings.Contains(out.String(), "Unheld package(s): hyprland") {
		t.Errorf("expected confirmation message, got: %s", out.String())
	}
}
