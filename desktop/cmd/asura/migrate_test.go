package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestMigrationOrderAndIgnore verifies that migrations are sorted in numeric order
// and any non-matching file names or directories are ignored.
func TestMigrationOrderAndIgnore(t *testing.T) {
	tmp := t.TempDir()
	sysDir := filepath.Join(tmp, "system")
	if err := os.MkdirAll(sysDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create valid migrations with out-of-order timestamps
	valid := []struct {
		name  string
		about string
	}{
		{"1795180020.sh", "Second migration"},
		{"1795180010.sh", "First migration"},
		{"1795180030.sh", "Third migration"},
	}
	for _, v := range valid {
		content := fmt.Sprintf("#!/bin/sh\necho %q\nexit 0\n", v.about)
		if err := os.WriteFile(filepath.Join(sysDir, v.name), []byte(content), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	// Create invalid/ignored files
	ignored := []string{
		".keep",
		"README.md",
		"not-a-number.sh",
		"12345.sh",       // too short (<10 digits)
		"1795180040.txt", // not .sh
	}
	for _, ign := range ignored {
		if err := os.WriteFile(filepath.Join(sysDir, ign), []byte("#!/bin/sh\necho \"ignore\"\n"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	// A directory named like a migration should also be ignored
	if err := os.Mkdir(filepath.Join(sysDir, "1795180050.sh"), 0o755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("ASURA_MIGRATIONS_DIR", tmp)

	list, err := listMigrations("system")
	if err != nil {
		t.Fatalf("listMigrations failed: %v", err)
	}

	if len(list) != 3 {
		t.Fatalf("expected 3 migrations, got %d", len(list))
	}

	expectedOrder := []string{"1795180010.sh", "1795180020.sh", "1795180030.sh"}
	expectedAbout := []string{"First migration", "Second migration", "Third migration"}
	for i, m := range list {
		if m.Name != expectedOrder[i] {
			t.Errorf("migration[%d] name = %s; want %s", i, m.Name, expectedOrder[i])
		}
		if m.About != expectedAbout[i] {
			t.Errorf("migration[%d] about = %q; want %q", i, m.About, expectedAbout[i])
		}
	}
}

// TestPendingRunAndRecord verifies that pending migrations execute, record state,
// and a subsequent run does nothing.
func TestPendingRunAndRecord(t *testing.T) {
	tree := t.TempDir()
	state := t.TempDir()
	sysDir := filepath.Join(tree, "system")
	if err := os.MkdirAll(sysDir, 0o755); err != nil {
		t.Fatal(err)
	}

	trace1 := filepath.Join(tree, "trace1.txt")
	trace2 := filepath.Join(tree, "trace2.txt")

	s1 := fmt.Sprintf("#!/bin/sh\necho \"Migration 1\"\ntouch %q\n", trace1)
	s2 := fmt.Sprintf("#!/bin/sh\necho \"Migration 2\"\ntouch %q\n", trace2)
	if err := os.WriteFile(filepath.Join(sysDir, "1795180010.sh"), []byte(s1), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sysDir, "1795180020.sh"), []byte(s2), 0o755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("ASURA_MIGRATIONS_DIR", tree)
	t.Setenv("ASURA_MIGRATION_STATE", state)

	var out, errOut bytes.Buffer
	code := runMigrations("system", &out, &errOut)
	if code != exitOK {
		t.Fatalf("runMigrations failed with code %d; err: %s", code, errOut.String())
	}

	if _, err := os.Stat(trace1); err != nil {
		t.Errorf("trace1 was not created: %v", err)
	}
	if _, err := os.Stat(trace2); err != nil {
		t.Errorf("trace2 was not created: %v", err)
	}

	// Markers must exist
	if !migrationDone(state, "1795180010.sh") {
		t.Errorf("marker for 1795180010.sh missing")
	}
	if !migrationDone(state, "1795180020.sh") {
		t.Errorf("marker for 1795180020.sh missing")
	}

	// Log must exist
	logPath := filepath.Join(filepath.Dir(state), "migrate.log")
	if _, err := os.Stat(logPath); err != nil {
		t.Errorf("migrate.log was not created at %s", logPath)
	}

	// Output check
	outStr := out.String()
	if !strings.Contains(outStr, "Migration 1 (1795180010.sh)") {
		t.Errorf("out missing migration 1 notice: %s", outStr)
	}
	if !strings.Contains(outStr, "ran 2") {
		t.Errorf("out missing ran 2: %s", outStr)
	}

	// Second run runs nothing
	out.Reset()
	errOut.Reset()
	code2 := runMigrations("system", &out, &errOut)
	if code2 != exitOK {
		t.Fatalf("second runMigrations returned %d", code2)
	}
	if !strings.Contains(out.String(), "nothing to run") {
		t.Errorf("second run did not report 'nothing to run': %s", out.String())
	}
}

// TestFailingMigrationStopsAndRetries verifies that a failure stops execution,
// writes a .failed marker, does not record success, and once repaired can run again.
func TestFailingMigrationStopsAndRetries(t *testing.T) {
	tree := t.TempDir()
	state := t.TempDir()
	sysDir := filepath.Join(tree, "system")
	if err := os.MkdirAll(sysDir, 0o755); err != nil {
		t.Fatal(err)
	}

	trace1 := filepath.Join(tree, "trace1.txt")
	trace3 := filepath.Join(tree, "trace3.txt")

	s1 := fmt.Sprintf("#!/bin/sh\necho \"Step 1\"\ntouch %q\n", trace1)
	s2Fail := "#!/bin/sh\necho \"Step 2\"\nexit 42\n"
	s3 := fmt.Sprintf("#!/bin/sh\necho \"Step 3\"\ntouch %q\n", trace3)

	f1 := filepath.Join(sysDir, "1795180010.sh")
	f2 := filepath.Join(sysDir, "1795180020.sh")
	f3 := filepath.Join(sysDir, "1795180030.sh")

	os.WriteFile(f1, []byte(s1), 0o755)
	os.WriteFile(f2, []byte(s2Fail), 0o755)
	os.WriteFile(f3, []byte(s3), 0o755)

	t.Setenv("ASURA_MIGRATIONS_DIR", tree)
	t.Setenv("ASURA_MIGRATION_STATE", state)

	var out, errOut bytes.Buffer
	code := runMigrations("system", &out, &errOut)
	if code != exitFail {
		t.Fatalf("expected code exitFail (%d), got %d", exitFail, code)
	}

	// 1 succeeded
	if !migrationDone(state, "1795180010.sh") {
		t.Errorf("step 1 should be done")
	}
	// 2 failed
	if migrationDone(state, "1795180020.sh") {
		t.Errorf("step 2 should NOT be marked done")
	}
	if !migrationFailed(state, "1795180020.sh") {
		t.Errorf("step 2 should have .failed marker")
	}
	// 3 was not run
	if _, err := os.Stat(trace3); err == nil {
		t.Errorf("step 3 should NOT have run")
	}
	if migrationDone(state, "1795180030.sh") {
		t.Errorf("step 3 should NOT be done")
	}

	// Fix script 2
	s2Fixed := "#!/bin/sh\necho \"Step 2 repaired\"\nexit 0\n"
	os.WriteFile(f2, []byte(s2Fixed), 0o755)

	out.Reset()
	errOut.Reset()
	code2 := runMigrations("system", &out, &errOut)
	if code2 != exitOK {
		t.Fatalf("re-run failed: %d, err: %s", code2, errOut.String())
	}

	// .failed marker must be removed
	if migrationFailed(state, "1795180020.sh") {
		t.Errorf("step 2 .failed marker should be cleared")
	}
	if !migrationDone(state, "1795180020.sh") {
		t.Errorf("step 2 should now be marked done")
	}
	if !migrationDone(state, "1795180030.sh") {
		t.Errorf("step 3 should now be marked done")
	}
	if _, err := os.Stat(trace3); err != nil {
		t.Errorf("step 3 should have run: %v", err)
	}
}

// TestPendingListAndMarkDone verifies flags --pending, --list and --mark-done.
func TestPendingListAndMarkDone(t *testing.T) {
	tree := t.TempDir()
	userStateBase := t.TempDir()
	userDir := filepath.Join(tree, "user")
	if err := os.MkdirAll(userDir, 0o755); err != nil {
		t.Fatal(err)
	}

	s1 := "#!/bin/sh\necho \"User fix 1\"\nexit 0\n"
	s2 := "#!/bin/sh\necho \"User fix 2\"\nexit 0\n"
	os.WriteFile(filepath.Join(userDir, "1795180010.sh"), []byte(s1), 0o755)
	os.WriteFile(filepath.Join(userDir, "1795180020.sh"), []byte(s2), 0o755)

	t.Setenv("ASURA_MIGRATIONS_DIR", tree)
	t.Setenv("XDG_STATE_HOME", userStateBase)

	stateDir := filepath.Join(userStateBase, "asura", "migrations")

	// 1. --pending
	var out bytes.Buffer
	codePending := reportMigrations("user", "pending", &out)
	if codePending != exitNoMatch {
		t.Errorf("pending report should return exitNoMatch (%d) when pending, got %d", exitNoMatch, codePending)
	}
	if !strings.Contains(out.String(), "1795180010.sh") || !strings.Contains(out.String(), "1795180020.sh") {
		t.Errorf("unexpected pending output: %s", out.String())
	}
	// Verify no markers exist
	if _, err := os.Stat(stateDir); err == nil {
		entries, _ := os.ReadDir(stateDir)
		if len(entries) > 0 {
			t.Errorf("pending created state entries: %d", len(entries))
		}
	}

	// 2. --list
	out.Reset()
	codeList := reportMigrations("user", "list", &out)
	if codeList != exitNoMatch {
		t.Errorf("list report with pending should return exitNoMatch, got %d", codeList)
	}
	if !strings.Contains(out.String(), "pending") {
		t.Errorf("list output missing 'pending': %s", out.String())
	}

	// 3. --mark-done
	codeMark := markMigrationsDone("user")
	if codeMark != exitOK {
		t.Errorf("markMigrationsDone failed with %d", codeMark)
	}
	if !migrationDone(stateDir, "1795180010.sh") || !migrationDone(stateDir, "1795180020.sh") {
		t.Errorf("markers were not created by markMigrationsDone")
	}

	// 4. --list after mark-done
	out.Reset()
	codeListDone := reportMigrations("user", "list", &out)
	if codeListDone != exitOK {
		t.Errorf("list report when done should return exitOK, got %d", codeListDone)
	}
	if strings.Contains(out.String(), "pending") || !strings.Contains(out.String(), "done") {
		t.Errorf("list output expected all done: %s", out.String())
	}

	// 5. --pending after mark-done
	out.Reset()
	codePendingDone := reportMigrations("user", "pending", &out)
	if codePendingDone != exitOK {
		t.Errorf("pending report when none pending should return exitOK, got %d", codePendingDone)
	}
	if !strings.Contains(out.String(), "nothing pending") {
		t.Errorf("pending report expected 'nothing pending': %s", out.String())
	}
}

// TestDoctorMigrationsCheck verifies doctor check pending vs FAILED reporting.
func TestDoctorMigrationsCheck(t *testing.T) {
	tree := t.TempDir()
	stateSys := t.TempDir()
	userStateBase := t.TempDir()

	os.MkdirAll(filepath.Join(tree, "system"), 0o755)
	os.MkdirAll(filepath.Join(tree, "user"), 0o755)

	os.WriteFile(filepath.Join(tree, "system", "1795180010.sh"), []byte("#!/bin/sh\necho \"Sys 1\"\n"), 0o755)
	os.WriteFile(filepath.Join(tree, "user", "1795180010.sh"), []byte("#!/bin/sh\necho \"User 1\"\n"), 0o755)

	t.Setenv("ASURA_MIGRATIONS_DIR", tree)
	t.Setenv("ASURA_MIGRATION_STATE", stateSys)
	t.Setenv("XDG_STATE_HOME", userStateBase)

	// Case 1: Pending migrations (doctor reports pending, exits OK)
	pendingSys, failedSys := pendingMigrations("system")
	if pendingSys != 1 || len(failedSys) != 0 {
		t.Errorf("expected 1 pending, 0 failed; got %d, %v", pendingSys, failedSys)
	}
	codeDoc := reportMigrationState("")
	if codeDoc != exitOK {
		t.Errorf("doctor with pending migrations should return exitOK, got %d", codeDoc)
	}

	// Case 2: Mark done (doctor reports all done, exits OK)
	markMigrationsDone("system")
	markMigrationsDone("user")
	pendingSys, failedSys = pendingMigrations("system")
	if pendingSys != 0 || len(failedSys) != 0 {
		t.Errorf("expected 0 pending, 0 failed; got %d, %v", pendingSys, failedSys)
	}
	codeDocDone := reportMigrationState("")
	if codeDocDone != exitOK {
		t.Errorf("doctor with done migrations should return exitOK, got %d", codeDocDone)
	}

	// Case 3: A migration has failed (.failed marker) -> doctor returns exitFail
	os.WriteFile(filepath.Join(stateSys, "1795180010.sh.failed"), []byte("exit 1\n"), 0o644)
	pendingSys, failedSys = pendingMigrations("system")
	if len(failedSys) != 1 || failedSys[0] != "1795180010.sh" {
		t.Errorf("expected failedSys=[1795180010.sh], got %v", failedSys)
	}
	codeDocFail := reportMigrationState("")
	if codeDocFail != exitFail {
		t.Errorf("doctor with failed migration must return exitFail (%d), got %d", exitFail, codeDocFail)
	}
}

// TestRepoMigrationsRules validates the repository's shipped migrations in migrations/.
func TestRepoMigrationsRules(t *testing.T) {
	repoMigrations := filepath.Join("..", "..", "..", "migrations")
	if _, err := os.Stat(repoMigrations); err != nil {
		t.Skipf("migrations dir not found at %s: %v", repoMigrations, err)
	}

	reName := regexp.MustCompile(`^[0-9]{10,}\.sh$`)
	seenTimestamps := make(map[string]string)

	for _, kind := range []string{"system", "user"} {
		dir := filepath.Join(repoMigrations, kind)
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatalf("failed to read %s: %v", dir, err)
		}

		for _, e := range entries {
			if e.Name() == ".keep" {
				continue
			}
			path := filepath.Join(dir, e.Name())

			// 1. Matches name rule
			if !reName.MatchString(e.Name()) {
				t.Errorf("%s: does not match migration name regex ^[0-9]{10,}\\.sh$", path)
			}

			// 2. Unique timestamp across migrations
			if prev, exists := seenTimestamps[e.Name()]; exists {
				t.Errorf("%s: duplicate timestamp with %s", path, prev)
			}
			seenTimestamps[e.Name()] = path

			// 3. Line 1 is #!/bin/sh and Line 2 is echo "..."
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("could not read %s: %v", path, err)
			}
			lines := strings.Split(string(data), "\n")
			if len(lines) < 2 {
				t.Errorf("%s: file has fewer than 2 lines", path)
				continue
			}
			if strings.TrimSpace(lines[0]) != "#!/bin/sh" {
				t.Errorf("%s: line 1 must be #!/bin/sh, got %q", path, lines[0])
			}
			line2 := strings.TrimSpace(lines[1])
			if !strings.HasPrefix(line2, "echo \"") || !strings.HasSuffix(line2, "\"") {
				t.Errorf("%s: line 2 must be echo \"One sentence: what it does\", got %q", path, line2)
			}

			// 4. Passes sh -n syntax check
			cmd := exec.Command("sh", "-n", path)
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Errorf("%s: sh -n syntax check failed: %v\nOutput: %s", path, err, string(out))
			}
		}
	}
}
