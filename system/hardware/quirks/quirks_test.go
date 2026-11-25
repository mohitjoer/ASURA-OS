package quirks_test

import (
	"testing"

	"toren/hardware/quirks"
)

func TestParseQuirkYAML(t *testing.T) {
	yamlContent := []byte(`id: test-hw
name: "Test Hardware"
description: "Test Hardware Description"
layer: vendor

match:
  sys_vendor: "TestVendor"
  chassis: "laptop"
  pci:
    - "10de:*"
  usb:
    - "06cb:00bd"
  module:
    - "test_mod"
  acpi:
    - "platform_profile"

packages:
  - pkg1
  - pkg2

kernel_parameters:
  - "param1=val"

files:
  udev:
    - target: "/etc/udev/rules.d/99-test.rules"
      content: |
        # test rule
        SUBSYSTEM=="power", ACTION=="add"
      reason: "test udev"
      override: true

notes:
  - "test note"
`)

	def, err := quirks.ParseQuirk(yamlContent, "test-hw.yaml")
	if err != nil {
		t.Fatalf("ParseQuirk failed: %v", err)
	}

	if def.ID != "test-hw" {
		t.Errorf("ID = %q, want test-hw", def.ID)
	}
	if def.Layer != quirks.LayerVendor {
		t.Errorf("Layer = %v, want LayerVendor", def.Layer)
	}
	if def.Match.SysVendor != "TestVendor" {
		t.Errorf("SysVendor = %q, want TestVendor", def.Match.SysVendor)
	}
	if len(def.Match.PCI) != 1 || def.Match.PCI[0] != "10de:*" {
		t.Errorf("PCI = %v, want [10de:*]", def.Match.PCI)
	}
	if len(def.Match.USB) != 1 || def.Match.USB[0] != "06cb:00bd" {
		t.Errorf("USB = %v, want [06cb:00bd]", def.Match.USB)
	}
	if len(def.Match.Module) != 1 || def.Match.Module[0] != "test_mod" {
		t.Errorf("Module = %v, want [test_mod]", def.Match.Module)
	}
	if len(def.Match.ACPI) != 1 || def.Match.ACPI[0] != "platform_profile" {
		t.Errorf("ACPI = %v, want [platform_profile]", def.Match.ACPI)
	}
	if len(def.Files) != 1 {
		t.Fatalf("Files len = %d, want 1", len(def.Files))
	}
	if !def.Files[0].Override {
		t.Errorf("Files[0].Override = false, want true")
	}
}

func TestResolvePlan_Overrides(t *testing.T) {
	genericLayer := quirks.MatchedLayer{
		Layer: quirks.LayerGeneric,
		Def: quirks.QuirkDef{
			ID: "generic-laptop",
			Files: []quirks.QuirkFile{
				{Target: "/etc/udev/rules.d/backlight.rules", Content: "generic backlight"},
				{Target: "/etc/udev/rules.d/power.rules", Content: "generic power"},
			},
			Packages:         []string{"pkg-generic"},
			KernelParameters: []string{"param1=1"},
		},
	}

	vendorLayer := quirks.MatchedLayer{
		Layer: quirks.LayerVendor,
		Def: quirks.QuirkDef{
			ID: "asus-laptop",
			Files: []quirks.QuirkFile{
				{Target: "/etc/udev/rules.d/backlight.rules", Content: "asus backlight", Override: true},
			},
			Packages:         []string{"pkg-asus"},
			KernelParameters: []string{"param2=2"},
		},
	}

	matched := []quirks.MatchedLayer{genericLayer, vendorLayer}
	files, overridden, pkgs, params, _ := quirks.ResolvePlan(matched)

	if len(files) != 2 {
		t.Fatalf("expected 2 resolved files, got %d", len(files))
	}
	if len(overridden) != 1 {
		t.Fatalf("expected 1 overridden file, got %d", len(overridden))
	}
	if overridden[0].OriginalQuirkID != "generic-laptop" || overridden[0].OverriddenByID != "asus-laptop" {
		t.Errorf("unexpected overridden info: %+v", overridden[0])
	}

	// Verify backlight target came from asus-laptop
	for _, f := range files {
		if f.Target == "/etc/udev/rules.d/backlight.rules" {
			if f.FromQuirkID != "asus-laptop" || f.Content != "asus backlight" {
				t.Errorf("backlight rule not overridden properly: %+v", f)
			}
		}
	}

	if len(pkgs) != 2 || pkgs[0] != "pkg-generic" || pkgs[1] != "pkg-asus" {
		t.Errorf("unexpected pkgs: %v", pkgs)
	}
	if len(params) != 2 || params[0] != "param1=1" || params[1] != "param2=2" {
		t.Errorf("unexpected params: %v", params)
	}
}
