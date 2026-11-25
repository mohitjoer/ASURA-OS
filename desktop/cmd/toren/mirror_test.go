package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestMirrorConfigKeepsTheHyprlandRepositoryFirst(t *testing.T) {
	conf := mirrorConfig("https://repo-de.voidlinux.org", []string{
		"# comment",
		"repository=https://raw.githubusercontent.com/example/hyprland/",
		"repository=https://repo-fastly.voidlinux.org/current",
	})
	lines := strings.Split(strings.TrimSpace(conf), "\n")
	if lines[1] != "repository=https://raw.githubusercontent.com/example/hyprland/" {
		t.Fatalf("the Hyprland repository is not first: %q", conf)
	}
	if strings.Contains(conf, "fastly") || !strings.Contains(conf, "repository=https://repo-de.voidlinux.org/current/multilib/nonfree") {
		t.Fatalf("%q", conf)
	}
}

func TestRankMirrorsDropsTheSilentAndSortsTheRest(t *testing.T) {
	got := rankMirrors(voidMirrors[:3], func(url string) (time.Duration, error) {
		switch url {
		case voidMirrors[0].URL:
			return 300 * time.Millisecond, nil
		case voidMirrors[1].URL:
			return 0, errors.New("timeout")
		}
		return 100 * time.Millisecond, nil
	})
	if len(got) != 2 || got[0].mirror.URL != voidMirrors[2].URL {
		t.Fatalf("%+v", got)
	}
}

func TestMirrorSetWritesTheFileAndResetRemovesIt(t *testing.T) {
	dir := t.TempDir()
	oldConf, oldASURA, oldWrite := mirrorConf, mirrorASURAConf, mirrorWrite
	t.Cleanup(func() { mirrorConf, mirrorASURAConf, mirrorWrite = oldConf, oldASURA, oldWrite })
	mirrorConf, mirrorASURAConf = filepath.Join(dir, "00-toren-mirror.conf"), filepath.Join(dir, "toren.conf")
	mirrorWrite = func(path, content string) error {
		if content == "" {
			return os.RemoveAll(path)
		}
		return os.WriteFile(path, []byte(content), 0o644)
	}
	if err := os.WriteFile(mirrorASURAConf, []byte("repository=https://repo-fastly.voidlinux.org/current\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if currentMirror() != "https://repo-fastly.voidlinux.org" {
		t.Fatalf("current %q", currentMirror())
	}
	if code := torenMirror([]string{"set", "de"}); code != exitOK || currentMirror() != "https://repo-de.voidlinux.org" {
		t.Fatalf("set: %d %q", code, currentMirror())
	}
	if code := torenMirror([]string{"set", "http://insecure.example"}); code != exitUsage {
		t.Fatalf("plain http accepted: %d", code)
	}
	if code := torenMirror([]string{"reset"}); code != exitOK || currentMirror() != "https://repo-fastly.voidlinux.org" {
		t.Fatalf("reset: %d %q", code, currentMirror())
	}
}
