package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const powerUsage = `Usage: toren power <verb> [arguments...]

Verbs:
  status                print current power supply state, governor, and profile
  profile [name]        get or set profile (balanced, performance, battery)
  threshold [pct]       get or set battery charge stop threshold (e.g. 80)
  sleep [mode]          get or set system suspend sleep mode (s2idle | deep)
  ac                    switch power settings to AC (balanced/performance)
  bat                   switch power settings to Battery (power-saving, D3cold)
`

var powerPrivileged = runPrivileged

func init() {
	register("power", "manage CPU governors, energy preference, battery thresholds, and sleep modes", torenPower)
}

func torenPower(args []string) int {
	if len(args) == 0 {
		return runPowerStatus()
	}

	switch args[0] {
	case "status":
		return runPowerStatus()
	case "profile":
		if len(args) > 1 {
			return setPowerProfile(args[1])
		}
		return getPowerProfile()
	case "threshold":
		if len(args) > 1 {
			return setBatteryThreshold(args[1])
		}
		return getBatteryThreshold()
	case "sleep":
		if len(args) > 1 {
			return setSleepMode(args[1])
		}
		return getSleepMode()
	case "ac":
		return setPowerProfile("balanced")
	case "bat", "battery":
		return setPowerProfile("battery")
	case "-h", "--help", "help":
		fmt.Fprint(os.Stdout, powerUsage)
		return exitOK
	default:
		fmt.Fprintf(os.Stderr, "toren power: unknown verb %q\n\n%s", args[0], powerUsage)
		return exitUsage
	}
}

func runPowerStatus() int {
	state := map[string]interface{}{}

	// Check power supply online status
	acOnline := false
	supplies, _ := filepath.Glob("/sys/class/power_supply/*/online")
	for _, s := range supplies {
		data, err := os.ReadFile(s)
		if err == nil && strings.TrimSpace(string(data)) == "1" {
			acOnline = true
			break
		}
	}
	state["ac_online"] = acOnline

	// Read CPU governor
	gov, _ := os.ReadFile("/sys/devices/system/cpu/cpu0/cpufreq/scaling_governor")
	state["governor"] = strings.TrimSpace(string(gov))

	// Read CPU EPP if available
	epp, _ := os.ReadFile("/sys/devices/system/cpu/cpu0/cpufreq/energy_performance_preference")
	if len(epp) > 0 {
		state["epp"] = strings.TrimSpace(string(epp))
	}

	// Read GPU runtime status
	nvPower, _ := os.ReadFile("/sys/bus/pci/devices/0000:01:00.0/power/runtime_status")
	if len(nvPower) > 0 {
		state["gpu_runtime_status"] = strings.TrimSpace(string(nvPower))
	}

	// Read sleep mode
	if memSleep, err := os.ReadFile("/sys/power/mem_sleep"); err == nil {
		state["mem_sleep"] = strings.TrimSpace(string(memSleep))
	}

	out, _ := json.MarshalIndent(state, "", "  ")
	fmt.Println(string(out))
	return exitOK
}

func getPowerProfile() int {
	if plat, err := os.ReadFile("/sys/firmware/acpi/platform_profile"); err == nil {
		curPlat := strings.TrimSpace(string(plat))
		switch curPlat {
		case "performance":
			fmt.Println("performance")
			return exitOK
		case "quiet", "low-power":
			fmt.Println("battery")
			return exitOK
		case "balanced":
			fmt.Println("balanced")
			return exitOK
		}
	}
	epp, _ := os.ReadFile("/sys/devices/system/cpu/cpu0/cpufreq/energy_performance_preference")
	cur := strings.TrimSpace(string(epp))
	switch cur {
	case "performance":
		fmt.Println("performance")
	case "power":
		fmt.Println("battery")
	default:
		fmt.Println("balanced")
	}
	return exitOK
}

func setPowerProfile(profile string) int {
	profile = strings.ToLower(strings.TrimSpace(profile))

	var governor string
	var epp string
	var platformProfile string

	switch profile {
	case "performance":
		governor = "performance"
		epp = "performance"
		platformProfile = "performance"
	case "battery", "bat", "powersave":
		governor = "powersave"
		epp = "power"
		platformProfile = "quiet"
	case "balanced", "balance":
		governor = "powersave"
		epp = "balance_performance"
		platformProfile = "balanced"
	default:
		fmt.Fprintf(os.Stderr, "toren power profile: unknown profile %q (choose: balanced, performance, battery)\n", profile)
		return exitUsage
	}

	if applyViaSysd(governor, epp, profile) {
		fmt.Printf("Power profile set to %s (via toren-sysd)\n", profile)
		return exitOK
	}

	applied := false
	cpus, _ := filepath.Glob("/sys/devices/system/cpu/cpu*/cpufreq/scaling_governor")
	for _, f := range cpus {
		if os.WriteFile(f, []byte(governor), 0644) == nil {
			applied = true
		}
	}

	epps, _ := filepath.Glob("/sys/devices/system/cpu/cpu*/cpufreq/energy_performance_preference")
	for _, f := range epps {
		_ = os.WriteFile(f, []byte(epp), 0644)
	}

	if platformProfile != "" {
		if os.WriteFile("/sys/firmware/acpi/platform_profile", []byte(platformProfile), 0644) == nil {
			applied = true
		}
	}

	gpus, _ := filepath.Glob("/sys/bus/pci/devices/*/power/control")
	for _, f := range gpus {
		_ = os.WriteFile(f, []byte("auto"), 0644)
	}

	if applied {
		fmt.Printf("Power profile set to %s\n", profile)
		return exitOK
	}

	fmt.Printf("Power profile requested: %s (requires toren-sysd or root privileges)\n", profile)
	return exitOK
}

