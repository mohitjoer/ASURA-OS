package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const kernelUsage = `Usage: toren kernel [list | set NAME | remove NAME]

Manages Void Linux kernel series on Toren OS.

Commands:
  list              list installed, running, and available Void kernel series
  set NAME          install kernel NAME, make it default in GRUB, and keep fallback
  remove NAME       remove an installed kernel (refuses running or last installed)

Supported Void kernel series:
  linux             current stable Void kernel
  linux-lts         long-term support kernel
  linux-mainline    mainline development kernel
`

var (
	kernelPrivileged = runPrivileged
	kernelUnameR     = func() (string, error) {
		out, err := exec.Command("uname", "-r").Output()
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(out)), nil
	}
	kernelXbpsQuery = func(pkg string) bool {
		return exec.Command("xbps-query", pkg).Run() == nil
	}
	kernelBootDir         = "/boot"
	supportedKernelSeries = map[string]bool{
		"linux":          true,
		"linux-lts":      true,
		"linux-mainline": true,
	}
)

func init() {
	register("kernel", "manage Void Linux kernel series (list, set, remove)", torenKernel)
	helpText["kernel"] = kernelUsage
}

func torenKernel(args []string) int {
	if len(args) == 0 || args[0] == "list" {
		return kernelList()
	}
	switch args[0] {
	case "set":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "toren kernel set: missing kernel name")
			fmt.Fprint(os.Stderr, kernelUsage)
			return exitUsage
		}
		return kernelSet(args[1])
	case "remove", "rm":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "toren kernel remove: missing kernel name")
			fmt.Fprint(os.Stderr, kernelUsage)
			return exitUsage
		}
		return kernelRemove(args[1])
	default:
		fmt.Fprintf(os.Stderr, "toren kernel: unknown subcommand %q\n\n", args[0])
		fmt.Fprint(os.Stderr, kernelUsage)
		return exitUsage
	}
}

func getInstalledKernels(bootDir string) []string {
	var installed []string
	entries, err := os.ReadDir(bootDir)
	if err != nil {
		return installed
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "vmlinuz-") {
			installed = append(installed, strings.TrimPrefix(e.Name(), "vmlinuz-"))
		}
	}
	return installed
}

func kernelList() int {
	running, err := kernelUnameR()
	if err != nil {
		running = "unknown"
	}

	installed := getInstalledKernels(kernelBootDir)

	fmt.Println("Void Linux Kernel Series:")
	for _, series := range []string{"linux", "linux-lts", "linux-mainline"} {
		status := "available"
		if kernelXbpsQuery(series) {
			status = "installed"
		}
		fmt.Printf("  %-16s [%s]\n", series, status)
	}

	fmt.Println("\nInstalled Kernels (/boot):")
	if len(installed) == 0 {
		fmt.Println("  (none found in /boot)")
	} else {
		for _, k := range installed {
			extra := ""
			if k == running {
				extra = " (running)"
			}
			fmt.Printf("  %-24s%s\n", k, extra)
		}
	}
	return exitOK
}

func hasNvidiaDriver() bool {
	return kernelXbpsQuery("nvidia") || kernelXbpsQuery("nvidia-dkms") || kernelXbpsQuery("nvidia-open")
}

func kernelSet(name string) int {
	name = strings.TrimSpace(name)
	if !supportedKernelSeries[name] && !strings.HasPrefix(name, "linux") {
		fmt.Fprintf(os.Stderr, "toren kernel set: unsupported kernel series %q (expected linux, linux-lts, or linux-mainline)\n", name)
		return exitFail
	}

	pkgs := []string{name}
	if hasNvidiaDriver() {
		headersPkg := name + "-headers"
		fmt.Printf("toren kernel set: NVIDIA driver detected, verifying matching headers package %s...\n", headersPkg)
		if err := exec.Command("xbps-query", "-R", headersPkg).Run(); err != nil {
			fmt.Fprintf(os.Stderr, "toren kernel set: matching kernel headers %s not found in repositories. Refusing to switch kernel to prevent NVIDIA build failure.\n", headersPkg)
			return exitFail
		}
		pkgs = append(pkgs, headersPkg)
	}

	fmt.Printf("toren kernel set: installing %s...\n", strings.Join(pkgs, ", "))
	installArgs := append([]string{"xbps-install", "-Sy"}, pkgs...)
	if err := kernelPrivileged(installArgs...); err != nil {
		fmt.Fprintf(os.Stderr, "toren kernel set: xbps-install failed: %v\n", err)
		return exitFail
	}

	if err := setGrubDefaultKernel(name); err != nil {
		fmt.Fprintf(os.Stderr, "toren kernel set: warning: configuring GRUB default: %v\n", err)
	}

	fmt.Println("toren kernel set: regenerating GRUB configuration...")
	if err := kernelPrivileged("grub-mkconfig", "-o", "/boot/grub/grub.cfg"); err != nil {
		fmt.Fprintf(os.Stderr, "toren kernel set: grub-mkconfig failed: %v\n", err)
		return exitFail
	}

	fmt.Printf("\n✓ Kernel %s successfully installed and set as GRUB default.\n", name)
	fmt.Println("  The previous kernel remains available as a fallback in the GRUB boot menu.")
	fmt.Println("  Please reboot your computer to boot into the new kernel.")
	return exitOK
}

func setGrubDefaultKernel(name string) error {
	grubDefaultFile := "/etc/default/grub"
	b, err := os.ReadFile(grubDefaultFile)
	if err != nil {
		return err
	}
	content := string(b)
	if !strings.Contains(content, "GRUB_DEFAULT=") {
		content += "\nGRUB_DEFAULT=0\n"
	}
	return os.WriteFile(grubDefaultFile, []byte(content), 0o644)
}

func kernelRemove(name string) int {
	name = strings.TrimSpace(name)
	running, _ := kernelUnameR()

	if running != "" && (name == running || strings.Contains(running, strings.TrimPrefix(name, "linux-"))) {
		fmt.Fprintf(os.Stderr, "toren kernel remove: cannot remove running kernel %q (%s). Reboot into another kernel first.\n", name, running)
		return exitFail
	}

	installed := getInstalledKernels(kernelBootDir)
	if len(installed) <= 1 {
		fmt.Fprintf(os.Stderr, "toren kernel remove: cannot remove %q: it is the only installed kernel on this machine.\n", name)
		return exitFail
	}

	pkgs := []string{name}
	if kernelXbpsQuery(name + "-headers") {
		pkgs = append(pkgs, name+"-headers")
	}

	fmt.Printf("toren kernel remove: removing %s...\n", strings.Join(pkgs, ", "))
	removeArgs := append([]string{"xbps-remove", "-Ry"}, pkgs...)
	if err := kernelPrivileged(removeArgs...); err != nil {
		fmt.Fprintf(os.Stderr, "toren kernel remove: xbps-remove failed: %v\n", err)
		return exitFail
	}

	fmt.Println("toren kernel remove: regenerating GRUB configuration...")
	if err := kernelPrivileged("grub-mkconfig", "-o", "/boot/grub/grub.cfg"); err != nil {
		fmt.Fprintf(os.Stderr, "toren kernel remove: grub-mkconfig failed: %v\n", err)
		return exitFail
	}

	fmt.Printf("✓ Kernel %s removed.\n", name)
	return exitOK
}
