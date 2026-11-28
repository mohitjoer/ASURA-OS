---
name: toren-prowl
description: >
  Use toren-prowl for fast cited SQLite code search (with automatic vendor
  secret masking), MCP tool wiring across coding agents, and Zero-LLM smart
  model routing with a localhost dashboard.
---

# toren-prowl

`toren-prowl` (`/usr/bin/toren-prowl`) provides two capabilities for coding agents:

1. **Cited SQLite Code Index + Secret Masking (`index`, `search`, `mcp`)**
   - Stores an isolated FTS5 index per project under `~/.cache/prowl/<project-hash>.db`.
   - Masks unambiguous vendor API keys (OpenAI, Anthropic, Google, GitHub, AWS, Slack, Stripe, OpenRouter, PEM keys, etc.) before writing to SQLite or returning snippets.
2. **Zero-LLM Smart Routing + Localhost Dashboard (`serve`, `route`)**
   - Inspects up to 24 KiB of the latest turn locally (0 ms network, \$0) to detect domain and effort, then routes across enabled models in `~/.config/prowl/config.json` using live latency, quotas, cooldowns, pricing, and Artificial Analysis benchmarks.
   - Exposes OpenAI (`/v1/chat/completions`, `/v1/responses`, `/v1/models`), Anthropic (`/v1/messages`), and a web UI at `http://127.0.0.1:4317/`.

## Common Commands

```sh
toren-prowl wire --installed   # Wire MCP into Claude Code, OpenCode, Codex, Gemini, Cursor, Windsurf, Zed, Antigravity
toren-prowl index --root .     # Index current repository
toren-prowl search <query>     # Cited search (path:start-end)
toren-prowl serve --root .     # Start localhost dashboard + router proxy
```
