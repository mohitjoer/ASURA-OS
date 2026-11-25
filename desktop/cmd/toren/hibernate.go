package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
)

// Hibernation support for Toren OS:
//   toren hibernate status
//   toren hibernate setup
//   toren hibernate remove
//
// Safe, opt-in hibernation with Btrfs subvolume (@swap) or ext4 swapfile,
// resume offset calculation, GRUB command line configuration, and dracut resume module.

var (
	hibernateGrubDefault = "/etc/default/grub"
	hibernateFstab       = "/etc/fstab"
	hibernateDracutConf  = "/etc/dracut.conf.d/resume.conf"
	hibernatePowerState  = "/sys/power/state"
	hibernatePowerResume = "/sys/power/resume"
	hibernateLockdown    = "/sys/kernel/security/lockdown"
	hibernateSwaps       = "/proc/swaps"
	hibernateMeminfo     = "/proc/meminfo"
	hibernateReadyMarker = "/run/toren/hibernate.ready"
	hibernatePrivileged  = runPrivileged
)

const hibernateUsage = `Usage: toren hibernate [status | setup | remove]

Configures or removes safe system hibernation.

  status   show whether hibernation is supported, configured, and active
  setup    prepare Btrfs swapfile (@swap), compute resume offset, configure
           GRUB and dracut initramfs (takes a system snapshot first)
  remove   undo swapfile, fstab, GRUB, and dracut configuration
`

func init() {
	register("hibernate", "manage system hibernation, swapfile, and resume configuration", torenHibernate)
}

func torenHibernate(args []string) int {
	if len(args) == 0 || args[0] == "status" {
		return runHibernateStatus()
	}

	switch args[0] {
	case "status":
		return runHibernateStatus()
	case "setup":
		return runHibernateSetup()
	case "remove":
		return runHibernateRemove()
	case "-h", "--help", "help":
		fmt.Print(hibernateUsage)
		return exitOK
	default:
		fmt.Fprintf(os.Stderr, "toren hibernate: unknown verb %q\n\n%s", args[0], hibernateUsage)
		return exitUsage
	}
}

// isHibernationSupported checks if the kernel power state includes "disk".
func isHibernationSupported(stateContent string) bool {
	fields := strings.Fields(stateContent)
	for _, f := range fields {
		if f == "disk" {
			return true
		}
	}
	return false
}

// isSecureBootLockedDown checks whether kernel lockdown is active.
func isSecureBootLockedDown(lockdownContent string) bool {
	return strings.Contains(lockdownContent, "[integrity]") || strings.Contains(lockdownContent, "[confidentiality]")
}

// parseMemTotalMB parses /proc/meminfo MemTotal into megabytes.
func parseMemTotalMB(meminfoContent string) (uint64, error) {
	scanner := bufio.NewScanner(strings.NewReader(meminfoContent))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "MemTotal:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				kb, err := strconv.ParseUint(fields[1], 10, 64)
				if err != nil {
					return 0, err
				}
				return kb / 1024, nil
			}
		}
	}
	return 0, fmt.Errorf("MemTotal line not found in meminfo")
}

// updateGrubCmdline updates GRUB_CMDLINE_LINUX_DEFAULT with resume and resume_offset parameters.
func updateGrubCmdline(content, resumeDev, resumeOffset string) string {
	re := regexp.MustCompile(`(?m)^(GRUB_CMDLINE_LINUX_DEFAULT=")([^"]*)(")`)
	return re.ReplaceAllStringFunc(content, func(m string) string {
		match := re.FindStringSubmatch(m)
		if len(match) < 4 {
			return m
		}
		prefix := match[1]
		cmdline := match[2]
		suffix := match[3]

		// Strip existing resume and resume_offset parameters
		words := strings.Fields(cmdline)
		var kept []string
		for _, w := range words {
			if strings.HasPrefix(w, "resume=") || strings.HasPrefix(w, "resume_offset=") {
				continue
			}
			kept = append(kept, w)
		}

		if resumeDev != "" {
			kept = append(kept, fmt.Sprintf("resume=%s", resumeDev))
			if resumeOffset != "" {
				kept = append(kept, fmt.Sprintf("resume_offset=%s", resumeOffset))
			}
		}

		return prefix + strings.Join(kept, " ") + suffix
	})
}

// parseBtrfsSwapfileOffset parses the output of `btrfs inspect-internal map-swapfile -r`.
func parseBtrfsSwapfileOffset(output string) (string, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if trimmed == "" {
			continue
		}
		fields := strings.Fields(trimmed)
		// Usually the command prints just the offset, or "resume_offset: <offset>"
		last := fields[len(fields)-1]
		if _, err := strconv.ParseUint(last, 10, 64); err == nil {
			return last, nil
		}
	}
	return "", fmt.Errorf("unable to parse resume offset from btrfs map-swapfile output: %q", output)
}

