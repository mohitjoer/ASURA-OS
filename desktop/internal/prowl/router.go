package prowl

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"
)

// ArtificialAnalysisScores holds optional benchmark scores (0..100 scale)
// from Artificial Analysis for a model.
type ArtificialAnalysisScores struct {
	Intelligence float64 `json:"intelligence,omitempty"`
	Coding       float64 `json:"coding,omitempty"`
	Math         float64 `json:"math,omitempty"`
	Reasoning    float64 `json:"reasoning,omitempty"`
}

// ModelCandidate describes one enabled model in the user's routing pool.
type ModelCandidate struct {
	ID               string                   `json:"id"`
	Provider         string                   `json:"provider"` // openai | anthropic | openrouter | google | custom
	BaseURL          string                   `json:"base_url"`
	APIKeyEnv        string                   `json:"api_key_env,omitempty"`
	APIKey           string                   `json:"api_key,omitempty"`
	UpstreamModel    string                   `json:"upstream_model"`
	Enabled          bool                     `json:"enabled"`
	InputPricePer1M  float64                  `json:"input_price_per_1m"`
	OutputPricePer1M float64                  `json:"output_price_per_1m"`
	MaxRPM           int                      `json:"max_rpm,omitempty"`
	MaxDailyRequests int                      `json:"max_daily_requests,omitempty"`
	Tier             string                   `json:"tier,omitempty"` // fast | balanced | frontier
	Benchmarks       ArtificialAnalysisScores `json:"benchmarks,omitempty"`
}

// RouteDecision explains which model was chosen and why.
type RouteDecision struct {
	Classification Classification `json:"classification"`
	Selected       ModelCandidate `json:"selected"`
	Score          float64        `json:"score"`
	Reason         string         `json:"reason"`
	Alternatives   []string       `json:"alternatives,omitempty"`
}

type modelRuntime struct {
	latencyEWMA   float64
	samples       int
	failures      int
	cooldownUntil time.Time
	rpmTimes      []time.Time
	dayKey        string
	dayCount      int
}

// Router performs Zero-LLM pool routing using classification, live latency,
// quotas, cooldowns, pricing, and optional Artificial Analysis scores.
type Router struct {
	mu      sync.Mutex
	pool    []ModelCandidate
	runtime map[string]*modelRuntime
}

// NewRouter creates a Router for the given model pool.
func NewRouter(pool []ModelCandidate) *Router {
	copied := make([]ModelCandidate, len(pool))
	copy(copied, pool)
	return &Router{
		pool:    copied,
		runtime: map[string]*modelRuntime{},
	}
}

// RouteTurn classifies the latest turn (capped at 24 KiB) and picks the best
// available model plus ordered fallbacks.
func (r *Router) RouteTurn(latestTurn string) (RouteDecision, []ModelCandidate, error) {
	cls := Classify(latestTurn)
	return r.Select(time.Now(), cls)
}

// Select scores enabled models and returns the top choice plus ordered fallbacks.
func (r *Router) Select(now time.Time, cls Classification) (RouteDecision, []ModelCandidate, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	type scored struct {
		cand   ModelCandidate
		score  float64
		reason string
	}
	var available []scored
	dayKey := now.UTC().Format("2006-01-02")

	for _, m := range r.pool {
		if !m.Enabled {
			continue
		}
		rt := r.ensureRuntime(m.ID, dayKey)
		if now.Before(rt.cooldownUntil) {
			continue
		}
		rt.pruneRPM(now)
		if m.MaxRPM > 0 && len(rt.rpmTimes) >= m.MaxRPM {
			continue
		}
		if m.MaxDailyRequests > 0 && rt.dayCount >= m.MaxDailyRequests {
			continue
		}

		sc, reason := scoreCandidate(m, rt, cls)
		available = append(available, scored{cand: m, score: sc, reason: reason})
	}

	if len(available) == 0 {
		return RouteDecision{}, nil, errors.New("prowl router: no enabled models available (all in cooldown or over quota)")
	}

	sort.Slice(available, func(i, j int) bool {
		if available[i].score == available[j].score {
			return available[i].cand.ID < available[j].cand.ID
		}
		return available[i].score > available[j].score
	})

	ordered := make([]ModelCandidate, len(available))
	var alts []string
	for i, item := range available {
		ordered[i] = item.cand
		if i > 0 {
			alts = append(alts, item.cand.ID)
		}
	}

	best := available[0]
	return RouteDecision{
		Classification: cls,
		Selected:       best.cand,
		Score:          math.Round(best.score*100) / 100,
		Reason:         best.reason,
		Alternatives:   alts,
	}, ordered, nil
}

