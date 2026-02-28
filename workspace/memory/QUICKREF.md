# PicoClaw Memory System - Quick Reference

## 🚀 Quick Start (5 commands)

```bash
# 1. Store a conversation turn
./bin/memory_write --role user --content "I prefer morning briefings at 7am"

# 2. Process it
./bin/memory_sync

# 3. Recall it later
./bin/memory_recall --query "briefing preferences"

# 4. Check status
./bin/memory_status

# 5. Backup
./bin/memory_export --output backup.tar.gz
```

## 📝 Common Patterns

### Store user preference
```bash
./bin/memory_write --role user --content "I prefer X" --type message
./bin/memory_sync
```

### Store assistant response
```bash
./bin/memory_write --role assistant --content "$RESPONSE" \
  --metadata '{"skill":"skill-name"}'
```

### Store workflow execution
```bash
./bin/memory_write --role system --content "Executed workflow: summary" \
  --type workflow --metadata '{"skill":"name","status":"success"}'
```

### Get context before LLM call
```bash
CONTEXT=$(./bin/memory_recall --query "relevant topic keywords" --format markdown)
# Then inject $CONTEXT into prompt
```

## 🔥 New: Hierarchical Memory & Distillation

### Memory Tiers

Memory is now organized in 3 tiers:

| Tier | Name | Retention | Speed | Use Case |
|------|------|-----------|-------|----------|
| 1 | Active | 7 days | <10ms | Recent & frequent |
| 2 | Working | 30 days | <100ms | Moderate access |
| 3 | Archive | Forever | <500ms | Long-term storage |

### Automatic Summaries

Daily, weekly, and monthly summaries are created automatically:

```bash
# Manual distillation
./bin/memory_distill daily --yesterday
./bin/memory_distill weekly
./bin/memory_distill monthly

# Automatic (runs all needed distillations)
./bin/memory_distill auto
```

### Tier Management

Optimize memory tiers based on access patterns:

```bash
# Promote recent/frequent to Tier 1
./bin/memory_tier promote

# Demote old items to Tier 2
./bin/memory_tier demote

# Archive items to Tier 3
./bin/memory_tier archive

# Run all optimizations
./bin/memory_tier auto

# View tier statistics
./bin/memory_tier stats
```

### Tier-Based Recall

Use hierarchical recall for better performance:

```bash
# Use tier-based recall (includes summaries)
./bin/memory_recall --query "topic" --use-tiers --format markdown
```

### System Dashboard

View comprehensive memory statistics:

```bash
./bin/memory_dashboard
```

### Archive Old Events

Compress old events to save space:

```bash
# Compress specific month
./bin/memory_archive compress --month 2026-01

# List archives
./bin/memory_archive list

# Search archive
./bin/memory_archive search --month 2026-01 --query "preferences"

# Auto-archive old months
./bin/memory_archive compress --auto
```

## 🎯 Skill Integration (3-Step Pattern)

```bash
# 1️⃣ BEFORE: Recall context
MEMORY=$(./bin/memory_recall --query "skill-specific topic" --format markdown)

# 2️⃣ DURING: Use context in your logic
# [Your skill logic here, informed by $MEMORY]

# 3️⃣ AFTER: Store results
./bin/memory_write --role assistant --content "$OUTPUT" --metadata '{"skill":"name"}'
./bin/memory_sync &
```

## 🔍 Pattern Detection (Automatic)

The system automatically detects and extracts:

| Pattern | Example | Priority |
|---------|---------|----------|
| Explicit | `/remember X`, `remember: X` | 10 (highest) |
| Preference | "I prefer X", "from now on X" | 5 |
| Decision | "we will X", "the plan is X" | 5 |
| Procedure | "always X", "every time X" | 5 |
| Fact | "note: X", "important: X" | 3 |

## 📊 Status & Maintenance

```bash
# Show status
./bin/memory_status

# Verify integrity
./bin/memory_status --verify

# Full sync (process all events)
./bin/memory_materialize --all
./bin/memory_remember --all

# Export backup
./bin/memory_export --output "backup_$(date +%Y%m%d).tar.gz"

# Import/restore
./bin/memory_import --input backup.tar.gz
```

## 🗂️ File Structure

```
memory/
├── log/
│   ├── events.ndjson          # Source of truth (append-only)
│   └── archive/               # Compressed archives (Tier 3)
├── index/
│   ├── memory.db              # Fast retrieval (SQLite with tiers)
│   ├── status.json            # Processing metadata
│   └── active_context.json    # Tier 1 cache
├── distilled/                 # Compressed summaries
│   ├── daily/                 # Daily summaries
│   ├── weekly/                # Weekly summaries
│   └── monthly/               # Monthly summaries
├── backups/                   # Daily backups (auto-created)
└── tests/                     # Test suites and benchmarks
```

## 💡 Tips & Best Practices

### ✅ DO:
- Store preferences explicitly: "remember: I prefer X"
- Use specific recall queries: "calendar morning routine" not just "calendar"
- Run `memory_sync` after important captures
- Check `memory_status` periodically
- Export backups daily (automated in `memory_sync`)

### ❌ DON'T:
- Store passwords/secrets (use `.secrets/` instead)
- Write every tiny interaction (be selective)
- Use vague queries ("help", "info")
- Skip the sync step after important writes

## 🔧 Troubleshooting

**No memories recalled?**
```bash
./bin/memory_status           # Check if events exist
./bin/memory_materialize      # Reprocess events
./bin/memory_remember         # Extract memories
```

**Chain integrity error?**
```bash
./bin/memory_status --verify  # Diagnose
```

**Want to see what's stored?**
```bash
# View raw events
tail -20 memory/log/events.ndjson

# View extracted memories
./bin/memory_status | grep -A 10 "Top Memories"
```

## 📚 Complete Documentation

- `memory/MEMORY.md` - Complete system documentation
- `memory/INTEGRATION.md` - Full integration examples
- `memory/IMPLEMENTATION.md` - Technical details
- `skills/memory-integration/SKILL.md` - Skill integration guide

## ⚡ Performance

- **RAM**: 50-75 MB (includes tier cache and active context)
- **Storage**: ~800 bytes/event, ~2-5 KB/indexed document
- **Compression**: 90-95% for archives, 80%+ for summaries
- **Speed**: 
  - Write: <10ms
  - Tier 1 recall: <10ms (active context)
  - Tier 2 recall: <100ms (working memory)
  - Tier 3 recall: <500ms (archive decompression)

### Run Benchmarks

Test your system's performance:

```bash
cd memory/tests
python3 benchmark.py
```

## 🎓 Learning Over Time

The system learns by:
1. Tracking which memories are retrieved and used
2. Counting usage per memory (`use_count`)
3. Recording positive/negative feedback
4. Re-ranking results based on historical usefulness

No configuration needed - it learns automatically from usage patterns.

---

**Need help?** See `memory/MEMORY.md` for full documentation or run `./bin/memory_status` to check system health.
