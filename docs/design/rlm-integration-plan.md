# RLM Integration Implementation Plan

> **Status**: PLANNING COMPLETE - Ready for Implementation  
> **Date**: 2026-02-26  
> **Agent**: Orchestrator → Memory Specialist → Architecture Specialist → Go Specialist  
> **Estimated Effort**: 5 days (~385 lines of code)

---

## ✅ LM Studio Compatibility Verified

**Research Confirmed:** RLMgw works with **any OpenAI-compatible endpoint**

- ✅ **LM Studio** - Your setup (`http://localhost:1234/v1`)
- ✅ **vLLM** - Default setup (`http://localhost:8000/v1`)
- ✅ **Ollama** - Local models
- ✅ **OpenAI** - Cloud API
- ✅ **Anthropic, Groq, etc.** - Any OpenAI-compatible provider

**Evidence from rlmgw source:**
- `upstream_base_url` is fully configurable via `RLMGW_UPSTREAM_BASE_URL`
- Uses `httpx.Client(base_url=config.upstream_base_url)` - accepts any URL
- No provider-specific logic - pure OpenAI protocol proxy

**Your configuration will be:**
```json
{
  "providers": {
    "rlm": {
      "enabled": true,
      "upstream_base_url": "http://localhost:1234/v1",  // ← Your LM Studio
      "upstream_model": "gpt-oss-20b"                   // ← Your model
    }
  }
}
```

---

## Problem Statement

Picoclaw experiences context window errors when initializing agents because the system prompt loads all context upfront:
- 28 skills (9,177 words)
- MEMORY.md (1,015 words)
- Bootstrap files (1,373 words)
- Total: ~11,500+ words loaded before user's first message

**Error:**
```
The number of tokens to keep from the initial prompt is greater than the context length
```

---

## Solution: RLM Provider Integration

**What is RLM?**
Recursive Language Models (RLMs) enable LLMs to handle near-infinite contexts by programmatically exploring and recursively calling themselves to select only relevant context.

**Repository:** https://github.com/mitkox/rlmgw

**RLMgw Component:**
- Python FastAPI server
- OpenAI-compatible gateway
- Intelligent context selection using RLM recursion
- Sits between picoclaw and actual LLM provider

---

## Integration Approach: Provider Wrapper Pattern ⭐

### Architecture

```
Agent Loop
    ↓
RLM Provider (new)
    ├─→ Spawns rlmgw subprocess (Python)
    │   │
    │   └─→ RLMgw Server (localhost:8010)
    │       ├─→ Context Selection (RLM exploration)
    │       └─→ Upstream Provider (OpenAI/Anthropic/etc)
    │
    └─→ Proxies Chat() requests
        ├─→ Adds workspace context intelligently
        └─→ Returns LLMResponse
```

### Why This Approach?

1. **Fits Existing Patterns** ✅
   - Uses existing `LLMProvider` interface
   - No changes to agent loop
   - No changes to existing providers

2. **Minimal Code** ✅
   - Estimated ~385 lines total
   - Under 500-line guideline
   - Subprocess wrapper pattern from IMPLEMENTATION_GUIDE

3. **Zero Regressions** ✅
   - Opt-in via config (disabled by default)
   - Existing providers unchanged
   - Backward compatible

4. **No New Go Dependencies** ✅
   - Uses stdlib only (`os/exec`, `net/http`)
   - Python rlmgw runs as external process

---

## Implementation Phases

### Phase 1: Configuration Schema (Day 1)

**Files Modified:**
- `pkg/config/config.go` (+20 lines)

