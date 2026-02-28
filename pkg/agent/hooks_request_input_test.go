package agent

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sipeed/picoclaw/pkg/bus"
	"github.com/sipeed/picoclaw/pkg/config"
)

func TestExecuteRequestInputHook_Success(t *testing.T) {
	// Create temporary workspace with test script
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "prompt.sh")

	scriptContent := `#!/bin/bash
echo "🤔 $1"
echo "Please enter your response:"
`
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	// Create message bus
	msgBus := bus.NewMessageBus()
	defer msgBus.Close()

	// Create hook configuration
	hook := config.LoopHook{
		Name:         "test_input",
		Command:      scriptPath + " '{prompt_text}'",
		Enabled:      true,
		Timeout:      5,
		ReturnAs:     "user_input",
		DefaultValue: "default_value",
	}

	// Create hook executor
	executor := NewHookExecutor(tmpDir)

	// Template variables
	vars := map[string]string{
		"prompt_text": "What is your favorite color?",
	}

	// Start goroutine to simulate user response
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		// Wait for request
		req, ok := msgBus.ConsumeInputRequest(ctx)
		if !ok {
			t.Logf("Failed to consume input request")
			return
		}

		// Send response
		msgBus.PublishInputResponse(bus.InputResponse{
			RequestID: req.RequestID,
			Input:     "Blue",
			TimedOut:  false,
		})
	}()

	// Execute hook
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := executor.ExecuteRequestInputHook(ctx, hook, vars, msgBus, "telegram", "user123")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result != "Blue" {
		t.Errorf("Expected result 'Blue', got: %s", result)
	}
}

func TestExecuteRequestInputHook_Timeout(t *testing.T) {
	// Create temporary workspace with test script
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "prompt.sh")

	scriptContent := `#!/bin/bash
echo "Test prompt"
`
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	// Create message bus
	msgBus := bus.NewMessageBus()
	defer msgBus.Close()

	// Create hook with short timeout
	hook := config.LoopHook{
		Name:         "test_timeout",
		Command:      scriptPath,
		Enabled:      true,
		Timeout:      1, // 1 second timeout
		ReturnAs:     "user_input",
		DefaultValue: "timeout_default",
	}

	// Create hook executor
	executor := NewHookExecutor(tmpDir)

	// Execute hook (no response will be sent)
	ctx := context.Background()

	result, err := executor.ExecuteRequestInputHook(ctx, hook, map[string]string{}, msgBus, "telegram", "user123")

	// Should return default value on timeout (no error)
	if err != nil {
		t.Errorf("Expected no error on timeout, got: %v", err)
	}

	if result != "timeout_default" {
		t.Errorf("Expected default value 'timeout_default', got: %s", result)
	}
}

func TestExecuteRequestInputHook_NilBus(t *testing.T) {
	tmpDir := t.TempDir()

	hook := config.LoopHook{
		Name:         "test_nil_bus",
		Command:      "echo test",
		Enabled:      true,
		Timeout:      5,
		ReturnAs:     "user_input",
		DefaultValue: "default",
	}

	executor := NewHookExecutor(tmpDir)

	ctx := context.Background()

	result, err := executor.ExecuteRequestInputHook(ctx, hook, map[string]string{}, nil, "telegram", "user123")

	// Should return default value and error
	if err == nil {
		t.Error("Expected error with nil bus")
	}

	if result != "default" {
		t.Errorf("Expected default value 'default', got: %s", result)
	}
}

func TestExecuteRequestInputHook_EmptyPrompt(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "empty.sh")

	// Script that produces empty output
	scriptContent := `#!/bin/bash
# Empty output
`
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	msgBus := bus.NewMessageBus()
	defer msgBus.Close()

	hook := config.LoopHook{
		Name:         "test_empty",
		Command:      scriptPath,
		Enabled:      true,
		Timeout:      5,
		ReturnAs:     "user_input",
		DefaultValue: "default",
	}

	executor := NewHookExecutor(tmpDir)

	ctx := context.Background()

	result, err := executor.ExecuteRequestInputHook(ctx, hook, map[string]string{}, msgBus, "telegram", "user123")

	// Should return default value when prompt is empty
	if err == nil {
		t.Error("Expected error with empty prompt")
	}

	if result != "default" {
		t.Errorf("Expected default value 'default', got: %s", result)
	}
}

