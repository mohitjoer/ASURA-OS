package quirks_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"toren/hardware/probe"
	"toren/hardware/quirks"
	"toren/hardware/redact"
)

func resolveMachinesDir() string {
	candidates := []string{
		"../testdata/machines",
		"testdata/machines",
		"system/hardware/testdata/machines",
		"../system/hardware/testdata/machines",
		"../../system/hardware/testdata/machines",
	}
	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && info.IsDir() {
			return c
		}
	}
	return "../testdata/machines"
}

func TestMachineFixtures(t *testing.T) {
	machinesDir := resolveMachinesDir()
	modelsDir, componentsDir, _ := resolveHardwareDirs()

	defs, err := quirks.LoadQuirks(modelsDir, componentsDir)
	if err != nil {
		t.Fatalf("failed to load quirks: %v", err)
	}

	entries, err := os.ReadDir(machinesDir)
	if err != nil {
		t.Fatalf("failed to read machines dir %s: %v", machinesDir, err)
	}

	if len(entries) == 0 {
		t.Fatalf("no machine fixtures found in %s", machinesDir)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		mID := entry.Name()
		mDir := filepath.Join(machinesDir, mID)

		t.Run(mID, func(t *testing.T) {
			// Read probe data
			fs, err := probe.New(mDir)
			if err != nil {
				t.Fatalf("probe.New failed: %v", err)
			}

			// Redaction check: ensure no leaked real serials in DMI
			dmiDir := filepath.Join(mDir, "class", "dmi", "id")
			if dmiEntries, err := os.ReadDir(dmiDir); err == nil {
				rawDMI := make(map[string]string)
				for _, de := range dmiEntries {
					if de.IsDir() {
						continue
					}
					content, _ := os.ReadFile(filepath.Join(dmiDir, de.Name()))
					rawDMI[de.Name()] = string(content)
				}
				findings := redact.Verify([]byte(rawDMI["product_name"]), redact.Context{
					RawDMI: rawDMI,
				})
				for _, f := range findings {
					t.Errorf("fixture %s failed redaction verify: %s (%s)", mID, f.Rule, f.Detail)
				}
			}

			// Expected layers
			expectedFile := filepath.Join(mDir, "expected_layers.json")
			expectedData, err := os.ReadFile(expectedFile)
			if err != nil {
				t.Fatalf("reading expected_layers.json: %v", err)
			}
			var expectedIDs []string
			if err := json.Unmarshal(expectedData, &expectedIDs); err != nil {
				t.Fatalf("parsing expected_layers.json: %v", err)
			}

			// Match
			matched, err := quirks.MatchMachine(defs, fs)
			if err != nil {
				t.Fatalf("MatchMachine failed: %v", err)
			}

			var gotIDs []string
			for _, m := range matched {
				gotIDs = append(gotIDs, m.Def.ID)
			}

			if !reflect.DeepEqual(gotIDs, expectedIDs) {
				t.Errorf("fixture %s:\n got matched layers: %v\nwant matched layers: %v", mID, gotIDs, expectedIDs)
			}
		})
	}
}
