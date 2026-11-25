package main

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestNoLegacyNames ensures that retired internal names
// do not reappear anywhere in the repository outside of shell-desktop and the migration script.
func TestNoLegacyNames(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatalf("resolving repository root: %v", err)
	}

	pat := "(^|[^a-zA-Z])[" + "R" + "r]" + "y" + "o[a-zA-Z]*|" + "R" + "Y" + "O"
	cmd := exec.Command("git", "grep", "-n", "-I", "-E", pat, "--", ":!shell-desktop", ":!migrations/user/1795364225.sh")
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		// git grep exits 1 when no matches are found
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return
		}
		t.Fatalf("running git grep: %v\noutput: %s", err, string(out))
	}

	var unexpected []string
	for _, line := range strings.Split(string(bytes.TrimSpace(out)), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.Contains(strings.ToLower(trimmed), "everyone") {
			continue
		}
		unexpected = append(unexpected, trimmed)
	}

	if len(unexpected) > 0 {
		t.Fatalf("legacy names found in repository:\n%s", strings.Join(unexpected, "\n"))
	}
}

// TestNoLegacyDesktopNames ensures that old desktop names
// do not reappear anywhere in the repository outside of vendored trees, legal/licence files,
// and the migration script.
func TestNoLegacyDesktopNames(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatalf("resolving repository root: %v", err)
	}

	pat := `\b` + "end" + `-?4\b|shell-` + "end4"
	cmd := exec.Command("git", "grep", "-n", "-I", "-E", pat, "--",
		":!desktop/cmd/toren-shell/legacy_names_test.go",
		":!shell-desktop",
		":!shell/vendor",
		":!shell/modules",
		":!shell/services",
		":!NOTICE.md",
		":!LICENSING.md",
		":!README.md",
		":!docs/architecture.md",
		":!shell/hub/pages/AboutPage.qml",
		":!desktop/internal/shell/vendor_test.go",
		":!migrations/user/1795442300.sh",
	)
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return
		}
		t.Fatalf("running git grep: %v\noutput: %s", err, string(out))
	}

	trimmed := strings.TrimSpace(string(out))
	if trimmed != "" {
		t.Fatalf("legacy desktop names found in repository outside allowed files:\n%s", trimmed)
	}
}
