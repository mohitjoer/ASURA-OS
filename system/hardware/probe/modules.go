package probe

import (
	"strings"
)

// Modules reads the loaded kernel module names from sysfs (module) or fallback proc/modules.
func Modules(fs *FS) ([]string, error) {
	names, err := fs.entries("module")
	if err == nil && len(names) > 0 {
		return names, nil
	}

	// Fallback to modules.txt or proc/modules in fixtures
	for _, rel := range []string{"modules.txt", "proc/modules"} {
		raw, present, err := fs.readString(rel)
		if err == nil && present && raw != "" {
			var mods []string
			for _, line := range strings.Split(raw, "\n") {
				line = strings.TrimSpace(line)
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}
				fields := strings.Fields(line)
				if len(fields) > 0 {
					mods = append(mods, fields[0])
				}
			}
			if len(mods) > 0 {
				return mods, nil
			}
		}
	}

	return nil, nil
}
