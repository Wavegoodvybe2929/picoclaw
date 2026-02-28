package tools

import (
	"context"
	"errors"
	"testing"
)

func TestRequestInputTool_Name(t *testing.T) {
	tool := NewRequestInputTool()
	if tool.Name() != "request_input" {
		t.Errorf("Expected name 'request_input', got: %s", tool.Name())
	}
}

func TestRequestInputTool_Description(t *testing.T) {
	tool := NewRequestInputTool()
	desc := tool.Description()
	if desc == "" {
		t.Error("Expected non-empty description")
	}
	// Check that description mentions key concepts
	if !contains(desc, "input") || !contains(desc, "user") {
		t.Error("Description should mention 'input' and 'user'")
	}
}

func TestRequestInputTool_Parameters(t *testing.T) {
	tool := NewRequestInputTool()
	params := tool.Parameters()

	// Verify structure
	if params["type"] != "object" {
		t.Error("Expected type 'object'")
	}

	props, ok := params["properties"].(map[string]any)
	if !ok {
		t.Fatal("Expected properties map")
	}

	// Verify prompt parameter exists
	prompt, ok := props["prompt"].(map[string]any)
	if !ok {
		t.Fatal("Expected 'prompt' parameter")
	}

	if prompt["type"] != "string" {
		t.Error("Expected prompt type to be 'string'")
	}

	// Verify required fields
	required, ok := params["required"].([]string)
	if !ok {
		t.Fatal("Expected required array")
	}

	if len(required) != 1 || required[0] != "prompt" {
		t.Error("Expected 'prompt' to be required")
	}
}

func TestRequestInputTool_Execute_MissingPrompt(t *testing.T) {
	tool := NewRequestInputTool()

	// Set a dummy callback
	tool.SetRequestCallback(func(ctx context.Context, prompt string) (string, error) {
		return "should not be called", nil
	})

	// Execute without prompt
	result := tool.Execute(context.Background(), map[string]any{})

	if !result.IsError {
		t.Error("Expected error result for missing prompt")
	}

	if result.ForLLM == "" {
		t.Error("Expected error message in ForLLM")
	}
}

func TestRequestInputTool_Execute_EmptyPrompt(t *testing.T) {
	tool := NewRequestInputTool()

	tool.SetRequestCallback(func(ctx context.Context, prompt string) (string, error) {
		return "should not be called", nil
	})

	// Execute with empty prompt
	result := tool.Execute(context.Background(), map[string]any{
		"prompt": "",
	})

	if !result.IsError {
		t.Error("Expected error result for empty prompt")
	}
}

func TestRequestInputTool_Execute_NoCallback(t *testing.T) {
	tool := NewRequestInputTool()

	// Don't set callback

	result := tool.Execute(context.Background(), map[string]any{
		"prompt": "Test prompt",
	})

	if !result.IsError {
		t.Error("Expected error result when callback not configured")
	}

	if !contains(result.ForLLM, "not configured") {
		t.Error("Expected error message to mention 'not configured'")
	}
}

func TestRequestInputTool_Execute_Success(t *testing.T) {
	tool := NewRequestInputTool()

	expectedPrompt := "What is your favorite color?"
	expectedResponse := "Blue"

	// Set callback that returns expected response
	tool.SetRequestCallback(func(ctx context.Context, prompt string) (string, error) {
		if prompt != expectedPrompt {
			t.Errorf("Expected prompt '%s', got '%s'", expectedPrompt, prompt)
		}
		return expectedResponse, nil
	})

	result := tool.Execute(context.Background(), map[string]any{
		"prompt": expectedPrompt,
	})

	if result.IsError {
		t.Errorf("Expected success, got error: %s", result.ForLLM)
	}

	if !contains(result.ForLLM, expectedResponse) {
		t.Errorf("Expected ForLLM to contain '%s', got: %s", expectedResponse, result.ForLLM)
	}

	if !result.Silent {
		t.Error("Expected result to be silent (user already saw prompt and responded)")
	}

	if result.ForUser != "" {
		t.Error("Expected ForUser to be empty for silent result")
	}
}

func TestRequestInputTool_Execute_CallbackError(t *testing.T) {
	tool := NewRequestInputTool()

	expectedError := errors.New("timeout waiting for user")

	tool.SetRequestCallback(func(ctx context.Context, prompt string) (string, error) {
		return "", expectedError
	})

	result := tool.Execute(context.Background(), map[string]any{
		"prompt": "Test prompt",
	})

	if !result.IsError {
		t.Error("Expected error result when callback returns error")
	}

	// Check that error message contains relevant keywords
	errorMsg := result.ForLLM
	if !contains(errorMsg, "Failed") && !contains(errorMsg, "failed") {
		t.Errorf("Expected error message to contain 'failed', got: %s", errorMsg)
	}

	if result.Err != expectedError {
		t.Error("Expected Err field to contain the callback error")
	}
}

func TestRequestInputTool_SetContext(t *testing.T) {
	tool := NewRequestInputTool()

	tool.SetContext("telegram", "user123")

	// Context is set internally - just verify no panic
	if tool.defaultChannel != "telegram" {
		t.Errorf("Expected channel 'telegram', got: %s", tool.defaultChannel)
	}

	if tool.defaultChatID != "user123" {
		t.Errorf("Expected chatID 'user123', got: %s", tool.defaultChatID)
	}
}

func TestRequestInputTool_Execute_WithContext(t *testing.T) {
	tool := NewRequestInputTool()
	tool.SetContext("slack", "channel456")

	called := false
	tool.SetRequestCallback(func(ctx context.Context, prompt string) (string, error) {
		called = true

		// Verify context is passed through
		select {
		case <-ctx.Done():
			t.Error("Context should not be cancelled")
		default:
		}

		return "response", nil
	})

	ctx := context.Background()
	result := tool.Execute(ctx, map[string]any{
		"prompt": "Test",
	})

	if !called {
		t.Error("Expected callback to be called")
	}

	if result.IsError {
		t.Errorf("Expected success, got error: %s", result.ForLLM)
	}
}

func TestRequestInputTool_Execute_InvalidPromptType(t *testing.T) {
	tool := NewRequestInputTool()

	tool.SetRequestCallback(func(ctx context.Context, prompt string) (string, error) {
		return "should not be called", nil
	})

	// Execute with non-string prompt
	result := tool.Execute(context.Background(), map[string]any{
		"prompt": 123, // Invalid type
	})

	if !result.IsError {
		t.Error("Expected error result for invalid prompt type")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())
}
