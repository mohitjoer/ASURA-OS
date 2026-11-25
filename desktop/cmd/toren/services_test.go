package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// serviceSandbox points runit's folders at a temp tree and records the root
// commands instead of running them (performing ln and rm itself).
func serviceSandbox(t *testing.T, installed bool) *[]string {
	t.Helper()
	base := t.TempDir()
	oldDef, oldEn, oldPriv, oldInst := svDefinitions, svEnabled, svPrivileged, svInstalled
	t.Cleanup(func() { svDefinitions, svEnabled, svPrivileged, svInstalled = oldDef, oldEn, oldPriv, oldInst })
	svDefinitions, svEnabled = filepath.Join(base, "sv"), filepath.Join(base, "service")
	for _, d := range []string{svEnabled, filepath.Join(svDefinitions, "sshd"), filepath.Join(svDefinitions, "docker")} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	var ran []string
	svInstalled = func(string) bool { return installed }
	svPrivileged = func(args ...string) error {
		ran = append(ran, strings.Join(args, " "))
		switch args[0] {
		case "ln":
			return os.Symlink(args[2], args[3])
		case "rm":
			return os.Remove(args[2])
		}
		return nil
	}
	t.Setenv("SUDO_USER", "")
	return &ran
}

func TestEnableLinksTheServiceAndDisableUnlinksIt(t *testing.T) {
	ran := serviceSandbox(t, true)
	if code := torenEnable([]string{"ssh"}); code != exitOK {
		t.Fatalf("enable: %d", code)
	}
	if serviceState(switchableServices["ssh"]) != "on" {
		t.Fatal("sshd is not linked")
	}
	if code := torenEnable([]string{"ssh"}); code != exitOK || len(*ran) != 1 {
		t.Fatalf("a second enable should change nothing: %v", *ran)
	}
	if code := torenDisable([]string{"ssh"}); code != exitOK || serviceState(switchableServices["ssh"]) != "off" {
		t.Fatalf("disable: %d", code)
	}
}

func TestEnableInstallsAMissingPackageFirst(t *testing.T) {
	ran := serviceSandbox(t, false)
	torenEnable([]string{"ssh"})
	if len(*ran) == 0 || (*ran)[0] != "xbps-install -Sy openssh" {
		t.Fatalf("ran %v", *ran)
	}
}

func TestEnableRefusesUnknownServicesAndOnesWithNoDefinition(t *testing.T) {
	serviceSandbox(t, true)
	if code := torenEnable([]string{"teleport"}); code != exitUsage {
		t.Fatalf("unknown: %d", code)
	}
	if code := torenEnable([]string{"tailscale"}); code != exitFail {
		t.Fatalf("no /etc/sv/tailscaled: %d", code)
	}
}
