package overlay

import (
	"os"
	"path/filepath"
	"testing"
)

// Every update preserves a full copy of the base; old ones must go, but never
// the current revision or the one rollback restores.
func TestPruneRevisionsKeepsTheNewestAndTheRollbackTarget(t *testing.T) {
	state := t.TempDir()
	paths := Paths{State: state}
	dir := filepath.Join(state, revisionsDir)
	ids := []string{"20260101T000000Z", "20260102T000000Z", "20260103T000000Z", "20260104T000000Z", "20260105T000000Z"}
	for _, id := range ids {
		if err := os.MkdirAll(filepath.Join(dir, id, "shell"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	// The current revision points back to an old one (a rollback target).
	if err := writeRevision(paths, Revision{ID: ids[4], Previous: ids[0]}); err != nil {
		t.Fatal(err)
	}
	if err := pruneRevisions(paths, 3); err != nil {
		t.Fatal(err)
	}
	for _, id := range ids {
		_, err := os.Stat(filepath.Join(dir, id))
		kept := err == nil
		want := id != ids[1] // the three newest, plus ids[0] as the rollback target
		if kept != want {
			t.Errorf("%s kept=%v, want %v", id, kept, want)
		}
	}
}
