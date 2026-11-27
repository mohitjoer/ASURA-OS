package prowl

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
)

const cliUsage = `prowl: secret-masked SQLite code index (CLI + MCP) & Zero-LLM smart router

Usage:
  prowl index  [--root .] [--db PATH]            Index workspace into SQLite (secrets masked)
  prowl search [--root .] [--db PATH] <query>    Search cited code index
  prowl stats  [--root .] [--db PATH]            Show index statistics
  prowl mcp    [--root .] [--db PATH]            Serve cited index & router over MCP stdio
  prowl serve  [--addr 127.0.0.1:4317]           Start OpenAI + Anthropic compatible router proxy
  prowl route  <prompt>                          Dry-run 24 KiB Zero-LLM classifier & router
  prowl setup                                    Print config snippets for OpenCode, Codex, Claude Code, Gemini
`

// Run executes the prowl CLI subcommand.
func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		_, _ = fmt.Fprint(stdout, cliUsage)
		return 2
	}
	sub := args[0]
	rest := args[1:]

	switch sub {
	case "index":
		return runIndex(rest, stdout, stderr)
	case "search":
		return runSearch(rest, stdout, stderr)
	case "stats":
		return runStats(rest, stdout, stderr)
	case "mcp":
		return runMCP(rest, os.Stdin, stdout, stderr)
	case "serve":
		return runServe(rest, stdout, stderr)
	case "route":
		return runRoute(rest, stdout, stderr)
	case "setup":
		_, _ = fmt.Fprint(stdout, setupSnippets())
		return 0
	case "-h", "--help", "help":
		_, _ = fmt.Fprint(stdout, cliUsage)
		return 0
	default:
		_, _ = fmt.Fprintf(stderr, "prowl: unknown subcommand %q\n\n%s", sub, cliUsage)
		return 2
	}
}

func defaultDBPath(root string) string {
	abs, err := filepath.Abs(root)
	if err != nil {
		abs = root
	}
	sum := sha256.Sum256([]byte(abs))
	cacheDir, err := os.UserCacheDir()
	if err != nil || cacheDir == "" {
		cacheDir = filepath.Join(os.TempDir(), "prowl-cache")
	}
	return filepath.Join(cacheDir, "prowl", hex.EncodeToString(sum[:8])+".db")
}

func defaultConfigPath() string {
	cfgDir, err := os.UserConfigDir()
	if err != nil || cfgDir == "" {
		return ""
	}
	return filepath.Join(cfgDir, "prowl", "config.json")
}

func loadConfig(path string) Config {
	cfg := DefaultConfig()
	if path == "" {
		path = defaultConfigPath()
	}
	if path == "" {
		return cfg
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return cfg
	}
	var loaded Config
	if err := json.Unmarshal(b, &loaded); err == nil {
		if loaded.ListenAddr != "" {
			cfg.ListenAddr = loaded.ListenAddr
		}
		if loaded.DBPath != "" {
			cfg.DBPath = loaded.DBPath
		}
		if len(loaded.Models) > 0 {
			cfg.Models = loaded.Models
		}
	}
	return cfg
}

func runIndex(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("index", flag.ContinueOnError)
	fs.SetOutput(stderr)
	root := fs.String("root", ".", "workspace root")
	dbPath := fs.String("db", "", "sqlite db path")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *dbPath == "" {
		*dbPath = defaultDBPath(*root)
	}
	idx, err := OpenIndex(*dbPath, *root)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}
	defer func() { _ = idx.Close() }()

	res, err := idx.IndexWorkspace(context.Background())
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}
	b, _ := json.MarshalIndent(res, "", "  ")
	_, _ = fmt.Fprintln(stdout, string(b))
	return 0
}

