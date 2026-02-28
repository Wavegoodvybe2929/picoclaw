# Core App Agent ↔ PicoClaw Integration Plan

**Created:** 2026-02-20  
**Status:** Planning Phase

## Executive Summary

This document outlines the integration strategy for connecting the **Core App Agent** (a Nuxt 4 layered monorepo framework) with **PicoClaw** (an ultra-lightweight Go-based AI assistant) to create a comprehensive development and personal operations system.

---

## System Overview

### Core App Agent
- **Architecture:** Nuxt 4 layered monorepo (TypeScript/Vue)
- **Purpose:** "Tech company in a box" framework for building applications
- **Key Features:**
  - MCP server with 12 tools (docs at `localhost:3000/mcp`)
  - Feature knowledge system (slug-based documentation)
  - Runtime configuration service (hot-reloadable settings)
  - Authentication (3-role RBAC)
  - Control plane app (`localhost:3001`)
  - Multi-layer inheritance (core → organization → apps)
- **Strengths:** Deep framework capabilities, structured development, AI-native documentation

### PicoClaw
- **Architecture:** Go-based, <10MB RAM, single binary
- **Purpose:** Ultra-lightweight personal AI assistant
- **Key Features:**
  - Skills-based extensibility
  - Memory system (guaranteed storage)
  - Calendar & task management
  - Web search integration
  - Workspace-centric operations
  - Chat app integrations (Telegram, Discord, etc.)
- **Strengths:** Minimal resources, fast boot, portable, sandboxed execution

---

## Integration Objectives

### Primary Goals
1. **Enable Core development workflow through PicoClaw**
   - Start/stop Core services from chat
   - Query Core documentation via MCP
   - Monitor Core app health and errors
   - Deploy and manage Core applications

2. **Augment PicoClaw with Core capabilities**
   - Access Core's feature knowledge system
   - Use Core's authentication for secure operations
   - Store PicoClaw configurations in Core's runtime config service

3. **Create unified AI-assisted operations**
   - Seamless context between personal tasks and development work
   - Single conversation interface for all operations
   - Memory continuity across both systems

### Use Cases

#### Developer Workflows
- "Check if the Core docs server is running"
- "Start the dashboard demo"
- "What does the layer-cascade feature do?" (queries Core MCP)
- "Show me runtime errors from the control app"
- "Deploy my SaaS app to production"

#### Personal + Development Integration
- "Add 'Review Core PR' to tomorrow's calendar at 2pm"
- "Search the web for Nuxt 4 layer best practices, then check our Core docs"
- "Remind me in 30 minutes to restart the dev server"

#### Knowledge Bridging
- "Remember: I prefer deploying with Bun runtime" → stored in both PicoClaw memory AND Core config
- "What features did we work on last week?" → PicoClaw memory + Core MCP census

---

## Architecture Options

### Option 1: MCP Client Integration ⭐ RECOMMENDED
**Approach:** Add MCP client capability to PicoClaw, connect to Core's MCP server

**Pros:**
- Leverages existing Core MCP infrastructure
- Type-safe, well-documented tool interface
- Non-invasive to both systems
- Enables rich feature knowledge access
- No API authentication needed (localhost)

**Cons:**
- Requires adding MCP client to Go codebase
- Limited to tools Core MCP exposes

**Implementation:**
```
PicoClaw Runtime
  ↓
  MCP Client (Go)
  ↓
  HTTP → localhost:3000/mcp (Core Docs Server)
  ↓
  12 MCP Tools (explain, introspect, census, get-page, etc.)
```

### Option 2: REST API Integration
**Approach:** PicoClaw skills call Core's API endpoints directly

**Pros:**
- Standard HTTP, easy to implement
- Can access runtime config, settings, auth
- Works even if MCP server is down

**Cons:**
- Requires authentication setup
- Need to manage API keys securely
- More brittle (URL changes break integration)

### Option 3: CLI Wrapper Skills
**Approach:** Create PicoClaw skills that shell out to Core CLI commands

**Pros:**
- Simple to start
- Uses existing Core tooling
- Easy debugging

**Cons:**
- Limited by Core CLI capabilities
- Text parsing fragility
- No structured responses