// RecordDispatch records an outgoing request against RPM/daily quotas.
func (r *Router) RecordDispatch(modelID string, now time.Time) {
	r.mu.Lock()
	defer r.mu.Unlock()
	rt := r.ensureRuntime(modelID, now.UTC().Format("2006-01-02"))
	rt.rpmTimes = append(rt.rpmTimes, now)
	rt.dayCount++
}

// RecordSuccess updates the EWMA latency and resets failure/cooldown state.
func (r *Router) RecordSuccess(modelID string, latency time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	rt := r.ensureRuntime(modelID, time.Now().UTC().Format("2006-01-02"))
	ms := float64(latency.Milliseconds())
	if ms < 1 {
		ms = 1
	}
	if rt.samples == 0 {
		rt.latencyEWMA = ms
	} else {
		rt.latencyEWMA = 0.7*rt.latencyEWMA + 0.3*ms
	}
	rt.samples++
	rt.failures = 0
	rt.cooldownUntil = time.Time{}
}

// RecordFailure puts a model into exponential cooldown on 429/5xx/transport errors.
func (r *Router) RecordFailure(modelID string, statusCode int, now time.Time) {
	r.mu.Lock()
	defer r.mu.Unlock()
	rt := r.ensureRuntime(modelID, now.UTC().Format("2006-01-02"))
	rt.failures++
	backoff := 15 * time.Second
	if statusCode == 429 {
		backoff = 30 * time.Second
	}
	for i := 1; i < rt.failures && backoff < 5*time.Minute; i++ {
		backoff *= 2
	}
	if backoff > 5*time.Minute {
		backoff = 5 * time.Minute
	}
	rt.cooldownUntil = now.Add(backoff)
}

func (r *Router) ensureRuntime(id, dayKey string) *modelRuntime {
	rt, ok := r.runtime[id]
	if !ok {
		rt = &modelRuntime{dayKey: dayKey}
		r.runtime[id] = rt
		return rt
	}
	if rt.dayKey != dayKey {
		rt.dayKey = dayKey
		rt.dayCount = 0
	}
	return rt
}

func (rt *modelRuntime) pruneRPM(now time.Time) {
	cutoff := now.Add(-time.Minute)
	n := 0
	for _, ts := range rt.rpmTimes {
		if ts.After(cutoff) {
			rt.rpmTimes[n] = ts
			n++
		}
	}
	rt.rpmTimes = rt.rpmTimes[:n]
}

func scoreCandidate(m ModelCandidate, rt *modelRuntime, cls Classification) (float64, string) {
	quality := benchmarkForDomain(m, cls.Domain)

	// Effort determines how much we weight quality vs price/latency.
	qualityWeight, priceWeight, latencyWeight := 0.8, 1.5, 0.015
	switch cls.Effort {
	case EffortHigh:
		qualityWeight, priceWeight, latencyWeight = 1.4, 0.35, 0.006
	case EffortLow:
		qualityWeight, priceWeight, latencyWeight = 0.45, 2.5, 0.025
	}

	blendedPrice := m.InputPricePer1M + 2.0*m.OutputPricePer1M
	pricePenalty := blendedPrice * priceWeight

	latencyPenalty := 0.0
	if rt.samples > 0 {
		latencyPenalty = (rt.latencyEWMA / 100.0) * (latencyWeight * 100.0)
	}

	quotaBonus := 0.0
	if m.MaxRPM > 0 {
		headroom := float64(m.MaxRPM-len(rt.rpmTimes)) / float64(m.MaxRPM)
		quotaBonus += headroom * 5.0
	}

	score := quality*qualityWeight + quotaBonus - pricePenalty - latencyPenalty
	reason := fmt.Sprintf("domain=%s effort=%s quality=%.1f price_pen=%.1f lat_pen=%.1f",
		cls.Domain, cls.Effort, quality, pricePenalty, latencyPenalty)
	return score, reason
}

func benchmarkForDomain(m ModelCandidate, domain Domain) float64 {
	b := m.Benchmarks
	var v float64
	switch domain {
	case DomainCoding, DomainAgentic:
		v = firstPositive(b.Coding, b.Intelligence)
	case DomainMath:
		v = firstPositive(b.Math, b.Reasoning, b.Intelligence)
	case DomainReasoning, DomainResearch:
		v = firstPositive(b.Reasoning, b.Intelligence)
	default:
		v = b.Intelligence
	}
	if v > 0 {
		return v
	}
	switch m.Tier {
	case "frontier":
		return 82
	case "balanced":
		return 68
	case "fast":
		return 55
	default:
		return 60
	}
}

func firstPositive(vals ...float64) float64 {
	for _, v := range vals {
		if v > 0 {
			return v
		}
	}
	return 0
}
