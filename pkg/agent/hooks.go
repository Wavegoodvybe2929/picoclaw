package agent

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sipeed/picoclaw/pkg/bus"
	"github.com/sipeed/picoclaw/pkg/config"
	"github.com/sipeed/picoclaw/pkg/logger"
)

// HookExecutor manages the execution of lifecycle hooks in the agent loop.
// It wraps the exec pattern with template variable substitution and context injection.
type HookExecutor struct {
	workspaceDir string
	pythonVenv   string
	timeout      time.Duration
}

// NewHookExecutor creates a new HookExecutor for the given workspace.
// It automatically detects Python virtual environment if present.
func NewHookExecutor(workspaceDir string) *HookExecutor {
	executor := &HookExecutor{
		workspaceDir: workspaceDir,
		timeout:      30 * time.Second, // Default timeout for hooks
	}

	// Detect Python virtual environment
	executor.pythonVenv = executor.detectPythonVenv()

	return executor
}

// SetTimeout sets the timeout for hook execution.
func (h *HookExecutor) SetTimeout(timeout time.Duration) {
	h.timeout = timeout
}

// detectPythonVenv checks for a Python virtual environment in the workspace.
// Returns the path to the venv activate script if found, empty string otherwise.
func (h *HookExecutor) detectPythonVenv() string {
	if h.workspaceDir == "" {
		return ""
	}

	// Common venv locations
	venvPaths := []string{
		filepath.Join(h.workspaceDir, ".venv"),
		filepath.Join(h.workspaceDir, "venv"),
		filepath.Join(h.workspaceDir, "env"),
	}

	for _, venvPath := range venvPaths {
		// Check if venv directory exists
		if stat, err := os.Stat(venvPath); err == nil && stat.IsDir() {
			// Check for activation script
			var activateScript string
			if runtime.GOOS == "windows" {
				activateScript = filepath.Join(venvPath, "Scripts", "activate.bat")
			} else {
				activateScript = filepath.Join(venvPath, "bin", "activate")
			}

			if _, err := os.Stat(activateScript); err == nil {
				return venvPath
			}
		}
	}

	return ""
}

// ExecuteHooks executes a list of hooks with the given template variables.
// Returns a map of injection results where the key is "context" for inject_as="context" hooks.
func (h *HookExecutor) ExecuteHooks(
	ctx context.Context,
	hooks []config.LoopHook,
	vars map[string]string,
) (map[string]string, error) {
	results := make(map[string]string)

	for _, hook := range hooks {
		// Skip disabled hooks
		if !hook.Enabled {
			continue
		}

		// Execute the hook
		output, err := h.executeHook(ctx, hook, vars)
		if err != nil {
			// Log error but don't stop execution of other hooks
			logger.WarnCF("agent", "Hook execution failed",
				map[string]any{
					"hook_name": hook.Name,
					"error":     err.Error(),
					"command":   hook.Command,
				})
			continue
		}

		// Store result if injection is requested
		if hook.InjectAs != "" {
			// Append to existing content if key already exists
			if existing, ok := results[hook.InjectAs]; ok {
				results[hook.InjectAs] = existing + "\n\n" + output
			} else {
				results[hook.InjectAs] = output
			}
		}
	}

	return results, nil
}

// executeHook executes a single hook with template variable substitution.
func (h *HookExecutor) executeHook(
	ctx context.Context,
	hook config.LoopHook,
	vars map[string]string,
) (string, error) {
	// Substitute template variables in command
	command := h.substituteVariables(hook.Command, vars)

	// Determine working directory
	workingDir := h.workspaceDir
	if workingDir == "" {
		var err error
		workingDir, err = os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get working directory: %w", err)
		}
	}

	// Create context with timeout
	execCtx, cancel := context.WithTimeout(ctx, h.timeout)
	defer cancel()

	// Build command
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(execCtx, "powershell", "-NoProfile", "-NonInteractive", "-Command", command)
	} else {
		// Check if this is a Python script and we have a venv
		if h.pythonVenv != "" && h.isPythonCommand(command) {
			// Wrap command to activate venv first
			command = h.wrapWithVenv(command)
		}
		cmd = exec.CommandContext(execCtx, "sh", "-c", command)
	}

	cmd.Dir = workingDir

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute command
	err := cmd.Run()
	if err != nil {
		// Include stderr in error for debugging
		stderrStr := stderr.String()
		if stderrStr != "" {
			return "", fmt.Errorf("command failed: %w\nStderr: %s", err, stderrStr)
		}
		return "", fmt.Errorf("command failed: %w", err)
	}

	// Return stdout as result
	output := stdout.String()

	// Trim trailing whitespace but preserve internal formatting
	output = strings.TrimSpace(output)

	return output, nil
}

// substituteVariables replaces template variables in the command string.
// Template variables are in the format {variable_name}.
func (h *HookExecutor) substituteVariables(command string, vars map[string]string) string {
	result := command

	for key, value := range vars {
		// Escape the value for shell safety
		escapedValue := h.shellEscape(value)

		// Replace {key} with the escaped value
		placeholder := "{" + key + "}"
		result = strings.ReplaceAll(result, placeholder, escapedValue)
	}

	return result
}

// shellEscape escapes a string for safe use in shell commands.
// This prevents command injection by properly quoting the string.
func (h *HookExecutor) shellEscape(s string) string {
	// For empty strings, return empty quotes
	if s == "" {
		return "''"
	}

	// Check if the string needs escaping
	needsEscape := false
	for _, ch := range s {
		if !isSafeChar(ch) {
			needsEscape = true
			break
		}
	}

	// If no special chars, return as-is
	if !needsEscape {
		return s
	}

	// Use single quotes and escape any single quotes in the string
	// Replace ' with '\''
	escaped := strings.ReplaceAll(s, "'", "'\\''")
	return "'" + escaped + "'"
}

