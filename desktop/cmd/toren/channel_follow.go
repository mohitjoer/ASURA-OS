package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

// How `toren update` follows its branch (dev-unstable by default):
//
//  1. fetch: bounded (30 s), never prompting for credentials, and an
//     unreachable remote is not a failure while an earlier fetch is on disk;
//  2. gate: the target is the newest commit on the branch whose GitHub checks
//     all passed, not the tip, so a commit that breaks CI never reaches a
//     machine (--force takes the tip unchecked, --wait waits for checks still
//     running on the tip);
//  3. never backwards: a target older than the commit already applied to this
//     system is not applied;
//  4. sync: the update's own clone is reset to the target; a developer's
//     checkout is only fast-forwarded, and local work in it is kept.

var (
	fetchTimeout   = 30 * time.Second
	ciAPIBase      = "https://api.github.com"
	ciHTTP         = &http.Client{Timeout: 10 * time.Second}
	ciPollEvery    = 20 * time.Second
	ciWaitFor      = 25 * time.Minute
	ciLookBack     = 20
	updateWait     bool
	githubRemoteRe = regexp.MustCompile(`github\.com[:/]+([^/]+)/([^/]+?)(?:\.git)?/?$`)
)

// fetchChannel brings refs/remotes/origin/BRANCH up to date and returns it.
func fetchChannel(repoDir, branch string) (string, error) {
	ref := "refs/remotes/origin/" + branch
	args := []string{"fetch", "--quiet", "--no-tags", "origin", "+refs/heads/" + branch + ":" + ref}
	// A shallow clone (the update's own) needs enough history to find the
	// last commit that passed CI and to list what changed.
	if gitOut(repoDir, "rev-parse", "--is-shallow-repository") == "true" {
		args = append(args[:1], append([]string{"--depth", "50"}, args[1:]...)...)
	}
	cmd := repoGit(repoDir, args...)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0", "GIT_ASKPASS=true")
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err == nil {
		timer := time.AfterFunc(fetchTimeout, func() { _ = cmd.Process.Kill() })
		err = cmd.Wait()
		timer.Stop()
	}
	if err != nil {
		if gitOut(repoDir, "rev-parse", "--verify", "--quiet", ref) != "" {
			fmt.Printf("    could not reach origin (%v); using the copy of %s fetched before\n", err, branch)
			return ref, nil
		}
		return "", fmt.Errorf("git fetch origin %s (is the network up?): %w", branch, err)
	}
	return ref, nil
}

// ciState is what GitHub's checks say about one commit.
type ciState int

const (
	ciUnchecked ciState = iota // no checks ran on it (e.g. inside a larger push)
	ciRunning
	ciFailed
	ciPassed
)

func (s ciState) String() string {
	return [...]string{"unchecked", "running", "failed", "passed"}[s]
}

// githubRepoOf is the owner/name of the checkout's origin on GitHub.
func githubRepoOf(repoDir string) (string, bool) {
	m := githubRemoteRe.FindStringSubmatch(gitOut(repoDir, "remote", "get-url", "origin"))
	if m == nil {
		return "", false
	}
	return m[1] + "/" + m[2], true
}

