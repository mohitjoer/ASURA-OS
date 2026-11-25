package main

import (
	"testing"
)

func TestUpdateGrubCmdline(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		resumeDev    string
		resumeOffset string
		expected     string
	}{
		{
			name:         "add to simple cmdline",
			input:        `GRUB_CMDLINE_LINUX_DEFAULT="loglevel=4 splash"`,
			resumeDev:    "UUID=abcd-1234",
			resumeOffset: "123456",
			expected:     `GRUB_CMDLINE_LINUX_DEFAULT="loglevel=4 splash resume=UUID=abcd-1234 resume_offset=123456"`,
		},
		{
			name:         "update existing resume and offset",
			input:        `GRUB_CMDLINE_LINUX_DEFAULT="loglevel=4 resume=UUID=old-uuid resume_offset=9999 splash"`,
			resumeDev:    "UUID=new-uuid",
			resumeOffset: "456789",
			expected:     `GRUB_CMDLINE_LINUX_DEFAULT="loglevel=4 splash resume=UUID=new-uuid resume_offset=456789"`,
		},
		{
			name:         "remove resume and offset completely",
			input:        `GRUB_CMDLINE_LINUX_DEFAULT="loglevel=4 resume=UUID=old-uuid resume_offset=9999 splash"`,
			resumeDev:    "",
			resumeOffset: "",
			expected:     `GRUB_CMDLINE_LINUX_DEFAULT="loglevel=4 splash"`,
		},
		{
			name:         "preserve comments and other lines",
			input:        "# GRUB Configuration\nGRUB_TIMEOUT=5\nGRUB_CMDLINE_LINUX_DEFAULT=\"quiet\"\nGRUB_DISABLE_OS_PROBER=false\n",
			resumeDev:    "/dev/mapper/cryptroot",
			resumeOffset: "777",
			expected:     "# GRUB Configuration\nGRUB_TIMEOUT=5\nGRUB_CMDLINE_LINUX_DEFAULT=\"quiet resume=/dev/mapper/cryptroot resume_offset=777\"\nGRUB_DISABLE_OS_PROBER=false\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := updateGrubCmdline(tc.input, tc.resumeDev, tc.resumeOffset)
			if got != tc.expected {
				t.Errorf("updateGrubCmdline() mismatch:\nGot:  %q\nWant: %q", got, tc.expected)
			}
		})
	}
}

func TestParseBtrfsSwapfileOffset(t *testing.T) {
	validOutputs := []string{
		"12345678\n",
		" 12345678 ",
		"12345678",
	}

	for _, out := range validOutputs {
		offset, err := parseBtrfsSwapfileOffset(out)
		if err != nil {
			t.Fatalf("unexpected error parsing %q: %v", out, err)
		}
		if offset != "12345678" {
			t.Errorf("expected 12345678, got %s", offset)
		}
	}

	invalidOutputs := []string{
		"",
		"ERROR: not a swapfile",
		"abc",
	}

	for _, out := range invalidOutputs {
		if _, err := parseBtrfsSwapfileOffset(out); err == nil {
			t.Errorf("expected error parsing invalid output %q, got nil", out)
		}
	}
}

func TestParseExt4SwapfileOffset(t *testing.T) {
	filefragOutput := `Filesystem type is: ef53
File size of /swapfile is 4294967296 (1048576 blocks of 4096 bytes)
 ext:     logical_offset:        physical_offset: length:   expected: flags:
   0:        0..   32767:      34816..     67583:  32768:            
   1:    32768..   65535:      67584..    100351:  32768:            
`
	offset, err := parseExt4SwapfileOffset(filefragOutput)
	if err != nil {
		t.Fatalf("unexpected error parsing filefrag output: %v", err)
	}
	if offset != "34816" {
		t.Errorf("expected physical offset 34816, got %s", offset)
	}

	if _, err := parseExt4SwapfileOffset("invalid filefrag"); err == nil {
		t.Error("expected error parsing invalid filefrag output, got nil")
	}
}

func TestParseMemTotalMB(t *testing.T) {
	meminfo := `MemTotal:       16301384 kB
MemFree:         4123456 kB
MemAvailable:    8901234 kB
Buffers:          345678 kB
`
	mb, err := parseMemTotalMB(meminfo)
	if err != nil {
		t.Fatalf("parseMemTotalMB: %v", err)
	}
	expected := uint64(16301384 / 1024)
	if mb != expected {
		t.Errorf("got %d MB, want %d MB", mb, expected)
	}
}

func TestIsSecureBootLockedDown(t *testing.T) {
	if !isSecureBootLockedDown("none [integrity] confidentiality") {
		t.Error("expected true for active integrity lockdown")
	}
	if !isSecureBootLockedDown("none integrity [confidentiality]") {
		t.Error("expected true for active confidentiality lockdown")
	}
	if isSecureBootLockedDown("[none] integrity confidentiality") {
		t.Error("expected false for unconfined lockdown")
	}
	if isSecureBootLockedDown("") {
		t.Error("expected false for empty lockdown")
	}
}

func TestIsHibernationSupported(t *testing.T) {
	if !isHibernationSupported("freeze mem disk\n") {
		t.Error("expected true when disk is present in power state")
	}
	if isHibernationSupported("freeze mem\n") {
		t.Error("expected false when disk is absent from power state")
	}
}
