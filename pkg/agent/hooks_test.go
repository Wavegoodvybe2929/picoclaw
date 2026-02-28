package agent

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/sipeed/picoclaw/pkg/config"
)

func TestNewHookExecutor(t *testing.T) {
	tempDir := t.TempDir()

	executor := NewHookExecutor(tempDir)

	if executor == nil {
		t.Fatal("NewHookExecutor returned nil")
	}

	if executor.workspaceDir != tempDir {
		t.Errorf("Expected workspaceDir %q, got %q", tempDir, executor.workspaceDir)
	}

	if executor.timeout == 0 {
		t.Error("Expected non-zero timeout")
	}
}

func TestDetectPythonVenv(t *testing.T) {
	tempDir := t.TempDir()

	// Create a mock venv structure
	venvPath := filepath.Join(tempDir, ".venv")
	if err := os.MkdirAll(venvPath, 0o755); err != nil {
		t.Fatalf("Failed to create venv dir: %v", err)
	}

	var activateScript string
	if runtime.GOOS == "windows" {
		scriptsDir := filepath.Join(venvPath, "Scripts")
		if err := os.MkdirAll(scriptsDir, 0o755); err != nil {
			t.Fatalf("Failed to create Scripts dir: %v", err)
		}
		activateScript = filepath.Join(scriptsDir, "activate.bat")
	} else {
		binDir := filepath.Join(venvPath, "bin")
		if err := os.MkdirAll(binDir, 0o755); err != nil {
			t.Fatalf("Failed to create bin dir: %v", err)
		}
		activateScript = filepath.Join(binDir, "activate")
	}

	// Create activate script
	if err := os.WriteFile(activateScript, []byte("# mock activate script"), 0o755); err != nil {
		t.Fatalf("Failed to create activate script: %v", err)
	}

	executor := NewHookExecutor(tempDir)

	if executor.pythonVenv == "" {
		t.Error("Expected Python venv to be detected")
	}

	if executor.pythonVenv != venvPath {
		t.Errorf("Expected venv path %q, got %q", venvPath, executor.pythonVenv)
	}
}

func TestDetectPythonVenv_NotFound(t *testing.T) {
	tempDir := t.TempDir()

	executor := NewHookExecutor(tempDir)

	if executor.pythonVenv != "" {
		t.Errorf("Expected no venv, got %q", executor.pythonVenv)
	}
}