// commitChecks reads a commit's check runs (GitHub Actions and any other
// checks app). Every run must be finished; a failure, cancellation or timeout
// in any of them fails the commit.
func commitChecks(repo, sha string) (ciState, error) {
	req, err := http.NewRequest(http.MethodGet,
		fmt.Sprintf("%s/repos/%s/commits/%s/check-runs?per_page=100", ciAPIBase, repo, sha), nil)
	if err != nil {
		return ciUnchecked, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "toren-update")
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := ciHTTP.Do(req)
	if err != nil {
		return ciUnchecked, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ciUnchecked, fmt.Errorf("GitHub answered %s", resp.Status)
	}
	var body struct {
		Total int `json:"total_count"`
		Runs  []struct {
			Name       string `json:"name"`
			Status     string `json:"status"`
			Conclusion string `json:"conclusion"`
		} `json:"check_runs"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 4<<20)).Decode(&body); err != nil {
		return ciUnchecked, err
	}
	counted := 0
	for _, r := range body.Runs {
		if !publishingCheck(r.Name) {
			counted++
		}
	}
	return ciVerdict(counted, func(yield func(status, conclusion string)) {
		for _, r := range body.Runs {
			if !publishingCheck(r.Name) {
				yield(r.Status, r.Conclusion)
			}
		}
	}), nil
}

// publishingCheck reports whether a check run is ci's publishing stage (the
// "Packages" job and the jobs it calls). Publishing ASURA's package
// repository happens after the code checks passed and says nothing about
// whether the commit works; a publishing problem (GitHub Pages, a slow
// compositor build) must not hold every machine back.
func publishingCheck(name string) bool {
	return strings.HasPrefix(name, "Packages")
}

// ciVerdict folds the runs into one state.
func ciVerdict(total int, runs func(func(status, conclusion string))) ciState {
	if total == 0 {
		return ciUnchecked
	}
	state := ciPassed
	runs(func(status, conclusion string) {
		switch {
		case status != "completed":
			if state == ciPassed {
				state = ciRunning
			}
		case conclusion == "success" || conclusion == "neutral" || conclusion == "skipped":
		default: // failure, cancelled, timed_out, action_required, startup_failure, stale
			state = ciFailed
		}
	})
	return state
}

// gatedTarget is the newest commit on ref (first-parent, the branch's own
// line) whose checks passed, with a note when that is not the tip. An empty
// target means none of the last ciLookBack commits passed.
func gatedTarget(repoDir, repo, ref string) (string, string, error) {
	commits := strings.Fields(gitOut(repoDir, "rev-list", "--first-parent", "-n", fmt.Sprint(ciLookBack), ref))
	if len(commits) == 0 {
		return "", "", fmt.Errorf("%s has no commits", ref)
	}
	tipState := ciUnchecked
	for i, sha := range commits {
		state, err := commitChecks(repo, sha)
		if err != nil {
			return "", "", err
		}
		if i == 0 {
			tipState = state
			if state == ciRunning && updateWait {
				state, err = waitForChecks(repo, sha)
				if err != nil {
					return "", "", err
				}
				tipState = state
			}
		}
		if state != ciPassed {
			continue
		}
		if i == 0 {
			return sha, "", nil
		}
		tip := short(commits[0])
		switch tipState {
		case ciRunning:
			return sha, fmt.Sprintf("checks for the newest commit %s are still running; taking the last commit that passed, %s (--wait waits for them)", tip, short(sha)), nil
		case ciFailed:
			return sha, fmt.Sprintf("the newest commit %s failed CI; taking the last commit that passed, %s", tip, short(sha)), nil
		default:
			return sha, fmt.Sprintf("the newest commit %s has no CI result of its own; taking the last commit that passed, %s", tip, short(sha)), nil
		}
	}
	return "", fmt.Sprintf("none of the last %d commits on %s passed CI", len(commits), strings.TrimPrefix(ref, "refs/remotes/origin/")), nil
}

func waitForChecks(repo, sha string) (ciState, error) {
	deadline := time.Now().Add(ciWaitFor)
	for time.Now().Before(deadline) {
		fmt.Printf("    waiting for the checks on %s...\n", short(sha))
		time.Sleep(ciPollEvery)
		state, err := commitChecks(repo, sha)
		if err != nil || state != ciRunning {
			return state, err
		}
	}
	return ciRunning, nil
}

// isAncestorIn reports whether commit a is in b's history.
func isAncestorIn(repoDir, a, b string) bool {
	return repoGit(repoDir, "merge-base", "--is-ancestor", a, b).Run() == nil
}

// syncToTarget moves the checkout to target and returns the commit to apply.
// The update's own clone follows the branch exactly (reset, even after a
// force-push); a developer's checkout is only fast-forwarded, and one with
// local changes or its own commits stays where it is, with a note.
func syncToTarget(repoDir, branch, target string) (string, string, error) {
	run := func(args ...string) error {
		cmd := repoGit(repoDir, args...)
		cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
		return cmd.Run()
	}
	if isUpdateCache(repoDir) {
		if err := run("checkout", "--quiet", "-B", branch, target); err != nil {
			return "", "", fmt.Errorf("git checkout: %w", err)
		}
		if err := run("reset", "--quiet", "--hard", target); err != nil {
			return "", "", fmt.Errorf("git reset: %w", err)
		}
		return gitOut(repoDir, "rev-parse", "HEAD"), "", nil
	}
	head := gitOut(repoDir, "rev-parse", "HEAD")
	// Untracked files do not count: they never block a fast-forward that does
	// not touch them, and counting them would freeze a checkout on one commit.
	if dirty := gitOut(repoDir, "status", "--porcelain", "--untracked-files=no"); dirty != "" {
		return head, fmt.Sprintf("%s has uncommitted changes; staying on %s (commit or stash them to follow %s)", repoDir, short(head), branch), nil
	}
	if isAncestorIn(repoDir, target, "HEAD") {
		return head, "", nil
	}
	if isAncestorIn(repoDir, "HEAD", target) {
		if err := run("merge", "--ff-only", "--quiet", target); err != nil {
			return "", "", fmt.Errorf("fast-forward to %s: %w", short(target), err)
		}
		return gitOut(repoDir, "rev-parse", "HEAD"), "", nil
	}
	return head, fmt.Sprintf("%s has commits of its own that %s does not; staying on %s", repoDir, branch, short(head)), nil
}
