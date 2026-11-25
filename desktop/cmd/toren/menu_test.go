package main

import (
	"testing"
)

func TestMenuUsage(t *testing.T) {
	if res := asuraMenu([]string{"--help"}); res != exitOK {
		t.Fatalf("asuraMenu(--help) returned %d, want %d", res, exitOK)
	}
	if res := asuraMenu([]string{"-h"}); res != exitOK {
		t.Fatalf("asuraMenu(-h) returned %d, want %d", res, exitOK)
	}
}
