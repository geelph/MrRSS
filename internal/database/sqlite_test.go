package database

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"MrRSS/internal/models"
)

func TestDatabaseInitialization(t *testing.T) {
	// Create temporary database
	dbFile := "test_init.db"
	defer os.Remove(dbFile)

	// Test database creation and initialization
	db, err := NewDB(dbFile)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Initialize database
	err = db.Init()
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Verify tables were created
	var tableCount int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name IN ('feeds', 'articles', 'settings', 'schema_version')").Scan(&tableCount)
	if err != nil {
		t.Fatalf("Failed to query tables: %v", err)
	}
	if tableCount != 4 {
		t.Errorf("Expected 4 tables, got %d", tableCount)
	}

	// Verify indexes were created
	var indexCount int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name LIKE 'idx_%'").Scan(&indexCount)
	if err != nil {
		t.Fatalf("Failed to query indexes: %v", err)
	}
	if indexCount < 8 {
		t.Errorf("Expected at least 8 indexes, got %d", indexCount)
	}

	// Verify schema version
	var version int
	err = db.QueryRow("SELECT MAX(version) FROM schema_version").Scan(&version)
	if err != nil {
		t.Fatalf("Failed to query schema version: %v", err)
	}
	if version < 5 {
		t.Errorf("Expected schema version >= 5, got %d", version)
	}
}

func TestDatabasePerformanceWithIndexes(t *testing.T) {
	// Create temporary database
	dbFile := "test_perf.db"
	defer os.Remove(dbFile)

	db, err := NewDB(dbFile)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	err = db.Init()
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Add test feed
	feed := &models.Feed{
		Title:       "Test Feed",
		URL:         "https://example.com/feed",
		Description: "Test Description",
		Category:    "test",
	}
	err = db.AddFeed(feed)
	if err != nil {
		t.Fatalf("Failed to add feed: %v", err)
	}

	// Get feed ID
	feeds, err := db.GetFeeds()
	if err != nil || len(feeds) == 0 {
		t.Fatalf("Failed to get feeds: %v", err)
	}
	feedID := feeds[0].ID

	// Insert many test articles
	ctx := context.Background()
	numArticles := 1000
	articles := make([]*models.Article, numArticles)
	for i := 0; i < numArticles; i++ {
		articles[i] = &models.Article{
			FeedID:      feedID,
			Title:       fmt.Sprintf("Article %d", i),
			URL:         fmt.Sprintf("https://example.com/article-%d", i),
			PublishedAt: time.Now().Add(-time.Duration(i) * time.Minute),
			IsRead:      i%2 == 0,
			IsFavorite:  i%10 == 0,
		}
	}

	// Measure insert time
	startInsert := time.Now()
	err = db.SaveArticles(ctx, articles)
	if err != nil {
		t.Fatalf("Failed to save articles: %v", err)
	}
	insertDuration := time.Since(startInsert)
	t.Logf("Inserted %d articles in %v", numArticles, insertDuration)

	// Measure query time with filter
	startQuery := time.Now()
	results, err := db.GetArticles("unread", feedID, "", 50, 0)
	if err != nil {
		t.Fatalf("Failed to get articles: %v", err)
	}
	queryDuration := time.Since(startQuery)
	t.Logf("Queried articles in %v, got %d results", queryDuration, len(results))

	// Query should be fast with indexes (under 50ms for 1000 articles)
	if queryDuration > 50*time.Millisecond {
		t.Logf("Warning: Query took longer than expected: %v", queryDuration)
	}

	// Test category query
	startCategoryQuery := time.Now()
	results, err = db.GetArticles("", 0, "test", 50, 0)
	if err != nil {
		t.Fatalf("Failed to get articles by category: %v", err)
	}
	categoryQueryDuration := time.Since(startCategoryQuery)
	t.Logf("Queried articles by category in %v, got %d results", categoryQueryDuration, len(results))

	// Test favorites query
	startFavQuery := time.Now()
	results, err = db.GetArticles("favorites", 0, "", 50, 0)
	if err != nil {
		t.Fatalf("Failed to get favorite articles: %v", err)
	}
	favQueryDuration := time.Since(startFavQuery)
	t.Logf("Queried favorite articles in %v, got %d results", favQueryDuration, len(results))
}

func TestMigrationIdempotency(t *testing.T) {
	// Create temporary database
	dbFile := "test_migration.db"
	defer os.Remove(dbFile)

	db, err := NewDB(dbFile)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Initialize database multiple times
	for i := 0; i < 3; i++ {
		err = db.Init()
		if err != nil {
			t.Fatalf("Failed to initialize database on iteration %d: %v", i, err)
		}
	}

	// Verify schema version is still correct
	var version int
	err = db.QueryRow("SELECT MAX(version) FROM schema_version").Scan(&version)
	if err != nil {
		t.Fatalf("Failed to query schema version: %v", err)
	}
	if version < 5 {
		t.Errorf("Expected schema version >= 5, got %d", version)
	}

	// Verify only one version 5 entry exists
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM schema_version WHERE version = 5").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count schema versions: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 version 5 entry, got %d", count)
	}
}

func BenchmarkGetArticles(b *testing.B) {
	// Create temporary database
	dbFile := "bench.db"
	defer os.Remove(dbFile)

	db, err := NewDB(dbFile)
	if err != nil {
		b.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	err = db.Init()
	if err != nil {
		b.Fatalf("Failed to initialize database: %v", err)
	}

	// Add test feed
	feed := &models.Feed{
		Title:       "Bench Feed",
		URL:         "https://example.com/bench",
		Description: "Bench Description",
	}
	err = db.AddFeed(feed)
	if err != nil {
		b.Fatalf("Failed to add feed: %v", err)
	}

	feeds, _ := db.GetFeeds()
	feedID := feeds[0].ID

	// Insert test articles
	ctx := context.Background()
	articles := make([]*models.Article, 500)
	for i := 0; i < 500; i++ {
		articles[i] = &models.Article{
			FeedID:      feedID,
			Title:       fmt.Sprintf("Bench Article %d", i),
			URL:         fmt.Sprintf("https://example.com/bench-%d", i),
			PublishedAt: time.Now().Add(-time.Duration(i) * time.Minute),
			IsRead:      i%3 == 0,
		}
	}
	db.SaveArticles(ctx, articles)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := db.GetArticles("", feedID, "", 50, 0)
		if err != nil {
			b.Fatalf("Failed to get articles: %v", err)
		}
	}
}
