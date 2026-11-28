package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"toren/hardware/probe"
	"toren/hardware/quirks"
)

// Hardware quirks engine with layered matching:
// generic -> vendor -> family -> model -> component
//
// Usage:
//   toren hardware quirks list
//   toren hardware quirks show [model-or-component-id]
//   toren hardware quirks apply [id]
//   toren hardware quirks remove [id]

const (
	appliedStateJSONFile   = "/var/lib/toren/hardware/applied.json"
	legacyAppliedStateFile = "/var/lib/toren/quirks.applied"
	appliedBackupsDir      = "/var/lib/toren/hardware/backups"
)

var (
	quirksPrivileged = runPrivileged
)

// resolveHardwareDirs finds the models and components directories.
func resolveHardwareDirs() (modelsDir, componentsDir string) {
	candidates := []string{
		"/usr/share/toren/hardware",
		"system/hardware",
		"../system/hardware",
		"../../system/hardware",
		"../../../system/hardware",
	}
	for _, dir := range candidates {
		m := filepath.Join(dir, "models")
		if info, err := os.Stat(m); err == nil && info.IsDir() {
			c := filepath.Join(dir, "components")
			return m, c
		}
	}
	return "system/hardware/models", "system/hardware/components"
}

func loadAllHardwareQuirks() ([]quirks.QuirkDef, error) {
	modelsDir, componentsDir := resolveHardwareDirs()
	return quirks.LoadQuirks(modelsDir, componentsDir)
}

func getAppliedState() (*quirks.AppliedState, error) {
	data, err := os.ReadFile(appliedStateJSONFile)
	if err == nil {
		var state quirks.AppliedState
		if err := json.Unmarshal(data, &state); err == nil {
			return &state, nil
		}
	}

	// Legacy fallback
	if legData, legErr := os.ReadFile(legacyAppliedStateFile); legErr == nil {
		id := strings.TrimSpace(string(legData))
		if id != "" {
			return &quirks.AppliedState{
				AppliedAt: time.Now().UTC().Format(time.RFC3339),
				Layers: []quirks.AppliedLayerInfo{
					{ID: id, Name: id, Layer: "legacy"},
				},
			}, nil
		}
	}

	return nil, err
}

func getAppliedQuirkIDs() map[string]bool {
	res := make(map[string]bool)
	state, err := getAppliedState()
	if err == nil && state != nil {
		for _, l := range state.Layers {
			res[l.ID] = true
		}
	}
	return res
}

const quirksUsage = `Usage: toren hardware quirks [list | show | apply | remove]

Manages hardware fixes and device tuning with layered precedence.

  list            list available hardware quirks and highlight matches for this machine
  show [id]       display configuration, matched layers, and tuning plan
  apply [id]      apply hardware fixes across all matched layers
  remove [id]     revert all applied hardware fixes
`

func hardwareQuirks(args []string) int {
	if len(args) == 0 || args[0] == "list" {
		return runQuirksList()
	}

	switch args[0] {
	case "list":
		return runQuirksList()
	case "show":
		return runQuirksShow(args[1:])
	case "apply":
		return runQuirksApply(args[1:])
	case "remove":
		return runQuirksRemove(args[1:])
	case "-h", "--help", "help":
		fmt.Print(quirksUsage)
		return exitOK
	default:
		fmt.Fprintf(os.Stderr, "toren hardware quirks: unknown verb %q\n\n%s", args[0], quirksUsage)
		return exitUsage
	}
}