### Option 4: Shared Database
**Approach:** Both systems read/write to shared SQLite database

**Pros:**
- Direct data access
- No network overhead
- Transactional consistency

**Cons:**
- Tight coupling
- Schema migration complexity
- Concurrency issues

---

## Recommended Approach: Hybrid MCP + Process Management

### Phase 1: Core Process Management Skill (Week 1)
Create a PicoClaw skill for basic Core lifecycle operations.

**Skill:** `skills/core-dev/SKILL.md`

**Capabilities:**
- Start/stop Core services (docs, control, demos, apps)
- Check service health (port probing)
- View recent logs (tail last N lines)
- Restart crashed services

**Implementation:**
```bash
# In PicoClaw workspace/bin/
./core_start app=docs     # Start docs server
./core_start app=control   # Start control plane
./core_start app=all       # Start all services
./core_stop app=docs
./core_status              # JSON output: {app, port, pid, status}
./core_logs app=docs lines=50
```

**Technical Notes:**
- Use PicoClaw's `exec` tool with workspace path to Core repo
- Respect PicoClaw's workspace sandbox (allow Core path exception)
- Store Core repo path in `workspace/state/core_config.json`

### Phase 2: MCP Knowledge Bridge (Week 2)
Integrate Core's MCP server as a PicoClaw tool.

**Tool:** `pkg/core/mcp_client.go` (new package in PicoClaw)

**Capabilities:**
- Query Core documentation: `explain(slug, aspect)`
- Inspect feature registry: `introspect(slug)`, `census()`
- Read runtime logs: `recent-logs(slug, since)`
- Record knowledge: `record(slug, aspect, content)`

**Skill:** `skills/core-knowledge/SKILL.md`

**Usage in conversations:**
```
User: "How does the layer cascade work in Core?"
PicoClaw → MCP client → Core MCP → explain("layer-cascade", "overview")
PicoClaw: "Here's the overview from Core docs: [content]"
```

**Implementation Steps:**
1. Add MCP HTTP client to PicoClaw (`go get` appropriate MCP library)
2. Create wrapper functions in `pkg/core/mcp.go`
3. Register as PicoClaw internal tool (not exec-based)
4. Add to `TOOLS.md` for AI awareness

### Phase 3: Configuration Sync (Week 3)
Bidirectional sync between PicoClaw preferences and Core runtime config.

**Use Case:**
```
User to PicoClaw: "I prefer dark mode"
→ PicoClaw memory stores it
→ Also writes to Core config service: PUT /api/settings/ui.theme
```

**Implementation:**
- Add Core API client to PicoClaw
- Create `./bin/core_config_sync` command
- Run during PicoClaw heartbeat (periodic sync)
- Store last sync timestamp in `workspace/state/core_sync.json`

### Phase 4: Unified Memory (Week 4)
Share context between PicoClaw memory and Core feature knowledge.

**Capabilities:**
- PicoClaw memories → Core knowledge annotations
- Core feature work → PicoClaw memory events
- Query across both: "What did I learn about rate-limiting?"

**Implementation:**
- Extend PicoClaw `./bin/memory_write` to optionally call Core MCP `record()`
- Create Core → PicoClaw webhook for feature updates
- Unified search via MCP census + PicoClaw memory_recall

---

## Technical Specifications

### File Structure

#### PicoClaw New Files
```
workspace/
├── skills/
│   ├── core-dev/
│   │   └── SKILL.md              # Core process management
│   ├── core-knowledge/
│   │   └── SKILL.md              # MCP knowledge queries
│   └── core-config/
│       └── SKILL.md              # Core config integration
├── bin/
│   ├── core_start                # Start Core services
│   ├── core_stop                 # Stop Core services
│   ├── core_status               # Service health check
│   ├── core_logs                 # Tail service logs
│   ├── core_config_get           # Read Core config
│   ├── core_config_set           # Write Core config
│   └── core_config_sync          # Bidirectional sync
└── state/
    ├── core_config.json          # Core repo path, port mappings
    └── core_sync.json            # Last sync timestamps
```

