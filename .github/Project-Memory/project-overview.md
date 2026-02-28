# Picoclaw Project Overview

> **📊 DYNAMIC DOCUMENT**: This document reflects the current state of the picoclaw project. Update after milestones, architectural changes, or significant decisions.

---

## Project Identity

**Name**: Picoclaw  
**Type**: Multi-channel AI Chat Platform  
**Primary Language**: Go  
**Current Version**: Development/Beta  
**Repository**: /Users/wavegoodvybe/Documents/GitHub/picoclaw

---

## Project Mission

Picoclaw is a flexible, multi-channel AI chat platform that connects various messaging platforms (Discord, Telegram, Slack, WeChat, etc.) with multiple LLM providers (OpenAI, Anthropic, local models, etc.), featuring:

- **Multi-Channel Support**: Connect to diverse chat platforms
- **Multi-Provider LLM Integration**: Support various AI providers
- **Agent System**: Sophisticated agent capabilities and coordination
- **MCP Tool Integration**: Model Context Protocol for enhanced capabilities
- **Session Management**: Persistent conversation context
- **Flexible Routing**: Intelligent message and agent routing

---

## Current Project Status

**Phase**: Active Development  
**Focus**: Enhanced agent system implementation  
**Recent Milestone**: Implemented dual-layer memory management system (2026-02-25)

### Active Work Streams
1. Enhanced agent system architecture
2. Memory management framework
3. Agent coordination patterns
4. Knowledge capture and promotion

### Completed Features
- Multi-channel integration framework
- LLM provider abstraction
- Basic agent system
- Configuration management
- Session handling
- Event bus architecture

### Planned Features
- Enhanced agent coordination
- Long-term memory system
- Pattern library
- Advanced routing capabilities
- Tool marketplace
- Performance optimizations

---

## Architecture Overview

### High-Level Architecture

```
┌─────────────────────────────────────────────────────┐
│                   Chat Platforms                    │
│  (Discord, Telegram, Slack, WeChat, WhatsApp, etc.) │
└────────────────────┬────────────────────────────────┘
                     │
                     ↓
┌─────────────────────────────────────────────────────┐
│                Channel Adapters                     │
│          (pkg/channels/[platform].go)               │
└────────────────────┬────────────────────────────────┘
                     │
                     ↓
┌─────────────────────────────────────────────────────┐
│                  Event Bus                          │
│              (pkg/bus/bus.go)                       │
└────────────────────┬────────────────────────────────┘
                     │
                     ↓
┌─────────────────────────────────────────────────────┐
│              Routing & Session                      │
│    (pkg/routing/, pkg/session/)                     │
└────────────────────┬────────────────────────────────┘
                     │
                     ↓
┌─────────────────────────────────────────────────────┐
│                Agent System                         │
│    (pkg/agent/ - coordination, memory, skills)      │
└────────────────────┬────────────────────────────────┘
                     │
                     ↓
┌─────────────────────────────────────────────────────┐
│               Provider Abstraction                  │
│            (pkg/providers/[provider]/)              │
└────────────────────┬────────────────────────────────┘
                     │
                     ↓
┌─────────────────────────────────────────────────────┐
│                  LLM Providers                      │
│     (OpenAI, Anthropic, Local Models, etc.)         │
└─────────────────────────────────────────────────────┘
```

### Key Components

**Channels** (`pkg/channels/`):
- Platform-specific adapters
- Event handling and normalization
- Message formatting
- Error recovery

**Event Bus** (`pkg/bus/`):
- Decoupled event distribution
- Async event processing
- Event subscription patterns

**Routing** (`pkg/routing/`):
- Message routing logic
- Agent ID mapping
- Context-aware routing

**Session** (`pkg/session/`):
- Conversation persistence
- Context window management
- Session lifecycle

**Agent System** (`pkg/agent/`):
- Agent orchestration
- Memory management
- Skill execution
- Tool integration

**Providers** (`pkg/providers/`):
- Provider abstraction layer
- API client management
- Token tracking
- Rate limiting
- Failover handling

**Tools** (`pkg/tools/`):
- MCP tool definitions
- Tool execution
- Result formatting

---

## Technology Stack

### Primary Technologies

**Core**:
- Go 1.21+ (primary language)
- JSON/YAML (configuration and data)
- Git (version control)

**Integrations**:
- Discord API
- Telegram Bot API
- Slack API
- WeChat/WeCom APIs
- LINE API
- OpenAI API
- Anthropic Claude API

**Development**:
- golangci-lint (linting)
- go test (testing)
- Docker (containerization)
- Make (build automation)

**Infrastructure**:
- JSON Schema (validation)
- YAML (configuration)
- Event-driven architecture
- Pub/Sub patterns

---

## Team Organization

### Agent System Teams

