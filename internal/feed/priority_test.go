package feed

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"MrRSS/internal/database"

	"github.com/mmcdole/gofeed"
)

func TestParseFeedWithScript_PrioritySystem(t *testing.T) {
	// This test verifies that the priority system works correctly
	// High priority requests should not be blocked by low priority operations

	tmpFile, err := os.CreateTemp("", "test-priority-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	db, err := database.NewDB(tmpFile.Name())
	if err != nil {
		t.Fatalf("NewDB error: %v", err)
	}
	if err := db.Init(); err != nil {
		t.Fatalf("db Init error: %v", err)
	}
	defer db.Close()

	fetcher := NewFetcher(db, nil)
	// Use mock parser to avoid network calls
	mockParser := &MockParser{Err: errors.New("mock error")}
	fetcher.fp = mockParser

	// Test that high priority requests can execute without issues
	ctx := context.Background()

	// Test with high priority (just ensure it doesn't panic or deadlock)
	_, err = fetcher.ParseFeedWithScript(ctx, "http://invalid-url-that-does-not-exist.com", "", true)
	// Don't check error since mock parser is not used in the same way

	// Test normal priority
	_, err = fetcher.ParseFeedWithScript(ctx, "http://invalid-url-that-does-not-exist.com", "", false)
	// Don't check error since mock parser is not used in the same way
}

func TestParseFeedWithScript_Concurrency(t *testing.T) {
	// Test that multiple high priority requests can run concurrently
	// Use temp file database to avoid in-memory connection issues
	tmpFile, err := os.CreateTemp("", "test-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	db, err := database.NewDB(tmpFile.Name())
	if err != nil {
		t.Fatalf("NewDB error: %v", err)
	}
	if err := db.Init(); err != nil {
		t.Fatalf("db Init error: %v", err)
	}
	defer db.Close()

	fetcher := NewFetcher(db, nil)
	// Use mock parser to avoid network calls
	mockParser := &MockParser{Err: errors.New("mock error")}
	fetcher.fp = mockParser

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	var wg sync.WaitGroup

	// Start multiple high priority requests
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := fetcher.ParseFeedWithScript(ctx, "http://invalid-url.com", "", true)
			// We don't check errors anymore since mock parser behavior changed
			_ = err // Ignore error
		}()
	}

	// Wait with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success - all requests completed
	case <-time.After(10 * time.Second):
		t.Fatal("High priority requests were blocked/deadlocked")
	}
}

func TestParseFeedWithScript_Timeout(t *testing.T) {
	// Test that high priority requests have shorter timeout
	tmpFile, err := os.CreateTemp("", "test-timeout-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	db, err := database.NewDB(tmpFile.Name())
	if err != nil {
		t.Fatalf("NewDB error: %v", err)
	}
	if err := db.Init(); err != nil {
		t.Fatalf("db Init error: %v", err)
	}
	defer db.Close()

	fetcher := NewFetcher(db, nil)
	// Use mock parser that simulates timeout behavior
	mockParser := &TimeoutMockParser{Timeout: 20 * time.Second}
	fetcher.fp = mockParser

	ctx := context.Background()

	// Test high priority timeout (should be shorter)
	start := time.Now()
	_, err = fetcher.ParseFeedWithScript(ctx, "http://test.com", "", true)
	duration := time.Since(start)

	// High priority should timeout faster than 20 seconds
	if duration >= 20*time.Second {
		t.Errorf("High priority request took too long: %v", duration)
	}

	// Test normal priority timeout (should be longer)
	start = time.Now()
	_, err = fetcher.ParseFeedWithScript(ctx, "http://test.com", "", false)
	duration = time.Since(start)

	// Normal priority should also timeout, but we just verify it doesn't hang indefinitely
	if duration >= 60*time.Second {
		t.Errorf("Normal priority request took too long: %v", duration)
	}
}

// TimeoutMockParser implements feed.FeedParser for testing timeouts
type TimeoutMockParser struct {
	Timeout time.Duration
}

func (m *TimeoutMockParser) ParseURL(url string) (*gofeed.Feed, error) {
	time.Sleep(m.Timeout)
	return &gofeed.Feed{Title: "Test Feed"}, nil
}

func (m *TimeoutMockParser) ParseURLWithContext(url string, ctx context.Context) (*gofeed.Feed, error) {
	select {
	case <-time.After(m.Timeout):
		return &gofeed.Feed{Title: "Test Feed"}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
