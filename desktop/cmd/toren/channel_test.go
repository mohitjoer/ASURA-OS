package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestXbpsChannel(t *testing.T) {
	tests := []struct {
		name    string
		sources string // empty: no file at all
		want    string
		wantErr string
	}{
		{"channel key", "channel=stable\nrepository=https://repo.torenos.dev/current\n", "stable", ""},
		{"repository key", "repository=https://repo.torenos.dev/current\n", "https://repo.torenos.dev/current", ""},
		{"comment and spacing", "# Toren OS\n  repository = https://repo.torenos.dev/testing  \n", "https://repo.torenos.dev/testing", ""},
		{"empty file", "# Only comments\n\n", "", "names no channel or repository"},
		{"not configured", "", "", "no Toren repository is configured"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			if tc.sources != "" {
				path := filepath.Join(root, xbpsRepoConf)
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte(tc.sources), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			got, err := xbpsChannel(root)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("got %q, %v; want an error with %q", got, err, tc.wantErr)
				}
				return
			}
			if err != nil || got != tc.want {
				t.Fatalf("got %q, %v; want %q", got, err, tc.want)
			}
		})
	}
}

// `set` is not built; it must say so rather than succeed.
func TestChannelOnlyHasGet(t *testing.T) {
	if code := torenChannel([]string{"set", "testing"}); code != exitUsage {
		t.Errorf("toren channel set exits %d, want %d", code, exitUsage)
	}
	if code := torenChannel(nil); code != exitUsage {
		t.Errorf("toren channel with no verb exits %d, want %d", code, exitUsage)
	}
}
