---
name: capture-note
description: Capture a quick note into notes/inbox.md (or the configured notes inbox) with timestamp and tags.
---

# SKILL: capture_note

## Goal
Capture a quick note into an inbox file for later review.

## When to use
User says 'note', 'remember this', 'capture'

## Inputs (from user)
- Note text (required)
- Optional: tag (work/personal/finance/etc.)

## Files used
- Read: None (create file if missing)
- Write: `notes/inbox.md` (append)

## Procedure
1) **Store in memory first**: If note starts with important keywords ("remember:", "I prefer", "always"), immediately write to memory: `./bin/memory_write --role user --content "remember: [note text]" --type note`
2) Create `notes/` if needed.
3) Append a timestamped bullet to `notes/inbox.md`.
4) **Process memory**: Run `./bin/memory_sync` to process and extract the memory item.
5) Echo back what was saved.

## Output format (must follow exactly)
**Saved note**
- YYYY-MM-DD HH:MM ET — [tag] note text

**Saved to**: notes/inbox.md

## Safety / constraints
- Never invent calendar events.
- Default timezone: America/New_York
- All times must be explicit and in `YYYY-MM-DD HH:MM`.
- If required fields are missing, ask ONLY for what is missing.
- If you wrote files, state what changed.
