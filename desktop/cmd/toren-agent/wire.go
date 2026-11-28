package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// skillNames is the ordered list of Toren skill directories to wire for every agent.
var skillNames = []string{"toren", "toren-create", "toren-prowl", "systemd-bridge"}

// AgentInfo describes a supported local coding agent harness.
type AgentInfo struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Binaries   []string          `json:"binaries"`
	InPath     bool              `json:"inPath"`
	SkillPath  string            `json:"skillPath"`  // primary (toren) skill path, kept for compat
	SkillPaths map[string]string `json:"skillPaths"` // name → full path for all skill dirs
	Wired      bool              `json:"wired"`      // true only if ALL skills are wired
}

// AgentDef defines how to resolve an agent's paths and detect its presence.
type AgentDef struct {
	ID           string
	Name         string
	Binaries     []string
	SkillsRoot   func() string // parent dir that contains the per-skill subdirs
}

func getHome() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return "/root"
}

func getConfigHome() string {
	if c := os.Getenv("XDG_CONFIG_HOME"); c != "" {
		return c
	}
	return filepath.Join(getHome(), ".config")
}

// SupportedAgents returns the definitions for all recognized local coding agents.
func SupportedAgents() []AgentDef {
	return []AgentDef{
		{
			ID:       "opencode",
			Name:     "OpenCode CLI",
			Binaries: []string{"opencode"},
			SkillsRoot: func() string {
				return filepath.Join(getConfigHome(), "opencode", "skills")
			},
		},
		{
			ID:       "pi",
			Name:     "Pi coding agent",
			Binaries: []string{"pi"},
			SkillsRoot: func() string {
				return filepath.Join(getHome(), ".pi", "agent", "skills")
			},
		},
		{
			ID:       "claude",
			Name:     "Claude Code",
			Binaries: []string{"claude"},
			SkillsRoot: func() string {
				return filepath.Join(getHome(), ".claude", "skills")
			},
		},
		{
			ID:       "gemini",
			Name:     "Gemini CLI / Antigravity",
			Binaries: []string{"gemini", "antigravity", "agy"},
			SkillsRoot: func() string {
				return filepath.Join(getHome(), ".gemini", "skills")
			},
		},
	}
}

// skillPathFor returns the full destination path for a given agent + skill name.
func skillPathFor(d AgentDef, name string) string {
	return filepath.Join(d.SkillsRoot(), name)
}

// resolveSkillsRoot locates the parent "skills" directory that contains all Toren skill subdirs.
// Priority:
// 1. ASURA_SKILLS_DIR environment override (the dir itself, or its parent)
// 2. Packaged system path /usr/share/toren/shell/agent/skills
// 3. Repository checkout fallback (shell/agent/skills relative to cwd / exe)
func resolveSkillsRoot() (string, error) {
	if env := os.Getenv("ASURA_SKILLS_DIR"); env != "" {
		// env might point to the root itself or to the "toren" subdirectory
		if hasSkillDir(env, "toren") {
			return env, nil
		}
		parent := filepath.Dir(env)
		if hasSkillDir(parent, "toren") {
			return parent, nil
		}
	}

	sysRoot := "/usr/share/toren/shell/agent/skills"
	if hasSkillDir(sysRoot, "toren") {
		return sysRoot, nil
	}

	roots := []string{
		"shell/agent/skills",
		"../shell/agent/skills",
		"../../shell/agent/skills",
	}

	if pwd, err := os.Getwd(); err == nil {
		roots = append(roots,
			filepath.Join(pwd, "shell", "agent", "skills"),
			filepath.Join(pwd, "..", "shell", "agent", "skills"),
			filepath.Join(pwd, "..", "..", "shell", "agent", "skills"),
		)
	}

	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		roots = append(roots,
			filepath.Join(exeDir, "..", "shell", "agent", "skills"),
			filepath.Join(exeDir, "..", "..", "shell", "agent", "skills"),
			filepath.Join(exeDir, "..", "..", "..", "shell", "agent", "skills"),
		)
	}

	for _, r := range roots {
		clean := filepath.Clean(r)
		if hasSkillDir(clean, "toren") {
			if abs, err := filepath.Abs(clean); err == nil {
				return abs, nil
			}
			return clean, nil
		}
	}

	return "", fmt.Errorf("could not locate Toren skills root (checked %s and repo fallbacks)", sysRoot)
}

// hasSkillDir returns true if root/name/SKILL.md exists.
func hasSkillDir(root, name string) bool {
	_, err := os.Stat(filepath.Join(root, name, "SKILL.md"))
	return err == nil
}

// ResolveSkillSource locates the primary "toren" skill directory (kept for compatibility).
func ResolveSkillSource() (string, error) {
	root, err := resolveSkillsRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "toren"), nil
}

func isValidSkillDir(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "SKILL.md"))
	return err == nil
}

// isOurSkillLink checks if a symlink points into the Toren skills tree.
func isOurSkillLink(link string) bool {
	fi, err := os.Lstat(link)
	if err != nil || fi.Mode()&os.ModeSymlink == 0 {
		return false
	}
	dest, err := os.Readlink(link)
	if err != nil {
		return false
	}
	if !filepath.IsAbs(dest) {
		dest = filepath.Join(filepath.Dir(link), dest)
	}
	return isValidSkillDir(filepath.Clean(dest))
}

