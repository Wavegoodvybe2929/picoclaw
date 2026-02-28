// PicoClaw - Ultra-lightweight personal AI agent
// License: MIT
//
// Copyright (c) 2026 PicoClaw contributors

package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/sipeed/picoclaw/pkg/config"
	"github.com/sipeed/picoclaw/pkg/logger"
)

// RLMProvider implements LLMProvider using RLMgw (Recursive Language Models Gateway).
// RLMgw is a Python-based OpenAI-compatible gateway that enables handling near-infinite
// contexts by intelligently selecting relevant context through recursive exploration
// before forwarding requests to an upstream OpenAI-compatible provider.
//
// The provider spawns an RLMgw subprocess on first use and maintains it for the lifetime
// of the provider instance. All requests are proxied through the RLMgw server which
// performs intelligent context selection before forwarding to the upstream provider.
type RLMProvider struct {
	config     config.RLMConfig
	cmd        *exec.Cmd
	httpClient *http.Client
	baseURL    string
	started    bool
}

// NewRLMProvider creates a new RLM provider with the given configuration.
// The provider will start the RLMgw subprocess on initialization.
func NewRLMProvider(cfg config.RLMConfig) (*RLMProvider, error) {
	if !cfg.Enabled {
		return nil, fmt.Errorf("RLM provider is not enabled in configuration")
	}

	if cfg.UpstreamBaseURL == "" {
		return nil, fmt.Errorf("upstream_base_url is required for RLM provider")
	}

	if cfg.UpstreamModel == "" {
		return nil, fmt.Errorf("upstream_model is required for RLM provider")
	}

	// Apply defaults
	if cfg.Host == "" {
		cfg.Host = "127.0.0.1"
	}
	if cfg.Port == 0 {
		cfg.Port = 8010
	}
	if cfg.MaxInternalCalls == 0 {
		cfg.MaxInternalCalls = 3
	}
	if cfg.MaxContextChars == 0 {
		cfg.MaxContextChars = 12000
	}

	p := &RLMProvider{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 300 * time.Second, // Longer timeout for RLM processing
		},
		baseURL: fmt.Sprintf("http://%s:%d", cfg.Host, cfg.Port),
	}

	// Start the RLMgw server
	if err := p.startServer(); err != nil {
		return nil, fmt.Errorf("failed to start RLMgw server: %w", err)
	}

	return p, nil
}

// startServer starts the RLMgw subprocess and waits for it to be ready.
func (p *RLMProvider) startServer() error {
	// Find python3 executable
	pythonPath := p.config.PythonPath
	if pythonPath == "" {
		var err error
		pythonPath, err = exec.LookPath("python3")
		if err != nil {
			return fmt.Errorf("python3 not found in PATH (set python_path in config): %w", err)
		}
	}

	// Expand ~ in python path
	if strings.HasPrefix(pythonPath, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("could not determine home directory: %w", err)
		}
		pythonPath = filepath.Join(homeDir, pythonPath[2:])
	}

	// Find rlmgw installation path
	rlmgwPath := p.config.RLMGWPath
	if rlmgwPath == "" {
		// Try default location: ~/rlmgw
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("could not determine home directory: %w", err)
		}
		rlmgwPath = filepath.Join(homeDir, "rlmgw")
	}

	// Expand ~ in path
	if strings.HasPrefix(rlmgwPath, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("could not determine home directory: %w", err)
		}
		rlmgwPath = filepath.Join(homeDir, rlmgwPath[2:])
	}

	// Verify rlmgw directory exists
	if _, err := os.Stat(rlmgwPath); os.IsNotExist(err) {
		return fmt.Errorf("rlmgw directory not found at %s (install from https://github.com/mitkox/rlmgw)", rlmgwPath)
	}

	// Setup workspace root
	workspaceRoot := p.config.WorkspaceRoot
	if workspaceRoot == "" {
		// Use current working directory as default
		var err error
		workspaceRoot, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("could not determine current directory: %w", err)
		}
	}

	// Expand ~ in workspace path
	if strings.HasPrefix(workspaceRoot, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("could not determine home directory: %w", err)
		}
		workspaceRoot = filepath.Join(homeDir, workspaceRoot[2:])
	}

	// Build command: python3 -m rlmgw.server
	p.cmd = exec.Command(pythonPath, "-m", "rlmgw.server")
	p.cmd.Dir = rlmgwPath

	// Set environment variables
	p.cmd.Env = append(os.Environ(),
		fmt.Sprintf("RLMGW_HOST=%s", p.config.Host),
		fmt.Sprintf("RLMGW_PORT=%d", p.config.Port),
		fmt.Sprintf("RLMGW_UPSTREAM_BASE_URL=%s", p.config.UpstreamBaseURL),
		fmt.Sprintf("RLMGW_UPSTREAM_MODEL=%s", p.config.UpstreamModel),
		fmt.Sprintf("RLMGW_REPO_ROOT=%s", workspaceRoot),
		fmt.Sprintf("RLMGW_USE_RLM_CONTEXT_SELECTION=%t", p.config.UseRLMSelection),
		fmt.Sprintf("RLMGW_MAX_INTERNAL_CALLS=%d", p.config.MaxInternalCalls),
		fmt.Sprintf("RLMGW_MAX_CONTEXT_PACK_CHARS=%d", p.config.MaxContextChars),
	)

	// Capture stderr for debugging
	var stderr bytes.Buffer
	p.cmd.Stderr = &stderr

	// Start the subprocess
	if err := p.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start rlmgw subprocess: %w", err)
	}

	logger.InfoF("RLMgw subprocess started", map[string]any{"pid": p.cmd.Process.Pid, "url": p.baseURL})

	// Wait for the server to be ready
	if err := p.waitForReady(30 * time.Second); err != nil {
		// Cleanup the process if startup failed
		_ = p.cmd.Process.Kill()
		_ = p.cmd.Wait()

		stderrStr := stderr.String()
		if stderrStr != "" {
			return fmt.Errorf("rlmgw server failed to start (stderr: %s): %w", stderrStr, err)
		}
		return fmt.Errorf("rlmgw server failed to start: %w", err)
	}

	p.started = true
	logger.InfoF("RLMgw server ready", map[string]any{"url": p.baseURL})
	return nil
}