func runQuirksList() int {
	allDefs, err := loadAllHardwareQuirks()
	if err != nil {
		fmt.Fprintf(os.Stderr, "toren hardware quirks: %v\n", err)
		return exitFail
	}

	fs, _ := probe.New(probe.RootDefault)
	matchedLayers, _ := quirks.MatchMachine(allDefs, fs)
	matchedMap := make(map[string]bool)
	for _, m := range matchedLayers {
		matchedMap[m.Def.ID] = true
	}
	appliedMap := getAppliedQuirkIDs()

	modelsDir, componentsDir := resolveHardwareDirs()
	fmt.Printf("Hardware Quirks (%d definition(s) across %s and %s):\n\n", len(allDefs), modelsDir, componentsDir)

	for _, d := range allDefs {
		tag := ""
		isMatched := matchedMap[d.ID]
		isApplied := appliedMap[d.ID]

		if isMatched && isApplied {
			tag = " [MATCHES THIS MACHINE, APPLIED]"
		} else if isMatched {
			tag = " [MATCHES THIS MACHINE]"
		} else if isApplied {
			tag = " [APPLIED]"
		}

		layerPrefix := fmt.Sprintf("[%s]", d.Layer)
		fmt.Printf("  %-12s %-24s %s%s\n", layerPrefix, d.ID, d.Name, tag)
	}

	fmt.Println()
	if len(matchedLayers) > 0 {
		allApplied := true
		for _, m := range matchedLayers {
			if !appliedMap[m.Def.ID] {
				allApplied = false
				break
			}
		}
		if !allApplied {
			fmt.Println("Run 'toren hardware quirks apply' to apply all matched hardware layers.")
		}
	}

	return exitOK
}

func runQuirksShow(args []string) int {
	allDefs, err := loadAllHardwareQuirks()
	if err != nil {
		fmt.Fprintf(os.Stderr, "toren hardware quirks: %v\n", err)
		return exitFail
	}

	if len(args) > 0 {
		targetID := args[0]
		var target *quirks.QuirkDef
		for _, d := range allDefs {
			if d.ID == targetID {
				cp := d
				target = &cp
				break
			}
		}
		if target == nil {
			fmt.Fprintf(os.Stderr, "toren hardware quirks: quirk %q not found\n", targetID)
			return exitFail
		}

		fmt.Printf("ID:          %s\n", target.ID)
		fmt.Printf("Name:        %s\n", target.Name)
		fmt.Printf("Layer:       %s\n", target.Layer)
		fmt.Printf("Description: %s\n", target.Description)
		fmt.Println()

		fmt.Println("Match Criteria:")
		if target.Match.Chassis != "" {
			fmt.Printf("  chassis:        %s\n", target.Match.Chassis)
		}
		if target.Match.SysVendor != "" {
			fmt.Printf("  sys_vendor:     %s\n", target.Match.SysVendor)
		}
		if target.Match.ProductName != "" {
			fmt.Printf("  product_name:   %s\n", target.Match.ProductName)
		}
		if target.Match.ProductFamily != "" {
			fmt.Printf("  product_family: %s\n", target.Match.ProductFamily)
		}
		if target.Match.BoardName != "" {
			fmt.Printf("  board_name:     %s\n", target.Match.BoardName)
		}
		if len(target.Match.PCI) > 0 {
			fmt.Printf("  pci:            %s\n", strings.Join(target.Match.PCI, ", "))
		}
		if len(target.Match.USB) > 0 {
			fmt.Printf("  usb:            %s\n", strings.Join(target.Match.USB, ", "))
		}
		if len(target.Match.Module) > 0 {
			fmt.Printf("  module:         %s\n", strings.Join(target.Match.Module, ", "))
		}
		if len(target.Match.ACPI) > 0 {
			fmt.Printf("  acpi:           %s\n", strings.Join(target.Match.ACPI, ", "))
		}
		fmt.Println()

		if len(target.Packages) > 0 {
			fmt.Println("Packages:")
			for _, p := range target.Packages {
				fmt.Printf("  - %s\n", p)
			}
			fmt.Println()
		}

		if len(target.KernelParameters) > 0 {
			fmt.Println("Kernel Parameters:")
			for _, kp := range target.KernelParameters {
				fmt.Printf("  - %s\n", kp)
			}
			fmt.Println()
		}

		if len(target.Files) > 0 {
			fmt.Println("Configuration Files:")
			for _, f := range target.Files {
				overrideStr := ""
				if f.Override {
					overrideStr = " (override)"
				}
				fmt.Printf("  - %s (%s)%s: %s\n", f.Target, f.Kind, overrideStr, f.Reason)
			}
			fmt.Println()
		}

		if len(target.Notes) > 0 {
			fmt.Println("Hardware Notes:")
			for _, n := range target.Notes {
				fmt.Printf("  * %s\n", n)
			}
		}
		return exitOK
	}

	// Show matched layers for current machine
	fs, _ := probe.New(probe.RootDefault)
	matchedLayers, _ := quirks.MatchMachine(allDefs, fs)
	if len(matchedLayers) == 0 {
		fmt.Println("No hardware quirks match this machine.")
		return exitOK
	}

	files, overridden, pkgs, params, _ := quirks.ResolvePlan(matchedLayers)

	overriddenMap := make(map[string]quirks.OverriddenFile)
	for _, o := range overridden {
		overriddenMap[o.Target] = o
	}

	fmt.Printf("Matched Hardware Layers (%d layer(s) ordered generic -> vendor -> family -> model -> component):\n\n", len(matchedLayers))

	for i, m := range matchedLayers {
		fmt.Printf("[%d/%d] %s layer: %s (%s)\n", i+1, len(matchedLayers), m.Layer, m.Def.Name, m.Def.ID)
		if len(m.Reasons) > 0 {
			fmt.Printf("  Match Reason:      %s\n", strings.Join(m.Reasons, "; "))
		}
		if len(m.Def.Packages) > 0 {
			fmt.Printf("  Packages:          %s\n", strings.Join(m.Def.Packages, ", "))
		}
		if len(m.Def.KernelParameters) > 0 {
			fmt.Printf("  Kernel Parameters: %s\n", strings.Join(m.Def.KernelParameters, ", "))
		}
		if len(m.Def.Files) > 0 {
			fmt.Println("  Files:")
			for _, f := range m.Def.Files {
				if ov, isOv := overriddenMap[f.Target]; isOv && ov.OriginalQuirkID == m.Def.ID {
					fmt.Printf("    - %s (%s) [OVERRIDDEN by %s layer %s]\n", f.Target, f.Kind, ov.OverriddenBy, ov.OverriddenByID)
				} else {
					fmt.Printf("    - %s (%s): %s\n", f.Target, f.Kind, f.Reason)
				}
			}
		}
		fmt.Println()
	}

	fmt.Println("Resolved Active Configuration Plan:")
	if len(pkgs) > 0 {
		fmt.Printf("  Packages to install:       %s\n", strings.Join(pkgs, ", "))
	}
	if len(params) > 0 {
		fmt.Printf("  Kernel parameters to set:  %s\n", strings.Join(params, ", "))
	}
	if len(files) > 0 {
		fmt.Println("  Files to lay down:")
		for _, f := range files {
			fmt.Printf("    - %s (from [%s] %s)\n", f.Target, f.FromLayer, f.FromQuirkID)
		}
	}

	return exitOK
}

