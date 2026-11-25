package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// `asura migrate` runs the one-time fixes an installed machine needs to get
// from an older ASURA to this one (migrations/ in the repository, installed at
// /usr/share/asura/migrations). Each is a shell script named by the Unix time
// it was written, so they run in order, whose first line after the shebang is
// an `echo` saying what it does.
//
//	system/  once per machine, as root; done ones recorded in /var/lib/asura/migrations
//	user/    once per user, as that user; recorded in ~/.local/state/asura/migrations
//
// A migration is recorded only when it exits 0, so a failed one runs again on
// the next update. A fresh install records every migration it ships as done
// (the ISO build), since it already starts in the current state.

const migrateUsage = `Usage: asura migrate [--system | --user] [--pending | --list | --mark-done]

Runs the one-time fixes this machine has not run yet: system migrations as
root (through sudo), then the user's. asura update runs it after applying ASURA.

  --system     only the machine's migrations (as root)
  --user       only this user's migrations
  --pending    list what would run, and run nothing
  --list       every migration, with done, pending or failed
  --mark-done  record every migration as done without running it (a fresh install)
`

var (
	migrationsDir        = "/usr/share/asura/migrations"
	systemMigrationState = "/var/lib/asura/migrations"
	migrationNameRe      = regexp.MustCompile(`^[0-9]{10,}\.sh$`)
)

func init() {
	register("migrate", "run the one-time fixes this machine has not had yet", asuraMigrate)
}

type migration struct {
	Name  string // "1795180000.sh"
	Path  string
	About string // the script's own one-line description
}

// migrationDirs are the installed tree and the state directory for a kind.
func migrationDirs(kind string) (string, string) {
	tree := migrationsDir
	if env := os.Getenv("ASURA_MIGRATIONS_DIR"); env != "" {
		tree = env
	}
	state := systemMigrationState
	if env := os.Getenv("ASURA_MIGRATION_STATE"); env != "" && kind == "system" {
		state = env
	}
	if kind == "user" {
		base := os.Getenv("XDG_STATE_HOME")
		if base == "" {
			base = filepath.Join(homeDir(), ".local", "state")
		}
		state = filepath.Join(base, "asura", "migrations")
	}
	return filepath.Join(tree, kind), state
}

// listMigrations returns the kind's migrations in the order they run.
func listMigrations(kind string) ([]migration, error) {
	dir, _ := migrationDirs(kind)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []migration
	for _, e := range entries {
		if e.IsDir() || !migrationNameRe.MatchString(e.Name()) {
			continue
		}
		path := filepath.Join(dir, e.Name())
		out = append(out, migration{Name: e.Name(), Path: path, About: migrationAbout(path)})
	}
	sort.Slice(out, func(i, j int) bool { return migrationTime(out[i].Name) < migrationTime(out[j].Name) })
	return out, nil
}

func migrationTime(name string) int64 {
	n, _ := strconv.ParseInt(strings.TrimSuffix(name, ".sh"), 10, 64)
	return n
}

// migrationAbout is the text of the script's leading `echo "..."`.
func migrationAbout(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if rest, ok := strings.CutPrefix(line, "echo "); ok {
			return strings.Trim(rest, `"'`)
		}
		return ""
	}
	return ""
}

func migrationDone(state, name string) bool {
	_, err := os.Stat(filepath.Join(state, name))
	return err == nil
}

func migrationFailed(state, name string) bool {
	_, err := os.Stat(filepath.Join(state, name+".failed"))
	return err == nil
}

func asuraMigrate(args []string) int {
	kinds := []string{"system", "user"}
	mode := "run"
	for _, a := range args {
		switch a {
		case "--system":
			kinds = []string{"system"}
		case "--user":
			kinds = []string{"user"}
		case "--pending":
			mode = "pending"
		case "--list":
			mode = "list"
		case "--mark-done":
			mode = "mark-done"
		case "-h", "--help":
			fmt.Print(migrateUsage)
			return exitOK
		default:
			fmt.Fprint(os.Stderr, migrateUsage)
			return exitUsage
		}
	}
	code := exitOK
	for _, kind := range kinds {
		var c int
		if mode == "list" || mode == "pending" {
			c = reportMigrations(kind, mode, os.Stdout)
		} else {
			c = migrateAs(kind, mode)
		}
		if c != exitOK {
			code = c
			if mode == "run" {
				break // a failed system migration stops before the user's
			}
		}
	}
	return code
}

// migrateAs runs (or marks) the kind as the right user: system as root, user
// as the user who ran sudo. It re-executes this binary when it is not.
func migrateAs(kind, mode string) int {
	flag := "--" + kind
	extra := []string{}
	if mode == "mark-done" {
		extra = append(extra, "--mark-done")
	}
	// Nothing to run means no sudo: an update with no new migration must not
	// ask for a password just to find that out.
	// (Under sudo the user's state is in their home, not root's: let the
	// re-run as that user decide.)
	if mode == "run" && !(kind == "user" && os.Geteuid() == 0) {
		if pending, _ := pendingMigrations(kind); pending == 0 {
			fmt.Printf("migrations: %s: nothing to run\n", kind)
			return exitOK
		}
	}
	switch {
	case kind == "system" && os.Geteuid() != 0:
		return runSelf(append([]string{"sudo", selfPath(), "migrate", flag}, extra...))
	case kind == "user" && os.Geteuid() == 0:
		name := os.Getenv("SUDO_USER")
		if name == "" || name == "root" {
			fmt.Println("migrations: user: run as root with no SUDO_USER; no user to migrate")
			return exitOK
		}
		u, err := user.Lookup(name)
		if err != nil {
			fmt.Fprintln(os.Stderr, "asura migrate:", err)
			return exitFail
		}
		return runSelf(append([]string{"sudo", "-u", name, "env", "HOME=" + u.HomeDir, selfPath(), "migrate", flag}, extra...))
	}
	if mode == "mark-done" {
		return markMigrationsDone(kind)
	}
	return runMigrations(kind, os.Stdout, os.Stderr)
}

