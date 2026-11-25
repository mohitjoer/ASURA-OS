package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"syscall"
	"time"

	"asura/desktop/internal/overlay"
	"asura/desktop/internal/runner"
)

// asuraUpdate runs the complete Void Linux system update lifecycle.
//
// Lifecycle sequence:
//  1. Snapshot ~/.config to git AND create an atomic Btrfs root snapshot
//     (/.snapshots/pre-update-<timestamp>) if / is on Btrfs.
//  2. Run system package updates via xbps-install -Su -y (routing through
//     asura-sysd socket /run/asura/sysd.sock or sudo xbps-install -Su -y).
//  3. Update installed Flatpaks via flatpak update -y (if flatpak is present).
//  4. Re-materialize the /usr/share/asura/ base overlay with ~/.config/asura/user_edits/
//     and trigger a live shell/daemon reload.
func asuraUpdate(args []string) int {
	set := newFlags("update", updateUsage)
	var (
		root       = set.String("root", "", "operate on a tree under this directory")
		dryRun     = set.Bool("dry-run", false, "do the snapshot and preserve steps, fetch nothing")
		noSnapshot = set.Bool("no-snapshot", false, "skip the config snapshot")
		unstable   = set.Bool("unstable", false, "pull the latest code from the dev-unstable branch")
		branch     = set.String("branch", "", "git branch to pull code updates from (default: dev-unstable)")
		channel    = set.String("channel", "", "channel to update from (stable, unstable, dev-unstable)")
		system     = set.Bool("system", false, "force update of all system packages (XBPS)")
		asuraOnly  = set.Bool("asura-only", false, "only update Toren packages and desktop components")
		check      = set.Bool("check", false, "only report whether a newer Toren is available")
		force      = set.Bool("force", false, "take the tip of the branch without checking CI, and reapply even when already applied")
		wait       = set.Bool("wait", false, "wait for CI still running on the newest commit instead of taking the last one that passed")
	)
	if err := set.Parse(args); err != nil {
		return exitUsage
	}

	targetBranch := ""
	if *unstable {
		targetBranch = "dev-unstable"
	} else if *branch != "" {
		targetBranch = *branch
	} else if *channel == "unstable" || *channel == "dev-unstable" {
		targetBranch = "dev-unstable"
	} else if *channel == "stable" {
		targetBranch = "main"
	} else if ch := pinnedChannel(*root); ch != "" {
		targetBranch = channelBranches[ch]
	} else {
		if ch, err := xbpsChannel(*root); err == nil {
			if ch == "unstable" || ch == "dev-unstable" {
				targetBranch = "dev-unstable"
			}
		}
	}

	updateForce = *force
	updateWait = *wait
	if *check {
		b := targetBranch
		if b == "" {
			b = "dev-unstable"
		}
		return checkForUpdate(b, *root)
	}

	paths, err := overlayPaths(*root)
	if err != nil {
		fmt.Fprintln(os.Stderr, "asura update:", err)
		return exitFail
	}

	isRoot := os.Geteuid() == 0
	updateSystemPkgs := (isRoot || *system) && !*asuraOnly

	var snapshot func(overlay.Paths) error
	if !*noSnapshot {
		snapshot = func(p overlay.Paths) error {
			if err := gitSnapshot(p); err != nil {
				return err
			}
			if !*dryRun && updateSystemPkgs {
				if snap, err := createBtrfsSnapshot(); err != nil {
					fmt.Fprintf(os.Stderr, "asura update: btrfs snapshot: %v\n", err)
				} else if snap != "" {
					fmt.Printf("created system Btrfs snapshot: %s\n", snap)
					pruneBtrfsSnapshots("/.snapshots", keepUpdateSnapshots)
				}
			}
			return nil
		}
	}

	var asuraErr error
	var fetch func() error
	if !*dryRun {
		runUserHooks("pre-update", nil)
		fetch = func() error {
			// System packages first, so ASURA is built and applied against the
			// libraries it will run with (only as root or with --system).
			if updateSystemPkgs {
				fmt.Println("==> Step 1/4: Updating system packages (XBPS)...")
				if err := voidFetch(); err != nil {
					fmt.Fprintf(os.Stderr, "asura update: system packages: %v\n", err)
				}
			} else {
				fmt.Println("==> Step 1/4: System packages: skipped (sudo asura update, or --system, updates them)")
			}
			// Then ASURA, from its branch (dev-unstable unless the channel or
			// a flag names another).
			b := targetBranch
			if b == "" {
				b = "dev-unstable"
			}
			fmt.Printf("==> Step 2/4: Updating Toren to the latest '%s'...\n", b)
			if err := pullGitUpdates(b, *root); err != nil {
				asuraErr = err
				fmt.Fprintf(os.Stderr, "asura update: Toren was not updated: %v\n", err)
			}
			// The one-time fixes this machine has not had yet, run by the
			// binary just applied, so a migration added in this update runs
			// with the code that knows about it.
			if *root == "" {
				fmt.Println("==> Running one-time fixes (migrations)...")
				if err := runUpdateMigrations(); err != nil && asuraErr == nil {
					asuraErr = err
				}
			}
			return nil
		}
	} else {
		fetch = nil
	}

	// Step 1 & 2 & 4: snapshot, package update, and re-materialization
	if err := overlay.Update(paths, snapshot, fetch); err != nil {
		fmt.Fprintln(os.Stderr, "asura update:", err)
		return exitFail
	}

	// Step 3: Flatpak updates (if flatpak is present)
	if !*dryRun {
		updateFlatpaks()
	}

	// Regenerate GRUB so the new Btrfs snapshot gets a boot entry (only on system updates)
	if !*dryRun && updateSystemPkgs && isBtrfsRoot() {
		if _, err := exec.LookPath("grub-mkconfig"); err == nil {
			fmt.Println("==> Regenerating GRUB boot menu (snapshot entries)...")
			cmd := exec.Command("grub-mkconfig", "-o", "/boot/grub/grub.cfg")
			cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "asura update: grub-mkconfig: %v\n", err)
			}
		}
	}

	status, err := overlay.StatusOf(paths)
	if err != nil {
		fmt.Fprintln(os.Stderr, "asura update:", err)
		return exitFail
	}
	fmt.Printf("updated to revision %s\n", status.Revision)
	fmt.Printf("config materialized; %d user edit(s), %d stale, %d missing\n",
		status.Edits, status.Stale, status.Missing)

	if !*dryRun {
		// The running desktop picks up what was just installed: services
		// re-synced, Hyprland reloaded, the shell and daemons restarted. No
		// logout needed.
		fmt.Println("==> Step 4/4: Refreshing the running desktop...")
		if os.Geteuid() == 0 {
			refreshAsSudoUser(os.Stdout, os.Stderr)
		} else {
			refreshMain(os.Stdout, os.Stderr, runner.Exec(20*time.Second))
		}
	}

	if asuraErr != nil {
		fmt.Fprintln(os.Stderr, "\n==> Toren update FAILED:", asuraErr)
		fmt.Fprintln(os.Stderr, "==> The desktop was refreshed, but runs the code it had before. `asura doctor` checks the system.")
		return exitFail
	}
	if !updateSystemPkgs && !*asuraOnly {
		fmt.Println("\n--------------------------------------------------------------------------------")
		fmt.Println("==> Toren packages and desktop components are up to date.")
		fmt.Println("==> Recommendation: To update all base system packages (XBPS) and OS, run:")
		fmt.Println("    sudo asura update")
		fmt.Println("--------------------------------------------------------------------------------")
	} else if *asuraOnly {
		fmt.Println("\n==> Toren update complete.")
	} else {
		fmt.Println("\n==> Full system and Toren update complete.")
		reportRebootNeeded()
	}

	if !*dryRun {
		runUserHooks("post-update", nil)
	}

	return exitOK
}

