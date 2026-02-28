# PicoClaw MCP Gateway Integration Plan

**Integrating Docker MCP Gateway to extend PicoClaw's tool ecosystem**

This document outlines the plan to integrate Docker's Model Context Protocol (MCP) Gateway into PicoClaw, enabling access to containerized MCP servers and the broader MCP ecosystem.

---

## Executive Summary

### What is MCP Gateway?

Docker's MCP Gateway is an open-source orchestration layer for Model Context Protocol (MCP) servers. It:
- Acts as a centralized proxy between MCP clients and servers
- Runs MCP servers in isolated Docker containers with security restrictions
- Manages server lifecycle, authentication, and routing
- Organizes servers into profiles (named collections of servers)
- Provides access to 200+ tools via Docker MCP Catalog

### Integration Goal

Enable PicoClaw agents to access MCP servers through Docker MCP Gateway as an **optional, experimental feature**. This provides:
- Access to 200+ containerized tools (GitHub, Slack, Puppeteer, etc.)
- Zero dependency management (tools run in containers)
- Profile-based tool organization
- OAuth authentication handled by Gateway
- Security through Docker's container isolation

### Scope

**In Scope:**
- Connect PicoClaw as MCP client to Docker MCP Gateway
- Support stdio transport protocol
- Profile selection via config
- Tool discovery and execution
- Error handling and fallback
- Documentation and examples

**Out of Scope:**
- Managing Docker MCP Gateway lifecycle (user installs separately)
- Creating custom MCP servers
- GUI for profile management
- OAuth flow handling (delegated to Gateway)

---

## Phase 1: Discovery & Understanding

### 1.1 Current State Analysis

**What PicoClaw Already Has:**
- ✅ Sophisticated tool system (`pkg/tools/`)
- ✅ Tool registration and discovery
- ✅ JSON-based tool parameter schemas
- ✅ Provider-agnostic architecture
- ✅ Config-driven feature management
- ✅ Subprocess execution patterns
- ✅ Error handling and logging

**Relevant Existing Patterns:**
- `pkg/tools/` - Tool registration and execution
- `pkg/config/` - Configuration management
- `pkg/agent/loop.go` - Tool invocation during agent loop
- `cmd/picoclaw/cmd_skills.go` - External tool integration

**What Doesn't Exist Yet:**
- MCP protocol client implementation
- MCP server/tool discovery
- MCP transport layer (stdio)
- MCP message serialization/deserialization
- Integration with Docker MCP Gateway

### 1.2 MCP Protocol Overview

**MCP Architecture:**
```
PicoClaw (MCP Client)
    ↓ stdio transport
Docker MCP Gateway
    ↓ container management
MCP Servers (in containers)
    ↓ tool execution
Results back to PicoClaw
```

**MCP Communication Pattern:**
- JSON-RPC 2.0 over stdio
- Initialize handshake with capabilities
- Tool discovery via `tools/list`
- Tool execution via `tools/call`
- Resource access via `resources/*`

**Gateway CLI Command:**
```bash
docker mcp gateway run --profile <profile_name>
```

### 1.3 Gap Analysis

**Missing Components:**
1. MCP client library or implementation
2. JSON-RPC 2.0 message handling
3. Gateway subprocess management
4. MCP tool → PicoClaw tool adapter
5. Configuration schema for MCP settings
6. Error handling for Gateway failures

**Minimal Changes Needed:**
- New package: `pkg/mcp/` (~300 lines)
- Config additions: ~30 lines
- Tool integration: ~50 lines
- Documentation: ~100 lines
- **Total: ~480 lines**

---

## Phase 2: Design Constraints

### 2.1 Zero Regressions ✅

- All existing tools work unchanged
- All existing commands work unchanged
- MCP integration is completely opt-in
- Default behavior: MCP disabled
- If Gateway unavailable, gracefully degrade (no crash)

### 2.2 No Breaking Changes ✅

- New config section: `mcp.enabled = false` by default
- Existing configs remain valid
- No changes to existing tool interfaces
- No new required dependencies

### 2.3 Minimal Code Addition ✅

**Target: <500 lines**

