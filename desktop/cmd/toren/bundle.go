package main

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
)

// `asura bundle` installs a curated set of apps in one go: a gaming setup, a
// development environment, editors, everyday apps. Each bundle is Void
// packages (XBPS) plus Flathub apps for what Void does not package, and may
// switch on services (asura enable). Installing takes a system snapshot first.
//
// Package names are checked against the repositories before anything is
// installed: one unknown name would make XBPS refuse the whole transaction,
// so a name a repository does not carry is reported and skipped instead.

type bundle struct {
	About    string
	Packages []string            // XBPS packages
	GPU      map[string][]string // extra XBPS packages per GPU vendor (amd, intel, nvidia)
	Flatpaks []string            // Flathub app ids
	Services []string            // asura enable names
	Note     string              // shown after installing
}

var bundles = map[string]bundle{
	"gaming": {
		About:    "Steam, Lutris, Heroic, RetroArch, MangoHud, GameMode, Gamescope, Wine and 32-bit graphics drivers",
		Packages: []string{"steam", "lutris", "retroarch", "mangohud", "gamemode", "gamescope", "wine", "winetricks", "vulkan-loader-32bit", "libgcc-32bit", "libstdc++-32bit", "libdrm-32bit", "libglvnd-32bit"},
		GPU: map[string][]string{
			"amd":    {"mesa-dri-32bit", "mesa-vulkan-radeon-32bit"},
			"intel":  {"mesa-dri-32bit", "mesa-vulkan-intel-32bit"},
			"nvidia": {"nvidia-libs-32bit"},
		},
		Flatpaks: []string{"com.heroicgameslauncher.hgl"},
		Note:     "Start games with gamemoderun %command% in Steam's launch options for GameMode, and mangohud for the overlay.",
	},
	"dev": {
		About:    "compilers and toolchains (C/C++, Go, Rust, Node.js, Python), Docker with Compose, and common CLI tools",
		Packages: []string{"base-devel", "git", "go", "rust", "cargo", "nodejs", "python3", "python3-pip", "python3-virtualenv", "docker", "docker-compose", "jq", "ripgrep", "fd", "github-cli", "lazygit", "direnv"},
		Services: []string{"docker"},
		Note:     "Log out and back in once so the docker group applies.",
	},
	"editors": {
		About:    "Helix, Emacs and Visual Studio Code (Neovim is already installed)",
		Packages: []string{"helix", "emacs-pgtk"},
		Flatpaks: []string{"com.visualstudio.code"},
	},
	"apps": {
		About:    "Signal, Discord, Spotify, Obsidian, Dropbox, Tailscale and the Bitwarden password manager",
		Packages: []string{"tailscale"},
		Flatpaks: []string{"org.signal.Signal", "com.discordapp.Discord", "com.spotify.Client", "md.obsidian.Obsidian", "com.dropbox.Client", "com.bitwarden.desktop"},
		Services: []string{"tailscale"},
		Note:     "Run sudo tailscale up to log this machine in to Tailscale.",
	},
	"creative": {
		About:    "GIMP, Inkscape, Krita, Blender, OBS Studio and Kdenlive",
		Packages: []string{"gimp", "inkscape", "krita", "blender", "obs", "kdenlive"},
	},
	"office": {
		About:    "LibreOffice, a PDF editor (Xournal++) and a scanner app (Simple Scan)",
		Packages: []string{"libreoffice", "xournalpp", "simple-scan"},
		Services: []string{"printing"},
	},
}

