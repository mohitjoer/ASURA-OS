# toren-prowl (Code Index, Secret Masking & Zero-LLM Router)

`toren-prowl` is installed at `/usr/bin/toren-prowl`. Use it to search codebases fast via its per-project SQLite FTS5 index, protect secrets, and route LLM requests without router overhead.

## Commands

```sh
toren-prowl index --root .                  # Index repo into ~/.cache/prowl/<hash>.db (secrets masked)
toren-prowl search --root . <query>         # Cited file:line-range search
toren-prowl stats --root .                  # Show indexed files, chunks, and masked secrets
toren-prowl wire --installed                # Register MCP into installed coding tools
toren-prowl status                          # Check installed & wired coding tools
toren-prowl serve --root .                  # Start localhost dashboard + OpenAI/Anthropic router (127.0.0.1:4317)
toren-prowl route "<prompt>"                # Dry-run 24 KiB Zero-LLM domain/effort classifier
```

## Rules

1. Prefer `toren-prowl search` (or MCP `prowl_search` / `prowl_read`) over broad `grep` loops when orienting in a repository.
2. Never disable secret masking; vendor API keys and PEM private keys are redacted before SQLite writes and `prowl_read` output.
3. Router pool settings live in `~/.config/prowl/config.json`.
