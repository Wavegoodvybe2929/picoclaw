// PicoClaw - Ultra-lightweight personal AI agent
// License: MIT
//
// Copyright (c) 2026 PicoClaw contributors

package providers

import (
	"testing"

	"github.com/sipeed/picoclaw/pkg/config"
)

// TestCreateProviderFromConfig_RLMProtocol tests that rlm/ protocol prefix returns helpful error
func TestCreateProviderFromConfig_RLMProtocol(t *testing.T) {
	t.Skip("DISABLED: RLM provider features are commented out in code")

	cfg := &config.ModelConfig{
		ModelName: "test-rlm",
		Model:     "rlm/gpt-4o",
	}

	_, _, err := CreateProviderFromConfig(cfg)
	if err == nil {
		t.Fatal("CreateProviderFromConfig() expected error for rlm/ protocol")
	}

	// Should suggest using provider="rlm" instead
	if !containsIgnoreCase(err.Error(), "provider") || !containsIgnoreCase(err.Error(), "agents.defaults") {
		t.Errorf("expected error message to mention using provider in agents.defaults, got: %v", err)
	}
}

// TestCreateProvider_RLMEnabled tests successful RLM provider creation
func TestCreateProvider_RLMEnabled(t *testing.T) {
	// Skip if rlmgw is not installed (integration test requirement)
	// This test will be run during integration testing
	t.Skip("Requires rlmgw installation - run with integration tests")

	cfg := &config.Config{
		Agents: config.AgentsConfig{
			Defaults: config.AgentDefaults{
				Provider:  "rlm",
				ModelName: "gpt-4o",
			},
		},
		Providers: config.ProvidersConfig{
			RLM: config.RLMConfig{
				Enabled:         true,
				UpstreamBaseURL: "http://localhost:1234/v1",
				UpstreamModel:   "gpt-4o",
			},
		},
	}

	// Note: This will attempt to spawn rlmgw subprocess
	provider, modelID, err := CreateProvider(cfg)
	if err != nil {
		t.Fatalf("CreateProvider() error = %v", err)
	}
	if provider == nil {
		t.Fatal("CreateProvider() returned nil provider")
	}
	if modelID != "gpt-4o" {
		t.Errorf("modelID = %q, want %q", modelID, "gpt-4o")
	}

	// Clean up
	if statefulProvider, ok := provider.(StatefulProvider); ok {
		statefulProvider.Close()
	}
}

// TestCreateProvider_RLMNotEnabled tests error when RLM is not enabled
func TestCreateProvider_RLMNotEnabled(t *testing.T) {
	t.Skip("DISABLED: RLM provider features are commented out in code")

	cfg := &config.Config{
		Agents: config.AgentsConfig{
			Defaults: config.AgentDefaults{
				Provider:  "rlm",
				ModelName: "gpt-4o",
			},
		},
		Providers: config.ProvidersConfig{
			RLM: config.RLMConfig{
				Enabled: false, // Not enabled
			},
		},
	}

	_, _, err := CreateProvider(cfg)
	if err == nil {
		t.Fatal("CreateProvider() expected error when RLM not enabled")
	}
	if !containsIgnoreCase(err.Error(), "not enabled") {
		t.Errorf("expected error message to mention 'not enabled', got: %v", err)
	}
}

// TestCreateProvider_RLMMissingUpstreamURL tests error when required config is missing
func TestCreateProvider_RLMMissingUpstreamURL(t *testing.T) {
	t.Skip("DISABLED: RLM provider features are commented out in code")

	cfg := &config.Config{
		Agents: config.AgentsConfig{
			Defaults: config.AgentDefaults{
				Provider:  "rlm",
				ModelName: "gpt-4o",
			},
		},
		Providers: config.ProvidersConfig{
			RLM: config.RLMConfig{
				Enabled: true,
				// Missing UpstreamBaseURL
				UpstreamModel: "gpt-4o",
			},
		},
	}

	_, _, err := CreateProvider(cfg)
	if err == nil {
		t.Fatal("CreateProvider() expected error when upstream_base_url missing")
	}
	if !containsIgnoreCase(err.Error(), "upstream_base_url") {
		t.Errorf("expected error message to mention 'upstream_base_url', got: %v", err)
	}
}

