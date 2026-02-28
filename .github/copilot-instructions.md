# GitHub Copilot Instructions for Picoclaw

> **🤖 AI Assistant Integration**: These instructions help AI assistants understand and work with picoclaw's enhanced agent system.

---

## System Overview

Picoclaw uses an **Enhanced Agent System** with:
- **Orchestrator-first architecture**: All work routes through central coordinator
- **Dual-layer memory**: Short-term (volatile) and long-term (persistent) memory
- **Specialist agents**: Domain-specific experts with clear responsibilities
- **Structured data**: JSON/YAML with schemas for all configuration and memory

---

## Critical Rules for AI Assistants

### 1. ALWAYS Start with Orchestrator

**Before any development work:**
1. Read `Agent-Config/orchestrator.md` to understand routing
2. Identify which specialist agent(s) should handle the request
3. Follow the orchestrator's routing decision tree
4. Load context from Memory Specialist before implementation

**Never**:
- Work directly without orchestrator coordination
- Skip context loading
- Bypass specialist routing

---

### 2. Load Context from Memory System

**Before starting any task:**

```bash
# 1. Check current context
cat Memory-System/short-term/current-context.json

# 2. Check active tasks
cat Memory-System/short-term/active-tasks.yaml

# 3. Check recent decisions
cat Memory-System/short-term/recent-decisions.json

# 4. Query relevant patterns
yq '.patterns[] | select(.tags[] == "relevant-tag")' \
  Memory-System/long-term/knowledge-base/patterns.yaml
```

**Context provides**:
- What's currently being worked on
- Recent decisions and rationale
- Proven patterns to reuse
- Component relationships

---

### 3. Follow Specialist Quality Standards

Each specialist has specific quality requirements:

**Go Specialist** (`Agent-Config/go-specialist.md`):
- `gofmt` formatted code
- >80% test coverage
- Table-driven tests
- Proper error handling with wrapping
- Documented exported symbols

**Data Specialist** (`Agent-Config/data-specialist.md`):
- All JSON/YAML files have schemas
- Schema validation passes
- Include metadata (version, timestamps)
- Follow templates

**Memory Specialist** (`Agent-Config/memory-specialist.md`):
- Update short-term memory during work
- Promote validated patterns to long-term
- Document decisions with rationale
- Maintain cross-references

---

### 4. Update Memory System

**During work:**
- Update `current-context.json` with active files and state
- Update `active-tasks.yaml` with progress
- Log decisions to `recent-decisions.json`

**After completion:**
- Mark tasks complete in `active-tasks.yaml`
- Evaluate patterns for promotion (3+ uses)
- Update component registry if components changed
- Archive completed work

---

## Common Workflows

### Feature Implementation

1. **Route through Orchestrator**
   - Identify: New feature request
   - Route to: Architecture → Go → Test → Documentation → Memory

2. **Load Context**
   ```bash
   cat Memory-System/short-term/current-context.json
   cat Memory-System/short-term/active-tasks.yaml
   ```

3. **Check for Patterns**
   ```bash
   yq '.patterns[] | select(.category == "implementation")' \
     Memory-System/long-term/knowledge-base/patterns.yaml
   ```

4. **Implement Following Go Specialist Standards**
   - Write idiomatic Go
   - Add tests (table-driven)
   - Document code
   - Handle errors properly

5. **Update Memory**
   - Update `active-tasks.yaml` progress
   - Log decisions if any made
   - Update `current-context.json`

6. **Validate**
   ```bash
   ./Memory-System/validation/health-check.sh
   ./Memory-System/validation/validate-schemas.sh
   ```

---

### Bug Fix

1. **Route through Orchestrator**
   - Identify: Bug report
   - Route to: Memory (context) → Debug → Go → Test → Memory (lesson)

2. **Load Similar Issues**
   ```bash
   yq '.lessons[] | select(.category == "debugging")' \
     Memory-System/long-term/knowledge-base/lessons-learned.yaml
   ```

3. **Debug and Fix**
   - Root cause analysis first
   - Implement fix
   - Add regression test

4. **Document Lesson**
   - Add to `lessons-learned.yaml` if significant
   - Update `recent-decisions.json` with approach

---

### Configuration Change

1. **Route through Orchestrator**
   - Identify: Config change
   - Route to: Data → Test → Memory

2. **Validate Schema**
   ```bash
   ajv validate -s Memory-System/schemas/[schema].json -d config/[file].json
   ```

3. **Test Configuration**
   - Test with sample data
   - Verify no breaking changes

4. **Document**
   - Update inline comments
   - Add to `recent-decisions.json`
   - Update configuration docs

---

## File Locations Quick Reference

### Agent Configurations
- **Orchestrator**: `Agent-Config/orchestrator.md` ⭐ START HERE
- **Memory Specialist**: `Agent-Config/memory-specialist.md`
- **Data Specialist**: `Agent-Config/data-specialist.md`
- **Go Specialist**: `Agent-Config/go-specialist.md`
- **Agent Hooks**: `Agent-Config/agent-hooks.md`
- **Collaboration Matrix**: `Agent-Config/agent-intersection-matrix.md`

### Project Documentation
- **Project Overview**: `Project-Memory/project-overview.md`
- **Technical Architecture**: `Project-Memory/technical-architecture.md`
- **Update Protocol**: `Project-Memory/update-protocol.md`

