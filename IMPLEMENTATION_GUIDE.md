# PicoClaw Implementation Guide

**A systematic approach to planning and implementing new features**

This document captures the methodology, principles, and patterns used in PicoClaw development. Follow this guide when planning any new implementation to ensure consistency with project standards.

---

## Core Philosophy

### The PicoClaw Way

1. **Simple over Complex** - Favor thin wrappers over reimplementation
2. **Small over Large** - Minimize code additions (<500 lines preferred)
3. **Safe over Fast** - Zero regressions, all changes additive
4. **Config-driven over Hardcoded** - User controls behavior via config.json
5. **Opt-in over Mandatory** - New features disabled by default (unless critical)
6. **Extensible over Specific** - Build systems that enable future features

---

## Planning Process

### Phase 1: Discovery & Understanding

**Before writing any code, thoroughly understand:**

#### 1.1 Current State
- What already exists in the codebase?
- What works well that shouldn't be changed?
- What are the existing patterns and conventions?
- Where are the integration points?

**Tools:**
```bash
# Explore the codebase
grep -r "pattern" pkg/
find . -name "*.go" | xargs grep "interface"
git log --oneline path/to/file.go

# Read existing implementations
cat pkg/agent/loop.go
cat pkg/config/config.go
```

#### 1.2 Requirements Gathering
Ask clarifying questions:
- What problem are we solving?
- Who is the user and what's their workflow?
- What are the must-haves vs nice-to-haves?
- Are there existing solutions we can leverage?
- What are the safety concerns?
- What are the performance implications?

#### 1.3 Gap Analysis
Identify what's missing:
- What functionality doesn't exist yet?
- What prevents the user from achieving their goal?
- What's the minimal change needed to bridge the gap?

---

### Phase 2: Design Constraints

**Every PicoClaw implementation must satisfy these constraints:**

#### 2.1 Zero Regressions
- ✅ All existing commands work unchanged
- ✅ All existing features work unchanged
- ✅ All existing tests pass
- ✅ All existing configs remain valid

**Test:** Can a user upgrade without changing anything and have it work exactly as before?

#### 2.2 No Breaking Changes
- ✅ New features are opt-in (via config)
- ✅ Default behavior matches previous version
- ✅ Graceful degradation if new features unavailable
- ✅ Clear migration path if breaking changes unavoidable

#### 2.3 Minimal Code Addition
Target: <500 lines of new code per feature

**Why:** 
- Easier to review
- Fewer bugs
- Easier to maintain
- Faster to implement

**How:**
- Leverage existing systems
- Use subprocess instead of reimplementing
- Prefer configuration over code
- Build thin wrappers, not new systems

#### 2.4 No New Dependencies
- ✅ Use Go standard library
- ✅ Use existing third-party packages already in go.mod
- ✅ No new external services required
- ❌ Avoid adding new dependencies unless critical

**Exception:** If a new dependency is truly needed, justify why existing solutions don't work.

#### 2.5 Config-Driven Behavior
User controls features via `~/.picoclaw/config.json`

**Pattern:**
```go
type FeatureConfig struct {
    Enabled bool   `json:"enabled"`
    // ... feature-specific settings
}

type AgentDefaults struct {
    // ... existing fields
    NewFeature FeatureConfig `json:"new_feature,omitempty"`
}
```

**Defaults:**
```go
func DefaultConfig() *Config {
    return &Config{
        Agents: AgentsConfig{
            Defaults: AgentDefaults{
                // New features disabled by default (unless critical)
                NewFeature: FeatureConfig{
                    Enabled: false,
                },
            },
        },
    }
}
```

---

### Phase 3: Architecture Design

#### 3.1 Integration Strategy

**Ask:** Where does this feature fit in the existing architecture?

**Options:**

**A. New Tool** (if feature is user-invocable)
```go
// pkg/tools/new_feature.go
type NewFeatureTool struct {
    config FeatureConfig
}

func (t *NewFeatureTool) Name() string { return "new_feature" }
func (t *NewFeatureTool) Execute(ctx context.Context, args map[string]any) *ToolResult
```

