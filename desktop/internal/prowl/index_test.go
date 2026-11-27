package prowl

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIndexMasksSecretsAndCitesSearch(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "prowl.db")

	src := "package auth\n\nconst apiKey = \"" + "sk-" + "proj-1234567890abcdefghijklmnopqrstuvwxyzABCD\"\n\nfunc VerifyToken(token string) bool {\n\treturn token != \"\"\n}\n"
	if err := os.WriteFile(filepath.Join(dir, "auth.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	idx, err := OpenIndex(dbPath, dir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = idx.Close() }()

	res, err := idx.IndexWorkspace(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if res.Indexed != 1 || res.MaskedSecrets != 1 {
		t.Fatalf("unexpected index result: %+v", res)
	}

	// Verify the raw SQLite table does NOT hold the unmasked secret.
	var stored string
	if err := idx.db.QueryRow(`SELECT content FROM chunks LIMIT 1`).Scan(&stored); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(stored, "1234567890abcdefghijklmnopqrstuvwxyz") {
		t.Fatalf("unmasked secret stored in sqlite:\n%s", stored)
	}

	hits, err := idx.Search(context.Background(), "VerifyToken", 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) == 0 {
		t.Fatal("expected cited search hit for VerifyToken")
	}
	if hits[0].Path != "auth.go" || hits[0].StartLine != 1 || hits[0].EndLine < 6 {
		t.Fatalf("unexpected citation: %+v", hits[0])
	}

	// Incremental re-index skips unchanged file.
	res2, err := idx.IndexWorkspace(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if res2.Skipped != 1 || res2.Indexed != 0 {
		t.Fatalf("expected unchanged file to be skipped, got %+v", res2)
	}
}
