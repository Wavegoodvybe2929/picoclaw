# Core Integration Quick Start

**Goal:** Get your first Core-integrated PicoClaw skill working in 30 minutes.

## Prerequisites

- [ ] PicoClaw installed and working (`picoclaw --version`)
- [ ] Core repo cloned at `/Users/wavegoodvybe/Documents/GitHub/core`
- [ ] Core dependencies installed (`cd core && bun install`)

## Step 1: Verify Core Services (5 min)

```bash
# From Core repo root
cd /Users/wavegoodvybe/Documents/GitHub/core

# Check if services can start
bun run dev:docs    # Should start on port 3000

# Open another terminal and test MCP endpoint
curl http://localhost:3000/mcp
# Should return MCP server info (not 404)

# Stop the dev server (Ctrl+C)
```

## Step 2: Configure PicoClaw Workspace (5 min)

Add Core integration config to PicoClaw:

```bash
# Edit PicoClaw config
vim ~/.picoclaw/config.json
```

Add this section (merge with existing JSON):

```json
{
  "agents": {
    "defaults": {
      "workspace": "~/.picoclaw/workspace",
      "restrict_to_workspace": false
    }
  },
  "integrations": {
    "core": {
      "enabled": true,
      "repo_path": "/Users/wavegoodvybe/Documents/GitHub/core",
      "mcp_endpoint": "http://localhost:3000/mcp",
      "auto_start_docs": false
    }
  }
}
```

**Important:** We set `restrict_to_workspace: false` to allow accessing the Core repo outside PicoClaw's workspace.

## Step 3: Create Core Status Checker (10 min)

Create a simple bash script to check Core service status:

```bash
cd ~/.picoclaw/workspace
mkdir -p bin
touch bin/core_status
chmod +x bin/core_status
```

Edit `bin/core_status`:

```bash
#!/usr/bin/env bash
# Check status of Core services

CORE_REPO="/Users/wavegoodvybe/Documents/GitHub/core"

# Function to check if a port is listening
check_port() {
  local port=$1
  local name=$2
  if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
    echo "✓ $name (port $port): RUNNING"
    return 0
  else
    echo "✗ $name (port $port): STOPPED"
    return 1
  fi
}

echo "Core Services Status"
echo "===================="
check_port 3000 "Docs + MCP Server"
check_port 3001 "Control Plane"
check_port 3010 "Dashboard Demo"
check_port 3011 "SaaS Demo"
check_port 3012 "Landing Demo"
check_port 3013 "Chat Demo"
echo ""
echo "Core repo: $CORE_REPO"
```

Test it:

```bash
./bin/core_status
# Should show all services as STOPPED (if nothing is running)
```

## Step 4: Create Core Dev Skill (5 min)

Create the skill definition:

```bash
mkdir -p ~/.picoclaw/workspace/skills/core-dev
```

Create `~/.picoclaw/workspace/skills/core-dev/SKILL.md`:

```markdown
---
name: core-dev
description: "Manage Core App Agent services (start/stop/status). Use `./bin/core_status` to check running services."
metadata: {"picoclaw":{"requires":{"bins":["lsof","bun"]}}}
---

# Core Dev Skill

Manage Core App Agent framework services running locally.

## Check Service Status

```bash
./bin/core_status
```

Returns status of all Core services (docs, control, demos).

## Start Services (Manual)

For now, start services manually in separate terminals:

```bash
# Terminal 1: Start docs + MCP server
cd /Users/wavegoodvybe/Documents/GitHub/core
bun run dev:docs

# Terminal 2: Start control plane
cd /Users/wavegoodvybe/Documents/GitHub/core
bun run dev:control
```

## Query MCP Server

Once docs are running, you can query the MCP endpoint:

```bash
curl -X POST http://localhost:3000/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "tools/list",
    "id": 1
  }'
```

Returns list of available MCP tools.
```

## Step 5: Update TOOLS.md (2 min)

Add the new tool to PicoClaw's tool awareness:

```bash
vim ~/.picoclaw/workspace/TOOLS.md
```

Add this section:

```markdown
## Core App Agent Integration (NEW)

Check status of local Core services:

```bash
./bin/core_status
```

Shows which Core services are running:
- Docs + MCP Server (port 3000)
- Control Plane (port 3001)
- Demo apps (ports 3010-3013)

**Note:** Start Core services manually for now:
- Docs: `cd /path/to/core && bun run dev:docs`
- Control: `cd /path/to/core && bun run dev:control`
```

## Step 6: Test with PicoClaw (3 min)

Start PicoClaw in interactive mode:

```bash
picoclaw agent
```

Try these commands:

```
> Check the status of Core services

> Run ./bin/core_status for me
```

PicoClaw should execute the script and show you service status.

**Expected Output:**
```
Core Services Status
====================
✗ Docs + MCP Server (port 3000): STOPPED
✗ Control Plane (port 3001): STOPPED
...
```

## Step 7: Start Core & Verify Integration

In a separate terminal:

```bash
cd /Users/wavegoodvybe/Documents/GitHub/core
bun run dev:docs
```

Wait for "Nuxt is ready" message, then back in PicoClaw:

```
> Check Core service status again
```

**Expected Output:**
```
Core Services Status
====================
✓ Docs + MCP Server (port 3000): RUNNING
✗ Control Plane (port 3001): STOPPED
...
```

## Success! What's Next?

You now have basic Core service monitoring from PicoClaw. 

### Next Enhancements:

1. **Auto-start services:** Create `bin/core_start` script
2. **MCP queries:** Add Go MCP client for knowledge access
3. **Service logs:** Create `bin/core_logs` to tail output
4. **Health checks:** Add to PicoClaw heartbeat tasks

### Reference

Full integration plan: `~/.picoclaw/workspace/docs/core-integration-plan.md`

## Troubleshooting

### "Command blocked by safety guard"

If you see this error, you need to disable workspace restrictions:

```json
{
  "agents": {
    "defaults": {
      "restrict_to_workspace": false
    }
  }
}
```

### "bun: command not found"

Install Bun runtime:

```bash
curl -fsSL https://bun.sh/install | bash
```

### "lsof: command not found"

Install lsof (Linux):

```bash
# Ubuntu/Debian
sudo apt-get install lsof

# macOS (should be pre-installed)
# If missing, install via brew
brew install lsof
```

### Port 3000 already in use

Something else is running on port 3000. Find and stop it:

```bash
lsof -ti:3000 | xargs kill
```

## Configuration Reference

### PicoClaw Config Location
`~/.picoclaw/config.json`

### Core Repo Location
`/Users/wavegoodvybe/Documents/GitHub/core`

### PicoClaw Workspace
`~/.picoclaw/workspace`

### Integration State (future)
`~/.picoclaw/workspace/state/core_config.json`
