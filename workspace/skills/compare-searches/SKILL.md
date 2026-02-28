---
name: compare-searches
description: Run multiple searches via ./bin/search and compare result overlap, relevance, and sources.
---

# SKILL: compare_searches

## Use when the user says:
- "compare what we found with X"
- "do another search and compare"
- "what's the overlap?"

## Goal
Compare a new search against the last saved search.

## Inputs
- Query (required)
- Top (optional, default 8)

## Files
- Reads: `web/last_search.json`
- Writes: `web/compare_search.json` (optional, if you save)



## Tool Policy
- Never use built-in web tools (search/fetch). They are disabled.
- Use exec with the workspace SearXNG CLI: `./bin/search "<query>" --top <N> [--out web/last_search.json]`.
- Prefer snippets; do not fetch pages unless the user explicitly asks.
- Fallback: if `./bin/search` is unavailable, use this skill’s bundled script in `scripts/bin/search`.

## Procedure
1) Ensure `web/last_search.json` exists; if not, run `web_search_save` first.
2) Run new search and save it:
   exec: `./bin/search "<Query>" --top <Top> --out web/compare_search.json`
3) Read both JSON files.
4) Compare:
   - overlap in domains
   - unique best links
   - differences in angles
5) Provide a combined recommendation.

## Output format
**Overlap**
- ...

**Best from previous**
- ...

**Best from new**
- ...

**Combined takeaways**
- ...

**Recommended next actions**
- ...
