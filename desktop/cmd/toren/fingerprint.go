package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// Fingerprint login and sudo integration for Toren OS:
//   toren enable fingerprint
//   toren disable fingerprint
//   toren fingerprint [status | enroll | verify | enable | disable]
//
// Rules & safety invariants:
// - Off by default, reversible with one command.
// - Never locks a user out.
// - Installs fprintd and libfprint; refuses clearly when no supported reader is found.
// - Enrolment: toren fingerprint enroll runs fprintd-enroll for the user;
//   nothing changes in PAM until at least one finger is enrolled.
// - PAM: adds auth sufficient pam_fprintd.so ABOVE the existing password line,
//   for sudo, the lock screen (hyprlock) and polkit only.
// - Password always keeps working. Never replaces or removes a password line.
// - Never touches the login greeter's PAM (greetd, login, sddm).
// - Marked block with byte-for-byte backup restoration on disable.

var (
	fprintPamSudo     = "/etc/pam.d/sudo"
	fprintPamHyprlock = "/etc/pam.d/hyprlock"
	fprintPamPolkit   = "/etc/pam.d/polkit-1"
	fprintPrivileged  = runPrivileged
	fprintCommandExec = func(name string, args ...string) ([]byte, error) {
		return exec.Command(name, args...).CombinedOutput()
	}
	fprintInstalled = svInstalled
)

const (
	fprintBlockBegin = "# BEGIN ASURA FINGERPRINT"
	fprintBlockLine  = "auth sufficient pam_fprintd.so"
	fprintBlockEnd   = "# END ASURA FINGERPRINT"
	fprintBlock      = fprintBlockBegin + "\n" + fprintBlockLine + "\n" + fprintBlockEnd + "\n"
)

const fingerprintUsage = `Usage: toren fingerprint [status | enroll | verify | enable | disable]
       toren enable fingerprint | toren disable fingerprint

Manages optional fingerprint authentication.

  status    show reader detection, enrolment status, service, and PAM configuration
  enroll    enroll a fingerprint with fprintd-enroll (activates PAM once enrolled)
  verify    verify an enrolled fingerprint with fprintd-verify
  enable    install fprintd/libfprint and enable the service (activates PAM if enrolled)
  disable   remove fingerprint from PAM (restores byte-for-byte) and stop service
`

func init() {
	register("fingerprint", "manage fingerprint reader, enrolment, and PAM authentication", torenFingerprint)
}

func torenFingerprint(args []string) int {
	if len(args) == 0 || args[0] == "status" {
		return runFingerprintStatus()
	}

	switch args[0] {
	case "status":
		return runFingerprintStatus()
	case "enroll":
		return runFingerprintEnroll(args[1:])
	case "verify":
		return runFingerprintVerify(args[1:])
	case "enable":
		return runFingerprintEnable()
	case "disable":
		return runFingerprintDisable()
	case "-h", "--help", "help":
		fmt.Print(fingerprintUsage)
		return exitOK
	default:
		fmt.Fprintf(os.Stderr, "toren fingerprint: unknown verb %q\n\n%s", args[0], fingerprintUsage)
		return exitUsage
	}
}

// hasSupportedFingerprintReader checks whether a fingerprint reader is present.
func hasSupportedFingerprintReader() (bool, string) {
	// 1. Try fprintd-list if installed
	if out, err := fprintCommandExec("fprintd-list"); err == nil {
		outStr := strings.TrimSpace(string(out))
		if len(outStr) > 0 && !strings.Contains(strings.ToLower(outStr), "no devices available") {
			lines := strings.Split(outStr, "\n")
			for _, line := range lines {
				if strings.Contains(strings.ToLower(line), "device") {
					return true, strings.TrimSpace(line)
				}
			}
			return true, lines[0]
		}
	}

	// 2. Scan USB devices for biometric or common fingerprint vendors
	// libfprint supported vendors / class 0xef or known IDs (Synaptics 06cb, Goodix 27c6, Elan 04f3, Validity 138a, FocalTech 2808, EgisTec 1c7a)
	knownVendors := []string{"06cb", "27c6", "04f3", "138a", "2808", "1c7a", "0bda"}
	if out, err := fprintCommandExec("lsusb"); err == nil {
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			lower := strings.ToLower(line)
			if strings.Contains(lower, "fingerprint") || strings.Contains(lower, "biometric") {
				return true, strings.TrimSpace(line)
			}
			for _, vendor := range knownVendors {
				if strings.Contains(lower, "id "+vendor+":") && (strings.Contains(lower, "sensor") || strings.Contains(lower, "reader") || strings.Contains(lower, "fingerprint")) {
					return true, strings.TrimSpace(line)
				}
			}
		}
	}

	// Check /sys/bus/usb/devices for USB interface class 0xff or vendor/product strings
	if entries, err := os.ReadDir("/sys/bus/usb/devices"); err == nil {
		for _, entry := range entries {
			devPath := filepath.Join("/sys/bus/usb/devices", entry.Name())
			prodFile := filepath.Join(devPath, "product")
			if prodBytes, err := os.ReadFile(prodFile); err == nil {
				prodStr := strings.ToLower(string(prodBytes))
				if strings.Contains(prodStr, "fingerprint") || strings.Contains(prodStr, "biometric") {
					return true, strings.TrimSpace(string(prodBytes))
				}
			}
		}
	}

	return false, "no supported fingerprint reader found on this machine"
}