**New Config:**
```go
type RLMConfig struct {
    Enabled            bool   `json:"enabled"`
    PythonPath         string `json:"python_path,omitempty"`
    RLMGWPath          string `json:"rlmgw_path,omitempty"`
    Host               string `json:"host,omitempty"`
    Port               int    `json:"port,omitempty"`
    UpstreamBaseURL    string `json:"upstream_base_url"`                // OpenAI-compatible endpoint (LM Studio, vLLM, etc)
    UpstreamModel      string `json:"upstream_model"`                   // Model name
    WorkspaceRoot      string `json:"workspace_root,omitempty"`         // Workspace to analyze (default: agent workspace)
    UseRLMSelection    bool   `json:"use_rlm_selection"`                // Use intelligent RLM selection (true) or simple (false)
    MaxInternalCalls   int    `json:"max_internal_calls,omitempty"`     // Max recursive calls (default: 3)
    MaxContextChars    int    `json:"max_context_pack_chars,omitempty"` // Max context size (default: 12000)
}
```

**Example User Config (LM Studio):**
```json
{
  "agents": {
    "defaults": {
      "provider": "rlm",
      "model": "openai/gpt-oss-20b"
    }
  },
  "providers": {
    "openai": {
      "api_key": "lm-studio",
      "api_base": "http://localhost:1234/v1"
    },
    "rlm": {
      "enabled": true,
      "upstream_base_url": "http://localhost:1234/v1",
      "upstream_model": "gpt-oss-20b",
      "use_rlm_selection": true,
      "max_internal_calls": 3
    }
  }
}
```

**Note:** RLMgw works with **any OpenAI-compatible endpoint** including LM Studio, vLLM, Ollama, etc.

**Testing:**
- Config loads without errors
- Defaults apply correctly
- Backward compatible (no rlm config = existing behavior)

---

### Phase 2: RLM Provider Implementation (Day 2-3)

**Files Created:**
- `pkg/providers/rlm_provider.go` (~200 lines)

**Key Components:**

1. **Subprocess Management**
   ```go
   func (p *RLMProvider) startServer() error {
       // Find python3
       // Find rlmgw path (auto-detect)
       // Start subprocess with env vars
       // Wait for readiness (/readyz endpoint)
       // Return ready or error
   }
   ```

2. **Health Checking**
   ```go
   func (p *RLMProvider) waitForReady(timeout time.Duration) error {
       // Poll /readyz endpoint
       // Retry with backoff
       // Return when ready or timeout
   }
   ```

3. **Chat Implementation**
   ```go
   func (p *RLMProvider) Chat(...) (*LLMResponse, error) {
       // Build OpenAI-compatible request
       // POST to http://localhost:8010/v1/chat/completions
       // Parse response
       // Return LLMResponse
   }
   ```

4. **Lifecycle Management**
   ```go
   func (p *RLMProvider) Close() {
       // Send SIGINT to subprocess
       // Wait for graceful shutdown
       // Clean up resources
   }
   ```

**Error Handling:**
- Graceful fallback if subprocess fails
- Detailed logging for debugging
- Auto-retry on transient errors

**Testing:**
- Unit tests for config validation
- Subprocess lifecycle tests
- Mock HTTP server tests

---

### Phase 3: Provider Registration (Day 4)

**Files Modified:**
- `pkg/providers/factory_provider.go` (+15 lines)

**Changes:**
```go
func createProviderFromRef(ref ModelRef, cfg *config.Config) (LLMProvider, error) {
    switch ref.Provider {
    // ... existing cases ...
    
    case "rlm":
        rlmCfg := cfg.Providers.RLM
        if !rlmCfg.Enabled {
            return nil, fmt.Errorf("rlm provider not enabled")
        }
        return NewRLMProvider(rlmCfg)
    
    default:
        return nil, fmt.Errorf("unknown provider: %s", ref.Provider)
    }
}
```

**Testing:**
- Factory creates RLM provider correctly
- Error handling for disabled config
- Integration with existing provider system

---

### Phase 4: Testing & Documentation (Day 5)

**Test Files Created:**
- `pkg/providers/rlm_provider_test.go` (~150 lines)

**Test Coverage:**
- Config validation
- Subprocess lifecycle
- HTTP request/response
- Error handling
- Graceful shutdown

**Documentation Created:**
- `docs/RLM_INTEGRATION.md` - User guide
- `config/config.example.json` - Updated with RLM example
- Code comments - All exported symbols documented

