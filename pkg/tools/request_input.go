package tools

import (
	"context"
	"fmt"

	"github.com/sipeed/picoclaw/pkg/bus"
	"github.com/sipeed/picoclaw/pkg/config"
)

// InputRequestCallback is called when the agent needs to request user input.
// It returns the user's response or an error.
type InputRequestCallback func(ctx context.Context, prompt string) (string, error)

// RequestInputTool enables agents to pause and request additional input from the user.
// This creates an interactive workflow where the agent can ask clarification questions,
// request confirmations, or gather structured input mid-conversation.
type RequestInputTool struct {
	requestCallback InputRequestCallback
	defaultChannel  string
	defaultChatID   string
}

// NewRequestInputTool creates a new request_input tool.
func NewRequestInputTool() *RequestInputTool {
	return &RequestInputTool{}
}

func (t *RequestInputTool) Name() string {
	return "request_input"
}

func (t *RequestInputTool) Description() string {
	return "Request additional input from the user. Use this when you need clarification, confirmation, or additional information. The conversation will pause until the user responds or timeout expires."
}

func (t *RequestInputTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"prompt": map[string]any{
				"type":        "string",
				"description": "The question or prompt to show the user. Be clear and specific.",
			},
		},
		"required": []string{"prompt"},
	}
}

// SetContext sets the default channel and chat ID for this tool.
func (t *RequestInputTool) SetContext(channel, chatID string) {
	t.defaultChannel = channel
	t.defaultChatID = chatID
}

// SetRequestCallback sets the callback function that handles the actual input request.
func (t *RequestInputTool) SetRequestCallback(callback InputRequestCallback) {
	t.requestCallback = callback
}

// Execute handles the tool execution.
// Note: The actual input request logic is handled by the agent loop using hooks.
// This tool serves as an interface for the LLM to trigger input requests.
func (t *RequestInputTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	prompt, ok := args["prompt"].(string)
	if !ok || prompt == "" {
		return &ToolResult{
			ForLLM:  "prompt is required and must be a non-empty string",
			IsError: true,
		}
	}

	if t.requestCallback == nil {
		return &ToolResult{
			ForLLM:  "Input request not configured - request_input hooks may not be enabled",
			IsError: true,
		}
	}

	// Call the request callback which will execute the hooks
	userInput, err := t.requestCallback(ctx, prompt)
	if err != nil {
		return &ToolResult{
			ForLLM:  fmt.Sprintf("Failed to request user input: %v", err),
			IsError: true,
			Err:     err,
		}
	}

	// Return the user's input to the LLM
	return &ToolResult{
		ForLLM:  fmt.Sprintf("User responded: %s", userInput),
		ForUser: "", // Don't send anything to user (they already saw the prompt and responded)
		Silent:  true,
	}
}

// InputRequestExecutor provides the implementation for handling input requests via hooks.
type InputRequestExecutor struct {
	hookExecutor *interface{} // Will be set to *agent.HookExecutor at runtime
	msgBus       *bus.MessageBus
	hooks        []config.LoopHook
}

// NewInputRequestExecutor creates a new executor for handling input requests.
func NewInputRequestExecutor(msgBus *bus.MessageBus, hooks []config.LoopHook) *InputRequestExecutor {
	return &InputRequestExecutor{
		msgBus: msgBus,
		hooks:  hooks,
	}
}