func getBatteryThreshold() int {
	nodes := []string{
		"/sys/class/power_supply/BAT0/charge_control_end_threshold",
		"/sys/class/power_supply/BAT1/charge_control_end_threshold",
		"/sys/class/power_supply/BAT0/charge_stop_threshold",
		"/sys/class/power_supply/BAT1/charge_stop_threshold",
		"/sys/devices/platform/lg-laptop/battery_care_limit",
		"/sys/bus/platform/drivers/ideapad_acpi/*/conservation_mode",
	}

	for _, n := range nodes {
		matches, _ := filepath.Glob(n)
		for _, m := range matches {
			if data, err := os.ReadFile(m); err == nil {
				val := strings.TrimSpace(string(data))
				if strings.Contains(m, "conservation_mode") {
					if val == "1" {
						fmt.Println("Battery charge threshold: 60-80% (Lenovo conservation mode enabled)")
					} else {
						fmt.Println("Battery charge threshold: 100% (Lenovo conservation mode disabled)")
					}
					return exitOK
				}
				fmt.Printf("Battery charge threshold: %s%%\n", val)
				return exitOK
			}
		}
	}

	fmt.Println("Battery charge threshold: not exposed or standard profile active")
	return exitOK
}

func setBatteryThreshold(arg string) int {
	pct, err := strconv.Atoi(strings.TrimSuffix(strings.TrimSpace(arg), "%"))
	if err != nil || pct < 20 || pct > 100 {
		fmt.Fprintf(os.Stderr, "toren power threshold: invalid percentage %q (must be 20-100)\n", arg)
		return exitUsage
	}

	applied := false
	valStr := strconv.Itoa(pct)

	nodes := []string{
		"/sys/class/power_supply/BAT0/charge_control_end_threshold",
		"/sys/class/power_supply/BAT1/charge_control_end_threshold",
		"/sys/class/power_supply/BAT0/charge_stop_threshold",
		"/sys/class/power_supply/BAT1/charge_stop_threshold",
		"/sys/devices/platform/lg-laptop/battery_care_limit",
	}

	for _, n := range nodes {
		if _, err := os.Stat(n); err == nil {
			_ = powerPrivileged("sh", "-c", fmt.Sprintf("echo %s > %s", valStr, n))
			applied = true
		}
	}

	// Lenovo conservation mode
	ideapadNodes, _ := filepath.Glob("/sys/bus/platform/drivers/ideapad_acpi/*/conservation_mode")
	for _, n := range ideapadNodes {
		mode := "1"
		if pct > 90 {
			mode = "0"
		}
		_ = powerPrivileged("sh", "-c", fmt.Sprintf("echo %s > %s", mode, n))
		applied = true
	}

	if applied {
		fmt.Printf("Battery charge limit set to %d%%\n", pct)
	} else {
		fmt.Printf("Battery charge limit requested (%d%%), no compatible sysfs node found on this machine.\n", pct)
	}
	return exitOK
}

func getSleepMode() int {
	data, err := os.ReadFile("/sys/power/mem_sleep")
	if err != nil {
		fmt.Println("Sleep modes: not available or ACPI sleep not supported")
		return exitOK
	}
	fmt.Printf("Suspend sleep modes: %s\n", strings.TrimSpace(string(data)))
	return exitOK
}

func setSleepMode(mode string) int {
	mode = strings.ToLower(strings.TrimSpace(mode))
	if mode != "s2idle" && mode != "deep" {
		fmt.Fprintf(os.Stderr, "toren power sleep: invalid mode %q (must be 's2idle' or 'deep')\n", mode)
		return exitUsage
	}

	_ = powerPrivileged("sh", "-c", fmt.Sprintf("echo %s > /sys/power/mem_sleep", mode))
	fmt.Printf("Sleep mode set to %s\n", mode)
	return exitOK
}

func applyViaSysd(gov, epp, profile string) bool {
	sockPath := "/run/toren/sysd.sock"
	conn, err := net.DialTimeout("unix", sockPath, 200*time.Millisecond)
	if err != nil {
		return false
	}
	defer conn.Close()

	req := fmt.Sprintf(`{"action":"set_power","profile":"%s","governor":"%s","epp":"%s"}`+"\n", profile, gov, epp)
	if _, err := conn.Write([]byte(req)); err != nil {
		return false
	}
	return true
}
