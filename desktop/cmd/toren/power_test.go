package main

import (
	"testing"
)

func TestPowerSubcommand(t *testing.T) {
	orig := powerPrivileged
	defer func() { powerPrivileged = orig }()
	powerPrivileged = func(args ...string) error { return nil }

	if code := torenPower([]string{"--help"}); code != exitOK {
		t.Fatalf("toren power --help returned code %d, want %d", code, exitOK)
	}

	if code := torenPower([]string{"status"}); code != exitOK {
		t.Fatalf("toren power status returned code %d, want %d", code, exitOK)
	}

	if code := torenPower([]string{"profile"}); code != exitOK {
		t.Fatalf("toren power profile returned code %d, want %d", code, exitOK)
	}

	if code := torenPower([]string{"profile", "balanced"}); code != exitOK {
		t.Fatalf("toren power profile balanced returned code %d, want %d", code, exitOK)
	}

	if code := torenPower([]string{"profile", "invalid"}); code != exitUsage {
		t.Fatalf("toren power profile invalid returned code %d, want %d", code, exitUsage)
	}

	if code := torenPower([]string{"threshold"}); code != exitOK {
		t.Fatalf("toren power threshold returned code %d, want %d", code, exitOK)
	}

	if code := torenPower([]string{"threshold", "80"}); code != exitOK {
		t.Fatalf("toren power threshold 80 returned code %d, want %d", code, exitOK)
	}

	if code := torenPower([]string{"threshold", "invalid"}); code != exitUsage {
		t.Fatalf("toren power threshold invalid returned code %d, want %d", code, exitUsage)
	}

	if code := torenPower([]string{"sleep"}); code != exitOK {
		t.Fatalf("toren power sleep returned code %d, want %d", code, exitOK)
	}

	if code := torenPower([]string{"sleep", "s2idle"}); code != exitOK {
		t.Fatalf("toren power sleep s2idle returned code %d, want %d", code, exitOK)
	}

	if code := torenPower([]string{"sleep", "invalid"}); code != exitUsage {
		t.Fatalf("toren power sleep invalid returned code %d, want %d", code, exitUsage)
	}

	if code := torenPower([]string{"unknown"}); code != exitUsage {
		t.Fatalf("toren power unknown returned code %d, want %d", code, exitUsage)
	}
}
