# Agent Hooks - Automated Integration System

> **🔄 AUTOMATED AGENT INTEGRATION**: This document defines hooks that automatically trigger agent coordination at key development lifecycle points.

---

## Overview

Agent hooks are automated triggers that ensure proper agent coordination throughout the development lifecycle. They prevent forgotten steps, enforce quality gates, and maintain memory system consistency.

---

## Hook Types

### Pre-Task Hooks
Triggered **before** starting any development task

### During-Task Hooks
Triggered **during** active development work

### Post-Task Hooks
Triggered **after** completing development task

### Validation Hooks
Triggered at quality gate checkpoints

### Memory Hooks
Triggered for memory system operations

---

## Pre-Task Hooks

### Hook: Memory Context Loading

**Trigger**: Start of any development task  
**Responsible**: Memory Specialist  
**Priority**: CRITICAL - Must complete before work begins

**Actions**:
1. Load `Memory-System/short-term/current-context.json`
2. Load `Memory-System/short-term/active-tasks.yaml` for this task
3. Query `Memory-System/long-term/knowledge-base/patterns.yaml` for relevant patterns
4. Query `Memory-System/long-term/entity-memory/components.json` for affected components
5. Provide consolidated context to Orchestrator
6. Orchestrator routes context to assigned specialist

**Output**: Context package with all relevant information

**Validation**: Context loaded successfully, no missing required information

---

### Hook: Conflict Check

**Trigger**: Before routing task to specialist  
**Responsible**: Orchestrator  
**Priority**: HIGH - Prevents conflicts

**Actions**:
1. Parse `Memory-System/short-term/active-tasks.yaml`
2. Identify tasks with status "in-progress"
3. Check for file conflicts (same files being modified)
4. Check for dependency conflicts (required components being changed)
5. Check for resource conflicts (competing for same resources)
6. If conflict detected:
   - Queue task as dependency, OR
   - Request user guidance on priority, OR
   - Coordinate parallel work if safe

**Output**: Conflict resolution decision

**Validation**: No uncoordinated conflicts

---

### Hook: Quality Standards Load

**Trigger**: Before specialist begins implementation  
**Responsible**: Orchestrator  
**Priority**: MEDIUM - Ensures awareness of standards

**Actions**:
1. Load specialist's quality standards from agent config
2. Load domain-specific best practices from long-term memory
3. Load relevant patterns from pattern library
4. Provide standards package to specialist

**Output**: Quality standards and expectations document

**Validation**: Specialist acknowledges standards

---

## During-Task Hooks

### Hook: Context Update

**Trigger**: File modifications, decisions made, status changes  
**Responsible**: Memory Specialist  
**Priority**: MEDIUM - Keep context current

**Actions**:
1. Update `Memory-System/short-term/current-context.json`:
   - Add modified files to `recent_files`
   - Update `conversation_summary`
   - Adjust `active_agents` list
2. Update `Memory-System/short-term/active-tasks.yaml`:
   - Change task status if needed
   - Add progress notes
   - Update timestamp
3. Log decisions to `Memory-System/short-term/recent-decisions.json`:
   - Record choice made
   - Document rationale
   - Tag for categorization

**Output**: Updated memory state

**Validation**: Memory accurately reflects current state

---

### Hook: Progress Checkpoint

**Trigger**: Every 30 minutes of active work OR major milestone  
**Responsible**: Orchestrator  
**Priority**: LOW - Progress visibility

**Actions**:
1. Request status update from active specialist
2. Update task notes in `active-tasks.yaml`
3. Identify if additional specialists needed
4. Check if task should be broken down further

**Output**: Progress status

**Validation**: Progress tracked, no blockers unaddressed

---

## Post-Task Hooks

### Hook: Memory Updates

**Trigger**: Completion of any development task  
**Responsible**: Memory Specialist  
**Priority**: CRITICAL - Capture outcomes

**Actions**:
1. Update `current-context.json`:
   - Clear completed task from context
   - Archive conversation summary to working-notes.md
   - Update session state
2. Update `active-tasks.yaml`:
   - Set task status to "completed"
   - Add completion timestamp
   - Document final outcomes
3. Evaluate for promotion to long-term:
   - Check if patterns emerged (3+ uses)
   - Check if decisions validated
   - Add to promotion queue if applicable

**Output**: Short-term memory updated, promotion candidates identified

**Validation**: All outcomes captured

---

### Hook: Quality Gate Validation

**Trigger**: Task marked complete  
**Responsible**: Orchestrator (coordinates validators)  
**Priority**: CRITICAL - Ensure quality

**Actions**:
1. **Code Quality** (if code written):
   - Go Specialist: Code formatted, linted, vet passed
   - Test Specialist: Tests written, coverage adequate
   - Documentation Specialist: Code documented
