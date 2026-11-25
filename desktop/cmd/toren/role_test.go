package main

import "testing"

func TestTorenRoleUsage(t *testing.T) {
	code := torenRole([]string{"--help"})
	if code != exitOK {
		t.Fatalf("toren role --help returned code %d, want %d", code, exitOK)
	}

	code = torenRole([]string{"status"})
	if code != exitOK {
		t.Fatalf("toren role status returned code %d, want %d", code, exitOK)
	}

	code = torenRole([]string{"invalid-role"})
	if code != exitUsage {
		t.Fatalf("toren role invalid-role returned code %d, want %d", code, exitUsage)
	}
}
