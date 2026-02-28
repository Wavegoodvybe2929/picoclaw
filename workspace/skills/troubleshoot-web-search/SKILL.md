---
name: troubleshoot-web-search
description: Troubleshoot local search issues with ./bin/search and SearXNG (localhost) without using built-in web tools.
---

# SKILL: troubleshoot_web_search

## Goal
Help fix the web-search system when errors happen.

## Checklist (ask/verify in this order)
1) Is SearXNG running?
   - user runs: `docker ps | grep searxng`
2) Is JSON allowed?
   - user runs: `curl -s "http://localhost:8080/search?q=test&format=json" | head -c 80; echo`
3) Does the CLI work?
   - user runs: `~/.picoclaw/workspace/skills/troubleshoot-web-search/scripts/bin/search "test" --top 3`
4) Port mismatch?
   - user runs: `SEARXNG_URL=http://localhost:PORT ~/.picoclaw/workspace/skills/troubleshoot-web-search/scripts/bin/search "test" --top 3`

## Common fixes
- Recreate container with mounted settings:
  Host: `/Users/wavegoodvybe/.searxng` -> Container: `/etc/searxng`
- Ensure `settings.yml` enables json:
  search.formats: [html, json]

## Tool Policy
- Never use built-in web tools (search/fetch). They are disabled.
- Use exec with the workspace SearXNG CLI: `./bin/search "<query>" --top <N> [--out web/last_search.json]`.
- Prefer snippets; do not fetch pages unless the user explicitly asks.
- Fallback: if `./bin/search` is unavailable, use this skill’s bundled script in `scripts/bin/search`.
