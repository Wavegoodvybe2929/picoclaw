---
name: daily-briefing
description: Generate a daily briefing from local workspace data (calendar, tasks, notes) without using web tools.
---

# SKILL: daily_briefing

## Goal
Produce a morning briefing: agenda, priorities, reminders, conflicts, and one optional win.

## When to use
User asks for plan/day/agenda/morning briefing, or at the start of a session.

## Inputs (from user)
- Optional: target date (default: today)
- Optional: constraints (travel buffer, focus blocks, priorities)

## Files used
- Read: `calendar/events.csv` (required), optionally `notes/inbox.md`, `tasks/todo.md` if present
- Write: Optional: `notes/daily/YYYY-MM-DD.md` (if user wants it saved)

## Procedure
1) **Check memory for preferences**: `./bin/memory_recall --query "daily briefing preferences calendar priorities"` to get user's preferences and context.
2) Read `calendar/events.csv`.
3) Filter events for target date where `status != canceled`.
4) Sort by start time.
5) Identify gaps >= 30 minutes.
6) Detect overlaps/conflicts.
7) Generate top 3 actionable priorities (derived from notes + event context + memory context; if none, ask what 1–3 outcomes matter today).
8) Provide reminders for any "prep" implied by notes.
9) Offer one small "optional win" task.
10) **Store interaction**: After generating briefing, optionally write to memory with `./bin/memory_write --role assistant --content "Generated daily briefing for [date]" --metadata '{"skill":"daily-briefing"}'`

## Output format (must follow exactly)
1) **Today’s agenda**
- HH:MM–HH:MM Title (Location)

2) **Top 3 priorities**
- …

3) **Reminders / prep**
- …

4) **Conflicts / gaps**
- …

5) **Optional win**
- …

## Safety / constraints
- Never invent calendar events.
- Default timezone: America/New_York
- All times must be explicit and in `YYYY-MM-DD HH:MM`.
- If required fields are missing, ask ONLY for what is missing.
- If you wrote files, state what changed.
