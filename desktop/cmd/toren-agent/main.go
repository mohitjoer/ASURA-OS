package main

import (
	"fmt"
	"os"
)

const usage = `asura-agent: Toren Linux local coding agent harness integration

Usage:
  asura-agent wire [agent]      Symlink the Asura skill suite into all or specified agent
  asura-agent unwire [agent]    Remove the Asura skill symlink from all or specified agent
  asura-agent status [--json]   Report detection in PATH and wiring status of local agents
  asura-agent wire --installed  Wire every supported agent that is installed (run at login)
  asura-agent ask QUESTION      Ask ASURA's assistant (opens OpenCode in a terminal)
  asura-agent chat MESSAGE      The same, for the shell's chat panel (JSON lines)

Supported agents:
  opencode    OpenCode CLI (~/.config/opencode/skills/asura)
  pi          Pi coding agent (~/.pi/agent/skills/asura)
  claude      Claude Code (~/.claude/skills/asura)
  gemini      Gemini CLI / Antigravity (~/.gemini/skills/asura)
`

func main() {
	if len(os.Args) < 2 {
		fmt.Print(usage)
		os.Exit(2)
	}

	switch os.Args[1] {
	case "wire":
		agent := ""
		if len(os.Args) > 2 {
			agent = os.Args[2]
		}
		if agent == "--installed" {
			if _, err := WireInstalled(); err != nil {
				fmt.Fprintf(os.Stderr, "asura-agent wire: %v\n", err)
				os.Exit(1)
			}
			return
		}
		if _, err := Wire(agent); err != nil {
			fmt.Fprintf(os.Stderr, "asura-agent wire: %v\n", err)
			os.Exit(1)
		}
	case "unwire":
		agent := ""
		if len(os.Args) > 2 {
			agent = os.Args[2]
		}
		if _, err := Unwire(agent); err != nil {
			fmt.Fprintf(os.Stderr, "asura-agent unwire: %v\n", err)
			os.Exit(1)
		}
	case "status":
		asJSON := false
		for _, arg := range os.Args[2:] {
			if arg == "--json" {
				asJSON = true
			}
		}
		if err := Status(asJSON); err != nil {
			fmt.Fprintf(os.Stderr, "asura-agent status: %v\n", err)
			os.Exit(1)
		}
	case "ask":
		os.Exit(Ask(os.Args[2:], os.Stdout))
	case "chat":
		os.Exit(Chat(os.Args[2:], os.Stdout))
	case "agent":
		os.Exit(0) // the chat panel's engine picker: OpenCode is the only one
	case "-h", "--help", "help":
		fmt.Print(usage)
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", os.Args[1])
		fmt.Print(usage)
		os.Exit(2)
	}
}
