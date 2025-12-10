package label

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"MrRSS/internal/feed"
	"MrRSS/internal/handlers/core"
	"MrRSS/internal/label"
	"MrRSS/internal/utils"
)

// HandleGenerateLabels generates labels for an article's content.
func HandleGenerateLabels(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ArticleID int64 `json:"article_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get the article content
	content, err := getArticleContent(h, req.ArticleID)
	if err != nil {
		log.Printf("Error getting article content for labeling: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if content == "" {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"labels":      []string{},
			"is_too_short": true,
			"error":        "No content available for this article",
		})
		return
	}

	// Get label provider from settings (with default)
	provider, err := h.DB.GetSetting("label_provider")
	if err != nil || provider == "" {
		provider = "local" // Default to local algorithm
	}

	// Get max label count from settings
	maxCountStr, _ := h.DB.GetSetting("label_max_count")
	maxCount, err := strconv.Atoi(maxCountStr)
	if err != nil || maxCount <= 0 || maxCount > label.MaxLabels {
		maxCount = label.MaxLabels
	}

	var result label.LabelResult

	if provider == "ai" {
		// Use AI labeling
		apiKey, err := h.DB.GetSetting("label_ai_api_key")
		if err != nil || apiKey == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "missing_ai_api_key",
			})
			return
		}

		// Get endpoint and model with fallback to defaults
		endpoint, _ := h.DB.GetSetting("label_ai_endpoint")
		model, _ := h.DB.GetSetting("label_ai_model")
		systemPrompt, _ := h.DB.GetSetting("label_ai_system_prompt")

		aiLabeler := label.NewAILabeler(apiKey, endpoint, model)
		if systemPrompt != "" {
			aiLabeler.SetSystemPrompt(systemPrompt)
		}
		aiResult, err := aiLabeler.GenerateLabels(content, maxCount)
		if err != nil {
			log.Printf("Error generating AI labels: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		result = aiResult
	} else {
		// Use local algorithm
		labeler := label.NewLabeler()
		result = labeler.GenerateLabels(content, maxCount)
	}

	// Save labels to database
	labelsJSON, err := json.Marshal(result.Labels)
	if err != nil {
		log.Printf("Error marshaling labels: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := h.DB.UpdateArticleLabels(req.ArticleID, string(labelsJSON)); err != nil {
		log.Printf("Error saving labels to database: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"labels":       result.Labels,
		"is_too_short": result.IsTooShort,
	})
}

// HandleUpdateLabels updates the labels for an article
func HandleUpdateLabels(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ArticleID int64    `json:"article_id"`
		Labels    []string `json:"labels"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate labels
	var validLabels []string
	for _, lbl := range req.Labels {
		if len(lbl) >= label.MinLabelLength && len(lbl) <= label.MaxLabelLength {
			validLabels = append(validLabels, lbl)
		}
	}

	// Limit number of labels
	if len(validLabels) > label.MaxLabels {
		validLabels = validLabels[:label.MaxLabels]
	}

	// Save to database
	labelsJSON, err := json.Marshal(validLabels)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := h.DB.UpdateArticleLabels(req.ArticleID, string(labelsJSON)); err != nil {
		log.Printf("Error updating labels: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"labels":  validLabels,
	})
}

// getArticleContent fetches the content of an article by ID
func getArticleContent(h *core.Handler, articleID int64) (string, error) {
	// Get the article directly by ID
	article, err := h.DB.GetArticleByID(articleID)
	if err != nil {
		return "", fmt.Errorf("failed to get article by ID: %w", err)
	}

	// Get the feed
	feeds, err := h.DB.GetFeeds()
	if err != nil {
		return "", err
	}

	var targetFeed *struct {
		URL        string
		ScriptPath string
	}
	for _, f := range feeds {
		if f.ID == article.FeedID {
			targetFeed = &struct {
				URL        string
				ScriptPath string
			}{
				URL:        f.URL,
				ScriptPath: f.ScriptPath,
			}
			break
		}
	}

	if targetFeed == nil {
		return "", nil
	}

	// Parse the feed to get fresh content
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	parsedFeed, err := h.Fetcher.ParseFeedWithScript(ctx, targetFeed.URL, targetFeed.ScriptPath)
	if err != nil {
		return "", err
	}

	// Find the article in the feed by URL
	for _, item := range parsedFeed.Items {
		if utils.URLsMatch(item.Link, article.URL) {
			// Use the centralized content extraction logic
			content := feed.ExtractContent(item)
			return utils.CleanHTML(content), nil
		}
	}

	return "", nil
}
