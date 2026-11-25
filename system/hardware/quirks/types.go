package quirks

import (
	"fmt"
	"strings"
)

// Layer represents the precedence hierarchy for hardware quirks:
// LayerGeneric -> LayerVendor -> LayerFamily -> LayerModel -> LayerComponent
type Layer int

const (
	LayerGeneric Layer = iota + 1
	LayerVendor
	LayerFamily
	LayerModel
	LayerComponent
)

func (l Layer) String() string {
	switch l {
	case LayerGeneric:
		return "generic"
	case LayerVendor:
		return "vendor"
	case LayerFamily:
		return "family"
	case LayerModel:
		return "model"
	case LayerComponent:
		return "component"
	default:
		return "unknown"
	}
}

// ParseLayer parses a layer name string into a Layer enum.
func ParseLayer(s string) (Layer, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "generic":
		return LayerGeneric, nil
	case "vendor":
		return LayerVendor, nil
	case "family":
		return LayerFamily, nil
	case "model":
		return LayerModel, nil
	case "component":
		return LayerComponent, nil
	default:
		return 0, fmt.Errorf("unknown layer %q", s)
	}
}

// QuirkMatch defines hardware match criteria across DMI and components.
type QuirkMatch struct {
	SysVendor     string   `yaml:"sys_vendor,omitempty"`
	ProductName   string   `yaml:"product_name,omitempty"`
	ProductFamily string   `yaml:"product_family,omitempty"`
	BoardName     string   `yaml:"board_name,omitempty"`
	Chassis       string   `yaml:"chassis,omitempty"`
	PCI           []string `yaml:"pci,omitempty"`
	USB           []string `yaml:"usb,omitempty"`
	Module        []string `yaml:"module,omitempty"`
	ACPI          []string `yaml:"acpi,omitempty"`
	// Legacy GPU field mapped to PCI:
	GPU string `yaml:"gpu,omitempty"`
}

// QuirkFile represents a configuration file to deploy.
type QuirkFile struct {
	Kind     string `yaml:"kind"`
	Target   string `yaml:"target"`
	Source   string `yaml:"source,omitempty"`
	Content  string `yaml:"content,omitempty"`
	Reason   string `yaml:"reason,omitempty"`
	Override bool   `yaml:"override,omitempty"`
}

// QuirkDef is a complete hardware quirk definition.
type QuirkDef struct {
	ID               string      `yaml:"id"`
	Name             string      `yaml:"name"`
	Description      string      `yaml:"description"`
	Layer            Layer       `yaml:"layer,omitempty"`
	Match            QuirkMatch  `yaml:"match"`
	Packages         []string    `yaml:"packages,omitempty"`
	KernelParameters []string    `yaml:"kernel_parameters,omitempty"`
	Files            []QuirkFile `yaml:"files,omitempty"`
	Services         []string    `yaml:"services,omitempty"`
	Notes            []string    `yaml:"notes,omitempty"`
	Overrides        []string    `yaml:"overrides,omitempty"`
	FilePath         string      `yaml:"-"`
}

// MatchedLayer captures a matched quirk definition and why it matched.
type MatchedLayer struct {
	Def     QuirkDef
	Layer   Layer
	Reasons []string
}

// AppliedLayerInfo is persisted in applied.json to describe applied layers.
type AppliedLayerInfo struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Layer   string   `json:"layer"`
	Reasons []string `json:"reasons"`
}

// AppliedFileInfo records files written to the host for exact rollback.
type AppliedFileInfo struct {
	Target     string `json:"target"`
	Kind       string `json:"kind"`
	FromQuirk  string `json:"from_quirk"`
	Reason     string `json:"reason"`
	BackupPath string `json:"backup_path,omitempty"`
}

// AppliedState records the full state of applied hardware quirks.
type AppliedState struct {
	AppliedAt        string             `json:"applied_at"`
	Layers           []AppliedLayerInfo `json:"layers"`
	Files            []AppliedFileInfo  `json:"files"`
	KernelParameters []string           `json:"kernel_parameters,omitempty"`
	Packages         []string           `json:"packages,omitempty"`
	Services         []string           `json:"services,omitempty"`
}
