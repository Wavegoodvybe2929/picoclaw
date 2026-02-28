package config

import (
	"encoding/json"
	"testing"
)

func TestResolveLoopHooks_RequestedProfile(t *testing.T) {
	defaults := AgentDefaults{
		LoopHooks: LoopHooks{
			BeforeLLM: []LoopHook{
				{Name: "old_hook", Enabled: true},
			},
		},
		LoopProfiles: map[string]LoopHooks{
			"default": {
				BeforeLLM: []LoopHook{
					{Name: "default_hook", Enabled: true},
				},
			},
			"memory_enabled": {
				BeforeLLM: []LoopHook{
					{Name: "memory_recall", Enabled: true},
				},
				AfterResponse: []LoopHook{
					{Name: "memory_write", Enabled: true},
				},
			},
		},
	}

	// Test requesting a specific profile
	resolved := defaults.ResolveLoopHooks("memory_enabled")
	if len(resolved.BeforeLLM) != 1 || resolved.BeforeLLM[0].Name != "memory_recall" {
		t.Errorf("Expected memory_recall hook, got %+v", resolved.BeforeLLM)
	}
	if len(resolved.AfterResponse) != 1 || resolved.AfterResponse[0].Name != "memory_write" {
		t.Errorf("Expected memory_write hook, got %+v", resolved.AfterResponse)
	}
}

func TestResolveLoopHooks_DefaultProfile(t *testing.T) {
	defaults := AgentDefaults{
		LoopHooks: LoopHooks{
			BeforeLLM: []LoopHook{
				{Name: "old_hook", Enabled: true},
			},
		},
		LoopProfiles: map[string]LoopHooks{
			"default": {
				BeforeLLM: []LoopHook{
					{Name: "default_hook", Enabled: true},
				},
			},
		},
	}

	// Test requesting empty profile name - should use "default"
	resolved := defaults.ResolveLoopHooks("")
	if len(resolved.BeforeLLM) != 1 || resolved.BeforeLLM[0].Name != "default_hook" {
		t.Errorf("Expected default_hook, got %+v", resolved.BeforeLLM)
	}
}

func TestResolveLoopHooks_FallbackToLoopHooks(t *testing.T) {
	defaults := AgentDefaults{
		LoopHooks: LoopHooks{
			BeforeLLM: []LoopHook{
				{Name: "old_hook", Enabled: true},
			},
		},
		LoopProfiles: nil, // No profiles defined
	}

	// Should fall back to loop_hooks field
	resolved := defaults.ResolveLoopHooks("")
	if len(resolved.BeforeLLM) != 1 || resolved.BeforeLLM[0].Name != "old_hook" {
		t.Errorf("Expected old_hook from loop_hooks, got %+v", resolved.BeforeLLM)
	}
}

func TestResolveLoopHooks_NonExistentProfile(t *testing.T) {
	defaults := AgentDefaults{
		LoopHooks: LoopHooks{
			BeforeLLM: []LoopHook{
				{Name: "fallback_hook", Enabled: true},
			},
		},
		LoopProfiles: map[string]LoopHooks{
			"default": {
				BeforeLLM: []LoopHook{
					{Name: "default_hook", Enabled: true},
				},
			},
		},
	}

	// Request non-existent profile - should fall back to "default"
	resolved := defaults.ResolveLoopHooks("nonexistent")
	if len(resolved.BeforeLLM) != 1 || resolved.BeforeLLM[0].Name != "default_hook" {
		t.Errorf("Expected fallback to default profile, got %+v", resolved.BeforeLLM)
	}
}

func TestResolveLoopHooks_NoProfilesNoLoopHooks(t *testing.T) {
	defaults := AgentDefaults{
		LoopHooks:    LoopHooks{},
		LoopProfiles: nil,
	}

	// Should return empty LoopHooks
	resolved := defaults.ResolveLoopHooks("")
	if len(resolved.BeforeLLM) != 0 {
		t.Errorf("Expected empty hooks, got %+v", resolved.BeforeLLM)
	}
}

