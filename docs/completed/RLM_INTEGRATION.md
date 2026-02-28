# RLM Integration User Guide

> **Recursive Language Models (RLM) Integration for PicoClaw**  
> Enables handling near-infinite contexts through intelligent context selection

---

## Overview

The RLM provider integration allows PicoClaw to handle large codebases and contexts that would normally exceed LLM context windows. It uses [RLMgw](https://github.com/mitkox/rlmgw) (Recursive Language Models Gateway) to intelligently select only relevant context before sending requests to your LLM provider.

### What is RLM?

RLM (Recursive Language Models) is a technique that enables LLMs to explore and recursively call themselves to select only the most relevant context from large datasets. This allows handling contexts far beyond typical token limits.

### How It Works

```
User Request
    ↓
PicoClaw Agent
    ↓
RLM Provider
    ↓
RLMgw Server (Python subprocess)
    ├─→ Intelligently explores workspace
    ├─→ Recursively selects relevant files/context
    └─→ Forwards optimized request to upstream provider
    ↓
Upstream Provider (OpenAI/LM Studio/Anthropic/etc)
    ↓
Response back to user
```

---

## Prerequisites

### 1. Python 3.11+

RLMgw requires Python 3.11 or later.

```bash
# Check Python version
python3 --version

# Should output: Python 3.11.x or higher
```

### 2. Install RLMgw

```bash
# Clone the rlmgw repository
cd ~
git clone https://github.com/mitkox/rlmgw
cd rlmgw

# Install uv package manager (if not already installed)
curl -LsSf https://astral.sh/uv/install.sh | sh

# Install dependencies
uv sync
```

### 3. Upstream LLM Provider

RLMgw works with **any OpenAI-compatible endpoint**, including:

- **LM Studio** - Local models (`http://localhost:1234/v1`)
- **vLLM** - Local inference server (`http://localhost:8000/v1`)
- **Ollama** - Local models with OpenAI compatibility
- **OpenAI** - Cloud API (`https://api.openai.com/v1`)
- **Anthropic, Groq, etc.** - Any OpenAI-compatible provider

---

## Configuration

### Basic Configuration (LM Studio Example)

Add the RLM provider configuration to your `~/.picoclaw/config.json`:

```json
{
  "agents": {
    "defaults": {
      "provider": "rlm",
      "model_name": "gpt-oss-20b"
    }
  },
  "providers": {
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

### Configuration Options

| Option | Type | Required | Default | Description |
|--------|------|----------|---------|-------------|
| `enabled` | boolean | Yes | `false` | Enable/disable RLM provider |
| `upstream_base_url` | string | Yes | - | OpenAI-compatible endpoint URL |
| `upstream_model` | string | Yes | - | Model name for upstream provider |
| `use_rlm_selection` | boolean | No | `false` | Use intelligent RLM context selection |
| `python_path` | string | No | auto-detect | Path to Python 3.11+ executable |
| `rlmgw_path` | string | No | `~/rlmgw` | Path to rlmgw installation |
| `host` | string | No | `127.0.0.1` | RLMgw server host (localhost only) |
| `port` | integer | No | `8010` | RLMgw server port |
| `workspace_root` | string | No | agent workspace | Workspace to analyze for context |
| `max_internal_calls` | integer | No | `3` | Max recursive calls for context selection |
| `max_context_pack_chars` | integer | No | `12000` | Max context size in characters |

### Advanced Configuration Examples

#### OpenAI Cloud Provider

```json
{
  "providers": {
    "openai": {
      "api_key": "sk-...",
      "api_base": "https://api.openai.com/v1"
    },
    "rlm": {
      "enabled": true,
      "upstream_base_url": "https://api.openai.com/v1",
      "upstream_model": "gpt-4o",
      "use_rlm_selection": true,
      "workspace_root": "~/.picoclaw/workspace"
    }
  }
}
```

#### vLLM Local Server

```json
{
  "providers": {
    "rlm": {
      "enabled": true,
      "upstream_base_url": "http://localhost:8000/v1",
      "upstream_model": "mistral-7b-instruct",
      "use_rlm_selection": true,
      "max_internal_calls": 2,
      "max_context_pack_chars": 8000
    }
  }
}
```

#### Custom Python/RLMgw Paths

```json
{
  "providers": {
    "rlm": {
      "enabled": true,
      "python_path": "/opt/homebrew/bin/python3",
      "rlmgw_path": "/Users/me/projects/rlmgw",
      "upstream_base_url": "http://localhost:1234/v1",
      "upstream_model": "llama-3-8b",
      "use_rlm_selection": true
    }
  }
}
```

#### Performance Tuning

For faster responses with reduced context quality:

```json
{
  "providers": {
    "rlm": {
      "enabled": true,
      "upstream_base_url": "http://localhost:1234/v1",
      "upstream_model": "gpt-oss-20b",
      "use_rlm_selection": true,
      "max_internal_calls": 1,
      "max_context_pack_chars": 6000
    }
  }
}
```

For maximum context quality (slower):

```json
{
  "providers": {
    "rlm": {
      "enabled": true,
      "upstream_base_url": "http://localhost:1234/v1",
      "upstream_model": "gpt-oss-20b",
      "use_rlm_selection": true,
      "max_internal_calls": 5,
      "max_context_pack_chars": 20000
    }
  }
}
```

---

## Usage

Once configured, use PicoClaw normally. The RLM provider will automatically:

1. Start the RLMgw subprocess on first use
2. Analyze your workspace when needed
3. Select relevant context intelligently
4. Forward optimized requests to your upstream provider
5. Return responses seamlessly

### Example Session

```bash
# Start PicoClaw agent
picoclaw agent

# The RLM provider will start automatically
# You'll see logs indicating RLMgw server startup

> hello
< Hello! I'm your PicoClaw agent with RLM context awareness.

> Can you help me understand the authentication system in this codebase?
< [RLM intelligently explores auth-related files and provides focused response]

> What does the auth/oauth.go file do?
< [RLM loads and explains the specific file with relevant context]
```

---

## Performance Characteristics

### Latency

| Operation | Time Impact | Notes |
|-----------|-------------|-------|
| First request | +3-5s | One-time RLMgw subprocess startup |
| Context selection | +1-3s per request | Varies based on `max_internal_calls` |
| Subsequent requests | +1-2s | RLM recursion overhead |

### Memory Usage

- **RLMgw subprocess**: ~50-100MB
- **Context cache**: Varies by workspace size
- **Total overhead**: ~100-150MB additional memory

### Trade-offs

| Benefit | Cost |
|---------|------|
| ✅ Handles large contexts (>100k tokens) | ⏱️ Additional 1-3s latency per request |
| ✅ Eliminates context window errors | 💾 ~100-150MB additional memory |
| ✅ Intelligent context selection | 🔧 Requires Python runtime |
| ✅ Works with any OpenAI-compatible provider | 📦 External dependency (rlmgw) |

---

## Troubleshooting

### Issue: "python3 not found in PATH"

**Cause:** Python 3 not installed or not in PATH

**Solution:**
```bash
# Install Python 3.11+
# macOS (Homebrew)
brew install python@3.11

# Ubuntu/Debian
sudo apt install python3.11

# Set python_path in config
{
  "providers": {
    "rlm": {
      "python_path": "/usr/local/bin/python3.11"
    }
  }
}
```

### Issue: "rlmgw not found"

**Cause:** RLMgw not installed or not at expected location

**Solution:**
```bash
# Clone and install rlmgw
cd ~
git clone https://github.com/mitkox/rlmgw
cd rlmgw
uv sync

# OR specify custom path in config
{
  "providers": {
    "rlm": {
      "rlmgw_path": "/path/to/your/rlmgw"
    }
  }
}
```

### Issue: "failed to start RLMgw server"

**Cause:** Missing Python dependencies

**Solution:**
```bash
# Reinstall dependencies
cd ~/rlmgw
uv sync

# Or try with pip
python3 -m pip install -r requirements.txt
```

### Issue: "RLMgw server not ready"

**Cause:** Server startup timeout or port already in use

**Solution:**
```bash
# Check if port is in use
lsof -i :8010

# Kill existing process if needed
kill -9 <PID>

# Or use different port in config
{
  "providers": {
    "rlm": {
      "port": 8011
    }
  }
}
```

### Issue: Still getting context window errors

**Cause:** Upstream provider's context limit exceeded even with RLM

**Solutions:**

1. **Reduce context size:**
```json
{
  "providers": {
    "rlm": {
      "max_context_pack_chars": 6000,
      "max_internal_calls": 1
    }
  }
}
```

2. **Increase upstream provider's context window:**
- For LM Studio: Increase `n_ctx` in model settings
- For vLLM: Use `--max-model-len` parameter
- For cloud providers: Use models with larger context windows

3. **Use a model with larger context:**
```json
{
  "providers": {
    "rlm": {
      "upstream_model": "claude-opus-4"  
    }
  }
}
```

### Issue: High latency / Slow responses

**Cause:** Too many recursive calls

**Solution:** Reduce `max_internal_calls`:
```json
{
  "providers": {
    "rlm": {
      "max_internal_calls": 1,
      "max_context_pack_chars": 8000
    }
  }
}
```

### Issue: "RLM provider is not enabled in configuration"

**Cause:** `enabled` is set to `false` or missing

**Solution:**
```json
{
  "providers": {
    "rlm": {
      "enabled": true
    }
  }
}
```

### Issue: "upstream_base_url is required"

**Cause:** Missing required configuration field

**Solution:**
```json
{
  "providers": {
    "rlm": {
      "enabled": true,
      "upstream_base_url": "http://localhost:1234/v1",
      "upstream_model": "your-model-name"
    }
  }
}
```

---

## Debug Mode

To see detailed RLM operation logs, set log level to debug:

```bash
# Set log level
export PICOCLAW_LOG_LEVEL=debug

# Run PicoClaw
picoclaw agent

# You'll see logs like:
# [DEBUG] RLM: Starting RLMgw server...
# [DEBUG] RLM: Server ready at http://127.0.0.1:8010
# [DEBUG] RLM: Sending request with 3 messages
# [DEBUG] RLM: Context selection took 1.2s
# [DEBUG] RLM: Received response with 245 tokens
```

---

## Security Considerations

### Localhost Only

RLMgw server binds to `127.0.0.1` by default and is **not exposed** to the network. This ensures:
- ✅ No external access to your workspace
- ✅ No network exposure of sensitive code
- ✅ Standard subprocess security model

### Workspace Access

RLMgw has **read-only access** to your workspace:
- ✅ Cannot modify files
- ✅ Cannot execute code
- ✅ Only reads files for context analysis

### Subprocess Security

The RLMgw subprocess:
- Runs with **same user permissions** as PicoClaw
- **No privilege escalation**
- **Isolated process** (can be killed independently)
- **Standard input/output**, no shell execution

---

## Performance Optimization Tips

### 1. Optimize Workspace Size

RLM analyzes your workspace for context. Smaller, focused workspaces = faster:

```json
{
  "agents": {
    "defaults": {
      "workspace": "~/.picoclaw/workspace/my-project",
      "restrict_to_workspace": true
    }
  }
}
```

### 2. Tune Recursion Depth

Balance between context quality and speed:

- **Fast mode** (`max_internal_calls: 1`): 1-2s latency, basic context
- **Balanced** (`max_internal_calls: 3`): 2-3s latency, good context (default)
- **Deep** (`max_internal_calls: 5`): 3-5s latency, comprehensive context

### 3. Adjust Context Size

Smaller context = faster processing:

```json
{
  "providers": {
    "rlm": {
      "max_context_pack_chars": 8000
    }
  }
}
```

### 4. Use Appropriate Upstream Model

- **Local models**: Lower latency, but may need more RLM calls
- **Cloud models**: Higher baseline latency, but better context understanding
- **Larger context models**: Reduce need for aggressive RLM pruning

### 5. Workspace Organization

Structure your workspace for better RLM analysis:

```
~/.picoclaw/workspace/
├── current-project/       # Active work
│   ├── src/
│   └── docs/
├── archive/              # Old projects (RLM will skip)
└── reference/            # Documentation (RLM can find)
```

---

## Comparison: RLM vs Direct Provider

| Scenario | Direct Provider | RLM Provider |
|----------|----------------|--------------|
| **Small contexts** (<4k tokens) | ✅ Fast, direct | ⚠️ Slower, unnecessary overhead |
| **Medium contexts** (4k-16k tokens) | ⚠️ Works, but uses full context | ✅ Optimizes context, faster upstream |
| **Large contexts** (16k-32k tokens) | ❌ May hit limits, slow | ✅ Selects relevant context only |
| **Very large contexts** (>32k tokens) | ❌ Context window errors | ✅ Handles seamlessly |

### When to Use RLM

✅ **Use RLM when:**
- Working with large codebases (>50 files)
- Frequently hitting context window limits
- Need to reference many files in one query
- Project has >10k lines of code
- Using local models with smaller context windows

❌ **Don't use RLM when:**
- Working with small projects (<10 files)
- Performing simple queries
- Context always fits in window
- Need minimal latency (real-time chat)
- Asking questions unrelated to workspace

---

## Upgrading RLMgw

To update to the latest version of RLMgw:

```bash
cd ~/rlmgw
git pull origin main
uv sync

# Restart PicoClaw agent
# RLMgw subprocess will use updated code
```

---

## Uninstalling RLM Integration

To disable RLM without removing code:

```json
{
  "agents": {
    "defaults": {
      "provider": "openai"  // Switch back to direct provider
    }
  },
  "providers": {
    "rlm": {
      "enabled": false
    }
  }
}
```

To completely remove:

```bash
# Remove rlmgw
rm -rf ~/rlmgw

# Remove RLM config from ~/.picoclaw/config.json
# PicoClaw will work normally with other providers
```

---

## FAQ

### Q: Does RLM work offline?

**A:** Yes, if you use a local upstream provider (LM Studio, vLLM, Ollama). RLMgw runs locally and only needs network access to communicate with the upstream provider.

### Q: Can I use multiple agents with different RLM configurations?

**A:** Currently, RLM is configured globally. All agents using `provider: rlm` will share the same configuration. Future versions may support per-agent RLM configs.

### Q: Does RLM support streaming responses?

**A:** Not yet. Streaming support is planned for a future release. Currently, responses are returned after complete generation.

### Q: Can I use RLM with function calling / tool use?

**A:** Yes! RLM fully supports OpenAI-compatible function calling. Tools are passed through transparently to the upstream provider.

### Q: Will RLM work if I switch between projects?

**A:** Yes. RLM analyzes the `workspace_root` configured for the agent. If you have different agents for different projects, each will have its own workspace context.

### Q: Does RLM store my code or send it anywhere?

**A:** No. RLM runs entirely locally on your machine. Your code stays on your computer. Only the selected context is sent to your configured upstream provider (which you control).

### Q: What happens if RLMgw crashes?

**A:** PicoClaw will detect the crash and return an error. You can restart RLMgw by restarting PicoClaw or by sending another request (RLM provider will attempt to restart the subprocess).

---

## Examples

### Example 1: Analyzing Large Codebase

```bash
picoclaw agent

> Can you give me an overview of how authentication works in this project?
< [RLM explores auth-related files]
< The authentication system has three main components:
< 1. OAuth flow in pkg/auth/oauth.go
< 2. Token management in pkg/auth/token.go
< 3. PKCE verification in pkg/auth/pkce.go
< ...

> Show me how the OAuth callback is handled
< [RLM focuses on oauth.go callback logic]
< The OAuth callback is handled in the HandleCallback function...
```

### Example 2: Cross-File Analysis

```bash
> How does the agent loop interact with the message bus?

< [RLM analyzes pkg/agent/loop.go and pkg/bus/bus.go]
< The agent loop publishes messages to the bus at three points:
< 1. On user input (agent.MessageReceived event)
< 2. After LLM response (agent.ResponseGenerated event)
< 3. On tool execution (agent.ToolCalled event)
< ...
```

### Example 3: Finding Related Code

```bash
> Where is the session state managed?

< [RLM searches across multiple packages]
< Session state is managed in two places:
< 1. In-memory cache: pkg/session/store.go
< 2. Persistent storage: pkg/state/db.go
< The Session interface defined in pkg/session/session.go...
```

---

## References

- **RLMgw Repository**: https://github.com/mitkox/rlmgw
- **RLM Paper**: https://arxiv.org/abs/2512.24601
- **RLM Blog Post**: https://alexzhang13.github.io/blog/2025/rlm/
- **PicoClaw Documentation**: https://github.com/sipeed/picoclaw/tree/main/docs

---

## Support

For issues or questions:

1. **Check logs**: Set `PICOCLAW_LOG_LEVEL=debug` for detailed output
2. **Check this guide**: Most common issues are covered in Troubleshooting
3. **File an issue**: https://github.com/sipeed/picoclaw/issues
4. **RLMgw issues**: https://github.com/mitkox/rlmgw/issues

---

**Document Version**: 1.0.0  
**Last Updated**: 2026-02-26  
**Applies to**: PicoClaw v0.1.0+
