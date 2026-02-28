---
name: web-search
description: Search the web via ./bin/search (SearXNG) and return a concise, sourced summary (no web_fetch).
---

# SKILL: web_search

## Say this skill name when you want a simple web search:
"Use SKILL: web_search. Query: <text>"

## Goal
Run a local web search for the query and return the best results + a short summary.

## Inputs
- Query (required)
- Top (optional, default 5)



## Tool Policy
- Never use built-in web tools (search/fetch). They are disabled.
- Use exec with the workspace SearXNG CLI: `./bin/search "<query>" --top <N> [--out web/last_search.json]`.
- Prefer snippets; do not fetch pages unless the user explicitly asks.
- Fallback: if `./bin/search` is unavailable, use this skill’s bundled script in `scripts/bin/search`.

## Procedure
1) Run:
   exec: `./bin/search "<Query>" --top <Top>`
2) From the output, pick the best 5 results (or fewer if fewer returned).
3) Summarize into 5–10 bullets.
4) Provide 3 next actions if appropriate.

## Output format
**Top results**
1) Title — URL
- snippet

**Summary**
- ...

**Next actions**
- ...
