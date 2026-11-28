package main

import (
	"os"
	"toren/desktop/internal/prowl"
)

func main() {
	os.Exit(prowl.Run(os.Args[1:], os.Stdout, os.Stderr))
}
