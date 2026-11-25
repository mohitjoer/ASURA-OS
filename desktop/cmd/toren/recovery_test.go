package main

import (
	"bytes"
	"strings"
	"testing"
)

// TestRecoveryMenuChoices verifies that each choice in the recovery menu:
// 1. Describes what it will do and asks for confirmation
// 2. Dispatches to the corresponding recovery handler when confirmed
// 3. Exits cleanly on option 7 or 'q'
func TestRecoveryMenuChoices(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantCalls  string
		wantOutput string
	}{
		{
			name:       "choice 1: doctor confirmed",
			input:      "1\ny\n7\n",
			wantCalls:  "doctor",
			wantOutput: "[Option 1: System Doctor]",
		},
		{
			name:       "choice 1: doctor declined",
			input:      "1\nn\n7\n",
			wantCalls:  "",
			wantOutput: "[Option 1: System Doctor]",
		},
		{
			name:       "choice 2: reset confirmed",
			input:      "2\ny\n7\n",
			wantCalls:  "reset",
			wantOutput: "[Option 2: Reset Desktop Configuration]",
		},
		{
			name:       "choice 3: reinstall confirmed",
			input:      "3\ny\n7\n",
			wantCalls:  "reinstall",
			wantOutput: "[Option 3: Reinstall Shipped ASURA Files]",
		},
		{
			name:       "choice 4: snapshot rollback confirmed",
			input:      "4\ny\n7\n",
			wantCalls:  "rollback",
			wantOutput: "[Option 4: System Snapshots]",
		},
		{
			name:       "choice 5: safe-mode confirmed",
			input:      "5\ny\n7\n",
			wantCalls:  "safemode",
			wantOutput: "[Option 5: Safe-Mode Session]",
		},
		{
			name:       "choice 6: show logs confirmed",
			input:      "6\ny\n7\n",
			wantCalls:  "logs",
			wantOutput: "[Option 6: View System & Session Logs]",
		},
		{
			name:       "choice 7: exit immediately",
			input:      "7\n",
			wantCalls:  "",
			wantOutput: "Exiting recovery menu.",
		},
		{
			name:       "quit with q",
			input:      "q\n",
			wantCalls:  "",
			wantOutput: "Exiting recovery menu.",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var called []string
			var outBuf bytes.Buffer

			runner := &recoveryRunner{
				in:  strings.NewReader(tc.input),
				out: &outBuf,
				doctor: func() int {
					called = append(called, "doctor")
					return exitOK
				},
				reset: func() int {
					called = append(called, "reset")
					return exitOK
				},
				reinstall: func() int {
					called = append(called, "reinstall")
					return exitOK
				},
				rollback: func() int {
					called = append(called, "rollback")
					return exitOK
				},
				safeMode: func() int {
					called = append(called, "safemode")
					return exitOK
				},
				showLogs: func() int {
					called = append(called, "logs")
					return exitOK
				},
			}

			code := runRecoveryMenu(runner)
			if code != exitOK {
				t.Errorf("runRecoveryMenu exited with %d, want %d", code, exitOK)
			}

			gotCalls := strings.Join(called, ",")
			if gotCalls != tc.wantCalls {
				t.Errorf("handlers called: %q, want %q", gotCalls, tc.wantCalls)
			}

			if !strings.Contains(outBuf.String(), tc.wantOutput) {
				t.Errorf("output missing expected string %q\nFull output:\n%s", tc.wantOutput, outBuf.String())
			}
		})
	}
}
