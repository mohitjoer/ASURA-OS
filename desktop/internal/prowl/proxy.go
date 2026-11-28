package prowl

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// Config configures Prowl's index, MCP server, and Zero-LLM HTTP router.
type Config struct {
	ListenAddr string           `json:"listen_addr"`
	DBPath     string           `json:"db_path"`
	Models     []ModelCandidate `json:"models"`
}

// DefaultConfig returns a starter configuration compatible with OpenCode, Codex, and Claude Code.
func DefaultConfig() Config {
	return Config{
		ListenAddr: "127.0.0.1:4317",
		Models: []ModelCandidate{
			{
				ID:               "openai-mini",
				Provider:         "openai",
				BaseURL:          "https://api.openai.com/v1",
				APIKeyEnv:        "OPENAI_API_KEY",
				UpstreamModel:    "gpt-4.1-mini",
				Enabled:          true,
				InputPricePer1M:  0.40,
				OutputPricePer1M: 1.60,
				Tier:             "fast",
				Benchmarks:       ArtificialAnalysisScores{Intelligence: 68, Coding: 66, Math: 67},
			},
			{
				ID:               "claude-sonnet",
				Provider:         "anthropic",
				BaseURL:          "https://api.anthropic.com",
				APIKeyEnv:        "ANTHROPIC_API_KEY",
				UpstreamModel:    "claude-sonnet-4-5",
				Enabled:          true,
				InputPricePer1M:  3.00,
				OutputPricePer1M: 15.00,
				Tier:             "frontier",
				Benchmarks:       ArtificialAnalysisScores{Intelligence: 88, Coding: 93, Reasoning: 89},
			},
		},
	}
}

// ProxyServer exposes OpenAI (/v1/chat/completions, /v1/responses, /v1/models)
// and Anthropic (/v1/messages) endpoints backed by Prowl's Zero-LLM Router.
type ProxyServer struct {
	router *Router
	client *http.Client
}

// NewProxyServer creates a ProxyServer for the given Router.
func NewProxyServer(router *Router) *ProxyServer {
	return &ProxyServer{
		router: router,
		client: &http.Client{Timeout: 5 * time.Minute},
	}
}