// voidFetch executes the system package update via Void Linux XBPS.
// It first attempts to route through the privileged asura-sysd socket.
// If the daemon is not active, it falls back to direct xbps-install.
func voidFetch() error {

	// Try sysd socket first
	if err := runSysdPkgUpdate(); err == nil {
		fmt.Println("system packages updated via asura-sysd")
		return nil
	}

	// Fallback to local xbps-install command
	return runLocalXbpsUpdate()
}

// runSysdPkgUpdate communicates with asura-sysd over /run/asura/sysd.sock.
func runSysdPkgUpdate() error {
	conn, err := net.DialTimeout("unix", "/run/asura/sysd.sock", 2*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()

	req := struct {
		Action string `json:"action"`
	}{
		Action: "pkg_update",
	}
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}
	if _, err := conn.Write(append(data, '\n')); err != nil {
		return err
	}

	scanner := bufio.NewScanner(conn)
	if scanner.Scan() {
		var res struct {
			Status  string `json:"status"`
			Message string `json:"message"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &res); err != nil {
			return err
		}
		if res.Status != "ok" {
			return fmt.Errorf("sysd returned error: %s", res.Message)
		}
		return nil
	}
	return fmt.Errorf("no response from asura-sysd")
}

// runLocalXbpsUpdate executes xbps-install -Su -y locally with privilege elevation.
//
// A package repository can lag behind a library soname
// bump in Void: the update then aborts with exit 8, "broken, unresolvable
// shlib", and nothing at all is updated. In that case the packages XBPS names
// as broken, and the installed library each of them needs, are held for one
// retry so everything else updates; the holds are released afterwards, so the
// next update tries them again once the repository has rebuilt.
func runLocalXbpsUpdate() error {
	elevate := func(prog string, args ...string) *exec.Cmd {
		if os.Geteuid() != 0 {
			return exec.Command("sudo", append([]string{prog}, args...)...)
		}
		return exec.Command(prog, args...)
	}
	run := func() (string, error) {
		var buf strings.Builder
		cmd := elevate("xbps-install", "-Su", "-y")
		cmd.Stdout = io.MultiWriter(os.Stdout, &buf)
		cmd.Stderr = io.MultiWriter(os.Stderr, &buf)
		cmd.Stdin = os.Stdin
		err := cmd.Run()
		return buf.String(), err
	}

	out, err := run()
	if err == nil {
		return nil
	}
	exitErr, ok := err.(*exec.ExitError)
	if !ok || exitErr.ExitCode() != 8 {
		return fmt.Errorf("xbps-install: %w", err)
	}

	broken, shlibs := parseBrokenShlibs(out)
	hold := append([]string(nil), broken...)
	for _, lib := range shlibs {
		if owner := shlibOwner(lib); owner != "" {
			hold = append(hold, owner)
		}
	}
	hold = uniqueStrings(hold)
	if len(hold) == 0 {
		return fmt.Errorf("xbps-install: unresolved shared libraries")
	}
	fmt.Fprintf(os.Stderr, "\n==> A repository has not caught up with a library update; holding back for now: %s\n", strings.Join(hold, " "))
	if err := elevate("xbps-pkgdb", append([]string{"-m", "hold"}, hold...)...).Run(); err != nil {
		return fmt.Errorf("xbps-pkgdb hold: %w", err)
	}
	_, retryErr := run()
	_ = elevate("xbps-pkgdb", append([]string{"-m", "unhold"}, hold...)...).Run()
	if retryErr != nil {
		fmt.Fprintln(os.Stderr, "==> The system packages could not be updated yet; the desktop update continues.")
		return nil
	}
	fmt.Printf("==> System packages updated; held back until their repository rebuilds: %s\n", strings.Join(hold, " "))
	return nil
}

// brokenShlibLine is xbps's "aquamarine-0.8.0_2: broken, unresolvable shlib `libdisplay-info.so.2'".
var brokenShlibLine = regexp.MustCompile("^\\s*(\\S+)-[^-\\s]+_[0-9]+: broken, unresolvable shlib [`'](\\S+?)'")

// parseBrokenShlibs returns the package names XBPS reports broken and the
// shared libraries they cannot resolve.
func parseBrokenShlibs(out string) (pkgs, libs []string) {
	for _, line := range strings.Split(out, "\n") {
		if m := brokenShlibLine.FindStringSubmatch(line); m != nil {
			pkgs = append(pkgs, m[1])
			libs = append(libs, m[2])
		}
	}
	return uniqueStrings(pkgs), uniqueStrings(libs)
}

// shlibOwner names the installed package that provides lib, the one whose
// update would remove it.
func shlibOwner(lib string) string {
	for _, dir := range []string{"/usr/lib", "/usr/lib64"} {
		out, err := exec.Command("xbps-query", "-o", filepath.Join(dir, lib)).Output()
		if err != nil {
			continue
		}
		// "libdisplay-info-0.2.0_1: /usr/lib/libdisplay-info.so.2 (...)"
		if pkgver, _, ok := strings.Cut(strings.TrimSpace(string(out)), ":"); ok {
			if name, err := exec.Command("xbps-uhelper", "getpkgname", pkgver).Output(); err == nil {
				return strings.TrimSpace(string(name))
			}
		}
	}
	return ""
}

func uniqueStrings(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}

// updateFlatpaks executes flatpak update -y if Flatpak is installed.
func updateFlatpaks() {
	if _, err := exec.LookPath("flatpak"); err != nil {
		return
	}
	fmt.Println("==> Step 3/4: Updating Flatpak applications...")
	cmd := exec.Command("flatpak", "update", "-y")
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	_ = cmd.Run()
}

// isBtrfsRoot checks if the root filesystem is Btrfs.
func isBtrfsRoot() bool {
	data, err := os.ReadFile("/proc/mounts")
	if err != nil {
		out, err := exec.Command("stat", "-f", "-c", "%T", "/").Output()
		return err == nil && strings.TrimSpace(string(out)) == "btrfs"
	}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 3 && fields[1] == "/" && fields[2] == "btrfs" {
			return true
		}
	}
	return false
}

// keepUpdateSnapshots is how many pre-update Btrfs snapshots an update keeps.
const keepUpdateSnapshots = 5

// oldUpdateSnapshots names the pre-update snapshots beyond the newest keep.
// Their names end in a sortable timestamp; snapshots made any other way (by
// hand, or a rollback's kept root) are never selected.
func oldUpdateSnapshots(names []string, keep int) []string {
	return oldSnapshots(names, snapshotPreUpdate, keep)
}

// oldSnapshots names the snapshots with prefix beyond the newest keep.
func oldSnapshots(names []string, prefix string, keep int) []string {
	var matching []string
	for _, name := range names {
		if strings.HasPrefix(name, prefix) {
			matching = append(matching, name)
		}
	}
	sort.Strings(matching)
	if len(matching) <= keep {
		return nil
	}
	return matching[:len(matching)-keep]
}

// pruneBtrfsSnapshots deletes, for each automatic kind (before an update,
// before an install), the snapshots beyond the newest keep. Each one pins
// every file changed since, so they grow without limit when nothing removes
// them. Snapshots made by hand are never removed.
func pruneBtrfsSnapshots(dir string, keep int) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	var names []string
	for _, entry := range entries {
		if entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	for _, prefix := range []string{snapshotPreUpdate, snapshotPreInstall} {
		for _, name := range oldSnapshots(names, prefix, keep) {
			args := []string{"btrfs", "subvolume", "delete", filepath.Join(dir, name)}
			if os.Geteuid() != 0 {
				args = append([]string{"sudo"}, args...)
			}
			if out, err := exec.Command(args[0], args[1:]...).CombinedOutput(); err != nil {
				fmt.Fprintf(os.Stderr, "asura: removing old snapshot %s: %v (%s)\n", name, err, strings.TrimSpace(string(out)))
				continue
			}
			fmt.Printf("removed old system snapshot: %s\n", name)
		}
	}
}

// createBtrfsSnapshot snapshots root before an update.
func createBtrfsSnapshot() (string, error) {
	return createBtrfsSnapshotAs(snapshotPreUpdate)
}

// createBtrfsSnapshotAs makes a read-only snapshot of root in /.snapshots,
// named prefix plus a sortable timestamp. It does nothing (and returns "")
// when root is not Btrfs or btrfs-progs is missing.
func createBtrfsSnapshotAs(prefix string) (string, error) {
	if !isBtrfsRoot() {
		return "", nil
	}
	if _, err := exec.LookPath("btrfs"); err != nil {
		return "", nil
	}

	snapDir := "/.snapshots"
	timestamp := time.Now().Format("20060102-150405")
	snapName := prefix + timestamp
	snapPath := filepath.Join(snapDir, snapName)

	sudo := "sudo"
	if os.Geteuid() == 0 {
		sudo = ""
	}

	// Ensure /.snapshots directory exists
	var mkdirCmd *exec.Cmd
	if sudo != "" {
		mkdirCmd = exec.Command(sudo, "mkdir", "-p", snapDir)
	} else {
		mkdirCmd = exec.Command("mkdir", "-p", snapDir)
	}
	_ = mkdirCmd.Run()

	var snapCmd *exec.Cmd
	if sudo != "" {
		snapCmd = exec.Command(sudo, "btrfs", "subvolume", "snapshot", "-r", "/", snapPath)
	} else {
		snapCmd = exec.Command("btrfs", "subvolume", "snapshot", "-r", "/", snapPath)
	}

	out, err := snapCmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s (%s)", err, strings.TrimSpace(string(out)))
	}
	return snapPath, nil
}

const defaultGitRepoURL = "https://github.com/mohitjoer/Toren-Linux.git"

// appliedCommitFile records the commit installation/apply-repo.sh last
// applied to this system, so an update knows whether there is anything new.
var appliedCommitFile = "/usr/share/asura/applied-commit"

// appliedSourceFile records the checkout that commit was applied from.
var appliedSourceFile = "/usr/share/asura/applied-source"

// updateForce reapplies the checkout even when it is the applied commit.
var updateForce bool

func appliedCommit(root string) string {
	b, err := os.ReadFile(filepath.Join(root, appliedCommitFile))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

func gitOut(repoDir string, args ...string) string {
	out, err := repoGit(repoDir, args...).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// isUpdateCache: a checkout asura update cloned for itself, which it may reset
// to the branch. A developer's checkout is only ever fast-forwarded, so local
// work in it is never lost.
func isUpdateCache(repoDir string) bool {
	return strings.HasSuffix(repoDir, filepath.Join(".cache", "asura", "source")) ||
		repoDir == "/var/cache/asura/source"
}

// runUpdateMigrations runs `asura migrate` from the installed binary: system
// migrations as root (through sudo), then the user's (as SUDO_USER when the
// update runs under sudo).
func runUpdateMigrations() error {
	bin := "/usr/bin/asura"
	if _, err := os.Stat(bin); err != nil {
		bin = selfPath()
	}
	if code := runSelf([]string{bin, "migrate"}); code != exitOK {
		return fmt.Errorf("a migration failed (asura migrate --list shows which); it runs again on the next update")
	}
	return nil
}

// followBranch fetches branch and picks the commit to apply: the newest one
// that passed CI (the tip with --force). It returns the target, or "" when
// there is nothing safe to take, with the reason printed.
func followBranch(repoDir, branch string) (string, error) {
	ref, err := fetchChannel(repoDir, branch)
	if err != nil {
		return "", err
	}
	if updateForce {
		fmt.Println("    --force: taking the tip of " + branch + " without checking CI")
		return gitOut(repoDir, "rev-parse", ref), nil
	}
	repo, ok := githubRepoOf(repoDir)
	if !ok {
		fmt.Println("    origin is not on GitHub, so its CI cannot be read; taking the tip of " + branch)
		return gitOut(repoDir, "rev-parse", ref), nil
	}
	target, note, err := gatedTarget(repoDir, repo, ref)
	if err != nil {
		// Without the gate nothing may be taken: an unchecked tip is exactly
		// what the gate keeps off machines.
		fmt.Printf("    could not read CI results from GitHub (%v)\n", err)
		fmt.Println("    keeping the installed Toren; `asura update --force` takes the tip unchecked")
		return "", nil
	}
	if note != "" {
		fmt.Println("    " + note)
	}
	if target == "" {
		fmt.Println("    keeping the installed Toren")
	}
	return target, nil
}

// pullGitUpdates brings ASURA to the newest commit of branch that passed CI:
// the checkout is synced to it and, unless that is the applied commit (or
// older than it), what changed is listed and the checkout is applied.
func pullGitUpdates(branch string, root string) error {
	repoDir, err := findOrCloneRepo(branch)
	if err != nil {
		return fmt.Errorf("locating Asura repository: %w", err)
	}
	fmt.Printf("    source: %s (%s)\n", repoDir, branch)
	target, err := followBranch(repoDir, branch)
	if err != nil || target == "" {
		return err
	}
	applied := appliedCommit(root)
	if applied != "" && !updateForce && gitOut(repoDir, "cat-file", "-t", applied) == "commit" {
		if gitOut(repoDir, "rev-parse", target) == gitOut(repoDir, "rev-parse", applied) {
			fmt.Printf("    Toren is already at the latest checked commit %s\n", short(applied))
			return nil
		}
		if isAncestorIn(repoDir, target, applied) {
			fmt.Printf("    the installed Toren (%s) is newer than %s; nothing was applied\n", short(applied), short(target))
			return nil
		}
	}
	head, note, err := syncToTarget(repoDir, branch, target)
	if err != nil {
		return err
	}
	if note != "" {
		fmt.Println("    " + note)
	}
	if head == applied && !updateForce {
		fmt.Printf("    Toren is already at %s\n", short(head))
		return nil
	}
	if applied != "" && applied != head {
		if log := gitOut(repoDir, "log", "--oneline", "--no-decorate", "-n", "25", applied+".."+head); log != "" {
			fmt.Printf("    new since %s:\n", short(applied))
			for _, l := range strings.Split(log, "\n") {
				fmt.Println("      " + l)
			}
		}
	}
	fmt.Printf("    applying %s\n", short(head))
	return applyRepo(repoDir, root)
}

// checkForUpdate reports whether branch has a commit that passed CI and is
// newer than the applied one, fetching but changing nothing.
func checkForUpdate(branch, root string) int {
	repoDir, err := findOrCloneRepo(branch)
	if err != nil {
		fmt.Fprintln(os.Stderr, "asura update --check:", err)
		return exitFail
	}
	target, err := followBranch(repoDir, branch)
	if err != nil {
		fmt.Fprintln(os.Stderr, "asura update --check:", err)
		return exitFail
	}
	if target == "" {
		return exitFail
	}
	applied := appliedCommit(root)
	switch {
	case applied == target:
		fmt.Printf("up to date: %s (%s)\n", short(target), branch)
		return exitOK
	case applied != "" && gitOut(repoDir, "cat-file", "-t", applied) == "commit" && isAncestorIn(repoDir, target, applied):
		fmt.Printf("up to date: %s is newer than the last checked commit %s (%s)\n", short(applied), short(target), branch)
		return exitOK
	default:
		fmt.Printf("update available: %s -> %s (%s)\n", orUnknown(short(applied)), short(target), branch)
		if applied != "" {
			if log := gitOut(repoDir, "log", "--oneline", "--no-decorate", "-n", "25", applied+".."+target); log != "" {
				fmt.Println(log)
			}
		}
		return exitNoMatch
	}
}

func short(sha string) string {
	if len(sha) > 8 {
		return sha[:8]
	}
	return sha
}

func orUnknown(s string) string {
	if s == "" {
		return "unknown"
	}
	return s
}

// applyRepoArgs is the command that applies repoDir to root. A package var so
// the update tests can see what would run without running it.
var applyRepoArgs = func(repoDir, root string) []string {
	args := []string{"sh", filepath.Join(repoDir, "installation", "apply-repo.sh"), repoDir}
	if root != "" && root != "/" {
		args = append(args, root)
	}
	if os.Geteuid() != 0 {
		args = append([]string{"sudo"}, args...)
	}
	return args
}

func applyRepo(repoDir, root string) error {
	script := filepath.Join(repoDir, "installation", "apply-repo.sh")
	if _, err := os.Stat(script); err != nil {
		return fmt.Errorf("%s is missing; the checkout is older than this asura", script)
	}
	args := applyRepoArgs(repoDir, root)
	fmt.Println("==> Applying the updated repository to this system...")
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout, cmd.Stderr, cmd.Stdin = os.Stdout, os.Stderr, os.Stdin
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("apply-repo: %w", err)
	}
	return nil
}

func findOrCloneRepo(branch string) (string, error) {
	// 1. Check ASURA_SRC_DIR environment variable override
	if env := os.Getenv("ASURA_SRC_DIR"); env != "" && isAsuraRepo(env) {
		return env, nil
	}

	// 1b. The checkout this system was last applied from (apply-repo records it).
	if b, err := os.ReadFile(appliedSourceFile); err == nil {
		if p := strings.TrimSpace(string(b)); p != "" && isAsuraRepo(p) && canWriteRepo(p) {
			return p, nil
		}
	}

	// 2. Check current directory or git root
	if out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output(); err == nil {
		top := strings.TrimSpace(string(out))
		if isAsuraRepo(top) {
			return top, nil
		}
	}

	// 3. Check well-known local paths
	home := os.Getenv("HOME")
	candidates := []string{
		filepath.Join(home, "code", "Toren-Linux"),
		filepath.Join(home, "code", "asura", "Toren-Linux"),
		filepath.Join(home, "Toren-Linux"),
		filepath.Join(home, ".cache", "asura", "source"),
		"/usr/src/asura",
		"/usr/share/asura",
		"/var/cache/asura/source",
	}
	if sudoUser := os.Getenv("SUDO_USER"); sudoUser != "" && sudoUser != "root" {
		userHome := filepath.Join("/home", sudoUser)
		candidates = append([]string{
			filepath.Join(userHome, "code", "Toren-Linux"),
			filepath.Join(userHome, "code", "asura", "Toren-Linux"),
			filepath.Join(userHome, "Toren-Linux"),
		}, candidates...)
	}
	// A checkout this user cannot write (a root-owned clone left by an
	// earlier `sudo asura update`) cannot be pulled into: skip it, and clone
	// a copy of the user's own below.
	for _, c := range candidates {
		if isAsuraRepo(c) && canWriteRepo(c) {
			return c, nil
		}
	}

	// 4. Clone to cache directory if not found locally. Under sudo the clone
	// is the invoking user's, in their cache, made as them: a root-owned clone
	// is one the user's own `asura update` can never pull into.
	cacheDir := filepath.Join(home, ".cache", "asura", "source")
	cloneAs := ""
	if os.Geteuid() == 0 {
		cacheDir = "/var/cache/asura/source"
		if u := os.Getenv("SUDO_USER"); u != "" && u != "root" {
			cacheDir = filepath.Join("/home", u, ".cache", "asura", "source")
			cloneAs = u
		}
	}
	if _, err := os.Stat(filepath.Join(cacheDir, ".git")); os.IsNotExist(err) {
		fmt.Printf("==> Cloning Toren Linux repository (%s) into %s...\n", branch, cacheDir)
		args := []string{"sh", "-c", `mkdir -p "$(dirname "$1")" && git clone --depth 1 --branch "$2" "$3" "$1"`, "sh", cacheDir, branch, defaultGitRepoURL}
		if cloneAs != "" {
			args = append([]string{"sudo", "-u", cloneAs}, args...)
		}
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("git clone failed: %w", err)
		}
		return cacheDir, nil
	}

	return cacheDir, nil
}

// repoGit runs git in repoDir as the checkout's owner. `sudo asura update`
// otherwise writes root-owned objects into the user's own checkout, after which
// the user's `asura update` fails with "Permission denied".
func repoGit(repoDir string, args ...string) *exec.Cmd {
	full := append([]string{"-c", "safe.directory=*", "-C", repoDir}, args...)
	if os.Geteuid() == 0 {
		if info, err := os.Stat(repoDir); err == nil {
			if st, ok := info.Sys().(*syscall.Stat_t); ok && st.Uid != 0 {
				return exec.Command("sudo", append([]string{"-u", fmt.Sprintf("#%d", st.Uid), "git"}, full...)...)
			}
		}
	}
	return exec.Command("git", full...)
}

// canWriteRepo reports whether git can update the checkout at path.
func canWriteRepo(path string) bool {
	return syscall.Access(filepath.Join(path, ".git"), 2 /* W_OK */) == nil
}

func isAsuraRepo(path string) bool {
	if path == "" {
		return false
	}
	if _, err := os.Stat(filepath.Join(path, ".git")); err != nil {
		return false
	}
	if _, err := os.Stat(filepath.Join(path, "shell", "hub")); err == nil {
		return true
	}
	return false
}
