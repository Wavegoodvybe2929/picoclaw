---
name: cancel-event
description: Cancel an existing calendar event in calendar/events.csv by matching date/title/time and setting status to cancelled.
---

# SKILL: cancel_event

## Goal
Cancel an event without deleting it (set status=canceled).

## When to use
User says 'cancel', 'remove from calendar'

## Inputs (from user)
- Event ID OR title + date

## Files used
- Read: `calendar/events.csv`
- Write: `calendar/events.csv`

## Procedure
1) Load CSV.
2) Match event.
3) Set `status` to `canceled`.
4) Write CSV.
5) Confirm cancellation.

## Output format (must follow exactly)
**Canceled event**
- ID: …
- Title: …
- When: …

**Saved to**: calendar/events.csv

## Safety / constraints
- Never invent calendar events.
- Default timezone: America/New_York
- All times must be explicit and in `YYYY-MM-DD HH:MM`.
- If required fields are missing, ask ONLY for what is missing.
- If you wrote files, state what changed.
