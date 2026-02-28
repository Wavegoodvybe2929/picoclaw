# Technical Architecture - Picoclaw

> **🏗️ LIVING DOCUMENT**: Technical decisions, patterns, and infrastructure. Update when architecture changes.

---

## Architectural Principles

### Core Principles

1. **Separation of Concerns**: Clear boundaries between layers
2. **Interface-Driven Design**: Depend on interfaces, not implementations
3. **Event-Driven Architecture**: Loose coupling via event bus
4. **Fail-Safe Defaults**: Graceful degradation, not catastrophic failure
5. **Configuration Over Code**: Externalize behavior where possible
6. **Testability First**: Design for easy testing

---

## System Architecture

### Layered Architecture

```
┌─────────────────────────────────────────┐
│        Presentation Layer               │
│  (Channel Adapters - platform specific) │
└──────────────┬─────────────────────────┘
               │
┌──────────────▼─────────────────────────┐
│         Integration Layer               │
│   (Event Bus, Message Normalization)    │
└──────────────┬─────────────────────────┘
               │
┌──────────────▼─────────────────────────┐
│         Business Logic Layer            │
│  (Routing, Session, Agent System)       │
└──────────────┬─────────────────────────┘
               │
┌──────────────▼─────────────────────────┐
│         Provider Layer                  │
│    (LLM Provider Abstraction)           │
└──────────────┬─────────────────────────┘
               │
┌──────────────▼─────────────────────────┐
│         External Services               │
│       (OpenAI, Claude, etc.)            │
└─────────────────────────────────────────┘
```

---

## Key Design Patterns

### Provider Pattern

**Purpose**: Abstract LLM provider differences

**Implementation**:
```go
type Provider interface {
    Name() string
    Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)
    Stream(ctx context.Context, req CompletionRequest) (<-chan StreamChunk, error)
}
```

**Benefits**:
- Easy to add new providers
- Uniform interface for routing layer
- Testable via mocks

**Usage**:
- OpenAI provider: `pkg/providers/openai/`
- Anthropic provider: `pkg/providers/anthropic/`
- Local model provider: `pkg/providers/local/`

---

### Channel Adapter Pattern

**Purpose**: Normalize different chat platform APIs

**Implementation**:
```go
type Channel interface {
    Name() string
    Start(ctx context.Context) error
    Stop() error
    Send(ctx context.Context, msg Message) error
}
```

**Benefits**:
- Platform-specific code isolated
- Uniform event emission
- Easy testing

**Adapters**:
- Discord: `pkg/channels/discord.go`
- Telegram: `pkg/channels/telegram.go`
- Slack: `pkg/channels/slack.go`

---

### Event Bus Pattern

**Purpose**: Decouple components via events

**Implementation**:
```go
type Bus interface {
    Publish(ctx context.Context, event Event) error
    Subscribe(eventType string, handler Handler) error
}
```

**Events**:
- `message.received`: New message from channel
- `message.sent`: Message sent to channel
- `agent.invoked`: Agent processing started
- `completion.received`: LLM response received

**Benefits**:
- Loose coupling
- Easy to add new event handlers
- Auditable event stream

---

### Repository Pattern

**Purpose**: Abstract data persistence (if used)

**Implementation**:
```go
type SessionRepository interface {
    Get(ctx context.Context, id string) (*Session, error)
    Save(ctx context.Context, session *Session) error
    Delete(ctx context.Context, id string) error
}
```

**Benefits**:
- Swap storage implementations
- Easy to test with mocks
- Consistent data access patterns

---

## Data Flow

### Message Processing Flow

```
1. Platform → Channel Adapter
   - Platform-specific webhook/polling
   - Parse platform message format
   
2. Channel Adapter → Event Bus
   - Normalize to internal Message format
   - Emit message.received event
   
3. Event Bus → Routing Layer
   - Subscribe to message.received
   - Determine target agent/session
   
4. Routing Layer → Session Manager
   - Load or create session
   - Add message to conversation
   
5. Session Manager → Agent System
   - Provide context to agent
   - Agent determines action
   
6. Agent System → Provider Layer
   - Format provider request
   - Call appropriate provider
   
7. Provider Layer → LLM Service
   - Make API call
   - Handle streaming if applicable
   
8. LLM Response → Agent System
   - Process response
   - Execute any tools/skills
   
9. Agent System → Channel Adapter
   - Format response for platform
   - Emit message.sent event
   
10. Channel Adapter → Platform
    - Send via platform API
    - Handle platform-specific formatting
```

