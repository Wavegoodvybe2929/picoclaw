---
name: email-check-triage-calendar
description: Every 30 minutes: check Gmail unread/new, write vault summaries, triage to tasks/todo.md, and upsert calendar/events.csv based on scheduling emails.
---

# SKILL: email_check_triage_calendar

## Goal
Run the operational pipeline:
1) Gmail unread/new check
2) Write summaries into the Obsidian vault (Daily + Email log)
3) Convert actionable emails into tasks (tasks/todo.md + vault)
4) Create/update calendar events in calendar/events.csv (with backups + vault logging)

## Files & folders (source of truth)
- Vault root: `vaults/` (all human notes go here)
- Calendar source of truth: `calendar/events.csv`
- Tasks: `tasks/todo.md`

## Tool policy
- Do **not** use built-in web tools.
- Use exec scripts in `./bin/` only.

## Procedure

### 0) Get relevant memory context (optional but recommended)
Run:
- exec: `./bin/memory_recall --query "email preferences calendar important contacts" --format markdown`
- Use this context to better triage emails (know what user cares about, how they like scheduling, etc.)

### 1) Check Gmail unread/new
Run:
- exec: `./bin/check_gmail_unread --max 10`

If the JSON indicates `new_unread_count == 0`, do nothing (avoid noisy logs).

### 2) Vault summary logging
When there is new unread mail:
- Append a short summary to: `vaults/Daily/YYYY-MM-DD.md`
- Append an entry to: `vaults/Email/Inbox Log.md` (append-only)

Use:
- exec: `./bin/vault_append "Daily/YYYY-MM-DD.md"`
- exec: `./bin/vault_append "Email/Inbox Log.md"`

Summary format:
- "New mail: N" + 3–7 bullets (sender + subject + 1-line gist)

### 3) Email triage → tasks
From the same Gmail JSON:
- Classify each message: Action / FYI / Ignore
- For Action emails:
  - Append checkboxes to `tasks/todo.md`
  - Append the same tasks (with context) to `vaults/Inbox/Email Tasks.md`
  - Add a one-line note to `vaults/Daily/YYYY-MM-DD.md`

### 4) Email → calendar upserts
Create or update an event **only** if the email contains:
- explicit date
- explicit start time
- title/subject usable as event title
- (optional) location or meeting link

Ignore:
- vague timing ("sometime next week")
- marketing/newsletters

Rules:
- If end time missing: default duration 30 minutes.
- Timezone default: America/New_York unless email explicitly says otherwise.

Event handling (safe + non-destructive):
- Prefer: if a clear matching event exists, update; otherwise create.
- Always use the scripts (they write `.bak` before edits):
  - exec: `./bin/calendar_add_event ...`
  - exec: `./bin/calendar_update_event ...`

Logging:
- Append changes to `vaults/Calendar/Calendar Changes.md` (append-only)
- Append one line to Daily note:
  - "Calendar updated: X added, Y updated."

### 5) Store workflow execution in memory (for learning)
After pipeline completes:
- exec: `./bin/memory_write --role system --content "Email pipeline completed: [summary]" --type workflow --metadata '{"skill":"email-check-triage-calendar"}'`
- exec: `./bin/memory_sync` (in background, processes new memories)

This helps the system learn patterns and preferences over time.

## Done when
- Script runs without interactive passwords.
- Only logs when new/unread exists.
- Action emails become tasks (both files).
- Meeting emails create/update `calendar/events.csv` with `.bak` backups and vault logs.
- Workflow execution stored in memory for future reference.
