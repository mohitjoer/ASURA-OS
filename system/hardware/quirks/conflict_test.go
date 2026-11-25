package quirks_test

import (
	"os"
	"path/filepath"
	"testing"

	"toren/hardware/quirks"
)

func resolveHardwareDirs() (modelsDir, componentsDir, repoRoot string) {
	candidates := []string{
		".",
		"..",
		"../..",
		"../../..",
		"system/hardware",
		"../system/hardware",
	}

	for _, c := range candidates {
		m := filepath.Join(c, "models")
		if info, err := os.Stat(m); err == nil && info.IsDir() {
			comp := filepath.Join(c, "components")
			root := filepath.Clean(filepath.Join(c, "..", ".."))
			if _, err := os.Stat(filepath.Join(root, "system")); err == nil {
				return m, comp, root
			}
			return m, comp, c
		}
		// Also try c/system/hardware/models
		m2 := filepath.Join(c, "system", "hardware", "models")
		if info, err := os.Stat(m2); err == nil && info.IsDir() {
			comp2 := filepath.Join(c, "system", "hardware", "components")
			return m2, comp2, c
		}
	}

	return "models", "components", "."
}

func TestConflictCheck(t *testing.T) {
	modelsDir, componentsDir, repoRoot := resolveHardwareDirs()

	defs, err := quirks.LoadQuirks(modelsDir, componentsDir)
	if err != nil {
		t.Fatalf("failed to load quirks: %v", err)
	}

	if len(defs) == 0 {
		t.Fatalf("no quirks loaded from %s and %s", modelsDir, componentsDir)
	}

	conflicts := quirks.CheckConflicts(defs, repoRoot)
	if len(conflicts) > 0 {
		for _, c := range conflicts {
			t.Errorf("Hardware quirk conflict on target %s: %s", c.Target, c.Reason)
		}
	}
}
