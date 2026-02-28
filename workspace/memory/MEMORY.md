# PicoClaw Memory System

**Low-RAM, reliable, plug-and-play personal memory for PicoClaw**

## Overview

The PicoClaw memory system provides guaranteed storage and reliable recall of conversation history, decisions, preferences, and learned knowledge. It's designed to run within tight RAM constraints (24GB alongside gpt-oss-20b) using on-disk storage with smart indexing.

## Architecture

### 1. Guaranteed Storage (Append-only Log)

**Location**: `memory/log/events.ndjson`

Every conversation turn, tool output, and significant event is written to an append-only log. Each event includes:
- Unique `event_id`
- Timestamp (UTC)
- Role (user/assistant/system)
- Content
- Metadata (conversation_id, thread_id, attachments)
- Tamper-evident chain (`hash`, `prev_hash`)

**Promise**: If `memory_write` succeeds, the event is durably stored and verifiable.

### 2. Lightweight Index (SQLite)

**Location**: `memory/index/memory.db`

Indexed for fast retrieval:
- **Documents table**: Chunked content with metadata
- **Memories table**: Pinned/procedural/semantic items extracted from events
- **Embeddings table**: Vector embeddings (optional, when sqlite-vec available)
- **Feedback table**: Learning signals (which memories were useful)
- **FTS5 index**: Keyword search fallback

### 3. Deterministic Recall (Memory Packet Contract)

When PicoClaw needs context, `memory_recall` builds a structured packet:
- **Pinned memories**: Always included (highest priority items)
- **Recent context**: Time-window slice (last N days)
- **Relevant hits**: Vector/keyword search results
- **Evidence pointers**: Event IDs, timestamps, sources (auditable)
- **Diagnostics**: Methods used, fallback status

**Promise**: The LLM always receives consistent, traceable, structured context.

## Command-Line Tools

All tools are in `bin/` and follow PicoClaw conventions.

### memory_write

Write an event to the append-only log (guaranteed storage).

```bash
# From command line
./bin/memory_write --role user --content "I prefer dark mode"

# From JSON
echo '{"role":"assistant","content":"Saved note"}' | ./bin/memory_write --json

# With metadata
./bin/memory_write --role user --content "Meeting at 3pm" --metadata '{"skill":"calendar"}'
```

**Always call this first** to ensure nothing is lost.

### memory_materialize

Process new events and update the database (chunking, embeddings, FTS).

```bash
# Process new events (idempotent)
./bin/memory_materialize

# Reprocess all events
./bin/memory_materialize --all

# Skip embeddings (faster, keyword-only)
./bin/memory_materialize --no-embed
```

**Run after writing events** to make them searchable.

### memory_remember

Extract memory items using explicit directives and deterministic rules.

```bash
# Process new events for memories
./bin/memory_remember

# Reprocess all
./bin/memory_remember --all
```

**Detects**:
- Explicit: `/remember ...`, `remember: ...`, "I will remember..."
- Preferences: "I prefer...", "from now on...", "always/never..."
- Decisions: "we will...", "the plan is...", "decided to..."
- Procedures: "do this every time...", "whenever X do Y..."
- Facts: "note:", "important:", "keep in mind..."

### memory_recall

Build and retrieve memory packet for LLM injection.

```bash
# Get relevant context (markdown format)
./bin/memory_recall --query "calendar preferences"

# JSON format with token budget
./bin/memory_recall --query "morning briefing" --budget-tokens 1500 --format json
```

**Retrieval strategy**:
1. Always include pinned memories
2. Try vector search first
3. Fallback to keyword search (FTS)
4. Fallback to recent time window
5. Always return evidence pointers

### memory_status

Show current system status.

```bash
# Human-readable status
./bin/memory_status

# JSON format
./bin/memory_status --json

# Verify chain integrity
./bin/memory_status --verify
```

### memory_export / memory_import

Backup and restore the entire memory system.

