---
name: list-today
description: List today’s calendar events from calendar/events.csv in a clean agenda format.
---

# SKILL: list_today

## Goal
List all events for today in time order, with next-up indicator.

## When to use
User asks: 'what’s on my schedule', 'today’s meetings', 'what do I have today'

## Inputs (from user)
- Optional: date (default: today)

## Files used
- Read: `calendar/events.csv`
- Write: None

## Procedure
1) Read CSV.
2) Filter to date, status scheduled/tentative/done.
3) Sort.
4) Identify the next event relative to current time (if current time known; otherwise omit).

## Output format (must follow exactly)
**Schedule (YYYY-MM-DD)**
- HH:MM–HH:MM Title — Status

**Next up**
- … (or 'No upcoming events')

## Safety / constraints
- Never invent calendar events.
- Default timezone: America/New_York
- All times must be explicit and in `YYYY-MM-DD HH:MM`.
- If required fields are missing, ask ONLY for what is missing.
- If you wrote files, state what changed.