---

## Concurrency Model

### Goroutine Usage

**Channel Adapters**:
- One goroutine per channel for event polling/receiving
- Graceful shutdown via context cancellation

**Event Bus**:
- Async event processing in goroutines
- Subscriber handlers run concurrently
- Event order not guaranteed (use correlation IDs if needed)

**Agent System**:
- One goroutine per active agent conversation
- Context propagation for cancellation
- Worker pool for tool execution (if needed)

**Provider Calls**:
- Concurrent calls to different providers OK
- Rate limiting per provider
- Timeout context for all calls

### Synchronization

**sync.Mutex**:
- Session state updates
- Provider token accounting
- Channel state management

**sync.RWMutex**:
- Configuration reading (frequent reads, rare writes)
- Component registry

**Channels**:
- Event distribution
- Worker pools
- Shutdown coordination

---

## Error Handling Strategy

### Error Propagation

**Principle**: Errors flow up, context wraps down

**Pattern**:
```go
func ProcessMessage(msg Message) error {
    if err := ValidateMessage(msg); err != nil {
        return fmt.Errorf("validate message: %w", err)
    }
    
    if err := StoreMessage(msg); err != nil {
        return fmt.Errorf("store message: %w", err)
    }
    
    return nil
}
```

### Error Types

**Recoverable Errors**:
- Network timeouts → Retry with backoff
- Rate limits → Queue and retry
- Validation errors → Return to user

**Non-Recoverable Errors**:
- Invalid configuration → Fail fast at startup
- Critical service unavailable → Circuit breaker
- Panic in goroutine → Recover, log, continue

### Error Responses

**To Users**:
- Friendly error messages
- No stack traces
- Actionable guidance

**To Logs**:
- Full error chain
- Context (user ID, session ID, etc.)
- Stack trace if panic
- Timestamp and severity

---

## Configuration Management

### Configuration Layers

1. **Defaults** (in code): Safe fallback values
2. **File** (`config/config.json`): Environment-independent
3. **Environment Variables**: Override for deployment
4. **Runtime**: Dynamic updates (if applicable)

### Configuration Structure

```json
{
  "server": {
    "port": 8080,
    "host": "localhost"
  },
  "channels": [...],
  "providers": [...],
  "agent": {
    "memory": {...},
    "skills": {...}
  }
}
```

### Validation

- Schema validation at load time
- Type safety via Go structs
- Required fields enforced
- Default values documented

---

## Performance Considerations

### Optimization Priorities

1. **Correctness First**: Working is better than fast-and-broken
2. **Measure Before Optimizing**: Profile to find bottlenecks
3. **Optimize Hot Paths**: Message processing, token counting
4. **Cache Judiciously**: Session context, provider metadata

### Known Bottlenecks

**Message Processing**:
- Target: < 10ms per message
- Profile: Token counting, validation

**Provider API Calls**:
- Dominated by network latency
- Use streaming for long responses
- Implement timeouts

**Memory Operations**:
- File I/O for memory system
- Cache frequently accessed patterns
- Async writes where possible

### Scaling Strategy

**Vertical Scaling**:
- More CPU cores → More concurrent conversations
- More memory → Larger context windows

**Horizontal Scaling** (future):
- Stateless routing layer
- Shared session storage (Redis)
- Load balancer for HTTP endpoints

---

## Security Architecture

### Authentication

**API Keys**:
- Per-provider API keys in configuration
- Never logged or exposed
- Rotatable without code changes

**Channel Authentication**:
- Platform-specific auth (tokens, webhooks)
- Signature verification where supported
- IP whitelisting (optional)

### Authorization

**User Permissions** (if applicable):
- Role-based access control
- Per-channel permissions
- Admin vs regular user separation