func runSearch(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("search", flag.ContinueOnError)
	fs.SetOutput(stderr)
	root := fs.String("root", ".", "workspace root")
	dbPath := fs.String("db", "", "sqlite db path")
	limit := fs.Int("limit", 10, "max hits")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	query := strings.TrimSpace(strings.Join(fs.Args(), " "))
	if query == "" {
		_, _ = fmt.Fprintln(stderr, "prowl search: query required")
		return 2
	}
	if *dbPath == "" {
		*dbPath = defaultDBPath(*root)
	}
	idx, err := OpenIndex(*dbPath, *root)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}
	defer func() { _ = idx.Close() }()

	hits, err := idx.Search(context.Background(), query, *limit)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}
	b, _ := json.MarshalIndent(hits, "", "  ")
	_, _ = fmt.Fprintln(stdout, string(b))
	return 0
}

func runStats(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("stats", flag.ContinueOnError)
	fs.SetOutput(stderr)
	root := fs.String("root", ".", "workspace root")
	dbPath := fs.String("db", "", "sqlite db path")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *dbPath == "" {
		*dbPath = defaultDBPath(*root)
	}
	idx, err := OpenIndex(*dbPath, *root)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}
	defer func() { _ = idx.Close() }()

	st, err := idx.Stats(context.Background())
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}
	b, _ := json.MarshalIndent(st, "", "  ")
	_, _ = fmt.Fprintln(stdout, string(b))
	return 0
}

func runMCP(args []string, in io.Reader, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("mcp", flag.ContinueOnError)
	fs.SetOutput(stderr)
	root := fs.String("root", ".", "workspace root")
	dbPath := fs.String("db", "", "sqlite db path")
	cfgPath := fs.String("config", "", "prowl config JSON")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *dbPath == "" {
		*dbPath = defaultDBPath(*root)
	}
	idx, err := OpenIndex(*dbPath, *root)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}
	defer func() { _ = idx.Close() }()

	cfg := loadConfig(*cfgPath)
	srv := NewMCPServer(idx, NewRouter(cfg.Models))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := srv.Serve(ctx, in, stdout); err != nil && ctx.Err() == nil {
		_, _ = fmt.Fprintf(stderr, "prowl mcp: %v\n", err)
		return 1
	}
	return 0
}

func runServe(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	fs.SetOutput(stderr)
	addr := fs.String("addr", "", "listen address (default 127.0.0.1:4317)")
	cfgPath := fs.String("config", "", "prowl config JSON")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	cfg := loadConfig(*cfgPath)
	if *addr != "" {
		cfg.ListenAddr = *addr
	}
	proxy := NewProxyServer(NewRouter(cfg.Models))
	_, _ = fmt.Fprintf(stdout, "prowl router listening on http://%s (OpenAI /v1/chat/completions, /v1/responses; Anthropic /v1/messages)\n", cfg.ListenAddr)
	if err := http.ListenAndServe(cfg.ListenAddr, proxy.Handler()); err != nil {
		_, _ = fmt.Fprintf(stderr, "prowl serve: %v\n", err)
		return 1
	}
	return 0
}

func runRoute(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("route", flag.ContinueOnError)
	fs.SetOutput(stderr)
	cfgPath := fs.String("config", "", "prowl config JSON")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	prompt := strings.Join(fs.Args(), " ")
	if strings.TrimSpace(prompt) == "" {
		_, _ = fmt.Fprintln(stderr, "prowl route: prompt required")
		return 2
	}
	cfg := loadConfig(*cfgPath)
	dec, _, err := NewRouter(cfg.Models).RouteTurn(prompt)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}
	b, _ := json.MarshalIndent(dec, "", "  ")
	_, _ = fmt.Fprintln(stdout, string(b))
	return 0
}

func setupSnippets() string {
	return `# 1. Start the Zero-LLM smart router (optional, for OpenCode / Codex / Claude Code):
prowl serve --addr 127.0.0.1:4317

# 2. Point your coding agents at the router:
# Claude Code:
export ANTHROPIC_BASE_URL="http://127.0.0.1:4317"
# Codex / OpenAI-compatible CLI:
export OPENAI_BASE_URL="http://127.0.0.1:4317/v1"

# 3. MCP server entry (OpenCode ~/.config/opencode/opencode.json, Claude ~/.claude.json, Codex ~/.codex/config.toml):
{
  "mcpServers": {
    "prowl": {
      "command": "prowl",
      "args": ["mcp", "--root", "."]
    }
  }
}
`
}