**Manual Integration Test:**
```bash
# 1. Install rlmgw
cd ~
git clone https://github.com/mitkox/rlmgw
cd rlmgw
uv sync

# 2. Configure picoclaw for LM Studio
cat > ~/.picoclaw/config.json <<EOF
{
  "agents": {
    "defaults": {
      "provider": "rlm",
      "model": "openai/gpt-oss-20b"
    }
  },
  "providers": {
    "openai": {
      "api_key": "lm-studio",
      "api_base": "http://localhost:1234/v1"
    },
    "rlm": {
      "enabled": true,
      "upstream_base_url": "http://localhost:1234/v1",
      "upstream_model": "gpt-oss-20b",
      "workspace_root": "~/.picoclaw/workspace",
      "use_rlm_selection": true
    }
  }
}
EOF

# Note: Make sure LM Studio server is running on localhost:1234

# 3. Test
picoclaw agent
> hello
> Can you help me understand this codebase?
```

**Expected Result:**
- No context window errors
- Intelligent context selection visible in logs
- Reduced initial prompt size
- Successful responses

---

## Code Size Estimate

```
New files:
  pkg/providers/rlm_provider.go        200 lines
  pkg/providers/rlm_provider_test.go   150 lines
  docs/RLM_INTEGRATION.md              100 lines (doc)

Modified files:
  pkg/config/config.go                 +20 lines
  pkg/providers/factory_provider.go    +15 lines
  config/config.example.json           +15 lines (doc)

Total Go Code: 385 lines
Total Changes: ~385 lines code + 115 lines docs = 500 lines total
```

**Status:** ✅ Under 500-line guideline

---

## Safety & Performance Analysis

### Concurrency Safety ✅
- Subprocess spawned once, reused across requests
- HTTP client is thread-safe (standard library)
- No shared mutable state
- Clean shutdown on context cancellation

### Performance Impact
| Metric | Impact | Mitigation |
|--------|--------|------------|
| **Latency** | +1-3s for RLM context selection | Configurable `max_internal_calls` |
| **Memory** | +50-100MB for Python subprocess | Acceptable trade-off vs context errors |
| **Startup** | +2-5s for subprocess init | One-time cost, amortized |
| **Throughput** | Minimal (subprocess reused) | Session caching in rlmgw |

**Trade-off Analysis:**
- **Cost:** Higher per-request latency (1-3s)
- **Benefit:** Solves context window errors, enables larger workspaces
- **Verdict:** Acceptable for users needing large context management

### Security ✅
- Read-only access to workspace
- Subprocess runs with same user permissions
- No user input executed directly
- Localhost-only HTTP (127.0.0.1, not exposed)
- Standard security model (same as other subprocess tools)

---

## Regression Test Plan

### Existing Behavior Preservation

**Test 1: Default Config (No RLM)**
```bash
# Use standard config
picoclaw agent
> hello
```
**Expected:** Existing behavior unchanged

**Test 2: Disabled RLM Config**
```json
{
  "providers": {
    "rlm": {
      "enabled": false
    }
  }
}
```
**Expected:** Falls back to direct provider

**Test 3: All Existing Tests Pass**
```bash
go test ./...
```
**Expected:** All 188 tests pass

---

## Dependencies & Prerequisites

### Runtime Dependencies
- **Python 3.11+** - For rlmgw subprocess
- **rlmgw repository** - Cloned and installed
- **Upstream provider** - OpenAI/Anthropic/etc API key

### Installation Steps for Users

```bash
# 1. Install Python dependencies
cd ~
git clone https://github.com/mitkox/rlmgw
cd rlmgw
curl -LsSf https://astral.sh/uv/install.sh | sh
uv sync

# 2. Configure picoclaw
# Edit ~/.picoclaw/config.json (see example above)

# 3. Test
picoclaw agent
```

---

## Troubleshooting Guide

### Common Issues

**Issue 1: "rlmgw not found"**
- **Cause:** rlmgw path not configured or not installed
- **Solution:** Set `rlmgw_path` in config or install to `~/rlmgw`

