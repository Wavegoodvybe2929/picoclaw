# PicoClaw App Agent Integration Plan

## Executive Summary

**Goal:** Integrate App Agent's control plane, MCP server, and web UI capabilities into PicoClaw, enabling users to manage and monitor their AI agent through a modern web interface while keeping the core Go binary as the agent runtime.

**Approach:** Lightweight integration layer that connects PicoClaw's Go binary with App Agent's Nuxt-based control plane. PicoClaw remains the agent runtime, App Agent provides the UI/management layer. All features are config-driven and opt-in via config.json.

**Code Impact:** ~280 lines added to PicoClaw, 0 lines removed

---

## Design Principles

✅ **Simple** - Thin HTTP/WebSocket bridge between PicoClaw and App Agent  
✅ **Small** - Minimal code additions (~280 lines total)  
✅ **Safe** - No regressions, all changes additive  
✅ **Config-driven** - User controls everything via `~/.picoclaw/config.json`  
✅ **Optional** - App Agent integration is completely opt-in  
✅ **Non-blocking** - PicoClaw works standalone, App Agent enhances it  

---

## Current State Analysis

### PicoClaw (Existing)
- ✅ Go-based AI agent with loop system
- ✅ Tool system (exec, read_file, write_file, etc.)
- ✅ Session management
- ✅ Workspace integration with memory, calendar, vault
- ✅ Loop hooks system (before_llm, after_response, etc.)
- ✅ Gateway server for message routing
- ✅ Multi-channel support (Telegram, Discord, Slack, etc.)

### App Agent (Existing)
- ✅ Nuxt 4 web framework with control plane
- ✅ Control plane UI at localhost:3001
- ✅ MCP server for tool integration
- ✅ Runtime config service (hot-reloadable, DB-backed)
- ✅ Feature tracking and monitoring
- ✅ Authentication and RBAC

### Gap
- ❌ No way to manage PicoClaw from App Agent's web UI
- ❌ No visualization of PicoClaw agent status/activity
- ❌ No integration between PicoClaw's tools and App Agent's MCP
- ❌ No web-based configuration management for PicoClaw
- ❌ No shared session/state visibility

---

## What Changes (User Perspective)

### Before Integration
```bash
# User manages PicoClaw via CLI only
picoclaw agent -m "hello"

# No visual dashboard
# No web-based configuration
# No activity monitoring
# Manual config.json editing required
```

### After Integration
```bash
# User can still use CLI
picoclaw agent -m "hello"

# But now also has web UI available
# Open http://localhost:3001 → See PicoClaw dashboard
# - Agent status and activity
# - Message history visualization
# - Tool usage statistics
# - Web-based config editor
# - Real-time logs and monitoring
```

### User Config (`~/.picoclaw/config.json`)
```json
{
  "agents": {
    "defaults": {
      "workspace": "~/.picoclaw/workspace",
      "model": "anthropic/claude-opus-4",
      "use_workspace_tools": true,
      "loop_hooks": {
        "before_llm": [
          {
            "name": "memory_recall",
            "command": "./bin/memory_recall --query '{query}'",
            "enabled": true,
            "inject_as": "context"
          }
        ],
        "after_response": [
          {
            "name": "memory_write",
            "command": "./bin/memory_write --role assistant --content '{assistant_message}'",
            "enabled": true
          }
        ]
      }
    }
  },
  "app_agent": {
    "enabled": true,
    "control_plane_url": "http://localhost:3001",
    "api_key": "optional-for-auth",
    "features": {
      "status_reporting": true,
      "config_sync": false,
      "log_streaming": true,
      "session_sharing": true
    }
  }
}
```

---

## Implementation Plan

### Phase 1: Config Schema Extension

**Files Modified:**
- `pkg/config/config.go` (+35 lines)
- `pkg/config/defaults.go` (+15 lines)