**B. Agent Loop Hook** (if feature affects processing)
```go
// Use existing hook system or extend it
"loop_hooks": {
  "before_llm": [...],
  "after_response": [...],
  "new_hook_point": [...]
}
```

**C. New Command** (if feature is CLI-invocable)
```go
// cmd/picoclaw/cmd_feature.go
func featureCmd() {
    // Implementation
}

// cmd/picoclaw/main.go
case "feature":
    featureCmd()
```

**D. Provider Extension** (if feature is LLM-related)
```go
// Extend existing provider interface
type Provider interface {
    // ... existing methods
    NewMethod() error  // Add to interface
}
```

#### 3.2 File Organization

**Standard structure:**
```
pkg/
  feature/           # New package for substantial features
    client.go        # Main implementation
    types.go         # Data structures
    feature_test.go  # Tests

cmd/picoclaw/
  cmd_feature.go     # CLI command (if needed)

config/
  config.go          # Add config structs (if needed)
  defaults.go        # Add defaults (if needed)
```

**Guideline:** Create new package only if >200 lines or reusable by multiple components.

#### 3.3 Error Handling Pattern

**PicoClaw standard:**
```go
// Return errors, don't panic
func NewFeature() (*Feature, error) {
    if err := validate(); err != nil {
        return nil, fmt.Errorf("failed to validate: %w", err)
    }
    return &Feature{}, nil
}

// Log and continue for non-critical errors
func (f *Feature) Process() {
    if err := f.tryOptional(); err != nil {
        logger.WarnCF("feature", "Optional step failed", 
            map[string]any{"error": err.Error()})
        // Continue processing
    }
}

// Graceful degradation
func (f *Feature) DoWork() string {
    result, err := f.tryMainWork()
    if err != nil {
        logger.ErrorCF("feature", "Main work failed, using fallback",
            map[string]any{"error": err.Error()})
        return f.fallback()
    }
    return result
}
```

---

### Phase 4: Implementation Planning

#### 4.1 Break Into Phases

**Each phase should be:**
- Independently testable
- Mergeable without breaking main
- Completable in 1-3 days

**Example breakdown:**
```
Phase 1: Config schema (Day 1)
  - Add structs to config.go
  - Add defaults
  - Test config loading

Phase 2: Core implementation (Day 2-3)
  - Create main feature file
  - Implement core logic
  - Unit tests

Phase 3: Integration (Day 4)
  - Wire into agent loop / tools / commands
  - Integration tests

Phase 4: Polish (Day 5)
  - Documentation
  - Examples
  - Edge case handling
```

#### 4.2 Code Size Budgeting

Before writing code, estimate:
```
New files:
  pkg/feature/client.go       ~150 lines
  cmd/picoclaw/cmd_feature.go ~100 lines

Modified files:
  pkg/config/config.go        +30 lines
  pkg/agent/loop.go           +20 lines
  cmd/picoclaw/main.go        +3 lines

Total: ~303 lines
```

**If estimate >500 lines:** Simplify the approach. Can you:
- Reuse existing code?
- Call external tools instead of implementing?
- Split into smaller features?
- Use configuration instead of code?

#### 4.3 Testing Strategy

**Required for every feature:**

```go
// Unit tests
func TestFeatureBasics(t *testing.T) { }
func TestFeatureEdgeCases(t *testing.T) { }
func TestFeatureErrors(t *testing.T) { }

// Integration tests
func TestFeatureIntegration(t *testing.T) { }

// End-to-end test (if user-facing)
func TestFeatureE2E(t *testing.T) { }
```

**Test coverage target:** >80% for new code

---

### Phase 5: Safety & Performance Analysis

#### 5.1 Concurrency Safety Checklist

**Questions to answer:**

- [ ] Does this introduce new goroutines?
  - If yes: Document lifecycle and cancellation
  - Use contexts for cancellation

- [ ] Does this share state between goroutines?
  - If yes: Use channels or mutexes
  - Prefer channels over mutexes

- [ ] Does this modify shared data structures?
  - If yes: Document locking strategy
  - Use sync.RWMutex if mostly reads

- [ ] Can this cause deadlocks?
  - If yes: Document lock acquisition order
  - Use defer for unlock

