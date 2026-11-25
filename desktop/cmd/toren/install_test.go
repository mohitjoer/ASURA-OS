package main

import (
	"testing"
)

func TestInstallSubcommand(t *testing.T) {
	cmd, taken := extraCommands["install"]
	if !taken {
		t.Fatalf("toren install command not registered")
	}

	if code := cmd.run([]string{"--help"}); code != exitOK {
		t.Errorf("toren install --help returned %d, want %d", code, exitOK)
	}
}
