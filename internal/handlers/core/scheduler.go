package core

import (
	"MrRSS/internal/models"
	"context"
	"log"
	"strconv"
	"time"

	"MrRSS/internal/cache"
	"MrRSS/internal/utils"
)

// StartBackgroundScheduler starts the background scheduler for auto-updates and cleanup.
func (h *Handler) StartBackgroundScheduler(ctx context.Context) {
	// Trigger initial cleanup on startup
	go func() {
		log.Println("Triggering initial cleanup on startup")
		h.Fetcher.GetCleanupManager().RequestCleanup()
	}()

	// Run initial cleanup only if auto_cleanup is enabled
	// This is now handled by the cleanup manager
	go func() {
		autoCleanup, _ := h.DB.GetSetting("auto_cleanup_enabled")
		if autoCleanup == "true" {
			log.Println("Auto cleanup enabled, will run after tasks complete")
		}

		// Run initial media cache cleanup if enabled
		mediaCacheEnabled, _ := h.DB.GetSetting("media_cache_enabled")
		if mediaCacheEnabled == "true" {
			log.Println("Running initial media cache cleanup...")
			h.cleanupMediaCache()
		}
	}()

	// Check refresh mode
	refreshMode, _ := h.DB.GetSetting("refresh_mode")

	if refreshMode == "intelligent" {
		// Use intelligent refresh mode with per-feed intervals
		h.startIntelligentScheduler(ctx)
	} else {
		// Use fixed interval mode (default)
		h.startFixedScheduler(ctx)
	}
}

// startFixedScheduler uses a fixed interval for all feeds (but respects per-feed custom intervals)
func (h *Handler) startFixedScheduler(ctx context.Context) {
	// Use a ticker to check feeds every minute
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	// Get global interval
	getGlobalInterval := func() time.Duration {
		intervalStr, _ := h.DB.GetSetting("update_interval")
		interval := 30
		if i, err := strconv.Atoi(intervalStr); err == nil && i > 0 {
			interval = i
		}
		return time.Duration(interval) * time.Minute
	}

	log.Printf("Starting fixed interval scheduler (global interval: %v)", getGlobalInterval())

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping fixed interval scheduler")
			return
		case <-ticker.C:
			// Check each feed to see if it needs refresh
			go h.refreshFeedsWithFixedInterval(ctx, getGlobalInterval())
		}
	}
}

// refreshFeedsWithFixedInterval checks and refreshes feeds based on fixed intervals
func (h *Handler) refreshFeedsWithFixedInterval(ctx context.Context, globalInterval time.Duration) {
	feeds, err := h.DB.GetFeeds()
	if err != nil {
		log.Printf("Error getting feeds for fixed refresh: %v", err)
		return
	}

	for i, feed := range feeds {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Create local copy to avoid loop variable capture issues
		currentFeed := feed

		// Determine refresh interval for this feed
		// Feed-level settings override global settings:
		// - RefreshInterval > 0: Use custom fixed interval (in minutes)
		// - RefreshInterval == -1: Use intelligent interval (even in fixed mode)
		// - RefreshInterval == 0: Use global interval (follows global mode)
		var refreshInterval time.Duration
		if currentFeed.RefreshInterval > 0 {
			// Use per-feed custom fixed interval
			refreshInterval = time.Duration(currentFeed.RefreshInterval) * time.Minute
		} else if currentFeed.RefreshInterval == -1 {
			// Feed explicitly requests intelligent interval
			// This allows individual feeds to use intelligent refresh even when global mode is fixed
			calculator := h.Fetcher.GetIntelligentRefreshCalculator()
			refreshInterval = calculator.CalculateInterval(currentFeed)
		} else {
			// Use global fixed interval (RefreshInterval == 0)
			refreshInterval = globalInterval
		}

		// Check if feed needs refresh based on last_updated time
		timeSinceUpdate := time.Since(currentFeed.LastUpdated)
		if timeSinceUpdate >= refreshInterval {
			// Apply staggered delay to avoid thundering herd
			staggerDelay := h.Fetcher.GetStaggeredDelay(currentFeed.ID, len(feeds))

			// Schedule feed refresh with stagger
			// If custom interval (>0 or -1), add to queue tail individually
			// If global interval (0), it will be handled by FetchAll instead
			if currentFeed.RefreshInterval != 0 {
				go func(f models.Feed, delay time.Duration, interval time.Duration) {
					time.Sleep(delay)
					select {
					case <-ctx.Done():
						return
					default:
						log.Printf("Auto-refreshing feed %s (interval: %v, fixed mode)", f.Title, interval)
						// Scheduled refresh = add to queue tail
						h.Fetcher.FetchSingleFeed(ctx, f, false)
					}

					// Run media cache cleanup if enabled
					mediaCacheEnabled, _ := h.DB.GetSetting("media_cache_enabled")
					if mediaCacheEnabled == "true" {
						h.cleanupMediaCache()
					}
				}(currentFeed, staggerDelay, refreshInterval)
			}
		}

		// Small delay between checking feeds to avoid CPU spikes
		if i < len(feeds)-1 {
			time.Sleep(100 * time.Millisecond)
		}
	}

	// For feeds with RefreshInterval == 0 (use global setting), trigger FetchAll once
	// This will add them all to queue tail together
	needsGlobalRefresh := false
	for _, feed := range feeds {
		if feed.RefreshInterval == 0 {
			needsGlobalRefresh = true
			break
		}
	}

	if needsGlobalRefresh {
		// Filter feeds that use global interval
		globalFeeds := make([]models.Feed, 0)
		for _, feed := range feeds {
			if feed.RefreshInterval == 0 {
				globalFeeds = append(globalFeeds, feed)
			}
		}
		if len(globalFeeds) > 0 {
			log.Printf("Triggering global refresh for %d feeds using global interval", len(globalFeeds))
			h.Fetcher.FetchAll(ctx)
		}
	}
}