// parseExt4SwapfileOffset parses physical offset from `filefrag -v`.
func parseExt4SwapfileOffset(output string) (string, error) {
	// filefrag -v prints header lines, then extent 0 with physical offset in 4th column
	scanner := bufio.NewScanner(strings.NewReader(output))
	inExtents := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "ext:") || strings.Contains(line, "physical_offset") {
			inExtents = true
			continue
		}
		if inExtents && len(line) > 0 && unicodeIsDigit(rune(line[0])) {
			fields := strings.Fields(line)
			if len(fields) >= 4 {
				offsetStr := strings.TrimSuffix(fields[3], "..")
				if _, err := strconv.ParseUint(offsetStr, 10, 64); err == nil {
					return offsetStr, nil
				}
			}
		}
	}
	return "", fmt.Errorf("unable to parse resume offset from filefrag output: %q", output)
}

func unicodeIsDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func runHibernateStatus() int {
	fmt.Println("Hibernation Status")

	// 1. Kernel state check
	powerStateBytes, err := os.ReadFile(hibernatePowerState)
	if err != nil || !isHibernationSupported(string(powerStateBytes)) {
		fmt.Println("  Kernel support:   no (/sys/power/state lacks 'disk')")
		return exitOK
	}
	fmt.Println("  Kernel support:   yes (/sys/power/state supports 'disk')")

	// 2. Lockdown check
	if data, err := os.ReadFile(hibernateLockdown); err == nil && isSecureBootLockedDown(string(data)) {
		fmt.Println("  Secure Boot:      locked down (hibernation blocked by kernel lockdown)")
		return exitOK
	}

	// 3. Swap check
	swapActive := false
	swapsData, err := os.ReadFile(hibernateSwaps)
	if err == nil {
		lines := strings.Split(string(swapsData), "\n")
		if len(lines) > 1 {
			for _, l := range lines[1:] {
				if strings.TrimSpace(l) != "" {
					swapActive = true
					break
				}
			}
		}
	}
	if swapActive {
		fmt.Println("  Swap active:      yes")
	} else {
		fmt.Println("  Swap active:      no")
	}

	// 4. Swapfile check
	swapfilePath := "/swap/swapfile"
	if fi, err := os.Stat(swapfilePath); err == nil {
		fmt.Printf("  Swapfile:         %s (%d MB)\n", swapfilePath, fi.Size()/(1024*1024))
	} else if fi, err := os.Stat("/swapfile"); err == nil {
		swapfilePath = "/swapfile"
		fmt.Printf("  Swapfile:         %s (%d MB)\n", swapfilePath, fi.Size()/(1024*1024))
	} else {
		fmt.Println("  Swapfile:         none")
	}

	// 5. GRUB configuration check
	hasResumeGrub := false
	if grubData, err := os.ReadFile(hibernateGrubDefault); err == nil {
		if strings.Contains(string(grubData), "resume=") {
			hasResumeGrub = true
		}
	}
	if hasResumeGrub {
		fmt.Println("  GRUB resume:      configured (resume= in GRUB_CMDLINE_LINUX_DEFAULT)")
	} else {
		fmt.Println("  GRUB resume:      not configured")
	}

	// 6. Dracut module check
	hasDracutResume := false
	if dracutData, err := os.ReadFile(hibernateDracutConf); err == nil {
		if strings.Contains(string(dracutData), "resume") {
			hasDracutResume = true
		}
	}
	if hasDracutResume {
		fmt.Println("  Dracut module:    configured (resume in dracut.conf.d)")
	} else {
		fmt.Println("  Dracut module:    not configured")
	}

	// Active resume device in running kernel
	if resumeDevBytes, err := os.ReadFile(hibernatePowerResume); err == nil {
		devStr := strings.TrimSpace(string(resumeDevBytes))
		if devStr != "" && devStr != "0:0" {
			fmt.Printf("  Active resume:    %s\n", devStr)
			fmt.Println("\nHibernation is fully configured and ready.")
			return exitOK
		}
	}

	if hasResumeGrub && hasDracutResume && swapActive {
		fmt.Println("\nHibernation is configured. (If recently setup, reboot to activate resume device).")
		return exitOK
	}

	fmt.Println("\nHibernation is not set up. Run 'sudo toren hibernate setup' to enable.")
	return exitOK
}

