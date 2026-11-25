package quirks

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strings"
)

// ParseQuirk parses a quirk YAML definition.
func ParseQuirk(data []byte, path string) (QuirkDef, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	def := QuirkDef{
		FilePath: path,
	}

	section := ""
	subSection := ""
	matchSubList := ""
	var currentFile *QuirkFile

	flushCurrentFile := func() {
		if currentFile != nil && currentFile.Target != "" {
			def.Files = append(def.Files, *currentFile)
			currentFile = nil
		}
	}

	inMultilineContent := false
	var contentBuilder strings.Builder

	for scanner.Scan() {
		rawLine := scanner.Text()
		trimmed := strings.TrimSpace(rawLine)

		if inMultilineContent {
			indent := len(rawLine) - len(strings.TrimLeft(rawLine, " "))
			if indent >= 8 && trimmed != "" {
				contentBuilder.WriteString(rawLine[8:])
				contentBuilder.WriteString("\n")
				continue
			}
			inMultilineContent = false
			if currentFile != nil {
				currentFile.Content = contentBuilder.String()
			}
			contentBuilder.Reset()
		}

		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		indent := len(rawLine) - len(strings.TrimLeft(rawLine, " "))

		// Top-level sections (indent == 0)
		if indent == 0 {
			flushCurrentFile()
			matchSubList = ""
			key, val, found := strings.Cut(trimmed, ":")
			if !found {
				continue
			}
			key = strings.TrimSpace(key)
			val = strings.Trim(strings.TrimSpace(val), "\"'")

			section = key
			subSection = ""
			switch key {
			case "id":
				def.ID = val
			case "name":
				def.Name = val
			case "description":
				def.Description = val
			case "layer":
				if l, err := ParseLayer(val); err == nil {
					def.Layer = l
				}
			}
			continue
		}

		// Indented level
		switch section {
		case "match":
			if indent == 2 {
				matchSubList = ""
				key, val, found := strings.Cut(trimmed, ":")
				if found {
					key = strings.TrimSpace(key)
					val = strings.Trim(strings.TrimSpace(val), "\"'")
					switch key {
					case "sys_vendor":
						def.Match.SysVendor = val
					case "product_name":
						def.Match.ProductName = val
					case "product_family":
						def.Match.ProductFamily = val
					case "board_name":
						def.Match.BoardName = val
					case "chassis", "chassis_type":
						def.Match.Chassis = val
					case "gpu":
						def.Match.GPU = val
						if val != "" {
							def.Match.PCI = append(def.Match.PCI, val)
						}
					case "pci":
						if val != "" {
							def.Match.PCI = append(def.Match.PCI, val)
						} else {
							matchSubList = "pci"
						}
					case "usb":
						if val != "" {
							def.Match.USB = append(def.Match.USB, val)
						} else {
							matchSubList = "usb"
						}
					case "module", "modules":
						if val != "" {
							def.Match.Module = append(def.Match.Module, val)
						} else {
							matchSubList = "module"
						}
					case "acpi":
						if val != "" {
							def.Match.ACPI = append(def.Match.ACPI, val)
						} else {
							matchSubList = "acpi"
						}
					}
				}
			} else if indent >= 4 {
				// Sub-list or map under match (e.g. pci: [ ... ] or pci: gpu: "...")
				if strings.HasPrefix(trimmed, "- ") {
					item := strings.Trim(strings.TrimSpace(trimmed[2:]), "\"'")
					if item != "" {
						switch matchSubList {
						case "pci":
							def.Match.PCI = append(def.Match.PCI, item)
						case "usb":
							def.Match.USB = append(def.Match.USB, item)
						case "module":
							def.Match.Module = append(def.Match.Module, item)
						case "acpi":
							def.Match.ACPI = append(def.Match.ACPI, item)
						}
					}
				} else if matchSubList == "pci" {
					k, v, found := strings.Cut(trimmed, ":")
					if found {
						k = strings.TrimSpace(k)
						v = strings.Trim(strings.TrimSpace(v), "\"'")
						if k == "gpu" && v != "" {
							def.Match.GPU = v
							def.Match.PCI = append(def.Match.PCI, v)
						}
					}
				}
			}

		case "packages":
			if strings.HasPrefix(trimmed, "- ") {
				pkg := strings.Trim(strings.TrimSpace(trimmed[2:]), "\"'")
				if pkg != "" {
					def.Packages = append(def.Packages, pkg)
				}
			}

		case "kernel_parameters":
			if strings.HasPrefix(trimmed, "- ") {
				param := strings.Trim(strings.TrimSpace(trimmed[2:]), "\"'")
				if param != "" {
					def.KernelParameters = append(def.KernelParameters, param)
				}
			}

		case "services":
			if strings.HasPrefix(trimmed, "- ") {
				svc := strings.Trim(strings.TrimSpace(trimmed[2:]), "\"'")
				if svc != "" {
					def.Services = append(def.Services, svc)
				}
			}

		case "notes":
			if strings.HasPrefix(trimmed, "- ") {
				note := strings.Trim(strings.TrimSpace(trimmed[2:]), "\"'")
				if note != "" {
					def.Notes = append(def.Notes, note)
				}
			}

		case "overrides":
			if strings.HasPrefix(trimmed, "- ") {
				ov := strings.Trim(strings.TrimSpace(trimmed[2:]), "\"'")
				if ov != "" {
					def.Overrides = append(def.Overrides, ov)
				}
			}

		case "files":
			if indent == 2 {
				flushCurrentFile()
				key, _, _ := strings.Cut(trimmed, ":")
				subSection = strings.TrimSpace(key)
				continue
			}
			if strings.HasPrefix(trimmed, "- target:") {
				flushCurrentFile()
				targetVal := strings.Trim(strings.TrimSpace(strings.TrimPrefix(trimmed, "- target:")), "\"'")
				currentFile = &QuirkFile{
					Kind:   subSection,
					Target: targetVal,
				}
				continue
			}
			if currentFile != nil {
				key, val, found := strings.Cut(trimmed, ":")
				if found {
					key = strings.TrimSpace(key)
					val = strings.TrimSpace(val)
					switch key {
					case "source":
						currentFile.Source = strings.Trim(val, "\"'")
					case "reason":
						currentFile.Reason = strings.Trim(val, "\"'")
					case "override":
						val = strings.Trim(val, "\"'")
						currentFile.Override = val == "true" || val == "yes" || val == "1"
					case "content":
						if val == "|" || val == ">" {
							inMultilineContent = true
							contentBuilder.Reset()
						} else {
							currentFile.Content = strings.Trim(val, "\"'")
						}
					}
				}
			}
		}
	}
	flushCurrentFile()

	if def.ID == "" {
		def.ID = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	}

	// Infer layer if not explicitly declared
	if def.Layer == 0 {
		def.Layer = inferLayer(def, path)
	}

	return def, nil
}

