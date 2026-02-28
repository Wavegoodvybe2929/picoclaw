# Workspace Integration Guide

## Overview

PicoClaw can automatically integrate with workspace scripts and tools, enabling powerful automation and customization capabilities. This integration allows you to:

- Execute custom scripts at specific lifecycle points in the agent loop
- Inject context from external sources (memory systems, databases, APIs)
- Automate logging, metrics collection, and notifications
- Replace built-in tools with workspace-specific implementations

This guide explains how to configure and use workspace integration features effectively.

---

## Table of Contents

1. [Quick Start](#quick-start)
2. [Enabling Workspace Tools](#enabling-workspace-tools)
3. [Configuring Loop Hooks](#configuring-loop-hooks)
4. [Template Variables](#template-variables)
5. [Hook Lifecycle](#hook-lifecycle)
6. [Use Cases and Examples](#use-cases-and-examples)
7. [Best Practices](#best-practices)
8. [Troubleshooting](#troubleshooting)

---

## Quick Start

### 1. Enable Workspace Integration

Add to your `config.json`:

```json
{
  "agents": {
    "defaults": {
      "workspace": "~/.picoclaw/workspace",
      "use_workspace_tools": true
    }
  }
}
```

### 2. Create a Simple Hook Script

Create `~/.picoclaw/workspace/bin/memory_recall`:

```bash
#!/bin/bash
# Simple memory recall hook
query="$1"
echo "Recalling context for: $query"
# Your memory lookup logic here
```

Make it executable:
```bash
chmod +x ~/.picoclaw/workspace/bin/memory_recall
```

### 3. Configure the Hook

```json
{
  "agents": {
    "defaults": {
      "loop_hooks": {
        "before_llm": [
          {
            "name": "memory_recall",
            "command": "./bin/memory_recall '{user_message}'",
            "enabled": true,
            "inject_as": "context"
          }
        ]
      }
    }
  }
}
```

---

## Enabling Workspace Tools

### Configuration

Set `use_workspace_tools` to `true` to enable workspace tool discovery:

```json
{
  "agents": {
    "defaults": {
      "use_workspace_tools": true,
      "workspace": "/path/to/workspace"
    }
  }
}
```

### How It Works

When `use_workspace_tools` is enabled:

1. PicoClaw scans `<workspace>/tools/` directory for tool definitions
2. Workspace tools are loaded in addition to built-in tools
3. If a workspace tool has the same name as a built-in tool, the workspace version takes precedence

### Creating Workspace Tools

See [TOOLS.md](../workspace/TOOLS.md) for detailed instructions on creating custom workspace tools.

---

## Configuring Loop Hooks

Loop hooks allow you to execute custom scripts at specific points in the agent lifecycle.

### Hook Types

PicoClaw supports four types of hooks:

#### 1. `before_llm`
Executed **before** the LLM is called, after the user message is received.

**Use Cases:**
- Memory recall and context injection
- Pre-processing user input
- Loading session-specific data
- Adding system context

**Example:**
```json
{
  "before_llm": [
    {
      "name": "memory_recall",
      "command": "./bin/memory_recall --query '{user_message}' --format markdown",
      "enabled": true,
      "inject_as": "context",
      "metadata": {
        "description": "Recall relevant context from long-term memory"
      }
    }
  ]
}
```

#### 2. `after_response`
Executed **after** the agent generates a response and saves it to the session.

**Use Cases:**
- Memory storage (user and assistant messages)
- Logging and analytics
- Post-processing responses
- External notifications

**Example:**
```json
{
  "after_response": [
    {
      "name": "memory_write_user",
      "command": "./bin/memory_write --role user --content '{user_message}'",
      "enabled": true,
      "inject_as": "",
      "metadata": {
        "description": "Store user message in memory"
      }
    },
    {
      "name": "memory_write_assistant",
      "command": "./bin/memory_write --role assistant --content '{assistant_message}'",
      "enabled": true,
      "inject_as": ""
    }
  ]
}
```

#### 3. `on_tool_call`
Executed **after** each tool is called, with access to tool name, arguments, and result.

**Use Cases:**
- Tool usage logging and metrics
- Tool result validation
- Conditional notifications
- Tool-specific post-processing

**Example:**
```json
{
  "on_tool_call": [
    {
      "name": "log_tool_usage",
      "command": "./bin/log_tool --name '{tool_name}' --args '{tool_args}' --result '{tool_result}'",
      "enabled": true,
      "inject_as": "",
      "metadata": {
        "description": "Log all tool usage to analytics system"
      }
    }
  ]
}
```

#### 4. `on_error`
Executed when an error occurs during agent processing.

**Use Cases:**
- Error logging and alerting
- Fallback mechanisms
- Error recovery
- Debugging support

**Example:**
```json
{
  "on_error": [
    {
      "name": "notify_error",
      "command": "./bin/notify_error --message '{error}' --session '{session_key}'",
      "enabled": true,
      "inject_as": "",
      "metadata": {
        "description": "Send error notifications to monitoring system"
      }
    }
  ]
}
```

#### 5. `request_input`
Executed when the agent needs additional input from the user. These hooks **block execution** until the user responds or timeout expires.

**Use Cases:**
- Confirming destructive actions before execution
- Gathering missing parameters or clarification
- Interactive multi-step workflows
- User approval for sensitive operations

**Example:**
```json
{
  "request_input": [
    {
      "name": "confirm_action",
      "command": "./bin/format_prompt '{prompt_text}'",
      "enabled": true,
      "timeout": 120,
      "return_as": "user_confirmation",
      "default_value": "no",
      "metadata": {
        "description": "Request user confirmation for actions"
      }
    }
  ]
}
```

**Special Fields for `request_input` Hooks:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `timeout` | integer | No | Seconds to wait for user response (default: 60) |
| `return_as` | string | No | Variable name to store user's response |
| `default_value` | string | No | Value to use if timeout expires or error occurs |

**Important Notes:**
- The hook command should output the prompt text to show the user
- The agent will pause and wait for the user to respond
- If the user doesn't respond within `timeout` seconds, `default_value` is used
- The user's response is returned to the agent as the tool result
- Use the `request_input` tool to trigger these hooks (agent calls `request_input(prompt="Your question?")`)

**Example Workflow:**
```
1. Agent: "I need to confirm before deploying"
2. Agent calls: request_input(prompt="Deploy to production? (yes/no)")
3. Hook executes: ./bin/format_prompt "Deploy to production? (yes/no)"
4. User sees: "🤔 Deploy to production? (yes/no)\n\nPlease respond with your input."
5. User responds: "yes"
6. Agent receives: "User responded: yes"
7. Agent continues: "Confirmed. Starting deployment..."
```

### Hook Configuration Fields

Each hook requires the following fields:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Unique identifier for the hook |
| `command` | string | Yes | Shell command to execute (supports template variables) |
| `enabled` | boolean | Yes | Whether the hook is active |
| `inject_as` | string | No | How to inject hook output (`""` or `"context"`) |
| `metadata` | object | No | Additional metadata (description, author, etc.) |

### Context Injection

The `inject_as` field controls how hook output is used:

- **`inject_as: ""`** (empty): Hook output is discarded (use for side effects like logging)
- **`inject_as: "context"`**: Hook output is injected into LLM context (only for `before_llm` hooks)

**Example with context injection:**

Script output:
```
Relevant memories:
- User prefers concise responses
- Previous conversation about Go error handling
- User's timezone: UTC-8
```

This content is automatically added to the LLM's system prompt before processing the user message.

---

## Template Variables

Hook commands support template variable substitution. Available variables depend on the hook type:

### Available in All Hooks

| Variable | Description | Example |
|----------|-------------|---------|
| `{query}` | User's current message (alias for `user_message`) | "How do I configure hooks?" |
| `{user_message}` | User message content | "How do I configure hooks?" |
| `{session_key}` | Current session identifier | "session-123-abc" |
| `{channel}` | Channel name (e.g., "slack", "telegram") | "telegram" |
| `{chat_id}` | Chat identifier for the current conversation | "chat-456-def" |

### Available in `after_response` Hooks

| Variable | Description | Example |
|----------|-------------|---------|
| `{assistant_message}` | Agent's response to the user | "To configure hooks, add a..." |

### Available in `on_tool_call` Hooks

| Variable | Description | Example |
|----------|-------------|---------|
| `{tool_name}` | Name of the tool that was called | "code_search" |
| `{tool_args}` | Tool arguments as JSON string | `{"query":"hooks","max_results":5}` |
| `{tool_result}` | Tool result content (for LLM) | "Found 3 files matching..." |

### Available in `on_error` Hooks

| Variable | Description | Example |
|----------|-------------|---------|
| `{error}` | Error message | "connection timeout" |

### Available in `request_input` Hooks

| Variable | Description | Example |
|----------|-------------|---------|
| `{prompt_text}` | The prompt/question to show the user | "Deploy to production?" |

All standard variables (`{user_message}`, `{session_key}`, `{channel}`, `{chat_id}`) are also available in `request_input` hooks.

### Usage Examples

**Simple substitution:**
```bash
./bin/script --query '{user_message}'
```

**Multiple variables:**
```bash
./bin/script --user '{user_message}' --session '{session_key}' --channel '{channel}'
```

**JSON arguments:**
```bash
./bin/log_tool --data '{"tool":"{tool_name}","args":{tool_args}}'
```

### Shell Escaping

Template variables are automatically escaped for safe shell execution. Special characters in user input are properly quoted.

---

## Hook Lifecycle

### Execution Order

When processing a user message, hooks execute in this order:

```
1. User message received
   ↓
2. before_llm hooks execute
   ↓
3. LLM generates response (may call tools)
   │
   ├─→ For each tool call:
   │   ├─ Tool executes
   │   └─ on_tool_call hooks execute
   ↓
4. Final response generated
   ↓
5. after_response hooks execute
   
[If error occurs at any point]
   ↓
6. on_error hooks execute
```

### Hook Execution Details

**Timeout:** Each hook has a default timeout of 30 seconds. Hooks that exceed this timeout are killed.

**Failure Handling:** If a hook fails:
- The failure is logged (visible in agent logs)
- Other hooks continue executing
- The agent loop continues normally
- For `before_llm` hooks, no context is injected from the failed hook

**Concurrency:** Hooks execute sequentially, not in parallel. Each hook completes before the next begins.

**Environment:**
- Hooks execute in the workspace directory
- Python virtual environments (`.venv`, `venv`, `env`) are automatically detected and activated
- Hooks inherit environment variables from the PicoClaw process

---

## Use Cases and Examples

### Use Case 1: Long-Term Memory System

**Goal:** Automatically recall and store conversation context in an external memory system.

**Configuration:**

```json
{
  "loop_hooks": {
    "before_llm": [
      {
        "name": "memory_recall",
        "command": "./bin/memory_recall --query '{user_message}' --session '{session_key}' --limit 5",
        "enabled": true,
        "inject_as": "context"
      }
    ],
    "after_response": [
      {
        "name": "memory_store_user",
        "command": "./bin/memory_store --role user --content '{user_message}' --session '{session_key}'",
        "enabled": true,
        "inject_as": ""
      },
      {
        "name": "memory_store_assistant",
        "command": "./bin/memory_store --role assistant --content '{assistant_message}' --session '{session_key}'",
        "enabled": true,
        "inject_as": ""
      }
    ]
  }
}
```

**Scripts:**

`bin/memory_recall`:
```python
#!/usr/bin/env python3
import argparse
import json
from memory_system import search_memories

parser = argparse.ArgumentParser()
parser.add_argument('--query', required=True)
parser.add_argument('--session', required=True)
parser.add_argument('--limit', type=int, default=5)
args = parser.parse_args()

# Search memory system
memories = search_memories(args.query, args.session, limit=args.limit)

# Format for LLM context
if memories:
    print("## Relevant Context from Memory\n")
    for mem in memories:
        print(f"- {mem['content']}")
```

`bin/memory_store`:
```python
#!/usr/bin/env python3
import argparse
from memory_system import store_memory

parser = argparse.ArgumentParser()
parser.add_argument('--role', required=True)
parser.add_argument('--content', required=True)
parser.add_argument('--session', required=True)
args = parser.parse_args()

store_memory(
    role=args.role,
    content=args.content,
    session_id=args.session
)
```

### Use Case 2: Tool Usage Analytics

**Goal:** Track which tools are used most frequently and measure their performance.

**Configuration:**

```json
{
  "loop_hooks": {
    "on_tool_call": [
      {
        "name": "analytics",
        "command": "./bin/log_tool_usage --name '{tool_name}' --session '{session_key}' --channel '{channel}'",
        "enabled": true,
        "inject_as": ""
      }
    ]
  }
}
```

**Script:**

`bin/log_tool_usage`:
```python
#!/usr/bin/env python3
import argparse
import json
import time
from pathlib import Path

parser = argparse.ArgumentParser()
parser.add_argument('--name', required=True)
parser.add_argument('--session', required=True)
parser.add_argument('--channel', required=True)
args = parser.parse_args()

# Log to analytics file
log_file = Path.home() / '.picoclaw' / 'analytics' / 'tool_usage.jsonl'
log_file.parent.mkdir(parents=True, exist_ok=True)

entry = {
    'timestamp': time.time(),
    'tool': args.name,
    'session': args.session,
    'channel': args.channel
}

with open(log_file, 'a') as f:
    f.write(json.dumps(entry) + '\n')
```

### Use Case 3: Context-Aware Responses

**Goal:** Load user preferences and conversation style from a profile system.

**Configuration:**

```json
{
  "loop_hooks": {
    "before_llm": [
      {
        "name": "load_user_profile",
        "command": "./bin/load_profile --chat '{chat_id}' --channel '{channel}'",
        "enabled": true,
        "inject_as": "context"
      }
    ]
  }
}
```

**Script:**

`bin/load_profile`:
```python
#!/usr/bin/env python3
import argparse
from profile_system import get_profile

parser = argparse.ArgumentParser()
parser.add_argument('--chat', required=True)
parser.add_argument('--channel', required=True)
args = parser.parse_args()

profile = get_profile(args.channel, args.chat)

if profile:
    print("## User Profile\n")
    print(f"- Response style: {profile['style']}")
    print(f"- Expertise level: {profile['expertise']}")
    print(f"- Preferred language: {profile['language']}")
    if profile['notes']:
        print(f"- Notes: {profile['notes']}")
```

### Use Case 4: Error Notifications

**Goal:** Send alerts when the agent encounters errors.

**Configuration:**

```json
{
  "loop_hooks": {
    "on_error": [
      {
        "name": "alert_errors",
        "command": "./bin/send_alert --message '{error}' --session '{session_key}' --severity high",
        "enabled": true,
        "inject_as": ""
      }
    ]
  }
}
```

**Script:**

`bin/send_alert`:
```bash
#!/bin/bash
MESSAGE="$1"
SESSION="$2"
SEVERITY="$3"

# Send to monitoring system (e.g., Slack webhook)
curl -X POST https://hooks.slack.com/your-webhook-url \
  -H 'Content-Type: application/json' \
  -d "{
    \"text\": \"🚨 PicoClaw Error\",
    \"attachments\": [{
      \"color\": \"danger\",
      \"fields\": [
        {\"title\": \"Error\", \"value\": \"$MESSAGE\", \"short\": false},
        {\"title\": \"Session\", \"value\": \"$SESSION\", \"short\": true},
        {\"title\": \"Severity\", \"value\": \"$SEVERITY\", \"short\": true}
      ]
    }]
  }"
```

---

## Best Practices

### 1. Keep Hooks Fast

Hooks execute synchronously and block the agent loop. Keep execution time under 1-2 seconds when possible.

**Good:**
```python
# Quick database lookup
SELECT * FROM memories WHERE user_id = ? LIMIT 5
```

**Bad:**
```python
# Heavy computation
for item in large_dataset:
    process_item(item)  # Takes 30 seconds
```

### 2. Handle Errors Gracefully

Hooks should never crash. Always handle exceptions and provide fallback behavior.

```python
#!/usr/bin/env python3
try:
    # Your hook logic
    result = do_something()
    print(result)
except Exception as e:
    # Log error but don't crash
    print(f"Warning: Hook failed: {e}", file=sys.stderr)
    # Optionally provide default output
    print("## Using default context")
```

### 3. Use Context Injection Sparingly

Only `before_llm` hooks should use `inject_as: "context"`. Too much injected context can:
- Increase token costs
- Slow down LLM responses
- Dilute important information

**Guideline:** Keep injected context under 500 tokens (~350 words).

### 4. Log Hook Activity

Include logging in your hook scripts for debugging:

```python
import logging
logging.basicConfig(
    filename='/var/log/picoclaw/hooks.log',
    level=logging.INFO
)

logging.info(f"memory_recall: query={query}, results={len(results)}")
```

### 5. Make Scripts Executable

Always set execute permissions on hook scripts:

```bash
chmod +x workspace/bin/*
```

### 6. Test Hooks Independently

Test hook scripts outside of PicoClaw first:

```bash
# Test memory recall script
./bin/memory_recall "how do I use hooks?"

# Test with substituted variables
./bin/memory_store --role user --content "test message" --session "test-123"
```

### 7. Use Python Virtual Environments

If your hooks use Python:

```bash
cd ~/.picoclaw/workspace
python3 -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
```

PicoClaw automatically detects and activates `.venv`, `venv`, or `env` directories.

### 8. Disable Hooks During Development

While developing or debugging the agent, disable hooks to isolate issues:

```json
{
  "name": "memory_recall",
  "enabled": false  // Temporarily disabled
}
```

---

## Troubleshooting

### Hook Not Executing

**Check:**
1. Is `enabled: true`?
2. Is the script executable? (`chmod +x script`)
3. Is the path correct? (relative to workspace directory)
4. Check PicoClaw logs for hook errors

**Debug:**
```bash
# Run hook manually
cd ~/.picoclaw/workspace
./bin/your_hook --arg value
```

### Hook Timeout

If hooks timeout (default: 30 seconds):

1. Optimize script performance
2. Move slow operations to background jobs
3. Consider using async tools instead of hooks

### Context Not Injecting

**Check:**
1. Hook must be in `before_llm`
2. Must set `inject_as: "context"`
3. Hook must output text to stdout
4. Check for script errors in logs

**Debug:**
```bash
# Test hook output
./bin/memory_recall "test query"
# Should print markdown text
```

### Template Variable Not Substituting

**Check:**
1. Variable name is correct (e.g., `{user_message}`, not `{message}`)
2. Variable is available for that hook type
3. Single quotes around variable in command: `'{user_message}'`

### Python Environment Not Activated

**Check:**
1. Virtual environment exists: `.venv/`, `venv/`, or `env/`
2. Activation script exists: `.venv/bin/activate` (Unix) or `.venv/Scripts/activate.bat` (Windows)

**Create if missing:**
```bash
cd ~/.picoclaw/workspace
python3 -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
```

### Hook Errors Not Appearing

Hook errors are logged with log level WARN. Check your log configuration:

```json
{
  "log": {
    "level": "debug",  // Use debug to see all hook activity
    "format": "json"
  }
}
```

---

## Advanced Topics

### Custom Timeout Configuration

Currently, all hooks use a 30-second timeout. This may become configurable in future versions:

```json
{
  "name": "slow_hook",
  "command": "./bin/slow_process",
  "enabled": true,
  "timeout": 60  // Future: custom timeout in seconds
}
```

### Hook Retry Logic

Currently, hooks do not retry on failure. You can implement retry logic in your script:

```python
import time

MAX_RETRIES = 3
for attempt in range(MAX_RETRIES):
    try:
        result = api_call()
        print(result)
        break
    except Exception as e:
        if attempt < MAX_RETRIES - 1:
            time.sleep(2 ** attempt)  # Exponential backoff
        else:
            print(f"Failed after {MAX_RETRIES} attempts", file=sys.stderr)
```

### Async Hook Execution

Currently, hooks execute synchronously. For truly async operations, consider:

1. Using async tools (via `tools/` directory)
2. Making hooks fire-and-forget (non-blocking):

```bash
#!/bin/bash
# Run actual work in background
./bin/slow_operation &
# Return immediately
echo "Initiated async operation"
```

---

## Related Documentation

- [Workspace Tools Guide](../workspace/TOOLS.md) - Creating custom tools
- [Memory System Guide](../workspace/memory/README.md) - Memory system integration
- [Configuration Reference](tools_configuration.md) - Full config documentation

---

## Version

**Guide Version:** 1.0.0  
**Last Updated:** February 25, 2026  
**Compatible with:** PicoClaw v0.1.0+