func selfPath() string {
	if p, err := os.Executable(); err == nil {
		return p
	}
	return "asura"
}

func runSelf(argv []string) int {
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Stdout, cmd.Stderr, cmd.Stdin = os.Stdout, os.Stderr, os.Stdin
	if err := cmd.Run(); err != nil {
		if exit, ok := err.(*exec.ExitError); ok {
			return exit.ExitCode()
		}
		fmt.Fprintln(os.Stderr, "asura migrate:", err)
		return exitFail
	}
	return exitOK
}

// runMigrations runs the kind's pending migrations, in order, stopping at the
// first that fails. Output goes to out and to migrate.log beside the state.
func runMigrations(kind string, out, errOut io.Writer) int {
	list, err := listMigrations(kind)
	if err != nil {
		fmt.Fprintln(errOut, "asura migrate:", err)
		return exitFail
	}
	_, state := migrationDirs(kind)
	var pending []migration
	for _, m := range list {
		if !migrationDone(state, m.Name) {
			pending = append(pending, m)
		}
	}
	if len(pending) == 0 {
		fmt.Fprintf(out, "migrations: %s: nothing to run\n", kind)
		return exitOK
	}
	if err := os.MkdirAll(state, 0o755); err != nil {
		fmt.Fprintln(errOut, "asura migrate:", err)
		return exitFail
	}
	logPath := filepath.Join(filepath.Dir(state), "migrate.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		logFile = nil
	}
	if logFile != nil {
		defer logFile.Close()
	}
	sink := func(w io.Writer) io.Writer {
		if logFile == nil {
			return w
		}
		return io.MultiWriter(w, logFile)
	}
	for _, m := range pending {
		// The script's own first line says what it does.
		fmt.Fprintf(out, "migrations: %s: %s\n", kind, m.Name)
		if logFile != nil {
			fmt.Fprintf(logFile, "== %s %s %s\n", time.Now().UTC().Format(time.RFC3339), kind, m.Name)
		}
		cmd := exec.Command("sh", m.Path)
		cmd.Stdout, cmd.Stderr = sink(out), sink(errOut)
		if err := cmd.Run(); err != nil {
			_ = os.WriteFile(filepath.Join(state, m.Name+".failed"), []byte(err.Error()+"\n"), 0o644)
			fmt.Fprintf(errOut, "migrations: %s: %s failed (%v); it runs again on the next update. Log: %s\n", kind, m.Name, err, logPath)
			return exitFail
		}
		_ = os.Remove(filepath.Join(state, m.Name+".failed"))
		if err := os.WriteFile(filepath.Join(state, m.Name), []byte(time.Now().UTC().Format(time.RFC3339)+"\n"), 0o644); err != nil {
			fmt.Fprintln(errOut, "asura migrate:", err)
			return exitFail
		}
	}
	fmt.Fprintf(out, "migrations: %s: ran %d\n", kind, len(pending))
	return exitOK
}

// markMigrationsDone records every migration of the kind as done.
func markMigrationsDone(kind string) int {
	list, err := listMigrations(kind)
	if err != nil {
		fmt.Fprintln(os.Stderr, "asura migrate:", err)
		return exitFail
	}
	_, state := migrationDirs(kind)
	if err := os.MkdirAll(state, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, "asura migrate:", err)
		return exitFail
	}
	for _, m := range list {
		if err := os.WriteFile(filepath.Join(state, m.Name), []byte("marked done\n"), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, "asura migrate:", err)
			return exitFail
		}
	}
	fmt.Printf("migrations: %s: %d marked done\n", kind, len(list))
	return exitOK
}

// reportMigrations prints every migration (list) or only the pending ones,
// and returns exitNoMatch when some are pending.
func reportMigrations(kind, mode string, out io.Writer) int {
	list, err := listMigrations(kind)
	if err != nil {
		fmt.Fprintln(os.Stderr, "asura migrate:", err)
		return exitFail
	}
	_, state := migrationDirs(kind)
	pending := 0
	for _, m := range list {
		status := "done"
		switch {
		case migrationFailed(state, m.Name):
			status = "FAILED"
			pending++
		case !migrationDone(state, m.Name):
			status = "pending"
			pending++
		}
		if mode == "list" || status != "done" {
			fmt.Fprintf(out, "%-8s %-7s %s  %s\n", kind, status, m.Name, m.About)
		}
	}
	if mode == "pending" && pending == 0 {
		fmt.Fprintf(out, "%-8s nothing pending\n", kind)
	}
	if pending > 0 {
		return exitNoMatch
	}
	return exitOK
}

// pendingMigrations counts the kind's pending and failed migrations, for doctor.
func pendingMigrations(kind string) (pending int, failed []string) {
	list, _ := listMigrations(kind)
	_, state := migrationDirs(kind)
	for _, m := range list {
		if migrationFailed(state, m.Name) {
			failed = append(failed, m.Name)
		}
		if !migrationDone(state, m.Name) {
			pending++
		}
	}
	return pending, failed
}
