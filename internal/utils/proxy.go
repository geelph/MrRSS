package utils

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// BuildProxyURL constructs a proxy URL from settings
func BuildProxyURL(proxyType, proxyHost, proxyPort, username, password string) string {
	if proxyHost == "" || proxyPort == "" {
		return ""
	}

	// Build auth string if username is provided
	auth := ""
	if username != "" {
		if password != "" {
			auth = username + ":" + password + "@"
		} else {
			auth = username + "@"
		}
	}

	return fmt.Sprintf("%s://%s%s:%s", proxyType, auth, proxyHost, proxyPort)
}

// CreateHTTPClient creates an HTTP client with optional proxy support
// This is the canonical implementation with proper TLS config and connection pooling
func CreateHTTPClient(proxyURL string, timeout time.Duration) (*http.Client, error) {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}

	// Configure proxy if provided
	if proxyURL != "" {
		parsedProxy, err := url.Parse(proxyURL)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy URL: %w", err)
		}
		transport.Proxy = http.ProxyURL(parsedProxy)
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}

	return client, nil
}
