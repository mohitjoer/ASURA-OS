package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSnapshotsListNewestFirstAcrossKinds(t *testing.T) {
	dir := t.TempDir()
	for _, n := range []string{
		"pre-update-20261101-120000",
		"pre-install-20261115-090000",
		"pre-update-20261120-080000",
		"manual-20261110-100000",
		"before-big-change",
	} {
		if err := os.MkdirAll(filepath.Join(dir, n), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	got, err := listSnapshotsIn(dir)
	if err != nil {
		t.Fatal(err)
	}
	want := "pre-update-20261120-080000 pre-install-20261115-090000 manual-20261110-100000 pre-update-20261101-120000"
	if !strings.HasPrefix(strings.Join(got, " "), want) || len(got) != 5 {
		t.Fatalf("got %v", got)
	}
}

func TestDescribeSnapshotSaysWhenWhyAndWhichASURA(t *testing.T) {
	dir := t.TempDir()
	name := "pre-install-20261115-090000"
	commitFile := filepath.Join(dir, name, appliedCommitFile)
	if err := os.MkdirAll(filepath.Dir(commitFile), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(commitFile, []byte("0123456789abcdef\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	line := describeSnapshot(dir, name)
	for _, want := range []string{name, "2026-11-15 09:00", "before installing or removing packages", "(ASURA 01234567)"} {
		if !strings.Contains(line, want) {
			t.Errorf("%q lacks %q", line, want)
		}
	}
	if err := os.MkdirAll(filepath.Join(dir, "by-hand"), 0o755); err != nil {
		t.Fatal(err)
	}
	if line := describeSnapshot(dir, "by-hand"); !strings.Contains(line, "by hand or by another tool") {
		t.Errorf("%q", line)
	}
}

func TestPruningKeepsTheNewestOfEachKindAndNeverManualOnes(t *testing.T) {
	names := []string{
		"pre-update-1", "pre-update-2", "pre-update-3",
		"pre-install-1", "pre-install-2",
		"manual-1", "manual-2", "manual-3",
	}
	if got := oldSnapshots(names, snapshotPreUpdate, 2); strings.Join(got, " ") != "pre-update-1" {
		t.Fatalf("update: %v", got)
	}
	if got := oldSnapshots(names, snapshotPreInstall, 2); len(got) != 0 {
		t.Fatalf("install: %v", got)
	}
}

func TestSnapshotCreateRejectsAnUnknownReason(t *testing.T) {
	if code := asuraSnapshot([]string{"create", "--reason", "whenever"}); code != exitUsage {
		t.Fatalf("code %d", code)
	}
}
