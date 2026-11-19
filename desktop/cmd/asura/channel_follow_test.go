package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCIVerdict(t *testing.T) {
	type run struct{ status, conclusion string }
	cases := []struct {
		name string
		runs []run
		want ciState
	}{
		{"none", nil, ciUnchecked},
		{"all passed", []run{{"completed", "success"}, {"completed", "skipped"}, {"completed", "neutral"}}, ciPassed},
		{"one running", []run{{"completed", "success"}, {"in_progress", ""}}, ciRunning},
		{"one failed", []run{{"completed", "success"}, {"completed", "failure"}}, ciFailed},
		{"failed beats running", []run{{"queued", ""}, {"completed", "cancelled"}}, ciFailed},
		{"timed out", []run{{"completed", "timed_out"}}, ciFailed},
	}
	for _, tc := range cases {
		got := ciVerdict(len(tc.runs), func(yield func(string, string)) {
			for _, r := range tc.runs {
				yield(r.status, r.conclusion)
			}
		})
		if got != tc.want {
			t.Errorf("%s: %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestGithubRepoOf(t *testing.T) {
	dir := t.TempDir()
	git(t, dir, "init", "-q")
	for url, want := range map[string]string{
		"https://github.com/mohitjoer/ASURA-OS.git": "mohitjoer/ASURA-OS",
		"https://github.com/mohitjoer/ASURA-OS":     "mohitjoer/ASURA-OS",
		"git@github.com:contibuter001/ASURA-OS.git": "contibuter001/ASURA-OS",
	} {
		git(t, dir, "remote", "remove", "origin")
		git(t, dir, "remote", "add", "origin", url)
		if got, ok := githubRepoOf(dir); !ok || got != want {
			t.Errorf("%s: %q %v, want %q", url, got, ok, want)
		}
	}
	git(t, dir, "remote", "set-url", "origin", "https://gitlab.com/x/y.git")
	if _, ok := githubRepoOf(dir); ok {
		t.Error("a non-GitHub origin was read as GitHub")
	}
}

// A branch of three commits, a clone of it, and a fake GitHub answering
// check-runs per commit from the states map.
func gateFixture(t *testing.T, states map[int]string) (clone string, commits []string, base string) {
	t.Helper()
	origin := t.TempDir()
	git(t, origin, "init", "-q", "-b", "dev-unstable")
	for i := 0; i < 3; i++ {
		os.WriteFile(filepath.Join(origin, "f"), fmt.Append(nil, i), 0o644)
		git(t, origin, "add", "f")
		git(t, origin, "-c", "user.name=t", "-c", "user.email=t@t", "commit", "-q", "-m", fmt.Sprint("c", i))
		commits = append(commits, strings.TrimSpace(git(t, origin, "rev-parse", "HEAD")))
	}
	clone = filepath.Join(t.TempDir(), "src")
	git(t, "", "clone", "-q", "-b", "dev-unstable", origin, clone)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for i, sha := range commits {
			if strings.Contains(r.URL.Path, "/commits/"+sha+"/check-runs") {
				switch states[i] {
				case "passed":
					fmt.Fprint(w, `{"total_count":2,"check_runs":[{"status":"completed","conclusion":"success"},{"status":"completed","conclusion":"skipped"}]}`)
				case "running":
					fmt.Fprint(w, `{"total_count":1,"check_runs":[{"status":"in_progress","conclusion":null}]}`)
				case "failed":
					fmt.Fprint(w, `{"total_count":1,"check_runs":[{"status":"completed","conclusion":"failure"}]}`)
				case "down":
					http.Error(w, "rate limited", http.StatusForbidden)
				default:
					fmt.Fprint(w, `{"total_count":0,"check_runs":[]}`)
				}
				return
			}
		}
		http.NotFound(w, r)
	}))
	t.Cleanup(srv.Close)
	return clone, commits, srv.URL
}

func TestGatedTarget(t *testing.T) {
	cases := []struct {
		name      string
		states    map[int]string // commit index (0 oldest) -> state
		want      int            // index of the expected target, -1 for none
		wantNote  string
		wantError bool
	}{
		{"tip passed", map[int]string{2: "passed", 1: "passed"}, 2, "", false},
		{"tip running", map[int]string{2: "running", 1: "passed"}, 1, "still running", false},
		{"tip failed", map[int]string{2: "failed", 1: "failed", 0: "passed"}, 0, "failed CI", false},
		{"tip unchecked", map[int]string{1: "passed"}, 1, "no CI result", false},
		{"nothing passed", map[int]string{2: "failed", 1: "failed", 0: "failed"}, -1, "none of the last", false},
		{"GitHub unreachable", map[int]string{2: "down"}, -1, "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			clone, commits, base := gateFixture(t, tc.states)
			old := ciAPIBase
			ciAPIBase = base
			t.Cleanup(func() { ciAPIBase = old })
			target, note, err := gatedTarget(clone, "o/r", "refs/remotes/origin/dev-unstable")
			if tc.wantError {
				if err == nil {
					t.Fatalf("want an error, got %q %q", target, note)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			want := ""
			if tc.want >= 0 {
				want = commits[tc.want]
			}
			if target != want {
				t.Errorf("target %q, want %q", short(target), short(want))
			}
			if !strings.Contains(note, tc.wantNote) {
				t.Errorf("note %q lacks %q", note, tc.wantNote)
			}
		})
	}
}

// A developer's checkout is fast-forwarded, never reset; one with local edits
// or commits of its own stays where it is.
func TestSyncToTargetKeepsLocalWork(t *testing.T) {
	clone, commits, _ := gateFixture(t, nil)
	git(t, clone, "reset", "-q", "--hard", commits[0])

	head, note, err := syncToTarget(clone, "dev-unstable", commits[1])
	if err != nil || head != commits[1] || note != "" {
		t.Fatalf("fast-forward: %q %q %v", short(head), note, err)
	}

	os.WriteFile(filepath.Join(clone, "f"), []byte("local edit"), 0o644)
	head, note, _ = syncToTarget(clone, "dev-unstable", commits[2])
	if head != commits[1] || !strings.Contains(note, "uncommitted changes") {
		t.Fatalf("dirty: %q %q", short(head), note)
	}
	git(t, clone, "-c", "user.name=t", "-c", "user.email=t@t", "commit", "-qam", "mine")
	head, note, _ = syncToTarget(clone, "dev-unstable", commits[2])
	if !strings.Contains(note, "commits of its own") {
		t.Fatalf("diverged: %q %q", short(head), note)
	}
	if b, _ := os.ReadFile(filepath.Join(clone, "f")); string(b) != "local edit" {
		t.Error("local work was lost")
	}
}

// Offline: an earlier fetch on disk is used rather than failing the update.
func TestFetchChannelOffline(t *testing.T) {
	clone, commits, _ := gateFixture(t, nil)
	ref, err := fetchChannel(clone, "dev-unstable")
	if err != nil || gitOut(clone, "rev-parse", ref) != commits[2] {
		t.Fatalf("online fetch: %q %v", ref, err)
	}
	git(t, clone, "remote", "set-url", "origin", filepath.Join(t.TempDir(), "gone"))
	if ref, err := fetchChannel(clone, "dev-unstable"); err != nil || ref == "" {
		t.Fatalf("offline with an earlier fetch: %q %v", ref, err)
	}
	if _, err := fetchChannel(clone, "no-such-branch"); err == nil {
		t.Fatal("offline with nothing fetched before must be an error")
	}
}

func git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	out, err := cmd.CombinedOutput()
	if err != nil && !(len(args) > 1 && args[0] == "remote" && args[1] == "remove") {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return string(out)
}
