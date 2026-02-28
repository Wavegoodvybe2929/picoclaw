# RLM Integration Guide for PicoClaw

> **Accurate as of**: February 26, 2026  
> **PicoClaw Version**: 0.1.0+  
> **Implementation**: Based on actual `pkg/providers/rlm_provider.go`

This guide explains how to integrate RLMgw (RLM Gateway) with PicoClaw to enable intelligent code context selection for AI agents working with large codebases.

---

## Table of Contents
- [Overview](#overview)
- [Architecture](#architecture)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
- [Advanced Configuration](#advanced-configuration)
- [Troubleshooting](#troubleshooting)
- [FAQ](#faq)

---

## Overview

### What is RLM?

RLM (Recursive Language Models) is a technique that enables LLMs to recursively explore and select only the most relevant context from large datasets. This allows handling contexts far beyond typical token limits.

### What is RLMgw?

RLMgw is a Python-based OpenAI-compatible HTTP gateway that uses RLM to intelligently select code context before forwarding requests to an upstream LLM provider.

### How PicoClaw Uses RLMgw

**Key Architectural Detail**: PicoClaw spawns RLMgw as a **local subprocess**, not a remote service. When you configure the RLM provider in PicoClaw:

1. PicoClaw starts RLMgw as a subprocess using `python3 -m rlmgw.main`
2. RLMgw binds to `127.0.0.1:8010` (localhost only)
3. PicoClaw communicates with RLMgw via HTTP on localhost
4. RLMgw analyzes your workspace and selects relevant context
5. RLMgw forwards requests with context to your upstream provider
6. When PicoClaw exits, RLMgw subprocess is gracefully terminated

**Benefits**:
- ✅ **Simple Setup**: Just configure one provider in PicoClaw config
- ✅ **Automatic Management**: PicoClaw handles subprocess lifecycle
- ✅ **Secure**: Localhost-only communication, no network exposure
- ✅ **Resource Efficient**: Subprocess shares system with PicoClaw
- ✅ **Works with Any Provider**: Upstream can be local or cloud-based

---

## Architecture

```
┌──────────────────────────────────────────────────────────────┐
│  Your Machine (where PicoClaw runs)                          │
│                                                               │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │  PicoClaw Process (Go)                                  │ │
│  │                                                         │ │
│  │  User Query → Agent Loop → RLM Provider                │ │
│  │                               ↓                         │ │
│  │                         Spawns subprocess               │ │
│  │                               ↓                         │ │
│  └───────────────────────────────┼──────────────────────────┘ │
│                                  │                            │
│                    HTTP  (localhost:8010)                     │
│                                  ↓                            │
│  ┌───────────────────────────────────────────────────────┐   │
│  │  RLMgw Subprocess (Python)                            │   │
│  │  • Analyzes workspace                                 │   │
│  │  • Selects relevant files                             │   │
│  │  • Packs context                                      │   │
│  └───────────────────┬───────────────────────────────────┘   │
│                      │                                        │
│                      │ HTTP to upstream provider              │
│                      ↓                                        │
└──────────────────────┼────────────────────────────────────────┘
                       │
           ┌───────────┴────────────┐
           │   Upstream Provider    │
           │  (Local or Remote)     │
           │                        │
           │  • LM Studio (local)   │
           │  • vLLM (local)        │
           │  • OpenAI (cloud)      │
           │  • Anthropic (cloud)   │
           │  • Groq (cloud)        │
           └────────────────────────┘
```

### Request Flow

1. **User Query** → PicoClaw agent
2. **Agent Loop** → Calls RLM Provider
3. **RLM Provider** → Sends HTTP request to RLMgw subprocess (localhost:8010)
4. **RLMgw Subprocess**:
   - Analyzes `workspace_root` for relevant context
   - Uses RLM algorithms to recursively select files
   - Packs selected code into context
5. **RLMgw** → Forwards request + context to upstream provider
6. **Upstream Provider** → Processes with full context, returns response
7. **Response** → RLMgw → PicoClaw → User

---

## Prerequisites

### Required
- **Python 3.11+** (required for RLMgw subprocess)
- **PicoClaw** installed and running
- **uv** or pip (Python package manager)

### Optional
- **Local LLM**: LM Studio, vLLM, or Ollama
- **OR Cloud API**: OpenAI, Anthropic, Groq, etc.

---

## Installation

### Step 1: Install RLMgw

PicoClaw expects RLMgw at `~/rlmgw` by default.

```bash
# Clone to default location
cd ~
git clone https://github.com/mitkox/rlmgw.git
cd rlmgw

# Install uv (if not installed)
curl -LsSf https://astral.sh/uv/install.sh | sh
export PATH="$HOME/.local/bin:$PATH"

# Install dependencies
uv sync

# Verify installation
python3 -m rlmgw.main --help
```

### Step 2: Set Up Upstream Provider (Choose One)

#### Option A: LM Studio (Local)

1. Download and install [LM Studio](https://lmstudio.ai)
2. Load a model (e.g., Qwen2.5-Coder, Llama-3, etc.)
3. Start server (default port: 1234)
4. Note the URL: `http://localhost:1234/v1`

#### Option B: vLLM (Local)

```bash
pip install vllm
vllm serve mistralai/Mixtral-8x7B-Instruct-v0.1 --port 8000
# URL: http://localhost:8000/v1
```

#### Option C: Cloud Provider (OpenAI, Anthropic, etc.)

- Get API key from provider
- Note the baseURL (e.g., `https://api.openai.com/v1`)

### Step 3: Configure PicoClaw

Edit `~/.picoclaw/config/config.json`:

```json
{
  "agents": {
    "defaults": {
      "provider": "rlm",
      "model_name": "model-at-upstream"
    }
  },
  "providers": {
    "rlm": {
      "enabled": true,
      "python_path": "",
      "rlmgw_path": "~/rlmgw",
      "host": "127.0.0.1",
      "port": 8010,
      "upstream_base_url": "http://localhost:1234/v1",
      "upstream_model": "model-at-upstream",
      "workspace_root": "~/.picoclaw/workspace",
      "use_rlm_selection": true,
      "max_internal_calls": 3,
      "max_context_pack_chars": 12000
    }
  }
}
```

### Step 4: Test

```bash
picoclaw agent

# You should see:
# [INFO] RLMgw subprocess started pid=12345 url=http://127.0.0.1:8010
# [INFO] RLMgw server ready url=http://127.0.0.1:8010

> Tell me about the authentication system
< [Response with intelligent context from your workspace]
```

---

## Configuration

### Configuration Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `enabled` | boolean | Yes | `false` | Enable RLM provider |
| `python_path` | string | No | auto-detect | Path to Python 3.11+ (blank = auto-detect) |
| `rlmgw_path` | string | No | `~/rlmgw` | Path to RLMgw installation |
| `host` | string | No | `127.0.0.1` | **Must be localhost** for security |
| `port` | integer | No | `8010` | RLMgw subprocess port |
| `upstream_base_url` | string | Yes | - | Upstream provider endpoint |
| `upstream_model` | string | Yes | - | Model name at upstream |
| `workspace_root` | string | No | agent workspace | Workspace to analyze for context |
| `use_rlm_selection` | boolean | No | `false` | Use intelligent RLM context selection |
| `max_internal_calls` | integer | No | `3` | RLM recursion depth (1-5) |
| `max_context_pack_chars` | integer | No | `12000` | Max context size in characters |

### Environment Variables Set by PicoClaw

PicoClaw automatically sets these when starting the RLMgw subprocess:

```bash
RLMGW_HOST=127.0.0.1              # From config.providers.rlm.host
RLMGW_PORT=8010                   # From config.providers.rlm.port
RLMGW_REPO_PATH=/workspace/path   # From config.providers.rlm.workspace_root
RLMGW_UPSTREAM_BASE_URL=...       # From config.providers.rlm.upstream_base_url
RLMGW_UPSTREAM_MODEL=...          # From config.providers.rlm.upstream_model
RLMGW_USE_RLM_CONTEXT_SELECTION=true  # From config.providers.rlm.use_rlm_selection
RLMGW_MAX_INTERNAL_CALLS=3        # From config.providers.rlm.max_internal_calls
RLMGW_MAX_CONTEXT_PACK_CHARS=12000    # From config.providers.rlm.max_context_pack_chars
```

**Note**: You do NOT need to set these manually. PicoClaw handles this automatically.

---

## Usage

### Basic Usage

Once configured, just use PicoClaw normally. RLMgw runs transparently:

```bash
picoclaw agent

> Explain how authentication works in this codebase
< The authentication system has three components:
< 1. OAuth flow in pkg/auth/oauth.go
< 2. Token management in pkg/auth/token.go
< 3. PKCE verification in pkg/auth/pkce.go
< ...

> Where is the login function called?
< The login function is called in 5 locations:
< - pkg/handlers/user.go:45
< - pkg/api/auth.go:89
< ...
```

### Example Workflows

#### Code Navigation

```bash
> What does pkg/agent/loop.go do?
< [RLMgw loads file with surrounding context]
< This file implements the main agent loop...
```

#### Cross-File Analysis

```bash
> How does the message bus interact with agents?
< [RLMgw analyzes multiple files]
< The bus publishes events at three points:
< 1. agent.MessageReceived (loop.go:123)
< 2. agent.ResponseGenerated (loop.go:234)
< ...
```

#### Bug Investigation

```bash
> Find all places where we validate user input
< [RLMgw searches for patterns]
< Input validation occurs in:
< - pkg/validation/sanitize.go (XSS protection)
< - pkg/handlers/user.go (form validation)
< ...
```

---

## Advanced Configuration

### Different Upstream Providers

#### LM Studio

```json
{
  "providers": {
    "rlm": {
      "enabled": true,
      "upstream_base_url": "http://localhost:1234/v1",
      "upstream_model": "qwen2.5-coder-7b",
      "use_rlm_selection": true
    }
  }
}
```

#### vLLM

```json
{
  "providers": {
    "rlm": {
      "enabled": true,
      "upstream_base_url": "http://localhost:8000/v1",
      "upstream_model": "mistralai/Mixtral-8x7B-Instruct-v0.1",
      "use_rlm_selection": true
    }
  }
}
```

#### OpenAI

```json
{
  "providers": {
    "rlm": {
      "enabled": true,
      "upstream_base_url": "https://api.openai.com/v1",
      "upstream_model": "gpt-4o",
      "use_rlm_selection": true
    }
  }
}
```

**Note**: Make sure you have `OPENAI_API_KEY` environment variable set.

### Performance Tuning

#### Fast Mode (Lower Latency)

```json
{
  "providers": {
    "rlm": {
      "use_rlm_selection": true,
      "max_internal_calls": 1,
      "max_context_pack_chars": 8000
    }
  }
}
```

#### Deep Mode (Better Context)

```json
{
  "providers": {
    "rlm": {
      "use_rlm_selection": true,
      "max_internal_calls": 5,
      "max_context_pack_chars": 16000
    }
  }
}
```

#### Simple Mode (No RLM, Keyword-Based)

```json
{
  "providers": {
    "rlm": {
      "use_rlm_selection": false,
      "max_context_pack_chars": 8000
    }
  }
}
```

### Custom RLMgw Location

```json
{
  "providers": {
    "rlm": {
      "enabled": true,
      "rlmgw_path": "/opt/rlmgw",
      "python_path": "/usr/local/bin/python3.12"
    }
  }
}
```

---

## Troubleshooting

### Issue: "python3 not found in PATH"

**Cause**: Python 3.11+ not installed or not in PATH

**Solution**:
```bash
# macOS (Homebrew)
brew install python@3.11

# Ubuntu/Debian
sudo apt install python3.11

# OR specify path in config
{
  "providers": {
    "rlm": {
      "python_path": "/usr/local/bin/python3.11"
    }
  }
}
```

### Issue: "rlmgw directory not found at ~/rlmgw"

**Cause**: RLMgw not installed at expected location

**Solution**:
```bash
# Install at default location
cd ~
git clone https://github.com/mitkox/rlmgw.git
cd rlmgw
uv sync

# OR specify custom path
{
  "providers": {
    "rlm": {
      "rlmgw_path": "/path/to/your/rlmgw"
    }
  }
}
```

### Issue: "failed to start RLMgw server"

**Cause**: Missing Python dependencies

**Solution**:
```bash
cd ~/rlmgw
uv sync

# OR try with pip
pip install -r requirements.txt
```

### Issue: "timeout waiting for rlmgw server to be ready"

**Cause**: RLMgw subprocess failed to start or port already in use

**Check port**:
```bash
lsof -i :8010
# If in use, kill process or change port in config
```

**View RLMgw logs**:
PicoClaw captures RLMgw stderr. Check PicoClaw logs for error messages.

### Issue: "RLM provider is not enabled"

**Cause**: `enabled: false` in config

**Solution**:
```json
{
  "providers": {
    "rlm": {
      "enabled": true
    }
  }
}
```

### Issue: Slow responses

**Cause**: Too many recursive calls

**Solution**: Reduce `max_internal_calls`:
```json
{
  "providers": {
    "rlm": {
      "max_internal_calls": 1
    }
  }
}
```

### Issue: Context window errors

**Cause**: Context pack too large for upstream model

**Solution**: Reduce context size:
```json
{
  "providers": {
    "rlm": {
      "max_context_pack_chars": 8000
    }
  }
}
```

---

## FAQ

### Q: Can I use RLM with any LLM provider?

**A**: Yes! RLM works with any OpenAI-compatible endpoint (LM Studio, vLLM, Ollama, OpenAI, Anthropic, Groq, etc.).

### Q: Does RLM work offline?

**A**: Yes, if you use a local upstream provider (LM Studio, vLLM, Ollama).

### Q: Where does RLMgw run?

**A**: On the same machine as PicoClaw, as a subprocess. It's not a separate service.

### Q: Is my code sent to OpenAI?

**A**: Only if you configure OpenAI as the upstream provider. If you use LM Studio or vLLM, everything stays local.

### Q: Can I disable context selection?

**A**: Yes, set `use_rlm_selection: false` for simple keyword-based context selection.

### Q: What happens if RLMgw crashes?

**A**: PicoClaw will detect the crash and return an error. Restart PicoClaw to restart the subprocess.

### Q: Can multiple PicoClaw instances share one RLMgw?

**A**: No. Each PicoClaw instance spawns its own RLMgw subprocess.

### Q: How much memory does RLMgw use?

**A**: Approximately 50-150MB depending on workspace size and context cache.

### Q: Can I manually start RLMgw?

**A**: You don't need to. PicoClaw manages it automatically. Manual start is not supported.

---

## Performance Characteristics

| Operation | Latency | Notes |
|-----------|---------|-------|
| First request | +3-5s | One-time subprocess startup |
| Context selection (RLM enabled) | +1-3s | Varies with `max_internal_calls` |
| Context selection (RLM disabled) | +0.5-1s | Simple keyword search |
| Subsequent requests | +1-2s | RLM recursion overhead |

| Resource | Usage | Notes |
|----------|-------|-------|
| Memory (RLMgw subprocess) | 50-150MB | Varies with workspace size |
| Memory (context cache) | 10-50MB | Depends on cache settings |
| CPU usage | Low | Only during context selection |

---

## When to Use RLM

✅ **Use RLM when**:
- Working with large codebases (>50 files)
- Frequently hitting context window limits
- Need to reference many files in queries
- Project has >10k lines of code
- Using smaller local models

❌ **Don't use RLM when**:
- Working with small projects (<10 files)
- Performing simple queries
- Context always fits in window
- Need minimal latency
- Questions unrelated to code

---

## Security Considerations

### Localhost Only

RLMgw subprocess binds to `127.0.0.1` (localhost) and is **not exposed to the network**.

**Security guarantees**:
- ✅ No external access to workspace
- ✅ No network exposure of code
- ✅ Standard subprocess security model
- ✅ Same permissions as PicoClaw process

### Workspace Access

RLMgw has **read-only access** to `workspace_root`:
- ✅ Cannot modify files
- ✅ Cannot execute code
- ✅ Only reads files for context analysis

---

## References

- **RLMgw Repository**: https://github.com/mitkox/rlmgw
- **RLM Paper**: https://arxiv.org/abs/2512.24601
- **RLM Blog Post**: https://alexzhang13.github.io/blog/2025/rlm/
- **PicoClaw Documentation**: https://github.com/sipeed/picoclaw/tree/main/docs
- **Implementation**: `pkg/providers/rlm_provider.go`
- **Configuration**: `pkg/config/config.go` (RLMConfig struct)

---

## Support

For issues or questions:

1. **Check this guide**: Most common issues covered in [Troubleshooting](#troubleshooting)
2. **Check logs**: PicoClaw logs show RLMgw subprocess output
3. **File an issue**: https://github.com/sipeed/picoclaw/issues
4. **RLMgw issues**: https://github.com/mitkox/rlmgw/issues

---

**Document Version**: 2.0.0 (Accurate Implementation-Based)  
**Last Updated**: February 26, 2026  
**Applies to**: PicoClaw v0.1.0+  
**Based on**: Actual `pkg/providers/rlm_provider.go` implementation
