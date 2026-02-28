package agent

import (
	"context"
	"testing"

	"github.com/sipeed/picoclaw/pkg/config"
	"github.com/sipeed/picoclaw/pkg/providers"
)

// mockProviderForLoopProfiles is a simple mock provider for testing loop profiles
type mockProviderForLoopProfiles struct{}

func (m *mockProviderForLoopProfiles) Chat(
	ctx context.Context,
	messages []providers.Message,
	tools []providers.ToolDefinition,
	model string,
	opts map[string]any,
) (*providers.LLMResponse, error) {
	return &providers.LLMResponse{
		Content:   "Mock response",
		ToolCalls: []providers.ToolCall{},
	}, nil
}

func (m *mockProviderForLoopProfiles) GetDefaultModel() string {
	return "mock-model"
}

func TestAgentInstance_LoopHooksResolution(t *testing.T) {
	// Mock provider
	mockProvider := &mockProviderForLoopProfiles{}

	defaults := &config.AgentDefaults{
		Workspace:         "~/.picoclaw/test",
		MaxTokens:         8192,
		MaxToolIterations: 20,
		LoopHooks: config.LoopHooks{
			BeforeLLM: []config.LoopHook{
				{Name: "fallback_hook", Enabled: true},
			},
		},
		LoopProfiles: map[string]config.LoopHooks{
			"default": {
				BeforeLLM: []config.LoopHook{
					{Name: "default_hook", Enabled: true},
				},
			},
			"memory": {
				BeforeLLM: []config.LoopHook{
					{Name: "memory_recall", Enabled: true},
				},
				AfterResponse: []config.LoopHook{
					{Name: "memory_write", Enabled: true},
				},
			},
		},
	}

	cfg := &config.Config{
		Agents: config.AgentsConfig{
			Defaults: *defaults,
		},
	}

	// Test 1: Agent with specific loop_profile
	agentCfg := &config.AgentConfig{
		ID:          "test_agent",
		LoopProfile: "memory",
	}

	instance := NewAgentInstance(agentCfg, defaults, cfg, mockProvider)

	if len(instance.LoopHooks.BeforeLLM) != 1 || instance.LoopHooks.BeforeLLM[0].Name != "memory_recall" {
		t.Errorf("Expected memory_recall hook, got %+v", instance.LoopHooks.BeforeLLM)
	}

	if len(instance.LoopHooks.AfterResponse) != 1 || instance.LoopHooks.AfterResponse[0].Name != "memory_write" {
		t.Errorf("Expected memory_write hook, got %+v", instance.LoopHooks.AfterResponse)
	}
}

func TestAgentInstance_LoopHooksResolution_DefaultProfile(t *testing.T) {
	mockProvider := &mockProviderForLoopProfiles{}

	defaults := &config.AgentDefaults{
		Workspace:         "~/.picoclaw/test",
		MaxTokens:         8192,
		MaxToolIterations: 20,
		LoopProfiles: map[string]config.LoopHooks{
			"default": {
				BeforeLLM: []config.LoopHook{
					{Name: "default_hook", Enabled: true},
				},
			},
		},
	}

	cfg := &config.Config{
		Agents: config.AgentsConfig{
			Defaults: *defaults,
		},
	}

	// Test 2: Agent without loop_profile (should use default)
	agentCfg := &config.AgentConfig{
		ID: "test_agent",
	}

	instance := NewAgentInstance(agentCfg, defaults, cfg, mockProvider)

	if len(instance.LoopHooks.BeforeLLM) != 1 || instance.LoopHooks.BeforeLLM[0].Name != "default_hook" {
		t.Errorf("Expected default_hook, got %+v", instance.LoopHooks.BeforeLLM)
	}
}

func TestAgentInstance_LoopHooksResolution_BackwardCompatibility(t *testing.T) {
	mockProvider := &mockProviderForLoopProfiles{}

	// Old config without loop_profiles
	defaults := &config.AgentDefaults{
		Workspace:         "~/.picoclaw/test",
		MaxTokens:         8192,
		MaxToolIterations: 20,
		LoopHooks: config.LoopHooks{
			BeforeLLM: []config.LoopHook{
				{Name: "old_style_hook", Enabled: true},
			},
		},
	}

	cfg := &config.Config{
		Agents: config.AgentsConfig{
			Defaults: *defaults,
		},
	}

	agentCfg := &config.AgentConfig{
		ID: "test_agent",
	}

	instance := NewAgentInstance(agentCfg, defaults, cfg, mockProvider)

	// Should fall back to loop_hooks
	if len(instance.LoopHooks.BeforeLLM) != 1 || instance.LoopHooks.BeforeLLM[0].Name != "old_style_hook" {
		t.Errorf("Backward compatibility: expected old_style_hook, got %+v", instance.LoopHooks.BeforeLLM)
	}
}