func getRootDeviceAndFS() (devPath string, fsType string, isLuks bool, err error) {
	// Find root device from /proc/mounts
	mounts, err := os.ReadFile("/proc/mounts")
	if err != nil {
		return "", "", false, fmt.Errorf("reading /proc/mounts: %w", err)
	}

	scanner := bufio.NewScanner(bytes.NewReader(mounts))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 3 && fields[1] == "/" {
			devPath = fields[0]
			fsType = fields[2]
			break
		}
	}

	if devPath == "" {
		return "", "", false, fmt.Errorf("unable to determine root device")
	}

	if strings.HasPrefix(devPath, "/dev/mapper/") {
		isLuks = true
	}

	return devPath, fsType, isLuks, nil
}

func getDeviceUUID(devPath string) (string, error) {
	out, err := exec.Command("blkid", "-s", "UUID", "-o", "value", devPath).Output()
	if err != nil {
		return "", fmt.Errorf("reading blkid for %s: %w", devPath, err)
	}
	uuid := strings.TrimSpace(string(out))
	if uuid == "" {
		return "", fmt.Errorf("empty UUID for %s", devPath)
	}
	return uuid, nil
}

func runHibernateSetup() int {
	if os.Geteuid() != 0 {
		fmt.Fprintln(os.Stderr, "toren hibernate setup: requires root privileges (run with sudo)")
		return exitFail
	}

	// 1. Refusal checks
	// Check Secure Boot lockdown
	if data, err := os.ReadFile(hibernateLockdown); err == nil && isSecureBootLockedDown(string(data)) {
		fmt.Fprintln(os.Stderr, "toren hibernate: Secure Boot lockdown prevents hibernation")
		return exitFail
	}

	// Check kernel power state
	powerStateBytes, err := os.ReadFile(hibernatePowerState)
	if err != nil || !isHibernationSupported(string(powerStateBytes)) {
		fmt.Fprintln(os.Stderr, "toren hibernate: running kernel does not support hibernation (disk state missing)")
		return exitFail
	}

	// Check root filesystem
	rootDev, fsType, isLuks, err := getRootDeviceAndFS()
	if err != nil {
		fmt.Fprintf(os.Stderr, "toren hibernate: %v\n", err)
		return exitFail
	}

	if fsType != "btrfs" && fsType != "ext4" {
		fmt.Fprintf(os.Stderr, "toren hibernate: root filesystem must be Btrfs or ext4 (found %s)\n", fsType)
		return exitFail
	}

	// Determine required swap size (RAM size in MB)
	memBytes, err := os.ReadFile(hibernateMeminfo)
	if err != nil {
		fmt.Fprintf(os.Stderr, "toren hibernate: reading meminfo: %v\n", err)
		return exitFail
	}
	ramMB, err := parseMemTotalMB(string(memBytes))
	if err != nil {
		fmt.Fprintf(os.Stderr, "toren hibernate: %v\n", err)
		return exitFail
	}
	swapSizeMB := ramMB + 512

	// Check disk free space
	var stat syscall.Statfs_t
	if err := syscall.Statfs("/", &stat); err != nil {
		fmt.Fprintf(os.Stderr, "toren hibernate: checking disk space: %v\n", err)
		return exitFail
	}
	freeMB := (stat.Bavail * uint64(stat.Bsize)) / (1024 * 1024)
	if freeMB < swapSizeMB+2048 {
		fmt.Fprintf(os.Stderr, "toren hibernate: not enough free space on root filesystem (requires %d MB, available %d MB)\n", swapSizeMB+2048, freeMB)
		return exitFail
	}

	// 2. Take a system snapshot before modifying system configuration
	fmt.Println("Taking pre-hibernation system snapshot...")
	snapPath, err := createBtrfsSnapshotAs(snapshotManual)
	if err != nil {
		fmt.Fprintf(os.Stderr, "toren hibernate: warning taking snapshot: %v\n", err)
	} else if snapPath != "" {
		fmt.Printf("System snapshot created: %s\n", snapPath)
	}

	// 3. Create swapfile
	var swapfilePath string
	var resumeOffset string

	if fsType == "btrfs" {
		// Btrfs swapfile requires its own subvolume (@swap) mounted at /swap with NOCOW
		swapfilePath = "/swap/swapfile"
		swapDir := "/swap"

		// Ensure /swap exists and is mounted as @swap if possible
		_ = os.MkdirAll(swapDir, 0o755)

		// Create subvolume if not already mounted
		subvolCreated := false
		if _, err := os.Stat(filepath.Join(swapDir, "swapfile")); os.IsNotExist(err) {
			// Check if /swap is a subvolume or mount point
			if exec.Command("btrfs", "subvolume", "show", swapDir).Run() != nil {
				_ = exec.Command("btrfs", "subvolume", "create", swapDir).Run()
				subvolCreated = true
			}
		}

		// Disable COW and compression
		_ = exec.Command("chattr", "+C", swapDir).Run()
		_ = exec.Command("btrfs", "property", "set", swapDir, "compression", "none").Run()

		// Create swapfile using btrfs mkswapfile if available
		fmt.Printf("Creating %d MB Btrfs swapfile at %s...\n", swapSizeMB, swapfilePath)
		if exec.Command("btrfs", "filesystem", "mkswapfile", "--size", fmt.Sprintf("%dm", swapSizeMB), swapfilePath).Run() != nil {
			// Fallback: truncate, chattr +C, fallocate, chmod 600, mkswap
			_ = exec.Command("truncate", "-s", "0", swapfilePath).Run()
			_ = exec.Command("chattr", "+C", swapfilePath).Run()
			_ = exec.Command("btrfs", "property", "set", swapfilePath, "compression", "none").Run()
			if err := exec.Command("fallocate", "-l", fmt.Sprintf("%dM", swapSizeMB), swapfilePath).Run(); err != nil {
				fmt.Fprintf(os.Stderr, "toren hibernate: creating swapfile: %v\n", err)
				return exitFail
			}
			_ = os.Chmod(swapfilePath, 0o600)
			if err := exec.Command("mkswap", swapfilePath).Run(); err != nil {
				fmt.Fprintf(os.Stderr, "toren hibernate: formatting swap: %v\n", err)
				return exitFail
			}
		}

		// Compute resume offset on Btrfs
		out, err := exec.Command("btrfs", "inspect-internal", "map-swapfile", "-r", swapfilePath).Output()
		if err != nil {
			fmt.Fprintf(os.Stderr, "toren hibernate: computing swapfile offset: %v\n", err)
			return exitFail
		}
		offset, err := parseBtrfsSwapfileOffset(string(out))
		if err != nil {
			fmt.Fprintf(os.Stderr, "toren hibernate: %v\n", err)
			return exitFail
		}
		resumeOffset = offset

		// Update /etc/fstab if /swap is not present
		fstabBytes, _ := os.ReadFile(hibernateFstab)
		fstabContent := string(fstabBytes)
		if subvolCreated && !strings.Contains(fstabContent, "/swap") {
			rootUUID, _ := getDeviceUUID(rootDev)
			if rootUUID != "" {
				fstabContent += fmt.Sprintf("\nUUID=%s /swap btrfs defaults,noatime,subvol=@swap 0 0\n", rootUUID)
			}
		}
		if !strings.Contains(fstabContent, swapfilePath) {
			fstabContent += fmt.Sprintf("%s none swap defaults 0 0\n", swapfilePath)
			_ = os.WriteFile(hibernateFstab, []byte(fstabContent), 0o644)
		}

	} else {
		// ext4 swapfile
		swapfilePath = "/swapfile"
		fmt.Printf("Creating %d MB ext4 swapfile at %s...\n", swapSizeMB, swapfilePath)
		if err := exec.Command("fallocate", "-l", fmt.Sprintf("%dM", swapSizeMB), swapfilePath).Run(); err != nil {
			if err := exec.Command("dd", "if=/dev/zero", "of="+swapfilePath, "bs=1M", fmt.Sprintf("count=%d", swapSizeMB)).Run(); err != nil {
				fmt.Fprintf(os.Stderr, "toren hibernate: creating ext4 swapfile: %v\n", err)
				return exitFail
			}
		}
		_ = os.Chmod(swapfilePath, 0o600)
		_ = exec.Command("mkswap", swapfilePath).Run()

		// Compute offset via filefrag
		out, err := exec.Command("filefrag", "-v", swapfilePath).Output()
		if err != nil {
			fmt.Fprintf(os.Stderr, "toren hibernate: computing ext4 swap offset: %v\n", err)
			return exitFail
		}
		offset, err := parseExt4SwapfileOffset(string(out))
		if err != nil {
			fmt.Fprintf(os.Stderr, "toren hibernate: %v\n", err)
			return exitFail
		}
		resumeOffset = offset

		// Update fstab
		fstabBytes, _ := os.ReadFile(hibernateFstab)
		if !strings.Contains(string(fstabBytes), swapfilePath) {
			newFstab := string(fstabBytes) + fmt.Sprintf("\n%s none swap defaults 0 0\n", swapfilePath)
			_ = os.WriteFile(hibernateFstab, []byte(newFstab), 0o644)
		}
	}

	// Activate swap
	_ = exec.Command("swapon", swapfilePath).Run()

	// 4. Determine resume device UUID / device mapper
	var resumeDevice string
	if isLuks {
		// With LUKS, dracut unlocks the mapper before resuming
		resumeDevice = rootDev
	} else {
		rootUUID, err := getDeviceUUID(rootDev)
		if err != nil {
			fmt.Fprintf(os.Stderr, "toren hibernate: finding root UUID: %v\n", err)
			return exitFail
		}
		resumeDevice = "UUID=" + rootUUID
	}

	// 5. Update GRUB configuration
	fmt.Printf("Configuring GRUB resume=%s resume_offset=%s...\n", resumeDevice, resumeOffset)
	grubData, err := os.ReadFile(hibernateGrubDefault)
	if err != nil {
		fmt.Fprintf(os.Stderr, "toren hibernate: reading %s: %v\n", hibernateGrubDefault, err)
		return exitFail
	}
	newGrubData := updateGrubCmdline(string(grubData), resumeDevice, resumeOffset)
	if err := os.WriteFile(hibernateGrubDefault, []byte(newGrubData), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "toren hibernate: writing %s: %v\n", hibernateGrubDefault, err)
		return exitFail
	}

	// Regenerate GRUB
	fmt.Println("Regenerating GRUB configuration...")
	if err := hibernatePrivileged("update-grub"); err != nil {
		if err := hibernatePrivileged("grub-mkconfig", "-o", "/boot/grub/grub.cfg"); err != nil {
			fmt.Fprintf(os.Stderr, "toren hibernate: updating grub: %v\n", err)
		}
	}

	// 6. Update Dracut configuration
	fmt.Println("Configuring dracut resume module...")
	_ = os.MkdirAll(filepath.Dir(hibernateDracutConf), 0o755)
	dracutContent := `add_dracutmodules+=" resume "` + "\n"
	if err := os.WriteFile(hibernateDracutConf, []byte(dracutContent), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "toren hibernate: writing %s: %v\n", hibernateDracutConf, err)
		return exitFail
	}

	// Rebuild initramfs
	fmt.Println("Rebuilding initramfs with dracut...")
	if err := hibernatePrivileged("dracut", "--force"); err != nil {
		fmt.Fprintf(os.Stderr, "toren hibernate: rebuilding initramfs: %v\n", err)
	}

	// Mark ready
	_ = os.MkdirAll(filepath.Dir(hibernateReadyMarker), 0o755)
	_ = os.WriteFile(hibernateReadyMarker, []byte("ready\n"), 0o644)

	fmt.Println("\nHibernation setup completed successfully!")
	fmt.Println("Reboot your system to activate kernel resume parameters.")
	return exitOK
}

