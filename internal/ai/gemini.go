// Package ai provides Gemini API format handlers
package ai

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

// GeminiHandler implements FormatHandler for Gemini API
type GeminiHandler struct{}

// NewGeminiHandler creates a new Gemini format handler
func NewGeminiHandler() *GeminiHandler {
	return &GeminiHandler{}
}

// BuildRequest builds a Gemini API request
func (h *GeminiHandler) BuildRequest(config RequestConfig) (map[string]interface{}, error) {
	contents := []map[string]interface{}{}

	// If messages are provided, convert them to Gemini format
	if len(config.Messages) > 0 {
		for _, msg := range config.Messages {
			role := msg["role"]
			content := msg["content"]

			// Skip empty messages
			if content == "" {
				continue
			}

			// Map roles to Gemini format
			geminiRole := "user"
			switch role {
			case "system", "user":
				geminiRole = "user"
			case "assistant":
				geminiRole = "model"
			}

			geminiContent := map[string]interface{}{
				"role": geminiRole,
				"parts": []map[string]string{
					{"text": content},
				},
			}
			contents = append(contents, geminiContent)
		}
	} else {
		// Build from system and user prompts
		// Add user message
		userContent := map[string]interface{}{
			"role": "user",
			"parts": []map[string]string{
				{"text": config.UserPrompt},
			},
		}
		contents = append(contents, userContent)
	}

	// Build request body
	request := map[string]interface{}{
		"contents": contents,
		"generationConfig": map[string]interface{}{
			"temperature":     0.3,
			"maxOutputTokens": 2048,
		},
	}

	// Override defaults if provided
	if config.Temperature > 0 {
		request["generationConfig"].(map[string]interface{})["temperature"] = config.Temperature
	}
	if config.MaxTokens > 0 {
		request["generationConfig"].(map[string]interface{})["maxOutputTokens"] = config.MaxTokens
	}

	// Add system instruction if provided (Gemini-specific)
	if config.SystemPrompt != "" {
		request["systemInstruction"] = map[string]interface{}{
			"role": "user",
			"parts": []map[string]string{
				{"text": config.SystemPrompt},
			},
		}
	}

	return request, nil
}

// ParseResponse parses a Gemini API response
func (h *GeminiHandler) ParseResponse(body []byte) (ResponseResult, error) {
	var response struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
			FinishReason string `json:"finishReason"`
		} `json:"candidates"`
		PromptFeedback struct {
			BlockReason string `json:"blockReason,omitempty"`
		} `json:"promptFeedback"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return ResponseResult{}, fmt.Errorf("failed to decode Gemini response: %w", err)
	}

	// Check if prompt was blocked
	if response.PromptFeedback.BlockReason != "" {
		return ResponseResult{}, fmt.Errorf("prompt blocked: %s", response.PromptFeedback.BlockReason)
	}

	// Check if we have candidates
	if len(response.Candidates) == 0 {
		return ResponseResult{}, fmt.Errorf("no candidates in Gemini response")
	}

	// Get the first candidate's content
	candidate := response.Candidates[0]
	if len(candidate.Content.Parts) == 0 {
		return ResponseResult{}, fmt.Errorf("no parts in candidate content")
	}

	// Check finish reason
	if candidate.FinishReason == "SAFETY" {
		return ResponseResult{}, fmt.Errorf("response blocked for safety reasons")
	}
	if candidate.FinishReason == "RECITATION" {
		return ResponseResult{}, fmt.Errorf("response blocked for recitation reasons")
	}

	content := strings.TrimSpace(candidate.Content.Parts[0].Text)
	return ResponseResult{
		Content:    content,
		FormatUsed: FormatTypeGemini,
	}, nil
}

// ValidateResponse validates the HTTP response status
func (h *GeminiHandler) ValidateResponse(statusCode int, body []byte) error {
	switch statusCode {
	case http.StatusOK:
		return nil
	case http.StatusUnauthorized, http.StatusForbidden:
		return fmt.Errorf("Gemini authentication failed")
	case http.StatusNotFound:
		return fmt.Errorf("Gemini model not found")
	default:
		return fmt.Errorf("Gemini API returned status %d: %s", statusCode, string(body))
	}
}

// FormatEndpoint formats the endpoint URL for Gemini API
func (h *GeminiHandler) FormatEndpoint(endpoint, model string) string {
	return FormatGeminiEndpoint(endpoint, model)
}

// DetectAPIProvider detects the AI provider from the endpoint URL
// Returns "gemini", "openai", or "unknown"
func DetectAPIProvider(endpoint string) string {
	endpoint = strings.ToLower(endpoint)

	// Gemini API endpoints
	if strings.Contains(endpoint, "googleapis.com") ||
		strings.Contains(endpoint, "generativelanguage.googleapis.com") ||
		strings.Contains(endpoint, "gemini") {
		return "gemini"
	}

	// OpenAI-compatible endpoints (default)
	if strings.Contains(endpoint, "openai.com") ||
		strings.Contains(endpoint, "api.openai.com") {
		return "openai"
	}

	return "unknown"
}

// IsGeminiEndpoint checks if the given endpoint is a Gemini API endpoint
func IsGeminiEndpoint(endpoint string) bool {
	return DetectAPIProvider(endpoint) == "gemini"
}

// IsGeminiError checks if an error message indicates a Gemini API format mismatch
func IsGeminiError(errorMessage string) bool {
	// Check for common Gemini error messages
	geminiErrorPatterns := []string{
		"Unknown name \"prompt\"",
		"Unknown name \"messages\"",
		"Cannot find field",
		"INVALID_ARGUMENT",
		"generativelanguage.googleapis.com",
	}

	for _, pattern := range geminiErrorPatterns {
		if strings.Contains(errorMessage, pattern) {
			return true
		}
	}

	return false
}

// ExtractModelFromEndpoint extracts the model name from a Gemini endpoint
// For example: "https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:generateContent"
// returns "gemini-pro"
func ExtractModelFromEndpoint(endpoint string) string {
	// Gemini endpoint format: .../models/{model}:generateContent
	re := regexp.MustCompile(`/models/([^:]+)`)
	matches := re.FindStringSubmatch(endpoint)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// FormatGeminiEndpoint formats a Gemini endpoint with the given model
// If the endpoint already contains a model, it's replaced
func FormatGeminiEndpoint(baseEndpoint, model string) string {
	// If endpoint already contains :generateContent or :streamGenerateContent, replace the model
	if strings.Contains(baseEndpoint, ":generateContent") {
		// Use simple pattern to replace model name
		re := regexp.MustCompile(`/models/[^:]+:generateContent`)
		if re.MatchString(baseEndpoint) {
			return re.ReplaceAllString(baseEndpoint, "/models/"+model+":generateContent")
		}
		return baseEndpoint
	}

	// If endpoint doesn't contain :generateContent, add it
	baseEndpoint = strings.TrimSuffix(baseEndpoint, "/")

	// If model is already in the endpoint, just add the method
	if strings.Contains(baseEndpoint, "/models/"+model) {
		return baseEndpoint + ":generateContent"
	}

	// Add model and method
	if !strings.Contains(baseEndpoint, "/models/") {
		return baseEndpoint + "/models/" + model + ":generateContent"
	}

	return baseEndpoint + ":generateContent"
}
