---
name: summarize-last-search
description: Summarize the most recent saved web search results into key takeaways and next actions.
---

# SKILL: summarize_last_search

## Use when the user says things like:
- "summarize the last web search"
- "what did we find?"
- "give me takeaways from last search"

## Goal
Summarize the most recently saved SearXNG JSON results.

## Preconditions
- `web/last_search.json` exists (if not, instruct to run `web_search_save` first)

## Files
- Reads: `web/last_search.json`

## Procedure
1) Read `web/last_search.json`.
2) Extract:
   - query string
   - top results (title/url/content)
3) Provide:
   - 5–10 takeaways
   - best 5 links with why they matter
   - 3 next actions

## Output format
**Query**
- ...

**Best links**
1) Title — URL
- Why it matters: ...

**Takeaways**
- ...

**Next actions**
- ...