// Seams for tests: what a repository carries, the GPU vendors present, and
// running a command (as root through sudo, or as the user).
var (
	bundleRepoHas = func(pkg string) bool {
		return exec.Command("xbps-query", "-R", pkg).Run() == nil
	}
	bundleInstalled = func(pkg string) bool {
		return exec.Command("xbps-query", pkg).Run() == nil
	}
	bundleGPUs = detectGPUVendors
	bundleRoot = runPrivileged
	bundleUser = func(args ...string) error {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stdout, cmd.Stderr, cmd.Stdin = os.Stdout, os.Stderr, os.Stdin
		return cmd.Run()
	}
	bundleHasFlatpak = func() bool { _, err := exec.LookPath("flatpak"); return err == nil }
	bundleHasXbps    = func() bool { _, err := exec.LookPath("xbps-install"); return err == nil }
	bundleEnable     = func(name string) int { return asuraEnable([]string{name}) }
)

const bundleUsage = `Usage: asura bundle [list | show NAME | install NAME...]

Installs a curated set of apps in one go.

  list            every bundle (the default)
  show NAME       what a bundle installs
  install NAME    install one or more bundles; a system snapshot is taken
                  first, so sudo asura rollback --system can undo it
`

func init() {
	register("bundle", "install a curated set of apps: gaming, dev, editors, apps, creative, office", asuraBundle)
}

