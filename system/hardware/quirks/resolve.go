package quirks

// ResolvedFile represents an active file to write to disk.
type ResolvedFile struct {
	QuirkFile
	FromQuirkID string
	FromLayer   Layer
}

// OverriddenFile tracks an earlier layer file that was replaced.
type OverriddenFile struct {
	QuirkFile
	OriginalQuirkID string
	OriginalLayer   Layer
	OverriddenByID  string
	OverriddenBy    Layer
}

// ResolvePlan merges all matched layers into an actionable configuration plan.
// Later layers override earlier layers for the same file target.
func ResolvePlan(matched []MatchedLayer) (files []ResolvedFile, overridden []OverriddenFile, pkgs []string, params []string, svcs []string) {
	fileMap := make(map[string]int) // target -> index in files

	seenPkgs := make(map[string]bool)
	seenParams := make(map[string]bool)
	seenSvcs := make(map[string]bool)

	for _, m := range matched {
		// Collect packages
		for _, pkg := range m.Def.Packages {
			if !seenPkgs[pkg] {
				seenPkgs[pkg] = true
				pkgs = append(pkgs, pkg)
			}
		}

		// Collect kernel parameters
		for _, param := range m.Def.KernelParameters {
			if !seenParams[param] {
				seenParams[param] = true
				params = append(params, param)
			}
		}

		// Collect services
		for _, svc := range m.Def.Services {
			if !seenSvcs[svc] {
				seenSvcs[svc] = true
				svcs = append(svcs, svc)
			}
		}

		// Merge files with later-layer override
		for _, f := range m.Def.Files {
			if idx, exists := fileMap[f.Target]; exists {
				prev := files[idx]
				overridden = append(overridden, OverriddenFile{
					QuirkFile:       prev.QuirkFile,
					OriginalQuirkID: prev.FromQuirkID,
					OriginalLayer:   prev.FromLayer,
					OverriddenByID:  m.Def.ID,
					OverriddenBy:    m.Layer,
				})
				files[idx] = ResolvedFile{
					QuirkFile:   f,
					FromQuirkID: m.Def.ID,
					FromLayer:   m.Layer,
				}
			} else {
				fileMap[f.Target] = len(files)
				files = append(files, ResolvedFile{
					QuirkFile:   f,
					FromQuirkID: m.Def.ID,
					FromLayer:   m.Layer,
				})
			}
		}
	}

	return files, overridden, pkgs, params, svcs
}
