package feed_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"MrRSS/internal/database"
	ff "MrRSS/internal/feed"
	"MrRSS/internal/handlers/core"
	fh "MrRSS/internal/handlers/feed"
	"MrRSS/internal/models"
)

func setupHandler(t *testing.T) *core.Handler {
	t.Helper()
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB error: %v", err)
	}
	if err := db.Init(); err != nil {
		t.Fatalf("db Init error: %v", err)
	}
	f := ff.NewFetcher(db, nil)
	return core.NewHandler(db, f, nil)
}

func TestHandleAddAndDeleteAndRefreshFeed(t *testing.T) {
	h := setupHandler(t)

	// Create a simple RSS server
	rss := `<?xml version="1.0" encoding="utf-8"?><rss version="2.0"><channel><title>Test Feed</title><link>http://example.com/</link><item><title>Item1</title><link>http://example.com/1</link><guid>1</guid><pubDate>Mon, 02 Jan 2006 15:04:05 MST</pubDate></item></channel></rss>`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		w.Write([]byte(rss))
	}))
	defer srv.Close()

	// Add feed via handler
	payload := map[string]interface{}{"url": srv.URL, "category": "test"}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/feeds", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	fh.HandleAddFeed(h, w, req)
	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK from add feed, got %d", res.StatusCode)
	}

	// Verify feed exists
	feeds, err := h.DB.GetFeeds()
	if err != nil || len(feeds) == 0 {
		t.Fatalf("expected feeds in DB after add, err=%v, count=%d", err, len(feeds))
	}

	// Delete feed
	id := feeds[0].ID
	delReq := httptest.NewRequest(http.MethodPost, "/api/feeds/delete?id="+strconv.FormatInt(id, 10), nil)
	w2 := httptest.NewRecorder()
	fh.HandleDeleteFeed(h, w2, delReq)
	if w2.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK from delete feed, got %d", w2.Result().StatusCode)
	}

	// Try refresh with bad method
	rreq := httptest.NewRequest(http.MethodGet, "/api/feeds/refresh?id=1", nil)
	rw := httptest.NewRecorder()
	fh.HandleRefreshFeed(h, rw, rreq)
	if rw.Result().StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405 for GET refresh, got %d", rw.Result().StatusCode)
	}

	// Create a feed and test refresh path
	feed := &models.Feed{Title: "f", URL: srv.URL}
	feedID, err := h.DB.AddFeed(feed)
	if err != nil {
		t.Fatalf("AddFeed error: %v", err)
	}

	// POST refresh with invalid id
	badReq := httptest.NewRequest(http.MethodPost, "/api/feeds/refresh?id=notint", nil)
	w3 := httptest.NewRecorder()
	fh.HandleRefreshFeed(h, w3, badReq)
	if w3.Result().StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid id, got %d", w3.Result().StatusCode)
	}

	// POST refresh with existing id
	okReq := httptest.NewRequest(http.MethodPost, "/api/feeds/refresh?id="+strconv.FormatInt(feedID, 10), nil)
	w4 := httptest.NewRecorder()
	fh.HandleRefreshFeed(h, w4, okReq)
	if w4.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for refresh, got %d", w4.Result().StatusCode)
	}

	// allow background fetch to run briefly
	time.Sleep(100 * time.Millisecond)
}
