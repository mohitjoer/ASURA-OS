package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"asura/hardware/probe"
	"asura/hardware/quirks"
)

func TestHardwareQuirks_LayeredMatching(t *testing.T) {
	modelsDir, componentsDir := resolveHardwareDirs()
	allDefs, err := quirks.LoadQuirks(modelsDir, componentsDir)
	if err != nil {
		t.Fatalf("failed to load quirks: %v", err)
	}

	// 1. ASUS TUF FX507ZC4 machine
	tufDMI := probe.Fields{
		SysVendor:     "ASUSTeK COMPUTER INC.",
		ProductName:   "ASUS TUF Gaming F15 FX507ZC4_FX507ZC4",
		ProductFamily: "ASUS TUF Gaming F15",
		BoardName:     "FX507ZC4",
		ChassisType:   "notebook",
	}
	tufPCI := []probe.Device{
		{Vendor: "0x10de", Device: "0x25a2"},
	}

	var matchedTuf []quirks.MatchedLayer
	for _, d := range allDefs {
		if ok, reasons := quirks.MatchQuirk(d, tufDMI, tufPCI, nil, probe.ACPIState{}, nil); ok {
			matchedTuf = append(matchedTuf, quirks.MatchedLayer{
				Def:     d,
				Layer:   d.Layer,
				Reasons: reasons,
			})
		}
	}

	if len(matchedTuf) < 4 {
		t.Fatalf("expected at least 4 layers matching TUF machine, got %d", len(matchedTuf))
	}

	files, overridden, pkgs, params, _ := quirks.ResolvePlan(matchedTuf)
	if len(files) == 0 {
		t.Errorf("expected resolved files, got 0")
	}
	_ = overridden
	_ = pkgs
	_ = params

	// 2. Lenovo ThinkPad T14
	tpDMI := probe.Fields{
		SysVendor:     "LENOVO",
		ProductName:   "ThinkPad T14 Gen 2",
		ProductFamily: "ThinkPad T14",
		ChassisType:   "notebook",
	}
	var matchedTP []quirks.MatchedLayer
	for _, d := range allDefs {
		if ok, reasons := quirks.MatchQuirk(d, tpDMI, nil, nil, probe.ACPIState{}, nil); ok {
			matchedTP = append(matchedTP, quirks.MatchedLayer{
				Def:     d,
				Layer:   d.Layer,
				Reasons: reasons,
			})
		}
	}
	hasGeneric := false
	hasThinkPad := false
	for _, m := range matchedTP {
		if m.Def.ID == "generic-laptop" {
			hasGeneric = true
		}
		if m.Def.ID == "lenovo-thinkpad" {
			hasThinkPad = true
		}
	}
	if !hasGeneric || !hasThinkPad {
		t.Errorf("ThinkPad matched layers missing expected generic/thinkpad: %+v", matchedTP)
	}

	// 3. Dell XPS
	xpsDMI := probe.Fields{
		SysVendor:     "Dell Inc.",
		ProductName:   "XPS 13 9310",
		ProductFamily: "XPS",
		ChassisType:   "notebook",
	}
	var matchedXPS []quirks.MatchedLayer
	for _, d := range allDefs {
		if ok, reasons := quirks.MatchQuirk(d, xpsDMI, nil, nil, probe.ACPIState{}, nil); ok {
			matchedXPS = append(matchedXPS, quirks.MatchedLayer{
				Def:     d,
				Layer:   d.Layer,
				Reasons: reasons,
			})
		}
	}
	hasDell := false
	hasXPS := false
	for _, m := range matchedXPS {
		if m.Def.ID == "dell-laptop" {
			hasDell = true
		}
		if m.Def.ID == "dell-xps" {
			hasXPS = true
		}
	}
	if !hasDell || !hasXPS {
		t.Errorf("XPS matched layers missing dell-laptop or dell-xps: %+v", matchedXPS)
	}

	// 4. Desktop ATX Tower
	desktopDMI := probe.Fields{
		SysVendor:     "Micro-Star International Co., Ltd.",
		ProductName:   "MS-7C37",
		ProductFamily: "Custom Desktop",
		ChassisType:   "desktop",
	}
	var matchedDesktop []quirks.MatchedLayer
	for _, d := range allDefs {
		if ok, _ := quirks.MatchQuirk(d, desktopDMI, nil, nil, probe.ACPIState{}, nil); ok {
			// Components like nvidia-gpu without chassis restriction might match if PCI given,
			// but laptop models must not match a desktop.
			if d.Layer < quirks.LayerComponent {
				matchedDesktop = append(matchedDesktop, quirks.MatchedLayer{Def: d})
			}
		}
	}
	if len(matchedDesktop) != 0 {
		t.Errorf("desktop machine matched laptop layers: %+v, want 0", matchedDesktop)
	}
}

func TestGrubKernelParameters(t *testing.T) {
	tempDir := t.TempDir()
	grubPath := filepath.Join(tempDir, "grub")

	initial := `GRUB_DEFAULT=0
GRUB_TIMEOUT=5
GRUB_DISTRIBUTOR="Asura"
GRUB_CMDLINE_LINUX_DEFAULT="loglevel=4 quiet"
`
	if err := os.WriteFile(grubPath, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	// Add params
	params := []string{"nvidia-drm.modeset=1", "mem_sleep_default=deep"}
	data, _ := os.ReadFile(grubPath)
	updated := updateGrubCmdlineParams(string(data), params, true)
	_ = os.WriteFile(grubPath, []byte(updated), 0644)

	afterAdd, _ := os.ReadFile(grubPath)
	if !strings.Contains(string(afterAdd), "nvidia-drm.modeset=1") || !strings.Contains(string(afterAdd), "mem_sleep_default=deep") {
		t.Errorf("grub cmdline missing params after add: %s", string(afterAdd))
	}

	// Remove params
	data, _ = os.ReadFile(grubPath)
	updated = updateGrubCmdlineParams(string(data), params, false)
	_ = os.WriteFile(grubPath, []byte(updated), 0644)

	afterRemove, _ := os.ReadFile(grubPath)
	if strings.Contains(string(afterRemove), "nvidia-drm.modeset=1") || strings.Contains(string(afterRemove), "mem_sleep_default=deep") {
		t.Errorf("grub cmdline still has params after remove: %s", string(afterRemove))
	}
	if !strings.Contains(string(afterRemove), "loglevel=4 quiet") {
		t.Errorf("original params lost: %s", string(afterRemove))
	}
}
