package main

// gpu.go: the `toren-hubd gpu ...` subcommand. The Hub GPU page calls these to
// read the hardware capability verdict, switch graphics mode and tune the card.
//
// mux, apply, hook and vm are the GPU-passthrough stack and the ASUS MUX helper.
// Both need tooling Toren does not ship (Looking Glass, a boot-config writer,
// the gpu-mux helper), so they answer "unavailable" instead of erroring.

import (
	"fmt"
	"os/exec"
)

func runGpu(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("gpu needs a subcommand: caps|mode|tune")
	}
	// detection, the render-mode writer, the MUX switch and the tuner all go through toren-gpu
	if _, err := exec.LookPath(torenGpuBin()); err != nil && args[0] != "apply" && args[0] != "hook" && args[0] != "vm" {
		return unavailable("gpu "+args[0], torenGpuBin()+" is not installed", args[1:])
	}
	switch args[0] {
	case "caps":
		report, err := detectCapability()
		if err != nil {
			return err
		}
		return printJSON(report)
	case "mode":
		return runGpuMode(args[1:])
	case "tune":
		return runGpuTune(args[1:])
	case "mux":
		return runGpuMux(args[1:])
	case "apply", "hook", "vm":
		return unavailable("gpu "+args[0], "GPU passthrough is not set up by Toren yet", args[1:])
	default:
		return fmt.Errorf("unknown gpu subcommand: %s", args[0])
	}
}
