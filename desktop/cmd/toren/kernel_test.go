package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetInstalledKernels(t *testing.T) {
	boot := t.TempDir()
	_ = os.WriteFile(filepath.Join(boot, "vmlinuz-6.12.10_1"), []byte("kernel"), 0o644)
	_ = os.WriteFile(filepath.Join(boot, "vmlinuz-6.6.68_1"), []byte("kernel"), 0o644)
	_ = os.WriteFile(filepath.Join(boot, "initramfs-6.12.10_1.img"), []byte("initrd"), 0o644)

	installed := getInstalledKernels(boot)
	if len(installed) != 2 {
		t.Fatalf("expected 2 kernels, got %d (%v)", len(installed), installed)
	}
}

func TestKernelSetUnsupported(t *testing.T) {
	code := kernelSet("freebsd-kernel")
	if code != exitFail {
		t.Errorf("expected exitFail for unsupported kernel, got %d", code)
	}
}

func TestKernelRemoveRefusesRunningOrLast(t *testing.T) {
	// Mock uname -r
	oldUname := kernelUnameR
	defer func() { kernelUnameR = oldUname }()
	kernelUnameR = func() (string, error) {
		return "6.12.10_1", nil
	}

	oldBoot := kernelBootDir
	defer func() { kernelBootDir = oldBoot }()
	boot := t.TempDir()
	kernelBootDir = boot
	_ = os.WriteFile(filepath.Join(boot, "vmlinuz-6.12.10_1"), []byte("kernel"), 0o644)

	// Refuses running kernel
	if code := kernelRemove("6.12.10_1"); code != exitFail {
		t.Errorf("expected exitFail when trying to remove running kernel, got %d", code)
	}

	// Refuses when 1 or fewer kernels in boot
	if code := kernelRemove("linux-lts"); code != exitFail {
		t.Errorf("expected exitFail when trying to remove last kernel, got %d", code)
	}
}
