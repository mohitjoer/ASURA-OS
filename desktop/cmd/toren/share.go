package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const shareUsage = `Usage: toren share send FILE DEVICE
       toren share receive
       toren share devices

File sharing over Tailscale Taildrop.

Commands:
  send FILE DEVICE  send FILE to a Tailscale DEVICE
  receive           fetch incoming files into ~/Downloads
  devices           list available Tailscale devices

Requirements:
  Requires Tailscale to be enabled and running (toren enable tailscale).
`

func init() {
	register("share", "share files across devices with Tailscale", torenShare)
	helpText["share"] = shareUsage
}

var (
	shareTailscaleStatus = func() ([]byte, error) {
		return exec.Command("tailscale", "status", "--json").Output()
	}
	shareExec = func(args ...string) error {
		cmd := exec.Command("tailscale", args...)
		cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
		return cmd.Run()
	}
)

func isTailscaleAvailable() bool {
	_, err := exec.LookPath("tailscale")
	if err != nil {
		return false
	}
	if _, err := os.Stat("/var/service/tailscaled"); err != nil {
		return false
	}
	return true
}

func torenShare(args []string) int {
	if !isTailscaleAvailable() {
		fmt.Fprintln(os.Stderr, "toren share: Tailscale is not enabled or running.")
		fmt.Fprintln(os.Stderr, "Enable it with:  toren enable tailscale")
		fmt.Fprintln(os.Stderr, "And log in with: sudo tailscale up")
		return exitFail
	}

	if len(args) == 0 || args[0] == "devices" || args[0] == "list" {
		return shareDevices()
	}

	switch args[0] {
	case "send":
		if len(args) < 3 {
			fmt.Fprintln(os.Stderr, "toren share send: missing FILE or DEVICE argument")
			fmt.Fprint(os.Stderr, shareUsage)
			return exitUsage
		}
		return shareSend(args[1], args[2])
	case "receive":
		return shareReceive()
	default:
		fmt.Fprintf(os.Stderr, "toren share: unknown subcommand %q\n\n", args[0])
		fmt.Fprint(os.Stderr, shareUsage)
		return exitUsage
	}
}

func shareDevices() int {
	out, err := shareTailscaleStatus()
	if err != nil {
		fmt.Fprintf(os.Stderr, "toren share: tailscale status failed: %v\n", err)
		return exitFail
	}

	var status struct {
		Self struct {
			HostName string `json:"HostName"`
			DNSName  string `json:"DNSName"`
			Online   bool   `json:"Online"`
		} `json:"Self"`
		Peer map[string]struct {
			HostName string `json:"HostName"`
			DNSName  string `json:"DNSName"`
			Online   bool   `json:"Online"`
			OS       string `json:"OS"`
		} `json:"Peer"`
	}

	if err := json.Unmarshal(out, &status); err != nil {
		fmt.Fprintf(os.Stderr, "toren share: could not parse tailscale status: %v\n", err)
		return exitFail
	}

	fmt.Printf("Self: %s (%s)\n\n", status.Self.HostName, status.Self.DNSName)
	fmt.Printf("%-24s %-10s %-10s\n", "DEVICE", "STATUS", "OS")
	fmt.Println(strings.Repeat("-", 50))
	for _, p := range status.Peer {
		st := "offline"
		if p.Online {
			st = "online"
		}
		fmt.Printf("%-24s %-10s %-10s\n", p.HostName, st, p.OS)
	}
	return exitOK
}

func shareSend(file, device string) int {
	if _, err := os.Stat(file); err != nil {
		fmt.Fprintf(os.Stderr, "toren share send: file %q not found\n", file)
		return exitFail
	}

	target := device
	if !strings.HasSuffix(target, ":") {
		target += ":"
	}

	fmt.Printf("toren share: sending %s to %s...\n", filepath.Base(file), device)
	if err := shareExec("file", "cp", file, target); err != nil {
		fmt.Fprintf(os.Stderr, "toren share: send failed: %v\n", err)
		return exitFail
	}
	fmt.Println("✓ File sent successfully.")
	return exitOK
}

func shareReceive() int {
	downloads := filepath.Join(homeDir(), "Downloads")
	_ = os.MkdirAll(downloads, 0o755)

	fmt.Println("toren share: checking for incoming Taildrop files...")
	cmd := exec.Command("tailscale", "file", "get", downloads)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "toren share: receive failed: %v\n%s\n", err, string(out))
		return exitFail
	}

	output := strings.TrimSpace(string(out))
	if output != "" {
		fmt.Println(output)
	}
	fmt.Printf("✓ Files received into %s\n", downloads)

	_ = exec.Command("notify-send",
		"-a", "Toren Share",
		"-i", "preferences-desktop-remote-desktop",
		"File Received",
		fmt.Sprintf("Tailscale files received into %s", downloads),
	).Run()

	return exitOK
}