**PicoClaw pattern:** Most of the codebase is single-threaded (agent loop). Preserve this where possible.

#### 5.2 Performance Impact Checklist

- [ ] Does this add latency to the main loop?
  - If yes: Make async or add timeout
  
- [ ] Does this increase memory usage?
  - If yes: Quantify and document (prefer <10MB)
  
- [ ] Does this do I/O in the hot path?
  - If yes: Add caching or move to background

- [ ] Does this scale with user data?
  - If yes: Add limits and pagination

**Benchmark new features:**
```go
func BenchmarkFeature(b *testing.B) {
    for i := 0; i < b.N; i++ {
        feature.Process()
    }
}
```

#### 5.3 Security Checklist

- [ ] Does this execute user input?
  - If yes: Validate and sanitize
  - Use allowlists, not denylists

- [ ] Does this access files?
  - If yes: Respect workspace restrictions
  - Validate paths (no ../ escaping)

- [ ] Does this make network requests?
  - If yes: Validate URLs, use timeouts
  - Respect proxy settings

- [ ] Does this handle secrets?
  - If yes: Never log secrets
  - Use secure storage (keychain/vault)

**PicoClaw has workspace restrictions enabled by default. Maintain this.**

---

### Phase 6: Documentation Requirements

#### 6.1 Code Documentation

**Every new package:**
```go
// Package feature provides [brief description].
//
// [Longer description of purpose and usage]
//
// Example usage:
//   f := feature.New()
//   result, err := f.Process()
package feature
```

**Every public function:**
```go
// ProcessData processes the input data and returns results.
// 
// Returns an error if data is invalid or processing fails.
// Errors are logged but processing continues for optional steps.
func ProcessData(data string) (Result, error) {
```

**Every config field:**
```go
type FeatureConfig struct {
    // Enabled determines whether the feature is active.
    // Default: false
    Enabled bool `json:"enabled"`
    
    // Timeout in seconds for feature operations.
    // Default: 30
    Timeout int `json:"timeout,omitempty"`
}
```

#### 6.2 User Documentation

**Required updates:**

1. **README.md** - Add feature to appropriate section
2. **Config example** - Show feature configuration
3. **User guide** - Explain when and how to use
4. **Migration guide** - If changing behavior

**Template:**
```markdown
## New Feature

### Overview
[What it does and why users care]

### Configuration
[Example config.json with feature enabled]

### Usage
[Command examples or workflow]

### Troubleshooting
[Common issues and solutions]
```

#### 6.3 Developer Documentation

**For complex features, create:**
```
docs/
  design/
    feature-design.md        # Architecture decisions
  implementation/
    feature-implementation.md # Implementation details
```

---

## Implementation Checklist

Before starting implementation, ensure you can answer:

### Design Phase
- [ ] What problem does this solve?
- [ ] What's the minimal viable implementation?
- [ ] How does this fit existing architecture?
- [ ] What files need to be created/modified?
- [ ] What's the estimated code size? (<500 lines?)
- [ ] What are the integration points?
- [ ] What existing code can be reused?

### Safety Phase
- [ ] Does this maintain backward compatibility?
- [ ] Does this preserve existing behavior by default?
- [ ] What happens if the feature fails?
- [ ] What are the security implications?
- [ ] What are the concurrency implications?
- [ ] What are the performance implications?

### Implementation Phase
- [ ] Config schema defined
- [ ] Default config set (feature disabled by default unless critical)
- [ ] Core implementation complete
- [ ] Unit tests written (>80% coverage)
- [ ] Integration tests written
- [ ] Error handling implemented
- [ ] Logging added (use logger.InfoCF/WarnCF/ErrorCF)
- [ ] Graceful degradation verified

### Documentation Phase
- [ ] Code comments complete
- [ ] User documentation updated
- [ ] Config example provided
- [ ] Migration guide (if needed)
- [ ] CHANGELOG.md updated

### Testing Phase
- [ ] All tests pass
- [ ] Manual testing complete
- [ ] Edge cases verified
- [ ] Error cases verified
- [ ] Performance acceptable
- [ ] No regressions in existing features

---

## Common Patterns

### Pattern 1: Subprocess Wrapper

**When:** You want to use an external tool instead of reimplementing

