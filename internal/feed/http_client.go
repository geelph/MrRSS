package feed

import (
	"MrRSS/internal/utils"
	"net/http"
	"time"
)

// CreateHTTPClient creates an HTTP client with optional proxy support
// Wrapper around utils.CreateHTTPClient with default timeout for feed fetching
func CreateHTTPClient(proxyURL string) (*http.Client, error) {
	return utils.CreateHTTPClient(proxyURL, 30*time.Second)
}

// BuildProxyURL constructs a proxy URL from settings
// Wrapper around utils.BuildProxyURL for backward compatibility
func BuildProxyURL(proxyType, proxyHost, proxyPort, username, password string) string {
	return utils.BuildProxyURL(proxyType, proxyHost, proxyPort, username, password)
}