**New Structures:**
```go
// pkg/config/config.go

type AppAgentFeatures struct {
    StatusReporting bool `json:"status_reporting"`
    ConfigSync      bool `json:"config_sync"`
    LogStreaming    bool `json:"log_streaming"`
    SessionSharing  bool `json:"session_sharing"`
}

type AppAgentConfig struct {
    Enabled         bool             `json:"enabled"`
    ControlPlaneURL string           `json:"control_plane_url"`
    APIKey          string           `json:"api_key,omitempty"`
    Features        AppAgentFeatures `json:"features"`
    ReportInterval  int              `json:"report_interval"` // seconds
}

type Config struct {
    // ... existing fields ...
    AppAgent AppAgentConfig `json:"app_agent,omitempty"`
}
```

**Defaults:**
```go
// pkg/config/defaults.go

func DefaultConfig() *Config {
    return &Config{
        // ... existing defaults ...
        AppAgent: AppAgentConfig{
            Enabled:         false, // Opt-in
            ControlPlaneURL: "http://localhost:3001",
            ReportInterval:  10, // 10 seconds
            Features: AppAgentFeatures{
                StatusReporting: true,
                ConfigSync:      false, // Conservative default
                LogStreaming:    true,
                SessionSharing:  true,
            },
        },
    }
}
```

---

### Phase 2: App Agent Client

**New File:** `pkg/appagent/client.go` (~120 lines)

**Responsibilities:**
- HTTP client for App Agent API
- Status reporting
- Session synchronization
- Log streaming

**Key Functions:**
```go
// pkg/appagent/client.go

package appagent

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
    
    "github.com/sipeed/picoclaw/pkg/config"
    "github.com/sipeed/picoclaw/pkg/logger"
)

type Client struct {
    config     config.AppAgentConfig
    httpClient *http.Client
}

func NewClient(cfg config.AppAgentConfig) *Client {
    return &Client{
        config: cfg,
        httpClient: &http.Client{
            Timeout: 5 * time.Second,
        },
    }
}

// AgentStatus represents the current state of PicoClaw
type AgentStatus struct {
    AgentID        string                 `json:"agent_id"`
    Status         string                 `json:"status"` // "running", "idle", "error"
    ActiveSessions int                    `json:"active_sessions"`
    TotalMessages  int                    `json:"total_messages"`
    Uptime         int64                  `json:"uptime"`
    Memory         map[string]interface{} `json:"memory,omitempty"`
    Tools          []string               `json:"tools"`
    Timestamp      int64                  `json:"timestamp"`
}

// ReportStatus sends agent status to App Agent control plane
func (c *Client) ReportStatus(ctx context.Context, status AgentStatus) error {
    if !c.config.Enabled || !c.config.Features.StatusReporting {
        return nil
    }
    
    url := fmt.Sprintf("%s/api/picoclaw/status", c.config.ControlPlaneURL)
    
    data, err := json.Marshal(status)
    if err != nil {
        return fmt.Errorf("failed to marshal status: %w", err)
    }
    
    req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
    if err != nil {
        return fmt.Errorf("failed to create request: %w", err)
    }
    
    req.Header.Set("Content-Type", "application/json")
    if c.config.APIKey != "" {
        req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
    }
    
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return fmt.Errorf("failed to send status: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("status report failed: %d", resp.StatusCode)
    }
    
    return nil
}

// StreamLog sends a log entry to App Agent
func (c *Client) StreamLog(ctx context.Context, entry logger.LogEntry) error {
    if !c.config.Enabled || !c.config.Features.LogStreaming {
        return nil
    }
    
    url := fmt.Sprintf("%s/api/picoclaw/logs", c.config.ControlPlaneURL)
    
    data, err := json.Marshal(entry)
    if err != nil {
        return err
    }
    
    req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
    if err != nil {
        return err
    }
    
    req.Header.Set("Content-Type", "application/json")
    if c.config.APIKey != "" {
        req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
    }
    
    // Fire and forget for logs (don't block on errors)
    go func() {
        resp, err := c.httpClient.Do(req)
        if err != nil {
            return
        }
        defer resp.Body.Close()
    }()
    
    return nil
}

// SyncSession sends session data to App Agent
func (c *Client) SyncSession(ctx context.Context, sessionData map[string]interface{}) error {
    if !c.config.Enabled || !c.config.Features.SessionSharing {
        return nil
    }
    
    url := fmt.Sprintf("%s/api/picoclaw/sessions", c.config.ControlPlaneURL)
    
    data, err := json.Marshal(sessionData)
    if err != nil {
        return err
    }
    
    req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
    if err != nil {
        return err
    }
    
    req.Header.Set("Content-Type", "application/json")
    if c.config.APIKey != "" {
        req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
    }
    
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    return nil
}
```

