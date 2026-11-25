package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTuneWindowsConf(t *testing.T) {
	tmp := t.TempDir()
	conf := filepath.Join(tmp, "windows-11.conf")
	initial := "guest_os=\"windows\"\n"
	if err := os.WriteFile(conf, []byte(initial), 0o644); err != nil {
		t.Fatal(err)
	}

	tuneWindowsConf(conf)

	b, err := os.ReadFile(conf)
	if err != nil {
		t.Fatal(err)
	}
	content := string(b)
	if !strings.Contains(content, `tpm="on"`) {
		t.Errorf("conf missing tpm=on: %s", content)
	}
	if !strings.Contains(content, `cpu_cores="4"`) {
		t.Errorf("conf missing cpu_cores=4: %s", content)
	}
	if !strings.Contains(content, `ram="8G"`) {
		t.Errorf("conf missing ram=8G: %s", content)
	}
}

func TestVmWindowsAbort(t *testing.T) {
	oldReader := vmStdinReader
	defer func() { vmStdinReader = oldReader }()
	vmStdinReader = bufio.NewReader(strings.NewReader("n\n"))

	code := vmWindows(nil)
	if code != exitOK {
		t.Errorf("expected exitOK when user aborts, got %d", code)
	}
}
