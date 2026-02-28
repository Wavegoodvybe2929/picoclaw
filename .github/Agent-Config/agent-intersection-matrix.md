# Agent Intersection Matrix

> **🔗 COLLABORATION GUIDE**: This matrix defines when and how specialist agents should collaborate, preventing gaps and overlaps in responsibilities.

---

## Overview

The Agent Intersection Matrix maps collaboration patterns between specialist agents. It answers:
- Which agents work together on specific tasks?
- What does each agent contribute?
- How do agents hand off work?
- What quality standards apply to joint work?

---

## Matrix Format

Each section defines:
- **Scenario**: Type of work requiring multiple agents
- **Primary Agent**: Agent with main responsibility
- **Supporting Agents**: Agents providing assistance
- **Handoff Protocol**: How work flows between agents
- **Quality Gates**: Validation required at each step
- **Common Pitfalls**: What to avoid

---

## Feature Implementation

### Scenario: New Go Feature Implementation

**Primary Agent**: Go Specialist  
**Supporting Agents**: Architecture Specialist, Test Specialist, Documentation Specialist, Memory Specialist

**Workflow**:
```
Orchestrator receives request
    ↓
Memory Specialist loads context
    ↓
Architecture Specialist reviews design
    ↓ (if approved)
Go Specialist implements feature
    ↓ (concurrent)
Test Specialist writes tests
    ↓
Documentation Specialist updates docs
    ↓
Memory Specialist captures patterns
```

**Agent Contributions**:

| Agent | Contribution | Deliverable |
|-------|--------------|-------------|
| Memory Specialist | Load relevant patterns, track progress | Context package, updated memory |
| Architecture Specialist | Design review, interface design | Approved design document |
| Go Specialist | Feature implementation | Working code, idiomatic Go |
| Test Specialist | Test coverage | Tests with >80% coverage |
| Documentation Specialist | Documentation | Updated docs and code comments |
| Memory Specialist | Pattern capture | Updated patterns if reusable |

**Handoff Protocol**:
1. Memory → Architecture: Provide design context and existing patterns
2. Architecture → Go: Approve design, provide interface specifications
3. Go → Test: Provide implementation for test coverage
4. Test → Documentation: Confirm feature working for documentation
5. Documentation → Memory: Provide docs for knowledge base

**Quality Gates**:
- Architecture approval before implementation
- Tests passing before documentation
- Documentation complete before task closure

---

## Bug Fixing

### Scenario: Production Bug Fix

**Primary Agent**: Debug Specialist  
**Supporting Agents**: Go Specialist, Test Specialist, Memory Specialist

**Workflow**:
```
Orchestrator receives bug report
    ↓
Memory Specialist loads debugging context + similar issues
    ↓
Debug Specialist analyzes root cause
    ↓
Go Specialist implements fix
    ↓
Test Specialist adds regression tests
    ↓
Memory Specialist documents in lessons-learned
```

**Agent Contributions**:

| Agent | Contribution | Deliverable |
|-------|--------------|-------------|
| Memory Specialist | Load similar issues, debugging context | Historical context |
| Debug Specialist | Root cause analysis, reproduction | RCA document |
| Go Specialist | Bug fix implementation | Corrected code |
| Test Specialist | Regression tests | Tests preventing recurrence |
| Memory Specialist | Lesson capture | Updated lessons-learned.yaml |

**Handoff Protocol**:
1. Memory → Debug: Historical similar issues, debugging patterns
2. Debug → Go: Root cause identified, fix approach recommended
3. Go → Test: Fix implemented, needs regression tests
4. Test → Memory: Tests validated, ready for documentation

**Quality Gates**:
- Root cause must be identified before fixing
- Regression tests required before merge
- Lesson learned documented in long-term memory

---

## Refactoring

### Scenario: Code Refactoring

**Primary Agent**: Architecture Specialist  
**Supporting Agents**: Go Specialist, Test Specialist, Memory Specialist

**Workflow**:
```
Orchestrator receives refactoring request
    ↓
Memory Specialist loads component relationships
    ↓
Architecture Specialist designs refactoring approach
    ↓
Go Specialist implements refactoring
    ↓
Test Specialist validates no regressions
    ↓
Memory Specialist updates component registry
```

**Agent Contributions**:

| Agent | Contribution | Deliverable |
|-------|--------------|-------------|
| Memory Specialist | Component relationships, dependencies | Dependency graph |
| Architecture Specialist | Refactoring design, impact analysis | Refactoring plan |
| Go Specialist | Code refactoring | Refactored code |
| Test Specialist | Comprehensive regression testing | Validation that nothing broke |
| Memory Specialist | Updated component registry | Current system state |

