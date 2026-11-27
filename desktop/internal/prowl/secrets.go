package prowl

import (
	"fmt"
	"regexp"
	"strings"
)

type secretPattern struct {
	vendor string
	re     *regexp.Regexp
	pem    bool
}

var secretPatterns = []secretPattern{
	{
		vendor: "private_key",
		re:     regexp.MustCompile(`-----BEGIN (?:[A-Z0-9 ]+ )?PRIVATE KEY-----[\s\S]*?-----END (?:[A-Z0-9 ]+ )?PRIVATE KEY-----`),
		pem:    true,
	},
	{
		vendor: "anthropic",
		re:     regexp.MustCompile(`\bsk-ant-(?:api\d{2}|admin\d{2}|sid\d{2})?-[A-Za-z0-9_-]{20,}`),
	},
	{
		vendor: "openrouter",
		re:     regexp.MustCompile(`\bsk-or-v1-[0-9a-fA-F]{40,}\b`),
	},
	{
		vendor: "openai",
		re:     regexp.MustCompile(`\bsk-(?:proj|svcacct|admin)-[A-Za-z0-9_-]{20,}\b|\bsk-[A-Za-z0-9]{20}T3BlbkFJ[A-Za-z0-9]{20}\b`),
	},
	{
		vendor: "google",
		re:     regexp.MustCompile(`\bAIza[0-9A-Za-z_-]{35}\b`),
	},
	{
		vendor: "github",
		re:     regexp.MustCompile(`\b(?:ghp|gho|ghu|ghs|ghr)_[A-Za-z0-9_]{36,}\b|\bgithub_pat_[A-Za-z0-9_]{22,}\b`),
	},
	{
		vendor: "aws",
		re:     regexp.MustCompile(`\b(?:AKIA|ASIA|AROA|AIDA)[0-9A-Z]{16}\b`),
	},
	{
		vendor: "slack",
		re:     regexp.MustCompile(`\bxox[bpars]-[0-9]{8,}-[0-9A-Za-z-]{10,}\b`),
	},
	{
		vendor: "stripe",
		re:     regexp.MustCompile(`\b(?:sk|rk)_live_[0-9a-zA-Z]{24,}\b`),
	},
	{
		vendor: "huggingface",
		re:     regexp.MustCompile(`\bhf_[A-Za-z0-9]{24,}\b`),
	},
	{
		vendor: "groq",
		re:     regexp.MustCompile(`\bgsk_[A-Za-z0-9]{32,}\b`),
	},
	{
		vendor: "perplexity",
		re:     regexp.MustCompile(`\bpplx-[A-Za-z0-9]{32,}\b`),
	},
	{
		vendor: "xai",
		re:     regexp.MustCompile(`\bxai-[A-Za-z0-9]{32,}\b`),
	},
	{
		vendor: "npm",
		re:     regexp.MustCompile(`\bnpm_[A-Za-z0-9]{36}\b`),
	},
	{
		vendor: "pypi",
		re:     regexp.MustCompile(`\bpypi-AgEI[A-Za-z0-9_-]{30,}\b`),
	},
}

// MaskReport summarizes how many secrets were masked by vendor.
type MaskReport struct {
	Total   int
	Vendors map[string]int
}

// MaskSecrets redacts unambiguous vendor secrets and API keys from s before
// anything is written to the SQLite index or returned over MCP.
func MaskSecrets(s string) (string, MaskReport) {
	report := MaskReport{Vendors: map[string]int{}}
	out := s
	for _, pat := range secretPatterns {
		out = pat.re.ReplaceAllStringFunc(out, func(match string) string {
			report.Total++
			report.Vendors[pat.vendor]++
			if pat.pem {
				return fmt.Sprintf("[REDACTED:%s]", pat.vendor)
			}
			return maskToken(pat.vendor, match)
		})
	}
	if len(report.Vendors) == 0 {
		report.Vendors = nil
	}
	return out, report
}

func maskToken(vendor, tok string) string {
	if len(tok) <= 10 {
		return fmt.Sprintf("[REDACTED:%s]", vendor)
	}
	// Preserve a short non-sensitive prefix + last 4 chars so code citations remain debuggable.
	prefixLen := 6
	if idx := strings.IndexByte(tok, '_'); idx > 0 && idx < 11 {
		prefixLen = idx + 1
	} else if strings.HasPrefix(tok, "sk-ant-") {
		prefixLen = 7
	} else if strings.HasPrefix(tok, "sk-or-v1-") {
		prefixLen = 9
	} else if strings.HasPrefix(tok, "sk-proj-") {
		prefixLen = 8
	}
	if prefixLen >= len(tok)-4 {
		return fmt.Sprintf("[REDACTED:%s]", vendor)
	}
	return fmt.Sprintf("%s***REDACTED:%s***%s", tok[:prefixLen], vendor, tok[len(tok)-4:])
}
