package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestChannelSetPinsTheBranchUpdatesFollow(t *testing.T) {
	root := t.TempDir()
	if code := channelSet(root, "stable"); code != exitOK {
		t.Fatalf("set stable: %d", code)
	}
	if got := pinnedChannel(root); got != "stable" || channelBranches[got] != "main" {
		t.Fatalf("pinned %q", got)
	}
	if code := channelSet(root, "nightly"); code != exitUsage {
		t.Fatalf("an unknown channel was accepted: %d", code)
	}
	if code := torenChannel([]string{"set", "unstable", "--root", root}); code != exitOK || pinnedChannel(root) != "unstable" {
		t.Fatalf("toren channel set: %d %q", code, pinnedChannel(root))
	}
	if err := os.WriteFile(filepath.Join(root, channelFile), []byte("garbage\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if pinnedChannel(root) != "" {
		t.Fatal("an unknown value in the file must be ignored")
	}
}

func withKernels(t *testing.T, running string, installed ...string) {
	t.Helper()
	dir := t.TempDir()
	for _, k := range installed {
		if err := os.MkdirAll(filepath.Join(dir, k), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	oldDir, oldRun := modulesDir, bootedKernel
	t.Cleanup(func() { modulesDir, bootedKernel = oldDir, oldRun })
	modulesDir = dir
	bootedKernel = func() string { return running }
}

func TestRebootNeededWhenANewerKernelIsInstalled(t *testing.T) {
	withKernels(t, "6.12.9_1", "6.12.9_1", "6.12.10_1", "6.6.70_1")
	if newest, ok := rebootNeeded(); !ok || newest != "6.12.10_1" {
		t.Fatalf("got %q %v", newest, ok)
	}
}

func TestNoRebootWhenRunningTheNewest(t *testing.T) {
	withKernels(t, "6.12.10_1", "6.12.9_1", "6.12.10_1")
	if _, ok := rebootNeeded(); ok {
		t.Fatal("asked for a reboot while running the newest kernel")
	}
	withKernels(t, "6.12.10_1", "6.6.70_1") // a running kernel newer than any modules dir
	if _, ok := rebootNeeded(); ok {
		t.Fatal("asked for a reboot to an older kernel")
	}
}

func TestNotifyUpdatesAnnouncesEachUpdateOnce(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	withKernels(t, "6.12.10_1", "6.12.10_1")
	var sent []string
	oldSend, oldCheck := notifySend, updateCheckOutput
	t.Cleanup(func() { notifySend, updateCheckOutput = oldSend, oldCheck })
	notifySend = func(title, body string) error { sent = append(sent, title+": "+body); return nil }
	check := "update available: aaaaaaaa -> bbbbbbbb\n"
	updateCheckOutput = func() (string, int) { return check, exitNoMatch }

	torenNotifyUpdates(nil)
	torenNotifyUpdates(nil)
	if len(sent) != 1 || !strings.Contains(sent[0], "bbbbbbbb") {
		t.Fatalf("sent %v", sent)
	}
	check = "update available: aaaaaaaa -> cccccccc\n"
	torenNotifyUpdates(nil)
	if len(sent) != 2 {
		t.Fatalf("a newer update was not announced: %v", sent)
	}
	updateCheckOutput = func() (string, int) { return "up to date\n", exitOK }
	torenNotifyUpdates(nil)
	if len(sent) != 2 {
		t.Fatalf("announced while up to date: %v", sent)
	}
}
