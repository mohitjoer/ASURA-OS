package main

import (
	"strings"
	"testing"
)

func TestTheShippedManualHasItsChapters(t *testing.T) {
	chapters, err := manualChapters("../../../docs/manual")
	if err != nil {
		t.Fatal(err)
	}
	if len(chapters) < 8 {
		t.Fatalf("%d chapters", len(chapters))
	}
	names := map[string]bool{}
	for _, c := range chapters {
		if c.Title == "" {
			t.Errorf("%s has no # heading", c.Path)
		}
		names[c.Name] = true
	}
	// The welcome chapter's table names chapters; each must exist.
	for _, want := range []string{"desktop", "apps", "updates", "looks", "settings", "services", "hardware", "assistant", "trouble"} {
		if !names[want] {
			t.Errorf("no chapter named %s", want)
		}
	}
}

func TestManualShowsAChapterByNameOrNumber(t *testing.T) {
	oldDir, oldPager := manualDir, manualPager
	t.Cleanup(func() { manualDir, manualPager = oldDir, oldPager })
	manualDir = "../../../docs/manual"
	var shown string
	manualPager = func(text string) error { shown = text; return nil }
	if code := torenManual([]string{"updates"}); code != exitOK || !strings.HasPrefix(shown, "# Updates") {
		t.Fatalf("by name: %d %.30q", code, shown)
	}
	if code := torenManual([]string{"4"}); code != exitOK || !strings.HasPrefix(shown, "# Updates") {
		t.Fatalf("by number: %d %.30q", code, shown)
	}
	if code := torenManual([]string{"--all"}); code != exitOK || !strings.Contains(shown, "# When something goes wrong") {
		t.Fatal("--all lacks the last chapter")
	}
	if code := torenManual([]string{"nonsense"}); code != exitUsage {
		t.Fatalf("unknown: %d", code)
	}
}
