package prowl

import (
	"strings"
	"testing"
)

func TestMaskSecrets(t *testing.T) {
	openAI := "sk-" + "proj-1234567890abcdefghijklmnopqrstuvwxyzABCD"
	anthropic := "sk-" + "ant-api03-1234567890abcdefghijklmnopqrstuvwxyzWXYZ"
	google := "AIza" + "SyDummyKey1234567890123456789012345"
	github := "ghp_" + "1234567890abcdefghijklmnopqrstuvwxyz12"
	aws := "AKIA" + "IOSFODNN7EXAMPLE"
	openRouter := "sk-or-v1-" + "0123456789abcdef0123456789abcdef0123456789abcdef"
	pem := "-----BEGIN RSA " + "PRIVATE KEY-----\nMIIEowIBAAKCAQEAtest\n-----END RSA " + "PRIVATE KEY-----"

	input := strings.Join([]string{
		`OPENAI_KEY="` + openAI + `"`,
		`ANTHROPIC_KEY="` + anthropic + `"`,
		`GOOGLE_KEY="` + google + `"`,
		`GITHUB_TOKEN="` + github + `"`,
		`AWS_KEY="` + aws + `"`,
		`OPENROUTER_KEY="` + openRouter + `"`,
		`SAFE_ID="sk_normal_variable_name_not_a_secret"`,
		pem,
	}, "\n")

	masked, report := MaskSecrets(input)
	if report.Total != 7 {
		t.Fatalf("expected 7 masked secrets, got %d (vendors: %v)", report.Total, report.Vendors)
	}

	forbidden := []string{
		"1234567890abcdefghijklmnopqrstuvwxyz",
		google,
		github,
		aws,
		"0123456789abcdef0123456789abcdef",
		"MIIEowIBAAKCAQEAtest",
	}
	for _, f := range forbidden {
		if strings.Contains(masked, f) {
			t.Errorf("secret %q leaked in masked output:\n%s", f, masked)
		}
	}

	if !strings.Contains(masked, `SAFE_ID="sk_normal_variable_name_not_a_secret"`) {
		t.Errorf("non-secret identifier was modified:\n%s", masked)
	}
}
