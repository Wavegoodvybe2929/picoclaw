---
name: list-week
description: List this week’s calendar events from calendar/events.csv grouped by day.
---

# SKILL: list_week

## Goal
Show a week view (Mon–Sun) with grouped events and key highlights.

## When to use
User asks for 'this week', 'week view', 'upcoming'

## Inputs (from user)
- Optional: week starting date (default: current week)

## Files used
- Read: `calendar/events.csv`
- Write: None

## Procedure
1) Read CSV.
2) Determine week window.
3) Group events by date.
4) Provide daily total hours (estimate) and top 1 highlight per day.

## Output format (must follow exactly)
**Week view (YYYY-MM-DD → YYYY-MM-DD)**

**Mon YYYY-MM-DD**
- …

**Tue …**
- …

**Highlights**
- …

## Safety / constraints
- Never invent calendar events.
- Default timezone: America/New_York
- All times must be explicit and in `YYYY-MM-DD HH:MM`.
- If required fields are missing, ask ONLY for what is missing.
- If you wrote files, state what changed.