2. **Functional Quality**:
   - Test Specialist: All tests passing
   - Integration tests passed
3. **Documentation Quality**:
   - Documentation Specialist: Docs updated
   - Memory Specialist: Memory updated
4. **Pattern Capture**:
   - Memory Specialist: Patterns documented

**Output**: Quality validation report

**Validation**: All gates passed OR issues documented

---

### Hook: Component Registry Update

**Trigger**: New component created or existing modified  
**Responsible**: Memory Specialist  
**Priority**: MEDIUM - Keep registry current

**Actions**:
1. Update `Memory-System/long-term/entity-memory/components.json`:
   - Add new component OR
   - Update existing component metadata
   - Update dependencies array
   - Update interfaces (provides/requires)
2. Update `Memory-System/long-term/entity-memory/dependencies.yaml`:
   - Add new dependency relationships
   - Update internal dependencies
3. Update `Memory-System/long-term/entity-memory/apis.yaml` if APIs changed

**Output**: Component registry reflects current state

**Validation**: All component relationships accurate

---

## Validation Hooks

### Hook: Schema Validation

**Trigger**: Creation or modification of JSON/YAML files  
**Responsible**: Data Specialist  
**Priority**: HIGH - Prevent invalid data

**Actions**:
1. Identify schema for file (from `$schema` or comment)
2. Validate against declared schema using `ajv` (JSON) or `yq` (YAML)
3. Check required fields present
4. Verify data types and constraints
5. Report validation errors with context
6. Block commit if validation fails

**Output**: Validation report

**Validation**: All structured data conforms to schemas

---

### Hook: Test Coverage Validation

**Trigger**: Go code modified  
**Responsible**: Test Specialist  
**Priority**: HIGH - Maintain quality

**Actions**:
1. Run `go test -cover ./...` for affected packages
2. Check coverage percentage
3. Identify uncovered lines
4. Fail if coverage < 80% (or project threshold)
5. Report coverage details

**Output**: Coverage report

**Validation**: Coverage meets or exceeds threshold

---

### Hook: Code Quality Validation

**Trigger**: Go code committed  
**Responsible**: Go Specialist  
**Priority**: HIGH - Code quality

**Actions**:
1. Run `gofmt -l .` - check formatting
2. Run `go vet ./...` - check for issues
3. Run `golangci-lint run` - comprehensive linting
4. Check for race conditions: `go test -race ./...`
5. Fail if any checks fail
6. Report specific issues

**Output**: Code quality report

**Validation**: All quality checks passed

---

## Memory System Hooks

### Hook: Pattern Recognition

**Trigger**: Solution successfully applied 3+ times  
**Responsible**: Memory Specialist  
**Priority**: MEDIUM - Capture patterns

**Actions**:
1. Scan `recent-decisions.json` for repeated solutions
2. Identify pattern candidates (3+ successful uses)
3. Extract pattern components:
   - Problem context
   - Solution approach
   - Examples
   - Success metrics
4. Validate pattern against quality standards
5. Promote to `long-term/knowledge-base/patterns.yaml`
6. Update pattern metadata (tags, relationships)
7. Archive source decisions

**Output**: New pattern in long-term memory

**Validation**: Pattern documented completely, examples included

---

### Hook: Decision Finalization

**Trigger**: Temporary decision validated by implementation  
**Responsible**: Memory Specialist  
**Priority**: MEDIUM - Preserve decisions

**Actions**:
1. Identify validated decisions in `recent-decisions.json`
2. Check validation criteria:
   - Implementation successful
   - No significant issues
   - Becomes project standard
3. Transform to ADR format
4. Promote to `long-term/knowledge-base/decisions.json`
5. Document consequences (positive and negative)
6. Link related decisions
7. Archive source decision

**Output**: ADR in long-term memory

**Validation**: Decision fully documented with rationale

---

### Hook: Archival Trigger

**Trigger**: Age-based (30 days) or size-based (>10MB short-term)  
**Responsible**: Memory Specialist  
**Priority**: LOW - Memory optimization

**Actions**:
1. Identify archival candidates:
   - Files older than 30 days
   - Completed tasks
   - Obsolete decisions
2. Run promotion protocol first (capture valuable knowledge)
3. Compress archival candidates (gzip)
4. Move to `Memory-System/archive/YYYY-MM/`
5. Update archive index
6. Clean short-term memory
7. Log archival operation in changelog

**Output**: Clean short-term memory, archived data preserved

**Validation**: No valuable knowledge lost, short-term size reduced

---

## Configuration Update Hooks

### Hook: Configuration Change Validation

**Trigger**: Modification of configuration files  
**Responsible**: Data Specialist  
**Priority**: CRITICAL - Prevent invalid configs

