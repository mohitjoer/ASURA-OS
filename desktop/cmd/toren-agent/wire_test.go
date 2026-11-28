package main

import (
	"os"
	"path/filepath"
	"testing"
)

// setupTestEnv sets up an isolated temporary HOME and mock Toren skills root with
// all four skill directories.
func setupTestEnv(t *testing.T) (tempHome string, skillsRoot string) {
	t.Helper()

	tempHome = t.TempDir()
	t.Setenv("HOME", tempHome)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tempHome, ".config"))

	skillsRoot = t.TempDir()

	// Create a SKILL.md stub in each skill subdirectory.
	for _, name := range skillNames {
		dir := filepath.Join(skillsRoot, name)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		content := "---\nname: " + name + "\ndescription: Toren skill " + name + "\n---\n# " + name + "\n"
		if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// Point ASURA_SKILLS_DIR at the root so resolveSkillsRoot finds all skills.
	t.Setenv("ASURA_SKILLS_DIR", skillsRoot)
	return tempHome, skillsRoot
}

func assertSymlink(t *testing.T, linkPath, expectedTarget string) {
	t.Helper()
	fi, err := os.Lstat(linkPath)
	if err != nil {
		t.Fatalf("expected symlink at %s: %v", linkPath, err)
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("%s is not a symlink", linkPath)
	}
	target, err := os.Readlink(linkPath)
	if err != nil {
		t.Fatalf("failed to read symlink %s: %v", linkPath, err)
	}
	if filepath.Clean(target) != filepath.Clean(expectedTarget) {
		t.Fatalf("symlink %s points to %q, want %q", linkPath, target, expectedTarget)
	}
}

// assertAllSkillsWired checks that every skill in skillNames is a symlink
// under the agent's skills root pointing into skillsRoot.
func assertAllSkillsWired(t *testing.T, agentRoot string, skillsRoot string) {
	t.Helper()
	for _, name := range skillNames {
		link := filepath.Join(agentRoot, name)
		assertSymlink(t, link, filepath.Join(skillsRoot, name))
	}
}

func TestWireAllAgents(t *testing.T) {
	h, root := setupTestEnv(t)

	n, err := Wire("")
	if err != nil {
		t.Fatalf("Wire(\"\") failed: %v", err)
	}
	if n != 4 {
		t.Fatalf("Wire(\"\") wired %d agents, want 4", n)
	}

	assertAllSkillsWired(t, filepath.Join(h, ".config", "opencode", "skills"), root)
	assertAllSkillsWired(t, filepath.Join(h, ".pi", "agent", "skills"), root)
	assertAllSkillsWired(t, filepath.Join(h, ".claude", "skills"), root)
	assertAllSkillsWired(t, filepath.Join(h, ".gemini", "skills"), root)

	// Test idempotency: wiring again must succeed without error.
	n2, err := Wire("")
	if err != nil {
		t.Fatalf("second Wire(\"\") failed: %v", err)
	}
	if n2 != 4 {
		t.Fatalf("second Wire(\"\") returned %d, want 4", n2)
	}
	assertAllSkillsWired(t, filepath.Join(h, ".claude", "skills"), root)
}

func TestWireSingleAgent(t *testing.T) {
	h, root := setupTestEnv(t)

	n, err := Wire("claude")
	if err != nil {
		t.Fatalf("Wire(\"claude\") failed: %v", err)
	}
	if n != 1 {
		t.Fatalf("Wire(\"claude\") wired %d, want 1", n)
	}

	assertAllSkillsWired(t, filepath.Join(h, ".claude", "skills"), root)

	// Other agents should not be wired yet.
	for _, name := range skillNames {
		p := filepath.Join(h, ".config", "opencode", "skills", name)
		if _, err := os.Lstat(p); !os.IsNotExist(err) {
			t.Fatalf("opencode skill %s should not exist yet", name)
		}
	}
}