// TestCreateProvider_RLMMissingUpstreamModel tests error when upstream model is missing
func TestCreateProvider_RLMMissingUpstreamModel(t *testing.T) {
	t.Skip("DISABLED: RLM provider features are commented out in code")

	cfg := &config.Config{
		Agents: config.AgentsConfig{
			Defaults: config.AgentDefaults{
				Provider:  "rlm",
				ModelName: "gpt-4o",
			},
		},
		Providers: config.ProvidersConfig{
			RLM: config.RLMConfig{
				Enabled:         true,
				UpstreamBaseURL: "http://localhost:1234/v1",
				// Missing UpstreamModel
			},
		},
	}

	_, _, err := CreateProvider(cfg)
	if err == nil {
		t.Fatal("CreateProvider() expected error when upstream_model missing")
	}
	if !containsIgnoreCase(err.Error(), "upstream_model") {
		t.Errorf("expected error message to mention 'upstream_model', got: %v", err)
	}
}

// TestCreateProvider_RLMUsesUpstreamModel tests that upstream model is returned as modelID
func TestCreateProvider_RLMUsesUpstreamModel(t *testing.T) {
	// Skip if rlmgw is not installed
	t.Skip("Requires rlmgw installation - run with integration tests")

	cfg := &config.Config{
		Agents: config.AgentsConfig{
			Defaults: config.AgentDefaults{
				Provider:  "rlm",
				ModelName: "anything", // Should be ignored
			},
		},
		Providers: config.ProvidersConfig{
			RLM: config.RLMConfig{
				Enabled:         true,
				UpstreamBaseURL: "http://localhost:1234/v1",
				UpstreamModel:   "custom-model-name",
			},
		},
	}

	provider, modelID, err := CreateProvider(cfg)
	if err != nil {
		t.Fatalf("CreateProvider() error = %v", err)
	}
	if modelID != "custom-model-name" {
		t.Errorf("modelID = %q, want %q", modelID, "custom-model-name")
	}

	// Clean up
	if statefulProvider, ok := provider.(StatefulProvider); ok {
		statefulProvider.Close()
	}
}

// TestCreateProvider_NonRLMProviderStillWorks tests that non-RLM providers are unaffected
func TestCreateProvider_NonRLMProviderStillWorks(t *testing.T) {
	cfg := &config.Config{
		Agents: config.AgentsConfig{
			Defaults: config.AgentDefaults{
				Provider:  "", // No explicit provider
				ModelName: "test-model",
			},
		},
		ModelList: []config.ModelConfig{
			{
				ModelName: "test-model",
				Model:     "openai/gpt-4o",
				APIKey:    "test-key",
				APIBase:   "https://api.example.com/v1",
			},
		},
	}

	provider, modelID, err := CreateProvider(cfg)
	if err != nil {
		t.Fatalf("CreateProvider() error = %v, expected non-RLM provider to work", err)
	}
	if provider == nil {
		t.Fatal("CreateProvider() returned nil provider for non-RLM config")
	}
	if modelID != "gpt-4o" {
		t.Errorf("modelID = %q, want %q", modelID, "gpt-4o")
	}
}

// containsIgnoreCase checks if s contains substr (case-insensitive)
func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > 0 && len(substr) > 0 &&
				findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	sLower := toLower(s)
	substrLower := toLower(substr)
	for i := 0; i <= len(sLower)-len(substrLower); i++ {
		if sLower[i:i+len(substrLower)] == substrLower {
			return true
		}
	}
	return false
}

func toLower(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			b[i] = c + ('a' - 'A')
		} else {
			b[i] = c
		}
	}
	return string(b)
}