// Handler returns the HTTP handler for OpenCode, Codex, and Claude Code.
func (p *ProxyServer) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"service":"prowl-router"}`))
	})
	mux.HandleFunc("/v1/models", p.handleModels)
	mux.HandleFunc("/v1/chat/completions", p.handleForward("openai", "/chat/completions"))
	mux.HandleFunc("/v1/responses", p.handleForward("openai", "/responses"))
	mux.HandleFunc("/v1/messages", p.handleForward("anthropic", "/v1/messages"))
	return mux
}

func (p *ProxyServer) handleModels(w http.ResponseWriter, r *http.Request) {
	p.router.mu.Lock()
	pool := make([]ModelCandidate, len(p.router.pool))
	copy(pool, p.router.pool)
	p.router.mu.Unlock()

	type modelObj struct {
		ID      string `json:"id"`
		Object  string `json:"object"`
		OwnedBy string `json:"owned_by"`
	}
	data := []modelObj{{ID: "auto", Object: "model", OwnedBy: "prowl"}}
	for _, m := range pool {
		if m.Enabled {
			data = append(data, modelObj{ID: m.ID, Object: "model", OwnedBy: m.Provider})
		}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"object": "list", "data": data})
}

func (p *ProxyServer) handleForward(endpointKind, suffix string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		rawBody, err := io.ReadAll(io.LimitReader(r.Body, 16*1024*1024))
		if err != nil {
			http.Error(w, "failed to read body", http.StatusBadRequest)
			return
		}

		var payload map[string]any
		if err := json.Unmarshal(rawBody, &payload); err != nil {
			http.Error(w, "invalid json body", http.StatusBadRequest)
			return
		}

		latestTurn := extractLatestTurn(payload)
		decision, ordered, err := p.router.RouteTurn(latestTurn)
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}

		// Filter to candidates matching the requested wire protocol when possible,
		// falling back to all ordered candidates if none match explicitly.
		candidates := filterByEndpoint(ordered, endpointKind)
		if len(candidates) == 0 {
			candidates = ordered
		}

		var lastStatus int
		var lastBody []byte
		var lastHeader http.Header

		for _, cand := range candidates {
			now := time.Now()
			p.router.RecordDispatch(cand.ID, now)

			payload["model"] = cand.UpstreamModel
			fwdBody, err := json.Marshal(payload)
			if err != nil {
				http.Error(w, "marshal error", http.StatusInternalServerError)
				return
			}

			targetURL := strings.TrimRight(cand.BaseURL, "/") + suffix
			req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, targetURL, bytes.NewReader(fwdBody))
			if err != nil {
				p.router.RecordFailure(cand.ID, 500, time.Now())
				continue
			}
			req.Header.Set("Content-Type", "application/json")
			applyUpstreamAuth(req, r, cand)

			start := time.Now()
			resp, err := p.client.Do(req)
			if err != nil {
				p.router.RecordFailure(cand.ID, 502, time.Now())
				lastStatus = http.StatusBadGateway
				lastBody = []byte(err.Error())
				continue
			}

			if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
				b, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
				_ = resp.Body.Close()
				p.router.RecordFailure(cand.ID, resp.StatusCode, time.Now())
				lastStatus = resp.StatusCode
				lastBody = b
				lastHeader = resp.Header
				continue
			}

			p.router.RecordSuccess(cand.ID, time.Since(start))
			for k, vals := range resp.Header {
				for _, v := range vals {
					w.Header().Add(k, v)
				}
			}
			w.Header().Set("X-Prowl-Routed-Model", cand.ID)
			w.Header().Set("X-Prowl-Domain", string(decision.Classification.Domain))
			w.Header().Set("X-Prowl-Effort", string(decision.Classification.Effort))
			w.WriteHeader(resp.StatusCode)
			_, _ = io.Copy(w, resp.Body)
			_ = resp.Body.Close()
			return
		}

		for k, vals := range lastHeader {
			for _, v := range vals {
				w.Header().Add(k, v)
			}
		}
		if lastStatus == 0 {
			lastStatus = http.StatusBadGateway
		}
		w.WriteHeader(lastStatus)
		_, _ = w.Write(lastBody)
	}
}

func filterByEndpoint(ordered []ModelCandidate, kind string) []ModelCandidate {
	var out []ModelCandidate
	for _, c := range ordered {
		if kind == "anthropic" && c.Provider == "anthropic" {
			out = append(out, c)
		} else if kind == "openai" && c.Provider != "anthropic" {
			out = append(out, c)
		}
	}
	return out
}

func applyUpstreamAuth(dst, src *http.Request, cand ModelCandidate) {
	key := cand.APIKey
	if key == "" && cand.APIKeyEnv != "" {
		key = os.Getenv(cand.APIKeyEnv)
	}
	if cand.Provider == "anthropic" {
		if key != "" {
			dst.Header.Set("x-api-key", key)
		} else if v := src.Header.Get("x-api-key"); v != "" {
			dst.Header.Set("x-api-key", v)
		}
		if ver := src.Header.Get("anthropic-version"); ver != "" {
			dst.Header.Set("anthropic-version", ver)
		} else {
			dst.Header.Set("anthropic-version", "2023-06-01")
		}
		return
	}
	if key != "" {
		dst.Header.Set("Authorization", "Bearer "+key)
	} else if auth := src.Header.Get("Authorization"); auth != "" {
		dst.Header.Set("Authorization", auth)
	}
}

func extractLatestTurn(payload map[string]any) string {
	// OpenAI chat/completions & Anthropic messages use "messages"
	if msgs, ok := payload["messages"].([]any); ok {
		for i := len(msgs) - 1; i >= 0; i-- {
			m, ok := msgs[i].(map[string]any)
			if !ok {
				continue
			}
			role, _ := m["role"].(string)
			if role != "" && role != "user" {
				continue
			}
			if txt := extractTextField(m["content"]); txt != "" {
				return txt
			}
		}
	}
	// OpenAI Responses API uses "input"
	if txt := extractTextField(payload["input"]); txt != "" {
		return txt
	}
	return ""
}

func extractTextField(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case []any:
		var parts []string
		for _, item := range x {
			if m, ok := item.(map[string]any); ok {
				if t, ok := m["text"].(string); ok && t != "" {
					parts = append(parts, t)
				}
			}
		}
		return strings.Join(parts, "\n")
	default:
		return ""
	}
}