func bundleNames() []string {
	var names []string
	for n := range bundles {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

func asuraBundle(args []string) int {
	if len(args) == 0 || args[0] == "list" {
		fmt.Print(bundleUsage + "\n")
		for _, n := range bundleNames() {
			fmt.Printf("  %-9s %s\n", n, bundles[n].About)
		}
		return exitOK
	}
	switch args[0] {
	case "-h", "--help", "help":
		fmt.Print(bundleUsage)
		return exitOK
	case "show":
		if len(args) != 2 {
			fmt.Fprint(os.Stderr, bundleUsage)
			return exitUsage
		}
		b, ok := bundles[args[1]]
		if !ok {
			return unknownBundle(args[1])
		}
		fmt.Printf("%s: %s\n", args[1], b.About)
		fmt.Printf("  packages: %s\n", strings.Join(b.Packages, " "))
		for _, v := range []string{"amd", "intel", "nvidia"} {
			if extra := b.GPU[v]; len(extra) > 0 {
				fmt.Printf("  with %s graphics: %s\n", v, strings.Join(extra, " "))
			}
		}
		if len(b.Flatpaks) > 0 {
			fmt.Printf("  from Flathub: %s\n", strings.Join(b.Flatpaks, " "))
		}
		if len(b.Services) > 0 {
			fmt.Printf("  switches on: %s\n", strings.Join(b.Services, " "))
		}
		return exitOK
	case "install":
		if len(args) < 2 {
			fmt.Fprint(os.Stderr, bundleUsage)
			return exitUsage
		}
		for _, n := range args[1:] {
			if _, ok := bundles[n]; !ok {
				return unknownBundle(n)
			}
		}
		return installBundles(args[1:])
	}
	return unknownBundle(args[0])
}

func unknownBundle(name string) int {
	fmt.Fprintf(os.Stderr, "asura bundle: no bundle %q; one of: %s\n", name, strings.Join(bundleNames(), ", "))
	return exitUsage
}

// bundlePlan is what installing the named bundles would do on this machine.
type bundlePlan struct {
	Packages, Missing, Flatpaks, Services, Notes []string
}

func planBundles(names []string) bundlePlan {
	var plan bundlePlan
	seen := map[string]bool{}
	add := func(list *[]string, v string) {
		if !seen[v] {
			seen[v] = true
			*list = append(*list, v)
		}
	}
	gpus := bundleGPUs()
	for _, n := range names {
		b := bundles[n]
		pkgs := append([]string{}, b.Packages...)
		for _, v := range gpus {
			// NVIDIA's 32-bit libraries match its own driver; with the open
			// driver (nouveau) the Mesa ones are what games need.
			if v == "nvidia" && !bundleInstalled("nvidia") {
				pkgs = append(pkgs, b.GPU["amd"][:1]...)
				continue
			}
			pkgs = append(pkgs, b.GPU[v]...)
		}
		for _, p := range pkgs {
			switch {
			case seen[p]:
			case bundleInstalled(p):
				seen[p] = true
			case bundleRepoHas(p):
				add(&plan.Packages, p)
			default:
				add(&plan.Missing, p)
			}
		}
		for _, f := range b.Flatpaks {
			add(&plan.Flatpaks, f)
		}
		for _, s := range b.Services {
			add(&plan.Services, "service:"+s)
		}
		if b.Note != "" {
			plan.Notes = append(plan.Notes, b.Note)
		}
	}
	for i, s := range plan.Services {
		plan.Services[i] = strings.TrimPrefix(s, "service:")
	}
	return plan
}

func installBundles(names []string) int {
	if !bundleHasXbps() {
		fmt.Fprintln(os.Stderr, "asura bundle: xbps-install is not installed")
		return exitFail
	}
	fmt.Println("==> Refreshing the package lists")
	if err := bundleRoot("xbps-install", "-S"); err != nil {
		fmt.Fprintln(os.Stderr, "asura bundle: refreshing the package lists:", err)
		return exitFail
	}
	plan := planBundles(names)
	for _, m := range plan.Missing {
		fmt.Printf("skipped: %s is not in the configured repositories\n", m)
	}
	if len(plan.Packages) > 0 {
		_ = bundleRoot(selfPath(), "snapshot", "create", "--reason", "install")
		fmt.Printf("==> Installing %d package(s): %s\n", len(plan.Packages), strings.Join(plan.Packages, " "))
		if err := bundleRoot(append([]string{"xbps-install", "-y"}, plan.Packages...)...); err != nil {
			fmt.Fprintln(os.Stderr, "asura bundle: installing packages:", err)
			return exitFail
		}
	} else {
		fmt.Println("==> Every package is installed already")
	}
	if len(plan.Flatpaks) > 0 {
		if !bundleHasFlatpak() {
			fmt.Printf("skipped: Flatpak is not installed, so %s were not installed\n", strings.Join(plan.Flatpaks, " "))
		} else {
			fmt.Printf("==> Installing from Flathub: %s\n", strings.Join(plan.Flatpaks, " "))
			_ = bundleUser("flatpak", "remote-add", "--user", "--if-not-exists", "flathub", "https://dl.flathub.org/repo/flathub.flatpakrepo")
			if err := bundleUser(append([]string{"flatpak", "install", "--user", "-y", "--noninteractive", "flathub"}, plan.Flatpaks...)...); err != nil {
				fmt.Fprintln(os.Stderr, "asura bundle: installing from Flathub:", err)
				return exitFail
			}
		}
	}
	for _, s := range plan.Services {
		if code := bundleEnable(s); code != exitOK {
			return code
		}
	}
	for _, n := range plan.Notes {
		fmt.Println("note:", n)
	}
	fmt.Printf("bundle installed: %s\n", strings.Join(names, ", "))
	return exitOK
}

// detectGPUVendors reads the display controllers lspci lists.
func detectGPUVendors() []string {
	out, err := exec.Command("lspci").Output()
	if err != nil {
		return nil
	}
	return gpuVendorsIn(string(out))
}

func gpuVendorsIn(lspci string) []string {
	found := map[string]bool{}
	for _, line := range strings.Split(lspci, "\n") {
		l := strings.ToLower(line)
		if !strings.Contains(l, "vga") && !strings.Contains(l, "3d controller") && !strings.Contains(l, "display controller") {
			continue
		}
		switch {
		case strings.Contains(l, "nvidia"):
			found["nvidia"] = true
		case strings.Contains(l, "amd") || strings.Contains(l, "ati "):
			found["amd"] = true
		case strings.Contains(l, "intel"):
			found["intel"] = true
		}
	}
	var out []string
	for _, v := range []string{"amd", "intel", "nvidia"} {
		if found[v] {
			out = append(out, v)
		}
	}
	return out
}