Breakdown:
- `pkg/mcp/client.go` - MCP client implementation (~150 lines)
- `pkg/mcp/gateway.go` - Gateway process management (~80 lines)
- `pkg/mcp/types.go` - MCP message types (~60 lines)
- `pkg/config/config.go` - Config additions (~30 lines)
- `pkg/tools/mcp_adapter.go` - Tool adapter (~50 lines)
- `cmd/picoclaw/cmd_mcp.go` - CLI command (~80 lines)
- Tests - (~50 lines)

**Total: ~500 lines**

### 2.4 No New Dependencies ✅

Use **only** Go standard library:
- `encoding/json` - JSON-RPC messages
- `os/exec` - Gateway subprocess
- `io` - stdio communication
- `bufio` - Line-based reading

External requirement (user-installed):
- Docker Desktop with MCP Toolkit OR
- Docker Engine + manually installed `docker-mcp` CLI plugin

### 2.5 Config-Driven Behavior ✅

```json
{
  "mcp": {
    "enabled": false,
    "gateway": {
      "command": "docker",
      "args": ["mcp", "gateway", "run", "--profile", "default"],
      "timeout_seconds": 30,
      "auto_start": false
    },
    "tools": {
      "expose_as_builtin": false,
      "prefix": "mcp_"
    }
  }
}
```

