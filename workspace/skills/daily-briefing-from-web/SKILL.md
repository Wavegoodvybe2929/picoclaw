---
name: daily-briefing-from-web
description: Generate a daily briefing using local workspace data plus optional web search via ./bin/search (no web_fetch).
---

# SKILL: daily_briefing_from_web

## Goal
Create a short daily briefing using a web search + your priorities.

## Inputs
- Topic (optional; if missing, use: "AI tools, local agents, dev workflows")
- Top (optional, default 8)

## Files
- Writes: `web/last_search.json` (optional)



## Tool Policy
- Never use built-in web tools (search/fetch). They are disabled.
- Use exec with the workspace SearXNG CLI: `./bin/search "<query>" --top <N> [--out web/last_search.json]`.
- Prefer snippets; do not fetch pages unless the user explicitly asks.
- Fallback: if `./bin/search` is unavailable, use this skill’s bundled script in `scripts/bin/search`.

## Procedure
1) Run a search:
   exec: `./bin/search "<Topic>" --top <Top> --out web/last_search.json`
2) Produce:
   - 5 key updates
   - 3 things to watch
   - 3 actions for today

## Output format
**Key updates**
- ...

**Watchlist**
- ...

**Today’s actions**
- ...