```go
type ExternalTool struct {
    binPath string
    timeout time.Duration
}

func (e *ExternalTool) Execute(ctx context.Context, args []string) (string, error) {
    ctx, cancel := context.WithTimeout(ctx, e.timeout)
    defer cancel()
    
    cmd := exec.CommandContext(ctx, e.binPath, args...)
    output, err := cmd.CombinedOutput()
    if err != nil {
        return "", fmt.Errorf("command failed: %w", err)
    }
    return string(output), nil
}
```

**Benefits:**
- No reimplementation needed
- Leverage existing tools
- Easy to swap implementations

### Pattern 2: Config-Driven Feature

**When:** Feature behavior should be user-customizable

```go
type FeatureConfig struct {
    Enabled  bool              `json:"enabled"`
    Settings map[string]string `json:"settings,omitempty"`
}

type Feature struct {
    config FeatureConfig
}

func NewFeature(cfg FeatureConfig) *Feature {
    return &Feature{config: cfg}
}

func (f *Feature) Process() error {
    if !f.config.Enabled {
        return nil // Feature disabled, no-op
    }
    // Process...
}
```

### Pattern 3: Hook System

**When:** Multiple customization points needed

```go
type HookConfig struct {
    Name     string `json:"name"`
    Command  string `json:"command"`
    Enabled  bool   `json:"enabled"`
}

type HookExecutor struct {
    hooks []HookConfig
}

func (h *HookExecutor) Execute(ctx context.Context, hookPoint string, vars map[string]string) error {
    for _, hook := range h.hooks {
        if !hook.Enabled {
            continue
        }
        // Execute hook...
    }
    return nil
}
```

### Pattern 4: Tool Registration

**When:** Adding new agent capabilities

```go
// pkg/tools/feature.go
type FeatureTool struct {
    config FeatureConfig
}

func NewFeatureTool(cfg FeatureConfig) *FeatureTool {
    return &FeatureTool{config: cfg}
}

func (t *FeatureTool) Name() string { 
    return "feature_name" 
}

func (t *FeatureTool) Description() string {
    return "What this tool does"
}

func (t *FeatureTool) Parameters() map[string]any {
    return map[string]any{
        "type": "object",
        "properties": map[string]any{
            "param": map[string]any{
                "type": "string",
                "description": "Parameter description",
            },
        },
        "required": []string{"param"},
    }
}

func (t *FeatureTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
    // Implementation
    return SuccessResult(result)
}

// Register in pkg/agent/loop.go
agent.Tools.Register(tools.NewFeatureTool(cfg.Tools.Feature))
```

---

## Anti-Patterns (Avoid These)

### ❌ Anti-Pattern 1: Reimplementation

**Don't:**
```go
// Reimplementing search in Go
func (a *Agent) Search(query string) ([]Result, error) {
    // 500 lines of HTTP client code
    // JSON parsing
    // Rate limiting
    // Error handling
}
```

**Do:**
```go
// Call existing workspace tool
func (a *Agent) Search(query string) ([]Result, error) {
    cmd := exec.Command("./bin/search", query)
    output, err := cmd.Output()
    if err != nil {
        return nil, err
    }
    var results []Result
    json.Unmarshal(output, &results)
    return results, nil
}
```

### ❌ Anti-Pattern 2: Hardcoded Behavior

**Don't:**
```go
func (a *Agent) Process(msg string) {
    // Always do this
    a.storeMemory(msg)
    // No way to disable
}
```

**Do:**
```go
func (a *Agent) Process(msg string) {
    if a.config.EnableMemory {
        a.storeMemory(msg)
    }
}
```

### ❌ Anti-Pattern 3: Breaking Changes

**Don't:**
```go
// Changed function signature
func Process(msg string, newParam bool) error {
    // Breaks all existing callers
}
```

**Do:**
```go
// Add new function, keep old one
func ProcessWithOptions(msg string, opts ProcessOptions) error {
    // New implementation
}

// Deprecated: Use ProcessWithOptions
func Process(msg string) error {
    return ProcessWithOptions(msg, DefaultOptions)
}
```

### ❌ Anti-Pattern 4: Silent Failures

