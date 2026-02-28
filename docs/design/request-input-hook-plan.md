# Request Input Hook - Implementation Plan

**Feature**: New loop hook type for gathering additional user input during agent processing  
**Status**: Planning Phase  
**Date**: 2026-02-25  
**Agent System Protocol**: Orchestrator → Go Specialist → Test Specialist → Documentation → Memory

---

## Executive Summary

Add a new `request_input` hook type that allows the agent to pause processing and request additional input from the user. This enables interactive workflows where the agent can ask clarification questions, request confirmations, or gather structured input mid-conversation.

**Core Principle**: Extend existing hook system with minimal code, maintain backward compatibility, make interactive capability opt-in.

---

## Phase 1: Discovery & Understanding

### 1.1 Current State Analysis

**Existing Hook System** (pkg/agent/hooks.go, pkg/agent/loop.go):
- ✅ 4 hook types: `BeforeLLM`, `AfterResponse`, `OnToolCall`, `OnError`
- ✅ Template variable substitution (e.g., `{user_message}`, `{tool_name}`)
- ✅ Context injection via `inject_as: "context"`
- ✅ Python venv auto-detection and activation
- ✅ 30-second timeout per hook
- ✅ Graceful failure (logs warning, continues processing)

**Hook Lifecycle**:
```
User Message → BeforeLLM hooks → LLM Processing → Tool Calls → OnToolCall hooks → AfterResponse hooks
                                                    ↓
                                              OnError hooks (if error)
```

**Message Bus** (pkg/bus/):
- InboundMessage: User → Agent
- OutboundMessage: Agent → User
- Channels: Slack, Telegram, Discord, etc.

**Current Limitations**:
- All hooks are fire-and-forget (no return interaction)
- No mechanism to pause and wait for user response
- No way to dynamically gather input during processing

### 1.2 Requirements Gathering

**Use Cases**:
1. **Clarification**: Agent asks user for missing details
   - "Which environment did you mean: staging or production?"
   
2. **Confirmation**: Agent seeks approval before action
   - "This will delete 50 files. Confirm? (yes/no)"
   
3. **Structured Input**: Agent collects specific data
   - "What's the priority level? (low/medium/high)"
   
4. **Multi-step Workflow**: Agent guides user through process
   - "Step 1/3: Enter server hostname..."