---

### Phase 3: Status Reporter Service

**New File:** `pkg/appagent/reporter.go` (~80 lines)

**Responsibilities:**
- Periodic status reporting to App Agent
- Collects metrics from agent loop
- Background goroutine with graceful shutdown

**Implementation:**
```go
// pkg/appagent/reporter.go

package appagent

import (
    "context"
    "sync"
    "time"
    
    "github.com/sipeed/picoclaw/pkg/agent"
    "github.com/sipeed/picoclaw/pkg/config"
    "github.com/sipeed/picoclaw/pkg/logger"
)

type Reporter struct {
    client      *Client
    agentLoop   *agent.AgentLoop
    config      config.AppAgentConfig
    startTime   time.Time
    stopChan    chan struct{}
    wg          sync.WaitGroup
}

func NewReporter(cfg config.AppAgentConfig, loop *agent.AgentLoop) *Reporter {
    return &Reporter{
        client:    NewClient(cfg),
        agentLoop: loop,
        config:    cfg,
        startTime: time.Now(),
        stopChan:  make(chan struct{}),
    }
}

// Start begins periodic status reporting
func (r *Reporter) Start(ctx context.Context) {
    if !r.config.Enabled {
        return
    }
    
    r.wg.Add(1)
    go r.runReporter(ctx)
    
    logger.InfoCF("appagent", "Status reporter started",
        map[string]any{
            "url":      r.config.ControlPlaneURL,
            "interval": r.config.ReportInterval,
        })
}

// Stop gracefully shuts down the reporter
func (r *Reporter) Stop() {
    close(r.stopChan)
    r.wg.Wait()
}

func (r *Reporter) runReporter(ctx context.Context) {
    defer r.wg.Done()
    
    ticker := time.NewTicker(time.Duration(r.config.ReportInterval) * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            r.sendStatus(ctx)
        case <-r.stopChan:
            return
        case <-ctx.Done():
            return
        }
    }
}

func (r *Reporter) sendStatus(ctx context.Context) {
    info := r.agentLoop.GetStartupInfo()
    
    status := AgentStatus{
        AgentID:        "picoclaw-default",
        Status:         "running",
        ActiveSessions: r.getActiveSessions(),
        TotalMessages:  r.getTotalMessages(),
        Uptime:         int64(time.Since(r.startTime).Seconds()),
        Tools:          info["tools"].(map[string]any)["names"].([]string),
        Timestamp:      time.Now().Unix(),
    }
    
    if err := r.client.ReportStatus(ctx, status); err != nil {
        logger.WarnCF("appagent", "Failed to report status",
            map[string]any{"error": err.Error()})
    }
}

func (r *Reporter) getActiveSessions() int {
    // TODO: Implement session counting
    return 0
}

func (r *Reporter) getTotalMessages() int {
    // TODO: Implement message counting
    return 0
}
```

---

### Phase 4: Agent Loop Integration

**Files Modified:**
- `pkg/agent/loop.go` (+25 lines)
- `cmd/picoclaw/cmd_agent.go` (+15 lines)

**Integration into Agent Loop:**
```go
// pkg/agent/loop.go

type AgentLoop struct {
    // ... existing fields ...
    appAgentReporter *appagent.Reporter // NEW
}

func NewAgentLoop(cfg *config.Config, msgBus *bus.MessageBus, provider providers.LLMProvider) *AgentLoop {
    // ... existing initialization ...
    
    al := &AgentLoop{
        bus:      msgBus,
        cfg:      cfg,
        registry: registry,
        state:    stateManager,
        // ... other fields ...
    }
    
    // NEW: Initialize App Agent reporter if enabled
    if cfg.AppAgent.Enabled {
        al.appAgentReporter = appagent.NewReporter(cfg.AppAgent, al)
    }
    
    return al
}

// StartAppAgentReporter starts status reporting to App Agent
func (al *AgentLoop) StartAppAgentReporter(ctx context.Context) {
    if al.appAgentReporter != nil {
        al.appAgentReporter.Start(ctx)
    }
}

// StopAppAgentReporter gracefully stops the reporter
func (al *AgentLoop) StopAppAgentReporter() {
    if al.appAgentReporter != nil {
        al.appAgentReporter.Stop()
    }
}
```