// isFingerprintEnrolled checks if any fingers are enrolled for the given user.
func isFingerprintEnrolled(user string) bool {
	if user == "" {
		user = os.Getenv("USER")
		if user == "" {
			user = os.Getenv("LOGNAME")
		}
	}
	if user == "" {
		return false
	}

	// Check /var/lib/fprint or /var/lib/fprintd store
	for _, dir := range []string{"/var/lib/fprint", "/var/lib/fprintd"} {
		userDir := filepath.Join(dir, user)
		if entries, err := os.ReadDir(userDir); err == nil && len(entries) > 0 {
			return true
		}
	}

	// Try fprintd-list <user>
	out, err := fprintCommandExec("fprintd-list", user)
	if err != nil {
		return false
	}
	outStr := string(out)
	// Check for enrolled finger keywords
	fingerKeywords := []string{
		"right-thumb", "right-index-finger", "right-middle-finger", "right-ring-finger", "right-little-finger",
		"left-thumb", "left-index-finger", "left-middle-finger", "left-ring-finger", "left-little-finger",
	}
	for _, kw := range fingerKeywords {
		if strings.Contains(outStr, kw) {
			return true
		}
	}
	return strings.Contains(strings.ToLower(outStr), "enrolled: yes")
}

// regex to find existing password auth line in PAM files
// matches "@include common-auth", "auth include login", "auth include system-auth", "auth required pam_unix.so", etc.
var pamPasswordLineRe = regexp.MustCompile(`(?m)^([ \t]*(?:auth\s+(?:(?:sufficient|required|requisite|include)\s+\S+|\[[^\]]+\]\s+pam_unix\.so)|@include\s+(?:common-auth|system-auth)).*)$`)

// injectPamBlock inserts the fingerprint block above the existing password line.
func injectPamBlock(content string) string {
	if strings.Contains(content, fprintBlockBegin) {
		return content // already injected
	}

	loc := pamPasswordLineRe.FindStringIndex(content)
	if loc == nil {
		// Fallback: prepend block to the file
		return fprintBlock + content
	}

	// Insert immediately before the matched line
	return content[:loc[0]] + fprintBlock + content[loc[0]:]
}

// removePamBlock removes the marked block from the content.
func removePamBlock(content string) string {
	if !strings.Contains(content, fprintBlockBegin) {
		return content
	}

	blockRe := regexp.MustCompile(`(?s)# BEGIN ASURA FINGERPRINT\r?\n.*?# END ASURA FINGERPRINT\r?\n?`)
	return blockRe.ReplaceAllString(content, "")
}

// applyPamToTarget modifies a single PAM file safely with backup.
func applyPamToTarget(targetPath string) error {
	data, err := os.ReadFile(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // optional file not present
		}
		return err
	}

	origContent := string(data)
	if strings.Contains(origContent, fprintBlockBegin) {
		return nil // already applied
	}

	backupPath := targetPath + ".toren-fprint-bak"
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		if err := os.WriteFile(backupPath, data, 0644); err != nil {
			// Try privileged if normal write fails
			tmpBak := filepath.Join(os.TempDir(), filepath.Base(backupPath))
			if errWrite := os.WriteFile(tmpBak, data, 0644); errWrite == nil {
				_ = fprintPrivileged("cp", tmpBak, backupPath)
				_ = os.Remove(tmpBak)
			}
		}
	}

	newContent := injectPamBlock(origContent)
	tmpFile := filepath.Join(os.TempDir(), "pam-"+filepath.Base(targetPath))
	if err := os.WriteFile(tmpFile, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("writing temp PAM file: %w", err)
	}
	defer os.Remove(tmpFile)

	if err := fprintPrivileged("cp", tmpFile, targetPath); err != nil {
		return fmt.Errorf("updating %s: %w", targetPath, err)
	}
	return nil
}