**Issue 2: "rlmgw server not ready"**
- **Cause:** Python dependencies not installed
- **Solution:** `cd ~/rlmgw && uv sync`

**Issue 3: Still getting context errors**
- **Cause:** Upstream provider's context limit exceeded
- **Solution:** 
  - Reduce `max_context_chars` in config
  - Increase LM Studio's `n_ctx` parameter
  - Use model with larger context window

**Issue 4: High latency**
- **Cause:** RLM making many recursive calls
- **Solution:** Reduce `max_internal_calls` (default: 3, try 1-2)

---

## Alternative Approaches Considered

### Option A: Tool Integration
**Approach:** Create `rlm_query` tool that agent calls explicitly

**Pros:**
- Less invasive
- Agent has explicit control

**Cons:**
- Requires agent to recognize when context is large
- Not transparent to user
- Requires changing agent behavior

**Verdict:** ❌ Not recommended

---

### Option B: External Service (Sidecar)
**Approach:** Run rlmgw as separate service, configure picoclaw to use it

**Pros:**
- Zero code changes
- Service can be shared

**Cons:**
- Requires manual service management
- Not portable (users must set up separately)
- No automatic lifecycle management

**Verdict:** ❌ Not recommended for v1 (could be future enhancement)

---

### Option C: Provider Wrapper (SELECTED ✅)
**Approach:** Create RLM provider that wraps existing providers

**Pros:**
- Transparent to agent loop
- Minimal code (~385 lines)
- Automatic lifecycle management
- Fits existing patterns
- Opt-in via config

**Cons:**
- Requires Python runtime
- Subprocess overhead

**Verdict:** ✅ **RECOMMENDED** - Best balance of simplicity and functionality

---

## Success Criteria

### Implementation Complete When:
- [x] Design documented and approved
- [ ] Configuration schema implemented
- [ ] RLM provider implemented and tested
- [ ] Provider registered in factory
- [ ] Unit tests pass (>80% coverage)
- [ ] Integration tests pass
- [ ] Manual testing successful
- [ ] Documentation complete
- [ ] All existing tests pass (no regressions)
- [ ] Code review approved

### User Success When:
- [ ] User can enable RLM with simple config change
- [ ] Context window errors eliminated
- [ ] Latency acceptable (documented)
- [ ] Troubleshooting guide covers common issues
- [ ] Works with existing providers (OpenAI, Anthropic, etc.)

---

## Timeline

| Phase | Duration | Deliverable |
|-------|----------|-------------|
| **Phase 1** | Day 1 | Config schema + tests |
| **Phase 2** | Day 2-3 | RLM provider implementation |
| **Phase 3** | Day 4 | Provider registration + integration |
| **Phase 4** | Day 5 | Testing + documentation |
| **Total** | **5 days** | Complete feature ready for PR |

---

## Next Steps

### Immediate Actions:
1. **Review plan** - Get stakeholder approval
2. **Set up dev environment** - Install rlmgw for testing
3. **Create feature branch** - `git checkout -b feature/rlm-provider`
4. **Begin Phase 1** - Implement config schema

### Follow-up:
1. Submit PR following CONTRIBUTING.md
2. Address review feedback
3. Update CHANGELOG.md
4. Monitor user feedback post-release

---

## References

- **RLMgw Repository:** https://github.com/mitkox/rlmgw
- **RLM Paper:** https://arxiv.org/abs/2512.24601
- **RLM Blogpost:** https://alexzhang13.github.io/blog/2025/rlm/
- **Picoclaw Implementation Guide:** `/IMPLEMENTATION_GUIDE.md`
- **Agent System Instructions:** `/.github/copilot-instructions.md`
- **Technical Architecture:** `/.github/Project-Memory/technical-architecture.md`

---

**Plan Status:** ✅ COMPLETE - Ready for Implementation  
**Next Agent:** Go Specialist (for implementation)  
**Approval Required:** Yes (review this plan before proceeding)
