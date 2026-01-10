// Package ai provides Ollama API format handlers
package ai

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// OllamaHandler implements FormatHandler for Ollama API
type OllamaHandler struct{}

// NewOllamaHandler creates a new Ollama format handler
func NewOllamaHandler() *OllamaHandler {
	return &OllamaHandler{}
}

// BuildRequest builds an Ollama API request
// Note: Ollama uses a simple prompt format instead of messages
func (h *OllamaHandler) BuildRequest(config RequestConfig) (map[string]interface{}, error) {
	// Combine system and user prompts for Ollama
	var fullPrompt strings.Builder

	if config.SystemPrompt != "" {
		fullPrompt.WriteString(config.SystemPrompt)
		fullPrompt.WriteString("\n\n")
	}

	if config.UserPrompt != "" {
		fullPrompt.WriteString(config.UserPrompt)
	} else if len(config.Messages) > 0 {
		// If messages are provided, convert them to prompt format
		for _, msg := range config.Messages {
			role := msg["role"]
			content := msg["content"]
			if content == "" {
				continue
			}
			fullPrompt.WriteString(role)
			fullPrompt.WriteString(": ")
			fullPrompt.WriteString(content)
			fullPrompt.WriteString("\n")
		}
	}

	request := map[string]interface{}{
		"model":  config.Model,
		"prompt": fullPrompt.String(),
		"stream": false,
	}

	// Ollama doesn't typically use temperature/max_tokens in basic requests,
	// but we can include them if provided
	if config.Temperature > 0 {
		request["temperature"] = config.Temperature
	}
	if config.MaxTokens > 0 {
		request["num_predict"] = config.MaxTokens
	}

	return request, nil
}

// ParseResponse parses an Ollama API response
func (h *OllamaHandler) ParseResponse(body []byte) (ResponseResult, error) {
	var response struct {
		Response string `json:"response"`
		Done     bool   `json:"done"`
		Error    string `json:"error,omitempty"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return ResponseResult{}, fmt.Errorf("failed to decode Ollama response: %w", err)
	}

	// Check for Ollama error
	if response.Error != "" {
		return ResponseResult{}, fmt.Errorf("Ollama API error: %s", response.Error)
	}

	// Check if response is complete
	if !response.Done {
		return ResponseResult{}, fmt.Errorf("Ollama response incomplete (done=false)")
	}

	content := strings.TrimSpace(response.Response)
	if content == "" {
		return ResponseResult{}, fmt.Errorf("empty response from Ollama")
	}

	return ResponseResult{
		Content:    content,
		FormatUsed: FormatTypeOllama,
	}, nil
}

// ValidateResponse validates the HTTP response status
func (h *OllamaHandler) ValidateResponse(statusCode int, body []byte) error {
	switch statusCode {
	case http.StatusOK:
		return nil
	case http.StatusUnauthorized:
		return fmt.Errorf("Ollama authentication failed")
	case http.StatusNotFound:
		return fmt.Errorf("Ollama endpoint or model not found")
	default:
		return fmt.Errorf("Ollama API returned status %d: %s", statusCode, string(body))
	}
}

// FormatEndpoint returns the endpoint as-is for Ollama format
func (h *OllamaHandler) FormatEndpoint(endpoint, model string) string {
	return strings.TrimSuffix(endpoint, "/")
}

// IsOllamaError checks if an error message indicates an Ollama API format
func IsOllamaError(errorMessage string) bool {
	ollamaErrorPatterns := []string{
		"ollama",
		"model not found",
		"pull model",
		"invalid model",
	}

	for _, pattern := range ollamaErrorPatterns {
		if strings.Contains(strings.ToLower(errorMessage), pattern) {
			return true
		}
	}

	return false
}