func TestExecuteRequestInputHook_ContextCancellation(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "prompt.sh")

	scriptContent := `#!/bin/bash
echo "Test prompt"
`
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	msgBus := bus.NewMessageBus()
	defer msgBus.Close()

	hook := config.LoopHook{
		Name:         "test_cancel",
		Command:      scriptPath,
		Enabled:      true,
		Timeout:      10,
		ReturnAs:     "user_input",
		DefaultValue: "cancelled_default",
	}

	executor := NewHookExecutor(tmpDir)

	// Create context that we'll cancel
	ctx, cancel := context.WithCancel(context.Background())

	// Start execution in goroutine
	done := make(chan struct{})
	var result string
	var err error

	go func() {
		result, err = executor.ExecuteRequestInputHook(ctx, hook, map[string]string{}, msgBus, "telegram", "user123")
		close(done)
	}()

	// Cancel immediately
	time.Sleep(100 * time.Millisecond)
	cancel()

	// Wait for completion
	select {
	case <-done:
		// Should return default value and either context error or command error
		// (context cancellation can kill the command, resulting in command error)
		if err == nil {
			t.Error("Expected error on cancellation")
		}

		if result != "cancelled_default" {
			t.Errorf("Expected default value 'cancelled_default', got: %s", result)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for cancellation")
	}
}

func TestExecuteRequestInputHook_DefaultTimeout(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "prompt.sh")

	scriptContent := `#!/bin/bash
echo "Test prompt"
`
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	msgBus := bus.NewMessageBus()
	defer msgBus.Close()

	// Hook with no timeout specified (should default to 60 seconds)
	hook := config.LoopHook{
		Name:         "test_default_timeout",
		Command:      scriptPath,
		Enabled:      true,
		Timeout:      0, // 0 means use default
		ReturnAs:     "user_input",
		DefaultValue: "default",
	}

	executor := NewHookExecutor(tmpDir)

	// Start goroutine to respond quickly
	go func() {
		time.Sleep(100 * time.Millisecond)
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		req, ok := msgBus.ConsumeInputRequest(ctx)
		if !ok {
			return
		}

		msgBus.PublishInputResponse(bus.InputResponse{
			RequestID: req.RequestID,
			Input:     "fast_response",
			TimedOut:  false,
		})
	}()

	ctx := context.Background()

	result, err := executor.ExecuteRequestInputHook(ctx, hook, map[string]string{}, msgBus, "telegram", "user123")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result != "fast_response" {
		t.Errorf("Expected result 'fast_response', got: %s", result)
	}
}

func TestExecuteRequestInputHook_VariableSubstitution(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "prompt.sh")

	scriptContent := `#!/bin/bash
echo "$1 - $2"
`
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	msgBus := bus.NewMessageBus()
	defer msgBus.Close()

	hook := config.LoopHook{
		Name:         "test_vars",
		Command:      scriptPath + " '{var1}' '{var2}'",
		Enabled:      true,
		Timeout:      2,
		ReturnAs:     "user_input",
		DefaultValue: "default",
	}

	executor := NewHookExecutor(tmpDir)

	vars := map[string]string{
		"var1": "Hello",
		"var2": "World",
	}

	// Respond to the request
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		req, ok := msgBus.ConsumeInputRequest(ctx)
		if !ok {
			return
		}

		// Verify prompt has substituted variables
		expectedPrompt := "Hello - World"
		if req.Prompt != expectedPrompt {
			t.Errorf("Expected prompt '%s', got '%s'", expectedPrompt, req.Prompt)
		}

		msgBus.PublishInputResponse(bus.InputResponse{
			RequestID: req.RequestID,
			Input:     "substituted",
			TimedOut:  false,
		})
	}()

	ctx := context.Background()

	result, err := executor.ExecuteRequestInputHook(ctx, hook, vars, msgBus, "telegram", "user123")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result != "substituted" {
		t.Errorf("Expected result 'substituted', got: %s", result)
	}
}