**Infrastructure Specialists**:
- Orchestrator: Request routing and coordination
- Memory Specialist: Memory system management
- Data Specialist: JSON/YAML and schemas

**Domain Specialists**:
- Go Specialist: Go implementation
- Agent System Specialist: Agent coordination
- API Specialist: API integrations
- Database Specialist: Data persistence (if applicable)
- Security Specialist: Authentication and security

**Quality Specialists**:
- Test Specialist: Test coverage and validation
- Debug Specialist: Issue investigation
- Documentation Specialist: Documentation maintenance

---

## Development Workflow

### Standard Workflow

1. **Request Intake**: All work starts with Orchestrator
2. **Context Loading**: Memory Specialist provides context
3. **Routing**: Orchestrator assigns to specialist(s)
4. **Implementation**: Specialist implements with quality standards
5. **Validation**: Multiple quality gates
6. **Documentation**: Documentation updated
7. **Memory Update**: Patterns captured, memory updated

### Quality Gates

**Pre-Implementation**:
- Requirements clear
- Design reviewed
- No conflicts
- Context loaded

**Implementation**:
- Code standards met
- Tests written
- Error handling proper
- Documentation inline

**Post-Implementation**:
- All tests passing
- Code reviewed
- Docs complete
- Memory updated

---

## Key Metrics

### Code Quality
- Test coverage target: >80%
- Linting: All golangci-lint checks pass
- No race conditions detected
- Documentation: All exported symbols documented

### Performance
- Startup time: < 100ms
- Message processing: < 10ms average
- Memory usage: Reasonable for workload
- No memory leaks

### Reliability
- Graceful degradation under load
- Channel failover working
- Provider failover working
- Error recovery tested

---

## Recent Decisions

### ADR-001: Enhanced Agent System Implementation
**Date**: 2026-02-25  
**Status**: In Progress  
**Decision**: Implement enhanced agent-config framework with dual-layer memory system for improved coordination and knowledge management.

**Rationale**:
- Need better agent coordination
- Knowledge preservation important
- Pattern recognition valuable
- Context management critical

**Impact**:
- New directory structure (Agent-Config, Project-Memory, Memory-System)
- Orchestrator-first routing required
- Memory system becomes core capability
- Documentation standards enhanced

---

## Known Issues and Challenges

### Current Challenges
1. Agent coordination complexity
2. Memory system performance at scale
3. Multi-provider token accounting
4. Rate limiting across providers
5. Session persistence strategy

### Technical Debt
- Some channel adapters need refactoring
- Test coverage needs improvement in providers
- Documentation gaps in some packages
- Performance profiling needed

---

## Project Structure

```
picoclaw/
├── cmd/
│   └── picoclaw/          # Main application entry
├── pkg/
│   ├── agent/             # Agent system
│   ├── bus/               # Event bus
│   ├── channels/          # Platform integrations
│   ├── config/            # Configuration
│   ├── providers/         # LLM providers
│   ├── routing/           # Message routing
│   ├── session/           # Session management
│   ├── skills/            # Agent skills
│   ├── tools/             # MCP tools
│   └── utils/             # Utilities
├── config/                # Configuration files
├── docs/                  # Documentation
├── workspace/             # Workspace files
├── Agent-Config/          # Agent configurations (NEW)
├── Project-Memory/        # Project documentation (NEW)
└── Memory-System/         # Memory management (NEW)
```

---

## Communication Channels

### Internal
- Project-Memory/ (this directory)
- Memory-System/ (context and patterns)
- Agent-Config/ (workflows)

### External  
- GitHub repository
- Issue tracker
- Pull requests
- Documentation

---

## Next Steps

### Immediate (Current Sprint)
1. Complete enhanced agent system implementation
2. Create validation scripts
3. Set up pre-commit hooks
4. Document agent workflows

### Short-term (Next 2-4 weeks)
1. Implement pattern recognition
2. Build out component registry
3. Expand test coverage
4. Performance profiling

### Medium-term (Next 1-3 months)
1. Advanced routing capabilities
2. Tool marketplace
3. Enhanced session management
4. Performance optimizations

---

## Update Protocol

**When to Update**:
- After major milestones
- After architectural decisions
- After significant feature completions
- Monthly review minimum

**Who Updates**:
- Memory Specialist coordinates updates
- Domain specialists provide content
- Orchestrator reviews for accuracy

**What to Update**:
- Current status section
- Recent decisions
- Known issues
- Metrics
- Team organization (if changed)

---

## References

- Technical Architecture: `Project-Memory/technical-architecture.md`
- Development History: `Project-Memory/development-history.md`
- Memory Management: `Project-Memory/memory-management.md`
- Agent Framework Guide: `ENHANCED_AGENT_SYSTEM_GUIDE.md`

---

**Last Updated**: 2026-02-25  
**Updated By**: Memory Specialist  
**Version**: 1.0.0