func runHibernateRemove() int {
	if os.Geteuid() != 0 {
		fmt.Fprintln(os.Stderr, "toren hibernate remove: requires root privileges (run with sudo)")
		return exitFail
	}

	fmt.Println("Removing hibernation configuration...")

	// 1. Turn off and remove swapfiles
	for _, sf := range []string{"/swap/swapfile", "/swapfile"} {
		if _, err := os.Stat(sf); err == nil {
			_ = exec.Command("swapoff", sf).Run()
			_ = os.Remove(sf)
		}
	}

	// 2. Remove /swap subvolume if present
	if _, err := os.Stat("/swap"); err == nil {
		_ = exec.Command("umount", "/swap").Run()
		_ = exec.Command("btrfs", "subvolume", "delete", "/swap").Run()
		_ = os.RemoveAll("/swap")
	}

	// 3. Remove swap entries from fstab
	if fstabData, err := os.ReadFile(hibernateFstab); err == nil {
		lines := strings.Split(string(fstabData), "\n")
		var kept []string
		for _, l := range lines {
			if strings.Contains(l, "/swap") || strings.Contains(l, "swapfile") {
				continue
			}
			kept = append(kept, l)
		}
		_ = os.WriteFile(hibernateFstab, []byte(strings.Join(kept, "\n")), 0o644)
	}

	// 4. Remove resume from GRUB
	if grubData, err := os.ReadFile(hibernateGrubDefault); err == nil {
		newGrubData := updateGrubCmdline(string(grubData), "", "")
		_ = os.WriteFile(hibernateGrubDefault, []byte(newGrubData), 0o644)
		_ = hibernatePrivileged("update-grub")
	}

	// 5. Remove dracut configuration
	_ = os.Remove(hibernateDracutConf)
	_ = os.Remove(hibernateReadyMarker)

	// Rebuild initramfs
	fmt.Println("Rebuilding initramfs...")
	_ = hibernatePrivileged("dracut", "--force")

	fmt.Println("Hibernation configuration removed.")
	return exitOK
}
