---
name: extract-links-last-search
description: Extract and deduplicate URLs from the most recent search output saved in web/search_results.json.
---

# SKILL: extract_links_last_search

## Goal
Create a clean link list (copy/paste friendly) from the saved JSON.

## Files
- Reads: `web/last_search.json`
- Writes (optional): `web/last_links.md`

## Procedure
1) Read `web/last_search.json`.
2) Output top N links (default 10) as:
   - [Title](URL) — one-line note
3) If asked, save to `web/last_links.md`.

## Output format
**Links**
- Title — URL
- Title — URL
...
