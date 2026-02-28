# PicoClaw Workspace 🦞

**Local-first personal AI assistant with guaranteed memory**

This workspace powers PicoClaw, a lightweight personal operations system that manages your calendar, tasks, notes, and learns from every interaction.

## ⚡ Initial Setup

**First time setup** (required before using the workspace):

```bash
cd ~/.picoclaw/workspace

# 1. Create Python virtual environment
python3 -m venv .venv

# 2. Activate it
source .venv/bin/activate

# 3. Install dependencies
pip install -r requirements.txt

# 4. Verify setup
./bin/verify_setup
```

**For complete setup instructions**, see [`docs/SETUP.md`](docs/SETUP.md).

## 🚀 Quick Start

### For Users

```bash
# Check what's in your calendar today
./bin/calendar_list_today

# Capture a quick note
./bin/vault_new_note --title "Quick idea" --content "..."

# Search the web
./bin/search "topic of interest"

# Check memory system status
./bin/memory_status
```

### For the AI

Key context files (read these first):
- **AGENTS.md** - Primary instructions and workflow
- **TOOLS.md** - All available commands
- **IDENTITY.md** - Workspace structure and philosophy

## 📁 Workspace Structure

```
workspace/
├── AGENTS.md              # AI instructions (primary)
├── TOOLS.md               # Command reference
├── IDENTITY.md            # System identity
├── WORKSPACE_INTEGRATION.md  # How everything works together
│
├── bin/                   # Executable tools
│   ├── calendar_*         # Calendar management
│   ├── vault_*            # Note management
│   ├── search*            # Web search
│   ├── memory_*           # Memory system (8 commands)
│   └── ...
│
├── calendar/              # Calendar system
│   └── events.csv         # Source of truth
│
├── tasks/                 # Task management
│   └── todo.md            # Source of truth
│
├── vaults/                # Obsidian vault (all notes)
│   ├── Daily/             # Daily notes
│   ├── Email/             # Email logs
│   ├── Calendar/          # Calendar logs
│   ├── Search/            # Search logs
│   └── Research/          # Research notes
│
├── memory/                # Personal memory system ⭐ NEW
│   ├── log/               # Append-only event log
│   ├── index/             # SQLite database
│   ├── MEMORY.md          # Complete docs
│   ├── INTEGRATION.md     # Integration guide
│   └── QUICKREF.md        # Quick reference
│
└── skills/                # Skill definitions
    ├── daily-briefing/    # Morning planning
    ├── capture-note/      # Quick capture
    ├── email-check-triage-calendar/  # Email automation
    ├── memory-integration/  # Memory integration guide ⭐ NEW
    └── ...
```

## 🧠 Memory System (NEW)

PicoClaw now has a complete personal memory system that:
- **Stores** every conversation reliably
- **Recalls** relevant context intelligently
- **Learns** preferences automatically
- **Personalizes** responses over time

### Quick test:
```bash
# 1. Store a preference
./bin/memory_write --role user --content "I prefer morning briefings at 7am"

# 2. Process it
./bin/memory_sync

# 3. Recall it
./bin/memory_recall --query "briefing preferences"
```

See `memory/QUICKREF.md` for commands or `memory/MEMORY.md` for complete docs.

## 🎯 Core Capabilities

### Calendar Management
- `calendar/events.csv` is the source of truth
- Add, update, cancel events via bin/calendar_* tools
- Auto-backup before edits
- All changes logged to vault

### Task Management
- `tasks/todo.md` is the source of truth
- Simple checkbox format
- Can be populated from emails

### Note Taking (Obsidian)
- All notes in `vaults/` folder
- Open `vaults/` in Obsidian to browse
- Automated logging of email, calendar, search

### Web Research
- Local search via `./bin/search`
- No external API keys required
- Results saved to vault

### Email Automation
- Gmail integration (OAuth)
- Automated triage to tasks
- Calendar event extraction
- All logged to vault

### Personal Memory ⭐
- Guaranteed storage (append-only log)
- Smart recall (pinned + search + recent)
- Pattern detection (preferences, decisions, procedures)
- Learning from usage

## 📚 Documentation

### Getting Started
- This README - Overview
- `WORKSPACE_INTEGRATION.md` - How everything works together

### For AI
- `AGENTS.md` - Primary instructions
- `TOOLS.md` - Command reference
- `IDENTITY.md` - System identity
- `skills/*/SKILL.md` - Skill definitions

### For Memory System
- `memory/README.md` - Quick start
- `memory/QUICKREF.md` - Command cheat sheet
- `memory/MEMORY.md` - Complete documentation
- `memory/INTEGRATION.md` - Integration patterns
- `skills/memory-integration/SKILL.md` - Skill integration guide

## ✅ Integration Status

All components integrated and tested:
- ✅ Memory system implemented (8 tools)
- ✅ Context files updated (AGENTS, TOOLS, IDENTITY)
- ✅ Key skills updated (daily-briefing, capture-note, email-triage)
- ✅ Integration guide created (memory-integration skill)
- ✅ Complete documentation (5 docs)
- ✅ Testing suite (11 tests, all passing)
- ✅ No regressions to existing functionality

## 🔧 Integration Test

Run the complete integration test:
```bash
./bin/test_memory_integration
```

Should show:
```
✓ Test 1: Write user preference
✓ Test 2: Write assistant response
✓ Test 3: Write explicit remember directive
...
✓ All integration tests passed!
```

## 🎨 Philosophy

**Local-first**: Everything runs on your machine  
**Privacy-first**: Your data never leaves your computer  
**Accuracy over creativity**: For dates/times, be precise  
**Traceability**: If it matters, log it to the vault  
**Learning**: System improves from every interaction

## 🛠️ System Requirements

- Python 3.7+ (for memory system and tools)
- Go (for PicoClaw agent itself)
- Obsidian (optional, for viewing vaults)
- LM Studio (for local LLM)

## 📞 Getting Help

- Check `./bin/memory_status` for memory system health
- Review `WORKSPACE_INTEGRATION.md` for how components work together
- Read `memory/MEMORY.md` for memory system details
- Look at `skills/memory-integration/SKILL.md` for skill integration patterns

## 🚀 Next Steps

1. **Use it**: Start conversations, capture notes, check calendar
2. **Learn from it**: System remembers preferences automatically
3. **Extend it**: Add memory to custom skills (see integration guide)
4. **Monitor it**: Check `./bin/memory_status` regularly
5. **Backup it**: Run `./bin/memory_export` periodically

---

**Status**: Production ready. Everything aligned and working together. 🦞
