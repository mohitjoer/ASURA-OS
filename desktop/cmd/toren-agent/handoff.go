package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

// ASURA's own shell still has a quick-ask panel (launcher), a chat panel (bar)
// and a chat service that run `toren-agent ask` and `toren-agent chat`. They
// were written for a resident assistant that ASURA does not ship. Rather than
// leave them failing, `toren-agent` answers them in their own line protocols
// by handing the question to the assistant ASURA does ship: OpenCode, opened in
// the user's terminal (toren-app), with the ASURA skills wired in.

// openInAssistant opens OpenCode in a terminal, with prompt as its first
// message when there is one. It returns a sentence for the panel to show.
var openInAssistant = func(prompt string) (string, error) {
	opencode, err := exec.LookPath("opencode")
	if err != nil {
		return "", fmt.Errorf("OpenCode is not installed")
	}
	argv := []string{"toren-app", "terminal", "--", opencode}
	if prompt != "" && opencodeTakesPrompt(opencode) {
		argv = append(argv, "--prompt", prompt)
	}
	if _, err := exec.LookPath("toren-app"); err != nil {
		argv = argv[3:] // no launcher: run it directly (it needs a terminal)
	}
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if home, err := os.UserHomeDir(); err == nil {
		cmd.Dir = home
	}
	if err := cmd.Start(); err != nil {
		return "", err
	}
	_ = cmd.Process.Release()
	if prompt == "" {
		return "Opened OpenCode, ASURA's assistant, in a terminal.", nil
	}
	return "Asked OpenCode, ASURA's assistant, in a terminal window: " + prompt, nil
}

// opencodeTakesPrompt reports whether this OpenCode's TUI accepts --prompt.
func opencodeTakesPrompt(opencode string) bool {
	out, err := exec.Command(opencode, "--help").CombinedOutput()
	return err == nil && strings.Contains(string(out), "--prompt")
}

// Ask answers `toren-agent ask [--recent | --cancel | QUESTION...]` in the
// quick-ask panel's protocol: `@answer {json}` or `@error message` lines.
func Ask(args []string, out io.Writer) int {
	if len(args) > 0 {
		switch args[0] {
		case "--recent":
			fmt.Fprintln(out, "[]") // nothing is cached: answers live in OpenCode
			return 0
		case "--cancel":
			return 0
		}
	}
	question := strings.TrimSpace(strings.Join(args, " "))
	text, err := openInAssistant(question)
	if err != nil {
		fmt.Fprintf(out, "@error %v\n", err)
		return 1
	}
	answer, _ := json.Marshal(map[string]any{"text": text, "images": []string{}, "actions": []any{}})
	fmt.Fprintf(out, "@answer %s\n", answer)
	return 0
}

// Chat answers `toren-agent chat ...` in the chat panel's protocol: one JSON
// object per line, a turn ending in {"type":"done"} or {"type":"error"}.
func Chat(args []string, out io.Writer) int {
	emit := func(v map[string]any) {
		b, _ := json.Marshal(v)
		fmt.Fprintln(out, string(b))
	}
	if len(args) > 0 {
		switch args[0] {
		case "--commands":
			emit(map[string]any{"type": "commands", "commands": []any{}})
			return 0
		case "--skills":
			emit(map[string]any{"type": "skills", "skills": skillList()})
			return 0
		case "--new", "--cancel", "--set-model", "--perm":
			return 0 // no session of ASURA's own to reset, cancel or configure
		}
	}
	var words []string
	for _, a := range args {
		if !strings.HasPrefix(a, "--") {
			words = append(words, a)
		}
	}
	text, err := openInAssistant(strings.TrimSpace(strings.Join(words, " ")))
	if err != nil {
		emit(map[string]any{"type": "error", "message": err.Error()})
		return 1
	}
	emit(map[string]any{"type": "delta", "text": text})
	emit(map[string]any{"type": "done"})
	return 0
}

// skillList names the ASURA skills OpenCode is given, for the chat panel.
func skillList() []map[string]string {
	src, err := ResolveSkillSource()
	if err != nil {
		return []map[string]string{}
	}
	var out []map[string]string
	for _, dir := range []string{src, filepath.Join(filepath.Dir(src), "systemd-bridge")} {
		name, desc := skillHeader(filepath.Join(dir, "SKILL.md"))
		if name != "" {
			out = append(out, map[string]string{"name": name, "description": desc})
		}
	}
	return out
}

// skillHeader reads a SKILL.md front matter's name and the first line of its
// description.
func skillHeader(path string) (string, string) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", ""
	}
	var name, desc string
	lines := strings.Split(string(b), "\n")
	for i, line := range lines {
		if v, ok := strings.CutPrefix(line, "name:"); ok {
			name = strings.TrimSpace(v)
		}
		if v, ok := strings.CutPrefix(line, "description:"); ok {
			desc = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(v), ">"))
			if desc == "" && i+1 < len(lines) {
				desc = strings.TrimSpace(lines[i+1])
			}
		}
	}
	return name, desc
}

// WireInstalled wires the skills into every supported assistant that is on
// PATH, leaving the others alone (no folders for tools that are not there).
func WireInstalled() (int, error) {
	n := 0
	for _, a := range DetectAgents() {
		if !a.InPath {
			continue
		}
		if _, err := Wire(a.ID); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}
