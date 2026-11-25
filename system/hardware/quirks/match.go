package quirks

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"toren/hardware/match"
	"toren/hardware/probe"
)

// IsLaptopChassis checks if the chassis type indicates a portable machine.
func IsLaptopChassis(chassis string) bool {
	c := strings.ToLower(strings.TrimSpace(chassis))
	switch c {
	case "laptop", "notebook", "portable", "sub-notebook", "convertible", "detachable":
		return true
	}
	return false
}

// MatchFieldPattern compares a pattern with a probe value.
func MatchFieldPattern(pattern, value string) bool {
	if pattern == "" || pattern == "*" {
		return true
	}
	p := strings.ToLower(strings.TrimSpace(pattern))
	v := strings.ToLower(strings.TrimSpace(value))
	if p == v {
		return true
	}
	if matched, _ := filepath.Match(p, v); matched {
		return true
	}
	if strings.Contains(p, "*") {
		trimmedPat := strings.Trim(p, "*")
		if trimmedPat != "" && strings.Contains(v, trimmedPat) {
			return true
		}
	}
	normP := match.Normalize(pattern)
	normV := match.Normalize(value)
	if normP != "" && normV != "" {
		if normP == normV || strings.HasPrefix(normV, normP) || strings.HasPrefix(normP, normV) {
			return true
		}
	}
	return false
}

func matchDeviceID(pattern, id string) bool {
	pat := strings.ToLower(strings.TrimSpace(pattern))
	val := strings.ToLower(strings.TrimSpace(id))
	if pat == "" || pat == "*" || pat == "*:*" {
		return true
	}
	if pat == val {
		return true
	}
	if matched, _ := filepath.Match(pat, val); matched {
		return true
	}
	// Support vendor-only wildcards like "10de:*"
	if strings.HasSuffix(pat, ":*") {
		vPrefix := strings.TrimSuffix(pat, ":*")
		if strings.HasPrefix(val, vPrefix+":") {
			return true
		}
	}
	return false
}