**Wire into CLI command:**
```go
// cmd/picoclaw/cmd_agent.go

func agentCmd() {
    // ... existing setup ...
    
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    agentLoop := agent.NewAgentLoop(cfg, msgBus, provider)
    
    // NEW: Start App Agent reporter
    agentLoop.StartAppAgentReporter(ctx)
    defer agentLoop.StopAppAgentReporter()
    
    // ... rest of command ...
}
```

---

### Phase 5: App Agent API Endpoints (App Agent Side)

**Note:** These endpoints need to be implemented in the App Agent repository to receive data from PicoClaw.

**Files to Create in App Agent:**
- `core/server/api/picoclaw/status.post.ts` (~40 lines)
- `core/server/api/picoclaw/logs.post.ts` (~30 lines)
- `core/server/api/picoclaw/sessions.post.ts` (~35 lines)
- `core/control/app/pages/picoclaw/dashboard.vue` (~100 lines)

**Example Status Endpoint:**
```typescript
// In App Agent repo: core/server/api/picoclaw/status.post.ts

export default defineFeatureHandler('picoclaw-status', async (feat, event) => {
  const body = await readBody(event);
  
  // Validate incoming data
  if (!body.agent_id || !body.status) {
    throw createError({
      statusCode: 400,
      message: 'Missing required fields'
    });
  }
  
  feat.log('status-received', {
    agentId: body.agent_id,
    status: body.status,
    uptime: body.uptime
  });
  
  // Store status in database or cache
  // TODO: Implement storage layer
  
  return {
    success: true,
    timestamp: Date.now()
  };
});
```

---

## Code Changes Summary

### PicoClaw Changes (This Repo)

**New Files (3 files, ~215 lines)**
1. `pkg/appagent/client.go` - HTTP client (~120 lines)
2. `pkg/appagent/reporter.go` - Status reporter (~80 lines)
3. `pkg/appagent/types.go` - Type definitions (~15 lines)

**Modified Files (4 files, ~65 lines)**
1. `pkg/config/config.go` - Add AppAgent config (+35 lines)
2. `pkg/config/defaults.go` - Default config (+15 lines)
3. `pkg/agent/loop.go` - Reporter integration (+10 lines)
4. `cmd/picoclaw/cmd_agent.go` - Start/stop reporter (+5 lines)

**Total: ~280 lines added, 0 lines removed**

---

### App Agent Changes (Separate Repo)

**New Files (4 files, ~205 lines)**
1. `core/server/api/picoclaw/status.post.ts` (~40 lines)
2. `core/server/api/picoclaw/logs.post.ts` (~30 lines)
3. `core/server/api/picoclaw/sessions.post.ts` (~35 lines)
4. `core/control/app/pages/picoclaw/dashboard.vue` (~100 lines)

**Total: ~205 lines added**

---

## Safety Guarantees

### ✅ No Regressions
- All existing PicoClaw commands work unchanged
- All existing features work unchanged
- All existing sessions preserved
- All existing configs compatible

### ✅ No Breaking Changes
- Everything is opt-in via `app_agent.enabled`
- Default behavior: integration disabled
- Graceful fallback if App Agent unavailable
- PicoClaw works standalone without App Agent

### ✅ No New Dependencies
- Uses existing Go standard library (net/http)
- No new third-party packages
- App Agent is external service, not dependency
- Fire-and-forget for non-critical reports

### ✅ No Performance Impact
- Status reporting in background goroutine
- Non-blocking HTTP calls
- Configurable report interval
- Minimal memory overhead (~1KB per status report)

### ✅ Network Fault Tolerance
- Graceful handling of connection failures
- No retries on status reports (fire-and-forget)
- Logs warnings but continues operation
- Auto-reconnect on next interval

---

## Integration Flow

