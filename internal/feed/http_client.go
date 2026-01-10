package feed

import (
	"MrRSS/internal/utils"
)

// BuildProxyURL constructs a proxy URL from settings
// Wrapper around utils.BuildProxyURL for backward compatibility
func BuildProxyURL(proxyType, proxyHost, proxyPort, username, password string) string {
	return utils.BuildProxyURL(proxyType, proxyHost, proxyPort, username, password)
}
