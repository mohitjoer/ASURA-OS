package main

import (
	"fmt"
	"os"
)

const menuUsage = `Usage: toren menu

Opens the ASURA "do everything" menu in the desktop shell over IPC.
`

func init() {
	register("menu", "open the ASURA \"do everything\" menu", torenMenu)
}

func torenMenu(args []string) int {
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help" || args[0] == "help") {
		fmt.Fprint(os.Stdout, menuUsage)
		return exitOK
	}
	return torenIpc([]string{"call", "toren", "togglePopover", "menu"})
}