#### Core New Files (Optional)
```
core/
├── server/
│   └── api/
│       └── picoclaw/
│           ├── webhook.post.ts   # Receive PicoClaw events
│           └── status.get.ts     # Health endpoint for PicoClaw
└── docs/
    └── knowledge/
        └── picoclaw-integration.md  # Integration docs
```

### Configuration

#### PicoClaw Config Addition
```json
{
  "integrations": {
    "core": {
      "enabled": true,
      "repo_path": "/Users/wavegoodvybe/Documents/GitHub/core",
      "mcp_endpoint": "http://localhost:3000/mcp",
      "api_endpoint": "http://localhost:3001/api",
      "auto_start_services": ["docs"],
      "sync_interval": 1800
    }
  }
}
```

**Location:** `~/.picoclaw/config.json`

#### Core Config Addition
```typescript
// organization/app/app.config.ts (or create if doesn't exist)
export default defineAppConfig({
  picoclaw: {
    enabled: true,
    webhookSecret: process.env.PICOCLAW_WEBHOOK_SECRET,
    allowedWorkspaces: [
      '/Users/wavegoodvybe/Documents/GitHub/core/workspace'
    ]
  }
})
```

### Environment Variables

#### PicoClaw
```bash
# .picoclaw/config.json or environment
PICOCLAW_CORE_REPO=/Users/wavegoodvybe/Documents/GitHub/core
PICOCLAW_CORE_ENABLED=true
```

#### Core
```bash
# Not in runtimeConfig - read from process.env
CORE_PICOCLAW_WEBHOOK_SECRET=your-secret-here
```

---

## Implementation Roadmap

### Milestone 1: Basic Process Control (3-5 days)
- [ ] Create `core-dev` skill
- [ ] Implement `core_start`, `core_stop`, `core_status` scripts
- [ ] Test starting/stopping docs and control apps
- [ ] Add to PicoClaw `TOOLS.md`
- [ ] Verify sandbox exceptions work

### Milestone 2: MCP Integration (5-7 days)
- [ ] Research Go MCP client libraries
- [ ] Implement Core MCP client in PicoClaw
- [ ] Create `core-knowledge` skill
- [ ] Test `explain()` queries in conversation
- [ ] Add error handling for when MCP server is down

### Milestone 3: Configuration Bridge (3-5 days)
- [ ] Create Core API client wrapper
- [ ] Implement `core_config_get/set` commands
- [ ] Build bidirectional sync logic
- [ ] Add to PicoClaw heartbeat tasks
- [ ] Test preference propagation

### Milestone 4: Memory Unification (7-10 days)
- [ ] Design unified memory schema
- [ ] Extend `memory_write` with Core hooks
- [ ] Create Core → PicoClaw webhook endpoint
- [ ] Build cross-system search
- [ ] Test end-to-end memory flow

### Milestone 5: Production Hardening (5 days)
- [ ] Add comprehensive error handling
- [ ] Implement health checks
- [ ] Create integration tests
- [ ] Write user documentation
- [ ] Set up monitoring/logging

**Total Estimated Time:** 4-6 weeks

---

## Security Considerations

### Workspace Sandbox
- PicoClaw defaults to `restrict_to_workspace: true`
- Core repo is OUTSIDE default workspace
- **Solution:** Add Core repo path to allowed paths in PicoClaw config
- Alternative: Keep skills in workspace but exec commands target Core

### Authentication
- MCP server: Localhost only, no auth needed
- Core API: Requires session/token
- **Solution:** 
  - Option A: Create service account in Core for PicoClaw
  - Option B: Use Core's session-only auth (cookie-based)
  - Option C: Shared secret in environment variable

### Secrets Management
- Don't store API keys in regular config
- Use PicoClaw's `.secrets/` directory (create if not exists)
- Follow Core's `multi-encrypt` pattern for encrypted storage

### Network Exposure
- All integrations are localhost-only
- No external network calls
- MCP/API only when Core services are running locally

---

## Risk Assessment

| Risk | Severity | Mitigation |
|------|----------|------------|
| Core services down breaks PicoClaw | Medium | Graceful degradation, cache results, status checks before calls |
| Port conflicts | Low | Use Core's standard ports (3000-3013), check before start |
| Workspace sandbox violations | High | Explicit configuration, validation on startup |
| Memory growth from bi-sync | Medium | Implement deduplication, periodic cleanup |
| MCP protocol changes | Low | Pin to stable MCP version, test before updates |
| Go dependency bloat | Medium | Use lightweight MCP client, vendor dependencies |