func TestSubstituteVariables(t *testing.T) {
	executor := &HookExecutor{}

	tests := []struct {
		name     string
		command  string
		vars     map[string]string
		expected string
	}{
		{
			name:     "simple substitution",
			command:  "echo {message}",
			vars:     map[string]string{"message": "hello"},
			expected: "echo hello",
		},
		{
			name:     "multiple substitutions",
			command:  "echo {greeting} {name}",
			vars:     map[string]string{"greeting": "hello", "name": "world"},
			expected: "echo hello world",
		},
		{
			name:     "no substitutions",
			command:  "echo hello",
			vars:     map[string]string{},
			expected: "echo hello",
		},
		{
			name:     "missing variable",
			command:  "echo {missing}",
			vars:     map[string]string{"other": "value"},
			expected: "echo {missing}",
		},
		{
			name:     "special characters",
			command:  "./script {query}",
			vars:     map[string]string{"query": "hello world"},
			expected: "./script 'hello world'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := executor.substituteVariables(tt.command, tt.vars)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestShellEscape(t *testing.T) {
	executor := &HookExecutor{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "safe string",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "string with spaces",
			input:    "hello world",
			expected: "'hello world'",
		},
		{
			name:     "string with single quote",
			input:    "it's",
			expected: "'it'\\''s'",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "''",
		},
		{
			name:     "string with special chars",
			input:    "hello; rm -rf /",
			expected: "'hello; rm -rf /'",
		},
		{
			name:     "safe path",
			input:    "/usr/bin/python",
			expected: "/usr/bin/python",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := executor.shellEscape(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestIsSafeChar(t *testing.T) {
	tests := []struct {
		char rune
		safe bool
	}{
		{'a', true},
		{'Z', true},
		{'0', true},
		{'9', true},
		{'-', true},
		{'_', true},
		{'.', true},
		{'/', true},
		{':', true},
		{'@', true},
		{' ', false},
		{'$', false},
		{'&', false},
		{'|', false},
		{';', false},
		{'\'', false},
		{'"', false},
	}

	for _, tt := range tests {
		t.Run(string(tt.char), func(t *testing.T) {
			result := isSafeChar(tt.char)
			if result != tt.safe {
				t.Errorf("Expected isSafeChar(%q) = %v, got %v", tt.char, tt.safe, result)
			}
		})
	}
}

func TestIsPythonCommand(t *testing.T) {
	executor := &HookExecutor{}

	tests := []struct {
		name     string
		command  string
		expected bool
	}{
		{
			name:     "python command",
			command:  "python script.py",
			expected: true,
		},
		{
			name:     "python3 command",
			command:  "python3 script.py",
			expected: true,
		},
		{
			name:     "workspace bin script",
			command:  "./bin/memory_recall",
			expected: true,
		},
		{
			name:     "bin without dot",
			command:  "bin/script",
			expected: true,
		},
		{
			name:     "not python",
			command:  "echo hello",
			expected: false,
		},
		{
			name:     "bash script",
			command:  "bash script.sh",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := executor.isPythonCommand(tt.command)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestExecuteHook_Success(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping on Windows")
	}

	tempDir := t.TempDir()
	executor := NewHookExecutor(tempDir)

	hook := config.LoopHook{
		Name:     "test_hook",
		Command:  "echo {message}",
		Enabled:  true,
		InjectAs: "context",
	}

	vars := map[string]string{
		"message": "hello world",
	}

	ctx := context.Background()
	output, err := executor.ExecuteHook(ctx, hook, vars)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if output != "hello world" {
		t.Errorf("Expected 'hello world', got %q", output)
	}
}

func TestExecuteHook_WithVariableSubstitution(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping on Windows")
	}

	tempDir := t.TempDir()
	executor := NewHookExecutor(tempDir)

	hook := config.LoopHook{
		Name:     "test_hook",
		Command:  "echo '{greeting} {name}'",
		Enabled:  true,
		InjectAs: "context",
	}

	vars := map[string]string{
		"greeting": "hello",
		"name":     "world",
	}

	ctx := context.Background()
	output, err := executor.ExecuteHook(ctx, hook, vars)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if output != "hello world" {
		t.Errorf("Expected 'hello world', got %q", output)
	}
}

func TestExecuteHook_Timeout(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping on Windows")
	}

	tempDir := t.TempDir()
	executor := NewHookExecutor(tempDir)
	executor.SetTimeout(100 * time.Millisecond)

	hook := config.LoopHook{
		Name:    "slow_hook",
		Command: "sleep 5",
		Enabled: true,
	}

	ctx := context.Background()
	_, err := executor.ExecuteHook(ctx, hook, map[string]string{})

	if err == nil {
		t.Error("Expected timeout error")
	}
}

func TestExecuteHooks_MultipleHooks(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping on Windows")
	}

	tempDir := t.TempDir()
	executor := NewHookExecutor(tempDir)

	hooks := []config.LoopHook{
		{
			Name:     "hook1",
			Command:  "echo first",
			Enabled:  true,
			InjectAs: "context",
		},
		{
			Name:     "hook2",
			Command:  "echo second",
			Enabled:  true,
			InjectAs: "context",
		},
	}

	ctx := context.Background()
	results, err := executor.ExecuteHooks(ctx, hooks, map[string]string{})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	contextResult, ok := results["context"]
	if !ok {
		t.Fatal("Expected context result")
	}

	// Both outputs should be combined
	if !contains(contextResult, "first") || !contains(contextResult, "second") {
		t.Errorf("Expected both outputs, got %q", contextResult)
	}
}

func TestExecuteHooks_DisabledHook(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping on Windows")
	}

	tempDir := t.TempDir()
	executor := NewHookExecutor(tempDir)

	hooks := []config.LoopHook{
		{
			Name:     "disabled_hook",
			Command:  "echo should not run",
			Enabled:  false,
			InjectAs: "context",
		},
	}

	ctx := context.Background()
	results, err := executor.ExecuteHooks(ctx, hooks, map[string]string{})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(results) > 0 {
		t.Errorf("Expected no results from disabled hook, got %v", results)
	}
}

func TestExecuteHooks_NoInjection(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping on Windows")
	}

	tempDir := t.TempDir()
	executor := NewHookExecutor(tempDir)

	hooks := []config.LoopHook{
		{
			Name:    "no_inject_hook",
			Command: "echo test",
			Enabled: true,
			// InjectAs is empty - no injection
		},
	}

	ctx := context.Background()
	results, err := executor.ExecuteHooks(ctx, hooks, map[string]string{})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(results) > 0 {
		t.Errorf("Expected no injection results, got %v", results)
	}
}

func TestWrapWithVenv(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping on Windows")
	}

	tempDir := t.TempDir()
	venvPath := filepath.Join(tempDir, ".venv")

	executor := &HookExecutor{
		workspaceDir: tempDir,
		pythonVenv:   venvPath,
	}

	command := "python script.py"
	wrapped := executor.wrapWithVenv(command)

	expectedActivate := filepath.Join(venvPath, "bin", "activate")
	expected := fmt.Sprintf(". %s && %s", expectedActivate, command)

	if wrapped != expected {
		t.Errorf("Expected %q, got %q", expected, wrapped)
	}
}

func TestSetTimeout(t *testing.T) {
	executor := NewHookExecutor("")

	newTimeout := 5 * time.Second
	executor.SetTimeout(newTimeout)

	if executor.timeout != newTimeout {
		t.Errorf("Expected timeout %v, got %v", newTimeout, executor.timeout)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					len(s) > len(substr)+1 &&
						containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 1; i < len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestOnToolCallHookVariables(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping shell script test on Windows")
	}

	tempDir := t.TempDir()

	// Create a test hook script that echoes variables
	scriptPath := filepath.Join(tempDir, "test_tool_hook.sh")
	scriptContent := `#!/bin/bash
echo "tool_name=$1"
echo "tool_args=$2"
echo "tool_result=$3"
echo "session=$4"
`
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0o755); err != nil {
		t.Fatalf("Failed to write test script: %v", err)
	}

	executor := NewHookExecutor(tempDir)

	hook := config.LoopHook{
		Name:     "test_tool_hook",
		Command:  fmt.Sprintf("%s '{tool_name}' '{tool_args}' '{tool_result}' '{session_key}'", scriptPath),
		Enabled:  true,
		InjectAs: "",
	}

	vars := map[string]string{
		"tool_name":   "code_search",
		"tool_args":   `{"query":"hooks","max":5}`,
		"tool_result": "Found 3 files",
		"session_key": "test-session-123",
	}

	ctx := context.Background()
	results, err := executor.ExecuteHooks(ctx, []config.LoopHook{hook}, vars)

	if err != nil {
		t.Fatalf("ExecuteHooks failed: %v", err)
	}

	// Hook has no inject_as, so results should be empty or not contain 'context'
	if results["context"] != "" {
		t.Errorf("Expected no context injection, got: %q", results["context"])
	}
}

func TestOnToolCallHookExecution(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping shell script test on Windows")
	}

	tempDir := t.TempDir()

	// Create a test hook that logs to a file
	logFile := filepath.Join(tempDir, "tool_calls.log")
	scriptPath := filepath.Join(tempDir, "log_tool.sh")
	scriptContent := fmt.Sprintf(`#!/bin/bash
echo "$(date +%%s)|$1|$2" >> %s
`, logFile)
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0o755); err != nil {
		t.Fatalf("Failed to write test script: %v", err)
	}

	executor := NewHookExecutor(tempDir)

	hook := config.LoopHook{
		Name:     "log_tool",
		Command:  fmt.Sprintf("%s '{tool_name}' '{tool_args}'", scriptPath),
		Enabled:  true,
		InjectAs: "",
	}

	// Simulate multiple tool calls
	toolCalls := []map[string]string{
		{
			"tool_name": "code_search",
			"tool_args": `{"query":"test"}`,
		},
		{
			"tool_name": "file_read",
			"tool_args": `{"path":"test.go"}`,
		},
	}

	ctx := context.Background()
	for _, vars := range toolCalls {
		_, err := executor.ExecuteHooks(ctx, []config.LoopHook{hook}, vars)
		if err != nil {
			t.Fatalf("ExecuteHooks failed: %v", err)
		}
	}

	// Verify log file was created and has entries
	logContent, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logStr := string(logContent)
	if !contains(logStr, "code_search") {
		t.Errorf("Expected log to contain 'code_search', got: %s", logStr)
	}
	if !contains(logStr, "file_read") {
		t.Errorf("Expected log to contain 'file_read', got: %s", logStr)
	}
}
