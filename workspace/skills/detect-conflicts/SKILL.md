---
name: detect-conflicts
description: Detect conflicting calendar events in calendar/events.csv and propose fixes.
---

# SKILL: detect_conflicts

## Goal
Find overlaps and too-tight transitions (buffer < 10 min).

## When to use
User asks 'any conflicts', 'double booked', 'overlaps'

## Inputs (from user)
- Optional: date range (default: next 7 days)

## Files used
- Read: `calendar/events.csv`
- Write: None

## Procedure
1) Load CSV.
2) Filter to window.
3) For each day, sort events.
4) Detect overlaps: next.start < current.end.
5) Detect tight transitions: next.start - current.end < 10 minutes.
6) Output conflicts with suggestions.

## Output format (must follow exactly)
**Conflicts**
- YYYY-MM-DD: Event A overlaps Event B by X min

**Tight transitions**
- YYYY-MM-DD: A → B has only X min

**Suggestions**
- …

## Safety / constraints
- Never invent calendar events.
- Default timezone: America/New_York
- All times must be explicit and in `YYYY-MM-DD HH:MM`.
- If required fields are missing, ask ONLY for what is missing.
- If you wrote files, state what changed.