**Handoff Protocol**:
1. Memory → Architecture: Component registry, dependency graph
2. Architecture → Go: Refactoring plan, new structure
3. Go → Test: Refactored code for validation
4. Test → Memory: Validation complete, update registry

**Quality Gates**:
- Architecture approval required before implementation
- All tests must pass (no regressions)
- Component registry updated to reflect changes

---

## Configuration Changes

### Scenario: Configuration Update

**Primary Agent**: Data Specialist  
**Supporting Agents**: Go Specialist, Test Specialist, Memory Specialist

**Workflow**:
```
Orchestrator receives config change request
    ↓
Memory Specialist loads current config state
    ↓
Data Specialist validates schema and semantics
    ↓
Test Specialist validates configuration
    ↓
Go Specialist updates loading code (if needed)
    ↓
Memory Specialist documents change
```

**Agent Contributions**:

| Agent | Contribution | Deliverable |
|-------|--------------|-------------|
| Memory Specialist | Configuration history | Config context |
| Data Specialist | Schema validation, format correctness | Valid configuration |
| Test Specialist | Configuration testing | Validation tests |
| Go Specialist | Config loading (if structure changed) | Updated loading code |
| Memory Specialist | Change documentation | Updated decisions.json |

**Handoff Protocol**:
1. Memory → Data: Configuration history and schema
2. Data → Test: Validated configuration for testing
3. Test → Go: If config structure changed, update loading
4. Go → Memory: Document change and rationale

**Quality Gates**:
- Schema validation must pass
- Configuration must be testable
- Breaking changes require migration guide

---

## API Endpoint Creation

### Scenario: New REST API Endpoint

**Primary Agent**: API Specialist (if exists) OR Go Specialist  
**Supporting Agents**: Security Specialist, Test Specialist, Documentation Specialist, Data Specialist, Memory Specialist

**Workflow**:
```
Orchestrator receives API request
    ↓
Memory Specialist loads API patterns and standards
    ↓
Data Specialist designs request/response schemas
    ↓
Go Specialist implements endpoint
    ↓
Security Specialist reviews authentication/authorization
    ↓
Test Specialist creates integration tests
    ↓
Documentation Specialist documents API
    ↓
Memory Specialist captures patterns and updates API registry
```

**Agent Contributions**:

| Agent | Contribution | Deliverable |
|-------|--------------|-------------|
| Memory Specialist | API patterns, existing endpoints | Context |
| Data Specialist | Schema design | JSON schemas |
| Go Specialist | Endpoint implementation | Working endpoint |
| Security Specialist | Auth validation | Security review |
| Test Specialist | Integration tests | API tests |
| Documentation Specialist | API documentation | OpenAPI/docs |
| Memory Specialist | API registry update | Updated apis.yaml |

**Quality Gates**:
- Schema validated before implementation
- Security review required
- Integration tests must pass
- API documented before release

---

## Database Schema Change

### Scenario: Database Migration

**Primary Agent**: Database Specialist (if exists) OR Go Specialist  
**Supporting Agents**: Data Specialist, Test Specialist, Memory Specialist

**Workflow**:
```
Orchestrator receives schema change request
    ↓
Memory Specialist loads current schema and data models
    ↓
Data Specialist designs new schema
    ↓
Database Specialist creates migration
    ↓
Test Specialist validates migration
    ↓
Memory Specialist updates data-models.yaml
```

**Agent Contributions**:

| Agent | Contribution | Deliverable |
|-------|--------------|-------------|
| Memory Specialist | Current schema, data models | Schema context |
| Data Specialist | Schema design | New schema design |
| Database Specialist | Migration implementation | Migration script |
| Test Specialist | Migration testing | Test with rollback |
| Memory Specialist | Data model updates | Updated data-models.yaml |

**Quality Gates**:
- Migration must be reversible
- Data integrity preserved
- Performance impact assessed
- Models documentation updated

---

## Channel Integration

### Scenario: New Chat Platform Channel

**Primary Agent**: Go Specialist (for picoclaw)  
**Supporting Agents**: Agent System Specialist, Test Specialist, Documentation Specialist, Memory Specialist

**Workflow**:
```
Orchestrator receives channel integration request
    ↓
Memory Specialist loads channel patterns
    ↓
Agent System Specialist designs integration approach
    ↓
Go Specialist implements channel adapter
    ↓
Test Specialist creates channel tests
    ↓
Documentation Specialist documents configuration
    ↓
Memory Specialist captures integration patterns
```

