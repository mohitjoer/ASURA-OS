package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"toren/desktop/internal/runner"
)

const holdUsage = `Usage: toren hold [package...]

Hold packages back from updates, or list packages currently held.

When called with package names, xbps-pkgdb places each named package on hold so
xbps-install will not upgrade it during system updates.
When called with no arguments, lists all packages currently on hold.
`

const unholdUsage = `Usage: toren unhold <package...>

Release held packages so system updates can upgrade them.
`

func init() {
	register("hold", "hold packages back from updates, or list held packages", torenHold)
	register("unhold", "release held packages so updates can upgrade them", torenUnhold)
}

func torenHold(args []string) int {
	return holdMain(args, os.Stdout, os.Stderr, runner.Exec(30*time.Minute))
}

func torenUnhold(args []string) int {
	return unholdMain(args, os.Stdout, os.Stderr, runner.Exec(30*time.Minute))
}

func holdMain(args []string, out, errOut io.Writer, run runner.Func) int {
	set := newFlags("hold", holdUsage)
	packages, err := parseFlags(set, args)
	if err != nil {
		fmt.Fprint(errOut, holdUsage)
		return exitUsage
	}

	if len(packages) == 0 {
		argv := []string{"xbps-query", "-H"}
		outText, err := run(argv[0], argv[1:]...)
		if err != nil {
			fmt.Fprintf(errOut, "toren hold: %v\n", err)
			return exitFail
		}
		trimmed := strings.TrimSpace(outText)
		if trimmed == "" {
			fmt.Fprintln(out, "No packages are currently held.")
			return exitOK
		}
		fmt.Fprintln(out, trimmed)
		return exitOK
	}

	argv := append(pkgSudoPick(), "xbps-pkgdb", "-m", "hold")
	argv = append(argv, packages...)
	outText, err := run(argv[0], argv[1:]...)
	if err != nil {
		if strings.TrimSpace(outText) != "" {
			fmt.Fprintln(errOut, strings.TrimSpace(outText))
		}
		fmt.Fprintf(errOut, "toren hold: %v\n", err)
		return exitFail
	}
	if strings.TrimSpace(outText) != "" {
		fmt.Fprintln(out, strings.TrimSpace(outText))
	}
	fmt.Fprintf(out, "Held package(s): %s\n", strings.Join(packages, " "))
	return exitOK
}

func unholdMain(args []string, out, errOut io.Writer, run runner.Func) int {
	set := newFlags("unhold", unholdUsage)
	packages, err := parseFlags(set, args)
	if err != nil {
		fmt.Fprint(errOut, unholdUsage)
		return exitUsage
	}

	if len(packages) == 0 {
		fmt.Fprint(errOut, unholdUsage)
		return exitUsage
	}

	argv := append(pkgSudoPick(), "xbps-pkgdb", "-m", "unhold")
	argv = append(argv, packages...)
	outText, err := run(argv[0], argv[1:]...)
	if err != nil {
		if strings.TrimSpace(outText) != "" {
			fmt.Fprintln(errOut, strings.TrimSpace(outText))
		}
		fmt.Fprintf(errOut, "toren unhold: %v\n", err)
		return exitFail
	}
	if strings.TrimSpace(outText) != "" {
		fmt.Fprintln(out, strings.TrimSpace(outText))
	}
	fmt.Fprintf(out, "Unheld package(s): %s\n", strings.Join(packages, " "))
	return exitOK
}
