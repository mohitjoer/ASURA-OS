package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const vmUsage = `Usage: toren vm [list | windows | launch NAME | stop NAME]

Manage virtual machines on Toren OS via toren-vmd and quickemu.

Commands:
  list              list local virtual machines and runtime status
  windows           guided setup and creation of an official Windows 11 VM
  launch NAME       launch a virtual machine
  stop NAME         stop a running virtual machine
`

func init() {
	register("vm", "manage virtual machines, including guided Windows setup", torenVm)
	helpText["vm"] = vmUsage
}

var (
	vmStdinReader = bufio.NewReader(os.Stdin)
	vmRunner      = func(args ...string) error {
		cmd := exec.Command("toren-vmd", args...)
		cmd.Stdout, cmd.Stderr, cmd.Stdin = os.Stdout, os.Stderr, os.Stdin
		return cmd.Run()
	}
)

func torenVm(args []string) int {
	if len(args) == 0 || args[0] == "list" {
		return vmList()
	}
	switch args[0] {
	case "windows":
		return vmWindows(args[1:])
	case "launch":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "toren vm launch: missing VM name")
			fmt.Fprint(os.Stderr, vmUsage)
			return exitUsage
		}
		if err := vmRunner("launch", args[1]); err != nil {
			fmt.Fprintf(os.Stderr, "toren vm launch: %v\n", err)
			return exitFail
		}
		return exitOK
	case "stop":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "toren vm stop: missing VM name")
			fmt.Fprint(os.Stderr, vmUsage)
			return exitUsage
		}
		if err := vmRunner("stop", args[1]); err != nil {
			fmt.Fprintf(os.Stderr, "toren vm stop: %v\n", err)
			return exitFail
		}
		return exitOK
	default:
		fmt.Fprintf(os.Stderr, "toren vm: unknown subcommand %q\n\n", args[0])
		fmt.Fprint(os.Stderr, vmUsage)
		return exitUsage
	}
}

func vmList() int {
	out, err := exec.Command("toren-vmd", "list").Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "toren vm: failed to query VMs: %v\n", err)
		return exitFail
	}
	var vms []struct {
		Name    string `json:"name"`
		OS      string `json:"os"`
		Running bool   `json:"running"`
		Disk    string `json:"disk"`
	}
	if err := json.Unmarshal(out, &vms); err != nil {
		fmt.Println(string(out))
		return exitOK
	}
	if len(vms) == 0 {
		fmt.Println("No virtual machines created yet.")
		fmt.Println("Run 'toren vm windows' to create a Windows VM.")
		return exitOK
	}
	fmt.Printf("%-20s %-12s %-10s %s\n", "NAME", "OS", "STATUS", "DISK")
	fmt.Println(strings.Repeat("-", 60))
	for _, v := range vms {
		st := "stopped"
		if v.Running {
			st = "running"
		}
		fmt.Printf("%-20s %-12s %-10s %s\n", v.Name, v.OS, st, v.Disk)
	}
	return exitOK
}

func vmWindows(args []string) int {
	autoConfirm := false
	for _, a := range args {
		if a == "-y" || a == "--yes" {
			autoConfirm = true
		}
	}

	fmt.Println("╭──────────────────────────────────────────────────────────╮")
	fmt.Println("│           Windows 11 Virtual Machine Setup               │")
	fmt.Println("╰──────────────────────────────────────────────────────────╯")
	fmt.Println("This wizard guides you through downloading an official")
	fmt.Println("Microsoft Windows 11 ISO and configuring a high-performance VM.")
	fmt.Println()
	fmt.Println("Sensible defaults configured for this machine:")
	fmt.Println("  - OS:           Windows 11 (English International)")
	fmt.Println("  - Cores:        4 vCPUs")
	fmt.Println("  - RAM:          8 GB")
	fmt.Println("  - Disk:         64 GB (sparse qcow2, grows as needed)")
	fmt.Println("  - Firmware:     UEFI (OVMF)")
	fmt.Println("  - Security:     TPM 2.0 (swtpm enabled)")
	fmt.Println("  - Graphics:     SPICE (clipboard & display auto-resize)")
	fmt.Println()

	if !autoConfirm {
		fmt.Print("Proceed with downloading Windows 11 ISO (~5.5 GB) and creating VM? [y/N]: ")
		resp, _ := vmStdinReader.ReadString('\n')
		resp = strings.ToLower(strings.TrimSpace(resp))
		if resp != "y" && resp != "yes" {
			fmt.Println("\nAborted. No files were downloaded.")
			return exitOK
		}
	}

	vmsDir := os.Getenv("ASURA_VM_HOME")
	if vmsDir == "" {
		vmsDir = filepath.Join(os.Getenv("HOME"), ".local", "share", "toren", "vm")
	}
	_ = os.MkdirAll(vmsDir, 0o755)

	fmt.Println("\n==> Step 1/2: Fetching Windows 11 configuration & media via quickget...")
	cmd := exec.Command("quickget", "windows", "11")
	cmd.Dir = vmsDir
	cmd.Stdout, cmd.Stderr, cmd.Stdin = os.Stdout, os.Stderr, os.Stdin
	if err := cmd.Run(); err != nil {
		cmd2 := exec.Command("toren-vmd", "create", "windows", "11")
		cmd2.Stdout, cmd2.Stderr, cmd2.Stdin = os.Stdout, os.Stderr, os.Stdin
		if err2 := cmd2.Run(); err2 != nil {
			fmt.Fprintf(os.Stderr, "toren vm windows: failed to download Windows 11: %v\n", err)
			return exitFail
		}
	}

	confPath := filepath.Join(vmsDir, "windows-11.conf")
	if _, err := os.Stat(confPath); err == nil {
		tuneWindowsConf(confPath)
	}

	fmt.Println("\n==> Step 2/2: Windows 11 VM created successfully!")
	fmt.Println("To launch your VM, run:")
	fmt.Println("    toren vm launch windows-11")
	return exitOK
}

func tuneWindowsConf(confPath string) {
	b, err := os.ReadFile(confPath)
	if err != nil {
		return
	}
	content := string(b)
	lines := strings.Split(content, "\n")
	hasTPM := false
	hasCores := false
	hasRAM := false
	for _, l := range lines {
		if strings.HasPrefix(l, "tpm=") {
			hasTPM = true
		}
		if strings.HasPrefix(l, "cpu_cores=") {
			hasCores = true
		}
		if strings.HasPrefix(l, "ram=") {
			hasRAM = true
		}
	}
	var add []string
	if !hasTPM {
		add = append(add, `tpm="on"`)
	}
	if !hasCores {
		add = append(add, `cpu_cores="4"`)
	}
	if !hasRAM {
		add = append(add, `ram="8G"`)
	}
	if len(add) > 0 {
		content = content + "\n" + strings.Join(add, "\n") + "\n"
		_ = os.WriteFile(confPath, []byte(content), 0o644)
	}
}