func inferLayer(def QuirkDef, path string) Layer {
	cleanPath := filepath.ToSlash(path)
	if strings.Contains(cleanPath, "/components/") || strings.HasPrefix(cleanPath, "components/") {
		return LayerComponent
	}

	m := def.Match
	hasDMI := m.SysVendor != "" || m.ProductName != "" || m.ProductFamily != "" || m.BoardName != ""
	hasComponentMatch := len(m.PCI) > 0 || len(m.USB) > 0 || len(m.Module) > 0 || len(m.ACPI) > 0

	if hasComponentMatch && !hasDMI && m.Chassis == "" {
		return LayerComponent
	}
	if m.ProductName != "" || m.BoardName != "" {
		return LayerModel
	}
	if m.ProductFamily != "" {
		return LayerFamily
	}
	if m.SysVendor != "" {
		return LayerVendor
	}
	return LayerGeneric
}

// LoadQuirks loads all definitions from both models and components directories.
func LoadQuirks(modelsDir, componentsDir string) ([]QuirkDef, error) {
	var all []QuirkDef
	dirs := []string{modelsDir, componentsDir}

	for _, dir := range dirs {
		if dir == "" {
			continue
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}

		for _, entry := range entries {
			if entry.IsDir() || (!strings.HasSuffix(entry.Name(), ".yaml") && !strings.HasSuffix(entry.Name(), ".yml")) {
				continue
			}
			p := filepath.Join(dir, entry.Name())
			data, err := os.ReadFile(p)
			if err != nil {
				continue
			}
			def, err := ParseQuirk(data, p)
			if err == nil && def.ID != "" {
				all = append(all, def)
			}
		}
	}

	return all, nil
}