// MatchQuirk tests a QuirkDef against probe data.
// Returns whether it matched and reasons why it matched.
func MatchQuirk(
	def QuirkDef,
	dmi probe.Fields,
	pciDevices []probe.Device,
	usbDevices []probe.USBDevice,
	acpi probe.ACPIState,
	modules []string,
) (bool, []string) {
	m := def.Match
	var reasons []string

	// Chassis matching
	if m.Chassis != "" {
		req := strings.ToLower(strings.TrimSpace(m.Chassis))
		if req == "laptop" || req == "notebook" {
			if !IsLaptopChassis(dmi.ChassisType) {
				return false, nil
			}
			reasons = append(reasons, fmt.Sprintf("chassis '%s'", dmi.ChassisType))
		} else {
			if !MatchFieldPattern(m.Chassis, dmi.ChassisType) {
				return false, nil
			}
			reasons = append(reasons, fmt.Sprintf("chassis '%s'", dmi.ChassisType))
		}
	}

	// DMI matching
	if m.SysVendor != "" {
		if !MatchFieldPattern(m.SysVendor, dmi.SysVendor) {
			return false, nil
		}
		reasons = append(reasons, fmt.Sprintf("sys_vendor '%s'", dmi.SysVendor))
	}
	if m.ProductName != "" {
		if !MatchFieldPattern(m.ProductName, dmi.ProductName) {
			return false, nil
		}
		reasons = append(reasons, fmt.Sprintf("product_name '%s'", dmi.ProductName))
	}
	if m.ProductFamily != "" {
		if !MatchFieldPattern(m.ProductFamily, dmi.ProductFamily) {
			return false, nil
		}
		reasons = append(reasons, fmt.Sprintf("product_family '%s'", dmi.ProductFamily))
	}
	if m.BoardName != "" {
		if !MatchFieldPattern(m.BoardName, dmi.BoardName) {
			return false, nil
		}
		reasons = append(reasons, fmt.Sprintf("board_name '%s'", dmi.BoardName))
	}

	// PCI matching
	if len(m.PCI) > 0 {
		matchedPattern := ""
		matchedDev := ""
		for _, pat := range m.PCI {
			for _, dev := range pciDevices {
				v := strings.TrimPrefix(strings.ToLower(dev.Vendor), "0x")
				d := strings.TrimPrefix(strings.ToLower(dev.Device), "0x")
				id := v + ":" + d
				if matchDeviceID(pat, id) {
					matchedPattern = pat
					matchedDev = id
					break
				}
			}
			if matchedPattern != "" {
				break
			}
		}
		if matchedPattern == "" {
			return false, nil
		}
		reasons = append(reasons, fmt.Sprintf("pci '%s' (%s)", matchedDev, matchedPattern))
	}

	// USB matching
	if len(m.USB) > 0 {
		matchedPattern := ""
		matchedDev := ""
		for _, pat := range m.USB {
			for _, dev := range usbDevices {
				v := strings.ToLower(dev.IDVendor)
				d := strings.ToLower(dev.IDProduct)
				id := v + ":" + d
				if matchDeviceID(pat, id) {
					matchedPattern = pat
					matchedDev = id
					break
				}
			}
			if matchedPattern != "" {
				break
			}
		}
		if matchedPattern == "" {
			return false, nil
		}
		reasons = append(reasons, fmt.Sprintf("usb '%s' (%s)", matchedDev, matchedPattern))
	}

	// Module matching
	if len(m.Module) > 0 {
		matchedMod := ""
		for _, pat := range m.Module {
			patLower := strings.ToLower(strings.TrimSpace(pat))
			for _, mod := range modules {
				modLower := strings.ToLower(strings.TrimSpace(mod))
				if patLower == modLower || MatchFieldPattern(patLower, modLower) {
					matchedMod = mod
					break
				}
			}
			if matchedMod != "" {
				break
			}
		}
		if matchedMod == "" {
			return false, nil
		}
		reasons = append(reasons, fmt.Sprintf("module '%s'", matchedMod))
	}

	// ACPI matching
	if len(m.ACPI) > 0 {
		matchedACPI := ""
		for _, pat := range m.ACPI {
			patLower := strings.ToLower(strings.TrimSpace(pat))
			if patLower == "platform_profile" && acpi.PlatformProfile != "" {
				matchedACPI = "platform_profile"
				break
			}
			for _, tbl := range acpi.Tables {
				if strings.EqualFold(tbl, pat) {
					matchedACPI = tbl
					break
				}
			}
			if matchedACPI != "" {
				break
			}
		}
		if matchedACPI == "" {
			return false, nil
		}
		reasons = append(reasons, fmt.Sprintf("acpi '%s'", matchedACPI))
	}

	return true, reasons
}

// MatchMachine evaluates all quirk definitions against probe data from fs.
// Returns all matched layers in sorted order: generic -> vendor -> family -> model -> component.
func MatchMachine(defs []QuirkDef, fs *probe.FS) ([]MatchedLayer, error) {
	dmi, err := probe.DMI(fs)
	if err != nil {
		// Non-fatal or empty DMI
		dmi = probe.Fields{}
	}
	pciDevices, _ := probe.PCI(fs, nil)
	usbDevices, _ := probe.USB(fs)
	acpi, _ := probe.ACPI(fs)
	modules, _ := probe.Modules(fs)

	var matched []MatchedLayer
	for _, def := range defs {
		ok, reasons := MatchQuirk(def, dmi, pciDevices, usbDevices, acpi, modules)
		if ok {
			matched = append(matched, MatchedLayer{
				Def:     def,
				Layer:   def.Layer,
				Reasons: reasons,
			})
		}
	}

	// Sort matched layers by Layer (generic < vendor < family < model < component),
	// then by ID for deterministic ordering.
	sort.SliceStable(matched, func(i, j int) bool {
		if matched[i].Layer != matched[j].Layer {
			return matched[i].Layer < matched[j].Layer
		}
		return matched[i].Def.ID < matched[j].Def.ID
	})

	return matched, nil
}
