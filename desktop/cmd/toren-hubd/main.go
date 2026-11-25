// toren-hubd is the Go backend for the Toren Settings GUI. The Quickshell front end
// (qs -c hub) shells out to it the same way the rest of the desktop talks to
// toren-shell: a subcommand prints data on stdout or mutates persisted state.
//
//	toren-hubd keybinds            print the keybind legend as JSON
//	toren-hubd config get <key>    print a stored config value
//	toren-hubd config set <k> <v>  persist a config value (TOML)
package main

import (
	"fmt"
	"os"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		usage()
		os.Exit(2)
	}

	switch args[0] {
	case "-h", "--help", "help":
		usage()
		return
	case "keybinds":
		if err := runKeybinds(args[1:]); err != nil {
			fmt.Fprintln(os.Stderr, "toren-hubd:", err)
			os.Exit(1)
		}
	case "apps":
		if err := printApps(); err != nil {
			fmt.Fprintln(os.Stderr, "toren-hubd:", err)
			os.Exit(1)
		}
		fmt.Println()
	case "config":
		if err := runConfig(args[1:]); err != nil {
			fmt.Fprintln(os.Stderr, "toren-hubd:", err)
			os.Exit(1)
		}
	case "shell":
		if err := runShellPref(args[1:]); err != nil {
			fmt.Fprintln(os.Stderr, "toren-hubd:", err)
			os.Exit(1)
		}
	case "desktop":
		if err := runDesktop(args[1:]); err != nil {
			fmt.Fprintln(os.Stderr, "toren-hubd:", err)
			os.Exit(1)
		}
	case "wm":
		if err := runWm(args[1:]); err != nil {
			fmt.Fprintln(os.Stderr, "toren-hubd:", err)
			os.Exit(1)
		}
	case "outputs":
		if err := runOutputs(args[1:]); err != nil {
			fmt.Fprintln(os.Stderr, "toren-hubd:", err)
			os.Exit(1)
		}
	case "lock":
		if err := runLock(args[1:]); err != nil {
			fmt.Fprintln(os.Stderr, "toren-hubd:", err)
			os.Exit(1)
		}
	case "gpu":
		if err := runGpu(args[1:]); err != nil {
			fmt.Fprintln(os.Stderr, "toren-hubd:", err)
			os.Exit(1)
		}
	case "cpu":
		if err := runCpu(args[1:]); err != nil {
			fmt.Fprintln(os.Stderr, "toren-hubd:", err)
			os.Exit(1)
		}
	case "lighting":
		if err := runLighting(args[1:]); err != nil {
			fmt.Fprintln(os.Stderr, "toren-hubd:", err)
			os.Exit(1)
		}
	case "voxtype":
		if err := unavailable("voxtype", "dictation needs the voxtype engine, which Toren does not ship", args[1:]); err != nil {
			fmt.Fprintln(os.Stderr, "toren-hubd:", err)
			os.Exit(1)
		}
	case "rice":
		if err := runRice(args[1:]); err != nil {
			fmt.Fprintln(os.Stderr, "toren-hubd:", err)
			os.Exit(1)
		}
	case "fastfetch":
		if err := runFastfetch(args[1:]); err != nil {
			fmt.Fprintln(os.Stderr, "toren-hubd:", err)
			os.Exit(1)
		}
	case "profile":
		if err := runProfile(args[1:]); err != nil {
			fmt.Fprintln(os.Stderr, "toren-hubd:", err)
			os.Exit(1)
		}
	case "reload-cover":
		if err := runReloadCover(args[1:]); err != nil {
			fmt.Fprintln(os.Stderr, "toren-hubd:", err)
			os.Exit(1)
		}
	case "import":
		if err := runImport(args[1:]); err != nil {
			fmt.Fprintln(os.Stderr, "toren-hubd:", err)
			os.Exit(1)
		}
	case "palette-bridge":
		if err := runPaletteBridge(args[1:]); err != nil {
			fmt.Fprintln(os.Stderr, "toren-hubd:", err)
			os.Exit(1)
		}
	case "share":
		if err := runShare(args[1:]); err != nil {
			fmt.Fprintln(os.Stderr, "toren-hubd:", err)
			os.Exit(1)
		}
	case "clipboard":
		if err := runClipboard(args[1:]); err != nil {
			fmt.Fprintln(os.Stderr, "toren-hubd:", err)
			os.Exit(1)
		}
	case "rules":
		if err := runRules(args[1:]); err != nil {
			fmt.Fprintln(os.Stderr, "toren-hubd:", err)
			os.Exit(1)
		}
	case "autostart":
		if err := runAutostart(args[1:]); err != nil {
			fmt.Fprintln(os.Stderr, "toren-hubd:", err)
			os.Exit(1)
		}
	case "env":
		if err := runEnv(args[1:]); err != nil {
			fmt.Fprintln(os.Stderr, "toren-hubd:", err)
			os.Exit(1)
		}
	case "look":
		if err := runLook(args[1:]); err != nil {
			fmt.Fprintln(os.Stderr, "toren-hubd:", err)
			os.Exit(1)
		}
	case "updates":
		if err := runUpdates(args[1:]); err != nil {
			fmt.Fprintln(os.Stderr, "toren-hubd:", err)
			os.Exit(1)
		}
	default:
		usage()
		os.Exit(2)
	}
}

