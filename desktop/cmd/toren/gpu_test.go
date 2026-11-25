package main

import (
	"testing"
)

func TestGpuSubcommand(t *testing.T) {
	orig := gpuPrivileged
	defer func() { gpuPrivileged = orig }()
	gpuPrivileged = func(args ...string) error { return nil }
	if code := asuraGpu([]string{"--help"}); code != exitOK {
		t.Fatalf("asura gpu --help returned code %d, want %d", code, exitOK)
	}

	if code := asuraGpu([]string{"status"}); code != exitOK {
		t.Fatalf("asura gpu status returned code %d, want %d", code, exitOK)
	}

	if code := asuraGpu([]string{"status", "--json"}); code != exitOK {
		t.Fatalf("asura gpu status --json returned code %d, want %d", code, exitOK)
	}

	if code := asuraGpu([]string{"switch"}); code != exitOK {
		t.Fatalf("asura gpu switch returned code %d, want %d", code, exitOK)
	}

	if code := asuraGpu([]string{"switch", "hybrid"}); code != exitOK {
		t.Fatalf("asura gpu switch hybrid returned code %d, want %d", code, exitOK)
	}

	if code := asuraGpu([]string{"switch", "invalid-mode"}); code != exitUsage {
		t.Fatalf("asura gpu switch invalid-mode returned code %d, want %d", code, exitUsage)
	}

	if code := asuraGpu([]string{"nonexistent-verb"}); code != exitUsage {
		t.Fatalf("asura gpu nonexistent-verb returned code %d, want %d", code, exitUsage)
	}
}
