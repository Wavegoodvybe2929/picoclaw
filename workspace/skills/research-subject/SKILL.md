---
name: research-subject
description: Research pipeline: search for links, scrape pages, and write a structured vault research note in vaults/Research/.
---

# SKILL: research_subject

## Goal
Given a subject/topic string, produce a single high-quality research note in the Obsidian vault.

## Tool policy
- Never use built-in web search/fetch tools.
- Always use `exec` with the workspace scripts in `./bin/`.

## Inputs
- Topic (required): a short subject string.

## Outputs
- Main research note: `vaults/Research/YYYY-MM-DD/HHMM - <topic>.md`
- Optional sources cache JSON (recommended): keep next to the note (or in workspace `web/`).

## Procedure

### A) Discover links
- exec: `./bin/research_links "<topic>" --top 25 --out web/research_links.json`

### B) Scrape links
- exec: `./bin/research_scrape --in web/research_links.json --out web/research_scraped.json --max-pages 15`

### C) Write note skeleton
- exec: `./bin/research_write_note "<topic>" --in web/research_scraped.json`
  (captures excerpts + sources list)

### D) Synthesize (PicoClaw)
Open the generated note and replace the placeholder bullets with:
- Executive summary (5–10 bullets)
- Key findings grouped into sections
- Practical takeaways / next steps
- Open questions

Keep the Sources list intact.

## Definition of done
When the user says: `Research <topic>`
- A new note appears under `vaults/Research/...` with:
  - summary + findings + takeaways + sources
- The pipeline works repeatedly without looping fetches.
