---
name: task-list
description: Review and update tasks/todo.md; produce a prioritized task list and next actions.
---

# SKILL: task_list

## Goal
Maintain a simple prioritized task list aligned to the calendar.

## When to use
User asks 'to-do', 'tasks', 'what should I do next'

## Inputs (from user)
- Optional: new tasks
- Optional: priorities or deadlines

## Files used
- Read: `tasks/todo.md` (create if missing), `calendar/events.csv`
- Write: `tasks/todo.md`

## Procedure
1) Read existing `tasks/todo.md` if present.
2) If user adds tasks, append under Inbox.
3) Re-rank top tasks for today based on calendar constraints.
4) Write back cleaned list.

## Output format (must follow exactly)
**Top tasks (today)**
1) …
2) …
3) …

**Updated**: tasks/todo.md

## Safety / constraints
- Never invent calendar events.
- Default timezone: America/New_York
- All times must be explicit and in `YYYY-MM-DD HH:MM`.
- If required fields are missing, ask ONLY for what is missing.
- If you wrote files, state what changed.
