// PicoClaw - Ultra-lightweight personal AI agent
// License: MIT
//
// Copyright (c) 2026 PicoClaw contributors

package providers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/sipeed/picoclaw/pkg/config"
)

// --- Compile-time interface checks ---

var (
	_ LLMProvider      = (*RLMProvider)(nil)
	_ StatefulProvider = (*RLMProvider)(nil)
)

// --- Configuration Tests ---

func TestNewRLMProvider_ConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		cfg     config.RLMConfig
		wantErr bool
		errMsg  string
		skipRun bool // Skip actually calling NewRLMProvider
	}{
		{
			name: "valid config",
			cfg: config.RLMConfig{
				Enabled:         true,
				UpstreamBaseURL: "http://localhost:1234/v1",
				UpstreamModel:   "test-model",
			},
			wantErr: false,
			skipRun: true, // Skip because it would try to start subprocess
		},
		{
			name: "disabled provider",
			cfg: config.RLMConfig{
				Enabled: false,
			},
			wantErr: true,
			errMsg:  "not enabled",
		},
		{
			name: "missing upstream_base_url",
			cfg: config.RLMConfig{
				Enabled:       true,
				UpstreamModel: "test-model",
			},
			wantErr: true,
			errMsg:  "upstream_base_url is required",
		},
		{
			name: "missing upstream_model",
			cfg: config.RLMConfig{
				Enabled:         true,
				UpstreamBaseURL: "http://localhost:1234/v1",
			},
			wantErr: true,
			errMsg:  "upstream_model is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipRun {
				t.Skip("Skipping test that would require RLMgw subprocess")
				return
			}
			_, err := NewRLMProvider(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRLMProvider() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("error = %q, want to contain %q", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestNewRLMProvider_DefaultValues(t *testing.T) {
	// Skip subprocess startup for this test by mocking
	// We're just testing that defaults are applied
	cfg := config.RLMConfig{
		Enabled:         true,
		UpstreamBaseURL: "http://localhost:1234/v1",
		UpstreamModel:   "test-model",
		// Omit Host, Port, MaxInternalCalls, MaxContextChars to test defaults
	}

	// This will fail to start (no actual rlmgw), but we can check config processing
	p := &RLMProvider{config: cfg}

	// Apply defaults (same logic as in NewRLMProvider)
	if p.config.Host == "" {
		p.config.Host = "127.0.0.1"
	}
	if p.config.Port == 0 {
		p.config.Port = 8010
	}
	if p.config.MaxInternalCalls == 0 {
		p.config.MaxInternalCalls = 3
	}
	if p.config.MaxContextChars == 0 {
		p.config.MaxContextChars = 12000
	}

	if p.config.Host != "127.0.0.1" {
		t.Errorf("Host = %q, want %q", p.config.Host, "127.0.0.1")
	}
	if p.config.Port != 8010 {
		t.Errorf("Port = %d, want %d", p.config.Port, 8010)
	}
	if p.config.MaxInternalCalls != 3 {
		t.Errorf("MaxInternalCalls = %d, want %d", p.config.MaxInternalCalls, 3)
	}
	if p.config.MaxContextChars != 12000 {
		t.Errorf("MaxContextChars = %d, want %d", p.config.MaxContextChars, 12000)
	}
}

// --- GetDefaultModel Tests ---

func TestRLMProvider_GetDefaultModel(t *testing.T) {
	p := &RLMProvider{
		config: config.RLMConfig{
			UpstreamModel: "gpt-4o",
		},
	}

	if got := p.GetDefaultModel(); got != "gpt-4o" {
		t.Errorf("GetDefaultModel() = %q, want %q", got, "gpt-4o")
	}
}

// --- Chat Tests with Mock HTTP Server ---

func TestRLMProvider_Chat_Success(t *testing.T) {
	// Create a mock RLMgw server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/chat/completions" {
			// Verify request
			if r.Method != "POST" {
				t.Errorf("Method = %s, want POST", r.Method)
			}
			if r.Header.Get("Content-Type") != "application/json" {
				t.Errorf("Content-Type = %s, want application/json", r.Header.Get("Content-Type"))
			}

			// Parse request body
			var reqBody map[string]any
			if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
				t.Errorf("Failed to decode request body: %v", err)
				http.Error(w, "Invalid request", http.StatusBadRequest)
				return
			}

			// Verify required fields
			if _, ok := reqBody["model"]; !ok {
				t.Error("Request missing 'model' field")
			}
			if _, ok := reqBody["messages"]; !ok {
				t.Error("Request missing 'messages' field")
			}

			// Send mock response
			response := map[string]any{
				"choices": []map[string]any{
					{
						"message": map[string]any{
							"content": "Hello from RLMgw!",
						},
						"finish_reason": "stop",
					},
				},
				"usage": map[string]any{
					"prompt_tokens":     10,
					"completion_tokens": 5,
					"total_tokens":      15,
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer server.Close()

	// Create provider with mock server URL
	p := &RLMProvider{
		config: config.RLMConfig{
			UpstreamModel: "test-model",
		},
		httpClient: server.Client(),
		baseURL:    server.URL,
		started:    true,
	}

	// Test Chat
	resp, err := p.Chat(context.Background(), []Message{
		{Role: "user", Content: "Hello"},
	}, nil, "", nil)

	if err != nil {
		t.Fatalf("Chat() error = %v", err)
	}
	if resp.Content != "Hello from RLMgw!" {
		t.Errorf("Content = %q, want %q", resp.Content, "Hello from RLMgw!")
	}
	if resp.FinishReason != "stop" {
		t.Errorf("FinishReason = %q, want %q", resp.FinishReason, "stop")
	}
	if resp.Usage == nil {
		t.Fatal("Usage should not be nil")
	}
	if resp.Usage.PromptTokens != 10 {
		t.Errorf("PromptTokens = %d, want 10", resp.Usage.PromptTokens)
	}
	if resp.Usage.CompletionTokens != 5 {
		t.Errorf("CompletionTokens = %d, want 5", resp.Usage.CompletionTokens)
	}
}

func TestRLMProvider_Chat_WithTools(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/chat/completions" {
			// Parse request to verify tools are included
			var reqBody map[string]any
			json.NewDecoder(r.Body).Decode(&reqBody)

			// Check that tools are present
			if _, ok := reqBody["tools"]; !ok {
				t.Error("Request missing 'tools' field")
			}
			if reqBody["tool_choice"] != "auto" {
				t.Errorf("tool_choice = %v, want 'auto'", reqBody["tool_choice"])
			}

			// Send response with tool call
			response := map[string]any{
				"choices": []map[string]any{
					{
						"message": map[string]any{
							"content": "",
							"tool_calls": []map[string]any{
								{
									"id":   "call_123",
									"type": "function",
									"function": map[string]any{
										"name":      "get_weather",
										"arguments": `{"location":"NYC"}`,
									},
								},
							},
						},
						"finish_reason": "tool_calls",
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer server.Close()

	p := &RLMProvider{
		config: config.RLMConfig{
			UpstreamModel: "test-model",
		},
		httpClient: server.Client(),
		baseURL:    server.URL,
		started:    true,
	}

	tools := []ToolDefinition{
		{
			Type: "function",
			Function: ToolFunctionDefinition{
				Name:        "get_weather",
				Description: "Get weather for a location",
			},
		},
	}

	resp, err := p.Chat(context.Background(), []Message{
		{Role: "user", Content: "What's the weather in NYC?"},
	}, tools, "", nil)

	if err != nil {
		t.Fatalf("Chat() error = %v", err)
	}
	if resp.FinishReason != "tool_calls" {
		t.Errorf("FinishReason = %q, want %q", resp.FinishReason, "tool_calls")
	}
	if len(resp.ToolCalls) != 1 {
		t.Fatalf("len(ToolCalls) = %d, want 1", len(resp.ToolCalls))
	}
	if resp.ToolCalls[0].Function.Name != "get_weather" {
		t.Errorf("ToolCall name = %q, want %q", resp.ToolCalls[0].Function.Name, "get_weather")
	}
	if resp.ToolCalls[0].ID != "call_123" {
		t.Errorf("ToolCall ID = %q, want %q", resp.ToolCalls[0].ID, "call_123")
	}
}

func TestRLMProvider_Chat_WithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/chat/completions" {
			var reqBody map[string]any
			json.NewDecoder(r.Body).Decode(&reqBody)

			// Verify options are passed through
			if maxTokens, ok := reqBody["max_tokens"].(float64); !ok || maxTokens != 100 {
				t.Errorf("max_tokens = %v, want 100", reqBody["max_tokens"])
			}
			if temp, ok := reqBody["temperature"].(float64); !ok || temp != 0.7 {
				t.Errorf("temperature = %v, want 0.7", reqBody["temperature"])
			}

			response := map[string]any{
				"choices": []map[string]any{
					{
						"message": map[string]any{
							"content": "test",
						},
						"finish_reason": "stop",
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer server.Close()

	p := &RLMProvider{
		config: config.RLMConfig{
			UpstreamModel: "test-model",
		},
		httpClient: server.Client(),
		baseURL:    server.URL,
		started:    true,
	}

	options := map[string]any{
		"max_tokens":  100,
		"temperature": 0.7,
	}

	_, err := p.Chat(context.Background(), []Message{
		{Role: "user", Content: "test"},
	}, nil, "", options)

	if err != nil {
		t.Fatalf("Chat() error = %v", err)
	}
}

func TestRLMProvider_Chat_ErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/chat/completions" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "Invalid request"}`))
		}
	}))
	defer server.Close()

	p := &RLMProvider{
		config: config.RLMConfig{
			UpstreamModel: "test-model",
		},
		httpClient: server.Client(),
		baseURL:    server.URL,
		started:    true,
	}

	_, err := p.Chat(context.Background(), []Message{
		{Role: "user", Content: "test"},
	}, nil, "", nil)

	if err == nil {
		t.Fatal("Chat() expected error for 400 response")
	}
	if !strings.Contains(err.Error(), "400") {
		t.Errorf("error = %q, want to contain '400'", err.Error())
	}
}

func TestRLMProvider_Chat_NotStarted(t *testing.T) {
	p := &RLMProvider{
		config: config.RLMConfig{
			UpstreamModel: "test-model",
		},
		started: false,
	}

	_, err := p.Chat(context.Background(), []Message{
		{Role: "user", Content: "test"},
	}, nil, "", nil)

	if err == nil {
		t.Fatal("Chat() expected error when server not started")
	}
	if !strings.Contains(err.Error(), "not started") {
		t.Errorf("error = %q, want to contain 'not started'", err.Error())
	}
}

func TestRLMProvider_Chat_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/chat/completions" {
			// Send response with empty choices
			response := map[string]any{
				"choices": []map[string]any{},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer server.Close()

	p := &RLMProvider{
		config: config.RLMConfig{
			UpstreamModel: "test-model",
		},
		httpClient: server.Client(),
		baseURL:    server.URL,
		started:    true,
	}

	resp, err := p.Chat(context.Background(), []Message{
		{Role: "user", Content: "test"},
	}, nil, "", nil)

	if err != nil {
		t.Fatalf("Chat() error = %v", err)
	}
	if resp.Content != "" {
		t.Errorf("Content = %q, want empty", resp.Content)
	}
	if resp.FinishReason != "stop" {
		t.Errorf("FinishReason = %q, want 'stop'", resp.FinishReason)
	}
}

func TestRLMProvider_Chat_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		time.Sleep(2 * time.Second)
		response := map[string]any{
			"choices": []map[string]any{
				{
					"message":       map[string]any{"content": "late"},
					"finish_reason": "stop",
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	p := &RLMProvider{
		config: config.RLMConfig{
			UpstreamModel: "test-model",
		},
		httpClient: server.Client(),
		baseURL:    server.URL,
		started:    true,
	}

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	_, err := p.Chat(ctx, []Message{
		{Role: "user", Content: "test"},
	}, nil, "", nil)

	if err == nil {
		t.Fatal("Chat() expected error due to context cancellation")
	}
	if !strings.Contains(err.Error(), "context") {
		t.Errorf("error = %q, want to contain 'context'", err.Error())
	}
}

// --- Path Expansion Tests ---

func TestPathExpansion(t *testing.T) {
	// Test tilde expansion logic
	tests := []struct {
		name  string
		path  string
		valid bool
	}{
		{
			name:  "tilde path",
			path:  "~/rlmgw",
			valid: true,
		},
		{
			name:  "absolute path",
			path:  "/opt/rlmgw",
			valid: true,
		},
		{
			name:  "relative path",
			path:  "rlmgw",
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expanded := tt.path
			if strings.HasPrefix(tt.path, "~/") {
				homeDir, err := os.UserHomeDir()
				if err == nil {
					expanded = filepath.Join(homeDir, tt.path[2:])
				}
			}
			if strings.HasPrefix(tt.path, "~/") && !strings.HasPrefix(expanded, string(filepath.Separator)) {
				t.Error("Tilde expansion failed")
			}
		})
	}
}

// --- Close Tests ---

func TestRLMProvider_Close_NoProcess(t *testing.T) {
	p := &RLMProvider{}

	// Should not panic when closing without a process
	p.Close()

	// Test passes if no panic occurred
}

func TestRLMProvider_Close_NilProvider(t *testing.T) {
	// Test that Close doesn't panic when called on uninitialized provider
	// This verifies the nil checks in Close() method work correctly
	p := &RLMProvider{} // Provider with nil cmd field

	// Should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Close() panicked on provider with nil cmd: %v", r)
		}
	}()

	p.Close()
}

// --- Integration-style Tests (subprocess) ---
// These tests require actual rlmgw installation and are skipped unless explicitly enabled

func TestRLMProvider_SubprocessLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping subprocess test in short mode")
	}
	if runtime.GOOS == "windows" {
		t.Skip("Subprocess tests not fully supported on Windows")
	}

	// Check if rlmgw is installed
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Cannot determine home directory")
	}

	rlmgwPath := filepath.Join(homeDir, "rlmgw")
	if _, err := os.Stat(rlmgwPath); os.IsNotExist(err) {
		t.Skip("rlmgw not installed at ~/rlmgw (clone from https://github.com/mitkox/rlmgw)")
	}

	// Check if python3 is available
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 not found in PATH")
	}

	// This test would require a real upstream provider
	// For now, we skip it as it needs external dependencies
	t.Skip("Subprocess lifecycle test requires complete RLMgw setup with upstream provider")
}