**Actions**:
1. Backup current configuration to timestamped file
2. Validate new configuration:
   - Schema validation
   - Semantic validation
   - Dependencies check
3. Test configuration in isolated environment if possible
4. If validation fails:
   - Report detailed errors
   - Rollback to backup
   - Block change
5. If validation succeeds:
   - Document change in changelog
   - Update Memory Specialist
   - Proceed with change

**Output**: Valid configuration OR rollback

**Validation**: Configuration validated, documented, functional

---

### Hook: Configuration Documentation

**Trigger**: Configuration file modified  
**Responsible**: Documentation Specialist + Memory Specialist  
**Priority**: MEDIUM - Keep docs current

**Actions**:
1. Update inline documentation in config file
2. Update configuration documentation
3. Update migration guide if breaking change
4. Update `recent-decisions.json` with change rationale
5. Update component registry if config affects components

**Output**: Documentation reflects configuration changes

**Validation**: All docs updated and accurate

---

## Git Integration Hooks

### Pre-Commit Hook

**Trigger**: `git commit`  
**Responsible**: Multiple specialists (automated)  
**Priority**: CRITICAL - Quality gate

**Implementation** (`.git/hooks/pre-commit`):
```bash
#!/bin/bash
set -e

echo "Running pre-commit hooks..."

# 1. Format check (Go Specialist)
echo "Checking Go formatting..."
if ! gofmt -l . | grep -q '^$'; then
    echo "❌ Code not formatted. Run: gofmt -w ."
    exit 1
fi

# 2. Schema validation (Data Specialist)
echo "Validating JSON/YAML files..."
for file in $(git diff --cached --name-only | grep '\.json$'); do
    if ! jq empty "$file" 2>/dev/null; then
        echo "❌ Invalid JSON: $file"
        exit 1
    fi
    
    schema=$(jq -r '."$schema" // empty' "$file" 2>/dev/null)
    if [ -n "$schema" ]; then
        schema_file="Memory-System/schemas/$(basename "$schema")"
        if [ -f "$schema_file" ]; then
            if ! ajv validate -s "$schema_file" -d "$file" 2>/dev/null; then
                echo "❌ Schema validation failed: $file"
                exit 1
            fi
        fi
    fi
done

# 3. Go vet (Go Specialist)
echo "Running go vet..."
if ! go vet ./... 2>/dev/null; then
    echo "❌ go vet found issues"
    exit 1
fi

# 4. Tests (Test Specialist)
echo "Running tests..."
if ! go test ./... 2>/dev/null; then
    echo "❌ Tests failed"
    exit 1
fi

echo "✅ All pre-commit hooks passed!"
```

---

### Post-Commit Hook

**Trigger**: `git commit` (after successful commit)  
**Responsible**: Memory Specialist  
**Priority**: LOW - Tracking

**Actions**:
1. Update changelog in long-term memory
2. Update `current-context.json` with commit info
3. Track modified files in entity memory

---

## Continuous Integration Hooks

### CI Pipeline Hooks

**Trigger**: Push to repository  
**Responsible**: Automated CI/CD  
**Priority**: CRITICAL - Quality gate

**Pipeline Stages**:
1. **Build**: `go build ./...`
2. **Test**: `go test -race -cover ./...`
3. **Lint**: `golangci-lint run`
4. **Schema Validation**: Validate all JSON/YAML
5. **Security Scan**: Check for vulnerabilities
6. **Coverage Report**: Generate and publish

**Failure Action**: Block merge, notify responsible agent

---

## Hook Management

### Enable Hook
```bash
# Create hook script in .git/hooks/
chmod +x .git/hooks/pre-commit
```

### Disable Hook
```bash
# Temporarily bypass
git commit --no-verify

# Or remove hook
rm .git/hooks/pre-commit
```

### Hook Debugging
```bash
# Run hook manually
.git/hooks/pre-commit

# Debug with verbose output
bash -x .git/hooks/pre-commit
```

---

## Best Practices

### Do's ✅
- Keep hooks fast (< 5 seconds)
- Provide clear error messages
- Make hooks idempotent
- Test hooks before deploying
- Document hook requirements
- Version control hook scripts
- Use hooks for automation, not restrictions

### Don'ts ❌
- Don't make hooks blocking unnecessarily
- Don't hide errors in hooks
- Don't make hooks too complex
- Don't bypass critical hooks
- Don't forget to document hook behavior

---

## Summary

Agent hooks ensure:
- **Consistency**: Same process every time
- **Quality**: Automated quality gates
- **Memory**: Context always current
- **Coordination**: Proper agent handoffs
- **Efficiency**: Automation reduces manual steps

By leveraging hooks throughout the development lifecycle, the agent system operates smoothly and maintains high quality standards automatically.
