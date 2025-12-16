package feed

import (
	"MrRSS/internal/database"
	"MrRSS/internal/translation"
	"context"
	"testing"
	"time"

	"github.com/mmcdole/gofeed"
)

func TestFetchSingleFeedProgressTracking(t *testing.T) {
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create db: %v", err)
	}
	if err := db.Init(); err != nil {
		t.Fatalf("Failed to init db: %v", err)
	}

	fetcher := NewFetcher(db, translation.NewMockTranslator())

	mockFeed := &gofeed.Feed{
		Title:       "Test Feed",
		Description: "Test Description",
		Items:       []*gofeed.Item{},
	}
	fetcher.fp = &MockParser{Feed: mockFeed}

	feedID1, _ := fetcher.AddSubscription("http://test1.com/rss", "Category 1", "")
	feedID2, _ := fetcher.AddSubscription("http://test2.com/rss", "Category 2", "")
	feedID3, _ := fetcher.AddSubscription("http://test3.com/rss", "Category 3", "")

	feed1, _ := db.GetFeedByID(feedID1)
	feed2, _ := db.GetFeedByID(feedID2)
	feed3, _ := db.GetFeedByID(feedID3)

	go fetcher.FetchSingleFeed(context.Background(), *feed1)
	go fetcher.FetchSingleFeed(context.Background(), *feed2)
	go fetcher.FetchSingleFeed(context.Background(), *feed3)

	time.Sleep(50 * time.Millisecond)

	progress := fetcher.GetProgress()
	if progress.Total < 1 || progress.Total > 3 {
		t.Errorf("Expected total between 1-3, got %d", progress.Total)
	}

	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			t.Fatal("Timeout waiting for feeds to complete")
		case <-ticker.C:
			progress := fetcher.GetProgress()
			if !progress.IsRunning {
				if progress.Current < 1 || progress.Current > 3 {
					t.Errorf("Expected current between 1 and 3, got %d", progress.Current)
				}
				return
			}
		}
	}
}

func TestFetchSingleFeedDuplicatePrevention(t *testing.T) {
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create db: %v", err)
	}
	if err := db.Init(); err != nil {
		t.Fatalf("Failed to init db: %v", err)
	}

	fetcher := NewFetcher(db, translation.NewMockTranslator())

	mockFeed := &gofeed.Feed{
		Title:       "Test Feed",
		Description: "Test Description",
		Items:       []*gofeed.Item{},
	}
	fetcher.fp = &MockParser{Feed: mockFeed}

	feedID, _ := fetcher.AddSubscription("http://test.com/rss", "Category", "")
	feed, _ := db.GetFeedByID(feedID)

	go fetcher.FetchSingleFeed(context.Background(), *feed)
	go fetcher.FetchSingleFeed(context.Background(), *feed)

	time.Sleep(100 * time.Millisecond)

	progress := fetcher.GetProgress()
	if progress.Total > 1 {
		t.Errorf("Expected total 1 (duplicate prevented), got %d", progress.Total)
	}

	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			t.Fatal("Timeout waiting for feed to complete")
		case <-ticker.C:
			progress := fetcher.GetProgress()
			if !progress.IsRunning {
				return
			}
		}
	}
}