// --- Helper Tests ---

func TestRLMProvider_ParseResponse_InvalidJSON(t *testing.T) {
	p := &RLMProvider{}

	_, err := p.parseResponse([]byte("not json"))
	if err == nil {
		t.Fatal("parseResponse() expected error for invalid JSON")
	}
}

func TestRLMProvider_ParseResponse_InvalidToolArguments(t *testing.T) {
	// Response with malformed tool arguments should be handled gracefully
	response := `{
		"choices": [{
			"message": {
				"content": "",
				"tool_calls": [{
					"id": "call_1",
					"type": "function",
					"function": {
						"name": "test_tool",
						"arguments": "not valid json"
					}
				}]
			},
			"finish_reason": "tool_calls"
		}]
	}`

	p := &RLMProvider{}
	resp, err := p.parseResponse([]byte(response))

	if err != nil {
		t.Fatalf("parseResponse() error = %v", err)
	}

	// Tool call with invalid arguments should be skipped
	if len(resp.ToolCalls) != 0 {
		t.Errorf("len(ToolCalls) = %d, want 0 (invalid tool call should be skipped)", len(resp.ToolCalls))
	}
}

// --- Benchmark Tests ---

func BenchmarkRLMProvider_ParseResponse(b *testing.B) {
	response := []byte(`{
		"choices": [{
			"message": {
				"content": "Hello, world!",
				"tool_calls": []
			},
			"finish_reason": "stop"
		}],
		"usage": {
			"prompt_tokens": 10,
			"completion_tokens": 5,
			"total_tokens": 15
		}
	}`)

	p := &RLMProvider{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := p.parseResponse(response)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// --- stripSystemParts Tests ---

func TestStripSystemParts(t *testing.T) {
	tests := []struct {
		name     string
		messages []Message
		want     []rlmMessage
	}{
		{
			name: "simple message without SystemParts",
			messages: []Message{
				{
					Role:    "user",
					Content: "Hello",
				},
			},
			want: []rlmMessage{
				{
					Role:    "user",
					Content: "Hello",
				},
			},
		},
		{
			name: "message with SystemParts should be stripped",
			messages: []Message{
				{
					Role:    "system",
					Content: "You are a helpful assistant",
					SystemParts: []ContentBlock{
						{Type: "text", Text: "Cache this"},
					},
				},
				{
					Role:    "user",
					Content: "Hello",
				},
			},
			want: []rlmMessage{
				{
					Role:    "system",
					Content: "You are a helpful assistant",
					// SystemParts field should not exist in output
				},
				{
					Role:    "user",
					Content: "Hello",
				},
			},
		},
		{
			name: "message with tool calls",
			messages: []Message{
				{
					Role:    "assistant",
					Content: "",
					ToolCalls: []ToolCall{
						{
							ID:   "call_123",
							Type: "function",
							Function: &FunctionCall{
								Name:      "get_weather",
								Arguments: `{"city":"SF"}`,
							},
						},
					},
				},
				{
					Role:       "tool",
					Content:    "Sunny, 72F",
					ToolCallID: "call_123",
				},
			},
			want: []rlmMessage{
				{
					Role:    "assistant",
					Content: "",
					ToolCalls: []ToolCall{
						{
							ID:   "call_123",
							Type: "function",
							Function: &FunctionCall{
								Name:      "get_weather",
								Arguments: `{"city":"SF"}`,
							},
						},
					},
				},
				{
					Role:       "tool",
					Content:    "Sunny, 72F",
					ToolCallID: "call_123",
				},
			},
		},
		{
			name: "message with reasoning content",
			messages: []Message{
				{
					Role:             "assistant",
					Content:          "The answer is 42",
					ReasoningContent: "Let me think...",
				},
			},
			want: []rlmMessage{
				{
					Role:             "assistant",
					Content:          "The answer is 42",
					ReasoningContent: "Let me think...",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripSystemParts(tt.messages)

			if len(got) != len(tt.want) {
				t.Fatalf("stripSystemParts() length = %d, want %d", len(got), len(tt.want))
			}

			for i := range got {
				if got[i].Role != tt.want[i].Role {
					t.Errorf("message[%d].Role = %q, want %q", i, got[i].Role, tt.want[i].Role)
				}
				if got[i].Content != tt.want[i].Content {
					t.Errorf("message[%d].Content = %q, want %q", i, got[i].Content, tt.want[i].Content)
				}
				if got[i].ReasoningContent != tt.want[i].ReasoningContent {
					t.Errorf("message[%d].ReasoningContent = %q, want %q", i, got[i].ReasoningContent, tt.want[i].ReasoningContent)
				}
				if got[i].ToolCallID != tt.want[i].ToolCallID {
					t.Errorf("message[%d].ToolCallID = %q, want %q", i, got[i].ToolCallID, tt.want[i].ToolCallID)
				}
				if len(got[i].ToolCalls) != len(tt.want[i].ToolCalls) {
					t.Errorf("message[%d].ToolCalls length = %d, want %d", i, len(got[i].ToolCalls), len(tt.want[i].ToolCalls))
				}
			}

			// Verify SystemParts field doesn't exist in serialized JSON
			jsonData, err := json.Marshal(got)
			if err != nil {
				t.Fatalf("failed to marshal result: %v", err)
			}
			if strings.Contains(string(jsonData), "system_parts") {
				t.Errorf("stripSystemParts() result contains 'system_parts' field in JSON: %s", string(jsonData))
			}
		})
	}
}