// symlinkForce idempotently points link to target.
// Returns (createdOrUpdated, error).
func symlinkForce(link, target string) (bool, error) {
	fi, err := os.Lstat(link)
	if err == nil {
		if fi.Mode()&os.ModeSymlink == 0 {
			return false, fmt.Errorf("%s exists and is a regular file/directory, refusing to overwrite", link)
		}
		dest, readErr := os.Readlink(link)
		if readErr == nil {
			absDest := dest
			if !filepath.IsAbs(absDest) {
				absDest = filepath.Join(filepath.Dir(link), absDest)
			}
			if filepath.Clean(absDest) == filepath.Clean(target) {
				return false, nil // Already wired identically
			}
		}
		if rmErr := os.Remove(link); rmErr != nil {
			return false, rmErr
		}
	} else if !os.IsNotExist(err) {
		return false, err
	}

	if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
		return false, err
	}
	if err := os.Symlink(target, link); err != nil {
		return false, err
	}
	return true, nil
}

// DetectAgents inspects the system for each supported agent and its wiring status.
func DetectAgents() []AgentInfo {
	defs := SupportedAgents()
	out := make([]AgentInfo, 0, len(defs))

	for _, d := range defs {
		inPath := false
		for _, bin := range d.Binaries {
			if _, err := exec.LookPath(bin); err == nil {
				inPath = true
				break
			}
		}

		root := d.SkillsRoot()
		paths := make(map[string]string, len(skillNames))
		allWired := true
		for _, name := range skillNames {
			p := filepath.Join(root, name)
			paths[name] = p
			if !isOurSkillLink(p) {
				allWired = false
			}
		}

		out = append(out, AgentInfo{
			ID:         d.ID,
			Name:       d.Name,
			Binaries:   d.Binaries,
			InPath:     inPath,
			SkillPath:  filepath.Join(root, "toren"), // compat field
			SkillPaths: paths,
			Wired:      allWired,
		})
	}

	return out
}

// Wire links all Toren skill directories into one agent or all agents.
func Wire(agentID string) (int, error) {
	root, err := resolveSkillsRoot()
	if err != nil {
		return 0, err
	}

	defs := SupportedAgents()
	count := 0

	for _, d := range defs {
		if agentID != "" && d.ID != agentID {
			continue
		}
		anyNew := false
		for _, name := range skillNames {
			src := filepath.Join(root, name)
			link := skillPathFor(d, name)
			created, werr := symlinkForce(link, src)
			if werr != nil {
				return count, fmt.Errorf("wiring %s/%s (%s): %w", d.ID, name, link, werr)
			}
			if created {
				fmt.Printf("  wired   %s/%-18s -> %s\n", d.ID, name, tildeAbbrev(link))
				anyNew = true
			} else {
				fmt.Printf("  current %s/%-18s -> %s\n", d.ID, name, tildeAbbrev(link))
			}
		}
		if anyNew {
			fmt.Printf("wired %s (%d skills)\n", d.ID, len(skillNames))
		} else {
			fmt.Printf("already wired %s (%d skills)\n", d.ID, len(skillNames))
		}
		count++
	}

	if agentID != "" && count == 0 {
		return 0, fmt.Errorf("unknown agent %q (supported: opencode, pi, claude, gemini)", agentID)
	}

	return count, nil
}

// Unwire removes all Toren skill symlinks for one agent or all agents.
func Unwire(agentID string) (int, error) {
	defs := SupportedAgents()
	count := 0

	for _, d := range defs {
		if agentID != "" && d.ID != agentID {
			continue
		}
		removed := 0
		for _, name := range skillNames {
			link := skillPathFor(d, name)
			if isOurSkillLink(link) {
				if err := os.Remove(link); err != nil {
					return count, fmt.Errorf("unwiring %s/%s (%s): %w", d.ID, name, link, err)
				}
				fmt.Printf("  removed %s/%s (%s)\n", d.ID, name, tildeAbbrev(link))
				removed++
			}
		}
		if removed > 0 {
			fmt.Printf("unwired %s (%d skills removed)\n", d.ID, removed)
			count++
		} else {
			fmt.Printf("not wired %s\n", d.ID)
		}
	}

	if agentID != "" && count == 0 && !isValidAgentID(agentID) {
		return 0, fmt.Errorf("unknown agent %q (supported: opencode, pi, claude, gemini)", agentID)
	}

	return count, nil
}

func isValidAgentID(id string) bool {
	for _, d := range SupportedAgents() {
		if d.ID == id {
			return true
		}
	}
	return false
}

func tildeAbbrev(path string) string {
	h := getHome()
	if strings.HasPrefix(path, h) {
		return "~" + path[len(h):]
	}
	return path
}

// Status prints the detection and wiring status of all agents.
func Status(asJSON bool) error {
	agents := DetectAgents()
	if asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(agents)
	}

	fmt.Printf("%-10s %-25s %-9s %-7s %s\n", "AGENT", "NAME", "IN PATH", "WIRED", "SKILLS ROOT")
	for _, a := range agents {
		inPathStr := "no"
		if a.InPath {
			inPathStr = "yes"
		}
		wiredStr := "no"
		if a.Wired {
			wiredStr = "yes (all)"
		} else {
			// count which ones are wired
			n := 0
			for _, p := range a.SkillPaths {
				if isOurSkillLink(p) {
					n++
				}
			}
			if n > 0 {
				wiredStr = fmt.Sprintf("partial (%d/%d)", n, len(skillNames))
			}
		}
		fmt.Printf("%-10s %-25s %-9s %-16s %s\n", a.ID, a.Name, inPathStr, wiredStr, tildeAbbrev(filepath.Dir(a.SkillPath)))
	}
	return nil
}
