package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

const gpuUsage = `Usage: toren gpu <verb> [arguments...]

Verbs:
  status                   report detected GPUs, drivers, and runtime power status
  switch [mode]            switch GPU mode (hybrid | integrated | discrete)
  run <cmd...>             execute command offloaded to discrete NVIDIA/AMD GPU
`

var gpuPrivileged = runPrivileged

func init() {
	register("gpu", "inspect GPUs and run commands offloaded to discrete graphics", torenGpu)
}

type GPUInfo struct {
	Address       string `json:"address"`
	Vendor        string `json:"vendor"`
	Model         string `json:"model"`
	Driver        string `json:"driver"`
	RuntimeStatus string `json:"runtime_status,omitempty"`
	PowerControl  string `json:"power_control,omitempty"`
	PowerState    string `json:"power_state,omitempty"`
}

func torenGpu(args []string) int {
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, gpuUsage)
		return exitUsage
	}

	switch args[0] {
	case "status":
		return runGpuStatus(args[1:])
	case "switch":
		return runGpuSwitch(args[1:])
	case "run":
		return runGpuOffload(args[1:])
	case "-h", "--help", "help":
		fmt.Fprint(os.Stdout, gpuUsage)
		return exitOK
	default:
		fmt.Fprintf(os.Stderr, "toren gpu: unknown verb %q\n\n%s", args[0], gpuUsage)
		return exitUsage
	}
}

func scanGPUs() []GPUInfo {
	var gpus []GPUInfo
	entries, err := filepath.Glob("/sys/bus/pci/devices/*")
	if err != nil {
		return gpus
	}

	for _, entry := range entries {
		classData, err := os.ReadFile(filepath.Join(entry, "class"))
		if err != nil {
			continue
		}
		class := strings.TrimSpace(string(classData))
		// Class 0x030000 is VGA compatible, 0x030200 is 3D controller (NVIDIA Optimus / AMD dGPU)
		if !strings.HasPrefix(class, "0x0300") && !strings.HasPrefix(class, "0x0302") {
			continue
		}

		addr := filepath.Base(entry)
		vendorID, _ := os.ReadFile(filepath.Join(entry, "vendor"))
		driverLink, _ := os.Readlink(filepath.Join(entry, "driver"))
		runtimeStatus, _ := os.ReadFile(filepath.Join(entry, "power/runtime_status"))
		powerControl, _ := os.ReadFile(filepath.Join(entry, "power/control"))

		v := strings.TrimSpace(string(vendorID))
		vendorName := "Unknown"
		switch v {
		case "0x10de":
			vendorName = "NVIDIA Corporation"
		case "0x8086":
			vendorName = "Intel Corporation"
		case "0x1002":
			vendorName = "Advanced Micro Devices, Inc. [AMD/ATI]"
		}

		rStatus := strings.TrimSpace(string(runtimeStatus))
		pState := "Unknown"
		if rStatus == "suspended" {
			pState = "D3cold (Suspended, 0W idle)"
		} else if rStatus == "active" {
			pState = "D0 (Active)"
		}

		info := GPUInfo{
			Address:       addr,
			Vendor:        vendorName,
			Driver:        filepath.Base(driverLink),
			RuntimeStatus: rStatus,
			PowerControl:  strings.TrimSpace(string(powerControl)),
			PowerState:    pState,
		}
		gpus = append(gpus, info)
	}
	return gpus
}

func runGpuStatus(args []string) int {
	gpus := scanGPUs()
	if len(args) > 0 && args[0] == "--json" {
		out, _ := json.MarshalIndent(gpus, "", "  ")
		fmt.Println(string(out))
		return exitOK
	}

	if len(gpus) == 0 {
		fmt.Println("No PCI display controllers detected.")
		return exitOK
	}

	for _, gpu := range gpus {
		fmt.Printf("[%s] %s\n", gpu.Address, gpu.Vendor)
		fmt.Printf("  Driver:         %s\n", gpu.Driver)
		if gpu.RuntimeStatus != "" {
			fmt.Printf("  Runtime Power:  %s (control: %s)\n", gpu.RuntimeStatus, gpu.PowerControl)
			fmt.Printf("  Power State:    %s\n", gpu.PowerState)
		}
	}
	return exitOK
}

