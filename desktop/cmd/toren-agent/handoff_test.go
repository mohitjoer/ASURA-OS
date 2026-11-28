package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func stubAssistant(t *testing.T, err error) *string {
	t.Helper()
	asked := new(string)
	old := openInAssistant
	openInAssistant = func(prompt string) (string, error) {
		*asked = prompt
		if err != nil {
			return "", err
		}
		return "opened: " + prompt, nil
	}
	t.Cleanup(func() { openInAssistant = old })
	return asked
}

// The quick-ask panel reads one `@answer {json}` or `@error message` line.
func TestAskSpeaksThePanelProtocol(t *testing.T) {
	asked := stubAssistant(t, nil)
	var out bytes.Buffer
	if code := Ask([]string{"why", "is", "my", "bar", "empty"}, &out); code != 0 {
		t.Fatalf("exit %d", code)
	}
	if *asked != "why is my bar empty" {
		t.Errorf("asked %q", *asked)
	}
	line := strings.TrimSpace(out.String())
	payload, ok := strings.CutPrefix(line, "@answer ")
	if !ok {
		t.Fatalf("not an @answer line: %q", line)
	}
	var a struct {
		Text    string `json:"text"`
		Actions []any  `json:"actions"`
	}
	if err := json.Unmarshal([]byte(payload), &a); err != nil || !strings.Contains(a.Text, "why is my bar empty") {
		t.Errorf("answer %q (%v)", payload, err)
	}

	stubAssistant(t, errors.New("OpenCode is not installed"))
	out.Reset()
	if code := Ask([]string{"x"}, &out); code == 0 || !strings.HasPrefix(out.String(), "@error OpenCode is not installed") {
		t.Errorf("missing assistant: exit %d, %q", code, out.String())
	}

	out.Reset()
	Ask([]string{"--recent"}, &out)
	if strings.TrimSpace(out.String()) != "[]" {
		t.Errorf("--recent must be a JSON array, got %q", out.String())
	}
}

// The chat panel reads JSON lines; a turn ends in done or error.
func TestChatSpeaksThePanelProtocol(t *testing.T) {
	stubAssistant(t, nil)
	var out bytes.Buffer
	if code := Chat([]string{"hello"}, &out); code != 0 {
		t.Fatalf("exit %d", code)
	}
	var types []string
	for _, line := range strings.Split(strings.TrimSpace(out.String()), "\n") {
		var f map[string]any
		if err := json.Unmarshal([]byte(line), &f); err != nil {
			t.Fatalf("not JSON: %q", line)
		}
		types = append(types, f["type"].(string))
	}
	if strings.Join(types, ",") != "delta,done" {
		t.Errorf("turn = %v", types)
	}
	for _, flag := range []string{"--commands", "--skills"} {
		out.Reset()
		Chat([]string{flag}, &out)
		var f map[string]any
		if err := json.Unmarshal(out.Bytes(), &f); err != nil || f["type"] != strings.TrimPrefix(flag, "--") {
			t.Errorf("%s: %q", flag, out.String())
		}
	}
}

func TestDiagnoseCrashHandoff(t *testing.T) {
	orig := openInAssistant
	defer func() { openInAssistant = orig }()

	var gotPrompt string
	openInAssistant = func(prompt string) (string, error) {
		gotPrompt = prompt
		return "handed off", nil
	}

	var out strings.Builder
	code := DiagnoseCrash([]string{"--app", "toren-store", "--summary", "toren-store crashed", "SIGSEGV at 0x0"}, &out)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d (%s)", code, out.String())
	}
	if !strings.Contains(gotPrompt, "toren-store") || !strings.Contains(gotPrompt, "SIGSEGV at 0x0") {
		t.Fatalf("crash details missing from prompt: %q", gotPrompt)
	}
}
