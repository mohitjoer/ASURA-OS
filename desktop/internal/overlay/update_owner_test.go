package overlay

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

// `sudo toren update` writes revisions into the user's state directory; they
// must end up the user's, or the user's own update cannot remove them.
func TestRevisionsMadeAsRootGoToTheStateOwner(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("needs root to chown")
	}
	const uid, gid = 1000, 1000
	base := t.TempDir()
	state := filepath.Join(t.TempDir(), "state")
	if err := os.MkdirAll(filepath.Join(base, "hyprland"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(base, "hyprland", "misc.conf"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(state, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chown(state, uid, gid); err != nil {
		t.Fatal(err)
	}
	revision, err := recordRevision(Paths{Base: base, State: state})
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{
		filepath.Join(state, revisionsDir),
		revision.Preserved,
		filepath.Join(revision.Preserved, "hyprland", "misc.conf"),
		filepath.Join(state, revisionFile),
	} {
		info, err := os.Lstat(p)
		if err != nil {
			t.Fatal(err)
		}
		if st := info.Sys().(*syscall.Stat_t); st.Uid != uid {
			t.Errorf("%s is owned by %d, want the state owner %d", p, st.Uid, uid)
		}
	}
}
