package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// migrationSandbox points the runner at a temp tree and temp state, and
// returns the tree's system directory, the system state and the user state.
func migrationSandbox(t *testing.T) (sysDir, sysState, userState string) {
	t.Helper()
	base := t.TempDir()
	tree := filepath.Join(base, "migrations")
	t.Setenv("ASURA_MIGRATIONS_DIR", tree)
	t.Setenv("ASURA_MIGRATION_STATE", filepath.Join(base, "var", "migrations"))
	t.Setenv("XDG_STATE_HOME", filepath.Join(base, "state"))
	for _, k := range []string{"system", "user"} {
		if err := os.MkdirAll(filepath.Join(tree, k), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	return filepath.Join(tree, "system"), filepath.Join(base, "var", "migrations"),
		filepath.Join(base, "state", "toren", "migrations")
}

func writeMigration(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
}

func TestMigrationsRunInNumericOrderAndIgnoreOtherNames(t *testing.T) {
	sysDir, _, _ := migrationSandbox(t)
	writeMigration(t, sysDir, "10000000000.sh", "#!/bin/sh\necho \"third\"\n")
	writeMigration(t, sysDir, "9999999999.sh", "#!/bin/sh\necho \"second\"\n")
	writeMigration(t, sysDir, "1000000000.sh", "#!/bin/sh\necho 'first'\n")
	writeMigration(t, sysDir, "123.sh", "#!/bin/sh\necho short\n")
	writeMigration(t, sysDir, "1000000001.txt", "not a migration\n")
	writeMigration(t, sysDir, ".keep", "")
	list, err := listMigrations("system")
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, m := range list {
		got = append(got, m.Name+"="+m.About)
	}
	want := "1000000000.sh=first 9999999999.sh=second 10000000000.sh=third"
	if strings.Join(got, " ") != want {
		t.Fatalf("got %v, want %s", got, want)
	}
}

func TestPendingMigrationsRunOnceAndAreRecorded(t *testing.T) {
	sysDir, sysState, _ := migrationSandbox(t)
	counter := filepath.Join(t.TempDir(), "runs")
	body := "#!/bin/sh\necho \"count a run\"\necho x >>" + counter + "\n"
	writeMigration(t, sysDir, "1000000000.sh", body)
	writeMigration(t, sysDir, "1000000001.sh", body)
	var out, errOut bytes.Buffer
	if code := runMigrations("system", &out, &errOut); code != exitOK {
		t.Fatalf("first run: %d\n%s%s", code, out.String(), errOut.String())
	}
	if !strings.Contains(out.String(), "migrations: system: 1000000000.sh\ncount a run") {
		t.Fatalf("the description was not shown:\n%s", out.String())
	}
	for _, n := range []string{"1000000000.sh", "1000000001.sh"} {
		if !migrationDone(sysState, n) {
			t.Fatalf("%s not recorded", n)
		}
	}
	out.Reset()
	if code := runMigrations("system", &out, &errOut); code != exitOK || !strings.Contains(out.String(), "nothing to run") {
		t.Fatalf("second run: %d %q", code, out.String())
	}
	if b, _ := os.ReadFile(counter); strings.Count(string(b), "x") != 2 {
		t.Fatalf("ran %d times, want 2", strings.Count(string(b), "x"))
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(sysState), "migrate.log")); err != nil {
		t.Fatalf("no log: %v", err)
	}
}

func TestAFailedMigrationStopsTheRunAndRunsAgain(t *testing.T) {
	sysDir, sysState, _ := migrationSandbox(t)
	flag := filepath.Join(t.TempDir(), "fixed")
	writeMigration(t, sysDir, "1000000000.sh", "#!/bin/sh\necho \"fails until fixed\"\n[ -e "+flag+" ]\n")
	writeMigration(t, sysDir, "1000000001.sh", "#!/bin/sh\necho \"after\"\n")
	var out, errOut bytes.Buffer
	if code := runMigrations("system", &out, &errOut); code == exitOK {
		t.Fatal("a failing migration passed")
	}
	if migrationDone(sysState, "1000000000.sh") || !migrationFailed(sysState, "1000000000.sh") {
		t.Fatal("the failed migration should be marked failed and not done")
	}
	if migrationDone(sysState, "1000000001.sh") {
		t.Fatal("the run went on past a failure")
	}
	if _, failed := pendingMigrations("system"); len(failed) != 1 {
		t.Fatalf("failed = %v", failed)
	}
	if err := os.WriteFile(flag, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if code := runMigrations("system", &out, &errOut); code != exitOK {
		t.Fatalf("the retry failed: %s", errOut.String())
	}
	if migrationFailed(sysState, "1000000000.sh") || !migrationDone(sysState, "1000000001.sh") {
		t.Fatal("the retry did not clear the failure and finish")
	}
}

func TestReportingChangesNothingAndMarkDoneRecordsAll(t *testing.T) {
	sysDir, sysState, _ := migrationSandbox(t)
	marker := filepath.Join(t.TempDir(), "ran")
	writeMigration(t, sysDir, "1000000000.sh", "#!/bin/sh\necho \"touch\"\ntouch "+marker+"\n")
	var out bytes.Buffer
	if code := reportMigrations("system", "pending", &out); code != exitNoMatch || !strings.Contains(out.String(), "pending") {
		t.Fatalf("--pending: %d %q", code, out.String())
	}
	out.Reset()
	if code := reportMigrations("system", "list", &out); code != exitNoMatch {
		t.Fatalf("--list: %d", code)
	}
	if _, err := os.Stat(marker); err == nil || migrationDone(sysState, "1000000000.sh") {
		t.Fatal("reporting ran or recorded a migration")
	}
	if code := markMigrationsDone("system"); code != exitOK {
		t.Fatal("--mark-done failed")
	}
	if _, err := os.Stat(marker); err == nil || !migrationDone(sysState, "1000000000.sh") {
		t.Fatal("--mark-done must record without running")
	}
	out.Reset()
	if code := reportMigrations("system", "list", &out); code != exitOK || !strings.Contains(out.String(), "done") {
		t.Fatalf("after mark-done: %d %q", code, out.String())
	}
}

func TestUserMigrationsUseTheUsersState(t *testing.T) {
	_, _, userState := migrationSandbox(t)
	userDir, _ := migrationDirs("user")
	writeMigration(t, userDir, "1000000000.sh", "#!/bin/sh\necho \"user fix\"\n")
	var out, errOut bytes.Buffer
	if code := runMigrations("user", &out, &errOut); code != exitOK {
		t.Fatal(errOut.String())
	}
	if !migrationDone(userState, "1000000000.sh") {
		t.Fatalf("not recorded in %s", userState)
	}
}

func TestDoctorReportsPendingAndFailedMigrations(t *testing.T) {
	sysDir, sysState, _ := migrationSandbox(t)
	writeMigration(t, sysDir, "1000000000.sh", "#!/bin/sh\necho \"x\"\n")
	text, code := captureMigrateStdout(func() int { return reportMigrationState("") })
	if code != exitOK || !strings.Contains(text, "1 pending") {
		t.Fatalf("pending: %d %q", code, text)
	}
	if err := os.MkdirAll(sysState, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sysState, "1000000000.sh.failed"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	text, code = captureMigrateStdout(func() int { return reportMigrationState("") })
	if code == exitOK || !strings.Contains(text, "FAILED") {
		t.Fatalf("failed: %d %q", code, text)
	}
}

// TestRepositoryMigrationsFollowTheRules checks the migrations ASURA ships:
// the name rule, the shebang and description lines, unique times, and syntax.
func TestRepositoryMigrationsFollowTheRules(t *testing.T) {
	repo, _ := filepath.Abs("../../../migrations")
	echoLine := regexp.MustCompile(`^echo "[^"]+"$`)
	seen := map[string]string{}
	count := 0
	for _, kind := range []string{"system", "user"} {
		entries, err := os.ReadDir(filepath.Join(repo, kind))
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range entries {
			if e.Name() == ".keep" {
				continue
			}
			path := filepath.Join(repo, kind, e.Name())
			if !migrationNameRe.MatchString(e.Name()) {
				t.Errorf("%s/%s: name must be a Unix time and .sh", kind, e.Name())
				continue
			}
			if other, dup := seen[e.Name()]; dup {
				t.Errorf("%s/%s: same time as %s", kind, e.Name(), other)
			}
			seen[e.Name()] = kind
			b, _ := os.ReadFile(path)
			lines := strings.SplitN(string(b), "\n", 3)
			if len(lines) < 3 || lines[0] != "#!/bin/sh" || !echoLine.MatchString(lines[1]) {
				t.Errorf("%s/%s: line 1 must be #!/bin/sh and line 2 echo \"what it does\"", kind, e.Name())
			}
			if out, err := exec.Command("sh", "-n", path).CombinedOutput(); err != nil {
				t.Errorf("%s/%s: %s", kind, e.Name(), out)
			}
			count++
		}
	}
	if count == 0 {
		t.Fatal("no migrations found")
	}
}

// captureMigrateStdout runs fn with os.Stdout sent to a file and returns it.
func captureMigrateStdout(fn func() int) (string, int) {
	f, err := os.CreateTemp("", "doctor-out")
	if err != nil {
		panic(err)
	}
	defer os.Remove(f.Name())
	saved := os.Stdout
	os.Stdout = f
	code := fn()
	os.Stdout = saved
	_ = f.Close()
	b, _ := os.ReadFile(f.Name())
	return string(b), code
}
