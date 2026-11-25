package main

import (
	"testing"
)

func TestUpdatesStatus(t *testing.T) {
	// updatesStatus should return without error
	if err := updatesStatus(); err != nil {
		t.Fatalf("updatesStatus failed: %v", err)
	}

	// listSnapshots should return without error
	if err := listSnapshots(); err != nil {
		t.Fatalf("listSnapshots failed: %v", err)
	}
}