func runConfig(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("config needs get|set")
	}
	switch args[0] {
	case "get":
		if len(args) < 2 {
			return fmt.Errorf("config get needs a key")
		}
		v, ok := configGet(args[1])
		if !ok {
			return fmt.Errorf("unknown config key: %s", args[1])
		}
		fmt.Println(v)
		return nil
	case "set":
		if len(args) < 3 {
			return fmt.Errorf("config set needs a key and value")
		}
		return configSet(args[1], args[2])
	default:
		return fmt.Errorf("config needs get|set")
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage:")
	fmt.Fprintln(os.Stderr, "  toren-hubd keybinds")
	fmt.Fprintln(os.Stderr, "  toren-hubd config get <key>")
	fmt.Fprintln(os.Stderr, "  toren-hubd config set <key> <value>")
	fmt.Fprintln(os.Stderr, "  toren-hubd shell get|set <fish|bash|zsh>")
	fmt.Fprintln(os.Stderr, "  toren-hubd desktop get|defaults|cursors|layouts")
	fmt.Fprintln(os.Stderr, "  toren-hubd desktop variants <layout>")
	fmt.Fprintln(os.Stderr, "  toren-hubd desktop save|preview <json>")
	fmt.Fprintln(os.Stderr, "  toren-hubd desktop restore")
	fmt.Fprintln(os.Stderr, "  toren-hubd desktop set-rebind <default> <chosen>")
	fmt.Fprintln(os.Stderr, "  toren-hubd desktop plugins list")
	fmt.Fprintln(os.Stderr, "  toren-hubd desktop plugins rebuild [--all|--stale] [--checkout <dir>] [<id>...]")
	fmt.Fprintln(os.Stderr, "  toren-hubd desktop plugins add [--inspect] <git-url> [<plugin>...]")
	fmt.Fprintln(os.Stderr, "  toren-hubd desktop plugins remove <id>")
	fmt.Fprintln(os.Stderr, "  toren-hubd wm list")
	fmt.Fprintln(os.Stderr, "  toren-hubd wm preview <name>")
	fmt.Fprintln(os.Stderr, "  toren-hubd outputs [list]")
	fmt.Fprintln(os.Stderr, "  toren-hubd outputs apply <json>")
	fmt.Fprintln(os.Stderr, "  toren-hubd outputs profiles")
	fmt.Fprintln(os.Stderr, "  toren-hubd outputs save <name> <json>")
	fmt.Fprintln(os.Stderr, "  toren-hubd outputs load|rm <name>")
	fmt.Fprintln(os.Stderr, "  toren-hubd lock list|set <slug>")
	fmt.Fprintln(os.Stderr, "  toren-hubd gpu caps|mode get|set <mode>")
	fmt.Fprintln(os.Stderr, "  toren-hubd gpu tune caps|get")
	fmt.Fprintln(os.Stderr, "  toren-hubd gpu tune set <gpu> <id> <value>")
	fmt.Fprintln(os.Stderr, "  toren-hubd gpu tune reset [<gpu>]")
	fmt.Fprintln(os.Stderr, "  toren-hubd cpu caps [<profile>]")
	fmt.Fprintln(os.Stderr, "  toren-hubd cpu active")
	fmt.Fprintln(os.Stderr, "  toren-hubd cpu set <scope> <id> <value>")
	fmt.Fprintln(os.Stderr, "  toren-hubd lighting state|scan|enable|disable|apply")
	fmt.Fprintln(os.Stderr, "  toren-hubd lighting set <device> <json>")
	fmt.Fprintln(os.Stderr, "  toren-hubd lighting accent [#RRGGBB]")
	fmt.Fprintln(os.Stderr, "  toren-hubd lighting save|release <device>")
	fmt.Fprintln(os.Stderr, "  toren-hubd voxtype ...   (answers \"unavailable\": no dictation engine on Toren)")
	fmt.Fprintln(os.Stderr, "  toren-hubd palette-bridge status|install|service|integration|doctor [<source>]")
	fmt.Fprintln(os.Stderr, "  toren-hubd share status")
	fmt.Fprintln(os.Stderr, "  toren-hubd rice list|preflight|capture|apply|restore|save|fork|delete|import|setwall|files|export")
	fmt.Fprintln(os.Stderr, "  toren-hubd fastfetch get|preview <json>")
	fmt.Fprintln(os.Stderr, "  toren-hubd fastfetch save <json>")
	fmt.Fprintln(os.Stderr, "  toren-hubd fastfetch import-logo <path>")
	fmt.Fprintln(os.Stderr, "  toren-hubd reload-cover import <path>")
	fmt.Fprintln(os.Stderr, "  toren-hubd reload-cover prune [<managed-path>]")
	fmt.Fprintln(os.Stderr, "  toren-hubd clipboard stats|prune")
	fmt.Fprintln(os.Stderr, "  toren-hubd rules [list|add|remove]")
	fmt.Fprintln(os.Stderr, "  toren-hubd autostart [list|add|remove]")
	fmt.Fprintln(os.Stderr, "  toren-hubd env [list|check|set|remove]")
	fmt.Fprintln(os.Stderr, "  toren-hubd look [list|set|mode|wallpapers]")
	fmt.Fprintln(os.Stderr, "  toren-hubd updates [status|check|channel|snapshots]")
	fmt.Fprintln(os.Stderr, "  toren-hubd import scan <path|url>")
	fmt.Fprintln(os.Stderr, "  toren-hubd import apply <decisions.json|->")
	fmt.Fprintln(os.Stderr, "  toren-hubd import undo [<ts>]")
}
