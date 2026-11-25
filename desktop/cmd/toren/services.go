package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"sort"
	"strings"
)

// `toren enable NAME` and `toren disable NAME` switch optional system
// services on and off: the package is installed if it is missing, the runit
// service is linked into /var/service, and the user joins the group the
// service needs. They are off by default because most machines do not need
// them, and each one running is memory, boot time and attack surface.

type optionalService struct {
	About    string
	Package  string
	Services []string // runit services under /etc/sv
	Group    string   // a group the user joins, if any
}

var switchableServices = map[string]optionalService{
	"docker":      {About: "Docker containers", Package: "docker", Services: []string{"docker"}, Group: "docker"},
	"bluetooth":   {About: "Bluetooth devices", Package: "bluez", Services: []string{"bluetoothd"}, Group: "bluetooth"},
	"printing":    {About: "printers (CUPS)", Package: "cups", Services: []string{"cupsd"}, Group: "lpadmin"},
	"ssh":         {About: "logging in to this machine over SSH", Package: "openssh", Services: []string{"sshd"}},
	"libvirt":     {About: "virtual machines (libvirt)", Package: "libvirt", Services: []string{"libvirtd", "virtlockd", "virtlogd"}, Group: "libvirt"},
	"tailscale":   {About: "a Tailscale VPN", Package: "tailscale", Services: []string{"tailscaled"}},
	"fingerprint": {About: "fingerprint authentication (fprintd)", Package: "fprintd", Services: []string{"fprintd"}},
}

// Where runit keeps service definitions and the enabled links, and how a
// root command runs; tests point these elsewhere.
var (
	svDefinitions = "/etc/sv"
	svEnabled     = "/var/service"
	svPrivileged  = runPrivileged
	svInstalled   = func(pkg string) bool { return exec.Command("xbps-query", pkg).Run() == nil }
)

const enableUsage = `Usage: toren enable NAME | toren disable NAME | toren enable --list

Switches an optional system service on or off. Enabling installs its package
if it is missing, starts it (and at every boot), and adds you to its group;
disabling stops it and keeps the package.

`

func init() {
	register("enable", "switch an optional service on (docker, bluetooth, ssh, ...)", torenEnable)
	register("disable", "switch an optional service off", torenDisable)
}

func serviceNames() []string {
	var names []string
	for name := range switchableServices {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func printServiceTable() {
	fmt.Print(enableUsage)
	for _, name := range serviceNames() {
		s := switchableServices[name]
		fmt.Printf("  %-10s %-4s %s\n", name, serviceState(s), s.About)
	}
}

// serviceState is "on" when every runit service is linked, else "off".
func serviceState(s optionalService) string {
	for _, svc := range s.Services {
		if _, err := os.Lstat(filepath.Join(svEnabled, svc)); err != nil {
			return "off"
		}
	}
	return "on"
}

func lookupService(cmd string, args []string) (string, optionalService, int) {
	if len(args) == 0 || args[0] == "--list" || args[0] == "list" {
		printServiceTable()
		return "", optionalService{}, exitOK
	}
	if args[0] == "-h" || args[0] == "--help" {
		printServiceTable()
		return "", optionalService{}, exitOK
	}
	s, ok := switchableServices[args[0]]
	if !ok || len(args) > 1 {
		fmt.Fprintf(os.Stderr, "toren %s: unknown service %q; one of: %s\n", cmd, strings.Join(args, " "), strings.Join(serviceNames(), ", "))
		return "", optionalService{}, exitUsage
	}
	return args[0], s, -1
}

func torenEnable(args []string) int {
	name, s, code := lookupService("enable", args)
	if code >= 0 {
		return code
	}
	if name == "fingerprint" {
		return runFingerprintEnable()
	}
	if !svInstalled(s.Package) {
		fmt.Printf("installing %s...\n", s.Package)
		if err := svPrivileged("xbps-install", "-Sy", s.Package); err != nil {
			fmt.Fprintf(os.Stderr, "toren enable: installing %s: %v\n", s.Package, err)
			return exitFail
		}
	}
	for _, svc := range s.Services {
		def := filepath.Join(svDefinitions, svc)
		if _, err := os.Stat(def); err != nil {
			fmt.Fprintf(os.Stderr, "toren enable: %s has no runit service (%s)\n", s.Package, def)
			return exitFail
		}
		link := filepath.Join(svEnabled, svc)
		if _, err := os.Lstat(link); err == nil {
			continue
		}
		if err := svPrivileged("ln", "-s", def, link); err != nil {
			fmt.Fprintf(os.Stderr, "toren enable: %v\n", err)
			return exitFail
		}
	}
	if s.Group != "" {
		if who := invokingUser(); who != "" && !inGroup(who, s.Group) {
			if err := svPrivileged("usermod", "-aG", s.Group, who); err != nil {
				fmt.Fprintf(os.Stderr, "toren enable: adding %s to %s: %v\n", who, s.Group, err)
				return exitFail
			}
			fmt.Printf("added %s to the %s group; log out and back in for it to apply\n", who, s.Group)
		}
	}
	fmt.Printf("%s is on (%s); runit starts it within a few seconds and at every boot\n", name, s.About)
	return exitOK
}

func torenDisable(args []string) int {
	name, s, code := lookupService("disable", args)
	if code >= 0 {
		return code
	}
	if name == "fingerprint" {
		return runFingerprintDisable()
	}
	for _, svc := range s.Services {
		link := filepath.Join(svEnabled, svc)
		if _, err := os.Lstat(link); err != nil {
			continue
		}
		if err := svPrivileged("rm", "-f", link); err != nil {
			fmt.Fprintf(os.Stderr, "toren disable: %v\n", err)
			return exitFail
		}
	}
	fmt.Printf("%s is off; the %s package stays installed (sudo xbps-remove %s removes it)\n", name, s.Package, s.Package)
	return exitOK
}

// invokingUser is the person behind the command: SUDO_USER under sudo.
func invokingUser() string {
	if u := os.Getenv("SUDO_USER"); u != "" && u != "root" {
		return u
	}
	if u, err := user.Current(); err == nil && u.Username != "root" {
		return u.Username
	}
	return ""
}

func inGroup(name, group string) bool {
	u, err := user.Lookup(name)
	if err != nil {
		return false
	}
	g, err := user.LookupGroup(group)
	if err != nil {
		return false
	}
	ids, err := u.GroupIds()
	if err != nil {
		return false
	}
	for _, id := range ids {
		if id == g.Gid {
			return true
		}
	}
	return false
}
