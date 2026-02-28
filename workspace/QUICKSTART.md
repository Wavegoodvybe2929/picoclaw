# PicoClaw Quick Reference Card

## Essential Commands

### Verification
```bash
./bin/verify_setup              # Check entire system health
./bin/memory_status            # Check memory system
```

### Memory
```bash
./bin/memory_write --role user --content "..."
./bin/memory_recall --query "..."
./bin/memory_remember "fact"
./bin/memory_sync
./bin/memory_export            # Backup
```

### Calendar
```bash
./bin/calendar_add_event "title" "YYYY-MM-DD" "HH:MM" "HH:MM" "location" "notes"
./bin/calendar_update_event "title" ...
cat calendar/events.csv
```

### Tasks
```bash
cat tasks/todo.md
nano tasks/todo.md             # Edit tasks
```

### Email (requires Gmail token)
```bash
source .venv/bin/activate
./bin/check_gmail_unread --max 10
```

### Search (requires SearXNG)
```bash
./bin/search "query"
./bin/search_save_note "query"  # Search and save to vault
```

### Research
```bash
source .venv/bin/activate
./bin/research_links "topic" > /tmp/links.json
./bin/research_scrape /tmp/links.json > /tmp/content.json
./bin/research_write_note "topic" /tmp/content.json
```

### Notes
```bash
./bin/vault_new_note --title "..." --content "..."
./bin/vault_append --file "path" --content "..."
```

## File Locations

| Item | Location |
|------|----------|
| **Config** | `~/.picoclaw/config.json` |
| **Workspace** | `~/.picoclaw/workspace/` |
| **Memory DB** | `memory/index/memory.db` |
| **Calendar** | `calendar/events.csv` |
| **Tasks** | `tasks/todo.md` |
| **Vaults** | `vaults/{Daily,Email,Calendar,Search,Research,Inbox}/` |
| **Secrets** | `.secrets/` |
| **Logs** | `state/*.log` |

## Setup Status

| Component | Status | Next Step |
|-----------|--------|-----------|
| **Workspace** | ✅ Ready | - |
| **Memory** | ✅ Working | - |
| **Calendar** | ✅ Working | - |
| **Tasks** | ✅ Working | - |
| **Vaults** | ✅ Ready | - |
| **Python** | ✅ Ready | - |
| **Gmail** | ⚠️ Needs config | See `.secrets/gmail_token.json.PLACEHOLDER` |
| **SearXNG** | ⚠️ Needs setup | See `.secrets/SEARXNG_SETUP.md` |

## Daily Workflow

### Morning
```bash
./bin/verify_setup             # Check system
source .venv/bin/activate
./bin/check_gmail_unread       # Check email
cat tasks/todo.md              # Review tasks
```

### During Day
```bash
./bin/calendar_add_event ...   # Add events
./bin/vault_new_note ...       # Capture ideas
./bin/search_save_note "..."   # Research topics
```

### Evening
```bash
./bin/memory_export            # Backup memories
nano tasks/todo.md             # Update tasks
```

## Documentation

| Doc | Purpose |
|-----|---------|
| `docs/SETUP.md` | Complete setup guide |
| `docs/SETUP_COMPLETE.md` | Setup summary (this install) |
| `memory/MEMORY.md` | Memory system details |
| `memory/QUICKREF.md` | Memory quick reference |
| `.secrets/gmail_token.json.PLACEHOLDER` | Gmail setup |
| `.secrets/SEARXNG_SETUP.md` | Search setup |

## Troubleshooting

### Script won't run
```bash
chmod +x ./bin/scriptname
```

### Python import error
```bash
source .venv/bin/activate
python -m pip install [package]
```

### Memory system error
```bash
./bin/memory_status
# Check error details
python memory/memory_core.py sync
```

## Security

- ✅ All packages in `.venv` (never sudo)
- ✅ Secrets in `.secrets/` directory
- ✅ Gmail token should be `chmod 600`
- ✅ All data stays local
- ✅ No cloud dependencies

## Quick Win Commands

```bash
# See what the system knows
./bin/memory_recall --query "preferences"

# Add a reminder
./bin/memory_remember "User prefers concise responses"

# Create today's note
./bin/vault_new_note --title "$(date +%Y-%m-%d)" --content "# Today"

# Export everything for backup
./bin/memory_export
```

---

**Last Updated:** February 23, 2026  
**Status:** Workspace setup complete ✅
