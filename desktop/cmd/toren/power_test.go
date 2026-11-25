package main

import (
	"testing"
)

func TestPowerSubcommand(t *testing.T) {
	orig := powerPrivileged
	defer func() { powerPrivileged = orig }()
	powerPrivileged = func(args ...string) error { return nil }

	if code := asuraPower([]string{"--help"}); code != exitOK {
		t.Fatalf("asura power --help returned code %d, want %d", code, exitOK)
	}

	if code := asuraPower([]string{"status"}); code != exitOK {
		t.Fatalf("asura power status returned code %d, want %d", code, exitOK)
	}

	if code := asuraPower([]string{"profile"}); code != exitOK {
		t.Fatalf("asura power profile returned code %d, want %d", code, exitOK)
	}

	if code := asuraPower([]string{"profile", "balanced"}); code != exitOK {
		t.Fatalf("asura power profile balanced returned code %d, want %d", code, exitOK)
	}

	if code := asuraPower([]string{"profile", "invalid"}); code != exitUsage {
		t.Fatalf("asura power profile invalid returned code %d, want %d", code, exitUsage)
	}

	if code := asuraPower([]string{"threshold"}); code != exitOK {
		t.Fatalf("asura power threshold returned code %d, want %d", code, exitOK)
	}

	if code := asuraPower([]string{"threshold", "80"}); code != exitOK {
		t.Fatalf("asura power threshold 80 returned code %d, want %d", code, exitOK)
	}

	if code := asuraPower([]string{"threshold", "invalid"}); code != exitUsage {
		t.Fatalf("asura power threshold invalid returned code %d, want %d", code, exitUsage)
	}

	if code := asuraPower([]string{"sleep"}); code != exitOK {
		t.Fatalf("asura power sleep returned code %d, want %d", code, exitOK)
	}

	if code := asuraPower([]string{"sleep", "s2idle"}); code != exitOK {
		t.Fatalf("asura power sleep s2idle returned code %d, want %d", code, exitOK)
	}

	if code := asuraPower([]string{"sleep", "invalid"}); code != exitUsage {
		t.Fatalf("asura power sleep invalid returned code %d, want %d", code, exitUsage)
	}

	if code := asuraPower([]string{"unknown"}); code != exitUsage {
		t.Fatalf("asura power unknown returned code %d, want %d", code, exitUsage)
	}
}