func TestUnwireAgents(t *testing.T) {
	h, _ := setupTestEnv(t)

	if _, err := Wire(""); err != nil {
		t.Fatalf("initial Wire failed: %v", err)
	}

	// Unwire single agent.
	n, err := Unwire("opencode")
	if err != nil {
		t.Fatalf("Unwire(\"opencode\") failed: %v", err)
	}
	if n != 1 {
		t.Fatalf("Unwire(\"opencode\") unwired %d agents, want 1", n)
	}

	// All opencode skill links must be gone.
	for _, name := range skillNames {
		p := filepath.Join(h, ".config", "opencode", "skills", name)
		if _, err := os.Lstat(p); !os.IsNotExist(err) {
			t.Fatalf("opencode skill %s was not removed", name)
		}
	}

	// Claude should still have all skills wired.
	for _, name := range skillNames {
		p := filepath.Join(h, ".claude", "skills", name)
		if _, err := os.Lstat(p); err != nil {
			t.Fatalf("claude skill %s should still exist: %v", name, err)
		}
	}

	// Unwire all remaining.
	if _, err := Unwire(""); err != nil {
		t.Fatalf("Unwire(\"\") failed: %v", err)
	}

	for _, d := range SupportedAgents() {
		for _, name := range skillNames {
			p := skillPathFor(d, name)
			if _, err := os.Lstat(p); !os.IsNotExist(err) {
				t.Fatalf("skill %s for %s should be removed", name, d.ID)
			}
		}
	}
}

func TestSymlinkSafetyWithForeignLink(t *testing.T) {
	h, root := setupTestEnv(t)

	// Create a foreign directory and symlink for the "toren" skill of claude.
	foreignDir := t.TempDir()
	claudeSkill := filepath.Join(h, ".claude", "skills", "toren")
	if err := os.MkdirAll(filepath.Dir(claudeSkill), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(foreignDir, claudeSkill); err != nil {
		t.Fatal(err)
	}

	// Unwire should NOT remove the foreign link (it's not a valid skill dir).
	if _, err := Unwire("claude"); err != nil {
		t.Fatalf("Unwire: %v", err)
	}
	if _, err := os.Lstat(claudeSkill); os.IsNotExist(err) {
		t.Fatalf("foreign symlink was unexpectedly removed")
	}

	// Wire should replace the foreign link cleanly with ours.
	if _, err := Wire("claude"); err != nil {
		t.Fatalf("Wire over foreign link failed: %v", err)
	}
	assertAllSkillsWired(t, filepath.Join(h, ".claude", "skills"), root)
}

func TestSymlinkSafetyWithRealDir(t *testing.T) {
	h, _ := setupTestEnv(t)

	// Plant a real directory at the "toren" skill path for claude.
	claudeSkill := filepath.Join(h, ".claude", "skills", "toren")
	if err := os.MkdirAll(claudeSkill, 0o755); err != nil {
		t.Fatal(err)
	}

	// symlinkForce should refuse to clobber a real directory.
	_, err := Wire("claude")
	if err == nil {
		t.Fatalf("expected error when wiring over a real directory")
	}
}

func TestDetectAgents(t *testing.T) {
	h, root := setupTestEnv(t)

	agents := DetectAgents()
	if len(agents) != 4 {
		t.Fatalf("DetectAgents returned %d, want 4", len(agents))
	}

	for _, a := range agents {
		if a.Wired {
			t.Fatalf("agent %s should not be wired initially", a.ID)
		}
	}

	if _, err := Wire("gemini"); err != nil {
		t.Fatal(err)
	}

	// Confirm gemini links point where we expect.
	assertAllSkillsWired(t, filepath.Join(h, ".gemini", "skills"), root)

	agentsAfter := DetectAgents()
	for _, a := range agentsAfter {
		if a.ID == "gemini" && !a.Wired {
			t.Fatalf("gemini should be wired after Wire(\"gemini\")")
		}
		if a.ID != "gemini" && a.Wired {
			t.Fatalf("agent %s should not be wired", a.ID)
		}
	}
}
