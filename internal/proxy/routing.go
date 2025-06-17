package proxy

import (
	"log/slog"
	"net/url"
	"strings"
	"sync"
)

// Global ad domains state
var (
	adDomainsMap   map[string]bool
	adDomainsMutex sync.RWMutex

	// Static file extensions map for O(1) lookup
	staticExtMap   map[string]bool
	staticExtMutex sync.RWMutex
)

// RoutingConfig contains configuration for routing decisions
type RoutingConfig struct {
	DirectExtensions []string
	DirectDomains    []string
	AdBlocking       struct {
		Enabled bool
	}
}

// SetAdDomainsMap updates the global ad domains map
func SetAdDomainsMap(adMap map[string]bool) {
	adDomainsMutex.Lock()
	adDomainsMap = adMap
	adDomainsMutex.Unlock()
}

// InitStaticExtensions initializes the static extensions map for O(1) lookup
func InitStaticExtensions(extensions []string) {
	staticExtMutex.Lock()
	defer staticExtMutex.Unlock()

	staticExtMap = make(map[string]bool, len(extensions))
	for _, ext := range extensions {
		// Store extensions in lowercase for case-insensitive matching
		staticExtMap[strings.ToLower(ext)] = true
	}
}

// IsStaticFile checks if URL is a static file
func IsStaticFile(urlStr string, config *RoutingConfig, logger *slog.Logger) bool {
	// Parse URL to get path
	u, err := url.Parse(urlStr)
	if err != nil {
		logger.Debug("Failed to parse URL for static file check", "url", urlStr, "error", err)
		return false
	}

	// Get path (already excludes query string and fragment)
	path := u.Path
	lowerPath := strings.ToLower(path)

	// First try O(1) map lookup if initialized
	staticExtMutex.RLock()
	if len(staticExtMap) > 0 {
		// Extract file extension
		lastDot := strings.LastIndex(lowerPath, ".")
		if lastDot != -1 && lastDot < len(lowerPath)-1 {
			ext := lowerPath[lastDot:]
			if staticExtMap[ext] {
				staticExtMutex.RUnlock()
				logger.Debug("URL identified as static file (map lookup)",
					"url", urlStr,
					"path", path,
					"extension", ext,
					"action", "direct_connection")
				return true
			}
		}
	}
	staticExtMutex.RUnlock()

	// Fallback to linear search if map not initialized
	if config != nil {
		for _, ext := range config.DirectExtensions {
			if strings.HasSuffix(lowerPath, ext) {
				logger.Debug("URL identified as static file (linear search)",
					"url", urlStr,
					"path", path,
					"extension", ext,
					"action", "direct_connection")
				return true
			}
		}
	}

	logger.Debug("URL not a static file", "url", urlStr, "path", path)
	return false
}

// IsCDNDomain checks if domain is a CDN domain
func IsCDNDomain(host string, config *RoutingConfig, logger *slog.Logger) bool {
	lowerHost := strings.ToLower(host)

	if config != nil {
		for _, cdn := range config.DirectDomains {
			if strings.Contains(lowerHost, cdn) {
				logger.Debug("Domain identified as CDN",
					"host", host,
					"pattern", cdn,
					"action", "direct_connection")
				return true
			}
		}
	}

	logger.Debug("Domain not a CDN", "host", host)
	return false
}

// IsAdDomain checks if domain is in ad blocking list (optimized with map)
func IsAdDomain(host string, config *RoutingConfig, logger *slog.Logger) bool {
	if config == nil || !config.AdBlocking.Enabled || adDomainsMap == nil {
		return false
	}

	lowerHost := strings.ToLower(host)

	// Use read lock for concurrent access
	adDomainsMutex.RLock()
	defer adDomainsMutex.RUnlock()

	// Check exact match first
	if adDomainsMap[lowerHost] {
		logger.Debug("Domain blocked (exact match)",
			"host", host,
			"action", "blocked",
			"reason", "ad_domain")
		return true
	}

	// Check if any parent domain is blocked
	parts := strings.Split(lowerHost, ".")
	for i := range parts {
		domain := strings.Join(parts[i:], ".")
		if adDomainsMap[domain] {
			logger.Debug("Domain blocked (parent match)",
				"host", host,
				"blocked_parent", domain,
				"action", "blocked",
				"reason", "ad_domain")
			return true
		}
	}

	logger.Debug("Domain not in ad list", "host", host)
	return false
}
