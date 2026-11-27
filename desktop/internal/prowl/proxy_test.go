package prowl

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProxyRoutesAndFailsOverOn429(t *testing.T) {
	rateLimited := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":"rate_limited"}`))
	}))
	defer rateLimited.Close()

	healthy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		_ = json.NewDecoder(r.Body).Decode(&req)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":    "chatcmpl-1",
			"model": req["model"],
		})
	}))
	defer healthy.Close()

	router := NewRouter([]ModelCandidate{
		{
			ID:            "primary",
			Provider:      "openai",
			BaseURL:       rateLimited.URL,
			UpstreamModel: "gpt-primary",
			Enabled:       true,
			Tier:          "frontier",
			Benchmarks:    ArtificialAnalysisScores{Coding: 95},
		},
		{
			ID:            "backup",
			Provider:      "openai",
			BaseURL:       healthy.URL,
			UpstreamModel: "gpt-backup",
			Enabled:       true,
			Tier:          "balanced",
			Benchmarks:    ArtificialAnalysisScores{Coding: 80},
		},
	})

	proxy := NewProxyServer(router)
	body := []byte(`{"model":"auto","messages":[{"role":"user","content":"refactor func main() in go"}]}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	proxy.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 after failover, got %d: %s", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("X-Prowl-Routed-Model"); got != "backup" {
		t.Fatalf("expected X-Prowl-Routed-Model=backup, got %q", got)
	}
	if got := rec.Header().Get("X-Prowl-Domain"); got != "coding" {
		t.Fatalf("expected X-Prowl-Domain=coding, got %q", got)
	}
}