**Defaults:**
- Disabled by default (experimental feature)
- Uses `default` profile
- 30-second timeout for Gateway operations
- Manual Gateway start (don't auto-start)
- MCP tools prefixed with `mcp_` to distinguish them

---

## Phase 3: Architecture Design

### 3.1 Integration Strategy

**Option A: MCP Tools as Native PicoClaw Tools** ✅ (Chosen)

**Why:**
- Seamless agent experience
- Reuses existing tool infrastructure
- No provider changes needed
- Easy to enable/disable

**How:**
1. Start Gateway subprocess (if configured)
2. Query Gateway for available tools
3. Create PicoClaw tool wrappers for each MCP tool
4. Register tools in agent's tool registry
5. Forward tool calls to Gateway via JSON-RPC

**Architecture:**
```
Agent requests tool execution
    ↓
Tool Registry finds MCP-wrapped tool
    ↓
MCPAdapter sends JSON-RPC to Gateway
    ↓
Gateway executes in container
    ↓
Result returned to Agent
```

### 3.2 File Organization

```
pkg/
  mcp/
    client.go         # MCP JSON-RPC client
    gateway.go        # Gateway subprocess management
    types.go          # MCP message types
    adapter.go        # Tool adaptation layer
    client_test.go    # Tests

cmd/picoclaw/
  cmd_mcp.go          # CLI: list-mcp-tools, test-mcp

config/
  config.go           # Add MCPConfig struct
  defaults.go         # Add default config
```

### 3.3 Component Design

#### 3.3.1 Gateway Manager (`pkg/mcp/gateway.go`)

```go
type GatewayManager struct {
    cmd     *exec.Cmd
    stdin   io.WriteCloser
    stdout  io.ReadCloser
    config  MCPConfig
    running bool
}

// Start launches the Gateway subprocess
func (g *GatewayManager) Start(ctx context.Context) error

// Stop gracefully terminates the Gateway
func (g *GatewayManager) Stop() error

// IsRunning checks if Gateway is responsive
func (g *GatewayManager) IsRunning() bool
```

#### 3.3.2 MCP Client (`pkg/mcp/client.go`)

```go
type Client struct {
    gateway *GatewayManager
    msgID   int64
}

// Initialize performs MCP handshake
func (c *Client) Initialize(ctx context.Context) error

// ListTools queries available MCP tools
func (c *Client) ListTools(ctx context.Context) ([]MCPTool, error)

// CallTool executes an MCP tool
func (c *Client) CallTool(ctx context.Context, name string, args map[string]any) (*MCPResult, error)

// SendRequest sends JSON-RPC request
func (c *Client) sendRequest(method string, params any) (json.RawMessage, error)
```

#### 3.3.3 MCP Tool Adapter (`pkg/mcp/adapter.go`)

```go
type MCPToolAdapter struct {
    client  *Client
    mcpTool MCPTool
    prefix  string
}

// Implements tools.Tool interface
func (a *MCPToolAdapter) Name() string
func (a *MCPToolAdapter) Description() string
func (a *MCPToolAdapter) Parameters() map[string]any
func (a *MCPToolAdapter) Execute(ctx context.Context, args map[string]any) *tools.ToolResult
```

#### 3.3.4 MCP Types (`pkg/mcp/types.go`)

```go
// JSON-RPC 2.0 message types
type JSONRPCRequest struct {
    JSONRPC string      `json:"jsonrpc"`
    ID      int64       `json:"id"`
    Method  string      `json:"method"`
    Params  interface{} `json:"params,omitempty"`
}

type JSONRPCResponse struct {
    JSONRPC string          `json:"jsonrpc"`
    ID      int64           `json:"id"`
    Result  json.RawMessage `json:"result,omitempty"`
    Error   *JSONRPCError   `json:"error,omitempty"`
}

type JSONRPCError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
}

// MCP-specific types
type MCPTool struct {
    Name        string         `json:"name"`
    Description string         `json:"description"`
    InputSchema map[string]any `json:"inputSchema"`
}

type MCPResult struct {
    Content []MCPContent `json:"content"`
    IsError bool         `json:"isError,omitempty"`
}

type MCPContent struct {
    Type string `json:"type"`
    Text string `json:"text,omitempty"`
}
```

### 3.4 Error Handling Pattern

**Graceful Degradation:**

```go
func (a *Agent) loadMCPTools(cfg config.MCPConfig) {
    if !cfg.Enabled {
        return // MCP disabled, skip
    }
    
    client, err := mcp.NewClient(cfg)
    if err != nil {
        logger.WarnCF("mcp", "Failed to initialize MCP, continuing without MCP tools",
            map[string]any{"error": err.Error()})
        return // Non-critical failure, continue
    }
    
    tools, err := client.ListTools(context.Background())
    if err != nil {
        logger.WarnCF("mcp", "Failed to list MCP tools",
            map[string]any{"error": err.Error()})
        return
    }
    
    // Register tools
    for _, tool := range tools {
        adapter := mcp.NewAdapter(client, tool, cfg.Tools.Prefix)
        a.Tools.Register(adapter)
    }
    
    logger.InfoCF("mcp", "Loaded MCP tools",
        map[string]any{"count": len(tools)})
}
```

**Tool Execution:**

```go
func (a *MCPToolAdapter) Execute(ctx context.Context, args map[string]any) *tools.ToolResult {
    result, err := a.client.CallTool(ctx, a.mcpTool.Name, args)
    if err != nil {
        return tools.ErrorResult(fmt.Sprintf("MCP tool failed: %v", err))
    }
    
    if result.IsError {
        return tools.ErrorResult(result.Content[0].Text)
    }
    
    return tools.SuccessResult(result.Content[0].Text)
}
```

---

## Phase 4: Implementation Planning

### 4.1 Phase Breakdown

#### Phase 1: Config & Foundation (Day 1)
**Goal:** Set up configuration and basic types

**Tasks:**
- [ ] Add `MCPConfig` to `pkg/config/config.go`
- [ ] Add default config to `pkg/config/defaults.go`
- [ ] Create `pkg/mcp/types.go` with JSON-RPC and MCP types
- [ ] Add unit tests for config loading

**Deliverable:** Config schema complete, tests passing

**Code estimate:** ~90 lines

---

#### Phase 2: Gateway Management (Day 2)
**Goal:** Manage Gateway subprocess lifecycle

**Tasks:**
- [ ] Implement `pkg/mcp/gateway.go`
  - `Start()` - Launch `docker mcp gateway run`
  - `Stop()` - Graceful shutdown
  - `IsRunning()` - Health check
- [ ] Add timeout handling
- [ ] Add unit tests with mock subprocess

**Deliverable:** Gateway can be started/stopped programmatically

**Code estimate:** ~80 lines

---

#### Phase 3: MCP Client (Day 3)
**Goal:** Implement JSON-RPC client

**Tasks:**
- [ ] Implement `pkg/mcp/client.go`
  - `sendRequest()` - JSON-RPC message handling
  - `Initialize()` - MCP handshake
  - `ListTools()` - Query available tools
  - `CallTool()` - Execute tools
- [ ] Add JSON-RPC ID generation
- [ ] Add response correlation
- [ ] Add unit tests

**Deliverable:** Can communicate with Gateway via JSON-RPC

**Code estimate:** ~150 lines

---

#### Phase 4: Tool Adaptation (Day 4)
**Goal:** Bridge MCP tools to PicoClaw tools

**Tasks:**
- [ ] Implement `pkg/mcp/adapter.go`
  - Implement `tools.Tool` interface
  - Map MCP tool schema to PicoClaw format
  - Forward execution to MCP client
- [ ] Add integration with agent tool registry
- [ ] Add unit tests

**Deliverable:** MCP tools appear as native PicoClaw tools

**Code estimate:** ~50 lines

---

#### Phase 5: CLI Commands (Day 5)
**Goal:** User-facing commands for management

**Tasks:**
- [ ] Implement `cmd/picoclaw/cmd_mcp.go`
  - `list-mcp-tools` - Show available MCP tools
  - `test-mcp` - Test Gateway connection
- [ ] Add to main command router
- [ ] Add CLI help text

**Deliverable:** Users can discover and test MCP integration

**Code estimate:** ~80 lines

---

#### Phase 6: Integration Testing (Day 6)
**Goal:** End-to-end validation

**Tasks:**
- [ ] Test with real Docker MCP Gateway
- [ ] Verify tool discovery
- [ ] Verify tool execution
- [ ] Test error scenarios (Gateway down, tool failure)
- [ ] Test profile switching
- [ ] Performance testing

**Deliverable:** Working end-to-end integration

---

#### Phase 7: Documentation (Day 7)
**Goal:** User and developer documentation

**Tasks:**
- [ ] Update README.md with MCP section
- [ ] Create docs/MCP_INTEGRATION.md
- [ ] Add config.example.json MCP section
- [ ] Add troubleshooting guide
- [ ] Update CHANGELOG.md

**Deliverable:** Complete documentation

**Code estimate:** ~150 lines (markdown)

---

### 4.2 Code Size Budget

| Component | Estimated Lines | Actual Lines |
|-----------|-----------------|--------------|
| pkg/mcp/types.go | 60 | TBD |
| pkg/mcp/gateway.go | 80 | TBD |
| pkg/mcp/client.go | 150 | TBD |
| pkg/mcp/adapter.go | 50 | TBD |
| pkg/config/config.go | +30 | TBD |
| cmd/picoclaw/cmd_mcp.go | 80 | TBD |
| Tests | 50 | TBD |
| **Total** | **~500** | **TBD** |

**Budget: <500 lines** ✅

### 4.3 Testing Strategy

#### Unit Tests

```go
// pkg/mcp/client_test.go
func TestClientInitialize(t *testing.T)
func TestClientListTools(t *testing.T)
func TestClientCallTool(t *testing.T)
func TestClientErrorHandling(t *testing.T)

// pkg/mcp/gateway_test.go
func TestGatewayStart(t *testing.T)
func TestGatewayStop(t *testing.T)
func TestGatewayHealthCheck(t *testing.T)

// pkg/mcp/adapter_test.go
func TestAdapterToolInterface(t *testing.T)
func TestAdapterExecution(t *testing.T)
```

#### Integration Tests

```go
// pkg/mcp/integration_test.go
func TestMCPEndToEnd(t *testing.T) {
    // Requires Docker MCP Gateway installed
    if testing.Short() {
        t.Skip("Skipping integration test")
    }
    
    // 1. Start Gateway
    // 2. List tools
    // 3. Execute tool
    // 4. Verify result
    // 5. Stop Gateway
}
```

#### Manual Testing Scenarios

1. **Gateway not installed** → Graceful warning, continue without MCP
2. **Gateway crashes** → Detect failure, log error, continue
3. **Tool execution timeout** → Return error result, don't hang
4. **Invalid profile** → Clear error message
5. **Multiple concurrent tool calls** → No race conditions

---

## Phase 5: Safety & Performance Analysis

### 5.1 Concurrency Safety

**Shared State:**
- Gateway subprocess (single instance)
- stdin/stdout streams (sequential access)
- Message ID counter (atomic increment)

**Safety Measures:**
```go
type Client struct {
    gateway *GatewayManager
    msgID   atomic.Int64  // Atomic counter
    mu      sync.Mutex    // Protect stdio access
}

func (c *Client) sendRequest(method string, params any) (json.RawMessage, error) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    // Safe stdio access
    id := c.msgID.Add(1)
    // ... send and receive
}
```

**No Deadlocks:**
- Single lock (stdio access)
- No nested locking
- Context-based timeouts

### 5.2 Performance Impact

**Latency:**
- Tool discovery: One-time cost at agent start (~100-500ms)
- Tool execution: Additional overhead ~10-50ms (subprocess communication)
- Impact: Negligible for typical LLM response times (seconds)

**Memory:**
- Gateway process: ~50MB (external process)
- MCP client: ~1MB (tool metadata + buffers)
- Total added: ~1MB in-process

**I/O:**
- Startup: One-time tool discovery (network call to Gateway)
- Runtime: One JSON-RPC request per MCP tool call
- No hot path impact (tools called infrequently)

**Benchmarks:**
```go
func BenchmarkMCPToolCall(b *testing.B) {
    client := setupTestClient()
    for i := 0; i < b.N; i++ {
        client.CallTool(context.Background(), "test_tool", nil)
    }
}
```

### 5.3 Security Checklist

**Process Execution:**
- ✅ Use `exec.CommandContext` with timeout
- ✅ Validate Gateway command path
- ✅ No shell interpretation (direct exec)

**Input Validation:**
- ✅ Validate tool names (alphanumeric + underscore)
- ✅ Validate JSON-RPC responses
- ✅ Sanitize error messages (no sensitive data)

**File System:**
- ✅ No direct file access from PicoClaw
- ✅ Gateway handles filesystem isolation
- ✅ Respect Gateway's security model

**Network:**
- ✅ No direct network calls from PicoClaw
- ✅ All requests proxied through Gateway
- ✅ Gateway enforces rate limits and access control

**Secrets:**
- ✅ Never log tool arguments (may contain sensitive data)
- ✅ OAuth handled by Gateway (not exposed to PicoClaw)
- ✅ No credential storage in PicoClaw

**Trust Model:**
- PicoClaw trusts Docker MCP Gateway
- Gateway trusts containerized MCP servers
- User must trust tools they add to profile
- Failures are non-critical (graceful degradation)

---

## Phase 6: Documentation Requirements

### 6.1 Code Documentation

**Package Documentation:**
```go
// Package mcp provides integration with Docker MCP Gateway.
//
// MCP (Model Context Protocol) is a protocol for connecting AI applications
// to external tools and data sources. This package enables PicoClaw agents
// to access containerized MCP servers via Docker's MCP Gateway.
//
// The Gateway must be installed separately:
//   - Docker Desktop with MCP Toolkit enabled, or
//   - Docker Engine + manual docker-mcp CLI plugin installation
//
// Example usage:
//   cfg := config.MCPConfig{Enabled: true}
//   client, err := mcp.NewClient(cfg)
//   tools, err := client.ListTools(ctx)
package mcp
```

**Function Documentation:**
```go
// NewClient creates an MCP client connected to Docker MCP Gateway.
//
// The Gateway subprocess is started automatically if auto_start is enabled.
// Returns an error if the Gateway is not installed or fails to start.
//
// The client must be closed with Close() to clean up the Gateway subprocess.
func NewClient(cfg config.MCPConfig) (*Client, error)
```

### 6.2 User Documentation

**README.md Addition:**

```markdown
## MCP Integration (Experimental)

PicoClaw can integrate with [Docker MCP Gateway](https://docs.docker.com/ai/mcp-catalog-and-toolkit/mcp-gateway/) to access 200+ containerized tools.

### Prerequisites

- Docker Desktop 4.62+ with MCP Toolkit, or
- Docker Engine + [manually installed MCP Gateway](https://docs.docker.com/ai/mcp-catalog-and-toolkit/mcp-gateway/#install-the-mcp-gateway-manually)

### Setup

1. Install Docker MCP Gateway (see prerequisites)
2. Create a profile in Docker Desktop or via CLI
3. Enable MCP in PicoClaw config:

```json
{
  "mcp": {
    "enabled": true,
    "gateway": {
      "args": ["mcp", "gateway", "run", "--profile", "my_profile"]
    }
  }
}
```

4. Restart PicoClaw agent

### Usage

MCP tools appear as regular tools with `mcp_` prefix:
```
> List available tools
- mcp_github_create_issue
- mcp_slack_send_message
- mcp_puppeteer_screenshot
```

### Troubleshooting

**Gateway not found:**
```
WARN[mcp] Failed to initialize MCP: exec: "docker": executable file not found
```
→ Install Docker or add to PATH

**Profile not found:**
```
ERROR[mcp] Gateway failed: profile "my_profile" not found
```
→ Create profile in Docker Desktop or use "default"
```

**New Document: `docs/MCP_INTEGRATION.md`**

```markdown
# MCP Integration Guide

## Overview

PicoClaw integrates with Docker MCP Gateway to provide access to containerized MCP servers. This enables agents to use tools like GitHub, Slack, Puppeteer, and 200+ others without manual installation.

## Architecture

[... detailed architecture diagrams and explanations ...]

## Configuration Reference

[... complete config options ...]

## Advanced Usage

### Custom Profiles
### OAuth Authentication
### Debugging

## Security Considerations

[... security model explanation ...]
```

### 6.3 Config Example

**config/config.example.json:**

```json
{
  "mcp": {
    "enabled": false,
    "gateway": {
      "command": "docker",
      "args": ["mcp", "gateway", "run", "--profile", "default"],
      "timeout_seconds": 30,
      "auto_start": false
    },
    "tools": {
      "expose_as_builtin": false,
      "prefix": "mcp_",
      "filter": {
        "allow_list": [],
        "deny_list": []
      }
    }
  }
}
```

---

## Implementation Checklist

### Design Phase
- [x] Problem definition: Enable MCP tool access via Gateway
- [x] Minimal viable implementation: Subprocess + JSON-RPC client
- [x] Architecture fit: Tool adapter pattern
- [x] File organization: `pkg/mcp/`, `cmd/picoclaw/cmd_mcp.go`
- [x] Code size estimate: ~500 lines ✅
- [x] Integration points: Tool registry, agent loop
- [x] Reuse existing: Tool system, config system, subprocess patterns

### Safety Phase
- [x] Backward compatibility: Feature disabled by default
- [x] Default behavior: No change (MCP opt-in)
- [x] Failure handling: Graceful degradation, warning logs
- [x] Security implications: Gateway manages isolation, no secrets in PicoClaw
- [x] Concurrency: Mutex for stdio, atomic message IDs
- [x] Performance: Minimal overhead, no hot path impact

### Implementation Phase
- [ ] Config schema defined
- [ ] Default config set (disabled by default)
- [ ] Core implementation (`pkg/mcp/`)
- [ ] Unit tests (>80% coverage target)
- [ ] Integration tests
- [ ] Error handling and logging
- [ ] Graceful degradation verified

### Documentation Phase
- [ ] Code comments complete
- [ ] User documentation (README.md)
- [ ] Integration guide (`docs/MCP_INTEGRATION.md`)
- [ ] Config example provided
- [ ] CHANGELOG.md updated

### Testing Phase
- [ ] All tests pass
- [ ] Manual testing with real Gateway
- [ ] Edge cases verified (Gateway down, bad profile, timeout)
- [ ] Error cases verified
- [ ] Performance acceptable (<100ms overhead per tool call)
- [ ] No regressions

---

## Success Criteria

### Functional Requirements
- [ ] Agent can discover MCP tools from Gateway
- [ ] Agent can execute MCP tools successfully
- [ ] Tool results are properly formatted
- [ ] Failures are gracefully handled
- [ ] Configuration is intuitive

### Non-Functional Requirements
- [ ] Code size: <500 lines ✅
- [ ] No new Go dependencies ✅
- [ ] Zero regressions ✅
- [ ] Tool call overhead: <100ms
- [ ] Memory overhead: <5MB

### User Experience
- [ ] Clear documentation
- [ ] Helpful error messages
- [ ] Easy to enable/disable
- [ ] Works with default profile out-of-box
- [ ] Troubleshooting guide available

---

## Risks & Mitigations

### Risk 1: Gateway Not Installed
**Impact:** Feature unusable

**Mitigation:**
- Clear documentation of prerequisites
- Graceful degradation with warning log
- CLI command to test Gateway availability
- Example: `picoclaw test-mcp`

### Risk 2: Gateway API Changes
**Impact:** Breaking changes in future Docker releases

**Mitigation:**
- Mark as experimental feature
- Version detection (future enhancement)
- Comprehensive error handling
- Regular testing with latest Docker Desktop

### Risk 3: Performance Overhead
**Impact:** Slower tool execution

**Mitigation:**
- Benchmark tool call latency
- Implement timeout limits
- Document performance characteristics
- Allow users to disable if slow

### Risk 4: Complex Debugging
**Impact:** Hard to diagnose failures

**Mitigation:**
- Verbose logging at each stage
- CLI test command
- Gateway health check
- Document common issues

---

## Future Enhancements (Out of Scope)

**Not in initial implementation, but potential future work:**

1. **Dynamic Tool Discovery**
   - Tools discovered on-demand during conversation
   - Agent asks for new tools when needed
   - Requires Dynamic MCP support

2. **Profile Management**
   - CLI commands to list/create/delete profiles
   - Config to map profiles to agents
   - Per-agent profile configuration

3. **Custom MCP Servers**
   - Documentation for building PicoClaw-aware servers
   - Example server implementations
   - Publishing to workspace

4. **Advanced Tool Filtering**
   - Allow/deny lists by tool name
   - Cost-based filtering
   - Permission-based filtering

5. **Caching & Performance**
   - Cache tool list (avoid repeated queries)
   - Connection pooling
   - Batch tool calls

6. **Gateway Lifecycle Management**
   - Auto-start Gateway if not running
   - Auto-restart on crash
   - Health monitoring

7. **Resource Access**
   - MCP resources support (beyond tools)
   - Prompts and prompt templates
   - Sampling support

---

## References

**Docker MCP Documentation:**
- [MCP Gateway](https://docs.docker.com/ai/mcp-catalog-and-toolkit/mcp-gateway/)
- [MCP Toolkit](https://docs.docker.com/ai/mcp-catalog-and-toolkit/toolkit/)
- [MCP Catalog](https://docs.docker.com/ai/mcp-catalog-and-toolkit/catalog/)
- [Dynamic MCP](https://docs.docker.com/ai/mcp-catalog-and-toolkit/dynamic-mcp/)

**Model Context Protocol:**
- [MCP Specification](https://spec.modelcontextprotocol.io/)
- [GitHub - MCP](https://github.com/modelcontextprotocol)

**PicoClaw References:**
- [IMPLEMENTATION_GUIDE.md](IMPLEMENTATION_GUIDE.md)
- [pkg/tools/](pkg/tools/) - Existing tool system
- [pkg/agent/loop.go](pkg/agent/loop.go) - Agent loop integration point

---

## Conclusion

This plan follows PicoClaw's implementation philosophy:

1. **Simple over Complex** - Thin wrapper around Gateway subprocess
2. **Small over Large** - ~500 lines of code, leveraging existing systems
3. **Safe over Fast** - Graceful degradation, experimental opt-in
4. **Config-driven** - Users control via config.json
5. **Opt-in** - Disabled by default, zero impact on existing users
6. **Extensible** - Enables future MCP ecosystem access

**Next Steps:**
1. Review and approve this plan
2. Begin Phase 1 implementation (Config & Foundation)
3. Iterate through phases with testing
4. Document and release as experimental feature

**Estimated Timeline:** 7 days for initial implementation + testing + documentation