func runGpuSwitch(args []string) int {
	gpus := scanGPUs()

	// Find discrete GPU
	var dGPU *GPUInfo
	for _, g := range gpus {
		if strings.Contains(g.Vendor, "NVIDIA") || strings.Contains(g.Vendor, "AMD") {
			cp := g
			dGPU = &cp
			break
		}
	}

	if len(args) == 0 {
		if dGPU == nil {
			fmt.Println("Current GPU mode: integrated (no discrete GPU detected)")
			return exitOK
		}
		if dGPU.PowerControl == "on" {
			fmt.Println("Current GPU mode: discrete (dGPU kept awake in D0)")
		} else {
			fmt.Println("Current GPU mode: hybrid (dynamic RTD3 / D3cold power management)")
		}
		return exitOK
	}

	mode := strings.ToLower(args[0])
	switch mode {
	case "hybrid", "auto":
		if dGPU != nil {
			ctrlPath := fmt.Sprintf("/sys/bus/pci/devices/%s/power/control", dGPU.Address)
			_ = gpuPrivileged("sh", "-c", fmt.Sprintf("echo auto > %s", ctrlPath))
			fmt.Println("GPU mode switched to hybrid: dGPU will enter D3cold dynamically when idle.")
		} else {
			fmt.Println("GPU mode set to hybrid.")
		}
	case "discrete", "performance":
		if dGPU != nil {
			ctrlPath := fmt.Sprintf("/sys/bus/pci/devices/%s/power/control", dGPU.Address)
			_ = gpuPrivileged("sh", "-c", fmt.Sprintf("echo on > %s", ctrlPath))
			fmt.Println("GPU mode switched to discrete: dGPU kept awake in D0.")
		} else {
			fmt.Println("GPU mode set to discrete.")
		}
	case "integrated":
		if dGPU != nil {
			ctrlPath := fmt.Sprintf("/sys/bus/pci/devices/%s/power/control", dGPU.Address)
			_ = gpuPrivileged("sh", "-c", fmt.Sprintf("echo auto > %s", ctrlPath))
			fmt.Println("GPU mode switched to integrated: dGPU dynamic power management active.")
		} else {
			fmt.Println("GPU mode set to integrated.")
		}
	default:
		fmt.Fprintf(os.Stderr, "toren gpu switch: unknown mode %q (must be hybrid, integrated, or discrete)\n", mode)
		return exitUsage
	}

	return exitOK
}

func runGpuOffload(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "toren gpu run: command required")
		return exitUsage
	}

	targetBin := args[0]
	path, err := exec.LookPath(targetBin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "toren gpu run: executable %q not found: %v\n", targetBin, err)
		return exitFail
	}

	// Environment variables for NVIDIA Prime render offloading + Mesa DRI PRIME
	env := os.Environ()
	offloadVars := []string{
		"__NV_PRIME_RENDER_OFFLOAD=1",
		"__GLX_VENDOR_LIBRARY_NAME=nvidia",
		"__VK_LAYER_NV_optimus=NVIDIA_only",
		"DRI_PRIME=1",
	}

	for _, v := range offloadVars {
		key := strings.Split(v, "=")[0]
		found := false
		for i, existing := range env {
			if strings.HasPrefix(existing, key+"=") {
				env[i] = v
				found = true
				break
			}
		}
		if !found {
			env = append(env, v)
		}
	}

	return execSubprocess(path, args, env)
}

func execSubprocess(path string, args []string, env []string) int {
	cmd := exec.Command(path, args[1:]...)
	cmd.Env = env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				return status.ExitStatus()
			}
			return exitFail
		}
		fmt.Fprintf(os.Stderr, "toren gpu run error: %v\n", err)
		return exitFail
	}
	return exitOK
}
