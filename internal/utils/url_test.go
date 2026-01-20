package utils

import "testing"

func TestNormalizeFeedURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "URL with https://",
			input:    "https://example.com/feed.xml",
			expected: "https://example.com/feed.xml",
		},
		{
			name:     "URL with http://",
			input:    "http://example.com/feed.xml",
			expected: "http://example.com/feed.xml",
		},
		{
			name:     "URL with rsshub://",
			input:    "rsshub://nytimes/sections",
			expected: "rsshub://nytimes/sections",
		},
		{
			name:     "URL with script://",
			input:    "script://custom_feed.js",
			expected: "script://custom_feed.js",
		},
		{
			name:     "URL with email://",
			input:    "email://user@example.com",
			expected: "email://user@example.com",
		},
		{
			name:     "URL without protocol",
			input:    "example.com/feed.xml",
			expected: "https://example.com/feed.xml",
		},
		{
			name:     "URL without protocol - with path",
			input:    "example.com/rss/feed",
			expected: "https://example.com/rss/feed",
		},
		{
			name:     "URL with extra spaces",
			input:    "  example.com/feed.xml  ",
			expected: "https://example.com/feed.xml",
		},
		{
			name:     "URL with uppercase protocol",
			input:    "HTTP://example.com/feed.xml",
			expected: "HTTP://example.com/feed.xml",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "feed:// protocol",
			input:    "feed://example.com/rss",
			expected: "feed://example.com/rss",
		},
		{
			name:     "ftp:// protocol",
			input:    "ftp://example.com/file.xml",
			expected: "ftp://example.com/file.xml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeFeedURL(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeFeedURL(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNormalizeURLForComparison(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "URL without query params",
			input:    "https://pubmed.ncbi.nlm.nih.gov/41345845/",
			expected: "https://pubmed.ncbi.nlm.nih.gov/41345845/",
		},
		{
			name:     "URL with tracking params",
			input:    "https://pubmed.ncbi.nlm.nih.gov/41345845/?utm_source=Edge&utm_medium=rss&fc=20251205031104",
			expected: "https://pubmed.ncbi.nlm.nih.gov/41345845/",
		},
		{
			name:     "URL with different tracking params",
			input:    "https://pubmed.ncbi.nlm.nih.gov/41345845/?fc=20251206000000&ff=20251206000001&v=2.18.0.post22",
			expected: "https://pubmed.ncbi.nlm.nih.gov/41345845/",
		},
		{
			name:     "HTTP URL with params",
			input:    "http://example.com/article/123?tracking=abc",
			expected: "http://example.com/article/123",
		},
		{
			name:     "Invalid URL returns original",
			input:    "not-a-valid-url",
			expected: "not-a-valid-url",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeURLForComparison(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeURLForComparison(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestURLsMatch(t *testing.T) {
	tests := []struct {
		name     string
		url1     string
		url2     string
		expected bool
	}{
		{
			name:     "Identical URLs",
			url1:     "https://pubmed.ncbi.nlm.nih.gov/41345845/",
			url2:     "https://pubmed.ncbi.nlm.nih.gov/41345845/",
			expected: true,
		},
		{
			name:     "Same base URL with different tracking params",
			url1:     "https://pubmed.ncbi.nlm.nih.gov/41345845/?utm_source=Edge&fc=20251205031104",
			url2:     "https://pubmed.ncbi.nlm.nih.gov/41345845/?utm_source=Chrome&fc=20251206000000",
			expected: true,
		},
		{
			name:     "One with params one without",
			url1:     "https://pubmed.ncbi.nlm.nih.gov/41345845/?fc=20251205031104",
			url2:     "https://pubmed.ncbi.nlm.nih.gov/41345845/",
			expected: true,
		},
		{
			name:     "Different article IDs",
			url1:     "https://pubmed.ncbi.nlm.nih.gov/41345845/",
			url2:     "https://pubmed.ncbi.nlm.nih.gov/41345846/",
			expected: false,
		},
		{
			name:     "Different hosts",
			url1:     "https://pubmed.ncbi.nlm.nih.gov/41345845/",
			url2:     "https://example.com/41345845/",
			expected: false,
		},
		{
			name:     "Empty strings",
			url1:     "",
			url2:     "",
			expected: true,
		},
		{
			name:     "WeChat same article different params",
			url1:     "https://mp.weixin.qq.com/s?__biz=MzIyODI1MzYyNA==&mid=2653550587&idx=1&sn=7674e614b87ed7d64644b944c709dc3a",
			url2:     "https://mp.weixin.qq.com/s?__biz=MzIyODI1MzYyNA==&mid=2653550587&idx=1&sn=999999999999999999999999999999",
			expected: true, // Same mid = same article
		},
		{
			name:     "WeChat different articles",
			url1:     "https://mp.weixin.qq.com/s?__biz=MzIyODI1MzYyNA==&mid=2653550587&idx=1&sn=7674e614b87ed7d64644b944c709dc3a",
			url2:     "https://mp.weixin.qq.com/s?__biz=MzIyODI1MzYyNA==&mid=2653550611&idx=1&sn=133a083423e1a3aa01ae86837d46d5c1",
			expected: false, // Different mid = different articles
		},
		{
			name:     "Generic ID parameter preservation",
			url1:     "https://example.com/article?id=123&utm_source=twitter",
			url2:     "https://example.com/article?id=123&utm_source=facebook",
			expected: true, // Same ID, different tracking params
		},
		{
			name:     "Numeric parameter preservation",
			url1:     "https://example.com/post/456?ref=homepage",
			url2:     "https://example.com/post/456?ref=search",
			expected: true, // Numeric path segment treated as ID
		},
		{
			name:     "Tracking parameter filtering",
			url1:     "https://example.com/page?utm_campaign=summer&utm_medium=email",
			url2:     "https://example.com/page?utm_campaign=winter&utm_medium=social",
			expected: true, // All tracking params ignored
		},
		{
			name:     "Mixed important and tracking params",
			url1:     "https://example.com/article?id=789&utm_source=newsletter&fbclid=abc123",
			url2:     "https://example.com/article?id=789&utm_source=podcast&gclid=def456",
			expected: true, // ID preserved, tracking params ignored
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := URLsMatch(tt.url1, tt.url2)
			if result != tt.expected {
				t.Errorf("URLsMatch(%q, %q) = %v, want %v", tt.url1, tt.url2, result, tt.expected)
			}
		})
	}
}
