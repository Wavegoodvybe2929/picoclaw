---
name: searxng-cli-conventions
description: Explain and enforce conventions for using ./bin/search (SearXNG CLI) and saving results locally.
---

# SKILL: searxng_cli_conventions

## Purpose
Teach the agent the local web-search system and how to keep commands short.

## System
- Web search is done by running: `skills/searxng-cli-conventions/scripts/bin/search "<query>" --top N [--out web/last_search.json]`
- This calls your local SearXNG at `http://localhost:8080` and returns:
  - Human-readable results (default), OR
  - Raw JSON (`--json`)

## Key files
- `web/last_search.json` : last saved JSON results
- `web/notes_last_search.md` : optional summary produced by the agent

## Rules
1) Prefer **short** user-facing instructions. Use the skills below.
2) Default `top = 5` unless user asks for more.
3) When the user says **"save"**, include `--out web/last_search.json`.
4) When user says **"summarize last search"**, read `web/last_search.json`.
5) If `web/last_search.json` does not exist, run a search with `--out` first.

## Failure handling
- If exec returns an error about SearXNG not reachable, ask user to confirm SearXNG is running and port.
- If JSON is missing/empty, rerun with `--out web/last_search.json` and then read it.

## Tool Policy
- Never use built-in web tools (search/fetch). They are disabled.
- Use exec with the workspace SearXNG CLI: `./bin/search "<query>" --top <N> [--out web/last_search.json]`.
- Prefer snippets; do not fetch pages unless the user explicitly asks.
- Fallback: if `./bin/search` is unavailable, use this skill’s bundled script in `scripts/bin/search`.
