package quirks

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Conflict describes a file target collision between two quirk definitions
// that can match the same machine without a declared override.
type Conflict struct {
	Target string
	DefA   string
	DefB   string
	Reason string
}

// CanCoexist checks if two quirk definitions could possibly match the same machine.
func CanCoexist(a, b QuirkDef) bool {
	// If one is generic, it can coexist with anything
	if a.Layer == LayerGeneric || b.Layer == LayerGeneric {
		return true
	}

	// Check chassis compatibility
	if a.Match.Chassis != "" && b.Match.Chassis != "" {
		cA := strings.ToLower(strings.TrimSpace(a.Match.Chassis))
		cB := strings.ToLower(strings.TrimSpace(b.Match.Chassis))
		if cA != cB && !(IsLaptopChassis(cA) && IsLaptopChassis(cB)) {
			return false
		}
	}

	// Check sys_vendor compatibility
	if a.Match.SysVendor != "" && b.Match.SysVendor != "" {
		if !vendorsOverlap(a.Match.SysVendor, b.Match.SysVendor) {
			return false
		}
	}

	// Check product_name compatibility
	if a.Match.ProductName != "" && b.Match.ProductName != "" {
		if !patternsOverlap(a.Match.ProductName, b.Match.ProductName) {
			return false
		}
	}

	// Check board_name compatibility
	if a.Match.BoardName != "" && b.Match.BoardName != "" {
		if !patternsOverlap(a.Match.BoardName, b.Match.BoardName) {
			return false
		}
	}

	// Check product_family compatibility
	if a.Match.ProductFamily != "" && b.Match.ProductFamily != "" {
		if !patternsOverlap(a.Match.ProductFamily, b.Match.ProductFamily) {
			return false
		}
	}

	return true
}

func vendorsOverlap(v1, v2 string) bool {
	norm1 := strings.ToLower(strings.Trim(v1, "* \t"))
	norm2 := strings.ToLower(strings.Trim(v2, "* \t"))
	if norm1 == "" || norm2 == "" {
		return true
	}
	return strings.Contains(norm1, norm2) || strings.Contains(norm2, norm1)
}

func patternsOverlap(p1, p2 string) bool {
	s1 := strings.ToLower(strings.Trim(p1, "* \t"))
	s2 := strings.ToLower(strings.Trim(p2, "* \t"))
	if s1 == "" || s2 == "" {
		return true
	}
	return strings.Contains(s1, s2) || strings.Contains(s2, s1)
}

// CheckConflicts validates all definitions to ensure that files matching
// the same machine do not write the same target with different content
// unless explicitly declared as an override.
func CheckConflicts(defs []QuirkDef, repoRoot string) []Conflict {
	var conflicts []Conflict

	// Resolve file contents for comparison
	getFileContent := func(f QuirkFile) string {
		if f.Content != "" {
			return strings.TrimSpace(f.Content)
		}
		if f.Source != "" {
			// Try relative to repoRoot
			candidates := []string{
				filepath.Join(repoRoot, f.Source),
				filepath.Join(repoRoot, "..", f.Source),
				f.Source,
			}
			for _, c := range candidates {
				if data, err := os.ReadFile(c); err == nil {
					return strings.TrimSpace(string(data))
				}
			}
			// If file cannot be read right now, use source path as surrogate
			return "source:" + f.Source
		}
		return ""
	}

	for i := 0; i < len(defs); i++ {
		for j := i + 1; j < len(defs); j++ {
			a := defs[i]
			b := defs[j]

			if !CanCoexist(a, b) {
				continue
			}

			// Check each file in a against b
			for _, fileA := range a.Files {
				for _, fileB := range b.Files {
					if fileA.Target != fileB.Target {
						continue
					}

					// Targets match: check if declared override
					isOverride := false
					if fileA.Override || fileB.Override {
						isOverride = true
					}
					for _, ov := range a.Overrides {
						if ov == b.ID {
							isOverride = true
						}
					}
					for _, ov := range b.Overrides {
						if ov == a.ID {
							isOverride = true
						}
					}

					if isOverride {
						continue
					}

					// Compare contents
					contentA := getFileContent(fileA)
					contentB := getFileContent(fileB)
					if contentA == contentB && contentA != "" {
						continue
					}

					conflicts = append(conflicts, Conflict{
						Target: fileA.Target,
						DefA:   a.ID,
						DefB:   b.ID,
						Reason: fmt.Sprintf("%s (layer %s) and %s (layer %s) write differing content to %s without declared override",
							a.ID, a.Layer, b.ID, b.Layer, fileA.Target),
					})
				}
			}
		}
	}

	return conflicts
}
