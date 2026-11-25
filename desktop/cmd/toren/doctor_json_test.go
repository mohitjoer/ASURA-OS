package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
)

func TestSectionFromTextTakesTheHeadingAsTitle(t *testing.T) {
	s := sectionFromText("\nPortals\n  configuration   none found\n\n  problem  x\n", exitFail, "check 1")
	if s.Section != "Portals" || s.OK || len(s.Lines) != 2 || s.Lines[0] != "configuration   none found" {
		t.Fatalf("got %+v", s)
	}
	if s := sectionFromText("", exitOK, "check 2"); s.Section != "check 2" || !s.OK || s.Lines == nil {
		t.Fatalf("an empty report: got %+v", s)
	}
}

func TestCaptureStdoutReturnsWhatACheckPrints(t *testing.T) {
	text, code := captureStdout(func() int { fmt.Println("Hub"); fmt.Println("  all fine"); return exitFail })
	if text != "Hub\n  all fine\n" || code != exitFail {
		t.Fatalf("got %q, %d", text, code)
	}
}

// TestDoctorJSONIsOneObject runs the real checks against an empty tree: it
// must print one JSON object with every section, and fail like the text report.
func TestDoctorJSONIsOneObject(t *testing.T) {
	root := t.TempDir()
	var out bytes.Buffer
	code := doctorJSON(root, false, &out)
	var report doctorReport
	if err := json.Unmarshal(out.Bytes(), &report); err != nil {
		t.Fatalf("not JSON: %v\n%s", err, out.String())
	}
	if want := 1 + len(sessionChecks(root, false)); len(report.Sections) != want {
		t.Fatalf("%d sections, want %d", len(report.Sections), want)
	}
	if report.OK || code == exitOK {
		t.Fatalf("an empty tree passed: ok=%v code=%d", report.OK, code)
	}
	if report.Sections[0].Section != "Config drift" {
		t.Fatalf("first section %q", report.Sections[0].Section)
	}
}
