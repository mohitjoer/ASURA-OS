package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestHardwareReportSubmit(t *testing.T) {
	fixture := filepath.Join("..", "..", "..", "system", "hardware", "probe",
		"testdata", "hybrid-laptop")

	var openedURL string
	oldOpenBrowser := openBrowser
	openBrowser = func(targetURL string) error {
		openedURL = targetURL
		return nil
	}
	defer func() { openBrowser = oldOpenBrowser }()

	code := hardwareReport([]string{"--root", fixture, "--submit"})
	if code != exitOK {
		t.Fatalf("hardwareReport --submit returned %d, want %d", code, exitOK)
	}

	if openedURL == "" {
		t.Fatal("expected browser to be opened with an issue URL, but got empty URL")
	}

	if !strings.Contains(openedURL, "https://github.com/mohitjoer/Toren-Linux/issues/new") {
		t.Errorf("URL does not point to new issue: %s", openedURL)
	}

	if !strings.Contains(openedURL, "template=hardware-support.yml") {
		t.Errorf("URL does not specify template: %s", openedURL)
	}

	if !strings.Contains(openedURL, "title=%5Bhardware%5D") {
		t.Errorf("URL does not contain [hardware] title: %s", openedURL)
	}
}
