package prowl

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMCPServerSearchAndMasking(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "prowl.db")
	fakeTok := "ghp_" + "1234567890abcdefghijklmnopqrstuvwxyz12"
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(
		"package main\nconst tok = \""+fakeTok+"\"\nfunc BootRouter() {}\n",
	), 0o644); err != nil {
		t.Fatal(err)
	}

	idx, err := OpenIndex(dbPath, dir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = idx.Close() }()
	if _, err := idx.IndexWorkspace(context.Background()); err != nil {
		t.Fatal(err)
	}

	srv := NewMCPServer(idx, nil)
	reqs := strings.Join([]string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`,
		`{"jsonrpc":"2.0","method":"notifications/initialized"}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/list"}`,
		`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"prowl_search","arguments":{"query":"BootRouter"}}}`,
		`{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"prowl_read","arguments":{"path":"main.go","start_line":1,"end_line":3}}}`,
	}, "\n") + "\n"

	var out bytes.Buffer
	if err := srv.Serve(context.Background(), strings.NewReader(reqs), &out); err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != 4 {
		t.Fatalf("expected 4 JSON-RPC responses (notification skipped), got %d:\n%s", len(lines), out.String())
	}

	var readResp struct {
		Result struct {
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(lines[3]), &readResp); err != nil {
		t.Fatal(err)
	}
	text := readResp.Result.Content[0].Text
	if strings.Contains(text, fakeTok) {
		t.Fatalf("prowl_read leaked unmasked secret:\n%s", text)
	}
	if !strings.Contains(text, "BootRouter") {
		t.Fatalf("expected BootRouter in prowl_read output:\n%s", text)
	}
}
