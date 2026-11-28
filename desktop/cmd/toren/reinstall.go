package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func init() {
	register("reinstall", "re-apply Toren files for currently installed commit", torenReinstall)
}

// reinstallPlan returns the command that will be executed to apply the commit.
func reinstallPlan(repoDir, _, root string) []string {
	return applyRepoArgs(repoDir, root)
}

// reinstallApply runs the apply step. Stubbed in tests.
var reinstallApply = applyRepo

// torenReinstall re-applies shipped files for the currently applied commit.
func torenReinstall(args []string) int {
	set := newFlags("reinstall", reinstallUsage)
	var root = set.String("root", "", "operate on a tree under this directory")
	if _, err := parseFlags(set, args); err != nil {
		return exitUsage
	}

	commit := appliedCommit(*root)
	if commit == "" {
		if *root == "" {
			if out := gitOut(".", "rev-parse", "HEAD"); out != "" {
				commit = out
			}
		}
		if commit == "" {
			fmt.Fprintln(os.Stderr, "toren reinstall: no applied commit recorded in /usr/share/toren/applied-commit")
			return exitFail
		}
	}

	repoDir, err := findOrCloneRepo("dev-unstable")
	if err != nil {
		// If running in development checkout, try using the current repository
		if _, errStat := os.Stat("installation/apply-repo.sh"); errStat == nil {
			repoDir, _ = filepath.Abs(".")
		} else {
			fmt.Fprintf(os.Stderr, "toren reinstall: %v\n", err)
			return exitFail
		}
	}

	// Verify the commit is available in repoDir, fetch if not
	if gitOut(repoDir, "cat-file", "-t", commit) != "commit" {
		fmt.Printf("    fetching commit %s from origin...\n", short(commit))
		if err := repoGit(repoDir, "fetch", "origin", commit).Run(); err != nil {
			_ = repoGit(repoDir, "fetch", "origin").Run()
		}
		if gitOut(repoDir, "cat-file", "-t", commit) != "commit" {
			fmt.Fprintf(os.Stderr, "toren reinstall: commit %s not found in repository\n", commit)
			return exitFail
		}
	}

	// Detach to the exact applied commit so nothing newer is applied
	if gitOut(repoDir, "rev-parse", "HEAD") != commit {
		if err := repoGit(repoDir, "checkout", "--quiet", "--detach", commit).Run(); err != nil {
			fmt.Fprintf(os.Stderr, "toren reinstall: checkout commit %s: %v\n", short(commit), err)
			return exitFail
		}
	}

	fmt.Printf("==> Reinstalling Toren files at commit %s\n", short(commit))
	if err := reinstallApply(repoDir, *root); err != nil {
		fmt.Fprintf(os.Stderr, "toren reinstall: apply-repo failed: %v\n", err)
		return exitFail
	}

	fmt.Printf("reinstalled Toren files at commit %s\n", short(commit))
	return exitOK
}
