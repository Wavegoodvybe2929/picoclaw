# PicoClaw Workspace Integration - Final Implementation Plan

## TL;DR

**Philosophy:** PicoClaw extends via scripts, not code. The `exec` tool runs anything. Hooks automate exec at lifecycle points.

**What's changing:** Adding ~340 lines to automate calling workspace scripts (memory, search, etc.) at key moments in the agent loop.

**User impact:** Write a script → agent can use it. Add a hook → agent uses it automatically. Zero Go knowledge required.

**Example:**
```bash
# 1. Write script (any language)
echo '#!/usr/bin/env python3' > workspace/bin/my_tool
chmod +x workspace/bin/my_tool

# 2. Agent uses it immediately
picoclaw agent
> exec("./bin/my_tool")

# 3. Or make it automatic via hook
{"before_llm": [{"command": "./bin/my_tool"}]}
```

**This already works.** This plan just adds the hook automation layer.

---

## Executive Summary

**Goal:** Make PicoClaw aware of workspace scripts it can already execute, and automate calling them at key lifecycle points.

**Core Insight:** PicoClaw's `exec` tool can already run any script. The workspace has 23+ Python tools. The gap isn't capability—it's automation and discoverability.

**Approach:** 
1. Hook system = automated exec calls at lifecycle points (before LLM, after response, etc.)
2. Template variable substitution (`{query}` → actual user message)
3. Output injection (exec output → LLM context)
4. Config-driven (user controls which scripts run when)

**Implementation:** ~340 lines of Go to:
- Parse hook configs
- Execute hooks via subprocess (same as exec tool)
- Inject outputs into agent context
- Zero breaking changes, all opt-in

**Result:** Write a script → agent can exec it. Add a hook → agent execs it automatically. Extend PicoClaw in any language without touching Go code.

**Follows:** [Implementation Guide](IMPLEMENTATION_GUIDE.md) principles - Simple, Small, Safe, Config-driven, Exec-first.

**Code Impact:** ~340 lines added, 0 lines removed

---

## Design Principles

✅ **Simple** - Thin wrapper around existing workspace tools  
✅ **Small** - Minimal code additions (~340 lines total)  
✅ **Safe** - No regressions, all changes additive  
✅ **Config-driven** - User controls everything via `~/.picoclaw/config.json`  
✅ **Exec-first** - The `exec` tool is the foundation for all extensibility  
✅ **Extensible** - Hook system = automated exec calls at lifecycle points

**Core Philosophy:** PicoClaw is extended by writing scripts, not code. The `exec` tool can run any script, and the hook system automates calling these scripts at key points in the agent loop. Want a new feature? Write a script. Want it automatic? Add a hook.

This follows the [Implementation Guide](IMPLEMENTATION_GUIDE.md) principles:
- **Simple over Complex** - Scripts, not reimplementation
- **Small over Large** - ~340 lines, not thousands
- **Config-driven** - Hooks configured in JSON, not hardcoded
- **Extensible over Specific** - Exec enables infinite extensions  

---

## The Exec Tool: Foundation of Extensibility

### What Is Exec?

The `exec` tool is PicoClaw's **primary extension mechanism**. It allows the agent to run any executable script or command.

**Existing exec tool capabilities:**
```go
// Agent can already do this:
exec("./bin/search", ["Go best practices"])  // Web search
exec("python", ["script.py", "--arg", "value"])  // Python scripts
exec("./workspace/bin/memory_recall", ["--query", "user preferences"])  // Custom tools
```

### Why Exec-First Design?

**Benefits:**
1. **Zero reimplementation** - Call existing tools instead of rewriting in Go
2. **Any language** - Python, Bash, Node.js, whatever you want
3. **Easy to extend** - Drop a new script in `workspace/bin/`, agent can use it
4. **Testable** - Scripts can be tested independently
5. **Shareable** - Share scripts between users, no Go knowledge needed

**Example Extension Flow:**
```bash
# User wants calendar integration
# 1. Write a script
echo '#!/usr/bin/env python3' > workspace/bin/calendar_today
chmod +x workspace/bin/calendar_today

# 2. Agent can immediately use it
picoclaw agent
> "What's on my calendar today?"
> Agent: exec("./bin/calendar_today")
> Returns: "Meeting at 2pm, Dentist at 4pm"
```

