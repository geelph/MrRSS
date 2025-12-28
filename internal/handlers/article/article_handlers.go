package article

import (
	"encoding/json"
	"net/http"
	"time"

	"MrRSS/internal/handlers/core"
	"MrRSS/internal/models"
)

// HandleProgress returns the current fetch progress with statistics.
func HandleProgress(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	progress := h.Fetcher.GetProgressWithStats()
	json.NewEncoder(w).Encode(progress)
}

// TaskDetailsResponse contains detailed task information
type TaskDetailsResponse struct {
	PoolTasks  []PoolTaskInfo  `json:"pool_tasks"`
	QueueTasks []QueueTaskInfo `json:"queue_tasks"`
}

// PoolTaskInfo contains information about a task in the pool
type PoolTaskInfo struct {
	FeedID    int64  `json:"feed_id"`
	FeedTitle string `json:"feed_title"`
	Reason    int    `json:"reason"`
	CreatedAt string `json:"created_at"`
}

// QueueTaskInfo contains information about a task in the queue
type QueueTaskInfo struct {
	FeedID    int64  `json:"feed_id"`
	FeedTitle string `json:"feed_title"`
	Position  int    `json:"position"`
}

// HandleTaskDetails returns detailed information about tasks in pool and queue
func HandleTaskDetails(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	tm := h.Fetcher.GetTaskManager()

	// Get pool tasks
	poolTasksRaw := tm.GetPoolTasks()
	poolTasks := make([]PoolTaskInfo, len(poolTasksRaw))
	for i, task := range poolTasksRaw {
		poolTasks[i] = PoolTaskInfo{
			FeedID:    task.FeedID,
			FeedTitle: task.FeedTitle,
			Reason:    int(task.Reason),
			CreatedAt: task.CreatedAt.Format(time.RFC3339),
		}
	}

	// Get queue tasks (limit to first 3)
	queueTasksRaw := tm.GetQueueTasks(3)
	queueTasks := make([]QueueTaskInfo, len(queueTasksRaw))
	for i, task := range queueTasksRaw {
		queueTasks[i] = QueueTaskInfo{
			FeedID:    task.FeedID,
			FeedTitle: task.FeedTitle,
			Position:  task.Position,
		}
	}

	response := TaskDetailsResponse{
		PoolTasks:  poolTasks,
		QueueTasks: queueTasks,
	}

	json.NewEncoder(w).Encode(response)
}

// HandleFilteredArticles returns articles filtered by advanced conditions from the database.
func HandleFilteredArticles(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req FilterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set default pagination values
	page := req.Page
	if page < 1 {
		page = 1
	}
	limit := req.Limit
	if limit < 1 {
		limit = 50
	}

	// Get show_hidden_articles setting
	showHiddenStr, _ := h.DB.GetSetting("show_hidden_articles")
	showHidden := showHiddenStr == "true"

	// Get all articles from database
	// Note: Using a high limit to fetch all articles for filtering
	// For very large datasets, consider implementing database-level filtering
	articles, err := h.DB.GetArticles("", 0, "", showHidden, 50000, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get feeds for category lookup
	feeds, err := h.DB.GetFeeds()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create a map of feed ID to category
	feedCategories := make(map[int64]string)
	for _, feed := range feeds {
		feedCategories[feed.ID] = feed.Category
	}

	// Apply filter conditions
	if len(req.Conditions) > 0 {
		var filteredArticles []models.Article
		for _, article := range articles {
			if evaluateArticleConditions(article, req.Conditions, feedCategories) {
				filteredArticles = append(filteredArticles, article)
			}
		}
		articles = filteredArticles
	}

	// Apply pagination
	total := len(articles)
	offset := (page - 1) * limit
	end := offset + limit

	// Handle edge cases for pagination
	var paginatedArticles []models.Article
	if offset >= total {
		// No more articles to show
		paginatedArticles = []models.Article{}
	} else {
		if end > total {
			end = total
		}
		paginatedArticles = articles[offset:end]
	}

	hasMore := end < total

	response := FilterResponse{
		Articles: paginatedArticles,
		Total:    total,
		Page:     page,
		Limit:    limit,
		HasMore:  hasMore,
	}

	json.NewEncoder(w).Encode(response)
}
