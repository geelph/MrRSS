package label

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"MrRSS/internal/config"
)

// AILabeler implements label generation using OpenAI-compatible APIs
type AILabeler struct {
	APIKey       string
	Endpoint     string
	Model        string
	SystemPrompt string
	client       *http.Client
}

// NewAILabeler creates a new AI labeler with the given credentials
func NewAILabeler(apiKey, endpoint, model string) *AILabeler {
	defaults := config.Get()
	// Use label-specific endpoint and model
	if endpoint == "" {
		endpoint = defaults.SummaryAIEndpoint // Reuse summary AI endpoint as default
	}
	if model == "" {
		model = defaults.SummaryAIModel // Reuse summary AI model as default
	}
	return &AILabeler{
		APIKey:       apiKey,
		Endpoint:     strings.TrimSuffix(endpoint, "/"),
		Model:        model,
		SystemPrompt: "", // Will be set from settings when used
		client:       &http.Client{Timeout: 30 * time.Second},
	}
}

// SetSystemPrompt sets a custom system prompt for the labeler
func (a *AILabeler) SetSystemPrompt(prompt string) {
	a.SystemPrompt = prompt
}

// GenerateLabels generates labels from the given text using an AI API
func (a *AILabeler) GenerateLabels(text string, maxLabels int) (LabelResult, error) {
	// Clean the text first
	cleanedText := cleanText(text)

	// Check if text is too short
	if len(cleanedText) < MinContentLength {
		return LabelResult{
			Labels:     []string{},
			IsTooShort: true,
		}, nil
	}

	// Validate maxLabels parameter
	if maxLabels <= 0 || maxLabels > MaxLabels {
		maxLabels = MaxLabels
	}

	// Truncate text if too long to save tokens
	runes := []rune(cleanedText)
	if len(runes) > MaxInputCharsForAI {
		cleanedText = string(runes[:MaxInputCharsForAI])
	}

	// Use custom system prompt if provided, otherwise use default
	systemPrompt := a.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = "You are a labeling assistant. Generate concise, relevant labels (keywords or short phrases) for the given text. Output ONLY a JSON array of labels, nothing else. Example: [\"Technology\", \"AI\", \"Machine Learning\"]"
	}

	userPrompt := fmt.Sprintf("Generate %d relevant labels for the following text. Return only a JSON array of strings:\n\n%s", maxLabels, cleanedText)

	requestBody := map[string]interface{}{
		"model": a.Model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"temperature": 0.3, // Low temperature for consistent labels
		"max_tokens":  200, // Limit output tokens for labels
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return LabelResult{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Validate endpoint URL to prevent SSRF attacks
	apiURL := a.Endpoint + "/chat/completions"
	parsedURL, err := url.Parse(apiURL)
	if err != nil {
		return LabelResult{}, fmt.Errorf("invalid API endpoint URL: %w", err)
	}
	// Allow HTTP for localhost/development environments, require HTTPS otherwise
	isLocalhost := parsedURL.Hostname() == "localhost" || parsedURL.Hostname() == "127.0.0.1" || parsedURL.Hostname() == "::1"
	if parsedURL.Scheme != "https" && !isLocalhost {
		return LabelResult{}, fmt.Errorf("API endpoint must use HTTPS for security (unless localhost)")
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return LabelResult{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.APIKey)

	resp, err := a.client.Do(req)
	if err != nil {
		return LabelResult{}, fmt.Errorf("ai api request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResp struct {
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil && errorResp.Error.Message != "" {
			return LabelResult{}, fmt.Errorf("ai api error: %s", errorResp.Error.Message)
		}
		return LabelResult{}, fmt.Errorf("ai api returned status: %d", resp.StatusCode)
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return LabelResult{}, fmt.Errorf("failed to decode ai response: %w", err)
	}

	if len(result.Choices) == 0 || result.Choices[0].Message.Content == "" {
		return LabelResult{}, fmt.Errorf("no labels found in ai response")
	}

	// Parse the JSON array from the response
	content := strings.TrimSpace(result.Choices[0].Message.Content)

	// Try to extract JSON array if wrapped in markdown code blocks
	if strings.Contains(content, "```") {
		// Extract content between ``` markers
		start := strings.Index(content, "[")
		end := strings.LastIndex(content, "]")
		if start != -1 && end != -1 && end > start {
			content = content[start : end+1]
		}
	}

	var labels []string
	if err := json.Unmarshal([]byte(content), &labels); err != nil {
		// If JSON parsing fails, try to extract labels from plain text
		labels = extractLabelsFromText(content, maxLabels)
	}

	// Normalize and validate labels
	var validLabels []string
	for _, label := range labels {
		normalized := normalizeLabel(label)
		if validateLabel(normalized) {
			validLabels = append(validLabels, normalized)
		}
		if len(validLabels) >= maxLabels {
			break
		}
	}

	return LabelResult{
		Labels:     validLabels,
		IsTooShort: false,
	}, nil
}

// extractLabelsFromText attempts to extract labels from plain text response
func extractLabelsFromText(text string, maxLabels int) []string {
	var labels []string

	// Try to split by common delimiters
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Remove common list markers
		line = strings.TrimPrefix(line, "-")
		line = strings.TrimPrefix(line, "*")
		line = strings.TrimPrefix(line, "•")
		line = strings.TrimSpace(line)

		// Remove numbering
		for i := 0; i < 10; i++ {
			prefix := fmt.Sprintf("%d.", i)
			if strings.HasPrefix(line, prefix) {
				line = strings.TrimPrefix(line, prefix)
				line = strings.TrimSpace(line)
				break
			}
		}

		if line != "" && len(line) >= MinLabelLength && len(line) <= MaxLabelLength {
			labels = append(labels, line)
			if len(labels) >= maxLabels {
				break
			}
		}
	}

	// If no labels found from lines, try comma-separated
	if len(labels) == 0 {
		parts := strings.Split(text, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part != "" && len(part) >= MinLabelLength && len(part) <= MaxLabelLength {
				labels = append(labels, part)
				if len(labels) >= maxLabels {
					break
				}
			}
		}
	}

	return labels
}
