---
name: memory-integration
description: Guide for integrating the memory system into any PicoClaw skill
---

# SKILL: Memory Integration Guide

## Goal
Show how to integrate PicoClaw's memory system into any skill for context awareness and learning.

## When to use memory

### Always use for:
- User preferences ("I prefer...", "always do...", "never...")
- Repeated workflows (daily briefings, email triage, etc.)
- Learning from feedback (what worked/didn't work)
- Personalizing responses based on past interactions

### Optional for:
- One-off tasks that don't need context
- Simple queries without personalization needs

## Integration Pattern (3 steps)

### Step 1: Recall Context (Before Main Logic)
```bash
# Get relevant memories for this task (enhanced with tiers)
MEMORY_CONTEXT=$(./bin/memory_recall \
  --query "skill-specific keywords and context" \
  --format markdown \
  --budget-tokens 1500 \
  --use-tiers)

# Use $MEMORY_CONTEXT in your prompts/logic
```

**Note**: The `--use-tiers` flag enables hierarchical memory recall, which:
- Includes active context (Tier 1) - recent and frequently accessed items
- Searches working memory (Tier 2) - recent events with semantic search
- Includes distilled summaries automatically
- Falls back to archive (Tier 3) if needed
- Provides better performance and more relevant context

**Example queries by skill**:
- Daily briefing: "daily briefing preferences calendar priorities morning routine"
- Email triage: "email preferences important contacts scheduling habits"
- Research: "research topics interests previous research on [topic]"
- Calendar: "calendar preferences meeting times scheduling habits"

### Step 2: Execute Main Skill Logic
Your normal skill procedure runs here, informed by $MEMORY_CONTEXT.

### Step 3: Store Results (After Main Logic)
```bash
# Store user input (if applicable)
./bin/memory_write --role user \
  --content "$USER_INPUT" \
  --type message \
  --metadata '{"skill":"skill-name"}'

# Store assistant response/output
./bin/memory_write --role assistant \
  --content "$SKILL_OUTPUT" \
  --type message \
  --metadata '{"skill":"skill-name","success":true}'

# Store workflow execution summary
./bin/memory_write --role system \
  --content "Executed [skill-name]: [brief summary]" \
  --type workflow \
  --metadata '{"skill":"skill-name"}'

# Process memories (can run in background)
./bin/memory_sync &
```

## Pattern Detection (Automatic)

The memory system automatically extracts these patterns:

### Explicit directives
- `/remember ...`
- `remember: ...`
- "I will remember..."
- User says: "Save this preference"

### Preferences (detected automatically)
- "I prefer..."
- "from now on..."
- "always..." / "never..."
- "my preference is..."

### Decisions
- "we will..." / "I will..."
- "the plan is..."
- "decided to..."

### Procedures
- "do this every time..."
- "whenever X, do Y..."
- "the process is..."

## Complete Example: Daily Briefing with Memory

```skill
## Procedure
1) **Recall memory context**:
   - exec: `./bin/memory_recall --query "daily briefing preferences priorities" --format markdown`
   - Store result in $MEMORY_CONTEXT

2) **Read calendar and tasks**:
   - exec: read `calendar/events.csv`
   - exec: read `tasks/todo.md`

3) **Generate briefing**:
   - Use $MEMORY_CONTEXT to personalize (preferred time, priority areas, etc.)
   - Generate formatted briefing

4) **Store interaction**:
   - exec: `./bin/memory_write --role assistant --content "$BRIEFING" --metadata '{"skill":"daily-briefing"}'`
   - exec: `./bin/memory_sync &`

5) **Return briefing** to user
```

## Complete Example: Capture Note with Memory

```skill
## Procedure
1) **Check if note is important**:
   - If note starts with "remember:", "I prefer", etc.

2) **Store in memory immediately**:
   - exec: `./bin/memory_write --role user --content "remember: $NOTE" --type note`

3) **Also store in vault**:
   - exec: append to `vaults/Inbox/notes.md`

4) **Process memory**:
   - exec: `./bin/memory_sync`
   - This extracts the preference/fact for future recall

5) **Confirm to user**: "Saved and remembered: [note]"
```

## Memory Commands Quick Reference

### Core workflow
```bash
# 1. Recall before action
./bin/memory_recall --query "topic" --format markdown

# 2. Write after action
./bin/memory_write --role <user|assistant|system> --content "..."

# 3. Sync periodically
./bin/memory_sync
```

### Helpful flags
```bash
# Write with type and metadata
./bin/memory_write --role user --content "..." \
  --type message \
  --metadata '{"skill":"name","tags":["tag1"]}'

# Recall with token budget
./bin/memory_recall --query "..." --budget-tokens 1000

# Get JSON instead of markdown
./bin/memory_recall --query "..." --format json

# Quiet sync (no output)
./bin/memory_sync --quiet
```

## Best Practices

### DO:
✓ Store user preferences explicitly: `remember: I prefer X`
✓ Use specific queries: "calendar preferences morning routine" vs "preferences"
✓ Include skill name in metadata for tracking
✓ Run `memory_sync` after important captures
✓ Use memory context to personalize responses

### DON'T:
✗ Store sensitive passwords/tokens in memory (use `.secrets/` instead)
✗ Write every tiny interaction (be selective)
✗ Forget to call `memory_sync` after capturing important info
✗ Use vague queries like "help" (be specific)

## Integration Checklist

When adding memory to a skill:

- [ ] Add memory recall at start (Step 1)
- [ ] Use memory context in logic/prompts (Step 2)
- [ ] Store user input if meaningful (Step 3)
- [ ] Store assistant output (Step 3)
- [ ] Run memory_sync at end (Step 3)
- [ ] Update skill metadata to indicate memory-aware
- [ ] Test that memories are recalled correctly
- [ ] Document what memories the skill uses/creates

## Testing Memory Integration

After integrating:

1. **Test write**:
   ```bash
   # Run skill, then check:
   ./bin/memory_status
   # Should show increased event count
   ```

2. **Test recall**:
   ```bash
   # Next time, should get context:
   ./bin/memory_recall --query "[skill topic]"
   # Should show relevant memories
   ```

3. **Test extraction**:
   ```bash
   # If you captured a preference:
   ./bin/memory_status --json | grep "memories"
   # Should show extracted memory
   ```

## Memory Maintenance (Optional)

For skills that run frequently, consider adding periodic maintenance:

```bash
#!/bin/bash
# Example: Daily briefing with memory maintenance

# 1. Get tiered context (automatically includes summaries)
MEMORY=$(../../bin/memory_recall --query "briefing preferences calendar priorities" \
  --format markdown --budget-tokens 1500 --use-tiers)

# 2. Use context in prompt
PROMPT="$MEMORY

---

Generate today's briefing based on my preferences and recent context.
Focus on priorities mentioned in memory.
"

# 3. Call LLM
RESPONSE=$(call_llm "$PROMPT")

# 4. Store interaction
echo "{\"role\":\"assistant\",\"content\":\"$RESPONSE\"}" | \
  ../../bin/memory_write --json

# 5. Sync memory (background)
../../bin/memory_sync &
```

### Automatic Maintenance Tasks

Run periodically (e.g., daily at 2am) for optimal performance:

```bash
# Tier management (promote/demote based on access)
../../bin/memory_tier auto

# Distillation (create summaries)
../../bin/memory_distill auto

# Archive (compress old months)
../../bin/memory_archive compress --auto
```

These tasks:
- **Tier management**: Keeps hot storage optimized for recent/frequent items
- **Distillation**: Creates daily/weekly/monthly summaries for compressed knowledge
- **Archive**: Compresses old events to save space (95%+ compression)

## See Also
- `memory/MEMORY.md` - Complete memory system docs
- `memory/INTEGRATION.md` - Full integration examples
- `memory/README.md` - Quick start guide

## Philosophy

Memory makes PicoClaw **learn and personalize** without manual configuration:
- User says "I prefer X" → System remembers → Future responses use X
- Workflow runs repeatedly → System learns patterns → Suggests improvements
- User gives feedback → System adjusts → Better results over time

This is the difference between a tool and an assistant.