func runQuirksApply(args []string) int {
	allDefs, err := loadAllHardwareQuirks()
	if err != nil {
		fmt.Fprintf(os.Stderr, "toren hardware quirks: %v\n", err)
		return exitFail
	}

	var layersToApply []quirks.MatchedLayer

	if len(args) > 0 {
		targetID := args[0]
		for _, d := range allDefs {
			if d.ID == targetID {
				layersToApply = append(layersToApply, quirks.MatchedLayer{
					Def:   d,
					Layer: d.Layer,
				})
				break
			}
		}
		if len(layersToApply) == 0 {
			fmt.Fprintf(os.Stderr, "toren hardware quirks: quirk %q not found\n", targetID)
			return exitFail
		}
	} else {
		fs, _ := probe.New(probe.RootDefault)
		layersToApply, err = quirks.MatchMachine(allDefs, fs)
		if err != nil || len(layersToApply) == 0 {
			fmt.Println("No hardware quirks match this machine.")
			return exitOK
		}
	}

	fmt.Printf("Applying %d hardware quirk layer(s)...\n", len(layersToApply))
	for _, m := range layersToApply {
		fmt.Printf("  + [%s] %s (%s)\n", m.Layer, m.Def.Name, m.Def.ID)
	}

	// Pre-apply snapshot
	if isBtrfsRoot() {
		fmt.Println("Taking pre-quirk system snapshot...")
		if snap, err := createBtrfsSnapshotAs(snapshotManual); err == nil {
			fmt.Printf("Created snapshot %s before applying quirks.\n", snap)
		}
	}

	files, _, pkgs, params, svcs := quirks.ResolvePlan(layersToApply)

	// Install packages
	for _, pkg := range pkgs {
		if !svInstalled(pkg) {
			fmt.Printf("Installing %s...\n", pkg)
			_ = quirksPrivileged("xbps-install", "-Sy", pkg)
		}
	}

	var appliedFiles []quirks.AppliedFileInfo
	_ = os.MkdirAll(appliedBackupsDir, 0755)

	// Deploy resolved files
	for _, rf := range files {
		var content []byte
		if rf.Content != "" {
			content = []byte(rf.Content)
		} else if rf.Source != "" {
			candidates := []string{
				rf.Source,
				filepath.Join("..", rf.Source),
				filepath.Join("../..", rf.Source),
			}
			for _, c := range candidates {
				if data, err := os.ReadFile(c); err == nil {
					content = data
					break
				}
			}
		}

		if len(content) == 0 {
			continue
		}

		backupPath := ""
		if orig, err := os.ReadFile(rf.Target); err == nil {
			// Save backup
			safeName := strings.ReplaceAll(strings.TrimPrefix(rf.Target, "/"), "/", "_")
			backupPath = filepath.Join(appliedBackupsDir, safeName+".bak")
			if _, err := os.Stat(backupPath); os.IsNotExist(err) {
				_ = os.WriteFile(backupPath, orig, 0644)
			}
		}

		_ = os.MkdirAll(filepath.Dir(rf.Target), 0755)
		tmp := filepath.Join(os.TempDir(), "quirk-"+filepath.Base(rf.Target))
		if err := os.WriteFile(tmp, content, 0644); err == nil {
			_ = quirksPrivileged("cp", tmp, rf.Target)
			_ = os.Remove(tmp)
			fmt.Printf("Laid down %s (from [%s] %s): %s\n", rf.Target, rf.FromLayer, rf.FromQuirkID, rf.Reason)
			appliedFiles = append(appliedFiles, quirks.AppliedFileInfo{
				Target:     rf.Target,
				Kind:       rf.Kind,
				FromQuirk:  rf.FromQuirkID,
				Reason:     rf.Reason,
				BackupPath: backupPath,
			})
		}
	}

	// Update kernel parameters in GRUB
	if len(params) > 0 {
		applyGrubKernelParameters(params)
	}

	// Reload udev
	_ = quirksPrivileged("udevadm", "control", "--reload-rules")
	_ = quirksPrivileged("udevadm", "trigger")

	// Save state to /var/lib/toren/hardware/applied.json
	var appliedLayers []quirks.AppliedLayerInfo
	for _, m := range layersToApply {
		appliedLayers = append(appliedLayers, quirks.AppliedLayerInfo{
			ID:      m.Def.ID,
			Name:    m.Def.Name,
			Layer:   m.Layer.String(),
			Reasons: m.Reasons,
		})
	}

	state := quirks.AppliedState{
		AppliedAt:        time.Now().UTC().Format(time.RFC3339),
		Layers:           appliedLayers,
		Files:            appliedFiles,
		KernelParameters: params,
		Packages:         pkgs,
		Services:         svcs,
	}

	_ = os.MkdirAll(filepath.Dir(appliedStateJSONFile), 0755)
	if data, err := json.MarshalIndent(state, "", "  "); err == nil {
		_ = os.WriteFile(appliedStateJSONFile, data, 0644)
	}
	// Write primary model ID to legacy file for backward compatibility
	if len(appliedLayers) > 0 {
		_ = os.WriteFile(legacyAppliedStateFile, []byte(appliedLayers[len(appliedLayers)-1].ID+"\n"), 0644)
	}

	fmt.Println("Successfully applied hardware quirks.")
	return exitOK
}

