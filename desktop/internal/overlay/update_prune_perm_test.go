package overlay

import (
	"os"
	"path/filepath"
	"testing"
)

// An old revision the update cannot remove (root's, made by an earlier
// `sudo asura update`) is skipped with a hint; it never fails the update.
// Run as a normal user against ASURA_TEST_PRUNE_STATE, a state directory
// whose oldest revision root owns (set up by the caller).
func TestPruneSkipsARevisionItCannotRemove(t *testing.T) {
	state := os.Getenv("ASURA_TEST_PRUNE_STATE")
	if state == "" || os.Geteuid() == 0 {
		t.Skip("needs ASURA_TEST_PRUNE_STATE and a non-root user")
	}
	if err := pruneRevisions(Paths{State: state}, 1); err != nil {
		t.Fatalf("prune failed the update: %v", err)
	}
	if _, err := os.Stat(filepath.Join(state, revisionsDir, "20260102T000000Z")); err == nil {
		t.Error("a removable old revision was kept")
	}
}