```bash
# Export to compressed archive
./bin/memory_export --output backup.tar.gz

# Import (with backup of existing data)
./bin/memory_import --input backup.tar.gz --backup

# Merge with existing data
./bin/memory_import --input backup.tar.gz --merge
```

## Integration with PicoClaw

### Typical workflow

1. **On conversation turn**:
   ```bash
   # Store user message
   ./bin/memory_write --role user --content "$USER_MESSAGE"
   
   # Store assistant response
   ./bin/memory_write --role assistant --content "$ASSISTANT_RESPONSE"
   ```

2. **After workflow execution**:
   ```bash
   # Materialize new events
   ./bin/memory_materialize --quiet
   
   # Extract memories
   ./bin/memory_remember --quiet
   ```

3. **Before LLM call (get context)**:
   ```bash
   # Retrieve relevant context
   CONTEXT=$(./bin/memory_recall --query "$TOPIC" --format markdown)
   
   # Inject into prompt
   ```

### Automated pipeline

For continuous operation, create a wrapper that:
- Writes events immediately (guaranteed storage)
- Materializes in background or on schedule
- Recalls on-demand when building LLM prompts

## What Gets Remembered

**Explicit (highest priority)**:
- Anything marked with `/remember` or `remember:`
- User statements like "I will remember..."
- Assistant confirmations like "saved to..."

**Deterministic patterns**:
- Preferences and settings
- Decisions and plans
- Procedures and workflows
- Important notes and facts
- File/vault operations

**Automatic**:
- All user messages (stored)
- All assistant responses (stored)
- Selective indexing (only meaningful content)

## Reliability Guarantees

✓ **Guaranteed**: Events are durably stored if `memory_write` succeeds  
✓ **Guaranteed**: Updates are not missed if writes are mandatory per turn  
✓ **Guaranteed**: LLM receives consistent memory packet format  
✓ **Guaranteed**: Evidence pointers make recall auditable  

✗ **Not guaranteed**: Perfect semantic recall via embeddings alone  
→ **Mitigated by**: Pinned memories + recency + keyword fallback

## Learning & Improvement

The system tracks usage to improve relevance:
- Records which memories were retrieved and used
- Tracks use counts and last-used timestamps
- Stores positive/negative feedback signals
- Re-ranks results based on historical usefulness

## Storage & RAM Footprint

**Disk**:
- Event log: ~1KB per event (grows linearly)
- SQLite DB: ~2-5KB per indexed document
- Embeddings: ~512 bytes per chunk (optional)

**RAM** (while running):
- SQLite: ~5-20MB working set (on-disk DB)
- Python process: ~10-30MB per tool invocation
- No persistent service required

**Total**: Minimal. Runs alongside gpt-oss-20b on 24GB easily.

## Advanced: Vector Search

To enable semantic vector search:

1. Install `sqlite-vec` extension:
   ```bash
   # macOS (Homebrew)
   brew install sqlite-vec
   
   # Or build from source and place in workspace/lib/
   ```

2. The system will auto-detect and use it

3. If unavailable, keyword search (FTS) is used automatically

## File Locations

```
memory/
├── MEMORY.md              # This file
├── memory_core.py         # Core Python module
├── log/
│   └── events.ndjson      # Append-only event log (source of truth)
└── index/
    ├── memory.db          # SQLite database (index + metadata)
    └── status.json        # Processing status tracking
```

## Troubleshooting

**No memories recalled?**
- Check `./bin/memory_status` to see if events were processed
- Run `./bin/memory_materialize` to index new events
- Run `./bin/memory_remember` to extract memory items

**Chain integrity error?**
- Run `./bin/memory_status --verify` to diagnose
- Check for manual edits to `events.ndjson`

**High RAM usage?**
- Embeddings are optional: use `--no-embed` flag
- FTS keyword search is always available as fallback

**Lost data?**
- Export regularly: `./bin/memory_export`
- Event log is append-only and includes hash chain for verification
