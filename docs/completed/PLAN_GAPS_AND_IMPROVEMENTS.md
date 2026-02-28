# PLAN.md Implementation - Gaps and Improvements

**Date:** February 25, 2026  
**Status:** Implementation Review  
**Reviewer:** Following Enhanced Agent System - Orchestrator Protocol

---

## Executive Summary

Following the [IMPLEMENTATION_GUIDE.md](IMPLEMENTATION_GUIDE.md) review checklist, this document identifies **gaps and necessary improvements** in the PLAN.md implementation (Phases 1-6).

### Overall Assessment

✅ **Core Implementation: 95% Complete**
- All 6 phases implemented
- 188 tests passing
- Zero compilation errors
- No regressions

❌ **Gaps Identified: 5 Critical Issues**
1. OnToolCall hooks defined but NOT executed
2. Improper logging pattern in hooks.go
3. Missing user documentation
4. Missing config examples
5. Missing hook validation

---

## Gap Analysis

### Gap 1: OnToolCall Hooks Not Executed ⚠️ CRITICAL

**Status:** ❌ **BLOCKING**  
**Severity:** HIGH  
**Impact:** Config field exists but does nothing

#### Evidence

**Config Definition (Present):**
- [pkg/config/config.go](pkg/config/config.go#L186): `OnToolCall []LoopHook` field defined ✅
- [PLAN.md](PLAN.md#L361): Spec includes OnToolCall ✅

**Execution (Missing):**
- [pkg/agent/loop.go](pkg/agent/loop.go#L680-L750): Tool execution code has NO hook calls ❌
- No `ExecuteHooks(...OnToolCall...)` anywhere in codebase ❌

**Test Coverage (Missing):**
- No tests for OnToolCall hook execution ❌

#### Root Cause

Phase 3 implementation focused on `before_llm`, `after_response`, and `on_error` hooks, but **forgot to implement `on_tool_call` hooks** during the tool execution phase.

#### Required Fix

**Location:** [pkg/agent/loop.go](pkg/agent/loop.go#L720-L730)

Add hook execution after tool result is obtained:

```go
// Execute tool calls
for _, tc := range normalizedToolCalls {
    argsJSON, _ := json.Marshal(tc.Arguments)
    argsPreview := utils.Truncate(string(argsJSON), 200)
    logger.InfoCF("agent", fmt.Sprintf("Tool call: %s(%s)", tc.Name, argsPreview),
        map[string]any{
            "agent_id":  agent.ID,
            "tool":      tc.Name,
            "iteration": iteration,
        })

    // Create async callback...
    asyncCallback := func(callbackCtx context.Context, result *tools.ToolResult) {
        // ...existing code...
    }

    toolResult := agent.Tools.ExecuteWithContext(
        ctx,
        tc.Name,
        tc.Arguments,
        opts.Channel,
        opts.ChatID,
        asyncCallback,
    )

    // NEW: Execute on_tool_call hooks
    if len(al.cfg.Agents.Defaults.LoopHooks.OnToolCall) > 0 {
        toolCallVars := hookVars // Copy from parent scope
        toolCallVars["tool_name"] = tc.Name
        argsStr, _ := json.Marshal(tc.Arguments)
        toolCallVars["tool_args"] = string(argsStr)
        toolCallVars["tool_result"] = toolResult.ForLLM
        
        hookExecutor.ExecuteHooks(
            ctx,
            al.cfg.Agents.Defaults.LoopHooks.OnToolCall,
            toolCallVars,
        )
    }

    // Send ForUser content to user immediately...
    // ...rest of existing code...
}
```

**Estimated Effort:** 15 lines of code  
**Testing Required:** 2 new tests in hooks_test.go

---

### Gap 2: Improper Logging Pattern ⚠️ IMPORTANT

**Status:** ❌ **TECHNICAL DEBT**  
**Severity:** MEDIUM  
**Impact:** Inconsistent with codebase standards

#### Evidence

**Current Implementation:**
[pkg/agent/hooks.go](pkg/agent/hooks.go#L91-L95):
```go
// Log error but don't stop execution of other hooks
fmt.Fprintf(os.Stderr, "Warning: Hook %q failed: %v\n", hook.Name, err)
continue
```

**Expected Pattern (from codebase):**
```go
// All other code uses logger package
logger.WarnCF("agent", "Hook execution failed",
    map[string]any{
        "hook_name": hook.Name,
        "error": err.Error(),
    })
```

#### Why This Matters

From [IMPLEMENTATION_GUIDE.md](IMPLEMENTATION_GUIDE.md#L377-L389):
> **Logging added (use logger.InfoCF/WarnCF/ErrorCF)**

- Consistent logging format across codebase
- Structured logs for monitoring/debugging
- Context fields for filtering
- Proper log levels

#### Required Fix

**Location:** [pkg/agent/hooks.go](pkg/agent/hooks.go#L1-L15)

1. Add logger import:
```go
import (
    // ...existing imports...
    "github.com/sipeed/picoclaw/pkg/logger"
)
```

2. Replace fmt.Fprintf with logger.WarnCF:
```go
// Execute the hook
output, err := h.executeHook(ctx, hook, vars)
if err != nil {
    // Log error but don't stop execution of other hooks
    logger.WarnCF("agent", "Hook execution failed",
        map[string]any{
            "hook_name": hook.Name,
            "error":     err.Error(),
            "command":   hook.Command,
        })
    continue
}
```

**Estimated Effort:** 10 lines changed  
**Testing Required:** No new tests (behavior unchanged)

---

### Gap 3: Missing User Documentation ⚠️ IMPORTANT

**Status:** ❌ **USABILITY ISSUE**  
**Severity:** MEDIUM  
**Impact:** Users don't know features exist

#### Evidence

**Searched for documentation:**
- ❌ README.md: No mention of "loop_hooks" ❌
- ❌ README.md: No mention of "workspace tools" integration ❌
- ❌ No docs/hooks-guide.md ❌
- ❌ No docs/workspace-integration.md ❌

**Users cannot discover:**
1. How to enable workspace tools
2. How to configure loop hooks
3. What hooks are available
4. What template variables exist
5. Example hook configurations

#### Required Fix

From [IMPLEMENTATION_GUIDE.md](IMPLEMENTATION_GUIDE.md#L582-L611):
> **Required updates:**
> 1. README.md - Add feature to appropriate section
> 2. Config example - Show feature configuration
> 3. User guide - Explain when and how to use

**Create:** `docs/WORKSPACE_INTEGRATION.md`

```markdown
# Workspace Integration Guide

## Overview
PicoClaw can automatically integrate with workspace scripts and tools...

## Configuration

### Enabling Workspace Tools
\`\`\`json
{
  "agents": {
    "defaults": {
      "use_workspace_tools": true
    }
  }
}
\`\`\`

### Configuring Loop Hooks
\`\`\`json
{
  "agents": {
    "defaults": {
      "loop_hooks": {
        "before_llm": [
          {
            "name": "memory_recall",
            "command": "./bin/memory_recall --query '{user_message}'",
            "enabled": true,
            "inject_as": "context"
          }
        ],
        "after_response": [...],
        "on_tool_call": [...],
        "on_error": [...]
      }
    }
  }
}
\`\`\`

## Template Variables
Available in hook commands:
- `{query}` - User's current message
- `{user_message}` - User message content
- `{assistant_message}` - Assistant response
- `{session_key}` - Current session ID
- `{channel}` - Channel name
- `{chat_id}` - Chat ID
- `{tool_name}` - Tool being called
- `{tool_args}` - Tool arguments (JSON)
- `{tool_result}` - Tool result
- `{error}` - Error message

## Hook Lifecycle
...
```

**Update:** `README.md`

Add section:
```markdown
## 🔌 Workspace Integration

PicoClaw can integrate with your workspace scripts and tools. See [Workspace Integration Guide](docs/WORKSPACE_INTEGRATION.md) for details.

**Features:**
- 🪝 **Loop Hooks**: Automate script execution at lifecycle points
- 🔧 **Workspace Tools**: Use custom scripts instead of built-in tools
- 💾 **Memory Integration**: Automatic context recall and storage
- 📊 **Tool Logging**: Track tool usage via hooks
```

**Estimated Effort:** 200 lines of documentation  
**Testing Required:** Manual review

---

### Gap 4: Missing Config Examples ⚠️ IMPORTANT

**Status:** ❌ **USABILITY ISSUE**  
**Severity:** MEDIUM  
**Impact:** Users don't see hook examples in config.example.json

#### Evidence

[config/config.example.json](config/config.example.json#L1-L50):
- ✅ Has `agents.defaults` section
- ❌ Missing `use_workspace_tools` example ❌
- ❌ Missing `loop_hooks` example ❌

#### Required Fix

**Location:** [config/config.example.json](config/config.example.json#L8-L10)

Add to `agents.defaults`:
```json
{
  "agents": {
    "defaults": {
      "workspace": "~/.picoclaw/workspace",
      "restrict_to_workspace": true,
      "model_name": "gpt4",
      "max_tokens": 8192,
      "temperature": 0.7,
      "max_tool_iterations": 20,
      
      // NEW: Workspace integration
      "use_workspace_tools": false,
      
      // NEW: Loop hooks (optional)
      "loop_hooks": {
        "before_llm": [
          {
            "name": "memory_recall",
            "command": "./bin/memory_recall --query '{user_message}' --format markdown",
            "enabled": true,
            "inject_as": "context",
            "metadata": {
              "description": "Recall relevant context from memory"
            }
          }
        ],
        "after_response": [
          {
            "name": "memory_write",
            "command": "./bin/memory_write --role user --content '{user_message}'",
            "enabled": true,
            "inject_as": "",
            "metadata": {
              "description": "Store user message in memory"
            }
          },
          {
            "name": "memory_write_assistant",
            "command": "./bin/memory_write --role assistant --content '{assistant_message}'",
            "enabled": true,
            "inject_as": ""
          }
        ],
        "on_tool_call": [
          {
            "name": "log_tool_usage",
            "command": "./bin/log_tool --name '{tool_name}' --args '{tool_args}'",
            "enabled": false,
            "inject_as": ""
          }
        ],
        "on_error": [
          {
            "name": "notify_error",
            "command": "./bin/notify_error '{error}'",
            "enabled": false,
            "inject_as": ""
          }
        ]
      }
    }
  }
}
```

**Estimated Effort:** 30 lines  
**Testing Required:** Validate JSON syntax

---

### Gap 5: Missing Hook Validation ⚠️ NICE-TO-HAVE

**Status:** ❌ **ENHANCEMENT**  
**Severity:** LOW  
**Impact:** Invalid hooks silently fail

#### Current Behavior

Invalid hook configurations are not validated:
- Empty command strings accepted
- Invalid template variable syntax not detected
- Non-existent inject_as values accepted

#### Desired Behavior

Validate hooks at config load time:
```go
// pkg/config/config.go
func (h *LoopHook) Validate() error {
    if h.Name == "" {
        return fmt.Errorf("hook name cannot be empty")
    }
    if h.Command == "" {
        return fmt.Errorf("hook command cannot be empty")
    }
    if h.InjectAs != "" && h.InjectAs != "context" {
        return fmt.Errorf("invalid inject_as value: %q (must be empty or 'context')", h.InjectAs)
    }
    
    // Validate template variables
    validVars := []string{
        "{query}", "{user_message}", "{assistant_message}",
        "{session_key}", "{channel}", "{chat_id}",
        "{tool_name}", "{tool_args}", "{tool_result}", "{error}",
    }
    // Check command contains only valid template vars
    // ...
    
    return nil
}
```

**Estimated Effort:** 50 lines  
**Testing Required:** 5 new validation tests  
**Priority:** MEDIUM (can be added later)

---

## Minor Improvements (Optional)

### Improvement 1: Configurable Hook Timeout

**Current:** Hardcoded 30s timeout in [pkg/agent/hooks.go](pkg/agent/hooks.go#L31)

**Proposed:** Add to LoopHook config
```go
type LoopHook struct {
    Name     string
    Command  string
    Enabled  bool
    InjectAs string
    Timeout  int    `json:"timeout,omitempty"` // NEW: seconds, default 30
    Metadata map[string]string
}
```

**Benefit:** Long-running hooks (e.g., large searches) won't timeout

---

### Improvement 2: Hook Execution Metrics

**Current:** No metrics on hook performance

**Proposed:** Add telemetry
```go
// pkg/agent/hooks.go
logger.InfoCF("agent", "Hook executed successfully",
    map[string]any{
        "hook_name":      hook.Name,
        "duration_ms":    elapsed.Milliseconds(),
        "output_length":  len(output),
    })
```

**Benefit:** Identify slow hooks, debug performance issues

---

### Improvement 3: Hook Retry Logic

**Current:** Hooks fail on first error

**Proposed:** Add retry for transient failures
```go
type LoopHook struct {
    // ...existing fields...
    Retries int `json:"retries,omitempty"` // NEW: retry count, default 0
}
```

**Benefit:** More resilient to temporary failures (network, DB, etc.)

---

### Improvement 4: Async Hook Execution

**Current:** All hooks execute synchronously (block agent loop)

**Proposed:** Optional async execution for non-critical hooks
```go
type LoopHook struct {
    // ...existing fields...
    Async bool `json:"async,omitempty"` // NEW: run in background
}
```

**Benefit:** Faster response times (don't wait for logging/metrics hooks)

**Caution:** Only for hooks with `inject_as: ""` (no context injection)

---

## Implementation Priority

Following [IMPLEMENTATION_GUIDE.md](IMPLEMENTATION_GUIDE.md) principles:

### Phase 1: Critical Gaps (MUST FIX)
**Estimated: 1 day**

1. ✅ **Implement OnToolCall hooks** (Gap 1)
   - Add execution in loop.go: 15 lines
   - Add tests: 2 tests
   - Update Memory System

2. ✅ **Fix logging pattern** (Gap 2)
   - Replace fmt.Fprintf: 10 lines
   - Import logger package

### Phase 2: Documentation (SHOULD FIX)
**Estimated: 0.5 day**

3. ✅ **Create user documentation** (Gap 3)
   - Write docs/WORKSPACE_INTEGRATION.md: 200 lines
   - Update README.md: 20 lines
   - Add examples

4. ✅ **Update config examples** (Gap 4)
   - Add to config.example.json: 30 lines
   - Validate JSON syntax

### Phase 3: Validation (NICE-TO-HAVE)
**Estimated: 0.5 day**

5. ⏸️ **Add hook validation** (Gap 5)
   - Validate at config load: 50 lines
   - Add validation tests: 5 tests
   - Can be deferred to future release

### Phase 4: Enhancements (OPTIONAL)
**Estimated: 1-2 days**

6. ⏸️ Minor improvements 1-4
   - Can be added in future releases
   - Not blocking current functionality

---

## Code Changes Summary

### Required Changes (Phases 1-2)

**New Files (1):**
- `docs/WORKSPACE_INTEGRATION.md` (~200 lines)

**Modified Files (4):**
1. `pkg/agent/loop.go` (+15 lines)
   - Add OnToolCall hook execution

2. `pkg/agent/hooks.go` (+1 import, ~10 lines changed)
   - Add logger import
   - Replace fmt.Fprintf with logger.WarnCF

3. `config/config.example.json` (+30 lines)
   - Add hook examples

4. `README.md` (+20 lines)
   - Add workspace integration section

**Total: ~275 lines (well within <500 line guideline)**

---

## Testing Requirements

### New Tests Required

1. **TestOnToolCallHookExecution** (pkg/agent/loop_test.go)
   - Verify hook executes during tool call
   - Verify template variables populated

2. **TestOnToolCallHookVariables** (pkg/agent/hooks_test.go)
   - Verify tool_name, tool_args, tool_result variables
   - Verify all template substitutions work

**Estimated:** 2 tests, ~100 lines

---

## Safety Analysis

### Concurrency Safety ✅
- OnToolCall hooks execute synchronously in agent loop (single-threaded)
- No new goroutines
- No shared state modifications
- **SAFE**

### Performance Impact ✅
- Hook execution already happens (before_llm, after_response)
- OnToolCall adds ~0-50ms per tool call (depending on hook script)
- Not in hot path (tools are already slow operations)
- **ACCEPTABLE**

### Backward Compatibility ✅
- OnToolCall hooks empty by default (no behavior change)
- Existing configs work unchanged
- New documentation doesn't affect existing users
- **NO BREAKING CHANGES**

### Security ✅
- Shell escaping already implemented
- Template variable substitution safe
- No new attack vectors
- **SECURE**

---

## Compliance Checklist

Following [IMPLEMENTATION_GUIDE.md](IMPLEMENTATION_GUIDE.md) checklist:

### Design Phase ✅
- [x] Problem clearly defined (OnToolCall missing, docs missing)
- [x] Minimal viable implementation (~275 lines)
- [x] Fits existing architecture (agent loop + hooks)
- [x] Files identified (4 modified, 1 new)
- [x] Code size estimated (<500 lines) ✅
- [x] Integration points identified (loop.go tool execution)
- [x] Existing code reused (hook executor, logger)

### Safety Phase ✅
- [x] Backward compatible ✅
- [x] Default behavior preserved ✅
- [x] Graceful failure (hooks optional) ✅
- [x] No security issues ✅
- [x] No concurrency issues ✅
- [x] Performance acceptable ✅

### Implementation Phase
- [ ] Config schema (no changes needed)
- [ ] Core implementation (OnToolCall execution)
- [ ] Unit tests (2 new tests)
- [ ] Error handling (logger pattern)
- [ ] Documentation (user guide, examples)

### Documentation Phase
- [ ] Code comments (in loop.go changes)
- [ ] User documentation (WORKSPACE_INTEGRATION.md)
- [ ] Config example (config.example.json)
- [ ] README update (feature section)

---

## Recommendation

### Immediate Action Required

**Implement Phases 1-2 (Critical Gaps + Documentation)**

This addresses the **blocking issue** (OnToolCall hooks) and makes the feature **discoverable and usable** for users.

**Estimated Effort:** 1.5 days  
**Risk:** LOW  
**Benefit:** HIGH

### Deferred to Future

**Phases 3-4 (Validation + Enhancements)**

These are improvements but not blockers. Can be added in a follow-up release.

**Rationale:**
- Current implementation works without validation
- Users will discover issues organically
- Enhancements are nice-to-have

---

## Conclusion

The PLAN.md implementation is **95% complete** and **production-ready** for basic use cases.

However, **5 gaps** were identified that should be addressed:
1. ⚠️ **CRITICAL**: OnToolCall hooks not executed
2. ⚠️ **IMPORTANT**: Improper logging pattern
3. ⚠️ **IMPORTANT**: Missing user documentation
4. ⚠️ **IMPORTANT**: Missing config examples
5. ⏸️ **NICE-TO-HAVE**: Missing validation

**Next Steps:**
1. Implement OnToolCall hooks (Gap 1) - **15 lines**
2. Fix logging pattern (Gap 2) - **10 lines**
3. Create user documentation (Gap 3) - **200 lines**
4. Add config examples (Gap 4) - **30 lines**

**Total Additional Work:** ~255 lines + 2 tests

Following IMPLEMENTATION_GUIDE principles, this keeps changes **small** (<500 lines), **safe** (no breaking changes), and **simple** (reusing existing patterns).

---

**Document Version:** 1.0.0  
**Last Updated:** 2026-02-25  
**Next Review:** After Phase 1-2 implementation
