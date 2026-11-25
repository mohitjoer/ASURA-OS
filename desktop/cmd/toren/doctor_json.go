package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"toren/desktop/internal/overlay"
)

// doctorSection is one check in `toren doctor --json`.
type doctorSection struct {
	Section string   `json:"section"`
	OK      bool     `json:"ok"`
	Lines   []string `json:"lines"`
}

// doctorReport is the whole of `toren doctor --json`.
type doctorReport struct {
	OK       bool            `json:"ok"`
	Sections []doctorSection `json:"sections"`
}

// stdoutMu serialises captureStdout: it swaps the process's stdout.
var stdoutMu sync.Mutex

// captureStdout runs fn with os.Stdout redirected and returns what it printed.
// The report functions print as they go; capturing them, rather than giving
// each a writer, keeps the text report exactly as it was.
func captureStdout(fn func() int) (string, int) {
	stdoutMu.Lock()
	defer stdoutMu.Unlock()
	r, w, err := os.Pipe()
	if err != nil {
		return "", fn()
	}
	saved := os.Stdout
	os.Stdout = w
	var buf bytes.Buffer
	done := make(chan struct{})
	go func() {
		_, _ = io.Copy(&buf, r)
		close(done)
	}()
	code := func() int {
		defer func() { os.Stdout = saved }()
		return fn()
	}()
	_ = w.Close()
	<-done
	_ = r.Close()
	return buf.String(), code
}

// sectionFromText turns one check's printed report into a section: its first
// unindented line is the title, every other non-blank line is kept trimmed.
func sectionFromText(text string, code int, fallback string) doctorSection {
	s := doctorSection{Section: fallback, OK: code == exitOK, Lines: []string{}}
	titled := false
	for _, line := range strings.Split(text, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if !titled && !strings.HasPrefix(line, " ") {
			s.Section = strings.TrimSpace(line)
			titled = true
			continue
		}
		s.Lines = append(s.Lines, strings.TrimSpace(line))
	}
	return s
}

// doctorJSON runs every doctor check and prints one JSON object. The exit code
// is the same as the text report's.
func doctorJSON(root string, fix bool, out io.Writer) int {
	report := doctorReport{OK: true}
	code := exitOK
	add := func(s doctorSection, c int) {
		report.Sections = append(report.Sections, s)
		if c != exitOK {
			report.OK = false
			code = c
		}
	}

	drift := doctorSection{Section: "Config drift", OK: true, Lines: []string{}}
	paths, err := overlayPaths(root)
	if err == nil {
		var findings []overlay.Drift
		findings, err = overlay.Doctor(paths)
		for _, f := range findings {
			drift.Lines = append(drift.Lines, fmt.Sprintf("%s %s", f.Reason, f.Path))
		}
	}
	if err != nil {
		drift.OK = false
		drift.Lines = append(drift.Lines, err.Error())
		add(drift, exitFail)
	} else {
		if len(drift.Lines) == 0 {
			drift.Lines = append(drift.Lines, "no drift: the config tree matches the base")
		}
		add(drift, exitOK) // drift is reported, not a failure, as in the text report
	}

	for i, check := range sessionChecks(root, fix) {
		text, c := captureStdout(check)
		add(sectionFromText(text, c, fmt.Sprintf("check %d", i+1)), c)
	}

	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	if err := enc.Encode(report); err != nil {
		fmt.Fprintln(os.Stderr, "toren doctor:", err)
		return exitFail
	}
	return code
}