// isSafeChar checks if a character is safe for shell without escaping.
func isSafeChar(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') ||
		(ch >= 'A' && ch <= 'Z') ||
		(ch >= '0' && ch <= '9') ||
		ch == '-' || ch == '_' || ch == '.' || ch == '/' ||
		ch == ':' || ch == '@'
}

// isPythonCommand checks if a command is invoking a Python script.
func (h *HookExecutor) isPythonCommand(command string) bool {
	trimmed := strings.TrimSpace(command)

	// Check for direct Python invocation
	if strings.HasPrefix(trimmed, "python") ||
		strings.HasPrefix(trimmed, "python3") ||
		strings.HasPrefix(trimmed, "python2") {
		return true
	}

	// Check for shebang-based script in workspace bin/
	if strings.HasPrefix(trimmed, "./bin/") || strings.HasPrefix(trimmed, "bin/") {
		return true // Assume workspace scripts might be Python
	}

	return false
}

// wrapWithVenv wraps a command to execute within the Python virtual environment.
func (h *HookExecutor) wrapWithVenv(command string) string {
	if runtime.GOOS == "windows" {
		activateScript := filepath.Join(h.pythonVenv, "Scripts", "activate.bat")
		return fmt.Sprintf("call %s && %s", activateScript, command)
	}

	// Unix-like systems: source the activate script
	activateScript := filepath.Join(h.pythonVenv, "bin", "activate")
	return fmt.Sprintf(". %s && %s", activateScript, command)
}

// ExecuteHook executes a single hook and returns its output.
// This is a convenience method for executing individual hooks.
func (h *HookExecutor) ExecuteHook(
	ctx context.Context,
	hook config.LoopHook,
	vars map[string]string,
) (string, error) {
	return h.executeHook(ctx, hook, vars)
}

// ExecuteRequestInputHook executes a hook that requests input from the user.
// This method blocks until the user responds or the timeout expires.
// It returns the user's input, or the default value if timeout expires or an error occurs.
//
// The hook command is executed to generate the prompt text, which is then sent to the user
// via the message bus. The method waits for the user's response using a subscription channel.
//
// Parameters:
//   - ctx: Context for cancellation
//   - hook: The request_input hook configuration
//   - vars: Template variables to substitute in the hook command
//   - msgBus: Message bus for sending prompts and receiving responses
//   - channel: Channel name where the request originated (e.g., "slack", "telegram")
//   - chatID: Chat ID where the request originated
//
// Returns:
//   - string: The user's response, or hook.DefaultValue if timeout/error
//   - error: Error if hook command fails or context is cancelled
//
// Concurrency: Safe for concurrent use. Each request gets a unique ID and response channel.
func (h *HookExecutor) ExecuteRequestInputHook(
	ctx context.Context,
	hook config.LoopHook,
	vars map[string]string,
	msgBus *bus.MessageBus,
	channel string,
	chatID string,
) (string, error) {
	// Validate inputs
	if msgBus == nil {
		logger.WarnCF("agent", "Message bus unavailable for input request",
			map[string]any{"hook": hook.Name})
		return hook.DefaultValue, fmt.Errorf("message bus is nil")
	}

	// Execute hook command to generate prompt
	prompt, err := h.executeHook(ctx, hook, vars)
	if err != nil {
		logger.WarnCF("agent", "Failed to generate input request prompt",
			map[string]any{
				"hook":  hook.Name,
				"error": err.Error(),
			})
		return hook.DefaultValue, fmt.Errorf("failed to generate prompt: %w", err)
	}

	// Trim whitespace from prompt
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		logger.WarnCF("agent", "Input request prompt is empty",
			map[string]any{"hook": hook.Name})
		return hook.DefaultValue, fmt.Errorf("prompt is empty")
	}

	// Create unique request ID
	requestID := uuid.New().String()

	// Determine timeout (default to 60 seconds if not specified)
	timeout := hook.Timeout
	if timeout <= 0 {
		timeout = 60
	}

	// Subscribe to response channel before publishing request
	responseChan := msgBus.SubscribeInputResponse(requestID)

	// Create and publish input request
	request := bus.InputRequest{
		RequestID: requestID,
		Channel:   channel,
		ChatID:    chatID,
		Prompt:    prompt,
		Timeout:   timeout,
	}

	logger.InfoCF("agent", "Sending input request to user",
		map[string]any{
			"hook":       hook.Name,
			"request_id": requestID,
			"timeout":    timeout,
		})

	msgBus.PublishInputRequest(request)

	// Wait for response with timeout
	select {
	case response := <-responseChan:
		if response.TimedOut {
			logger.WarnCF("agent", "Input request timed out",
				map[string]any{
					"hook":       hook.Name,
					"request_id": requestID,
				})
			return hook.DefaultValue, nil
		}
		logger.InfoCF("agent", "Received user input",
			map[string]any{
				"hook":       hook.Name,
				"request_id": requestID,
			})
		return response.Input, nil

	case <-time.After(time.Duration(timeout) * time.Second):
		// Timeout expired - clean up and return default
		logger.WarnCF("agent", "Input request timeout expired",
			map[string]any{
				"hook":       hook.Name,
				"request_id": requestID,
				"timeout":    timeout,
			})
		return hook.DefaultValue, nil

	case <-ctx.Done():
		// Context cancelled
		logger.WarnCF("agent", "Input request cancelled",
			map[string]any{
				"hook":       hook.Name,
				"request_id": requestID,
			})
		return hook.DefaultValue, ctx.Err()
	}
}
