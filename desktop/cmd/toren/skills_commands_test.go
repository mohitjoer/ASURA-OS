package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// The agent skills (shell/agent/skills) tell an assistant which commands to run
// on an ASURA machine. A command they name that does not exist sends the
// assistant down a dead end, so every `toren <sub>` and `toren-<tool>` written
// in a skill's code must be a real command.

var (
	codeSpan   = regexp.MustCompile("(?s)```.*?```|`[^`\n]+`")
	torenSub   = regexp.MustCompile("(?:^|[\\s(|;&`])toren ([a-z][a-z-]*)")
	torenTool  = regexp.MustCompile(`\b(toren-[a-z]+(?:-[a-z]+)*)`)
	notCommand = map[string]bool{
		"toren-plugins":    true, // the folder new plugins are created in
		"toren-shell.sock": true,
	}
)

func TestSkillsNameRealCommands(t *testing.T) {
	repo, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	skills := filepath.Join(repo, "shell", "agent", "skills")
	files, _ := filepath.Glob(filepath.Join(skills, "*", "*.md"))
	// The user manual makes the same promises to people.
	manual, _ := filepath.Glob(filepath.Join(repo, "docs", "manual", "*.md"))
	files = append(files, manual...)
	if len(files) == 0 {
		t.Fatalf("no skills under %s", skills)
	}
	toolExists := func(name string) bool {
		candidates := []string{name}
		if strings.HasPrefix(name, "toren-") {
			candidates = append(candidates, "toren-"+strings.TrimPrefix(name, "toren-"))
		} else if name == "toren" {
			candidates = append(candidates, "toren")
		}
		for _, cand := range candidates {
			// Commands, scripts, and the user services (a name like toren-updates).
			for _, dir := range []string{"desktop/cmd", "shell/scripts", "session", "system/runit/user-services"} {
				if _, err := os.Stat(filepath.Join(repo, dir, cand)); err == nil {
					return true
				}
			}
			// Helpers apply-repo.sh installs to /usr/bin from system/helpers/*/.
			helpers, _ := filepath.Glob(filepath.Join(repo, "system", "helpers", "*", cand))
			if len(helpers) > 0 {
				return true
			}
		}
		return false
	}
	for _, file := range files {
		b, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		rel, _ := filepath.Rel(repo, file)
		for _, code := range codeSpan.FindAllString(string(b), -1) {
			for _, m := range torenSub.FindAllStringSubmatch(code, -1) {
				_, builtIn := commands[m[1]]
				_, registered := extraCommands[m[1]]
				if !builtIn && !registered {
					t.Errorf("%s: `toren %s` is not an toren command", rel, m[1])
				}
			}
			for _, m := range torenTool.FindAllStringSubmatch(code, -1) {
				name := strings.TrimSuffix(m[1], "-")
				if notCommand[name] || strings.Contains(code, name+".") || strings.Contains(code, name+"/") {
					continue // a file or folder name, not a command
				}
				if !toolExists(name) {
					t.Errorf("%s: `%s` is not a command ASURA ships", rel, name)
				}
			}
		}
	}
}
