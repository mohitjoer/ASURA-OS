package main

import (
	"fmt"

	"toren/desktop/cmd/toren-hubd/internal/wm"
)

// toren-hubd wm answers what the Hub's compositor control asks. Toren runs one
// compositor, so list has one row and there is no switch to preview.
//
//	wm list             the compositor: name, active, capabilities
//	wm preview <name>   unknown for anything but the running compositor

func runWm(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("wm needs list|preview")
	}
	switch args[0] {
	case "list":
		return wmList()
	case "preview":
		if len(args) < 2 {
			return fmt.Errorf("wm preview needs a name")
		}
		return fmt.Errorf("unknown compositor %q: Toren runs %s only", args[1], wm.ProviderHyprland)
	default:
		return fmt.Errorf("wm needs list|preview")
	}
}

// wmProvider is one row of `wm list`.
type wmProvider struct {
	Name      string   `json:"name"`
	ConfigDir string   `json:"configDir"`
	Installed bool     `json:"installed"`
	Deployed  bool     `json:"deployed"`
	Active    bool     `json:"active"`
	Caps      *wm.Caps `json:"caps,omitempty"`
}

func wmList() error {
	row := wmProvider{
		Name:      wm.ProviderHyprland,
		ConfigDir: wm.ConfigDir(wm.ProviderHyprland),
		Installed: true,
		Deployed:  true,
		Active:    true,
	}
	if caps, err := wm.OpenNamed(wm.ProviderHyprland).Caps(); err == nil {
		row.Caps = &caps
	}
	return printJSON([]wmProvider{row})
}
