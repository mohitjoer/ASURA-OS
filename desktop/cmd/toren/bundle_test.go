package main

import (
	"strings"
	"testing"
)

type bundleFake struct {
	root, user []string
	enabled    []string
}

func fakeBundles(t *testing.T, repo map[string]bool, installed map[string]bool, gpus []string) *bundleFake {
	t.Helper()
	f := &bundleFake{}
	oldRepo, oldInst, oldGPU, oldRoot, oldUser, oldFP, oldEn := bundleRepoHas, bundleInstalled, bundleGPUs, bundleRoot, bundleUser, bundleHasFlatpak, bundleEnable
	t.Cleanup(func() {
		bundleRepoHas, bundleInstalled, bundleGPUs, bundleRoot, bundleUser, bundleHasFlatpak, bundleEnable = oldRepo, oldInst, oldGPU, oldRoot, oldUser, oldFP, oldEn
	})
	bundleRepoHas = func(p string) bool { return repo == nil || repo[p] }
	bundleInstalled = func(p string) bool { return installed[p] }
	bundleGPUs = func() []string { return gpus }
	bundleRoot = func(a ...string) error { f.root = append(f.root, strings.Join(a, " ")); return nil }
	bundleUser = func(a ...string) error { f.user = append(f.user, strings.Join(a, " ")); return nil }
	bundleHasFlatpak = func() bool { return true }
	oldX := bundleHasXbps
	t.Cleanup(func() { bundleHasXbps = oldX })
	bundleHasXbps = func() bool { return true }
	bundleEnable = func(n string) int { f.enabled = append(f.enabled, n); return exitOK }
	return f
}

func TestBundlePlanSkipsMissingAndInstalledAndAddsGPUDrivers(t *testing.T) {
	fakeBundles(t, map[string]bool{"steam": true, "lutris": true, "mesa-dri-32bit": true, "mesa-vulkan-intel-32bit": true, "nvidia-libs-32bit": true},
		map[string]bool{"lutris": true, "nvidia": true}, []string{"intel", "nvidia"})
	plan := planBundles([]string{"gaming"})
	got := strings.Join(plan.Packages, " ")
	for _, want := range []string{"steam", "mesa-vulkan-intel-32bit", "nvidia-libs-32bit"} {
		if !strings.Contains(got, want) {
			t.Errorf("plan lacks %s: %s", want, got)
		}
	}
	if strings.Contains(got, "lutris") {
		t.Errorf("an installed package was planned: %s", got)
	}
	if !strings.Contains(strings.Join(plan.Missing, " "), "retroarch") {
		t.Errorf("a package the repo lacks was not reported: %v", plan.Missing)
	}
}

func TestBundleUsesMesaWhenNVIDIAsDriverIsAbsent(t *testing.T) {
	fakeBundles(t, nil, nil, []string{"nvidia"})
	got := strings.Join(planBundles([]string{"gaming"}).Packages, " ")
	if strings.Contains(got, "nvidia-libs-32bit") || !strings.Contains(got, "mesa-dri-32bit") {
		t.Fatalf("%s", got)
	}
}

func TestBundleInstallSnapshotsInstallsFlatpaksAndEnablesServices(t *testing.T) {
	f := fakeBundles(t, nil, nil, nil)
	if code := installBundles([]string{"dev", "apps"}); code != exitOK {
		t.Fatalf("code %d", code)
	}
	root := strings.Join(f.root, "\n")
	if !strings.Contains(root, "snapshot create --reason install") || !strings.Contains(root, "xbps-install -y base-devel") {
		t.Fatalf("root commands:\n%s", root)
	}
	if strings.Index(root, "snapshot create") > strings.Index(root, "xbps-install -y") {
		t.Fatal("the snapshot must come before the install")
	}
	if !strings.Contains(strings.Join(f.user, "\n"), "flatpak install --user -y --noninteractive flathub org.signal.Signal") {
		t.Fatalf("user commands: %v", f.user)
	}
	if strings.Join(f.enabled, " ") != "docker tailscale" {
		t.Fatalf("enabled %v", f.enabled)
	}
}

func TestBundleRejectsUnknownNames(t *testing.T) {
	fakeBundles(t, nil, nil, nil)
	if code := asuraBundle([]string{"install", "dev", "everything"}); code != exitUsage {
		t.Fatalf("code %d", code)
	}
}

func TestGPUVendorsFromLspci(t *testing.T) {
	out := "00:02.0 VGA compatible controller: Intel Corporation Alder Lake-P GT2\n01:00.0 3D controller: NVIDIA Corporation GA107M [GeForce RTX 3050 Mobile]\n00:1f.3 Audio device: Intel Corporation\n"
	if got := strings.Join(gpuVendorsIn(out), " "); got != "intel nvidia" {
		t.Fatalf("got %q", got)
	}
}

func TestEveryBundleServiceIsSwitchable(t *testing.T) {
	for name, b := range bundles {
		for _, s := range b.Services {
			if _, ok := switchableServices[s]; !ok {
				t.Errorf("bundle %s switches on %q, which asura enable does not know", name, s)
			}
		}
	}
}