### Status Reporting Flow
```
PicoClaw Agent Loop
    ↓ (every 10s)
Reporter goroutine
    ↓ HTTP POST
App Agent API (/api/picoclaw/status)
    ↓
Store in cache/DB
    ↓
Control Plane UI displays status
```

### Log Streaming Flow
```
PicoClaw Logger
    ↓ (real-time)
App Agent Client
    ↓ HTTP POST (async)
App Agent API (/api/picoclaw/logs)
    ↓
Store in logs DB
    ↓
Control Plane UI displays logs
```

### Session Sharing Flow
```
PicoClaw Session Save
    ↓ (on message)
App Agent Client
    ↓ HTTP POST
App Agent API (/api/picoclaw/sessions)
    ↓
Store session data
    ↓
Control Plane UI browses sessions
```

---

## Usage Examples

### Example 1: Enable App Agent Integration
```json
{
  "agents": {
    "defaults": {
      "workspace": "~/.picoclaw/workspace",
      "model": "anthropic/claude-opus-4"
    }
  },
  "app_agent": {
    "enabled": true,
    "control_plane_url": "http://localhost:3001"
  }
}
```

### Example 2: Custom Report Interval
```json
{
  "app_agent": {
    "enabled": true,
    "control_plane_url": "http://localhost:3001",
    "report_interval": 30,
    "features": {
      "status_reporting": true,
      "log_streaming": true,
      "session_sharing": false
    }
  }
}
```

### Example 3: With Authentication
```json
{
  "app_agent": {
    "enabled": true,
    "control_plane_url": "https://appagent.example.com",
    "api_key": "your-secure-api-key",
    "features": {
      "status_reporting": true,
      "config_sync": false,
      "log_streaming": true,
      "session_sharing": true
    }
  }
}
```

---

## Implementation Timeline

### Week 1: PicoClaw Side (Phases 1-3)
- **Day 1:** Config schema extension
- **Day 2-3:** App Agent client implementation
- **Day 4:** Status reporter service
- **Day 5:** Testing and verification

### Week 2: App Agent Side (Phase 5)
- **Day 1-2:** API endpoints for receiving data
- **Day 3-4:** Control plane dashboard UI
- **Day 5:** Integration testing

### Week 3: Integration & Polish (Phase 4)
- **Day 1-2:** Wire reporter into agent loop
- **Day 3:** End-to-end testing
- **Day 4:** Documentation
- **Day 5:** User testing and refinement

---

## Testing Strategy

### Unit Tests
- App Agent client HTTP requests
- Status report serialization
- Config parsing with app_agent section
- Reporter goroutine lifecycle

### Integration Tests
- PicoClaw → App Agent status reporting
- Network failure handling
- Authentication header passing
- Report interval timing

### End-to-End Tests
1. Fresh install → enable integration → status appears in UI
2. Run PicoClaw agent → see real-time updates
3. Disable integration → PicoClaw still works
4. App Agent unavailable → PicoClaw logs warnings but continues
5. Custom config → intervals and features respected

---

## Documentation Updates

### PicoClaw Documentation
1. **README.md** - Add App Agent integration section
2. **CONFIG_REFERENCE.md** - Document app_agent config
3. **INTEGRATION.md** - New file explaining integration

### App Agent Documentation
1. **README.md** - Add PicoClaw integration section
2. **API.md** - Document PicoClaw endpoints
3. **DASHBOARD.md** - Explain PicoClaw dashboard

---

## Success Criteria

✅ PicoClaw can report status to App Agent  
✅ App Agent control plane displays PicoClaw status  
✅ Zero impact when integration disabled  
✅ Graceful handling of network failures  
✅ Users can monitor PicoClaw from web UI  
✅ Total code addition under 300 lines (PicoClaw side)  
✅ All tests pass  
✅ Documentation complete  

---

## Future Extensions (Post-Implementation)

### Enhanced Monitoring
- Tool usage analytics
- Performance metrics (latency, tokens/s)
- Error rate tracking
- Memory usage graphs

### Bidirectional Communication
- Remote config updates from App Agent UI
- Remote agent control (pause/resume)
- Remote tool invocation from UI
- Session management from UI

