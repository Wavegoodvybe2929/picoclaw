---
name: import-events-text
description: Parse a pasted text list of events and import them into calendar/events.csv with validation.
---

# SKILL: import_events_text

## Goal
Convert a pasted list of events into CSV rows.

## When to use
User pastes meeting list or says 'import these events'

## Inputs (from user)
- Free text list of events

## Files used
- Read: `calendar/events.csv` (optional)
- Write: `calendar/events.csv` (append)

## Procedure
1) Parse each line into title, date, start, end, timezone.
2) If any item ambiguous, ask clarifying questions grouped by event.
3) Append rows for clear items.
4) Summarize what was imported.

## Output format (must follow exactly)
**Imported events**
- evt-#### Title — YYYY-MM-DD HH:MM–HH:MM ET

**Questions (if needed)**
- …

**Saved to**: calendar/events.csv

## Safety / constraints
- Never invent calendar events.
- Default timezone: America/New_York
- All times must be explicit and in `YYYY-MM-DD HH:MM`.
- If required fields are missing, ask ONLY for what is missing.
- If you wrote files, state what changed.
