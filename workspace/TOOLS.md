# Tools

## Local search (preferred)
Use `exec` to run:

- `./bin/search "<query>"`

Return a concise JSON list of results: `{title, url, snippet}`.

### Rules
- Use this instead of fetching whole pages.
- Built-in web search/fetch tools are treated as **disabled**.
- Do not open URLs unless the user explicitly asks; confirm first.

### Search + save to Obsidian
- `./bin/search "<query>"`  
  Writes results JSON to `web/last_search.json`.
- `./bin/search_save_note "<query>" [--top N]`  
  Runs `./bin/search`, then writes a markdown note to `vaults/Search/YYYY-MM-DD/HHMM - <query>.md` and updates `vaults/Search/Search Index.md`.

## Memory system (personal context & learning)

### Core commands
- `./bin/memory_write --role <user|assistant|system> --content "..."`  
  Guaranteed storage of conversation turns and important events.
  
- `./bin/memory_recall --query "topic" [--format markdown|json]`  
  Retrieve relevant context (pinned memories + search results + recent context).
  
- `./bin/memory_sync`  
  Process new events, extract memories, create daily backup.
  
- `./bin/memory_status [--verify]`  
  Show system status and verify chain integrity.

### Advanced commands
- `./bin/memory_materialize [--all] [--no-embed]`  
  Process events into searchable database.
  
- `./bin/memory_remember [--all]`  
  Extract memories using pattern detection.
  
- `./bin/memory_export --output <file.tar.gz>`  
  Backup entire memory system.
  
- `./bin/memory_import --input <file.tar.gz>`  
  Restore from backup.

### When to use memory
- **Before answering**: Recall context to personalize response
- **After conversation**: Write turns for future reference
- **For preferences**: Automatically detects "I prefer...", "always...", etc.
- **For learning**: System tracks what memories are useful

### Memory workflow
1. Start: `./bin/memory_recall --query "current topic"` → inject into prompt
2. During: Collect conversation turns
3. End: `./bin/memory_write` for each turn → `./bin/memory_sync`

See `memory/MEMORY.md` and `memory/INTEGRATION.md` for complete docs.

