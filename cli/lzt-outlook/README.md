# `cli/lzt-outlook/` — unified CLI (dispatcher)

This subtree contains the **unified CLI dispatcher** that routes
calls to either Path A (`graph/outlook-mcp-go`) or Path C
(`playwright/outlook-scraper-go`) based on `~/.config/lzt-outlook/config.toml`.

## What lives here (planned, Phase 4)

```
cli/lzt-outlook/
├── cmd/lzt-outlook/main.go         # CLI entry point
├── internal/
│   ├── dispatch/                   # Read config.toml + route to active path
│   ├── envelope/                   # JSON envelope (success/data/error/metadata)
│   ├── auth/                       # device-code + browser-session commands
│   └── shared/                     # args parser, validation, formatting
├── tests/                          # dispatcher unit tests
└── smoke/                          # end-to-end against sacrificial account
```

## Why a dispatcher

Two reasons:

1. **Single binary.** Users install one CLI (`lzt-outlook`), not two.
   They pick a path via config and the dispatcher routes internally.
2. **Same JSON contract.** The dispatcher emits the same JSON
   envelope regardless of which path handles the call. MCP servers
   consuming the envelope don't care about the underlying path.

## Status

This subtree is **structurally scaffolded only** as of Phase 0. No
Go code yet. Lands in Phase 4 (~1 hour of work).

## Config dispatch example

```toml
# ~/.config/lzt-outlook/config.toml
[server]
path = "graph"     # or "playwright"

[accounts]
allowed = ["personal", "sacrificial"]
default = "personal"
```

The dispatcher reads `[server].path` at startup and imports the
matching backend package. Calls like `lzt-outlook mail list
--account personal` are translated to a backend-specific request.

## CLI surface (planned)

See [`/ARCHITECTURE.md`](../ARCHITECTURE.md) § Unified CLI dispatch
for the full surface. Phase 5 will publish a `--help` reference.

## License

Apache-2.0, same as the repo root. See [`/LICENSE`](../LICENSE).