### Advanced Features
- WebSocket for real-time updates
- Multi-agent coordination via App Agent
- Shared tool registry
- Cross-agent session handoff

---

## Risks & Mitigations

### Risk: App Agent unavailable
**Mitigation:** Fire-and-forget reporting, log warnings, continue operation

### Risk: Network latency
**Mitigation:** Async HTTP calls, configurable timeout (5s default)

### Risk: Authentication issues
**Mitigation:** Optional API key, clear error messages, fallback to unauthenticated

### Risk: Data privacy
**Mitigation:** User controls what data is shared via features config, option to disable entirely

### Risk: Version mismatch
**Mitigation:** API versioning, backward compatibility, graceful degradation

---

## Comparison with PLAN.md (Workspace Integration)

This integration follows similar patterns to the workspace integration:

| Aspect | Workspace Integration | App Agent Integration |
|--------|----------------------|----------------------|
| **Purpose** | Memory, calendar, search tools | Web UI, monitoring, control |
| **Integration** | Loop hooks call Python tools | HTTP client reports to API |
| **Config Location** | config.json `loop_hooks` | config.json `app_agent` |
| **Execution** | Synchronous in agent loop | Asynchronous background |
| **Data Flow** | PicoClaw → workspace tools | PicoClaw → App Agent API |
| **Fallback** | Disable hooks gracefully | Disable reporting gracefully |

**Key Similarities:**
- Config-driven and opt-in
- Zero impact when disabled
- Graceful degradation on failure
- Minimal code additions

**Key Differences:**
- Workspace runs locally, App Agent may be remote
- Workspace tools run synchronously, App Agent client async
- Workspace provides data to agent, App Agent displays agent data

---

## Final Summary

This plan delivers a **lightweight, HTTP-based integration** that:
- Connects PicoClaw's Go runtime with App Agent's web UI
- Enables web-based monitoring and management
- Maintains PicoClaw's standalone functionality
- Adds minimal code (~280 lines to PicoClaw)
- Changes zero existing behavior
- Provides foundation for future bidirectional features

**The result:** PicoClaw users get a modern web interface for monitoring their AI agent, viewing logs, browsing sessions, and managing configuration—all while keeping the robust Go binary as the core agent runtime.

---

## Appendix: Configuration Reference

### Complete App Agent Config Schema

```go
type AppAgentFeatures struct {
    // StatusReporting enables periodic status updates to App Agent
    // Default: true (when app_agent.enabled is true)
    StatusReporting bool `json:"status_reporting"`
    
    // ConfigSync allows App Agent to push config changes
    // Default: false (security consideration)
    ConfigSync bool `json:"config_sync"`
    
    // LogStreaming sends real-time logs to App Agent
    // Default: true
    LogStreaming bool `json:"log_streaming"`
    
    // SessionSharing sends session data to App Agent
    // Default: true  
    SessionSharing bool `json:"session_sharing"`
}

type AppAgentConfig struct {
    // Enabled controls whether App Agent integration is active
    // Default: false
    Enabled bool `json:"enabled"`
    
    // ControlPlaneURL is the App Agent control plane endpoint
    // Default: "http://localhost:3001"
    ControlPlaneURL string `json:"control_plane_url"`
    
    // APIKey for authenticated requests (optional)
    // Default: ""
    APIKey string `json:"api_key,omitempty"`
    
    // Features controls what data is shared
    Features AppAgentFeatures `json:"features"`
    
    // ReportInterval in seconds between status updates
    // Default: 10
    // Minimum: 5
    ReportInterval int `json:"report_interval"`
}
```

### Example Minimal Config

```json
{
  "app_agent": {
    "enabled": true
  }
}
```

All other values use defaults (localhost:3001, 10s interval, all features enabled).

### Example Production Config

```json
{
  "app_agent": {
    "enabled": true,
    "control_plane_url": "https://control.mycompany.com",
    "api_key": "${APP_AGENT_API_KEY}",
    "report_interval": 30,
    "features": {
      "status_reporting": true,
      "config_sync": false,
      "log_streaming": true,
      "session_sharing": false
    }
  }
}
```

Conservative settings for production: longer intervals, no config sync, no session sharing (privacy).