**Don't:**
```go
func (a *Agent) Execute() {
    err := a.criticalOperation()
    if err != nil {
        // Silently ignore
        return
    }
}
```

**Do:**
```go
func (a *Agent) Execute() error {
    if err := a.criticalOperation(); err != nil {
        logger.ErrorCF("agent", "Critical operation failed",
            map[string]any{"error": err.Error()})
        return fmt.Errorf("critical operation failed: %w", err)
    }
    return nil
}
```

---

## Review Criteria

Before submitting a PR, verify:

### Code Quality
- [ ] Follows Go conventions (gofmt, golint clean)
- [ ] No commented-out code
- [ ] No debug print statements
- [ ] Consistent naming with existing code
- [ ] Error messages are actionable

### Functionality
- [ ] Feature works as specified
- [ ] Edge cases handled
- [ ] Error cases handled
- [ ] Graceful degradation implemented
- [ ] No regressions in existing features

### Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing done
- [ ] Performance acceptable
- [ ] Security reviewed

### Documentation
- [ ] Code comments complete
- [ ] User documentation updated
- [ ] Examples provided
- [ ] Config documented

---

## Example: Planning a New Feature

Let's walk through planning a hypothetical "Auto-summary" feature:

### Step 1: Discovery
**Current state:** Agent processes messages but doesn't automatically summarize long conversations.

**Requirements:** User wants automatic summaries every N messages to keep context window manageable.

**Existing solutions:** Agent already has session management and summarization capability (used for long threads).

### Step 2: Design Constraints
- ✅ Must not break existing behavior
- ✅ Must be opt-in (disabled by default)
- ✅ Must be configurable (threshold, summary style)
- ✅ Must handle failures gracefully (continue without summary)
- ✅ Target: <200 lines of code

### Step 3: Architecture
**Integration point:** Agent loop, after message processing

**Config structure:**
```go
type AutoSummaryConfig struct {
    Enabled       bool   `json:"enabled"`
    MessageCount  int    `json:"message_count"`   // Summarize every N messages
    Style         string `json:"style"`           // "concise", "detailed"
}
```

**Implementation approach:**
- Add counter to session
- Check counter after each message
- Trigger existing summarization code
- Store summary in session

### Step 4: Code Plan
```
New code: 0 files, 0 lines (reuse existing summarization)

Modified:
  pkg/config/config.go          +10 lines (config struct)
  pkg/config/defaults.go        +5 lines (default config)
  pkg/session/session.go        +20 lines (message counter)
  pkg/agent/loop.go             +30 lines (auto-trigger logic)

Total: 65 lines
```

### Step 5: Safety Analysis
- No concurrency issues (single-threaded loop)
- No performance issues (reuses existing code)
- No security issues (internal feature)
- Graceful degradation: if summarization fails, continue normally

### Step 6: Implementation
```go
// pkg/config/config.go
type AutoSummaryConfig struct {
    Enabled      bool   `json:"enabled"`
    MessageCount int    `json:"message_count"`
    Style        string `json:"style,omitempty"`
}

// pkg/config/defaults.go
AutoSummary: AutoSummaryConfig{
    Enabled:      false,  // Opt-in
    MessageCount: 20,     // Reasonable default
    Style:        "concise",
},

// pkg/agent/loop.go - in runAgentLoop
if al.cfg.Agents.Defaults.AutoSummary.Enabled {
    count := agent.Sessions.GetMessageCount(opts.SessionKey)
    if count > 0 && count % al.cfg.Agents.Defaults.AutoSummary.MessageCount == 0 {
        al.summarizeSession(agent, opts.SessionKey, opts.Channel, opts.ChatID)
    }
}
```

**Result:** 65 lines, reuses existing code, opt-in, configurable, safe.

---

## Conclusion

The PicoClaw way is to:
1. **Understand deeply** before changing anything
2. **Simplify ruthlessly** - prefer small over large
3. **Reuse extensively** - leverage existing code and tools
4. **Configure flexibly** - let users control behavior
5. **Test thoroughly** - verify everything works
6. **Document completely** - explain what and why

**Follow this guide for every new feature to maintain PicoClaw's simplicity, safety, and quality.**