func runQuirksRemove(_ []string) int {
	state, err := getAppliedState()
	if err != nil || state == nil {
		fmt.Println("No hardware quirks are currently applied.")
		return exitOK
	}

	fmt.Printf("Reverting %d applied hardware quirk layer(s)...\n", len(state.Layers))

	// Revert files
	for _, f := range state.Files {
		if f.BackupPath != "" {
			if _, err := os.Stat(f.BackupPath); err == nil {
				_ = quirksPrivileged("cp", f.BackupPath, f.Target)
				_ = quirksPrivileged("rm", "-f", f.BackupPath)
				fmt.Printf("Restored %s from backup\n", f.Target)
				continue
			}
		}
		// Also check legacy .toren-quirk-bak
		legacyBak := f.Target + ".toren-quirk-bak"
		if _, err := os.Stat(legacyBak); err == nil {
			_ = quirksPrivileged("cp", legacyBak, f.Target)
			_ = quirksPrivileged("rm", "-f", legacyBak)
			fmt.Printf("Restored %s from backup\n", f.Target)
			continue
		}

		_ = quirksPrivileged("rm", "-f", f.Target)
		fmt.Printf("Removed %s\n", f.Target)
	}

	// Revert kernel parameters
	if len(state.KernelParameters) > 0 {
		removeGrubKernelParameters(state.KernelParameters)
	}

	// Reload udev
	_ = quirksPrivileged("udevadm", "control", "--reload-rules")
	_ = quirksPrivileged("udevadm", "trigger")

	// Remove state files
	_ = os.Remove(appliedStateJSONFile)
	_ = os.Remove(legacyAppliedStateFile)

	fmt.Println("Hardware quirks removed.")
	return exitOK
}