**No code changes needed.** This already works.

### What This Plan Adds: Automated Exec via Hooks

The hook system is **just automated exec calls** at specific lifecycle points:

```json
{
  "loop_hooks": {
    "before_llm": [
      {
        "command": "./bin/memory_recall --query '{query}'",  // <- Just exec
        "inject_as": "context"
      }
    ]
  }
}
```

**Translation:** "Before each LLM call, exec this command and inject the output."

**Why hooks?** To avoid the agent needing explicit instructions. Instead of:
```
User: "Remember to check my memory first"
Agent: <exec memory_recall> <thinks> <responds>
```

With hooks:
```
User: "What's my favorite color?"
Agent: <hook auto-execs memory_recall> <already knows> <responds>
```

### Extensibility Examples

**Without hooks (manual exec):**
```
User: "Search for Go tutorials"
Agent: exec("./bin/search", ["Go tutorials"])
```

**With hooks (automatic exec):**
```json
{
  "before_llm": [
    {"command": "./bin/memory_recall --query '{query}'"},  // Auto-recall
    {"command": "./bin/calendar_context"}                    // Auto-calendar
  ]
}
```

User: "What should I do today?"
→ Hook 1 execs `memory_recall` (remembers past todos)
→ Hook 2 execs `calendar_context` (gets today's events)
→ Agent has full context automatically

**The pattern:** Exec is always available. Hooks just automate when exec runs.

---

## Current State Analysis

### Workspace (Already Complete)
- ✅ Memory system (append-only log + SQLite index)
- ✅ Calendar management (CSV-based)
- ✅ Vault/notes (Obsidian-compatible)
- ✅ Search integration (SearXNG)
- ✅ Research pipeline
- ✅ Gmail integration
- ✅ 23 Python CLI tools in `workspace/bin/`

### PicoClaw Binary (Existing)
- ✅ Agent loop + LLM providers
- ✅ **Tool system with exec** (the extensibility foundation)
- ✅ Other tools: read_file, write_file, etc.
- ✅ Session management
- ✅ Gateway server
- ✅ Skills system
- ✅ Built-in web search tools (Brave/Tavily/DuckDuckGo)

### Gap
- ❌ Agent **can** use workspace tools via exec, but doesn't know they exist
- ❌ Agent uses built-in web search instead of workspace search (also via exec)
- ❌ Memory system not **automatically** integrated (could be called via exec manually)
- ❌ No **automatic** memory context injection before LLM calls

**Key insight:** The exec tool can already do everything. We just need:
1. **Discoverability** - Tell agent about workspace scripts
2. **Automation** - Run certain execs automatically (hooks)
3. **Context injection** - Inject exec output into LLM context

---

## 5-Minute Extension Example

**Scenario:** You want PicoClaw to check your local database before answering questions.

### Traditional Approach (Other AI tools)
1. Learn the AI tool's SDK/API
2. Write integration code (50-200 lines)
3. Compile/build the project
4. Debug integration issues
5. Update when AI tool changes
**Time:** Hours to days

### PicoClaw Approach (Exec-First)

**Step 1: Write a script (2 minutes)**
```bash
cat > workspace/bin/check_database << 'EOF'
#!/usr/bin/env python3
import sys
import sqlite3

query = sys.argv[1]
conn = sqlite3.connect('~/mydata.db')
results = conn.execute("SELECT * FROM facts WHERE content LIKE ?", (f"%{query}%",))
for row in results:
    print(f"- {row[1]}")
EOF

chmod +x workspace/bin/check_database
```

**Step 2: Test it (30 seconds)**
```bash
./workspace/bin/check_database "customer preferences"
# Output:
# - Customer prefers morning meetings
# - Customer likes detailed reports
```

**Step 3: Make it automatic (2 minutes)**
Edit `~/.picoclaw/config.json`:
```json
{
  "loop_hooks": {
    "before_llm": [
      {
        "command": "./bin/check_database '{query}'",
        "inject_as": "context"
      }
    ]
  }
}
```

**Step 4: Use it (30 seconds)**
```bash
picoclaw agent
> "What does the customer prefer?"
> Agent: <hook auto-execs check_database>
>        <sees "Customer prefers morning meetings">
>        <responds with context>
```

**Total time:** ~5 minutes  
**Code in PicoClaw:** 0 lines  
**Compilation:** None  
**Language:** Your choice (Python, Bash, Node.js, Ruby, etc.)  

**This is the power of exec-first design.**

---

## What Changes (User Perspective)

### Before Integration
```bash
# User runs agent
picoclaw agent
> Agent uses built-in web search
> No memory context
> No workspace awareness
> Forgets previous conversations
```

### After Integration
```bash
# User runs agent (same command)
picoclaw agent
> BEFORE LLM: Recalls relevant context from memory
> Uses workspace search (./bin/search) instead of built-in
> AFTER RESPONSE: Stores conversation in memory
> Maintains context across sessions
> Can use calendar, vault, research tools via exec
```

### User Config (`~/.picoclaw/config.json`)
```json
{
  "agents": {
    "defaults": {
      "workspace": "~/.picoclaw/workspace",
      "restrict_to_workspace": true,
      "use_workspace_tools": true,
      "loop_hooks": {
        "before_llm": [
          {
            "name": "memory_recall",
            "command": "./bin/memory_recall --query '{query}' --format markdown",
            "enabled": true,
            "inject_as": "context"
          }
        ],
        "after_response": [
          {
            "name": "memory_write_user",
            "command": "./bin/memory_write --role user --content '{user_message}'",
            "enabled": true
          },
          {
            "name": "memory_write_assistant",
            "command": "./bin/memory_write --role assistant --content '{assistant_message}'",
            "enabled": true
          },
          {
            "name": "memory_sync",
            "command": "./bin/memory_sync",
            "enabled": true
          }
        ]
      }
    }
  }
}
```

---

## Implementation Plan

### Phase 1: Config Schema Extension
**Files Modified:**
- `pkg/config/config.go` (+30 lines)
- `pkg/config/defaults.go` (+25 lines)

**New Structures:**
```go
type LoopHook struct {
    Name     string            `json:"name"`
    Command  string            `json:"command"`
    Enabled  bool              `json:"enabled"`
    InjectAs string            `json:"inject_as,omitempty"` // "context", "" 
    Metadata map[string]string `json:"metadata,omitempty"`
}

type LoopHooks struct {
    BeforeLLM      []LoopHook `json:"before_llm,omitempty"`
    AfterResponse  []LoopHook `json:"after_response,omitempty"`
    OnToolCall     []LoopHook `json:"on_tool_call,omitempty"`
    OnError        []LoopHook `json:"on_error,omitempty"`
}

type AgentDefaults struct {
    Workspace           string
    RestrictToWorkspace bool
    // ... existing fields ...
    
    // NEW
    UseWorkspaceTools   bool      `json:"use_workspace_tools,omitempty"`
    LoopHooks           LoopHooks `json:"loop_hooks,omitempty"`
}
```

**Template Variables (for hook commands):**
- `{query}` - user's current message
- `{user_message}` - user message content
- `{assistant_message}` - assistant response
- `{session_key}` - current session ID
- `{channel}` - channel name
- `{chat_id}` - chat ID
- `{tool_name}` - tool being called
- `{error}` - error message

**How hooks work:**
1. User message arrives
2. Substitute template variables in hook command
3. **Execute the command via subprocess** (same as exec tool)
4. Capture output
5. Inject into context if `inject_as: "context"`

**This is just exec with automation** - no new execution model, just lifecycle integration.

---

### Phase 2: Hook Executor System
**New File:** `pkg/agent/hooks.go` (~150 lines)

**What it does:** Wraps the exec pattern with lifecycle integration.

**Responsibilities:**
- Execute hook commands (via subprocess, like exec tool)
- Template variable substitution (`{query}` → actual query)
- Handle workspace bin/ path resolution
- Activate Python venv for Python scripts (if detected)
- Capture output for context injection
- Shell-safe escaping

**Core insight:** This is NOT a new execution model. It's the same subprocess execution that exec uses, just with:
- Automatic triggering at lifecycle points
- Template variable substitution
- Output capture for injection

**Key Functions:**
```go
type HookExecutor struct {
    workspaceDir string
    pythonVenv   string
}

func NewHookExecutor(workspaceDir string) *HookExecutor

func (h *HookExecutor) ExecuteHooks(
    ctx context.Context,
    hooks []config.LoopHook,
    vars map[string]string,
) (map[string]string, error)
```

**Example Hook Execution:**
```go
// Hook config
{
  "command": "./bin/memory_recall --query '{query}' --format markdown",
  "inject_as": "context"
}

// Executes as (same as if agent called exec tool manually):
cd ~/.picoclaw/workspace
./bin/memory_recall --query 'user's message here' --format markdown

// Returns output for injection
```

**This is exec + automation** - the command execution is identical to the exec tool, just triggered automatically.

---

### Phase 3: Agent Loop Integration
**Files Modified:**
- `pkg/agent/loop.go` (+30 lines)
- `pkg/agent/context.go` (+5 lines)

**Integration Points in `runAgentLoop`:**

```go
func (al *AgentLoop) runAgentLoop(...) (string, error) {
    // Existing: Update tool contexts
    al.updateToolContexts(agent, opts.Channel, opts.ChatID)

    // Existing: Build initial messages
    var history []providers.Message
    var summary string
    if !opts.NoHistory {
        history = agent.Sessions.GetHistory(opts.SessionKey)
        summary = agent.Sessions.GetSummary(opts.SessionKey)
    }
    
    // NEW: Execute before_llm hooks
    hookVars := map[string]string{
        "query":        opts.UserMessage,
        "user_message": opts.UserMessage,
        "session_key":  opts.SessionKey,
        "channel":      opts.Channel,
        "chat_id":      opts.ChatID,
    }
    
    hookExecutor := NewHookExecutor(agent.Workspace)
    hookResults, _ := hookExecutor.ExecuteHooks(
        ctx,
        al.cfg.Agents.Defaults.LoopHooks.BeforeLLM,
        hookVars,
    )
    
    // NEW: Inject memory context
    contextFromHooks := hookResults["context"]
    
    messages := agent.ContextBuilder.BuildMessages(
        history,
        summary,
        opts.UserMessage,
        contextFromHooks, // Memory context injected here
        opts.Channel,
        opts.ChatID,
    )

    // Existing: Save user message
    agent.Sessions.AddMessage(opts.SessionKey, "user", opts.UserMessage)

    // Existing: Run LLM iteration
    finalContent, iteration, err := al.runLLMIteration(ctx, agent, messages, opts)
    if err != nil {
        // NEW: Execute on_error hooks
        errorVars := hookVars
        errorVars["error"] = err.Error()
        hookExecutor.ExecuteHooks(ctx, al.cfg.Agents.Defaults.LoopHooks.OnError, errorVars)
        return "", err
    }

    // Existing: Handle empty response
    if finalContent == "" {
        finalContent = opts.DefaultResponse
    }

    // Existing: Save assistant message
    agent.Sessions.AddMessage(opts.SessionKey, "assistant", finalContent)
    agent.Sessions.Save(opts.SessionKey)
    
    // NEW: Execute after_response hooks (memory write)
    hookVars["assistant_message"] = finalContent
    hookExecutor.ExecuteHooks(
        ctx,
        al.cfg.Agents.Defaults.LoopHooks.AfterResponse,
        hookVars,
    )

    // Existing: Summarization, bus publish, logging
    // ... rest of function unchanged ...
}
```

**Modified `BuildMessages` signature:**
```go
func (cb *ContextBuilder) BuildMessages(
    history []providers.Message,
    summary string,
    userMessage string,
    hookContext string, // NEW: injected memory context
    channel string,
    chatID string,
) []providers.Message
```

---

### Phase 4: Prefer Workspace Tools (via Exec)
**File Modified:** `pkg/agent/loop.go` (+3 lines)

**Philosophy:** When workspace tools exist, prefer them over built-ins.

**Why?** Workspace tools are:
- Customizable by the user
- Integrated with user's data (search history, preferences)
- Extensible via simple scripts

The agent can call them via exec:
```
Agent: exec("./bin/search", ["query"])  // Workspace search
```

Instead of:
```
Agent: web_search("query")  // Built-in Brave/Tavily
```

**Change:**
```go
func registerSharedTools(...) {
    for _, agentID := range registry.ListAgentIDs() {
        agent, ok := registry.GetAgent(agentID)
        if !ok {
            continue
        }

        // NEW: Skip web tools if workspace tools preferred
        if !cfg.Agents.Defaults.UseWorkspaceTools {
            if searchTool := tools.NewWebSearchTool(...) {
                agent.Tools.Register(searchTool)
            }
            agent.Tools.Register(tools.NewWebFetchToolWithProxy(...))
        }
        
        // Existing: Hardware, message, spawn tools (unchanged)
        agent.Tools.Register(tools.NewI2CTool())
        agent.Tools.Register(tools.NewSPITool())
        // ... rest unchanged ...
    }
}
```

**Result:** When `use_workspace_tools: true`, agent naturally uses `exec ./bin/search` instead of built-in web search.

**Key point:** The agent ALWAYS has exec. This just changes which tools are available for the LLM to choose from. The LLM will prefer workspace scripts when they're the only search option.

---

### Phase 5: Workspace Commands
**New File:** `cmd/picoclaw/cmd_workspace.go` (~100 lines)

**Commands:**
```bash
picoclaw workspace status    # Calls ./bin/verify_setup
picoclaw workspace verify    # Calls ./bin/verify_setup --verify
picoclaw workspace tools     # Lists all bin/* tools
picoclaw workspace memory    # Shows memory system status
```

**Implementation:**
```go
func workspaceCmd() {
    if len(os.Args) < 3 {
        workspaceHelp()
        return
    }
    
    cfg, err := loadConfig()
    if err != nil {
        fmt.Printf("Error loading config: %v\n", err)
        os.Exit(1)
    }
    
    workspace := cfg.WorkspacePath()
    subcommand := os.Args[2]
    
    switch subcommand {
    case "status":
        workspaceStatusCmd(workspace)
    case "verify":
        workspaceVerifyCmd(workspace)
    case "tools":
        workspaceToolsCmd(workspace)
    case "memory":
        workspaceMemoryCmd(workspace)
    default:
        fmt.Printf("Unknown workspace command: %s\n", subcommand)
        workspaceHelp()
    }
}
```

**Modified:** `cmd/picoclaw/main.go` (+3 lines)
```go
switch command {
case "onboard":
    onboard()
case "agent":
    agentCmd()
// ... existing cases ...
case "workspace": // NEW
    workspaceCmd()
case "version":
    printVersion()
}
```

---

### Phase 6: Enhanced Onboarding
**File Modified:** `cmd/picoclaw/cmd_onboard.go` (~20 lines)

**Enhancement:** Write config.json with hooks pre-configured

```go
func onboard() {
    // ... existing onboarding ...
    
    // Create config with memory hooks enabled by default
    cfg := config.DefaultConfig()
    cfg.Providers.OpenAI.BaseURL = "http://localhost:1234/v1" // or user input
    
    configPath := getConfigPath()
    if err := cfg.SaveToPath(configPath); err != nil {
        fmt.Printf("Error saving config: %v\n", err)
        os.Exit(1)
    }
    
    fmt.Println("✓ Configuration created:", configPath)
    fmt.Println("✓ Memory hooks enabled by default")
    fmt.Println("  Edit config.json to customize loop hooks")
}
```

---

## Code Changes Summary

### New Files (2 files, ~250 lines)
1. `pkg/agent/hooks.go` - Hook executor system (~150 lines)
2. `cmd/picoclaw/cmd_workspace.go` - Workspace commands (~100 lines)

### Modified Files (5 files, ~90 lines)
1. `pkg/config/config.go` - Add hook structs (+30 lines)
2. `pkg/config/defaults.go` - Default hooks config (+25 lines)
3. `pkg/agent/loop.go` - Execute hooks in loop (+30 lines)
4. `pkg/agent/context.go` - Accept hook context param (+5 lines)
5. `cmd/picoclaw/main.go` - Add workspace case (+3 lines)

### Enhanced Files (1 file, ~20 lines)
1. `cmd/picoclaw/cmd_onboard.go` - Write hooks to config (+20 lines)

### **Total: ~340 lines added, 0 lines removed**

---

## Safety Guarantees

### ✅ No Regressions
- All existing commands work unchanged
- All existing tools work unchanged
- All existing sessions preserved
- All existing configs compatible

### ✅ No Breaking Changes
- Everything is opt-in
- Default behavior: hooks disabled if workspace doesn't exist
- Graceful fallback if hook commands fail
- Users without workspace: no errors, features just disabled

### ✅ No New Dependencies
- Uses existing Go standard library
- Uses existing subprocess execution
- No new third-party packages
- No new Python dependencies

### ✅ No Concurrency Issues
- Agent loop remains single-threaded
- Hooks execute synchronously
- Workspace tools handle own locking
- No new race conditions

### ✅ No File Conflicts
- Go binary calls workspace tools via subprocess
- No direct file access to memory DB or calendar CSV
- Python tools maintain exclusive control
- No duplicate implementations

---

## Agent Behavior Changes

### Memory Integration (Automated via Hooks)

**Reminder:** These are all just **exec calls** that happen automatically.

**Before each LLM call:**
1. Execute `before_llm` hooks
2. **Exec:** `./bin/memory_recall --query '{user_message}'` (subprocess)
3. Inject results into system prompt as context
4. LLM sees relevant past conversations and preferences

**After each response:**
1. Execute `after_response` hooks
2. **Exec:** `./bin/memory_write --role user --content '{user_message}'` (subprocess)
3. **Exec:** `./bin/memory_write --role assistant --content '{response}'` (subprocess)
4. **Exec:** `./bin/memory_sync` (subprocess)
5. Memory system stores and indexes conversation

**Without hooks,** the agent could do this manually:
```
Agent: <decides to check memory>
Agent: exec("./bin/memory_recall", ["--query", "user preferences"])
Agent: <reads result> <thinks> <responds>
```

**With hooks,** it's automatic:
```
Agent: <hook auto-execs memory_recall> <context injected> <responds>
```

### Web Search via Exec

**Before:**
```
User: "Search for Go best practices"
Agent uses: Brave/Tavily built-in search tool (Go implementation)
```

**After (with use_workspace_tools: true):**
```
User: "Search for Go best practices"
Agent uses: exec("./bin/search", ["Go best practices"])  // Calls workspace script
Gets: JSON results from workspace SearXNG
```

**Why better?** The workspace search:
- Uses your configured SearXNG instance
- Can be customized by editing `./bin/search` script
- Integrates with search history
- No code changes needed to modify behavior

### Other Workspace Tools Available (All via Exec)

Agent can use any workspace script:

**Manual exec (agent decides when):**
```
User: "Add meeting to calendar"
Agent: exec("./bin/calendar_add_event", ["Meeting", "2024-03-15", "14:00"])
```

**Available tools:**
- `./bin/calendar_add_event` - Add calendar events
- `./bin/vault_new_note` - Create notes
- `./bin/research_links` - Research topics
- `./bin/search_save_note` - Search and save
- `./bin/gmail_send` - Send email
- Any custom script you add

**Agent learns these from:**
- `workspace/AGENTS.md` - Loaded via context builder
- `workspace/TOOLS.md` - Tool documentation
- Auto-discovery: List `workspace/bin/*` files

**Extending PicoClaw = Writing Scripts:**
```bash
# Want weather integration?
echo '#!/usr/bin/env python3' > workspace/bin/weather
echo 'import requests; print(requests.get("...").json())' >> workspace/bin/weather
chmod +x workspace/bin/weather

# Agent can now use it:
User: "What's the weather?"
Agent: exec("./bin/weather", ["San Francisco"])
```

**No Go code. No compilation. Just write a script.**

---

## User Configuration Examples

### Minimal (Memory Only)
```json
{
  "agents": {
    "defaults": {
      "workspace": "~/.picoclaw/workspace",
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
            "name": "memory_write_user",
            "command": "./bin/memory_write --role user --content '{user_message}'",
            "enabled": true
          },
          {
            "name": "memory_write_assistant",
            "command": "./bin/memory_write --role assistant --content '{assistant_message}'",
            "enabled": true
          },
          {
            "name": "memory_sync",
            "command": "./bin/memory_sync",
            "enabled": true
          }
        ]
      }
    }
  }
}
```

### Extended (Calendar + Memory)
```json
{
  "agents": {
    "defaults": {
      "workspace": "~/.picoclaw/workspace",
      "use_workspace_tools": true,
      "loop_hooks": {
        "before_llm": [
          {
            "name": "memory_recall",
            "command": "./bin/memory_recall --query '{query}'",
            "enabled": true,
            "inject_as": "context"
          },
          {
            "name": "calendar_context",
            "command": "./bin/calendar_list_today",
            "enabled": true,
            "inject_as": "context"
          }
        ],
        "after_response": [
          {
            "name": "memory_write_user",
            "command": "./bin/memory_write --role user --content '{user_message}'",
            "enabled": true
          },
          {
            "name": "memory_write_assistant",
            "command": "./bin/memory_write --role assistant --content '{assistant_message}'",
            "enabled": true
          },
          {
            "name": "memory_sync",
            "command": "./bin/memory_sync",
            "enabled": true
          }
        ]
      }
    }
  }
}
```

### Disable Memory (Override Default)
```json
{
  "agents": {
    "defaults": {
      "workspace": "~/.picoclaw/workspace",
      "use_workspace_tools": false,
      "loop_hooks": {
        "before_llm": [],
        "after_response": []
      }
    }
  }
}
```

---

## Implementation Timeline

### Week 1: Foundation (Phases 1-2)
- **Day 1-2:** Config schema extension
- **Day 3-5:** Hook executor implementation

### Week 2: Integration (Phases 3-4)
- **Day 1-3:** Agent loop integration
- **Day 4-5:** Disable web tools, testing

### Week 3: Polish (Phases 5-6)
- **Day 1-2:** Workspace commands
- **Day 3:** Enhanced onboarding
- **Day 4-5:** Documentation and testing

---

## Testing Strategy

### Unit Tests
- Hook executor template substitution
- Shell escaping safety (prevent injection)
- Config parsing with hooks
- Subprocess execution patterns (match existing exec tool)

### Integration Tests
- Memory recall before LLM call (exec via hook)
- Memory write after response (exec via hook)
- Hook failure handling (graceful degradation)
- Workspace script execution (Python, Bash, etc.)

### End-to-End Tests
1. Fresh install → onboard → config has hooks
2. Agent conversation → memory stores properly (via exec)
3. Second conversation → recalls context from first (via exec)
4. Disable hooks → agent works without memory
5. Custom hooks → execute correctly (any script)
6. Manual exec → agent can call workspace scripts directly

---

## Documentation Updates

### User Documentation
1. **README.md** - Add "Extending PicoClaw" section emphasizing exec-first approach
2. **QUICKSTART.md** - Show 5-minute extension example
3. **EXTENDING.md** - New doc: How to write scripts and use hooks
4. **CONFIG_REFERENCE.md** - Document loop_hooks schema

**Key messages:**
- PicoClaw extends via scripts, not code
- Exec tool is the foundation
- Hooks automate exec calls
- Any language works

### Developer Documentation
1. **CONTRIBUTING.md** - Explain hook executor implementation
2. **ARCHITECTURE.md** - Document exec-first philosophy
3. **TESTING.md** - Subprocess testing patterns
4. **IMPLEMENTATION_GUIDE.md** - Already exists, reference it

---

## Success Criteria

### Functional Requirements
✅ Agent stores every conversation in memory (via exec hooks)  
✅ Agent recalls relevant context automatically (via exec hooks)  
✅ Agent uses workspace search instead of built-in (via exec)  
✅ Users can extend by writing scripts in any language  
✅ Hooks are just automated exec calls (no new execution model)  
✅ Users can customize hooks via config  
✅ Users can disable hooks if needed  

### Technical Requirements
✅ Zero regressions in existing functionality  
✅ Total code addition under 400 lines  
✅ All tests pass  
✅ No new dependencies  
✅ Follows Implementation Guide principles  
✅ Subprocess execution is safe (shell escaping, timeouts)  

### User Experience
✅ 5-minute extensions possible (write script + optional hook)  
✅ Clear documentation with examples  
✅ Helpful error messages when hooks fail  
✅ Graceful degradation (no crashes if workspace missing)  

---

## Future Extensions (Post-Implementation)

**Core principle:** Every extension is just a script + optional hook.

### More Hook Types (More Automation Points)
- `on_tool_call` - Exec script after any tool call (e.g., log tool usage)
- `on_session_start` - Exec script when session starts (e.g., load user preferences)
- `on_session_end` - Exec script when session ends (e.g., save session summary)

**Example:**
```json
{
  "on_session_start": [
    {"command": "./bin/load_user_context"}
  ]
}
```

### More Workspace Scripts (User-Created)
- Calendar extraction: `./bin/extract_events_from_conversation`
- Task creation: `./bin/create_task`
- Daily briefing: `./bin/generate_briefing`
- Email handling: `./bin/process_email`

**No code changes needed** - just add scripts to `workspace/bin/`

### Community Hook Patterns
- Webhook notifications: `./bin/send_webhook`
- External API integrations: `./bin/call_api`
- Custom logging: `./bin/log_conversation`
- Slack notifications: `./bin/notify_slack`

**Sharing scripts:**
```bash
# User 1 creates script
cat > workspace/bin/my_tool << 'EOF'
#!/usr/bin/env python3
# Custom functionality
EOF

# User 2 copies script
wget https://github.com/user/scripts/blob/main/my_tool -O workspace/bin/my_tool
chmod +x workspace/bin/my_tool

# Both users can now use it
```

**Extension ecosystem** - users can share scripts, not Go code.

---

## Risks & Mitigations

### Risk: Hook commands fail
**Mitigation:** Silent failure, log warning, continue without context

### Risk: Memory tools not installed
**Mitigation:** Check workspace/bin/ exists, disable hooks gracefully

### Risk: Python venv not activated
**Mitigation:** Auto-detect venv path, activate before execution

### Risk: User config malformed
**Mitigation:** Validate on load, fall back to defaults

### Risk: Performance impact
**Mitigation:** Hooks run in parallel where possible, timeouts on long commands

---

## Final Summary

This plan delivers a **simple, exec-first integration** following the [Implementation Guide](IMPLEMENTATION_GUIDE.md):

### What It Does
- **Exec is the foundation** - All workspace functionality accessed via subprocess execution
- **Hooks = automated exec** - Config-driven automation of script execution at lifecycle points
- **Zero reimplementation** - No workspace logic duplicated in Go
- **Script-based extensibility** - Users extend by writing scripts, not code

### How It Works
1. Workspace has 23+ Python scripts in `workspace/bin/`
2. Agent already has `exec` tool to run any script
3. **New:** Hook system automates exec at key lifecycle points
4. **New:** Memory hooks enabled by default (via config)
5. Scripts output context, hooks inject it into LLM

### Code Changes
- **New:** ~340 lines (hook executor + config)
- **Modified:** 0 lines of existing functionality
- **Removed:** 0 lines
- **Dependencies:** 0 new packages

### Guarantees
✅ **Zero regressions** - All existing functionality unchanged
✅ **Opt-in** - Hooks disabled if workspace doesn't exist
✅ **Extensible** - Drop new script in `bin/`, agent can use it
✅ **Shareable** - Scripts can be shared between users
✅ **Simple** - Subprocess execution, template substitution, that's it

### The Result

PicoClaw becomes a **script-extensible AI assistant** that:
- Remembers conversations (via memory scripts + hooks)
- Learns user preferences (via memory recall)
- Uses custom tools (via workspace scripts + exec)
- Extends infinitely (via user scripts)
- Requires zero Go knowledge to extend

**Philosophy:** The best integration is the one that doesn't exist. We're not integrating workspace INTO PicoClaw - we're making PicoClaw AWARE of workspace tools it can already execute.

**Following Implementation Guide:**
- ✅ Simple over Complex - Scripts, not reimplementation
- ✅ Small over Large - 340 lines, not thousands
- ✅ Safe over Fast - Zero breaking changes
- ✅ Config-driven - User controls everything
- ✅ Extensible over Specific - Any script works