**Agent Contributions**:

| Agent | Contribution | Deliverable |
|-------|--------------|-------------|
| Memory Specialist | Channel patterns, existing integrations | Pattern library |
| Agent System Specialist | Integration design | Integration approach |
| Go Specialist | Channel implementation | Working channel adapter |
| Test Specialist | Channel testing | Integration tests |
| Documentation Specialist | Configuration docs | Setup guide |
| Memory Specialist | Pattern capture | Updated patterns |

**Quality Gates**:
- Event handling complete
- Message formatting validated
- Error recovery implemented
- Configuration documented

---

## Provider Integration

### Scenario: New LLM Provider

**Primary Agent**: Go Specialist  
**Supporting Agents**: API Specialist (if exists), Test Specialist, Security Specialist, Memory Specialist

**Workflow**:
```
Orchestrator receives provider integration request
    ↓
Memory Specialist loads provider patterns
    ↓
Go Specialist implements provider adapter
    ↓
Security Specialist reviews API key handling
    ↓
Test Specialist creates provider tests
    ↓
Memory Specialist documents provider patterns
```

**Agent Contributions**:

| Agent | Contribution | Deliverable |
|-------|--------------|-------------|
| Memory Specialist | Provider patterns | Integration patterns |
| Go Specialist | Provider implementation | Provider adapter |
| Security Specialist | Security review | Security validation |
| Test Specialist | Provider testing | Mock tests |
| Memory Specialist | Pattern documentation | Updated patterns |

**Quality Gates**:
- API abstraction maintained
- Rate limiting respected
- Token tracking accurate
- Failover tested

---

## Collaboration Patterns

### Sequential Collaboration

**Pattern**: Work passes from one agent to next  
**When**: Dependencies between agent work  
**Example**: Design → Implementation → Testing

**Protocol**:
1. Agent A completes work
2. Agent A validates deliverable
3. Agent A hands off to Agent B with context
4. Agent B validates received context
5. Agent B begins work

---

### Parallel Collaboration

**Pattern**: Multiple agents work simultaneously  
**When**: Independent work streams  
**Example**: Implementation + Documentation (after initial code)

**Protocol**:
1. Orchestrator identifies parallel-safe work
2. Assigns to agents simultaneously
3. Each agent works independently
4. Agents sync at checkpoints
5. Final integration and validation

---

### Consultative Collaboration

**Pattern**: Primary agent consults expert for specific aspect  
**When**: Need specialized expertise  
**Example**: Go Specialist consults Security Specialist on auth

**Protocol**:
1. Primary agent encounters specialized need
2. Primary agent requests consultation via Orchestrator
3. Specialist provides guidance
4. Primary agent implements with guidance
5. Specialist validates implementation

---

## Common Pitfalls

### Pitfall: Responsibility Gaps

**Problem**: Task requires collaboration but agents unclear on ownership  
**Example**: Who updates documentation when config changes?

**Solution**:
- Orchestrator clearly assigns primary responsibility
- Matrix defines support agent obligations
- Handoff protocol ensures no gaps

---

### Pitfall: Duplicate Work

**Problem**: Multiple agents doing same work  
**Example**: Both Go and Test Specialist write tests

**Solution**:
- Clear role definitions
- Orchestrator prevents overlap
- Active-tasks.yaml shows current assignments

---

### Pitfall: Missing Handoffs

**Problem**: Agent completes work but doesn't notify next agent  
**Example**: Implementation done but Test Specialist not notified

**Solution**:
- Explicit handoff protocol in matrix
- Agent hooks enforce handoffs
- Orchestrator tracks handoff completion

---

### Pitfall: Incomplete Context

**Problem**: Supporting agent lacks context from primary agent  
**Example**: Test Specialist doesn't know implementation details

**Solution**:
- Handoff includes context package
- Memory Specialist maintains context
- Supporting agent validates context before starting

---

## Success Metrics

**Collaboration Quality**:
- No responsibility gaps
- No duplicate work
- All handoffs complete
- Context always sufficient
- Quality gates enforced

**Coordination Efficiency**:
- Clear role assignments
- Fast handoffs
- Minimal rework
- High specialist satisfaction

---

## Summary

The Agent Intersection Matrix ensures:
- **Clarity**: Each agent knows their role
- **Completeness**: All aspects covered
- **Efficiency**: No duplicate work
- **Quality**: Proper validation at handoffs
- **Coordination**: Smooth workflows

By defining collaboration patterns explicitly, agents work together seamlessly to deliver high-quality results.