func updateGrubCmdlineParams(content string, params []string, add bool) string {
	paramSet := map[string]bool{}
	for _, p := range params {
		paramSet[p] = true
	}

	re := regexp.MustCompile(`(?m)^(GRUB_CMDLINE_LINUX_DEFAULT=")([^"]*)(")`)
	return re.ReplaceAllStringFunc(content, func(m string) string {
		match := re.FindStringSubmatch(m)
		opts := strings.Fields(match[2])
		if add {
			optSet := map[string]bool{}
			for _, opt := range opts {
				optSet[opt] = true
			}
			for _, p := range params {
				if !optSet[p] {
					opts = append(opts, p)
				}
			}
			return match[1] + strings.Join(opts, " ") + match[3]
		}
		// remove
		var filtered []string
		for _, opt := range opts {
			if !paramSet[opt] {
				filtered = append(filtered, opt)
			}
		}
		return match[1] + strings.Join(filtered, " ") + match[3]
	})
}

func applyGrubKernelParameters(params []string) {
	data, err := os.ReadFile(hibernateGrubDefault)
	if err != nil {
		return
	}
	content := string(data)
	updated := updateGrubCmdlineParams(content, params, true)

	if updated != content {
		tmp := filepath.Join(os.TempDir(), "grub-quirks")
		if err := os.WriteFile(tmp, []byte(updated), 0644); err == nil {
			_ = quirksPrivileged("cp", tmp, hibernateGrubDefault)
			_ = os.Remove(tmp)
			_ = quirksPrivileged("update-grub")
		}
	}
}

func removeGrubKernelParameters(params []string) {
	data, err := os.ReadFile(hibernateGrubDefault)
	if err != nil {
		return
	}
	content := string(data)
	updated := updateGrubCmdlineParams(content, params, false)

	if updated != content {
		tmp := filepath.Join(os.TempDir(), "grub-quirks-clean")
		if err := os.WriteFile(tmp, []byte(updated), 0644); err == nil {
			_ = quirksPrivileged("cp", tmp, hibernateGrubDefault)
			_ = os.Remove(tmp)
			_ = quirksPrivileged("update-grub")
		}
	}
}