// startIntelligentScheduler uses per-feed intervals with staggered refresh
func (h *Handler) startIntelligentScheduler(ctx context.Context) {
	log.Println("Starting intelligent refresh scheduler")

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping intelligent scheduler")
			return
		case <-ticker.C:
			// Check each feed to see if it needs refresh
			go h.refreshFeedsIntelligently(ctx)
		}
	}
}

// refreshFeedsIntelligently checks and refreshes feeds based on their individual intervals
func (h *Handler) refreshFeedsIntelligently(ctx context.Context) {
	feeds, err := h.DB.GetFeeds()
	if err != nil {
		log.Printf("Error getting feeds for intelligent refresh: %v", err)
		return
	}

	// Use intelligent refresh calculator
	calculator := h.Fetcher.GetIntelligentRefreshCalculator()

	for i, feed := range feeds {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Create local copy to avoid loop variable capture issues
		currentFeed := feed

		// Determine refresh interval for this feed
		// Feed-level settings override global settings:
		// - RefreshInterval > 0: Use custom fixed interval (in minutes)
		// - RefreshInterval == -1: Use intelligent interval (explicit request)
		// - RefreshInterval == 0: Use global interval (intelligent in this mode)
		var refreshInterval time.Duration
		if currentFeed.RefreshInterval > 0 {
			// Use per-feed custom fixed interval
			refreshInterval = time.Duration(currentFeed.RefreshInterval) * time.Minute
		} else if currentFeed.RefreshInterval == -1 {
			// Feed explicitly requests intelligent interval
			refreshInterval = calculator.CalculateInterval(currentFeed)
		} else {
			// Use global intelligent interval (RefreshInterval == 0)
			// In intelligent mode, 0 means "use intelligent calculation"
			refreshInterval = calculator.CalculateInterval(currentFeed)
		}

		// Check if feed needs refresh based on last_updated time
		timeSinceUpdate := time.Since(currentFeed.LastUpdated)
		if timeSinceUpdate >= refreshInterval {
			// Apply staggered delay to avoid thundering herd
			staggerDelay := h.Fetcher.GetStaggeredDelay(currentFeed.ID, len(feeds))

			// Schedule feed refresh with stagger
			// All feeds go to queue tail in intelligent mode
			go func(f models.Feed, delay time.Duration, interval time.Duration) {
				time.Sleep(delay)
				select {
				case <-ctx.Done():
					return
				default:
					log.Printf("Intelligently refreshing feed %s (interval: %v)", f.Title, interval)
					// Scheduled refresh = add to queue tail
					h.Fetcher.FetchSingleFeed(ctx, f, false)
				}
			}(currentFeed, staggerDelay, refreshInterval)
		}

		// Small delay between checking feeds to avoid CPU spikes
		if i < len(feeds)-1 {
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// runCleanup runs the cleanup routine if enabled
func (h *Handler) runCleanup() {
	autoCleanup, _ := h.DB.GetSetting("auto_cleanup_enabled")
	if autoCleanup == "true" {
		count, err := h.DB.CleanupOldArticles()
		if err != nil {
			log.Printf("Error during automatic cleanup: %v", err)
		} else if count > 0 {
			log.Printf("Automatic cleanup: removed %d old articles", count)
		}
	}
}

// cleanupMediaCache performs media cache cleanup based on settings
func (h *Handler) cleanupMediaCache() {
	cacheDir, err := utils.GetMediaCacheDir()
	if err != nil {
		log.Printf("Failed to get media cache directory: %v", err)
		return
	}

	mediaCache, err := cache.NewMediaCache(cacheDir)
	if err != nil {
		log.Printf("Failed to initialize media cache: %v", err)
		return
	}

	// Get settings
	maxAgeDaysStr, _ := h.DB.GetSetting("media_cache_max_age_days")
	maxSizeMBStr, _ := h.DB.GetSetting("media_cache_max_size_mb")

	maxAgeDays, err := strconv.Atoi(maxAgeDaysStr)
	if err != nil || maxAgeDays <= 0 {
		maxAgeDays = 7 // Default
	}

	maxSizeMB, err := strconv.Atoi(maxSizeMBStr)
	if err != nil || maxSizeMB <= 0 {
		maxSizeMB = 100 // Default
	}

	// Cleanup by age
	ageCount, err := mediaCache.CleanupOldFiles(maxAgeDays)
	if err != nil {
		log.Printf("Failed to cleanup old media files: %v", err)
	} else if ageCount > 0 {
		log.Printf("Media cache cleanup: removed %d old files", ageCount)
	}

	// Cleanup by size
	sizeCount, err := mediaCache.CleanupBySize(maxSizeMB)
	if err != nil {
		log.Printf("Failed to cleanup media files by size: %v", err)
	} else if sizeCount > 0 {
		log.Printf("Media cache cleanup: removed %d files to stay under size limit", sizeCount)
	}
}
