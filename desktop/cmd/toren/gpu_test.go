package main

import (
	"testing"
)

func TestGpuSubcommand(t *testing.T) {
	orig := gpuPrivileged
	defer func() { gpuPrivileged = orig }()
	gpuPrivileged = func(args ...string) error { return nil }
	if code := torenGpu([]string{"--help"}); code != exitOK {
		t.Fatalf("toren gpu --help returned code %d, want %d", code, exitOK)
	}

	if code := torenGpu([]string{"status"}); code != exitOK {
		t.Fatalf("toren gpu status returned code %d, want %d", code, exitOK)
	}

	if code := torenGpu([]string{"status", "--json"}); code != exitOK {
		t.Fatalf("toren gpu status --json returned code %d, want %d", code, exitOK)
	}

	if code := torenGpu([]string{"switch"}); code != exitOK {
		t.Fatalf("toren gpu switch returned code %d, want %d", code, exitOK)
	}

	if code := torenGpu([]string{"switch", "hybrid"}); code != exitOK {
		t.Fatalf("toren gpu switch hybrid returned code %d, want %d", code, exitOK)
	}

	if code := torenGpu([]string{"switch", "invalid-mode"}); code != exitUsage {
		t.Fatalf("toren gpu switch invalid-mode returned code %d, want %d", code, exitUsage)
	}

	if code := torenGpu([]string{"nonexistent-verb"}); code != exitUsage {
		t.Fatalf("toren gpu nonexistent-verb returned code %d, want %d", code, exitUsage)
	}
}
