package opml

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"MrRSS/internal/database"
	"MrRSS/internal/feed"
	corepkg "MrRSS/internal/handlers/core"
)

func TestHandleOPMLImport_RawBody(t *testing.T) {
	xmlData := `<?xml version="1.0"?>
<opml version="1.0">
  <head><title>Test</title></head>
  <body>
    <outline text="Tech" title="Tech">
      <outline type="rss" text="Hacker News" title="Hacker News" xmlUrl="https://news.ycombinator.com/rss" />
    </outline>
    <outline type="rss" text="Go Blog" title="Go Blog" xmlUrl="https://blog.golang.org/feed.atom" />
  </body>
</opml>`

	// Use a real fetcher that writes to an in-memory DB (ImportSubscription uses DB.AddFeed)
	db := func() *database.DB {
		db, err := database.NewDB(":memory:")
		if err != nil {
			t.Fatalf("failed to create db: %v", err)
		}
		if err := db.Init(); err != nil {
			t.Fatalf("failed to init db: %v", err)
		}
		return db
	}()

	f := feed.NewFetcher(db, nil)
	h := &corepkg.Handler{Fetcher: f}

	req := httptest.NewRequest(http.MethodPost, "/opml/import", strings.NewReader(xmlData))
	rr := httptest.NewRecorder()

	HandleOPMLImport(h, rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rr.Code)
	}

	// Verify feeds were added
	feeds, err := db.GetFeeds()
	if err != nil {
		t.Fatalf("GetFeeds failed: %v", err)
	}
	if len(feeds) != 2 {
		t.Fatalf("expected 2 feeds in DB, got %d", len(feeds))
	}
}

func TestHandleOPMLExport(t *testing.T) {
	db := func() *database.DB {
		db, err := database.NewDB(":memory:")
		if err != nil {
			t.Fatalf("failed to create db: %v", err)
		}
		if err := db.Init(); err != nil {
			t.Fatalf("failed to init db: %v", err)
		}
		return db
	}()

	// insert a feed via SQL to keep test simple (provide non-null description and last_updated)
	_, _ = db.Exec("INSERT INTO feeds (title, url, description, last_updated) VALUES (?, ?, ?, datetime('now'))", "F1", "http://f1", "")
	// Sanity-check DB: try GetFeeds before calling handler
	if feeds, err := db.GetFeeds(); err != nil {
		t.Fatalf("GetFeeds before handler failed: %v", err)
	} else if len(feeds) == 0 {
		// continue — handler should still return data
	}

	h := &corepkg.Handler{DB: db}

	req := httptest.NewRequest(http.MethodGet, "/opml/export", nil)
	rr := httptest.NewRecorder()

	HandleOPMLExport(h, rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rr.Code)
	}

	ct := rr.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/xml") {
		t.Fatalf("expected text/xml content type, got %s", ct)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "http://f1") {
		t.Fatalf("exported OPML missing feed URL: %s", body)
	}
}
