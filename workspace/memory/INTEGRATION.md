# PicoClaw Memory Integration Guide

This guide shows how to integrate the memory system into PicoClaw workflows.

## Core Principle

**Write first, process async, recall on-demand**

- `memory_write`: Called immediately for every turn (guaranteed storage)
- `memory_materialize` + `memory_remember`: Run in background or after conversation
- `memory_recall`: Called when building LLM prompts (retrieve context)

## Integration Points

### 1. On Every Conversation Turn

**Store user input:**
```bash
# Option A: Command-line args
./bin/memory_write --role user --content "$USER_INPUT" \
  --conversation-id "$CONV_ID" --metadata '{"source":"chat"}'

# Option B: JSON pipe
echo "{\"role\":\"user\",\"content\":\"$USER_INPUT\"}" | ./bin/memory_write --json
```

**Store assistant response:**
```bash
./bin/memory_write --role assistant --content "$ASSISTANT_RESPONSE" \
  --conversation-id "$CONV_ID"
```

### 2. After Workflow/Skill Execution

**Store workflow output:**
```bash
./bin/memory_write --role system --content "$WORKFLOW_RESULT" \
  --type workflow --metadata "{\"skill\":\"$SKILL_NAME\"}"
```

**Process new events (async):**
```bash
# Run these in background or on schedule
./bin/memory_materialize --quiet &
./bin/memory_remember --quiet &
```

### 3. Before LLM Call (Get Context)

**Retrieve relevant memories:**
```bash
# Get context as markdown (for prompt injection)
MEMORY_CONTEXT=$(./bin/memory_recall --query "$TOPIC_OR_QUESTION" --format markdown)

# Or as JSON (for programmatic use)
MEMORY_JSON=$(./bin/memory_recall --query "$TOPIC" --format json --budget-tokens 1500)
```

**Inject into prompt:**
```bash
FULL_PROMPT="
$MEMORY_CONTEXT

---

User: $USER_INPUT
"
```

## Example: Daily Briefing Skill

```bash
#!/bin/bash
# skills/daily-briefing/run.sh

# 1. Store that user requested briefing
echo '{"role":"user","content":"Give me daily briefing"}' | \
  ../../bin/memory_write --json

# 2. Get relevant context from memory
MEMORY=$(../../bin/memory_recall --query "daily briefing preferences calendar" \
  --format markdown)

# 3. Build prompt with memory context
PROMPT="$MEMORY

---

Generate a daily briefing for February 19, 2026.
Include calendar, tasks, and priorities.
"

# 4. Call LLM (your existing method)
RESPONSE=$(call_llm "$PROMPT")

# 5. Store assistant response
echo "{\"role\":\"assistant\",\"content\":\"$RESPONSE\"}" | \
  ../../bin/memory_write --json

# 6. Process new events (background)
../../bin/memory_materialize --quiet &
../../bin/memory_remember --quiet &

# 7. Return response
echo "$RESPONSE"
```

## Example: Capture Note Skill

```bash
#!/bin/bash
# skills/capture-note/run.sh

NOTE="$1"

# 1. Store with explicit remember directive
../../bin/memory_write --role user \
  --content "remember: $NOTE" \
  --type note

# 2. Save to vault (your existing method)
../../bin/vault_new_note --title "Quick Note" --content "$NOTE"

# 3. Process immediately (this is high-priority)
../../bin/memory_materialize --quiet
../../bin/memory_remember --quiet

echo "Note captured and remembered."
```

## Example: Email Triage Skill

```bash
#!/bin/bash
# skills/email-check-triage-calendar/run.sh

# 1. Get emails
EMAILS=$(../../bin/check_gmail_unread)

# 2. Store that we checked email
../../bin/memory_write --role system \
  --content "Checked email: $EMAIL_COUNT new messages" \
  --type workflow --metadata '{"skill":"email-check-triage"}'

# 3. Get memory context about email preferences
MEMORY=$(../../bin/memory_recall --query "email preferences important contacts" \
  --format markdown)

# 4. Build triage prompt
PROMPT="$MEMORY

---

Triage these emails:
$EMAILS
"

# 5. Call LLM for triage
TRIAGE=$(call_llm "$PROMPT")

# 6. Store triage result
../../bin/memory_write --role assistant \
  --content "Email triage: $TRIAGE" \
  --metadata '{"skill":"email-check-triage"}'

# 7. Process async
../../bin/memory_materialize --quiet &
../../bin/memory_remember --quiet &

echo "$TRIAGE"
```

## Periodic Maintenance

Run these periodically (cron, launchd, or manual):

```bash
# Daily: Export backup
./bin/memory_export --output "backups/memory_$(date +%Y%m%d).tar.gz"

# Weekly: Verify integrity
./bin/memory_status --verify

# After bulk operations: Reprocess all
./bin/memory_materialize --all
./bin/memory_remember --all
```

## Best Practices

### 1. Always write first
Never skip `memory_write` - it's your durability guarantee.

### 2. Process async when possible
`memory_materialize` and `memory_remember` can run in background.

### 3. Query specifically
Better queries = better recall. Include context words:
- Good: "calendar preferences morning routine"
- Weak: "preferences"

### 4. Use explicit directives for important info
Prefix important statements with `remember:` or `/remember` to ensure high priority.

### 5. Track conversation IDs
Use `--conversation-id` to group related turns for better context.

### 6. Check status regularly
Run `./bin/memory_status` to ensure processing is keeping up.

## Troubleshooting

**Memory not recalled?**
```bash
# Check if it was written
grep "preference" memory/log/events.ndjson

# Check if it was processed
./bin/memory_status

# Reprocess if needed
./bin/memory_materialize
./bin/memory_remember
```

**Too much or too little recalled?**
- Adjust `--budget-tokens` to control volume
- Use more specific queries
- Pin important memories by marking explicit

**Processing lag?**
```bash
# Check status
./bin/memory_status

# Catch up processing
./bin/memory_materialize
./bin/memory_remember
```

## Environment Variables (Optional)

You can set these to customize behavior:

```bash
export MEMORY_DIR="/custom/path"          # Override memory directory
export MEMORY_MAX_CHUNK_SIZE=500          # Chunk size for indexing
export MEMORY_DEFAULT_BUDGET_TOKENS=2000  # Default recall budget
```

## API-Style Usage (Python)

If you're calling from Python directly:

```python
import sys
sys.path.insert(0, 'memory')
from memory_core import write_event, get_db_connection, build_memory_packet

# Write event
result = write_event(
    role="user",
    content="I prefer dark mode",
    event_type="message"
)

# Recall memories
from bin.memory_recall import build_memory_packet
packet = build_memory_packet(
    query="ui preferences",
    budget_tokens=1500
)
```

## Integration Checklist

- [ ] Add `memory_write` calls for every user turn
- [ ] Add `memory_write` calls for every assistant response
- [ ] Add `memory_recall` calls before LLM prompts
- [ ] Set up async processing (materialize + remember)
- [ ] Add explicit `/remember` support in user input handler
- [ ] Set up periodic backups (memory_export)
- [ ] Add status monitoring (memory_status)
- [ ] Update skills to include memory context
- [ ] Test recall quality with real queries
- [ ] Document any memory-related preferences in prompts

## Next Steps

1. Start with one skill (e.g., daily-briefing)
2. Add memory integration step-by-step
3. Test recall quality
4. Expand to other skills
5. Monitor and refine

The memory system is designed to be incrementally adopted - start small and expand as you see value.