// waitForReady polls the RLMgw /readyz endpoint until the server is ready or timeout occurs.
func (p *RLMProvider) waitForReady(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	retryInterval := 500 * time.Millisecond

	for time.Now().Before(deadline) {
		resp, err := p.httpClient.Get(p.baseURL + "/readyz")
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}

		time.Sleep(retryInterval)
	}

	return fmt.Errorf("timeout waiting for rlmgw server to be ready after %v", timeout)
}

// Chat implements LLMProvider.Chat by forwarding requests to the RLMgw server.
// RLMgw performs intelligent context selection before forwarding to the upstream provider.
func (p *RLMProvider) Chat(
	ctx context.Context,
	messages []Message,
	tools []ToolDefinition,
	model string,
	options map[string]any,
) (*LLMResponse, error) {
	if !p.started {
		return nil, fmt.Errorf("RLMgw server not started")
	}

	// Use upstream model if no model specified
	if model == "" {
		model = p.config.UpstreamModel
	}

	// Build OpenAI-compatible request
	requestBody := map[string]any{
		"model":    model,
		"messages": messages,
	}

	if len(tools) > 0 {
		requestBody["tools"] = tools
		requestBody["tool_choice"] = "auto"
	}

	// Add optional parameters
	if maxTokens, ok := options["max_tokens"].(int); ok && maxTokens > 0 {
		requestBody["max_tokens"] = maxTokens
	}
	if temperature, ok := options["temperature"].(float64); ok {
		requestBody["temperature"] = temperature
	}

	// Marshal request to JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/v1/chat/completions", bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request to RLMgw
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to rlmgw: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("rlmgw request failed (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse OpenAI-compatible response
	return p.parseResponse(body)
}

// parseResponse parses the OpenAI-compatible response from RLMgw.
func (p *RLMProvider) parseResponse(body []byte) (*LLMResponse, error) {
	var apiResponse struct {
		Choices []struct {
			Message struct {
				Content   string `json:"content"`
				ToolCalls []struct {
					ID       string `json:"id"`
					Type     string `json:"type"`
					Function struct {
						Name      string `json:"name"`
						Arguments string `json:"arguments"`
					} `json:"function"`
				} `json:"tool_calls"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Usage *UsageInfo `json:"usage"`
	}

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(apiResponse.Choices) == 0 {
		return &LLMResponse{
			Content:      "",
			FinishReason: "stop",
		}, nil
	}

	choice := apiResponse.Choices[0]

	// Parse tool calls if present
	toolCalls := make([]ToolCall, 0, len(choice.Message.ToolCalls))
	for _, tc := range choice.Message.ToolCalls {
		if tc.Type != "function" {
			continue
		}

		// Validate arguments JSON
		if tc.Function.Arguments != "" {
			var testUnmarshal map[string]any
			if err := json.Unmarshal([]byte(tc.Function.Arguments), &testUnmarshal); err != nil {
				logger.WarnF("Failed to parse tool call arguments", map[string]any{"error": err, "arguments": tc.Function.Arguments})
				continue
			}
		}

		toolCalls = append(toolCalls, ToolCall{
			ID:   tc.ID,
			Type: tc.Type,
			Function: &FunctionCall{
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments,
			},
		})
	}

	return &LLMResponse{
		Content:      choice.Message.Content,
		ToolCalls:    toolCalls,
		FinishReason: choice.FinishReason,
		Usage:        apiResponse.Usage,
	}, nil
}

// GetDefaultModel returns the configured upstream model.
func (p *RLMProvider) GetDefaultModel() string {
	return p.config.UpstreamModel
}

// Close implements StatefulProvider.Close by gracefully shutting down the RLMgw subprocess.
func (p *RLMProvider) Close() {
	if p.cmd == nil || p.cmd.Process == nil {
		return
	}

	logger.InfoF("Shutting down RLMgw subprocess", map[string]any{"pid": p.cmd.Process.Pid})

	// Send SIGINT for graceful shutdown
	if err := p.cmd.Process.Signal(os.Interrupt); err != nil {
		logger.WarnF("Failed to send interrupt signal to rlmgw", map[string]any{"error": err})
		// Force kill if interrupt fails
		_ = p.cmd.Process.Kill()
	}

	// Wait for process to exit (with timeout)
	done := make(chan error, 1)
	go func() {
		done <- p.cmd.Wait()
	}()

	select {
	case <-done:
		logger.Info("RLMgw subprocess exited cleanly")
	case <-time.After(5 * time.Second):
		logger.Warn("RLMgw subprocess did not exit in time, force killing")
		_ = p.cmd.Process.Kill()
		_ = p.cmd.Wait()
	}

	p.started = false
}
