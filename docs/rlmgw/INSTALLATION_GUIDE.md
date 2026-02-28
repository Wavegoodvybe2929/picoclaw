# Installation and Usage Guide

This guide covers installation and usage for both components in this repository:
1. **RLM (Recursive Language Models)** - Core inference engine
2. **RLMgw (RLM Gateway)** - OpenAI-compatible HTTP gateway for intelligent code context selection

## Table of Contents
- [Prerequisites](#prerequisites)
- [Installation](#installation)
  - [Basic Installation (RLM Core)](#basic-installation-rlm-core)
  - [Optional Dependencies](#optional-dependencies)
  - [Development Setup](#development-setup)
- [Quick Start: RLM Core](#quick-start-rlm-core)
- [RLM Usage Guide](#rlm-usage-guide)
  - [Basic Completion](#basic-completion)
  - [REPL Environments](#repl-environments)
  - [Model Providers](#model-providers)
  - [Logging and Visualization](#logging-and-visualization)
- [RLMgw Gateway Setup](#rlmgw-gateway-setup)
- [RLMgw Usage](#rlmgw-usage)
- [Configuration Reference](#configuration-reference)
- [Troubleshooting](#troubleshooting)

---

## Prerequisites

### Required
- **Python 3.11 or higher** (Python 3.12 recommended)
- **uv** package manager (recommended) or pip
- **API key** from a supported LLM provider:
  - OpenAI (`OPENAI_API_KEY`)
  - Anthropic (`ANTHROPIC_API_KEY`)
  - Portkey (`PORTKEY_API_KEY`)
  - Or other compatible providers

### Optional (for specific features)
- **Docker** (for DockerREPL environment)
- **Modal account** (for Modal sandboxes)
- **Node.js** (for trajectory visualizer)
- **vLLM server** (for RLMgw gateway)

---

## Installation

### Basic Installation (RLM Core)

#### Using uv (Recommended)

```bash
# Install uv if not already installed
curl -LsSf https://astral.sh/uv/install.sh | sh

# If you already have the repository locally:
cd /path/to/rlmgw  # Navigate to your existing repo

# OR if you need to clone it:
# git clone https://github.com/alexzhang13/rlm.git
# cd rlm

# Create virtual environment
uv init && uv venv --python 3.12
source .venv/bin/activate  # On Windows: .venv\Scripts\activate

# Install in editable mode
uv pip install -e .
```

#### Using pip

```bash
# Navigate to the repository directory
cd /path/to/rlmgw  # Use your actual repo path

# Create virtual environment
python3 -m venv .venv
source .venv/bin/activate  # On Windows: .venv\Scripts\activate

# Install in editable mode
pip install -e .
```

### Optional Dependencies

#### Modal Sandboxes (Cloud Isolation)

```bash
# Install Modal extra
uv pip install -e ".[modal]"

# Authenticate Modal account
modal setup
```

#### RLMgw Gateway (OpenAI-compatible API)

```bash
# Install gateway dependencies
uv pip install -e ".[gw]"
```

#### Development Tools

```bash
# Install dev + test dependencies
uv sync --group dev --group test

# Install pre-commit hooks
uv run pre-commit install
```

### Development Setup

For contributors:

```bash
# Install all optional dependencies
uv sync --group dev --group test

# Run code formatting
uv run ruff format .

# Run linting
uv run ruff check --fix .

# Run tests
uv run pytest
```

---

## Quick Start: RLM Core

### 1. Set Up Environment Variables

Create a `.env` file in your project root:

```bash
# .env
OPENAI_API_KEY=sk-...
# Or use other providers:
# ANTHROPIC_API_KEY=sk-ant-...
# PORTKEY_API_KEY=...
```

### 2. Run the Quickstart Example

```bash
uv run examples/quickstart.py
```

This will:
- Execute an RLM completion using your configured LLM
- Print execution logs to console
- Save trajectory logs to `./logs/` directory

### 3. Basic Python Usage

```python
import os
from dotenv import load_dotenv
from rlm import RLM

load_dotenv()

# Create RLM instance
rlm = RLM(
    backend="openai",
    backend_kwargs={
        "api_key": os.getenv("OPENAI_API_KEY"),
        "model_name": "gpt-4o",
    },
    verbose=True,  # Enable console output
)

# Make a completion call
result = rlm.completion("Print the first 100 powers of two, each on a newline.")
print(result.response)
```

---

## RLM Usage Guide

### Basic Completion

The core RLM pattern replaces standard LLM completion calls:

```python
from rlm import RLM

rlm = RLM(
    backend="openai",          # LLM provider
    backend_kwargs={"model_name": "gpt-4o"},
    environment="local",        # REPL environment
    max_depth=5,               # Max recursive depth
    max_iterations=10,         # Max iterations per depth
)

result = rlm.completion("Your prompt here")
print(result.response)
```

### REPL Environments

RLM supports multiple execution environments:

#### Local (Default)
Runs in the same process with limited isolation:

```python
rlm = RLM(
    environment="local",
    environment_kwargs={},
)
```

**Use when**: Quick testing, low-risk tasks
**Security**: Limited (same process)

#### Docker
Runs in Docker container:

```python
rlm = RLM(
    environment="docker",
    environment_kwargs={
        "image": "python:3.11-slim",  # Optional, default shown
    },
)
```

**Use when**: Need isolation but local execution
**Requirements**: Docker installed and running

#### Modal Sandboxes
Runs in cloud-based sandboxes:

```bash
# First-time setup
uv pip install -e ".[modal]"
modal setup
```

```python
rlm = RLM(
    environment="modal",
    environment_kwargs={},
)
```

**Use when**: Maximum isolation, production workloads
**Requirements**: Modal account + API key

#### Prime Intellect Sandboxes
⚠️ **Currently in beta** (not yet fully implemented)

```bash
export PRIME_API_KEY=...
```

```python
rlm = RLM(
    environment="prime",
    environment_kwargs={},
)
```

### Model Providers

#### OpenAI

```python
rlm = RLM(
    backend="openai",
    backend_kwargs={
        "api_key": os.getenv("OPENAI_API_KEY"),
        "model_name": "gpt-4o",
    },
)
```

#### Anthropic

```python
rlm = RLM(
    backend="anthropic",
    backend_kwargs={
        "api_key": os.getenv("ANTHROPIC_API_KEY"),
        "model_name": "claude-3-5-sonnet-20241022",
    },
)
```

#### Portkey (Multi-provider)

```python
rlm = RLM(
    backend="portkey",
    backend_kwargs={
        "api_key": os.getenv("PORTKEY_API_KEY"),
        "model_name": "gpt-4o",
    },
)
```

#### LiteLLM (Router)

```python
rlm = RLM(
    backend="litellm",
    backend_kwargs={
        "model_name": "gpt-4o",
    },
)
```

#### Local Models (via vLLM)

```python
# Start vLLM server first:
# vllm serve <model> --port 8000

rlm = RLM(
    backend="openai",  # vLLM is OpenAI-compatible
    backend_kwargs={
        "api_key": "EMPTY",
        "base_url": "http://localhost:8000/v1",
        "model_name": "your-model-name",
    },
)
```

### Logging and Visualization

#### Enable Logging

```python
from rlm import RLM
from rlm.logger import RLMLogger

logger = RLMLogger(log_dir="./logs")

rlm = RLM(
    backend="openai",
    backend_kwargs={"model_name": "gpt-4o"},
    logger=logger,
    verbose=True,  # Console output with rich formatting
)

result = rlm.completion("Your prompt")
# Logs saved to ./logs/<timestamp>.jsonl
```

#### Visualize Trajectories

```bash
cd visualizer/
npm install
npm run dev  # Starts on localhost:3001
```

Then:
1. Open http://localhost:3001 in your browser
2. Upload a `.jsonl` log file from `./logs/`
3. Explore the execution trajectory, code, and LLM calls

---

## RLMgw Gateway Setup

RLMgw sits between Claude Code (or other OpenAI-compatible clients) and a vLLM server, intelligently selecting relevant code context using RLM.

### Architecture

```
Claude Code
  ↓ OpenAI-compatible request
RLMgw Server (FastAPI)
  ↓ RLM-based context selection
  ↓ Inject context into system message
vLLM Server (e.g., MiniMax-M2.1)
  ↓
Response → Claude Code
```

### Prerequisites

1. **vLLM server** running (e.g., with MiniMax-M2.1 or other model)
2. **Target repository** to analyze
3. **RLMgw dependencies** installed

### Installation

```bash
# Install gateway dependencies
uv pip install -e ".[gw]"
```

### Start vLLM Server (Example)

```bash
# Install vLLM (if not already installed)
pip install vllm

# Start vLLM with MiniMax-M2.1 (example)
vllm serve mistralai/Mixtral-8x7B-Instruct-v0.1 --port 8000
# Or use your preferred model
```

### Start RLMgw Server

```bash
# Basic usage
python3 -m rlmgw.server \
  --host 127.0.0.1 \
  --port 8010 \
  --repo-root /path/to/target-repo \
  --upstream-base-url http://localhost:8000/v1 \
  --upstream-model your-model-name

# With RLM context selection enabled (default)
python3 -m rlmgw.server \
  --host 0.0.0.0 \
  --port 8010 \
  --repo-root /path/to/your/codebase \
  --use-rlm-context-selection \
  --max-internal-calls 3 \
  --max-context-pack-chars 12000
```

### Environment Variables (Alternative)

```bash
# RLMgw configuration
export RLMGW_HOST="0.0.0.0"
export RLMGW_PORT="8010"
export RLMGW_REPO_ROOT="/path/to/target-repo"
export RLMGW_UPSTREAM_BASE_URL="http://localhost:8000/v1"
export RLMGW_UPSTREAM_MODEL="minimax-m2-1"

# Context selection
export RLMGW_USE_RLM_CONTEXT_SELECTION="true"
export RLMGW_MAX_INTERNAL_CALLS="3"
export RLMGW_MAX_CONTEXT_PACK_CHARS="12000"

# Session management
export RLMGW_SESSION_TTL_HOURS="24"
export RLMGW_MAX_SESSIONS="50"
export RLMGW_STORAGE_DIR=".rlmgw"

# Run server
python3 -m rlmgw.server
```

---

## RLMgw Usage

### Test with curl

```bash
curl -X POST http://localhost:8010/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "minimax-m2-1",
    "messages": [
      {"role": "user", "content": "What authentication methods are used in this repo?"}
    ]
  }'
```

### Health Checks

```bash
# Check server status
curl http://localhost:8010/health

# Check upstream vLLM connection
curl http://localhost:8010/readyz
```

### Use with Claude Code

Configure Claude Code to use RLMgw as a custom model:

1. In Claude Code settings, add custom model endpoint:
   - **Base URL**: `http://localhost:8010/v1`
   - **Model**: `minimax-m2-1` (or your vLLM model name)

2. RLMgw will automatically:
   - Analyze user queries
   - Select relevant code context from target repo
   - Inject context into system message
   - Forward to vLLM for completion

### Context Selection Modes

#### RLM Mode (Default)
Intelligently explores repo using RLM:

```bash
export RLMGW_USE_RLM_CONTEXT_SELECTION="true"
export RLMGW_MAX_INTERNAL_CALLS="3"
```

**How it works**:
1. User query analyzed by RLM
2. RLM uses `repo.grep()`, `repo.read_file()`, `repo.list_files()`
3. Recursively refines file selection
4. Returns context pack with relevant files

**Best for**: Large codebases, complex queries

#### Simple Mode (Fallback)
Keyword-based grep search:

```bash
export RLMGW_USE_RLM_CONTEXT_SELECTION="false"
```

**How it works**:
1. Extract keywords from query
2. Grep for keywords in repo
3. Include common project files (README, etc.)

**Best for**: Small codebases, quick responses

### Session Management

RLMgw caches context packs per session to avoid reprocessing:

```bash
# Configure sessions
export RLMGW_SESSION_TTL_HOURS="24"
export RLMGW_MAX_SESSIONS="50"
export RLMGW_STORAGE_DIR=".rlmgw"
```

Session storage location: `.rlmgw/sessions.db`

---

## Configuration Reference

### RLM Core Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `backend` | str | `"openai"` | LLM provider: `openai`, `anthropic`, `portkey`, `litellm` |
| `backend_kwargs` | dict | `{}` | Provider-specific config (API key, model name, etc.) |
| `environment` | str | `"local"` | REPL environment: `local`, `docker`, `modal`, `prime` |
| `environment_kwargs` | dict | `{}` | Environment-specific config |
| `max_depth` | int | `5` | Maximum recursive depth |
| `max_iterations` | int | `10` | Max iterations per depth level |
| `verbose` | bool | `False` | Enable console output with rich formatting |
| `logger` | RLMLogger | `None` | Optional logger for trajectory logging |

### RLMgw Configuration

| Environment Variable | CLI Argument | Default | Description |
|---------------------|--------------|---------|-------------|
| `RLMGW_HOST` | `--host` | `127.0.0.1` | Server bind address |
| `RLMGW_PORT` | `--port` | `8010` | Server port |
| `RLMGW_REPO_ROOT` | `--repo-root` | `.` | Target repository path |
| `RLMGW_UPSTREAM_BASE_URL` | `--upstream-base-url` | Required | vLLM server URL (e.g., `http://localhost:8000/v1`) |
| `RLMGW_UPSTREAM_MODEL` | `--upstream-model` | Required | Model name in vLLM |
| `RLMGW_USE_RLM_CONTEXT_SELECTION` | `--use-rlm-context-selection` | `true` | Enable RLM-based context selection |
| `RLMGW_MAX_INTERNAL_CALLS` | `--max-internal-calls` | `3` | Max RLM recursive calls for context selection |
| `RLMGW_MAX_CONTEXT_PACK_CHARS` | `--max-context-pack-chars` | `12000` | Max context pack size |
| `RLMGW_SESSION_TTL_HOURS` | `--session-ttl-hours` | `24` | Session cache TTL |
| `RLMGW_MAX_SESSIONS` | `--max-sessions` | `50` | Max cached sessions |
| `RLMGW_STORAGE_DIR` | `--storage-dir` | `.rlmgw` | Storage directory |

---

## Troubleshooting

### RLM Core Issues

#### Import Errors

```python
# Error: ModuleNotFoundError: No module named 'rlm'
```

**Solution**: Ensure you've installed in editable mode:
```bash
uv pip install -e .
```

#### API Key Issues

```python
# Error: OpenAI API key not found
```

**Solution**: Set environment variable or pass explicitly:
```bash
export OPENAI_API_KEY=sk-...
```

Or in code:
```python
rlm = RLM(
    backend="openai",
    backend_kwargs={"api_key": "sk-..."},
)
```

#### Docker REPL Not Working

**Check Docker is running**:
```bash
docker --version
docker ps
```

**Check Docker daemon**:
```bash
# macOS: Ensure Docker Desktop is running
# Linux: sudo systemctl start docker
```

#### Modal Setup Issues

```bash
# Error: Modal not authenticated
```

**Solution**:
```bash
modal setup
# Follow authentication prompts
```

### RLMgw Issues

#### vLLM Connection Failed

```bash
curl http://localhost:8010/readyz
# Returns: {"ready": false, "upstream_error": "..."}
```

**Solutions**:
1. Check vLLM is running: `curl http://localhost:8000/v1/models`
2. Verify `RLMGW_UPSTREAM_BASE_URL` is correct
3. Check network/firewall settings

#### Context Selection Taking Too Long

**Reduce RLM internal calls**:
```bash
export RLMGW_MAX_INTERNAL_CALLS="1"  # Faster but less thorough
```

**Or use simple mode**:
```bash
export RLMGW_USE_RLM_CONTEXT_SELECTION="false"
```

#### Session Database Growing Large

**Clear old sessions**:
```bash
rm -rf .rlmgw/sessions.db
# Or adjust limits:
export RLMGW_SESSION_TTL_HOURS="6"
export RLMGW_MAX_SESSIONS="20"
```

#### RLM Context Selection Not Available

```bash
# Warning: RLM context pack builder not available
```

**Solution**: Ensure all dependencies are installed:
```bash
uv pip install -e ".[gw]"
```

### Code Quality Issues

#### Linting Failures

```bash
uv run ruff check --fix .
uv run ruff format .
```

#### Test Failures

```bash
# Run specific test
uv run pytest tests/test_imports.py -v

# Run with verbose output
uv run pytest -vv

# Check test coverage
uv run pytest --cov=rlm tests/
```

---

## Additional Resources

- **RLM Paper**: https://arxiv.org/abs/2512.24601
- **RLM Blog Post**: https://alexzhang13.github.io/blog/2025/rlm/
- **Documentation**: https://alexzhang13.github.io/rlm/
- **GitHub**: https://github.com/alexzhang13/rlm
- **RLM Minimal**: https://github.com/alexzhang13/rlm-minimal

## Getting Help

- **GitHub Issues**: https://github.com/alexzhang13/rlm/issues
- **Contributing**: See [CONTRIBUTING.md](CONTRIBUTING.md)
- **Agent Guidelines**: See [AGENTS.md](AGENTS.md) for development guidelines

---

## Citation

If you use this repository in your research, please cite:

```bibtex
@misc{zhang2025recursivelanguagemodels,
      title={Recursive Language Models}, 
      author={Alex L. Zhang and Tim Kraska and Omar Khattab},
      year={2025},
      eprint={2512.24601},
      archivePrefix={arXiv},
      primaryClass={cs.AI},
      url={https://arxiv.org/abs/2512.24601}, 
}
```
