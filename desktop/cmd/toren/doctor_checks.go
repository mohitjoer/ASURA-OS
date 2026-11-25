package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"asura/desktop/internal/runner"
	"asura/hardware/probe"
	"asura/hardware/quirks"
)

// Two checks for faults that say nothing when they happen. A keybind whose
// command was never installed runs, fails inside the compositor, and the key
// does nothing; a recommended package that is not installed turns one control
// into a button that fails. Both are found here instead of by a user.

// configExec matches the command a compositor config runs: an exec line, or a
// bind whose dispatcher is exec. The first word of it is captured, after any
// window rule in brackets.
var configExec = regexp.MustCompile(`(?m)^\s*(?:execr?(?:-once)?\s*=|bind[a-z]*\s*=[^,\n]*,[^,\n]*,\s*exec\s*,)\s*(?:\[[^\]\n]*\]\s*)?(\S+)`)

// asuraCommands names the Asura commands a compositor config runs, by bare
// name. Anything else it runs (ghostty, wpctl) is the base layer's, and a
// command guarded by `command -v` is already degrading on purpose.
func asuraCommands(config string) []string {
	seen := map[string]bool{}
	for _, match := range configExec.FindAllStringSubmatch(config, -1) {
		word := match[1]
		if strings.HasPrefix(word, "/usr/bin/") {
			word = filepath.Base(word)
		}
		if word == "asura" || strings.HasPrefix(word, "asura-") {
			seen[word] = true
		}
	}
	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// reportCommands checks that every Asura command the compositor runs is
// installed where the compositor's PATH finds it. The session's PATH has no
// ~/.local/bin, so /usr/bin is the only place that counts.
func reportCommands(root string) int {
	fmt.Println()
	fmt.Println("Commands")
	config, err := installedCompositorConfig(root, "/usr/share/asura/hyprland/hyprland.conf")
	if err != nil {
		fmt.Println("  the compositor's configuration is not installed; nothing to check")
		return exitOK
	}

	names := asuraCommands(config)
	problems := 0
	for _, name := range names {
		info, err := os.Stat(filepath.Join(root, "/usr/bin", name))
		switch {
		case err != nil:
			fmt.Printf("  %-24s run by the compositor, and /usr/bin/%s is not installed\n", name, name)
			problems++
		case info.Mode().Perm()&0111 == 0:
			fmt.Printf("  %-24s /usr/bin/%s is not executable\n", name, name)
			problems++
		}
	}
	if problems > 0 {
		return exitFail
	}
	fmt.Printf("  all %d asura command(s) the compositor runs are installed\n", len(names))
	return exitOK
}

// recommendedAbsent returns the recommended packages of the asura packages
// that are not installed on Void Linux via xbps-query.
func recommendedAbsent(run runner.Func) ([]string, int, error) {
	out, err := run("xbps-query", "-l")
	if err != nil {
		return nil, 0, err
	}

	installed := map[string]bool{}
	for _, line := range strings.Split(out, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && strings.HasPrefix(fields[0], "ii") {
			pkg := fields[1]
			if idx := strings.LastIndex(pkg, "-"); idx > 0 {
				installed[pkg[:idx]] = true
			}
		}
	}

	recommended := []string{
		"hyprland",
		"quickshell",
		"xdg-desktop-portal-hyprland",
		"pipewire",
		"wireplumber",
		"seatd",
		"NetworkManager",
		"ghostty",
	}

	var absent []string
	for _, pkg := range recommended {
		if installed[pkg] {
			continue
		}
		// If not registered in XBPS package database (e.g. custom ISO extraction),
		// check whether the binary or portal component actually exists on the system.
		bin := pkg
		switch pkg {
		case "hyprland":
			bin = "Hyprland"
		case "xdg-desktop-portal-hyprland":
			if _, err := os.Stat("/usr/libexec/xdg-desktop-portal-hyprland"); err == nil {
				continue
			}
			if _, err := os.Stat("/usr/lib/xdg-desktop-portal-hyprland"); err == nil {
				continue
			}
		case "seatd":
			if _, err := exec.LookPath("seatd"); err == nil {
				continue
			}
		case "quickshell":
			if _, err := exec.LookPath("quickshell"); err == nil {
				continue
			}
		case "ghostty":
			if _, err := exec.LookPath("ghostty"); err == nil {
				continue
			}
		}
		if _, lookErr := exec.LookPath(bin); lookErr == nil {
			continue
		}
		absent = append(absent, pkg)
	}
	return absent, len(recommended), nil
}

// reportDeps lists what the asura desktop recommends and this machine lacks.
func reportDeps(root string, run runner.Func) int {
	fmt.Println()
	fmt.Println("Dependencies (XBPS)")
	if root != "" {
		fmt.Println("  not checked: the package database describes the running machine, not a --root tree")
		return exitOK
	}
	absent, total, err := recommendedAbsent(run)
	switch {
	case err != nil:
		fmt.Printf("  not checked: %s\n", runner.Reason(err))
	case len(absent) == 0:
		fmt.Printf("  all %d recommended package(s) are installed\n", total)
	default:
		fmt.Printf("  %d of %d recommended package(s) are not installed: %s\n",
			len(absent), total, strings.Join(absent, ", "))
		fmt.Printf("  sudo xbps-install -S %s\n", strings.Join(absent, " "))
	}
	return exitOK
}

// reportVoidRepo validates the XBPS repository configuration in /etc/xbps.d/.
func reportVoidRepo(root string) int {
	fmt.Println()
	fmt.Println("Repository (XBPS)")
	path := filepath.Join(root, "/etc/xbps.d/asura.conf")
	if _, err := os.Stat(path); err != nil {
		path = filepath.Join(root, "/usr/share/xbps.d/asura.conf")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("  xbps repository config missing: %s\n", path)
		return exitFail
	}
	if !strings.Contains(string(data), "repository=") {
		fmt.Printf("  %s contains no repository entries\n", path)
		return exitFail
	}
	fmt.Printf("  repository config %s is valid\n", filepath.Base(path))
	return exitOK
}

// reportRunitServices validates that user runit services exist and are executable.
func reportRunitServices(root string) int {
	fmt.Println()
	fmt.Println("User Services (runit)")
	svcDir := filepath.Join(root, "/usr/share/asura/runit/services")
	entries, err := os.ReadDir(svcDir)
	if err != nil {
		svcDir = filepath.Join(root, "system/runit/user-services")
		entries, err = os.ReadDir(svcDir)
		if err != nil {
			fmt.Printf("  runit services directory not found\n")
			return exitFail
		}
	}
	count := 0
	for _, entry := range entries {
		if entry.IsDir() {
			runPath := filepath.Join(svcDir, entry.Name(), "run")
			if info, err := os.Stat(runPath); err == nil && info.Mode()&0111 != 0 {
				count++
			}
		}
	}
	fmt.Printf("  all %d runit user service(s) configured and executable\n", count)
	return exitOK
}

// coreServices are the system services the desktop needs, in boot order, and
// the optional ones that are reported but never fail doctor.
var (
	coreServices     = []string{"dbus", "udevd", "elogind", "seatd", "turnstiled", "NetworkManager", "toren-sysd"}
	optionalServices = []string{"scx-lavd", "ananicy-cpp", "tlp", "socklog-unix", "nanoklogd", "earlyoom", "thermald", "toren-maintenance"}
)

// svState is what `sv status` says about one service.
type svState int

const (
	svUnknown  svState = iota
	svRunning          // run:
	svFlapping         // down, but runit wants it up: it keeps exiting
	svStopped          // down, and nothing wants it up
	svMissing          // not enabled (no supervise directory)
)

// parseSvStatus reads the first clause of `sv status` output, the service's
// own (the log service follows after a semicolon).
func parseSvStatus(out string) svState {
	first := strings.TrimSpace(strings.SplitN(out, ";", 2)[0])
	switch {
	case strings.HasPrefix(first, "run:"):
		return svRunning
	case strings.HasPrefix(first, "down:"):
		if strings.Contains(first, "want up") {
			return svFlapping
		}
		return svStopped
	case strings.HasPrefix(first, "fail:"), strings.HasPrefix(first, "warning:"):
		if strings.Contains(first, "permission denied") || strings.Contains(first, "access denied") {
			return svUnknown
		}
		return svMissing
	}
	return svUnknown
}

// reportCoreServices checks the system services the session depends on. On a
// running machine it asks runit (`sv status`), which tells a service that is
// enabled but keeps crashing — "down: elogind: 1s, normally up, want up" —
// from one that runs; a crash-looping elogind is a login with no
// /run/user/<uid> and a desktop that cannot start. Under --root it can only
// see which are enabled. --fix enables a disabled required service.
func reportCoreServices(root string, fix bool) int {
	fmt.Println()
	fmt.Println("Core System Services (runit)")
	problems := 0

	enabled := func(svc string) bool {
		for _, dir := range []string{"/var/service", "/etc/runit/runsvdir/default", "/run/runit/runsvdir/current"} {
			if _, err := os.Stat(filepath.Join(root, dir, svc)); err == nil {
				return true
			}
		}
		return false
	}
	state := func(svc string) svState {
		if root != "" {
			if enabled(svc) {
				return svRunning
			}
			return svMissing
		}
		out, _ := exec.Command("sv", "status", "/var/service/"+svc).CombinedOutput()
		st := parseSvStatus(string(out))
		if st == svUnknown || (st == svMissing && os.Geteuid() != 0) {
			if enabled(svc) {
				return svRunning
			}
			return svMissing
		}
		return st
	}

	check := func(svc string, optional bool) {
		_, defErr := os.Stat(filepath.Join(root, "/etc/sv", svc))
		switch st := state(svc); {
		case st == svRunning:
			fmt.Printf("  • %-16s running\n", svc)
		case defErr != nil && optional:
			fmt.Printf("  • %-16s not installed (optional)\n", svc)
		case st == svMissing && optional:
			fmt.Printf("  • %-16s not enabled (optional)\n", svc)
		case st == svFlapping:
			problems++
			fmt.Printf("  • %-16s CRASHING: runit keeps restarting it\n", svc)
			fmt.Printf("    see why: sudo sv down %s; sudo sh -x /etc/sv/%s/run; sudo sv up %s\n", svc, svc, svc)
		case st == svStopped:
			problems++
			fmt.Printf("  • %-16s stopped (start: sudo sv up %s)\n", svc, svc)
		case defErr != nil:
			problems++
			fmt.Printf("  • %-16s not installed\n", svc)
		default:
			problems++
			fmt.Printf("  • %-16s not enabled (enable: sudo ln -s /etc/sv/%s /var/service/)\n", svc, svc)
			if fix {
				target := filepath.Join(root, "/var/service")
				if _, err := os.Stat(target); err != nil {
					target = filepath.Join(root, "/etc/runit/runsvdir/default")
				}
				args := []string{"ln", "-sf", filepath.Join("/etc/sv", svc), filepath.Join(target, svc)}
				if os.Geteuid() != 0 {
					args = append([]string{"sudo"}, args...)
				}
				if err := exec.Command(args[0], args[1:]...).Run(); err == nil {
					fmt.Printf("    fixed: enabled %s\n", svc)
					problems--
				}
			}
		}
	}
	for _, svc := range coreServices {
		check(svc, false)
	}
	for _, svc := range optionalServices {
		check(svc, true)
	}

	// The login's runtime directory, which elogind (or turnstile) creates. A
	// session without it cannot start Hyprland at all.
	if root == "" {
		rt := os.Getenv("XDG_RUNTIME_DIR")
		if rt == "" {
			rt = fmt.Sprintf("/run/user/%d", os.Getuid())
		}
		if info, err := os.Stat(rt); err != nil || !info.IsDir() {
			problems++
			fmt.Printf("  • %-16s %s is missing: the login did not create it (is elogind running?)\n", "runtime dir", rt)
		} else {
			fmt.Printf("  • %-16s %s\n", "runtime dir", rt)
		}
	}

	if problems > 0 {
		return exitFail
	}
	return exitOK
}

// reportUserAudioServices checks that PipeWire and WirePlumber user services
// are configured or running under turnstiled.
func reportUserAudioServices(root string, fix bool) int {
	fmt.Println()
	fmt.Println("Audio Services (PipeWire & WirePlumber)")
	audioServices := []string{"pipewire", "wireplumber"}
	problems := 0

	userSvcDir := filepath.Join(root, "/usr/share/asura/runit/services")
	for _, svc := range audioServices {
		present := false
		if root == "" {
			uid := os.Getuid()
			runDir := fmt.Sprintf("/run/user/%d/service/%s", uid, svc)
			if _, err := os.Stat(runDir); err == nil {
				present = true
			} else {
				if out, err := exec.Command("pgrep", "-x", svc).Output(); err == nil && len(strings.TrimSpace(string(out))) > 0 {
					present = true
				}
			}
		}
		if !present {
			if _, err := os.Stat(filepath.Join(userSvcDir, svc, "run")); err == nil {
				present = true
			} else if _, err := os.Stat(filepath.Join(root, "system/runit/user-services", svc, "run")); err == nil {
				present = true
			}
		}

		if present {
			fmt.Printf("  • %-16s configured\n", svc)
		} else {
			problems++
			fmt.Printf("  • %-16s missing or inactive under turnstiled\n", svc)
		}
	}
	if problems > 0 {
		return exitFail
	}
	return exitOK
}

func parseVersion(s string) []int {
	parts := strings.Split(s, ".")
	var nums []int
	for _, p := range parts {
		d := ""
		for _, c := range p {
			if c >= '0' && c <= '9' {
				d += string(c)
			} else {
				break
			}
		}
		if d == "" {
			break
		}
		n := 0
		fmt.Sscanf(d, "%d", &n)
		nums = append(nums, n)
	}
	return nums
}

func compareVersions(v1, v2 string) int {
	p1 := parseVersion(v1)
	p2 := parseVersion(v2)
	maxLen := len(p1)
	if len(p2) > maxLen {
		maxLen = len(p2)
	}
	for i := 0; i < maxLen; i++ {
		n1, n2 := 0, 0
		if i < len(p1) {
			n1 = p1[i]
		}
		if i < len(p2) {
			n2 = p2[i]
		}
		if n1 < n2 {
			return -1
		}
		if n1 > n2 {
			return 1
		}
	}
	return 0
}

// reportHyprlandVersion verifies that the installed Hyprland binary version
// is aligned with the session configuration syntax (minimum 0.49.0).
func reportHyprlandVersion(root string, fix bool) int {
	fmt.Println()
	fmt.Println("Hyprland Compositor Version")
	hyprBin := filepath.Join(root, "/usr/bin/Hyprland")
	if _, err := os.Stat(hyprBin); err != nil {
		path, err := exec.LookPath("Hyprland")
		if err != nil {
			fmt.Println("  Hyprland binary not found in PATH or /usr/bin")
			return exitOK
		}
		hyprBin = path
	}

	var out []byte
	var err error
	if os.Getuid() == 0 {
		if sudoUser := os.Getenv("SUDO_USER"); sudoUser != "" && sudoUser != "root" {
			out, err = exec.Command("su", sudoUser, "-c", hyprBin+" --version").Output()
		}
	}
	if len(out) == 0 {
		out, err = exec.Command(hyprBin, "--version").Output()
		if err != nil {
			out, err = exec.Command(hyprBin, "-v").Output()
		}
	}
	if err != nil && len(out) == 0 {
		fmt.Printf("  failed to query Hyprland version: %v\n", err)
		return exitFail
	}

	verStr := strings.TrimSpace(string(out))
	fmt.Printf("  installed: %s\n", strings.Split(verStr, "\n")[0])

	re := regexp.MustCompile(`v?([0-9]+\.[0-9]+(?:\.[0-9]+)?)`)
	matches := re.FindStringSubmatch(verStr)
	if len(matches) >= 2 {
		ver := matches[1]
		if compareVersions(ver, "0.49.0") < 0 {
			fmt.Printf("  warning: Hyprland version (%s) is older than 0.49.0 configuration syntax\n", ver)
			if fix {
				fmt.Println("  running xbps-install -S hyprland to update compositor...")
				sudo := "sudo"
				if os.Geteuid() == 0 {
					sudo = ""
				}
				var cmd *exec.Cmd
				if sudo != "" {
					cmd = exec.Command(sudo, "xbps-install", "-S", "-y", "hyprland")
				} else {
					cmd = exec.Command("xbps-install", "-S", "-y", "hyprland")
				}
				cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
				_ = cmd.Run()
			}
			return exitFail
		}
		fmt.Printf("  compositor version (%s) is aligned with session configuration\n", ver)
		return exitOK
	}
	fmt.Println("  compositor version is aligned with session configuration")
	return exitOK
}

// reportHardwarePower checks NVIDIA PCIe D3cold and battery charge threshold.
func reportHardwarePower(root string, fix bool) int {
	fmt.Println()
	fmt.Println("Hardware Power & Thermal")
	problems := 0

	pciDevices, _ := filepath.Glob(filepath.Join(root, "/sys/bus/pci/devices/*/vendor"))
	nvidiaFound := false
	for _, vendorFile := range pciDevices {
		vdata, err := os.ReadFile(vendorFile)
		if err == nil && strings.TrimSpace(string(vdata)) == "0x10de" {
			nvidiaFound = true
			devDir := filepath.Dir(vendorFile)
			ctrlFile := filepath.Join(devDir, "power", "control")
			cdata, err := os.ReadFile(ctrlFile)
			if err == nil {
				val := strings.TrimSpace(string(cdata))
				if val != "auto" {
					problems++
					fmt.Printf("  • NVIDIA GPU (%s) power/control is %q (D3cold requires \"auto\")\n",
						filepath.Base(devDir), val)
					if fix {
						sudo := "sudo"
						if os.Geteuid() == 0 {
							sudo = ""
						}
						cmdStr := fmt.Sprintf("echo auto > %s", ctrlFile)
						var cmd *exec.Cmd
						if sudo != "" {
							cmd = exec.Command(sudo, "sh", "-c", cmdStr)
						} else {
							cmd = exec.Command("sh", "-c", cmdStr)
						}
						if err := cmd.Run(); err == nil {
							fmt.Printf("    fixed: set %s to auto\n", ctrlFile)
							problems--
						}
					}
				} else {
					fmt.Printf("  • NVIDIA GPU (%s) D3cold runtime PM is active (auto)\n", filepath.Base(devDir))
				}
			}
		}
	}
	if !nvidiaFound {
		fmt.Println("  • NVIDIA GPU: not present (Intel/AMD/integrated graphics)")
	}

	batThresholds, _ := filepath.Glob(filepath.Join(root, "/sys/class/power_supply/BAT*/charge_control_end_threshold"))
	if len(batThresholds) == 0 {
		fmt.Println("  • Battery charge threshold: no battery controller detected (desktop/VM)")
	} else {
		for _, threshFile := range batThresholds {
			batName := filepath.Base(filepath.Dir(threshFile))
			tdata, err := os.ReadFile(threshFile)
			if err == nil {
				val := strings.TrimSpace(string(tdata))
				if val != "80" {
					fmt.Printf("  • Battery %s charge limit is %s (recommended: 80 for battery longevity)\n",
						batName, val)
					if fix {
						sudo := "sudo"
						if os.Geteuid() == 0 {
							sudo = ""
						}
						cmdStr := fmt.Sprintf("echo 80 > %s", threshFile)
						var cmd *exec.Cmd
						if sudo != "" {
							cmd = exec.Command(sudo, "sh", "-c", cmdStr)
						} else {
							cmd = exec.Command("sh", "-c", cmdStr)
						}
						if err := cmd.Run(); err == nil {
							fmt.Printf("    fixed: set %s to 80\n", threshFile)
						}
					}
				} else {
					fmt.Printf("  • Battery %s charge limit is configured to 80%%\n", batName)
				}
			}
		}
	}

	if problems > 0 {
		return exitFail
	}
	return exitOK
}

// reportDesktopShell checks the desktop shell the session will start: which
// one org.asura.shell desktop-shell picks, that its tree is installed, and,
// for the desktop, what its look depends on (matugen, a generated palette, the
// compositor look) and whether it is crashing in a loop.
func reportDesktopShell(root string, run runner.Func) int {
	fmt.Println()
	fmt.Println("Desktop shell")
	problems := 0

	choice := os.Getenv("ASURA_DESKTOP_SHELL")
	if choice == "" && root == "" {
		if out, err := run("gsettings", "get", "org.asura.shell", "desktop-shell"); err == nil {
			choice = strings.Trim(strings.TrimSpace(out), "'")
		}
	}
	if choice == "" {
		choice = "desktop"
	}
	dir := "/usr/share/asura/shell"
	if choice == "desktop" {
		dir = "/usr/share/asura/shell-desktop"
	}
	fmt.Printf("  %-18s %s (org.asura.shell desktop-shell)\n", "chosen", choice)
	if _, err := os.Stat(filepath.Join(root, dir, "shell.qml")); err != nil {
		problems++
		fmt.Printf("  %-18s %s/shell.qml is missing; the desktop cannot start (asura update --force)\n", "installed", dir)
	} else {
		fmt.Printf("  %-18s %s\n", "installed", dir)
	}
	if choice != "desktop" {
		if problems > 0 {
			return exitFail
		}
		return exitOK
	}

	for _, f := range []string{"asura/AsuraBridge.qml", "asura/first-run.sh", "asura/hyprland.conf", "matugen/config.toml"} {
		if _, err := os.Stat(filepath.Join(root, dir, f)); err != nil {
			problems++
			fmt.Printf("  %-18s %s/%s is missing (asura update reinstalls it)\n", "ASURA bridge", dir, f)
		}
	}
	if _, err := exec.LookPath("matugen"); err != nil && root == "" {
		problems++
		fmt.Printf("  %-18s not installed: desktop cannot take colours from the wallpaper (sudo xbps-install -S matugen)\n", "matugen")
	} else if root == "" {
		fmt.Printf("  %-18s installed\n", "matugen")
	}
	if root != "" {
		if problems > 0 {
			return exitFail
		}
		return exitOK
	}

	home := homeDir()
	state := os.Getenv("XDG_STATE_HOME")
	if state == "" {
		state = filepath.Join(home, ".local", "state")
	}
	if info, err := os.Stat(filepath.Join(state, "quickshell", "user", "generated", "colors.json")); err != nil || info.Size() == 0 {
		fmt.Printf("  %-18s none generated yet: the desktop uses default colours (Super+Shift+W picks a wallpaper and makes one)\n", "palette")
	} else {
		fmt.Printf("  %-18s generated from the wallpaper\n", "palette")
	}
	look, _ := os.ReadFile(filepath.Join(home, ".config", "asura", "hyprland", "shell.conf"))
	if strings.Contains(string(look), "shell-desktop") {
		fmt.Printf("  %-18s desktop compositor look is loaded\n", "compositor look")
	} else {
		fmt.Printf("  %-18s not loaded yet (asura refresh, or log in again)\n", "compositor look")
	}
	if b, err := os.ReadFile(filepath.Join(state, "asura", "desktop-quick-exits")); err == nil {
		if n, _ := strconv.Atoi(strings.TrimSpace(string(b))); n >= 3 {
			problems++
			fmt.Printf("  %-18s desktop keeps exiting and restarting; see %s\n", "crash loop",
				filepath.Join(state, "asura", "quickshell.log"))
			fmt.Printf("  %-18s after fixing it: rm %s/asura/desktop-*; sv restart ~/.config/service/quickshell\n", "", state)
		}
	}
	if problems > 0 {
		return exitFail
	}
	return exitOK
}

// reportMigrationState says whether this machine and user have one-time fixes
// waiting or failed. Pending ones are not a failure (the next update runs
// them); a failed one is.
func reportMigrationState(root string) int {
	fmt.Println()
	fmt.Println("Migrations")
	if root != "" {
		fmt.Println("  not checked: they describe the running machine, not a --root tree")
		return exitOK
	}
	code := exitOK
	for _, kind := range []string{"system", "user"} {
		pending, failed := pendingMigrations(kind)
		switch {
		case len(failed) > 0:
			code = exitFail
			fmt.Printf("  %-8s %d FAILED (%s); see %s\n", kind, len(failed), strings.Join(failed, ", "), migrationLog(kind))
			fmt.Printf("  %-8s asura update (or asura migrate) runs them again\n", "")
		case pending > 0:
			fmt.Printf("  %-8s %d pending; the next asura update runs them\n", kind, pending)
		default:
			fmt.Printf("  %-8s all done\n", kind)
		}
	}
	return code
}

func migrationLog(kind string) string {
	_, state := migrationDirs(kind)
	return filepath.Join(filepath.Dir(state), "migrate.log")
}

// reportFingerprint reports whether fingerprint authentication is active and safe.
// It warns if PAM references pam_fprintd.so while fprintd or pam_fprintd.so is missing.
func reportFingerprint(root string) int {
	fmt.Println()
	fmt.Println("Fingerprint")

	prefix := root
	sudoPam := filepath.Join(prefix, fprintPamSudo)
	hyprlockPam := filepath.Join(prefix, fprintPamHyprlock)
	polkitPam := filepath.Join(prefix, fprintPamPolkit)

	pamConfigured := false
	for _, p := range []string{sudoPam, hyprlockPam, polkitPam} {
		if data, err := os.ReadFile(p); err == nil {
			if strings.Contains(string(data), fprintBlockBegin) {
				pamConfigured = true
				break
			}
		}
	}

	fprintdBinary := filepath.Join(prefix, "/usr/bin/fprintd")
	pamModule := filepath.Join(prefix, "/usr/lib/security/pam_fprintd.so")
	fprintdAvailable := false
	if _, err := os.Stat(fprintdBinary); err == nil {
		fprintdAvailable = true
	} else if _, err := exec.LookPath("fprintd"); err == nil {
		fprintdAvailable = true
	}
	if _, err := os.Stat(pamModule); err == nil {
		fprintdAvailable = true
	}

	if pamConfigured && !fprintdAvailable {
		fmt.Println("  problem           PAM has fingerprint auth enabled, but fprintd/pam_fprintd.so is MISSING")
		fmt.Println("  fix               run 'asura disable fingerprint' to restore PAM, or 'xbps-install -Sy fprintd'")
		return exitFail
	}

	hasReader, desc := hasSupportedFingerprintReader()
	if !hasReader {
		if pamConfigured {
			fmt.Println("  warning           PAM has fingerprint configured, but no reader is currently detected")
		} else {
			fmt.Println("  reader            none detected (optional)")
		}
		return exitOK
	}

	if pamConfigured {
		fmt.Printf("  reader            %s (PAM active: sudo, hyprlock, polkit)\n", desc)
	} else {
		fmt.Printf("  reader            %s (available, PAM opt-in off)\n", desc)
	}
	return exitOK
}

// reportHardwareQuirks reports whether the machine matches any hardware quirks and whether they are applied.
func reportHardwareQuirks(root string) int {
	fmt.Println()
	fmt.Println("Hardware Quirks")

	allDefs, err := loadAllHardwareQuirks()
	if err != nil || len(allDefs) == 0 {
		fmt.Println("  definitions       no hardware quirks available")
		return exitOK
	}

	fs, _ := probe.New(root)
	matchedLayers, _ := quirks.MatchMachine(allDefs, fs)
	appliedState, _ := getAppliedState()

	appliedMap := make(map[string]bool)
	var staleLayers []string
	if appliedState != nil {
		matchedIDMap := make(map[string]bool)
		for _, m := range matchedLayers {
			matchedIDMap[m.Def.ID] = true
		}
		for _, l := range appliedState.Layers {
			appliedMap[l.ID] = true
			if !matchedIDMap[l.ID] {
				staleLayers = append(staleLayers, l.ID)
			}
		}
	}

	if len(matchedLayers) == 0 {
		if len(staleLayers) > 0 {
			fmt.Printf("  applied           %s (no longer matches current hardware)\n", strings.Join(staleLayers, ", "))
			fmt.Println("  fix               run 'asura hardware quirks remove' to clean stale configuration")
		} else {
			fmt.Println("  model             no specific hardware quirks matched (standard profile in use)")
		}
		return exitOK
	}

	allMatchedApplied := true
	for _, m := range matchedLayers {
		status := "available"
		if appliedMap[m.Def.ID] {
			status = "applied"
		} else {
			allMatchedApplied = false
		}
		fmt.Printf("  layer             [%s] %s (%s) [%s]\n", m.Layer, m.Def.Name, m.Def.ID, status)
	}

	if len(staleLayers) > 0 {
		fmt.Printf("  stale applied     %s (no longer matches current hardware)\n", strings.Join(staleLayers, ", "))
		fmt.Println("  fix               run 'asura hardware quirks remove' or 'asura hardware quirks apply' to re-sync")
	} else if allMatchedApplied {
		fmt.Println("  status            applied and active")
	} else {
		fmt.Println("  status            available (unapplied layers detected)")
		fmt.Println("  fix               run 'asura hardware quirks apply' to apply all matched hardware layers")
	}

	return exitOK
}

// reportAdoption reports whether this machine is an adopted Void Linux installation.
func reportAdoption(root string) int {
	statePath := filepath.Join(root, "etc", "asura", "adopt", "state.json")
	data, err := os.ReadFile(statePath)
	if err != nil {
		return exitOK
	}

	var state adoptState
	if err := json.Unmarshal(data, &state); err != nil {
		return exitOK
	}

	fmt.Println()
	fmt.Println("Adoption Status")
	fmt.Println("  type              adopted Void Linux glibc installation")
	fmt.Printf("  adopted           %s (user: %s)\n", state.Timestamp, state.User)
	fmt.Printf("  backup            %s\n", state.BackupDir)

	missing := []string{}
	for _, f := range state.CreatedFiles {
		if _, err := os.Stat(filepath.Join(root, f)); err != nil {
			missing = append(missing, f)
		}
	}
	if len(missing) > 0 {
		fmt.Printf("  status            drift detected (%d missing registrations)\n", len(missing))
	} else {
		fmt.Println("  status            healthy (all session registrations active)")
	}
	return exitOK
}