func TestLoopProfiles_JSONParsing(t *testing.T) {
	jsonData := `{
		"workspace": "~/.picoclaw",
		"restrict_to_workspace": true,
		"max_tokens": 8192,
		"max_tool_iterations": 20,
		"loop_profiles": {
			"default": {
				"before_llm": [
					{
						"name": "test_hook",
						"command": "echo test",
						"enabled": true
					}
				]
			},
			"memory": {
				"before_llm": [
					{
						"name": "memory_recall",
						"command": "./bin/memory_recall",
						"enabled": true
					}
				],
				"after_response": [
					{
						"name": "memory_write",
						"command": "./bin/memory_write",
						"enabled": true
					}
				]
			}
		}
	}`

	var defaults AgentDefaults
	err := json.Unmarshal([]byte(jsonData), &defaults)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if defaults.LoopProfiles == nil {
		t.Fatal("LoopProfiles is nil")
	}

	if len(defaults.LoopProfiles) != 2 {
		t.Errorf("Expected 2 profiles, got %d", len(defaults.LoopProfiles))
	}

	defaultProfile, ok := defaults.LoopProfiles["default"]
	if !ok {
		t.Fatal("default profile not found")
	}

	if len(defaultProfile.BeforeLLM) != 1 {
		t.Errorf("Expected 1 hook in default profile, got %d", len(defaultProfile.BeforeLLM))
	}

	if defaultProfile.BeforeLLM[0].Name != "test_hook" {
		t.Errorf("Expected test_hook, got %s", defaultProfile.BeforeLLM[0].Name)
	}

	memoryProfile, ok := defaults.LoopProfiles["memory"]
	if !ok {
		t.Fatal("memory profile not found")
	}

	if len(memoryProfile.BeforeLLM) != 1 || len(memoryProfile.AfterResponse) != 1 {
		t.Errorf("Memory profile has wrong number of hooks")
	}
}

func TestAgentConfig_LoopProfile(t *testing.T) {
	jsonData := `{
		"id": "test_agent",
		"loop_profile": "memory_enabled"
	}`

	var agentCfg AgentConfig
	err := json.Unmarshal([]byte(jsonData), &agentCfg)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if agentCfg.LoopProfile != "memory_enabled" {
		t.Errorf("Expected loop_profile='memory_enabled', got '%s'", agentCfg.LoopProfile)
	}
}

func TestLoopProfiles_BackwardCompatibility(t *testing.T) {
	// Old config format without loop_profiles
	jsonData := `{
		"workspace": "~/.picoclaw",
		"loop_hooks": {
			"before_llm": [
				{
					"name": "old_hook",
					"command": "echo old",
					"enabled": true
				}
			]
		}
	}`

	var defaults AgentDefaults
	err := json.Unmarshal([]byte(jsonData), &defaults)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Should work with old format
	if len(defaults.LoopHooks.BeforeLLM) != 1 {
		t.Errorf("Expected 1 hook in loop_hooks, got %d", len(defaults.LoopHooks.BeforeLLM))
	}

	// Resolution should fall back to loop_hooks
	resolved := defaults.ResolveLoopHooks("")
	if len(resolved.BeforeLLM) != 1 || resolved.BeforeLLM[0].Name != "old_hook" {
		t.Errorf("Backward compatibility broken: expected old_hook, got %+v", resolved.BeforeLLM)
	}
}

func TestLoopProfiles_MixedConfig(t *testing.T) {
	// Config with both loop_hooks and loop_profiles
	jsonData := `{
		"workspace": "~/.picoclaw",
		"loop_hooks": {
			"before_llm": [
				{
					"name": "old_hook",
					"command": "echo old",
					"enabled": true
				}
			]
		},
		"loop_profiles": {
			"new_profile": {
				"before_llm": [
					{
						"name": "new_hook",
						"command": "echo new",
						"enabled": true
					}
				]
			}
		}
	}`

	var defaults AgentDefaults
	err := json.Unmarshal([]byte(jsonData), &defaults)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// When profiles exist, they should take precedence
	resolved := defaults.ResolveLoopHooks("new_profile")
	if len(resolved.BeforeLLM) != 1 || resolved.BeforeLLM[0].Name != "new_hook" {
		t.Errorf("Expected new_hook from profile, got %+v", resolved.BeforeLLM)
	}

	// If no matching profile, fall back to loop_hooks
	resolvedFallback := defaults.ResolveLoopHooks("nonexistent")
	if len(resolvedFallback.BeforeLLM) != 1 || resolvedFallback.BeforeLLM[0].Name != "old_hook" {
		t.Errorf("Expected fallback to loop_hooks, got %+v", resolvedFallback.BeforeLLM)
	}
}
