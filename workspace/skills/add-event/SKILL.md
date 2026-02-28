---
name: add-event
description: Add a new event to calendar/events.csv from user-provided title/date/start/end and optional details.
---

# SKILL: add_event

## Goal
Add a new event to `calendar/events.csv` safely and consistently.

## When to use
User says 'add', 'schedule', 'put on my calendar', 'remind me'

## Inputs (from user)
- Title (required)
- Date (required)
- Start time (required)
- End time (required)
- Location (optional)
- Notes (optional)
- Priority 1–3 (optional, default 2)
- Status (optional, default scheduled)

## Files used
- Read: `calendar/events.csv` (create if missing)
- Write: `calendar/events.csv` (append)

## Procedure
1) If any required field missing, ask only for missing.
2) Load CSV and append new row.
3) Re-read to confirm.
4) Echo the saved event details and where it was written.

## Output format (must follow exactly)
**Added event**
- Title: …
- When: YYYY-MM-DD HH:MM–HH:MM
- Location: …
- Notes: …

**Saved to**: calendar/events.csv

## Safety / constraints
- Never invent calendar events.
- Default timezone: America/New_York
- All times must be explicit and in `YYYY-MM-DD HH:MM`.
- If required fields are missing, ask ONLY for what is missing.
- If you wrote files, state what changed.
