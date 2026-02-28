# PicoClaw Memory System - Quick Start

Tiny, reliable personal memory for PicoClaw. Runs on 24GB RAM alongside gpt-oss-20b.

## What It Does

- **Stores everything**: Append-only log with tamper-evident chain
- **Recalls reliably**: Pinned memories + smart search + deterministic fallback
- **Learns from usage**: Tracks what memories are helpful
- **Zero external dependencies**: Pure Python 3 standard library

## Quick Test

```bash
# 1. Write a memory
./bin/memory_write --role user --content "I prefer morning briefings at 7am"

# 2. Process it
./bin/memory_materialize
./bin/memory_remember

# 3. Recall it
./bin/memory_recall --query "briefing preferences"

# 4. Check status
./bin/memory_status
```

## Integration Pattern

```bash
# Before LLM call: get context
CONTEXT=$(./bin/memory_recall --query "daily planning")

# After conversation: store turns
echo "$USER_MSG" | ./bin/memory_write --role user --json
echo "$ASSISTANT_MSG" | ./bin/memory_write --role assistant --json

# Periodically: process new events
./bin/memory_materialize --quiet
./bin/memory_remember --quiet
```

## Files Created

```
memory/
├── log/events.ndjson       # Source of truth (append-only)
├── index/memory.db         # Fast retrieval (SQLite)
└── index/status.json       # Processing metadata
```

## Full Documentation

See [MEMORY.md](MEMORY.md) for complete details.

## Backup

```bash
# Export
./bin/memory_export --output backup.tar.gz

# Restore
./bin/memory_import --input backup.tar.gz
```