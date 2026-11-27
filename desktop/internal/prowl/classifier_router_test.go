package prowl

import (
	"strings"
	"testing"
	"time"
)

func TestClassifierBounded24KiBAndDomains(t *testing.T) {
	huge := strings.Repeat("a", 30*1024)
	cls := Classify(huge)
	if cls.InspectedBytes != MaxClassifierBytes || !cls.Truncated {
		t.Fatalf("expected 24 KiB inspection cap, got %d (truncated=%v)", cls.InspectedBytes, cls.Truncated)
	}

	coding := Classify("Please refactor func HandleAuth in auth.go and add a unit test for the SQLite patch.")
	if coding.Domain != DomainCoding {
		t.Fatalf("expected coding domain, got %s", coding.Domain)
	}

	mathTurn := Classify("Prove that \\sum_{i=1}^n i = \\frac{n(n+1)}{2} using induction.")
	if mathTurn.Domain != DomainMath || mathTurn.Effort != EffortHigh {
		t.Fatalf("expected high-effort math, got %+v", mathTurn)
	}

	agentic := Classify("Run tests, edit the file, search the repo, and fix the CI step by step in the workspace.")
	if agentic.Domain != DomainAgentic {
		t.Fatalf("expected agentic domain, got %s", agentic.Domain)
	}
}

func TestRouterScoresCooldownQuotasAndBenchmarks(t *testing.T) {
	now := time.Now()
	pool := []ModelCandidate{
		{
			ID:               "fast-cheap",
			Provider:         "openai",
			UpstreamModel:    "gpt-4o-mini",
			Enabled:          true,
			InputPricePer1M:  0.15,
			OutputPricePer1M: 0.60,
			MaxRPM:           2,
			Tier:             "fast",
			Benchmarks:       ArtificialAnalysisScores{Intelligence: 60, Coding: 58},
		},
		{
			ID:               "frontier-coder",
			Provider:         "anthropic",
			UpstreamModel:    "claude-sonnet-4-5",
			Enabled:          true,
			InputPricePer1M:  3.0,
			OutputPricePer1M: 15.0,
			MaxRPM:           10,
			Tier:             "frontier",
			Benchmarks:       ArtificialAnalysisScores{Intelligence: 88, Coding: 94, Reasoning: 90},
		},
	}

	r := NewRouter(pool)

	// High-effort coding should prefer frontier-coder due to coding benchmark weight.
	dec, _, err := r.Select(now, Classification{Domain: DomainCoding, Effort: EffortHigh})
	if err != nil {
		t.Fatal(err)
	}
	if dec.Selected.ID != "frontier-coder" {
		t.Fatalf("expected frontier-coder for high-effort coding, got %s (%s)", dec.Selected.ID, dec.Reason)
	}

	// Put frontier-coder in cooldown (429) -> must fail over to fast-cheap.
	r.RecordFailure("frontier-coder", 429, now)
	dec2, _, err := r.Select(now.Add(5*time.Second), Classification{Domain: DomainCoding, Effort: EffortHigh})
	if err != nil {
		t.Fatal(err)
	}
	if dec2.Selected.ID != "fast-cheap" {
		t.Fatalf("expected fallback to fast-cheap during cooldown, got %s", dec2.Selected.ID)
	}

	// Exhaust fast-cheap RPM quota -> no models left.
	r.RecordDispatch("fast-cheap", now.Add(6*time.Second))
	r.RecordDispatch("fast-cheap", now.Add(7*time.Second))
	if _, _, err := r.Select(now.Add(8*time.Second), Classification{Domain: DomainGeneral, Effort: EffortLow}); err == nil {
		t.Fatal("expected quota/cooldown exhaustion error")
	}
}