**Must-Haves**:
- Synchronous blocking wait for user response
- Timeout mechanism (don't wait forever)
- Template variable support in prompts
- Integration with existing hook system
- Works across all channels (Slack, Telegram, etc.)

**Nice-to-Haves**:
- Validation of user responses
- Retry on invalid input
- Default values if timeout expires
- Multiple choice prompts

### 1.3 Gap Analysis

**Missing Functionality**:
1. ❌ Hook type that blocks and waits for user input
2. ❌ Mechanism to send prompt to user via bus
3. ❌ Mechanism to receive and route user response back to hook
4. ❌ Timeout handling for abandoned prompts
5. ❌ State management for pending input requests

**Integration Points**:
- `pkg/config/config.go` - Add RequestInput hook type
- `pkg/agent/hooks.go` - Add interactive execution method
- `pkg/agent/loop.go` - Wire request_input hooks
- `pkg/bus/` - May need input request/response types

---

## Phase 2: Design Constraints

### 2.1 Zero Regressions ✅
- ✅ New hook type only used if configured
- ✅ Existing 4 hook types unchanged
- ✅ Existing tests continue passing
- ✅ No changes to existing tool behavior
- ✅ Default config has no request_input hooks

### 2.2 No Breaking Changes ✅
- ✅ Opt-in feature (disabled by default)
- ✅ Config schema backward compatible (omitempty)
- ✅ Graceful degradation if user doesn't respond
- ✅ Works with existing channels without modification

### 2.3 Minimal Code Addition ✅
**Estimated**: ~350 lines total

**Breakdown**:
```
pkg/config/config.go           +8 lines   (add RequestInput []LoopHook)
pkg/agent/hooks.go             +120 lines (ExecuteRequestInputHook method)
pkg/agent/loop.go              +80 lines  (integration point, state management)
pkg/bus/types.go               +20 lines  (InputRequest/InputResponse types)
config/config.example.json     +20 lines  (example config)
docs/WORKSPACE_INTEGRATION.md  +100 lines (documentation)
Tests                          +150 lines (unit + integration tests)

Total: ~498 lines (under 500 ✅)
```

### 2.4 No New Dependencies ✅
- ✅ Uses existing Go stdlib (context, sync, time)
- ✅ Uses existing bus package for messaging
- ✅ Uses existing hook execution infrastructure

### 2.5 Config-Driven Behavior ✅

**Configuration Example**:
```json
{
  "agents": {
    "defaults": {
      "loop_hooks": {
        "request_input": [
          {
            "name": "confirm_action",
            "command": "./bin/format_prompt '{prompt_text}'",
            "enabled": true,
            "timeout": 300,
            "return_as": "user_input",
            "metadata": {
              "description": "Request user confirmation",
              "required_vars": ["prompt_text"]
            }
          }
        ]
      }
    }
  }
}
```

**New Fields**:
- `timeout`: Seconds to wait for user response (default: 60)
- `return_as`: Variable name to store user's response
- `default_value`: Value to use if timeout expires (optional)

---

## Phase 3: Architecture Design

### 3.1 Integration Strategy

**Option A: New Hook Point** (RECOMMENDED)
Add 5th hook type: `RequestInput`

**When triggered**: Explicitly via special tool or hook condition

**Execution flow**:
```
Agent determines it needs input
    ↓
Execute RequestInput hook
    ↓
Hook sends prompt to user (via bus)
    ↓
Block and wait for user response (with timeout)
    ↓
Return user response to agent
    ↓
Continue processing with response
```

**Option B: Extend Existing Hooks** (NOT RECOMMENDED)
Add `wait_for_response: true` flag to existing hooks

**Rejected because**: Adds complexity to simple hooks, violates single responsibility

### 3.2 Detailed Design

#### 3.2.1 Config Schema Extension

```go
// pkg/config/config.go

type LoopHook struct {
    Name        string            `json:"name"`
    Command     string            `json:"command"`
    Enabled     bool              `json:"enabled"`
    InjectAs    string            `json:"inject_as,omitempty"`
    Metadata    map[string]string `json:"metadata,omitempty"`
    
    // NEW: For request_input hooks only
    Timeout      int    `json:"timeout,omitempty"`       // Seconds to wait (default: 60)
    ReturnAs     string `json:"return_as,omitempty"`     // Variable name for response
    DefaultValue string `json:"default_value,omitempty"` // Value if timeout
}

type LoopHooks struct {
    BeforeLLM     []LoopHook `json:"before_llm,omitempty"`
    AfterResponse []LoopHook `json:"after_response,omitempty"`
    OnToolCall    []LoopHook `json:"on_tool_call,omitempty"`
    OnError       []LoopHook `json:"on_error,omitempty"`
    RequestInput  []LoopHook `json:"request_input,omitempty"` // NEW
}
```

#### 3.2.2 Bus Message Types

```go
// pkg/bus/types.go

type InputRequest struct {
    RequestID string // UUID for tracking
    Channel   string
    ChatID    string
    Prompt    string // Question/prompt for user
    Timeout   int    // Seconds to wait
}

type InputResponse struct {
    RequestID string
    Input     string // User's response
    TimedOut  bool   // True if expired
}
```

#### 3.2.3 Hook Executor Extension

```go
// pkg/agent/hooks.go

// ExecuteRequestInputHook executes a hook that waits for user response.
// Returns the user's input or default value if timeout.
func (h *HookExecutor) ExecuteRequestInputHook(
    ctx context.Context,
    hook config.LoopHook,
    vars map[string]string,
    bus *bus.MessageBus,
    channel string,
    chatID string,
) (string, error) {
    // 1. Execute hook command to generate prompt
    prompt, err := h.executeHook(ctx, hook, vars)
    if err != nil {
        return hook.DefaultValue, fmt.Errorf("failed to generate prompt: %w", err)
    }
    
    // 2. Create input request
    requestID := uuid.New().String()
    timeout := hook.Timeout
    if timeout == 0 {
        timeout = 60 // Default 60 seconds
    }
    
    request := bus.InputRequest{
        RequestID: requestID,
        Channel:   channel,
        ChatID:    chatID,
        Prompt:    prompt,
        Timeout:   timeout,
    }
    
    // 3. Send request to user via bus
    bus.PublishInputRequest(request)
    
    // 4. Wait for response (with timeout)
    responseChan := bus.SubscribeInputResponse(requestID)
    
    select {
    case response := <-responseChan:
        if response.TimedOut {
            logger.WarnCF("agent", "Input request timed out",
                map[string]any{"request_id": requestID})
            return hook.DefaultValue, nil
        }
        return response.Input, nil
        
    case <-time.After(time.Duration(timeout) * time.Second):
        logger.WarnCF("agent", "Input request timeout expired",
            map[string]any{"request_id": requestID})
        return hook.DefaultValue, nil
        
    case <-ctx.Done():
        return hook.DefaultValue, ctx.Err()
    }
}
```

#### 3.2.4 Agent Loop Integration

```go
// pkg/agent/loop.go - in runAgentLoop

// NEW: Support for request_input hooks via special tool
// Option A: Via explicit tool call (e.g., agent calls "request_input" tool)
if toolName == "request_input" {
    prompt := args["prompt"].(string)
    hookVars["prompt_text"] = prompt
    
    // Execute request_input hooks
    for _, hook := range al.cfg.Agents.Defaults.LoopHooks.RequestInput {
        if !hook.Enabled {
            continue
        }
        
        userInput, err := hookExecutor.ExecuteRequestInputHook(
            ctx, hook, hookVars, al.bus, opts.Channel, opts.ChatID,
        )
        
        if err != nil {
            // Handle error - use default or fail gracefully
            toolResult = tools.ErrorResult(err)
        } else {
            // Return user input to agent
            hookVars[hook.ReturnAs] = userInput
            toolResult = tools.SuccessResult(userInput)
        }
    }
}

// Option B: Via hook evaluation (hooks can trigger themselves)
// This would require hooks to return special "request_input" directive
```

### 3.3 Error Handling Pattern

```go
// Graceful degradation
func (h *HookExecutor) ExecuteRequestInputHook(...) (string, error) {
    // If bus unavailable
    if bus == nil {
        logger.WarnCF("agent", "Bus unavailable for input request",
            map[string]any{"hook": hook.Name})
        return hook.DefaultValue, nil
    }
    
    // If timeout expires
    if response.TimedOut {
        logger.WarnCF("agent", "User did not respond in time",
            map[string]any{"hook": hook.Name, "timeout": timeout})
        return hook.DefaultValue, nil
    }
    
    // If channel doesn't support interactive input
    if !supportsInteractiveInput(channel) {
        logger.WarnCF("agent", "Channel doesn't support interactive input",
            map[string]any{"channel": channel})
        return hook.DefaultValue, nil
    }
}
```

---

## Phase 4: Implementation Planning

### 4.1 Phase Breakdown

**Phase 1: Bus Extensions** (Day 1)
- Add InputRequest and InputResponse types
- Add PublishInputRequest method
- Add SubscribeInputResponse method
- Unit tests for bus changes

**Phase 2: Hook Executor Extension** (Day 2)
- Implement ExecuteRequestInputHook method
- Add timeout and response handling
- Unit tests for hook executor

**Phase 3: Config Schema** (Day 3)
- Add RequestInput to LoopHooks
- Add timeout, return_as, default_value fields
- Update config loading and validation
- Config tests

**Phase 4: Agent Loop Integration** (Day 4)
- Add request_input tool (optional)
- Wire hooks into loop
- Integration tests

**Phase 5: Documentation & Examples** (Day 5)
- Update WORKSPACE_INTEGRATION.md
- Add config.example.json examples
- Create example scripts
- User guide

### 4.2 Code Size Validation

```
New files:
  None (all modifications to existing files)

Modified files:
  pkg/bus/types.go               +20 lines  (new types)
  pkg/bus/bus.go                 +30 lines  (pub/sub methods)
  pkg/agent/hooks.go             +120 lines (ExecuteRequestInputHook)
  pkg/agent/loop.go              +80 lines  (integration code)
  pkg/config/config.go           +8 lines   (schema extension)
  config/config.example.json     +20 lines  (examples)
  docs/WORKSPACE_INTEGRATION.md  +100 lines (documentation)

Tests:
  pkg/bus/bus_test.go            +40 lines
  pkg/agent/hooks_test.go        +60 lines
  pkg/agent/loop_test.go         +50 lines

Total: ~528 lines (slightly over target)
```

**Optimization opportunities**:
- Simplify bus implementation (reuse existing patterns)
- Combine some test cases
- Target: Reduce to <500 lines

### 4.3 Testing Strategy

**Unit Tests**:
```go
// pkg/agent/hooks_test.go
func TestExecuteRequestInputHook(t *testing.T) { }
func TestRequestInputTimeout(t *testing.T) { }
func TestRequestInputDefaultValue(t *testing.T) { }
func TestRequestInputCancellation(t *testing.T) { }

// pkg/bus/bus_test.go
func TestInputRequestPublish(t *testing.T) { }
func TestInputResponseRouting(t *testing.T) { }
```

**Integration Tests**:
```go
// pkg/agent/loop_test.go
func TestRequestInputHookIntegration(t *testing.T) {
    // Full flow: hook → bus → response → continue
}

func TestRequestInputWithChannels(t *testing.T) {
    // Test with different channel types
}
```

**Manual Testing Scenarios**:
1. Configure request_input hook
2. Agent calls request_input tool
3. User receives prompt
4. User responds within timeout → success
5. User doesn't respond → default value used
6. Multiple concurrent requests → proper routing

---

## Phase 5: Safety & Performance Analysis

### 5.1 Concurrency Safety ✅

**Concerns**:
- Multiple pending input requests
- Concurrent access to response channels
- Race conditions in bus subscription

**Solutions**:
- Use request ID for routing (no collisions)
- Each request gets unique response channel
- Mutex-protected subscription map in bus
- Context cancellation on timeout

**Pattern**:
```go
type MessageBus struct {
    mu               sync.RWMutex
    inputSubscribers map[string]chan InputResponse
}

func (mb *MessageBus) SubscribeInputResponse(requestID string) <-chan InputResponse {
    mb.mu.Lock()
    defer mb.mu.Unlock()
    
    ch := make(chan InputResponse, 1)
    mb.inputSubscribers[requestID] = ch
    return ch
}
```

### 5.2 Performance Impact ✅

**Agent Loop**:
- ✅ No impact when feature not used
- ✅ Blocking wait only during active request
- ⚠️ Could delay response if user slow to respond

**Latency**:
- Request creation: <1ms
- Bus routing: <5ms  
- User response time: Variable (1-300s)
- Timeout handling: 0ms (channel select)

**Memory**:
- Each pending request: ~100 bytes (lightweight)
- Response channels: 1 per request (cleaned up after)
- Expected concurrent requests: <10
- **Total overhead**: <1KB

**Mitigation**:
- Set reasonable default timeout (60s)
- Clean up expired subscriptions
- Limit concurrent requests per agent

### 5.3 Security Checklist ✅

- ✅ Request IDs are UUIDs (no enumeration)
- ✅ Responses routed by ID (no eavesdropping)
- ✅ Workspace restrictions still apply to hook scripts
- ✅ User input sanitized by LLM context builder
- ✅ Timeout prevents infinite waiting
- ✅ No secrets in hooks (same as existing)

**Additional Considerations**:
- User input should be validated before use
- Hooks should not execute arbitrary user input
- Rate limiting on request_input calls (prevent abuse)

---

## Phase 6: Documentation Requirements

### 6.1 User Documentation

**WORKSPACE_INTEGRATION.md** (new section):
```markdown
## Request Input Hooks

Request input hooks allow the agent to pause and gather additional 
information from the user during processing.

### Configuration

{
  "loop_hooks": {
    "request_input": [
      {
        "name": "confirm_deployment",
        "command": "./bin/format_confirmation '{environment}' '{service}'",
        "enabled": true,
        "timeout": 120,
        "return_as": "user_confirmation",
        "default_value": "no"
      }
    ]
  }
}

### Usage

The agent can trigger input requests in two ways:

1. Via special tool call:
   - Agent: "I need to confirm before proceeding"
   - Tool call: request_input(prompt="Deploy to production?")
   - User receives: "Deploy to production? (yes/no)"
   - User responds: "yes"
   - Agent continues with user's response

2. Via conditional hook:
   - Hook evaluates condition
   - If true, sends prompt
   - Waits for response
   - Returns response to agent

### Template Variables

- {prompt_text} - The prompt text from tool args
- {user_message} - Original user message
- All standard hook variables

### Best Practices

- Set appropriate timeouts (60-300s)
- Always provide default_value
- Keep prompts clear and concise
- Validate user responses in hooks
- Handle timeout gracefully
```

### 6.2 Code Documentation

Every new function documented with:
- Purpose and behavior
- Parameters and return values
- Error conditions
- Concurrency safety notes
- Example usage

### 6.3 Config Schema Docs

```go
// RequestInput hooks execute when the agent needs additional input
// from the user. These hooks block execution until user responds or
// timeout expires.
//
// Example use cases:
//   - Confirming destructive actions
//   - Gathering missing parameters
//   - Interactive multi-step workflows
RequestInput []LoopHook `json:"request_input,omitempty"`
```

---

## Implementation Checklist

### Design Phase ✅
- [x] Problem clearly defined
- [x] Minimal viable implementation designed
- [x] Architecture fits existing system
- [x] Files to modify identified
- [x] Code size estimated (<500 lines, ~528 currently)
- [x] Integration points identified
- [x] Existing code reuse maximized

### Safety Phase ✅
- [x] Backward compatibility maintained
- [x] Existing behavior preserved (feature opt-in)
- [x] Failure modes designed (timeout, default value)
- [x] Security implications analyzed
- [x] Concurrency safety designed
- [x] Performance impact minimal

### Pre-Implementation Phase
- [ ] Config schema finalized
- [ ] Default config set (feature disabled)
- [ ] Error handling strategy confirmed
- [ ] Test plan created (>80% coverage target)
- [ ] Documentation outline created

### Implementation Phase (Future)
- [ ] Bus types added
- [ ] Hook executor extended
- [ ] Config schema updated
- [ ] Agent loop integration complete
- [ ] Unit tests written
- [ ] Integration tests written
- [ ] Manual testing complete

### Documentation Phase (Future)
- [ ] Code comments complete
- [ ] WORKSPACE_INTEGRATION.md updated
- [ ] Config examples provided
- [ ] Example scripts created
- [ ] CHANGELOG.md entry added

---

## Risk Assessment

### High Risk
None identified - feature is additive and opt-in

### Medium Risk
1. **User confusion about when to use**
   - *Mitigation*: Clear documentation with examples
   - *Mitigation*: Start disabled by default

2. **Timeout tuning challenges**
   - *Mitigation*: Configurable per hook
   - *Mitigation*: Reasonable default (60s)
   - *Mitigation*: Always provide default_value

3. **Channel compatibility**
   - *Mitigation*: Works with existing OutboundMessage
   - *Mitigation*: Graceful degradation if unsupported

### Low Risk
1. **Performance impact of blocking**
   - *Impact*: Only affects agents using feature
   - *Mitigation*: Timeout prevents infinite wait

2. **Memory leak from abandoned requests**
   - *Mitigation*: Timeout cleanup
   - *Mitigation*: Context cancellation

---

## Alternatives Considered

### Alternative 1: Async Callbacks
Instead of blocking, use callbacks when user responds.

**Rejected**: More complex, harder to reason about, breaks sequential flow

### Alternative 2: External State Machine
Store request state externally, poll for responses.

**Rejected**: More moving parts, introduces external dependency

### Alternative 3: Tool-Only Approach
Only allow via tools, no hook system integration.

**Accepted for MVP**: Simpler, can add hook integration later

---

## Success Criteria

**Feature is successful if**:
1. User can configure request_input hooks
2. Agent can pause and request user input
3. User response is captured and returned to agent
4. Timeout mechanism works reliably
5. Zero regressions in existing functionality
6. <500 lines of code (streamlined version)
7. >80% test coverage
8. Clear documentation and examples

---

## Next Steps

Following Orchestrator protocol:

1. **Review this plan** with stakeholders
2. **Route to Go Specialist** for implementation
3. **Route to Test Specialist** for test strategy review
4. **Route to Documentation Specialist** for docs review
5. **Route to Memory Specialist** to log this plan

**Estimated Timeline**: 5 days for full implementation

**Dependencies**: None - feature is self-contained

**Approval Required**: Design review before implementation begins

---

## Appendix

### Example Scripts

**bin/format_prompt**:
```bash
#!/bin/bash
# Format a prompt for user

prompt="$1"
echo "🤔 $prompt"
echo ""
echo "Please respond with your input."
```

**bin/validate_response**:
```bash
#!/bin/bash
# Validate user response against allowed values

response="$1"
allowed="$2"  # comma-separated

if echo "$allowed" | grep -q "$response"; then
    echo "$response"
    exit 0
else
    echo "Invalid response. Please try again." >&2
    exit 1
fi
```

### Example Configuration

```json
{
  "agents": {
    "defaults": {
      "loop_hooks": {
        "request_input": [
          {
            "name": "confirm_action",
            "command": "./bin/format_prompt '{prompt_text}'",
            "enabled": true,
            "timeout": 120,
            "return_as": "user_confirmation",
            "default_value": "no",
            "metadata": {
              "description": "Request user confirmation for actions",
              "example": "Deploy to production? (yes/no)"
            }
          },
          {
            "name": "get_priority",
            "command": "./bin/format_prompt 'Priority level? (low/medium/high)'",
            "enabled": false,
            "timeout": 60,
            "return_as": "priority",
            "default_value": "medium"
          }
        ]
      }
    }
  }
}
```

### Example Usage Flow

```
1. User: "Deploy the new version to production"

2. Agent analyzes and determines deployment needed

3. Agent calls request_input tool:
   request_input(prompt="Deploy v2.1.0 to production? This affects 1000 users.")

4. request_input hook executes:
   - Formats prompt via ./bin/format_prompt
   - Sends to user: "🤔 Deploy v2.1.0 to production? This affects 1000 users.\n\nPlease respond with your input."
   - Waits up to 120s

5. User responds: "yes, proceed"

6. Hook returns: "yes, proceed"

7. Agent continues: "Confirmed. Starting deployment..."

8. Agent executes deployment tools

9. Agent responds: "✅ Deployment completed successfully"
```

---

**Plan Version**: 1.0.0  
**Author**: Orchestrator Agent (following Enhanced Agent System protocol)  
**Reviewed By**: Pending  
**Approved By**: Pending  
**Implementation Status**: Planning Complete, Ready for Review
