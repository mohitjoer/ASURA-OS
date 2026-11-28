package prowl

import (
	"strings"
	"unicode/utf8"
)

// MaxClassifierBytes caps how much of the latest turn the Zero-LLM classifier inspects.
const MaxClassifierBytes = 24 * 1024

// Domain represents the detected task domain.
type Domain string

const (
	DomainCoding    Domain = "coding"
	DomainReasoning Domain = "reasoning"
	DomainAgentic   Domain = "agentic"
	DomainMath      Domain = "math"
	DomainResearch  Domain = "research"
	DomainWriting   Domain = "writing"
	DomainGeneral   Domain = "general"
)

// Effort represents estimated reasoning/compute effort.
type Effort string

const (
	EffortLow    Effort = "low"
	EffortMedium Effort = "medium"
	EffortHigh   Effort = "high"
)

// Classification is the output of the bounded local classifier.
type Classification struct {
	Domain         Domain   `json:"domain"`
	Effort         Effort   `json:"effort"`
	Confidence     float64  `json:"confidence"`
	InspectedBytes int      `json:"inspected_bytes"`
	Truncated      bool     `json:"truncated"`
	Signals        []string `json:"signals,omitempty"`
}

// Classify inspects up to MaxClassifierBytes (24 KiB) of the latest turn
// using deterministic lexical and structural signals (zero LLM calls).
func Classify(turn string) Classification {
	truncated := false
	sample := turn
	if len(sample) > MaxClassifierBytes {
		sample = sample[:MaxClassifierBytes]
		for len(sample) > 0 && !utf8.ValidString(sample) {
			sample = sample[:len(sample)-1]
		}
		truncated = true
	}

	lower := strings.ToLower(sample)
	scores := map[Domain]int{
		DomainCoding:    0,
		DomainReasoning: 0,
		DomainAgentic:   0,
		DomainMath:      0,
		DomainResearch:  0,
		DomainWriting:   0,
	}
	var signals []string

	addSignal := func(d Domain, weight int, label string) {
		scores[d] += weight
		signals = append(signals, label)
	}

	// Coding signals
	if strings.Contains(sample, "```") || strings.Contains(sample, "func ") ||
		strings.Contains(sample, "package ") || strings.Contains(sample, "import ") ||
		strings.Contains(sample, "def ") || strings.Contains(sample, "const ") {
		addSignal(DomainCoding, 4, "code_syntax")
	}
	for _, kw := range []string{
		"refactor", "compile", "stack trace", "panic:", "traceback", "segfault",
		"unit test", "dockerfile", "sql", "regex", "typescript", "golang", "rust",
		"patch", "diff", "bug", "endpoint", "sqlite", "mcp",
	} {
		if strings.Contains(lower, kw) {
			addSignal(DomainCoding, 2, "kw:"+kw)
		}
	}

	// Agentic / tool-use signals
	for _, kw := range []string{
		"run tests", "edit the file", "fix the ci", "search the repo", "create a pr",
		"step by step", "execute", "deploy", "migrate", "orchestrate", "subagent",
		"tool_call", "workspace", "terminal",
	} {
		if strings.Contains(lower, kw) {
			addSignal(DomainAgentic, 3, "agentic:"+kw)
		}
	}

	// Math signals
	if strings.Contains(sample, "$$") || strings.Contains(sample, `\frac`) ||
		strings.Contains(sample, `\sum`) || strings.Contains(sample, `\int`) {
		addSignal(DomainMath, 5, "latex_math")
	}
	for _, kw := range []string{
		"prove that", "theorem", "eigenvalue", "derivative", "integral", "matrix",
		"probability", "combinatorics", "modulo", "polynomial", "equation",
	} {
		if strings.Contains(lower, kw) {
			addSignal(DomainMath, 3, "math:"+kw)
		}
	}

	// Reasoning signals
	for _, kw := range []string{
		"why does", "tradeoff", "trade-off", "architecture", "root cause",
		"analyze", "deduce", "constraint", "invariant", "edge case", "compare",
	} {
		if strings.Contains(lower, kw) {
			addSignal(DomainReasoning, 2, "reason:"+kw)
		}
	}

	// Research signals
	for _, kw := range []string{
		"benchmark", "paper", "citation", "literature", "survey", "docs",
		"documentation", "latest version", "release notes", "what changed",
	} {
		if strings.Contains(lower, kw) {
			addSignal(DomainResearch, 2, "research:"+kw)
		}
	}

	// Writing signals
	for _, kw := range []string{
		"draft an email", "rewrite", "summarize", "blog post", "readme",
		"tone", "grammar", "changelog",
	} {
		if strings.Contains(lower, kw) {
			addSignal(DomainWriting, 2, "writing:"+kw)
		}
	}

	bestDomain := DomainGeneral
	bestScore := 0
	secondScore := 0
	for d, sc := range scores {
		if sc > bestScore {
			secondScore = bestScore
			bestScore = sc
			bestDomain = d
		} else if sc > secondScore {
			secondScore = sc
		}
	}

	conf := 0.5
	if bestScore > 0 {
		margin := float64(bestScore - secondScore)
		conf = 0.55 + minFloat(0.42, margin*0.08+float64(bestScore)*0.03)
	}

	effort := estimateEffort(sample, lower, bestDomain, bestScore)

	return Classification{
		Domain:         bestDomain,
		Effort:         effort,
		Confidence:     conf,
		InspectedBytes: len(sample),
		Truncated:      truncated,
		Signals:        signals,
	}
}

func estimateEffort(sample, lower string, domain Domain, score int) Effort {
	n := len(sample)
	hardCues := 0
	for _, kw := range []string{
		"deep", "thorough", "all edge cases", "end-to-end", "prove",
		"architecture", "security", "concurrency", "race condition", "optimize",
	} {
		if strings.Contains(lower, kw) {
			hardCues++
		}
	}
	if n > 2048 || hardCues >= 2 || (domain == DomainMath && score >= 5) || (domain == DomainAgentic && score >= 6) {
		return EffortHigh
	}
	if n > 320 || score >= 4 || hardCues == 1 {
		return EffortMedium
	}
	return EffortLow
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
