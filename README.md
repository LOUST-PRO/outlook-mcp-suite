# outlook-mcp-suite

A multi-path AI-friendly bridge between MCP-equipped agents and an
Outlook / Microsoft 365 account. Two implementations share the same
surface contract: **a JSON-on-stdio MCP server** plus a **unified CLI**
that dispatches to the active path.

## What's in the box

```
outlook-mcp-suite/
├── graph/      Path A: Microsoft Graph (official, primary, maintained)
├── playwright/ Path C: web.outlook.com scraping (community, may break)
├── cli/        `lzt-outlook` unified CLI (dispatch)
├── docs/       Onboarding, shadow-mode, ledger-format, performance-cost
└── .claude/    Hooks + skills (Claude Code convention)
```

Read [`LEGAL-NOTICE.md`](./LEGAL-NOTICE.md) before doing anything.

## Quick start (Path A, MSA personal account)

```bash
# 1. Install (once path A is in release)
curl -L -o lzt-outlook.tar.gz \
  https://github.com/.../outlook-mcp-suite/releases/download/v0.1.0/lzt-outlook-graph-v0.1.0-linux-amd64.tar.gz
tar -xzf lzt-outlook.tar.gz
sudo mv lzt-outlook /usr/local/bin/

# 2. Register personal Azure app (5 minutes, manual, one-time)
#    → https://portal.azure.com/#blade/Microsoft_AAD_RegisteredApps/ApplicationsListBlade
#    → "New registration" → name "outlook-mcp-personal"
#    → "Authentication" → enable "Device code flow" (public client)
#    → "API permissions" → Microsoft Graph → delegated:
#         Mail.Read, Mail.ReadWrite, Calendars.Read, Calendars.ReadWrite,
#         User.Read
#    → copy Application (client) ID

# 3. Configure
mkdir -p ~/.config/lzt-outlook
cat > ~/.config/lzt-outlook/config.toml <<'EOF'
[server]
path = "graph"

[accounts]
allowed = ["personal", "sacrificial"]
default = "personal"

[scopes]
graph = [
  "Mail.Read",
  "Mail.ReadWrite",
  "Calendars.Read",
  "Calendars.ReadWrite",
  "User.Read"
]

[apply]
rate_limit_per_5min = 10
require_literal_phrase = "outlook apply: "

[html_sanitizer]
tracking_pixel_blocklist = [
  "*.google-analytics.com",
  "*.scorecardresearch.com",
  "*.doubleclick.net",
  "*.facebook.com/tr",
  "*.linkedin.com/li/track"
]
iframe_external_blocklist = []
EOF

# 4. Device-code login (one-time, shows URL + code)
lzt-outlook auth login --account personal --client-id <YOUR_CLIENT_ID>

# 5. Smoke
lzt-outlook mail folders --account personal
lzt-outlook mail list --account personal --limit 5
lzt-outlook calendar list --account personal --start today --end "+7d"

# 6. (Optional) wire into Claude Code
#    Add to ~/.claude/settings.json mcpServers:
#      "outlook": { "command": "/usr/local/bin/lzt-outlook", "args": ["mcp", "serve"] }
```

## Status: PHASE 0 SCAFFOLD

This repository is in **Phase 0 (scaffold only)**. No path is implemented yet.

| Path | Status | Plan |
|---|---|---|
| `graph/` | empty dir, no code | Phase 1 (~4h) — read-only Go wrapper, 11 tools |
| `playwright/` | empty dir, no code | Phase 3 (~30m) — structure + LEGAL.md only |
| `cli/lzt-outlook/` | empty dir, no code | Phase 4 (~1h) — dispatcher |
| `.claude/hooks/` | spec only, no script | Phase 0.5 (~1h) — 4 hooks shell scripts |
| `docs/` | empty dir | Phase 5 — onboarding, shadow-mode, ledger-format, perf-cost |

See [`docs/`] (placeholder) or the project issues for the current status
of each phase.

## Why two paths?

Microsoft Graph (Path A) is the official API but requires a one-time app
registration in the Azure portal. For users who can't or won't do that,
the only remaining option is browser automation against
`outlook.live.com` / `outlook.office.com` (Path C). Path C is fragile
and has ToS implications — see `LEGAL-NOTICE.md` before considering it.

## License

Apache License 2.0. See [`LICENSE`](./LICENSE).

Patent grant clause (Section 3) was chosen deliberately because OAuth
flows against Microsoft Graph touch Microsoft's patent portfolio; the
grant protects contributors from patent-claim trolling by any party,
including but not limited to Microsoft.

## Contributing

See [`CONTRIBUTING.md`](./CONTRIBUTING.md). This repository operates
under a **delegated PR-approval model** with strict criteria — read it
before opening a PR.