### Data Security

**At Rest**:
- Configuration encrypted if contains secrets
- Session data protected by file permissions
- Memory system in version control (no secrets)

**In Transit**:
- HTTPS for all external APIs
- TLS for internal services (if distributed)

**Secrets Management**:
- Environment variables preferred
- Secret management service (future)
- Never commit secrets to Git

---

## Testing Strategy

### Test Levels

**Unit Tests**:
- Per-package test files
- Mock interfaces for dependencies
- Table-driven tests
- Coverage target: >80%

**Integration Tests**:
- Channel → Event Bus → Routing
- Provider abstraction layer
- Build tag: `//go:build integration`

**End-to-End Tests**:
- Full message flow
- Mock external services
- Build tag: `//go:build e2e`

### Mocking Strategy

**Interfaces**:
- Provider interface → Mock provider
- Channel interface → Mock channel
- Repository interfaces → In-memory repos

**External Services**:
- HTTPTest for API mocking
- Mock servers for development

---

## Observability

### Logging

**Levels**:
- ERROR: Unrecoverable errors
- WARN: Recoverable issues
- INFO: Important events
- DEBUG: Detailed diagnostics

**Structured Logging**:
```go
logger.Info("message received",
    "channel", "discord",
    "user_id", userID,
    "session_id", sessionID)
```

### Metrics (Future)

**Key Metrics**:
- Messages processed per second
- Average processing latency
- Provider response times
- Error rates by type
- Active sessions count

**Tools**:
- Prometheus (future)
- Grafana dashboards (future)

### Tracing (Future)

**Distributed Tracing**:
- OpenTelemetry integration
- Trace entire message flow
- Visualize bottlenecks

---

## Deployment Architecture

### Current: Single Binary

**Deployment**:
```bash
go build -o picoclaw ./cmd/picoclaw
./picoclaw --config=config.json
```

**Benefits**:
- Simple deployment
- No orchestration needed
- Easy debugging

**Limitations**:
- Single point of failure
- Vertical scaling only

### Future: Containerized

**Docker**:
- Dockerfile included
- Multi-stage build
- Minimal base image (alpine)

**Docker Compose**:
- Picoclaw service
- Redis (for sessions)
- Monitoring stack

### Future: Kubernetes (Optional)

**Deployment**:
- StatefulSet for sessions (if not using Redis)
- Deployment for stateless routing
- ConfigMaps for configuration
- Secrets for API keys

---

## Technology Decisions

### Why Go?

**Strengths**:
- Excellent concurrency model
- Fast startup and execution
- Strong standard library
- Easy deployment (single binary)
- Good tooling (gofmt, go vet, etc.)

**Trade-offs**:
- Verbose error handling
- Limited generics (improving)

### Why Event Bus?

**Benefits**:
- Decoupling
- Easy to add handlers
- Auditable event stream
- Natural async processing

**Trade-offs**:
- Event order not guaranteed
- Harder to reason about flow
- Need good logging/tracing

### Why Provider Abstraction?

**Benefits**:
- Uniform interface
- Easy to add providers
- Testable

**Trade-offs**:
- Lowest-common-denominator API
- Provider-specific features harder

---

## Migration Strategies

### Configuration Changes

**Process**:
1. Add new field with default
2. Support old field (deprecated)
3. Migration script to update configs
4. Remove old field after deprecation period

### Schema Changes

**Process**:
1. Design new schema
2. Create migration script
3. Test migration with backup
4. Rollback plan ready
5. Apply migration
6. Validate

### Code Refactoring

**Process**:
1. Architecture Specialist designs approach
2. Create interfaces for new design
3. Implement alongside old code
4. Gradually migrate callers
5. Remove old code when unused

---

## References

- Go Best Practices: `Agent-Config/go-specialist.md`
- Memory Architecture: `Project-Memory/memory-management.md`
- Component Registry: `Memory-System/long-term/entity-memory/components.json`
- API Documentation: `docs/`

---

**Last Updated**: 2026-02-25  
**Updated By**: Architecture Specialist + Memory Specialist  
**Version**: 1.0.0
