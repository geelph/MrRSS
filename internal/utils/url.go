package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// NormalizeFeedURL ensures the feed URL has a protocol prefix.
// If the URL doesn't start with http://, https://, rsshub://, or other protocols,
// it will automatically prepend https://
func NormalizeFeedURL(feedURL string) string {
	if feedURL == "" {
		return feedURL
	}

	// Trim whitespace
	feedURL = strings.TrimSpace(feedURL)

	// Check if URL already has a protocol
	protocols := []string{
		"http://",
		"https://",
		"rsshub://",
		"script://",
		"email://",
		"feed://",
		"ftp://",
		"file://",
	}

	for _, protocol := range protocols {
		if strings.HasPrefix(strings.ToLower(feedURL), protocol) {
			return feedURL
		}
	}

	// If no protocol found, prepend https://
	return "https://" + feedURL
}

// NormalizeURLForComparison returns a normalized URL for comparison purposes.
// It strips query parameters that often change between feed fetches (like tracking params).
// This helps match articles even when feeds use dynamic URL parameters.
func NormalizeURLForComparison(rawURL string) string {
	if rawURL == "" {
		return ""
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	// If no scheme, return original (likely invalid URL)
	if parsed.Scheme == "" {
		return rawURL
	}
	// Return just scheme + host + path (without query parameters)
	return parsed.Scheme + "://" + parsed.Host + parsed.Path
}

// URLsMatch checks if two URLs refer to the same article by comparing their normalized forms.
// It first tries exact match, then falls back to intelligent normalization that preserves
// important query parameters while ignoring tracking parameters.
func URLsMatch(url1, url2 string) bool {
	// Try exact match first
	if url1 == url2 {
		return true
	}

	// Fall back to intelligent normalization
	return normalizeURLForMatching(url1) == normalizeURLForMatching(url2)
}

// normalizeURLForMatching normalizes URLs for comparison by preserving important query parameters
// and removing tracking parameters and other non-essential parameters.
func normalizeURLForMatching(rawURL string) string {
	if rawURL == "" {
		return ""
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	// If no scheme, return original (likely invalid URL)
	if parsed.Scheme == "" {
		return rawURL
	}

	result := parsed.Scheme + "://" + parsed.Host + parsed.Path

	// Process query parameters
	query := parsed.Query()
	if len(query) > 0 {
		importantParams := make(url.Values)

		for key, values := range query {
			if isImportantParameter(key, values) {
				importantParams[key] = values
			}
		}

		if len(importantParams) > 0 {
			result += "?" + importantParams.Encode()
		}
	}

	return result
}

// isImportantParameter determines if a query parameter should be preserved for URL matching
func isImportantParameter(key string, values []string) bool {
	if len(values) == 0 {
		return false
	}

	value := values[0] // Use first value

	// Always preserve parameters that look like IDs
	if isIDParameter(key) {
		return true
	}

	// Ignore known tracking parameters
	if isTrackingParameter(key) {
		return false
	}

	// Ignore parameters with very long random-looking values (likely tracking)
	if len(value) > 50 && looksLikeTrackingToken(value) {
		return false
	}

	// Preserve parameters with numeric values (likely IDs) - but not if they look like tracking tokens
	if isNumeric(value) && !looksLikeTrackingToken(value) {
		return true
	}

	// For other parameters, use heuristics
	// Short parameters are more likely to be important
	if len(key) <= 3 && len(value) <= 20 {
		return true
	}

	// Parameters with meaningful names
	if containsMeaningfulWords(key) {
		return true
	}

	// Default: preserve if not obviously tracking
	return !looksLikeTrackingToken(value)
}

// isIDParameter checks if parameter name suggests it's an ID
func isIDParameter(key string) bool {
	keyLower := strings.ToLower(key)

	// Exact matches for common ID parameter names
	exactMatches := []string{"id", "mid", "cid", "uid", "pid", "tid", "aid", "bid", "did", "eid", "fid", "gid", "hid", "iid", "jid", "kid", "lid", "nid", "oid", "qid", "rid", "sid", "vid", "wid", "xid", "yid", "zid"}
	for _, match := range exactMatches {
		if keyLower == match {
			return true
		}
	}

	// Prefix/suffix patterns for ID parameters
	idPatterns := []string{"_id", "id_", "article", "post", "entry", "item", "thread", "topic", "page", "__biz", "idx", "pmid"}
	for _, pattern := range idPatterns {
		if strings.Contains(keyLower, pattern) {
			return true
		}
	}

	return false
}

// isNumeric checks if a string represents a number
func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// isTrackingParameter checks if parameter is a known tracking parameter
func isTrackingParameter(key string) bool {
	keyLower := strings.ToLower(key)
	trackingPrefixes := []string{"utm_", "fbclid", "gclid", "msclkid", "ttclid", "_ga", "_gid", "_gat"}
	exactMatches := []string{"ref", "referrer", "source", "campaign", "medium", "term", "content", "fc", "sn"}

	for _, prefix := range trackingPrefixes {
		if strings.HasPrefix(keyLower, prefix) {
			return true
		}
	}

	for _, match := range exactMatches {
		if keyLower == match {
			return true
		}
	}

	return false
}

// looksLikeTrackingToken checks if a value looks like a tracking token
func looksLikeTrackingToken(value string) bool {
	if len(value) < 10 {
		return false
	}

	// Count different character types
	hasLower := strings.ContainsAny(value, "abcdefghijklmnopqrstuvwxyz")
	hasUpper := strings.ContainsAny(value, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	hasDigit := strings.ContainsAny(value, "0123456789")
	hasSpecial := strings.ContainsAny(value, "-_.")

	// Tracking tokens often have mixed case, digits, and special chars
	charTypeCount := 0
	if hasLower {
		charTypeCount++
	}
	if hasUpper {
		charTypeCount++
	}
	if hasDigit {
		charTypeCount++
	}
	if hasSpecial {
		charTypeCount++
	}

	// High entropy (many different character types) suggests tracking token
	if charTypeCount >= 3 {
		return true
	}

	// Long numeric strings (like timestamps) are also likely tracking
	if charTypeCount == 1 && hasDigit && len(value) > 12 {
		return true
	}

	return false
}

// containsMeaningfulWords checks if parameter name contains meaningful words
func containsMeaningfulWords(key string) bool {
	keyLower := strings.ToLower(key)
	meaningfulWords := []string{"lang", "locale", "format", "type", "category", "tag", "section", "view", "mode"}

	for _, word := range meaningfulWords {
		if strings.Contains(keyLower, word) {
			return true
		}
	}

	return false
}

// GenerateArticleUniqueID generates a unique identifier for an article based on title + feed_id + published_date.
// This provides better deduplication than URL-based approaches, especially when feeds use tracking parameters
// or when the same article appears in multiple feeds with different URLs.
// Note: Uses date only (not full timestamp) to group articles published on the same day.
func GenerateArticleUniqueID(title string, feedID int64, publishedAt time.Time, hasValidPublishedTime bool) string {
	// Trim whitespace from title
	title = strings.TrimSpace(title)

	// Format: title|feed_id|published_date (YYYY-MM-DD format)
	// Using date only instead of full timestamp to allow articles published on the same day
	// with minor time differences to be treated as duplicates
	var dateStr string
	if hasValidPublishedTime {
		dateStr = publishedAt.Format("2006-01-02")
	} else {
		dateStr = "" // Use empty string if no published date
	}

	data := fmt.Sprintf("%s|%d|%s", title, feedID, dateStr)

	// Generate MD5 hash and convert to lowercase hex string
	hash := md5.Sum([]byte(data))
	return strings.ToLower(hex.EncodeToString(hash[:]))
}
