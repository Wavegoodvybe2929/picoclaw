# PicoClaw Complete Setup Guide

**Complete instructions for setting up your local-first personal AI assistant**

Last updated: February 21, 2026

---

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Initial Setup](#initial-setup)
3. [Component Configuration](#component-configuration)
   - [Memory System](#memory-system)
   - [Calendar](#calendar)
   - [Tasks](#tasks)
   - [Vault (Notes)](#vault-notes)
   - [Email Integration](#email-integration)
   - [Web Search](#web-search)
   - [Research Pipeline](#research-pipeline)
4. [Automation Setup](#automation-setup)
5. [Verification](#verification)
6. [Troubleshooting](#troubleshooting)
7. [Next Steps](#next-steps)

---

## Prerequisites

### Required Software

- **Python 3.7+** - For memory system and tools (any version 3.7 or higher works)
  ```bash
  python3 --version
  ```
  
  **Installation (if needed)**:
  ```bash
  # macOS: Install via Homebrew
  brew install python3
  ```

- **Go** - For PicoClaw agent
  ```bash
  go version
  ```
  
  **Installation (if needed)**:
  ```bash
  # macOS: Install via Homebrew
  brew install go
  
  # Or download from: https://go.dev/dl/
  ```

- **Git** - For version control
  ```bash
  git --version
  ```
  
  **Installation (if needed)**:
  ```bash
  # macOS: Install via Homebrew
  brew install git
  
  # Or install Xcode Command Line Tools:
  # xcode-select --install
  ```

### Optional Software

- **Obsidian** - For viewing and editing vaults
  - Download from: https://obsidian.md

- **LM Studio** - For local LLM
  - Download from: https://lmstudio.ai

### System Requirements

- macOS (launchd scheduling examples are macOS-specific)
- 8GB+ RAM recommended
- 2GB+ free disk space

---

## Initial Setup

### 1. Workspace Location

Your PicoClaw workspace should be at:
```bash
~/.picoclaw/workspace/
```

### 2. Verify Workspace Structure

```bash
cd ~/.picoclaw/workspace
ls -la
```

Expected folders:
```
bin/        # Executable tools
calendar/   # Calendar data
docs/       # Documentation
memory/     # Memory system
sessions/   # Session data
skills/     # Skill definitions
state/      # Runtime state
tasks/      # Task lists
vaults/     # Obsidian notes
```

### 3. Create Missing Folders

```bash
cd ~/.picoclaw/workspace
mkdir -p state .secrets calendar tasks
mkdir -p vaults/{Daily,Email,Calendar,Search,Research,Inbox}
```

### 4. Make Scripts Executable

```bash
chmod +x ./bin/*
```

### 5. Setup Python Environment

Create a virtual environment and install dependencies:

```bash
cd ~/.picoclaw/workspace

# Create virtual environment
python3 -m venv .venv

# Activate it
source .venv/bin/activate

# Upgrade pip
python -m pip install --upgrade pip

# Install all dependencies
pip install -r requirements.txt
```

**Note**: The requirements.txt includes:
- Core dependencies (requests)
- Optional Gmail integration (google-api-python-client, google-auth, google-auth-oauthlib)
- Optional web scraping (beautifulsoup4, readability-lxml)

**Skip individual packages** if you've already run the above command.

---

## Component Configuration

### Memory System

The memory system stores conversations, recalls context, and learns preferences.

#### 1. Verify Memory System Files

```bash
cd ~/.picoclaw/workspace/memory
ls -la
```

Expected files:
- `memory_core.py` - Core memory implementation
- `MEMORY.md` - Complete documentation
- `QUICKREF.md` - Quick reference
- `INTEGRATION.md` - Integration guide

#### 2. Test Memory Commands

```bash
# Check status
./bin/memory_status

# Write a test memory
./bin/memory_write --role user --content "Testing memory system"

# Sync to index
./bin/memory_sync

# Recall it
./bin/memory_recall --query "testing"
```

#### 3. Expected Memory Structure

After first use:
```
memory/
├── log/
│   └── events.ndjson       # Append-only log
├── index/
│   ├── memory.db           # SQLite database
│   └── status.json         # Sync status
└── backups/                # Timestamped backups
```

#### 4. Memory Commands Reference

| Command | Purpose |
|---------|---------|
| `memory_status` | Check system health |
| `memory_write` | Add conversation turn |
| `memory_sync` | Process log → index |
| `memory_recall` | Search memories |
| `memory_remember` | Pin important fact |
| `memory_materialize` | Show conversation |
| `memory_export` | Backup everything |
| `memory_import` | Restore from backup |

📚 **Documentation**: See `memory/QUICKREF.md` for command details

---

### Calendar

The calendar system manages events in CSV format.

#### 1. Initialize Calendar File

```bash
cd ~/.picoclaw/workspace
touch calendar/events.csv
```

#### 2. Add Your First Event

```bash
./bin/calendar_add_event \
  "Setup Complete" \
  "2026-02-21" \
  "14:00" \
  "14:30" \
  "Home" \
  "Completed PicoClaw setup"
```

#### 3. Verify Calendar Entry

```bash
cat calendar/events.csv
```

Expected format:
```csv
title,date,start_time,end_time,location,notes
Setup Complete,2026-02-21,14:00,14:30,Home,Completed PicoClaw setup
```

#### 4. Calendar Commands Reference

| Command | Purpose |
|---------|---------|
| `calendar_add_event` | Add new event |
| `calendar_update_event` | Update existing event |
| List commands | (To be added) |

⚠️ **Backup**: Calendar scripts auto-create `.bak` files before edits

---

### Tasks

The task system uses simple Markdown checkboxes.

#### 1. Initialize Task File

```bash
cd ~/.picoclaw/workspace
cat > tasks/todo.md << 'EOF'
# TODO

## Today
- [ ] Complete PicoClaw setup
- [ ] Test memory system

## This Week
- [ ] Set up email automation
- [ ] Configure calendar sync

## Backlog
- [ ] Explore skills
- [ ] Customize workflows
EOF
```

#### 2. Verify Task File

```bash
cat tasks/todo.md
```

#### 3. Task Management

Tasks can be:
- Manually edited in `tasks/todo.md`
- Auto-populated from emails
- Checked off as completed
- Organized by priority (Today / This Week / Backlog)

📝 **Source of Truth**: `tasks/todo.md`

---

### Vault (Notes)

The vault system organizes notes in Obsidian-compatible format.

#### 1. Vault Structure

```
vaults/
├── Daily/          # Daily notes (YYYY-MM-DD.md)
├── Email/          # Email summaries
├── Calendar/       # Calendar logs
├── Search/         # Web search results
├── Research/       # Research notes
└── Inbox/          # Quick captures
```

#### 2. Create Your First Note

```bash
./bin/vault_new_note \
  --title "Setup Notes" \
  --content "PicoClaw workspace configured successfully on $(date)"
```

#### 3. Open Vault in Obsidian (Optional)

1. Open Obsidian
2. Click "Open folder as vault"
3. Select: `~/.picoclaw/workspace/vaults`
4. Browse your notes with full linking and search

#### 4. Vault Commands Reference

| Command | Purpose |
|---------|---------|
| `vault_new_note` | Create new note |
| `vault_append` | Append to note |

---

### Email Integration

Email integration requires Gmail OAuth setup.

#### 1. Install Python Dependencies

**IMPORTANT**: Use virtual environment (never `sudo pip`)

```bash
cd ~/.picoclaw/workspace
source .venv/bin/activate

# Install all dependencies (if not already done in Initial Setup)
pip install -r requirements.txt

# Or install Gmail dependencies individually:
# pip install google-api-python-client google-auth google-auth-oauthlib
```

#### 2. Create Gmail OAuth Token

You need to create an OAuth token with Gmail readonly scope:

1. Go to Google Cloud Console: https://console.cloud.google.com
2. Create a new project or select existing
3. Enable Gmail API
4. Create OAuth 2.0 credentials (Desktop app)
5. Download credentials JSON
6. Run OAuth flow to generate token:

```bash
# Use Google's quickstart or your own OAuth script
# Token should have Gmail readonly scope
```

#### 3. Place Token Securely

```bash
# Copy your token to .secrets/
cp /path/to/your/token.json ~/.picoclaw/workspace/.secrets/gmail_token.json

# Lock it down
chmod 600 ~/.picoclaw/workspace/.secrets/gmail_token.json
```

#### 4. Test Email Check

```bash
source ~/.picoclaw/workspace/.venv/bin/activate
./bin/check_gmail_unread --max 10
```

Expected output:
- JSON with unread messages
- Creates `state/gmail_seen.json` for de-duplication

#### 5. Email Commands Reference

| Command | Purpose |
|---------|---------|
| `check_gmail_unread` | Fetch unread emails |

📧 **Privacy**: Token gives read-only access. Data never leaves your machine.

---

### Web Search

Web search requires a running SearXNG instance.

#### 1. Verify Search Configuration

Check your `config.json` for SearXNG endpoint:

```bash
cd ~/.picoclaw/workspace/..
cat config.json | grep -A5 searxng
```

Expected:
```json
{
  "searxng": {
    "url": "http://localhost:8080"
  }
}
```

#### 2. Test Search

```bash
cd ~/.picoclaw/workspace
./bin/search "hello world"
```

Expected:
- Writes `web/last_search.json`
- Returns search results JSON

#### 3. Save Search to Vault

```bash
./bin/search_save_note "picoclaw setup guide"
```

Expected:
- Creates note in `vaults/Search/YYYY-MM-DD/HHMM - picoclaw setup guide.md`
- Updates `vaults/Search/Search Index.md`

#### 4. Search Commands Reference

| Command | Purpose |
|---------|---------|
| `search` | Execute web search |
| `search_save_note` | Search and save to vault |

🔍 **Local-first**: Requires local SearXNG instance (no external API keys)

---

### Research Pipeline

The research pipeline extracts links, scrapes content, and creates research notes.

#### 1. Install Scraping Dependencies

```bash
cd ~/.picoclaw/workspace
source .venv/bin/activate

# Install all dependencies (if not already done in Initial Setup)
pip install -r requirements.txt

# Or install scraping dependencies individually:
# pip install requests beautifulsoup4 readability-lxml
```

#### 2. Verify Installation

```bash
python -m pip show requests beautifulsoup4 lxml readability-lxml
```

Make sure packages are installed in `.venv`:
```
Location: /Users/wavegoodvybe/.picoclaw/workspace/.venv/lib/python3.x/site-packages
```

#### 3. Test Research Pipeline

```bash
source .venv/bin/activate

# Step 1: Search and extract links
./bin/research_links "picoclaw github" > /tmp/links.json

# Step 2: Scrape content
./bin/research_scrape /tmp/links.json > /tmp/scraped.json

# Step 3: Write research note
./bin/research_write_note "picoclaw github" /tmp/scraped.json
```

Expected:
- Research note in `vaults/Research/YYYY-MM-DD/HHMM - picoclaw github.md`
- Includes scraped content and sources

#### 4. Research Commands Reference

| Command | Purpose |
|---------|---------|
| `research_links` | Extract links from search |
| `research_scrape` | Scrape content from links |
| `research_write_note` | Create research note |

---

## Automation Setup

### Scheduled Email Pipeline (macOS launchd)

Run email checks and triage every 30 minutes.

#### 1. Create Pipeline Runner Script

```bash
cat > ~/.picoclaw/workspace/bin/run_30m_pipeline << 'BASH'
#!/usr/bin/env bash
set -euo pipefail
cd "$HOME/.picoclaw/workspace"

# Activate venv for email and research tools
if [ -f ".venv/bin/activate" ]; then
  source ".venv/bin/activate"
fi

# Check Gmail unread
./bin/check_gmail_unread --max 10 > /tmp/unread.json

# TODO: Replace with your actual PicoClaw command
# Example: picoclaw run skill email-check-triage-calendar

# For now, just log that we ran
echo "$(date): Pipeline executed" >> state/pipeline.log
BASH

chmod +x ~/.picoclaw/workspace/bin/run_30m_pipeline
```

#### 2. Create launchd Job

```bash
cat > ~/Library/LaunchAgents/com.picoclaw.pipeline.plist << 'XML'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
  <dict>
    <key>Label</key>
    <string>com.picoclaw.pipeline</string>

    <key>ProgramArguments</key>
    <array>
      <string>/bin/bash</string>
      <string>-lc</string>
      <string>~/.picoclaw/workspace/bin/run_30m_pipeline</string>
    </array>

    <key>StartInterval</key>
    <integer>1800</integer>

    <key>StandardOutPath</key>
    <string>~/.picoclaw/workspace/state/launchd.out.log</string>
    <key>StandardErrorPath</key>
    <string>~/.picoclaw/workspace/state/launchd.err.log</string>

    <key>RunAtLoad</key>
    <true/>
  </dict>
</plist>
XML
```

#### 3. Load and Start

```bash
launchctl unload ~/Library/LaunchAgents/com.picoclaw.pipeline.plist 2>/dev/null || true
launchctl load ~/Library/LaunchAgents/com.picoclaw.pipeline.plist
launchctl start com.picoclaw.pipeline
```

#### 4. Verify Automation

```bash
# Check if loaded
launchctl list | grep com.picoclaw.pipeline

# Check logs
tail -f ~/.picoclaw/workspace/state/launchd.out.log
tail -f ~/.picoclaw/workspace/state/launchd.err.log
```

#### 5. Stop Automation

```bash
launchctl unload ~/Library/LaunchAgents/com.picoclaw.pipeline.plist
```

---

## Verification

### Complete System Check

Run all verification steps:

```bash
cd ~/.picoclaw/workspace

echo "=== 1. Memory System ==="
./bin/memory_status

echo -e "\n=== 2. Calendar ==="
ls -lh calendar/

echo -e "\n=== 3. Tasks ==="
cat tasks/todo.md

echo -e "\n=== 4. Vault ==="
ls -R vaults/

echo -e "\n=== 5. Email (if configured) ==="
source .venv/bin/activate 2>/dev/null || true
./bin/check_gmail_unread --max 1 2>/dev/null && echo "✓ Email working" || echo "✗ Email not configured"

echo -e "\n=== 6. Search ==="
./bin/search "test" > /dev/null 2>&1 && echo "✓ Search working" || echo "✗ Search not configured"

echo -e "\n=== 7. Python Environment ==="
which python
python --version
python -m pip freeze | grep -E "requests|beautifulsoup4|google-api"

echo -e "\n=== Setup Complete ==="
```

### Key Indicators of Success

✅ **Memory**
- `./bin/memory_status` shows healthy status
- `memory/log/events.ndjson` exists
- `memory/index/memory.db` exists

✅ **Calendar**
- `calendar/events.csv` exists
- Can add/update events without errors

✅ **Tasks**
- `tasks/todo.md` exists and is readable

✅ **Vault**
- All vault folders exist
- Can create notes with `vault_new_note`

✅ **Email** (if configured)
- Gmail token exists and is chmod 600
- `check_gmail_unread` returns JSON

✅ **Search**
- `./bin/search` writes to `web/last_search.json`

✅ **Research**
- Python packages installed in `.venv`
- Can scrape and write research notes

---

## Troubleshooting

### Memory System Issues

**Problem**: `memory_status` fails
```bash
# Check if memory files exist
ls -la memory/log/events.ndjson
ls -la memory/index/

# Recreate if needed
mkdir -p memory/log memory/index
touch memory/log/events.ndjson
```

**Problem**: Sync fails
```bash
# Check Python is working
python memory/memory_core.py sync
```

### Email Issues

**Problem**: Gmail token expired
- Regenerate OAuth token with Gmail readonly scope
- Replace `.secrets/gmail_token.json`
- Ensure `chmod 600`

**Problem**: Permission denied
```bash
chmod 600 ~/.picoclaw/workspace/.secrets/gmail_token.json
```

### Search Issues

**Problem**: Search returns no results
- Verify SearXNG is running: `curl http://localhost:8080`
- Check `config.json` has correct URL
- Restart SearXNG instance

### Python Environment Issues

**Problem**: Wrong packages installed
```bash
# Verify you're using venv
which python
# Should show: ~/.picoclaw/workspace/.venv/bin/python

# If not, activate venv
source ~/.picoclaw/workspace/.venv/bin/activate
```

**Problem**: Permission issues
- Never use `sudo pip install`
- Always use `python -m pip install` inside activated venv

### Calendar Issues

**Problem**: No backup created
- Check script has execute permissions: `chmod +x ./bin/calendar_*`

**Problem**: CSV corrupted
- Restore from `calendar/events.csv.bak`

### Automation Issues

**Problem**: launchd job not running
```bash
# Check if loaded
launchctl list | grep com.picoclaw.pipeline

# Check logs for errors
tail -100 ~/.picoclaw/workspace/state/launchd.err.log
```

**Problem**: Script fails in cron but works manually
- Ensure full paths in script
- Check venv activation works
- Verify working directory is correct

---

## Next Steps

### 1. Start Using PicoClaw

```bash
# Daily workflow
./bin/memory_status                    # Check memory health
./bin/check_gmail_unread              # Check emails
./bin/search_save_note "topic"        # Research and save
./bin/calendar_add_event ...          # Add events
./bin/vault_new_note --title "..."   # Quick capture
```

### 2. Explore Skills

```bash
# List available skills
ls -la skills/

# Read skill documentation
cat skills/daily-briefing/SKILL.md
cat skills/email-check-triage-calendar/SKILL.md
cat skills/research-subject/SKILL.md
```

### 3. Customize Configuration

Edit workspace files:
- `AGENTS.md` - AI instructions
- `TOOLS.md` - Tool reference  
- `IDENTITY.md` - System identity

### 4. Monitor System Health

```bash
# Weekly checks
./bin/memory_status                   # Memory system
./bin/memory_export                   # Backup memories
du -sh memory/                        # Check size
tail -100 state/launchd.err.log      # Check automation
```

### 5. Integrate with Obsidian

1. Open Obsidian
2. Open `~/.picoclaw/workspace/vaults` as vault
3. Browse notes, use search, create links
4. PicoClaw will sync changes automatically

### 6. Learn Memory Integration

Read memory integration guides:
- `memory/MEMORY.md` - Complete documentation
- `memory/INTEGRATION.md` - Integration patterns
- `skills/memory-integration/SKILL.md` - Skill integration

### 7. Create Custom Skills

Use the skill-creator skill to make your own:
```bash
cat skills/skill-creator/SKILL.md
```

---

## Security Checklist

Before going into production:

- [ ] Gmail token is `chmod 600`
- [ ] Never used `sudo pip install`
- [ ] All Python packages in `.venv`
- [ ] No `curl | bash` commands run
- [ ] Calendar backups working (`.bak` files created)
- [ ] Vault notes only write to `vaults/`
- [ ] Memory exports run regularly
- [ ] Automation logs are monitoring for errors

---

## References

### Quick Links

- **Main README**: `README.md`
- **AI Instructions**: `AGENTS.md`
- **Tool Reference**: `TOOLS.md`
- **Memory Docs**: `memory/MEMORY.md`
- **Memory Quick Ref**: `memory/QUICKREF.md`

### Documentation Files

- This guide: `docs/SETUP.md`
- Email scheduling: `docs/launchd-email-pipeline.md`
- Safe setup plan: `docs/PicoClaw_Workspace_Safe_Setup_Plan.md`
- Memory enhancement: `docs/memory-enhancement-hybrid-plan.md`

### Support

For issues:
1. Check `./bin/memory_status`
2. Review logs in `state/`
3. Verify component setup above
4. Check skill documentation in `skills/*/SKILL.md`

---

**Status**: Your PicoClaw workspace is ready! 🦞

Everything is local-first, private, and built to learn from every interaction.
