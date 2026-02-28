---
name: focus-block
description: Add a focused work block to calendar/events.csv with clear start/end and purpose.
---

# SKILL: focus_block

## Goal
Schedule a focus block into the calendar based on available gaps.

## When to use
User says 'block time', 'schedule deep work'

## Inputs (from user)
- Duration (e.g., 60/90/120 minutes)
- Date (default today)
- Time preference (morning/afternoon/any)

## Files used
- Read: `calendar/events.csv`
- Write: `calendar/events.csv` (append)

## Procedure
1) Read events for the date.
2) Find gaps that fit duration (include 10 min buffer).
3) Propose 1–3 options.
4) When user chooses, append new event 'Focus block' with priority 1.

## Output format (must follow exactly)
**Focus block options**
A) …
B) …
C) …

(After selection)
**Added event** …

## Safety / constraints
- Never invent calendar events.
- Default timezone: America/New_York
- All times must be explicit and in `YYYY-MM-DD HH:MM`.
- If required fields are missing, ask ONLY for what is missing.
- If you wrote files, state what changed.