func TestAgentInstance_LoopHooksResolution_NonExistentProfile(t *testing.T) {
	mockProvider := &mockProviderForLoopProfiles{}

	defaults := &config.AgentDefaults{
		Workspace:         "~/.picoclaw/test",
		MaxTokens:         8192,
		MaxToolIterations: 20,
		LoopProfiles: map[string]config.LoopHooks{
			"default": {
				BeforeLLM: []config.LoopHook{
					{Name: "default_hook", Enabled: true},
				},
			},
		},
	}

	cfg := &config.Config{
		Agents: config.AgentsConfig{
			Defaults: *defaults,
		},
	}

	// Agent requesting non-existent profile
	agentCfg := &config.AgentConfig{
		ID:          "test_agent",
		LoopProfile: "nonexistent",
	}

	instance := NewAgentInstance(agentCfg, defaults, cfg, mockProvider)

	// Should fall back to default profile
	if len(instance.LoopHooks.BeforeLLM) != 1 || instance.LoopHooks.BeforeLLM[0].Name != "default_hook" {
		t.Errorf("Expected fallback to default profile, got %+v", instance.LoopHooks.BeforeLLM)
	}
}

func TestAgentInstance_LoopHooksResolution_NilAgent(t *testing.T) {
	mockProvider := &mockProviderForLoopProfiles{}

	defaults := &config.AgentDefaults{
		Workspace:         "~/.picoclaw/test",
		MaxTokens:         8192,
		MaxToolIterations: 20,
		LoopProfiles: map[string]config.LoopHooks{
			"default": {
				BeforeLLM: []config.LoopHook{
					{Name: "default_hook", Enabled: true},
				},
			},
		},
	}

	cfg := &config.Config{
		Agents: config.AgentsConfig{
			Defaults: *defaults,
		},
	}

	// Nil agent config (backward compatibility)
	instance := NewAgentInstance(nil, defaults, cfg, mockProvider)

	// Should use default profile
	if len(instance.LoopHooks.BeforeLLM) != 1 || instance.LoopHooks.BeforeLLM[0].Name != "default_hook" {
		t.Errorf("Expected default_hook for nil agent, got %+v", instance.LoopHooks.BeforeLLM)
	}
}

func TestAgentInstance_MultipleAgents_DifferentProfiles(t *testing.T) {
	mockProvider := &mockProviderForLoopProfiles{}

	defaults := &config.AgentDefaults{
		Workspace:         "~/.picoclaw/test",
		MaxTokens:         8192,
		MaxToolIterations: 20,
		LoopProfiles: map[string]config.LoopHooks{
			"default": {
				BeforeLLM: []config.LoopHook{
					{Name: "default_hook", Enabled: true},
				},
			},
			"memory": {
				BeforeLLM: []config.LoopHook{
					{Name: "memory_hook", Enabled: true},
				},
			},
			"debug": {
				OnToolCall: []config.LoopHook{
					{Name: "debug_hook", Enabled: true},
				},
			},
		},
	}

	cfg := &config.Config{
		Agents: config.AgentsConfig{
			Defaults: *defaults,
		},
	}

	// Create multiple agents with different profiles
	agent1 := &config.AgentConfig{
		ID:          "agent1",
		LoopProfile: "memory",
	}
	agent2 := &config.AgentConfig{
		ID:          "agent2",
		LoopProfile: "debug",
	}
	agent3 := &config.AgentConfig{
		ID: "agent3",
		// No profile specified - should use default
	}

	instance1 := NewAgentInstance(agent1, defaults, cfg, mockProvider)
	instance2 := NewAgentInstance(agent2, defaults, cfg, mockProvider)
	instance3 := NewAgentInstance(agent3, defaults, cfg, mockProvider)

	// Verify each agent has correct hooks
	if len(instance1.LoopHooks.BeforeLLM) != 1 || instance1.LoopHooks.BeforeLLM[0].Name != "memory_hook" {
		t.Errorf("Agent1 should have memory_hook, got %+v", instance1.LoopHooks.BeforeLLM)
	}

	if len(instance2.LoopHooks.OnToolCall) != 1 || instance2.LoopHooks.OnToolCall[0].Name != "debug_hook" {
		t.Errorf("Agent2 should have debug_hook, got %+v", instance2.LoopHooks.OnToolCall)
	}

	if len(instance3.LoopHooks.BeforeLLM) != 1 || instance3.LoopHooks.BeforeLLM[0].Name != "default_hook" {
		t.Errorf("Agent3 should have default_hook, got %+v", instance3.LoopHooks.BeforeLLM)
	}
}
