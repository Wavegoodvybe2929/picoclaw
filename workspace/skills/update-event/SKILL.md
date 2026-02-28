---
name: update-event
description: Update an existing calendar event in calendar/events.csv (time/title/location/notes/status) safely.
---

# SKILL: update_event

## Goal
Update an existing event (time/title/location/notes/status).

## When to use
User says 'move', 'reschedule', 'change', 'update'

## Inputs (from user)
- Event identifier (prefer ID) OR title + date
- Desired changes (fields)

## Files used
- Read: `calendar/events.csv`
- Write: `calendar/events.csv` (edit row in place)

## Procedure
1) Load CSV.
2) Match by title/date and show top 2–3 candidates if ambiguous.
3) Apply requested edits.
4) Write updated CSV.
5) Confirm old vs new fields.

## Output format (must follow exactly)
**Updated event**
- Before: YYYY-MM-DD HH:MM–HH:MM Title
- After:  YYYY-MM-DD HH:MM–HH:MM Title

**Saved to**: calendar/events.csv

## Safety / constraints
- Never invent calendar events.
- Default timezone: America/New_York
- All times must be explicit and in `YYYY-MM-DD HH:MM`.
- If required fields are missing, ask ONLY for what is missing.
- If you wrote files, state what changed.
