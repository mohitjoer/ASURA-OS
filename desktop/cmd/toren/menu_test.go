package main

import (
	"testing"
)

func TestMenuUsage(t *testing.T) {
	if res := torenMenu([]string{"--help"}); res != exitOK {
		t.Fatalf("torenMenu(--help) returned %d, want %d", res, exitOK)
	}
	if res := torenMenu([]string{"-h"}); res != exitOK {
		t.Fatalf("torenMenu(-h) returned %d, want %d", res, exitOK)
	}
}
