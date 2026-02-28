---
name: web-search-save
description: Search via ./bin/search, save raw JSON to web/last_search.json, and (optionally) save a markdown note into vaults/Search/.
---

# SKILL: web_search_save

## Say this skill name when you want search + save:
"Use SKILL: web_search_save. Query: <text>"

## Goal
Run a local web search and save raw JSON to `web/last_search.json`, then summarize.

## Inputs
- Query (required)
- Top (optional, default 8)

## Files
- Writes: `web/last_search.json`
- (Optional) Writes: `vaults/Search/YYYY-MM-DD/HHMM - <query>.md`
- (Optional) Appends: `vaults/Search/Search Index.md`



## Tool Policy
- Never use built-in web tools (search/fetch). They are disabled.
- Use exec with the workspace SearXNG CLI: `./bin/search "<query>" --top <N> [--out web/last_search.json]`.
- Prefer snippets; do not fetch pages unless the user explicitly asks.
- Fallback: if `./bin/search` is unavailable, use this skill’s bundled script in `scripts/bin/search`.

## Procedure
1) Ensure folder exists (create if needed):
   `web/`
2) Run (standard, JSON only):
   exec: `./bin/search "<Query>" --top <Top> --out web/last_search.json`
3) If the user wants searches logged into the Obsidian vault, prefer the wrapper:
   exec: `./bin/search_save_note "<Query>" --top <Top>`
   (This still writes `web/last_search.json` and also creates a vault note + updates Search Index.)
4) Summarize the printed results.
5) If user asks for deeper extraction, open and read `web/last_search.json`.

## Output format
Same as `web_search`.
