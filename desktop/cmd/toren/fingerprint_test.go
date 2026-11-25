package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInjectAndRemovePamBlock(t *testing.T) {
	cases := []struct {
		name     string
		original string
	}{
		{
			name: "sudo pam with common-auth",
			original: `#%PAM-1.0
session    required   pam_limits.so
session    required   pam_env.so readenv=1 user_readenv=0

@include common-auth
@include common-account
@include common-session-noninteractive
`,
		},
		{
			name: "hyprlock pam with auth include login",
			original: `#%PAM-1.0
auth include login
account include login
`,
		},
		{
			name: "polkit pam with system-auth",
			original: `#%PAM-1.0
auth       include      system-auth
account    include      system-auth
password   include      system-auth
session    include      system-auth
`,
		},
		{
			name: "direct pam_unix auth line",
			original: `#%PAM-1.0
auth required pam_unix.so nullok
account required pam_unix.so
`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			injected := injectPamBlock(tc.original)
			if !strings.Contains(injected, fprintBlockBegin) {
				t.Fatalf("injected content missing block begin")
			}
			if !strings.Contains(injected, fprintBlockLine) {
				t.Fatalf("injected content missing pam_fprintd line")
			}
			if !strings.Contains(injected, fprintBlockEnd) {
				t.Fatalf("injected content missing block end")
			}

			// Ensure injected block appears BEFORE the auth line
			blockIdx := strings.Index(injected, fprintBlockBegin)
			if strings.Contains(tc.original, "@include common-auth") {
				authIdx := strings.Index(injected, "@include common-auth")
				if blockIdx >= authIdx {
					t.Errorf("fingerprint block must appear ABOVE password line (@include common-auth)")
				}
			}
			if strings.Contains(tc.original, "auth include login") {
				authIdx := strings.Index(injected, "auth include login")
				if blockIdx >= authIdx {
					t.Errorf("fingerprint block must appear ABOVE password line (auth include login)")
				}
			}

			// Idempotency: injecting again must not duplicate the block
			reInjected := injectPamBlock(injected)
			if reInjected != injected {
				t.Errorf("injectPamBlock is not idempotent:\nFirst:\n%s\nSecond:\n%s", injected, reInjected)
			}

			// Removal must restore byte-for-byte exact original content
			restored := removePamBlock(injected)
			if restored != tc.original {
				t.Errorf("removePamBlock did not restore original byte-for-byte:\nGot:\n%s\nWant:\n%s", restored, tc.original)
			}
		})
	}
}

func TestPamByteForByteRestoreWithBackup(t *testing.T) {
	tmpDir := t.TempDir()
	testPam := filepath.Join(tmpDir, "sudo")
	origBytes := []byte(`#%PAM-1.0
session    required   pam_limits.so

@include common-auth
@include common-account
`)

	if err := os.WriteFile(testPam, origBytes, 0644); err != nil {
		t.Fatalf("writing test pam file: %v", err)
	}

	// Mock privileged runner for temp directory operations
	oldPriv := fprintPrivileged
	fprintPrivileged = func(args ...string) error {
		if len(args) == 0 {
			return nil
		}
		name := args[0]
		subArgs := args[1:]
		if name == "cp" && len(subArgs) == 2 {
			data, err := os.ReadFile(subArgs[0])
			if err != nil {
				return err
			}
			return os.WriteFile(subArgs[1], data, 0644)
		}
		if name == "rm" {
			for _, arg := range subArgs {
				if arg != "-f" {
					_ = os.Remove(arg)
				}
			}
			return nil
		}
		return nil
	}
	defer func() { fprintPrivileged = oldPriv }()

	// Apply PAM changes
	if err := applyPamToTarget(testPam); err != nil {
		t.Fatalf("applyPamToTarget: %v", err)
	}

	// Check backup file was created
	bakFile := testPam + ".toren-fprint-bak"
	bakData, err := os.ReadFile(bakFile)
	if err != nil {
		t.Fatalf("backup file not created: %v", err)
	}
	if !bytes.Equal(bakData, origBytes) {
		t.Fatalf("backup file does not match original bytes")
	}

	// Revert PAM changes
	if err := revertPamFromTarget(testPam); err != nil {
		t.Fatalf("revertPamFromTarget: %v", err)
	}

	// Verify restored file is byte-for-byte identical to original
	restoredData, err := os.ReadFile(testPam)
	if err != nil {
		t.Fatalf("reading restored file: %v", err)
	}
	if !bytes.Equal(restoredData, origBytes) {
		t.Errorf("restored file mismatch:\nGot:\n%s\nWant:\n%s", string(restoredData), string(origBytes))
	}

	// Verify backup file was removed after successful restoration
	if _, err := os.Stat(bakFile); !os.IsNotExist(err) {
		t.Errorf("backup file %s was not cleaned up after restore", bakFile)
	}
}

func TestTargetPamFilesExcludesGreeters(t *testing.T) {
	targets := targetPamFiles()
	for _, target := range targets {
		lower := strings.ToLower(target)
		if strings.Contains(lower, "greetd") || strings.Contains(lower, "sddm") || strings.Contains(lower, "lightdm") || strings.Contains(lower, "login") {
			t.Errorf("targetPamFiles must NEVER include login greeter, but contains: %s", target)
		}
	}
}