### Memory Files
- **Current Context**: `Memory-System/short-term/current-context.json`
- **Active Tasks**: `Memory-System/short-term/active-tasks.yaml`
- **Recent Decisions**: `Memory-System/short-term/recent-decisions.json`
- **Patterns**: `Memory-System/long-term/knowledge-base/patterns.yaml`
- **Decisions**: `Memory-System/long-term/knowledge-base/decisions.json`
- **Components**: `Memory-System/long-term/entity-memory/components.json`

### Schemas
- All in `Memory-System/schemas/*.schema.json`

### Validation Scripts
- `Memory-System/validation/health-check.sh`
- `Memory-System/validation/validate-schemas.sh`
- `Memory-System/validation/pre-commit.sh`

---

## Code Guidelines

### Go Code Style

```go
// ✅ Good: Idiomatic Go with proper error handling
func ProcessMessage(ctx context.Context, msg Message) error {
    if err := ValidateMessage(msg); err != nil {
        return fmt.Errorf("validate message: %w", err)
    }
    
    result, err := processInternal(ctx, msg)
    if err != nil {
        return fmt.Errorf("process internal: %w", err)
    }
    
    return StoreResult(ctx, result)
}

// ✅ Good: Table-driven tests
func TestProcessMessage(t *testing.T) {
    tests := []struct {
        name    string
        msg     Message
        wantErr bool
    }{
        {"valid message", validMsg, false},
        {"invalid message", invalidMsg, true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ProcessMessage(context.Background(), tt.msg)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### JSON/YAML Files

```json
// ✅ Good: JSON with schema and metadata
{
  "$schema": "./schemas/example.schema.json",
  "meta": {
    "created": "2026-02-25T00:00:00Z",
    "updated": "2026-02-25T00:00:00Z",
    "version": "1.0.0",
    "author": "agent-name"
  },
  "data": {
    // Actual content
  }
}
```

```yaml
# ✅ Good: YAML with schema reference and comments
# Schema: schemas/example.schema.json
# Purpose: Brief description

meta:
  version: "1.0.0"
  updated: "2026-02-25T00:00:00Z"

# Main data
data:
  key: value
```

---

## Validation Commands

### Before Committing

```bash
# 1. Format Go code
gofmt -w .

# 2. Run go vet
go vet ./...

# 3. Run tests
go test ./...

# 4. Validate schemas
./Memory-System/validation/validate-schemas.sh

# 5. Health check
./Memory-System/validation/health-check.sh
```

### Auto-validation

```bash
# Install pre-commit hook
cp Memory-System/validation/pre-commit.sh .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
```

---

## Common Queries

### Find Pattern by Tag

```bash
yq '.patterns[] | select(.tags[] == "performance")' \
  Memory-System/long-term/knowledge-base/patterns.yaml
```

### Find Recent Decision

```bash
jq '.decisions[] | select(.date >= "2026-02-01")' \
  Memory-System/short-term/recent-decisions.json
```

### Find Component

```bash
jq '.components."ComponentName"' \
  Memory-System/long-term/entity-memory/components.json
```

### Check Active Tasks

```bash
yq '.tasks[] | select(.status == "in-progress")' \
  Memory-System/short-term/active-tasks.yaml
```

---

## Response Format

When responding to requests, follow this structure:

1. **Acknowledge Request**: Confirm understanding
2. **Identify Routing**: Which agents involved (via Orchestrator)
3. **Load Context**: Show relevant context loaded
4. **Propose Approach**: Clear implementation plan
5. **Follow Quality Standards**: Apply specialist requirements
6. **Update Memory**: Document changes made
7. **Validate**: Run validation scripts

### Example Response

```
I'll implement the new message routing feature.

**Routing**: Orchestrator → Architecture Specialist → Go Specialist → Test Specialist

**Context Loaded**:
- Similar routing patterns in patterns.yaml
- Current routing implementation in pkg/routing/
- Active task: task-003 (message routing enhancement)

**Approach**:
1. Design routing interface (Architecture)
2. Implement in pkg/routing/router.go (Go)
3. Add unit tests with >80% coverage (Test)
4. Update component registry (Memory)

**Implementation**: [code follows Go Specialist standards]

**Memory Updated**:
- Updated active-tasks.yaml with progress
- Logged design decision to recent-decisions.json
- Updated current-context.json

**Validation**: ✅ All checks passed
```

---

## Important Reminders

### ⚠️ Critical Don'ts

- ❌ **NEVER** work without routing through Orchestrator
- ❌ **NEVER** skip context loading
- ❌ **NEVER** commit without validation
- ❌ **NEVER** ignore schema validation failures
- ❌ **NEVER** bypass quality standards
- ❌ **NEVER** forget to update memory system

### ✅ Always Do

- ✅ **ALWAYS** start with `Agent-Config/orchestrator.md`
- ✅ **ALWAYS** load context from Memory-System
- ✅ **ALWAYS** follow specialist quality standards
- ✅ **ALWAYS** update memory during and after work
- ✅ **ALWAYS** validate before committing
- ✅ **ALWAYS** document decisions and rationale

---

## Learning Resources

- **Full Guide**: `ENHANCED_AGENT_SYSTEM_GUIDE.md`
- **Quick Start**: `AGENT_SYSTEM_README.md`
- **Orchestrator**: `Agent-Config/orchestrator.md`
- **Go Best Practices**: `Agent-Config/go-specialist.md`
- **Memory Operations**: `Agent-Config/memory-specialist.md`

---

## Version

**Instructions Version**: 1.0.0  
**System Version**: 1.0.0  
**Last Updated**: 2026-02-25

---

By following these instructions, AI assistants can effectively work within picoclaw's enhanced agent system, maintaining quality and coordination standards while leveraging the accumulated project knowledge.