---

## Testing Strategy

### Unit Tests
- Core script functionality (start/stop/status)
- MCP client requests/responses
- Config sync logic

### Integration Tests
1. Start Core docs → verify PicoClaw can query MCP
2. Write Core config → verify sync to PicoClaw memory
3. PicoClaw command → verify affects Core state

### Manual Test Scenarios
1. Cold start: Boot PicoClaw, ensure Core detection works
2. Service failure: Kill Core docs, verify graceful error
3. Concurrent requests: Both systems accessing config simultaneously
4. Cross-system search: Query finds results from both memories

---

## Success Metrics

### Developer Experience
- [ ] Can start Core services via chat in <10 seconds
- [ ] Can query Core docs without leaving conversation
- [ ] Config changes propagate in <30 seconds
- [ ] Zero manual port/URL management needed

### System Performance
- [ ] PicoClaw memory footprint stays <20MB (with MCP client)
- [ ] MCP queries complete in <500ms
- [ ] No memory leaks over 24-hour operation
- [ ] Service health checks complete in <100ms

### Feature Completeness
- [ ] All 12 Core MCP tools accessible via PicoClaw
- [ ] Bidirectional config sync works for 10+ settings
- [ ] Unified search covers both memory systems
- [ ] Integration docs complete (this document)

---

## Future Enhancements

### Beyond MVP
1. **Visual Dashboard:** PicoClaw shows Core service status in chat UI
2. **Deployment Skills:** Deploy Core apps from PicoClaw commands
3. **Error Monitoring:** PicoClaw alerts on Core errors via heartbeat
4. **Code Generation:** PicoClaw uses Core MCP to scaffold new features
5. **Multi-Machine:** Sync between PicoClaw instances on different machines
6. **Voice Control:** "Hey PicoClaw, start the docs server"
7. **Browser Integration:** PicoClaw controls Chrome DevTools via Core's MCP

### Research Questions
- Can PicoClaw run INSIDE a Core app as an embedded assistant?
- Could Core feature wrappers emit events PicoClaw consumes?
- Is there value in Core apps having PicoClaw memory access?

---

## Next Steps

### Immediate Actions (This Week)
1. **Review this plan** with stakeholders
2. **Choose starting point:** Recommend beginning with Milestone 1 (Process Control)
3. **Set up development environment:**
   - Confirm Core repo path
   - Ensure PicoClaw has exec permissions
   - Test basic bash scripts from PicoClaw workspace
4. **Create first skill:** `skills/core-dev/SKILL.md`
5. **Prototype `core_status` script** as proof-of-concept

### Questions to Resolve
- [ ] Do we want PicoClaw to auto-start Core services on boot?
- [ ] Should config sync be push (PicoClaw → Core) or pull (Core → PicoClaw)?
- [ ] What's the preferred MCP client library for Go?
- [ ] Do we need write access to Core knowledge files from PicoClaw?
- [ ] Should integration be opt-in (config flag) or always-on?

---

## Resources

### Documentation
- Core AGENTS.md: `/Users/wavegoodvybe/Documents/GitHub/core/AGENTS.md`
- PicoClaw Workspace README: `~/workspace/README.md`
- Core MCP Docs: http://localhost:3000/internal/app-agent/ai/mcp
- PicoClaw GitHub: https://github.com/sipeed/picoclaw

### Key Files to Study
- `core/docs/server/mcp/tools/explain.ts` - MCP tool implementation example
- `workspace/skills/github/SKILL.md` - PicoClaw skill structure
- `workspace/bin/memory_write` - PicoClaw tool example
- `core/server/api/settings/[key].put.ts` - Core API endpoint example

### Go MCP Client Options (To Research)
- https://github.com/mark3labs/mcp-go (native Go implementation)
- HTTP client + JSON-RPC manual implementation
- gRPC if Core switches protocols

---

**Document Version:** 1.0  
**Last Updated:** 2026-02-20  
**Owner:** Integration Team  
**Review Date:** 2026-02-27