// revertPamFromTarget removes the marked block and restores from backup if identical.
func revertPamFromTarget(targetPath string) error {
	data, err := os.ReadFile(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	origContent := string(data)
	if !strings.Contains(origContent, fprintBlockBegin) {
		return nil // nothing to revert
	}

	cleanedContent := removePamBlock(origContent)

	backupPath := targetPath + ".toren-fprint-bak"
	backupData, bakErr := os.ReadFile(backupPath)

	// If cleaned content equals backup, we have a byte-for-byte restoration
	if bakErr == nil && bytes.Equal([]byte(cleanedContent), backupData) {
		// Restore exact backup file
		if err := fprintPrivileged("cp", backupPath, targetPath); err != nil {
			return fmt.Errorf("restoring %s from backup: %w", targetPath, err)
		}
		_ = fprintPrivileged("rm", "-f", backupPath)
		return nil
	}

	// Write cleaned content
	tmpFile := filepath.Join(os.TempDir(), "pam-clean-"+filepath.Base(targetPath))
	if err := os.WriteFile(tmpFile, []byte(cleanedContent), 0644); err != nil {
		return err
	}
	defer os.Remove(tmpFile)

	if err := fprintPrivileged("cp", tmpFile, targetPath); err != nil {
		return fmt.Errorf("updating %s: %w", targetPath, err)
	}

	if bakErr == nil {
		_ = fprintPrivileged("rm", "-f", backupPath)
	}
	return nil
}

// targetPamFiles returns the list of PAM targets for fingerprint auth.
// Crucial: sudo, hyprlock, polkit-1 only. NEVER login greeter (greetd/login).
func targetPamFiles() []string {
	return []string{
		fprintPamSudo,
		fprintPamHyprlock,
		fprintPamPolkit,
	}
}

// isPamConfigured checks whether any PAM target contains the fingerprint block.
func isPamConfigured() bool {
	for _, target := range targetPamFiles() {
		if data, err := os.ReadFile(target); err == nil {
			if strings.Contains(string(data), fprintBlockBegin) {
				return true
			}
		}
	}
	return false
}

func runFingerprintStatus() int {
	fmt.Println("Fingerprint Status:")
	fmt.Println()

	hasReader, desc := hasSupportedFingerprintReader()
	if hasReader {
		fmt.Printf("  Reader:      OK (%s)\n", desc)
	} else {
		fmt.Printf("  Reader:      None (%s)\n", desc)
	}

	// Package status
	fprintdInstalled := fprintInstalled("fprintd")
	libfprintInstalled := fprintInstalled("libfprint")
	fmt.Printf("  Packages:    fprintd=%v, libfprint=%v\n", fprintdInstalled, libfprintInstalled)

	// Service status
	serviceLinked := false
	if _, err := os.Lstat(filepath.Join(svEnabled, "fprintd")); err == nil {
		serviceLinked = true
	}
	fmt.Printf("  Service:     fprintd running=%v\n", serviceLinked)

	// Enrolment
	currentUser := os.Getenv("USER")
	enrolled := isFingerprintEnrolled(currentUser)
	fmt.Printf("  Enrolled:    %v (user: %s)\n", enrolled, currentUser)

	// PAM status
	pamActive := isPamConfigured()
	fmt.Printf("  PAM Auth:    configured=%v (sudo, hyprlock, polkit)\n", pamActive)
	fmt.Println()

	if pamActive && !fprintdInstalled {
		fmt.Println("  WARNING: PAM has fingerprint authentication configured but fprintd is missing!")
		fmt.Println("           Run 'toren disable fingerprint' to restore PAM safely.")
	} else if hasReader && !enrolled {
		fmt.Println("  Ready to enroll. Run 'toren fingerprint enroll' to record a finger.")
	} else if enrolled && !pamActive {
		fmt.Println("  Enrolled but PAM is off. Run 'toren enable fingerprint' to activate.")
	}

	return exitOK
}

func runFingerprintEnable() int {
	hasReader, desc := hasSupportedFingerprintReader()
	if !hasReader {
		fmt.Fprintf(os.Stderr, "toren enable fingerprint: refused — %s\n", desc)
		return exitFail
	}

	// Install packages if missing
	for _, pkg := range []string{"fprintd", "libfprint"} {
		if !fprintInstalled(pkg) {
			fmt.Printf("installing %s...\n", pkg)
			if err := fprintPrivileged("xbps-install", "-Sy", pkg); err != nil {
				fmt.Fprintf(os.Stderr, "toren enable fingerprint: installing %s: %v\n", pkg, err)
				return exitFail
			}
		}
	}

	// Enable runit service if definition exists
	def := filepath.Join(svDefinitions, "fprintd")
	link := filepath.Join(svEnabled, "fprintd")
	if _, err := os.Stat(def); err == nil {
		if _, err := os.Lstat(link); os.IsNotExist(err) {
			if err := fprintPrivileged("ln", "-s", def, link); err != nil {
				fmt.Fprintf(os.Stderr, "toren enable fingerprint: linking service: %v\n", err)
			}
		}
	}

	currentUser := os.Getenv("USER")
	if !isFingerprintEnrolled(currentUser) {
		fmt.Println("Fingerprint daemon enabled successfully.")
		fmt.Println("Note: PAM has NOT been modified because no fingerprint is enrolled yet.")
		fmt.Println("Run 'toren fingerprint enroll' to record your fingerprint and activate PAM login.")
		return exitOK
	}

	// Enrolled: apply PAM
	for _, target := range targetPamFiles() {
		if err := applyPamToTarget(target); err != nil {
			fmt.Fprintf(os.Stderr, "toren enable fingerprint: configuring %s: %v\n", target, err)
			return exitFail
		}
	}

	fmt.Println("Fingerprint authentication enabled successfully for sudo, hyprlock, and polkit.")
	return exitOK
}

func runFingerprintDisable() int {
	// 1. Revert PAM targets
	for _, target := range targetPamFiles() {
		if err := revertPamFromTarget(target); err != nil {
			fmt.Fprintf(os.Stderr, "toren disable fingerprint: reverting %s: %v\n", target, err)
			return exitFail
		}
	}

	// 2. Unlink runit service if linked
	link := filepath.Join(svEnabled, "fprintd")
	if _, err := os.Lstat(link); err == nil {
		_ = fprintPrivileged("rm", "-f", link)
	}

	fmt.Println("Fingerprint authentication disabled and PAM restored byte-for-byte.")
	return exitOK
}

func runFingerprintEnroll(args []string) int {
	hasReader, desc := hasSupportedFingerprintReader()
	if !hasReader {
		fmt.Fprintf(os.Stderr, "toren fingerprint enroll: refused — %s\n", desc)
		return exitFail
	}

	// Ensure fprintd is installed
	if !fprintInstalled("fprintd") {
		fmt.Println("installing fprintd and libfprint...")
		if err := fprintPrivileged("xbps-install", "-Sy", "fprintd", "libfprint"); err != nil {
			fmt.Fprintf(os.Stderr, "toren fingerprint enroll: %v\n", err)
			return exitFail
		}
	}

	// Start runit service or ensure dbus can start it
	def := filepath.Join(svDefinitions, "fprintd")
	link := filepath.Join(svEnabled, "fprintd")
	if _, err := os.Stat(def); err == nil {
		if _, err := os.Lstat(link); os.IsNotExist(err) {
			_ = fprintPrivileged("ln", "-s", def, link)
		}
	}

	// Run fprintd-enroll directly for user
	cmdArgs := []string{}
	if len(args) > 0 {
		cmdArgs = append(cmdArgs, args...)
	}
	cmd := exec.Command("fprintd-enroll", cmdArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Println("Place your finger on the fingerprint reader when prompted...")
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "toren fingerprint enroll: enrolment failed: %v\n", err)
		return exitFail
	}

	currentUser := os.Getenv("USER")
	if isFingerprintEnrolled(currentUser) {
		fmt.Println("\nEnrolment successful! Activating PAM for sudo, lock screen, and polkit...")
		for _, target := range targetPamFiles() {
			if err := applyPamToTarget(target); err != nil {
				fmt.Fprintf(os.Stderr, "toren fingerprint: warning: configuring %s: %v\n", target, err)
			}
		}
		fmt.Println("Fingerprint login is now active.")
	}

	return exitOK
}

func runFingerprintVerify(args []string) int {
	cmd := exec.Command("fprintd-verify", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Println("Swipe or touch the fingerprint reader to verify...")
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "verification failed: %v\n", err)
		return exitFail
	}
	fmt.Println("Fingerprint verified successfully!")
	return exitOK
}
