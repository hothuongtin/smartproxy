package proxy

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/proxy"
)

// Constants for transport management
const (
	DefaultTimeout = 30 * time.Second
	
	// Buffer sizes
	SmallBufferSize  = 16 * 1024  // 16KB for small requests
	MediumBufferSize = 64 * 1024  // 64KB for normal requests
	LargeBufferSize  = 256 * 1024 // 256KB for large/media requests
)

// Chrome/Chromium user agent patterns
var chromeUserAgentPatterns = []string{
	"Chrome/",
	"Chromium/",
	"CriOS/", // Chrome on iOS
	"Edg/",   // Edge (Chromium-based)
}

// IsChromeBrowser detects if the request is from Chrome/Chromium
func IsChromeBrowser(userAgent string) bool {
	lowerUA := strings.ToLower(userAgent)
	for _, pattern := range chromeUserAgentPatterns {
		if strings.Contains(lowerUA, strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}

// DetermineBufferSize determines optimal buffer size based on request
func DetermineBufferSize(r *http.Request, isStatic bool, logger *slog.Logger) (readBuffer, writeBuffer int) {
	// Default to medium buffers
	readBuffer = MediumBufferSize
	writeBuffer = MediumBufferSize
	
	// Check if it's a media file based on URL extension
	urlPath := strings.ToLower(r.URL.Path)
	isMedia := false
	mediaExtensions := []string{".mp4", ".webm", ".mp3", ".wav", ".ogg", ".avi", ".mov", ".mkv"}
	for _, ext := range mediaExtensions {
		if strings.HasSuffix(urlPath, ext) {
			isMedia = true
			break
		}
	}
	
	// Check Content-Type header for media
	contentType := r.Header.Get("Accept")
	if strings.Contains(contentType, "video/") || strings.Contains(contentType, "audio/") {
		isMedia = true
	}
	
	// Adjust buffer sizes based on request type
	if isMedia {
		// Large buffers for media streaming
		readBuffer = LargeBufferSize
		writeBuffer = LargeBufferSize
		logger.Debug("Using large buffers for media content",
			"url", r.URL.Path,
			"read_buffer", readBuffer,
			"write_buffer", writeBuffer)
	} else if isStatic {
		// Medium buffers for static files
		readBuffer = MediumBufferSize
		writeBuffer = MediumBufferSize
		logger.Debug("Using medium buffers for static content",
			"url", r.URL.Path,
			"read_buffer", readBuffer,
			"write_buffer", writeBuffer)
	} else if r.Method == "POST" || r.Method == "PUT" {
		// Larger write buffer for uploads
		writeBuffer = LargeBufferSize
		logger.Debug("Using large write buffer for upload",
			"method", r.Method,
			"url", r.URL.Path,
			"write_buffer", writeBuffer)
	} else if strings.Contains(urlPath, "api/") || strings.Contains(urlPath, "/api") {
		// Small buffers for API calls
		readBuffer = SmallBufferSize
		writeBuffer = SmallBufferSize
		logger.Debug("Using small buffers for API request",
			"url", r.URL.Path,
			"read_buffer", readBuffer,
			"write_buffer", writeBuffer)
	}
	
	// Chrome gets slightly larger buffers
	if IsChromeBrowser(r.Header.Get("User-Agent")) {
		if readBuffer == SmallBufferSize {
			readBuffer = MediumBufferSize
		}
		if writeBuffer == SmallBufferSize {
			writeBuffer = MediumBufferSize
		}
		logger.Debug("Adjusted buffers for Chrome browser")
	}
	
	return readBuffer, writeBuffer
}

// TransportConfig contains configuration for transport management
type TransportConfig struct {
	MaxIdleConns          int
	MaxIdleConnsPerHost   int
	IdleConnTimeout       int
	TLSHandshakeTimeout   int
	ExpectContinueTimeout int
	ReadBufferSize        int
	WriteBufferSize       int
}

// Global transport cache with last used tracking
var (
	upstreamCache     sync.Map // map[string]*transportCacheEntry
	cacheCleanupOnce  sync.Once
	cacheCleanupStop  chan struct{}
)

// transportCacheEntry holds a transport and its last used time
type transportCacheEntry struct {
	transport *http.Transport
	lastUsed  time.Time
	mu        sync.Mutex
}

// CreateOptimizedTransport creates optimized transport for direct connections
func CreateOptimizedTransport(config *TransportConfig) *http.Transport {
	return &http.Transport{
		// Connection pooling settings
		MaxIdleConns:        config.MaxIdleConns,
		MaxIdleConnsPerHost: config.MaxIdleConnsPerHost,
		IdleConnTimeout:     time.Duration(config.IdleConnTimeout) * time.Second,

		// Timeouts
		TLSHandshakeTimeout:   time.Duration(config.TLSHandshakeTimeout) * time.Second,
		ExpectContinueTimeout: time.Duration(config.ExpectContinueTimeout) * time.Second,

		// Buffer settings
		ReadBufferSize:  config.ReadBufferSize,
		WriteBufferSize: config.WriteBufferSize,

		// Connection settings
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,

		// TLS settings
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify:     false, // Fixed security issue
			SessionTicketsDisabled: false, // Enable TLS session resumption
		},

		// Enable HTTP/2
		ForceAttemptHTTP2: true,

		// Disable compression to reduce CPU usage
		DisableCompression: true,
	}
}

// CreateChromeOptimizedTransport creates transport optimized for Chrome browsers
func CreateChromeOptimizedTransport(config *TransportConfig, logger *slog.Logger) *http.Transport {
	// Chrome typically opens more connections per host
	// and benefits from higher connection limits
	chromeConfig := *config
	
	// Chrome opens up to 6 connections per host by default
	// We'll increase this to handle Chrome's aggressive connection behavior
	if chromeConfig.MaxIdleConnsPerHost < 20 {
		chromeConfig.MaxIdleConnsPerHost = 20
	}
	
	// Increase total idle connections for Chrome's multi-tab behavior
	if chromeConfig.MaxIdleConns < 200 {
		chromeConfig.MaxIdleConns = 200
	}
	
	logger.Debug("Chrome transport optimization applied",
		"original_max_idle_conns", config.MaxIdleConns,
		"optimized_max_idle_conns", chromeConfig.MaxIdleConns,
		"original_max_idle_conns_per_host", config.MaxIdleConnsPerHost,
		"optimized_max_idle_conns_per_host", chromeConfig.MaxIdleConnsPerHost)
	
	transport := &http.Transport{
		// Connection pooling settings optimized for Chrome
		MaxIdleConns:        chromeConfig.MaxIdleConns,
		MaxIdleConnsPerHost: chromeConfig.MaxIdleConnsPerHost,
		IdleConnTimeout:     time.Duration(chromeConfig.IdleConnTimeout) * time.Second,

		// Timeouts
		TLSHandshakeTimeout:   time.Duration(chromeConfig.TLSHandshakeTimeout) * time.Second,
		ExpectContinueTimeout: time.Duration(chromeConfig.ExpectContinueTimeout) * time.Second,

		// Buffer settings - Chrome handles its own buffering well
		ReadBufferSize:  chromeConfig.ReadBufferSize,
		WriteBufferSize: chromeConfig.WriteBufferSize,

		// Connection settings
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,

		// TLS settings
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify:     false,
			SessionTicketsDisabled: false, // Chrome benefits from session resumption
		},

		// Chrome supports HTTP/2
		ForceAttemptHTTP2: true,

		// Chrome handles its own compression
		DisableCompression: true,
		
		// Increase max response header size for Chrome's verbose headers
		MaxResponseHeaderBytes: 65536, // 64KB
	}
	
	return transport
}

// CreateHTTPProxyTransport creates transport for upstream HTTP proxy
func CreateHTTPProxyTransport(proxyURL string, username, password string, config *TransportConfig, logger *slog.Logger) (*http.Transport, error) {
	logger.Debug("Creating HTTP proxy transport",
		"proxy_url", proxyURL,
		"has_auth", username != "")

	parsedURL, err := url.Parse(proxyURL)
	if err != nil {
		logger.Debug("Failed to parse proxy URL", "error", err)
		return nil, err
	}

	// Add authentication if provided
	if username != "" && password != "" {
		parsedURL.User = url.UserPassword(username, password)
		logger.Debug("Added authentication to proxy URL")
	}

	transport := CreateOptimizedTransport(config)
	transport.Proxy = http.ProxyURL(parsedURL)

	logger.Debug("HTTP proxy transport created successfully",
		"proxy_host", parsedURL.Host)

	return transport, nil
}

// CreateSOCKS5ProxyTransport creates transport for upstream SOCKS5 proxy
func CreateSOCKS5ProxyTransport(proxyAddr string, username, password string, config *TransportConfig, logger *slog.Logger) (*http.Transport, error) {
	logger.Debug("Creating SOCKS5 proxy transport",
		"proxy_addr", proxyAddr,
		"has_auth", username != "")

	var auth *proxy.Auth
	if username != "" && password != "" {
		auth = &proxy.Auth{
			User:     username,
			Password: password,
		}
		logger.Debug("Using SOCKS5 authentication")
	}

	// Parse SOCKS5 URL if needed
	if strings.HasPrefix(proxyAddr, "socks5://") {
		u, err := url.Parse(proxyAddr)
		if err != nil {
			logger.Debug("Failed to parse SOCKS5 URL", "error", err)
			return nil, err
		}
		proxyAddr = u.Host
		logger.Debug("Extracted SOCKS5 host", "host", proxyAddr)
	}

	dialer, err := proxy.SOCKS5("tcp", proxyAddr, auth, proxy.Direct)
	if err != nil {
		logger.Debug("Failed to create SOCKS5 dialer", "error", err)
		return nil, err
	}

	transport := CreateOptimizedTransport(config)
	transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		logger.Debug("SOCKS5 dialing",
			"network", network,
			"addr", addr,
			"via", proxyAddr)
		return dialer.Dial(network, addr)
	}

	logger.Debug("SOCKS5 proxy transport created successfully",
		"proxy_addr", proxyAddr)

	return transport, nil
}

// DialThroughHTTPProxy dials through an HTTP proxy using CONNECT method
func DialThroughHTTPProxy(network, targetAddr string, proxyHost, proxyPort, username, password string, logger *slog.Logger) (net.Conn, error) {
	proxyAddr := net.JoinHostPort(proxyHost, proxyPort)

	logger.Debug("Dialing through HTTP proxy",
		"proxy", proxyAddr,
		"target", targetAddr,
		"has_auth", username != "")

	// Connect to proxy
	conn, err := net.DialTimeout("tcp", proxyAddr, DefaultTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to proxy: %w", err)
	}

	// Send CONNECT request
	connectReq := fmt.Sprintf("CONNECT %s HTTP/1.1\r\nHost: %s\r\n", targetAddr, targetAddr)

	// Add authentication if provided
	if username != "" && password != "" {
		auth := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
		connectReq += fmt.Sprintf("Proxy-Authorization: Basic %s\r\n", auth)
	}

	connectReq += "\r\n"

	if _, err := conn.Write([]byte(connectReq)); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to send CONNECT request: %w", err)
	}

	// Read response
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to read CONNECT response: %w", err)
	}

	response := string(buf[:n])
	logger.Debug("HTTP proxy CONNECT response", "response", strings.Split(response, "\r\n")[0])

	// Check if connection was successful
	if !strings.Contains(response, "200") {
		conn.Close()
		return nil, fmt.Errorf("proxy CONNECT failed: %s", strings.Split(response, "\r\n")[0])
	}

	return conn, nil
}

// DialThroughSOCKS5Proxy dials through a SOCKS5 proxy
func DialThroughSOCKS5Proxy(network, targetAddr string, proxyHost, proxyPort, username, password string, logger *slog.Logger) (net.Conn, error) {
	proxyAddr := net.JoinHostPort(proxyHost, proxyPort)

	logger.Debug("Dialing through SOCKS5 proxy",
		"proxy", proxyAddr,
		"target", targetAddr,
		"has_auth", username != "")

	var auth *proxy.Auth
	if username != "" && password != "" {
		auth = &proxy.Auth{
			User:     username,
			Password: password,
		}
	}

	dialer, err := proxy.SOCKS5("tcp", proxyAddr, auth, proxy.Direct)
	if err != nil {
		return nil, fmt.Errorf("failed to create SOCKS5 dialer: %w", err)
	}

	conn, err := dialer.Dial(network, targetAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial through SOCKS5: %w", err)
	}

	return conn, nil
}

// InitTransportCacheCleanup starts the periodic cache cleanup
func InitTransportCacheCleanup(interval time.Duration, maxAge time.Duration, logger *slog.Logger) {
	cacheCleanupOnce.Do(func() {
		cacheCleanupStop = make(chan struct{})
		go runCacheCleanup(interval, maxAge, logger)
		logger.Info("Transport cache cleanup initialized",
			"interval", interval,
			"max_age", maxAge)
	})
}

// StopTransportCacheCleanup stops the cache cleanup goroutine
func StopTransportCacheCleanup() {
	if cacheCleanupStop != nil {
		close(cacheCleanupStop)
	}
}

// runCacheCleanup periodically removes stale transports from cache
func runCacheCleanup(interval time.Duration, maxAge time.Duration, logger *slog.Logger) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cleanupTransportCache(maxAge, logger)
		case <-cacheCleanupStop:
			logger.Debug("Transport cache cleanup stopped")
			return
		}
	}
}

// cleanupTransportCache removes stale entries from the cache
func cleanupTransportCache(maxAge time.Duration, logger *slog.Logger) {
	now := time.Now()
	var cleaned int
	var checked int

	upstreamCache.Range(func(key, value interface{}) bool {
		checked++
		entry, ok := value.(*transportCacheEntry)
		if !ok {
			return true
		}

		entry.mu.Lock()
		age := now.Sub(entry.lastUsed)
		entry.mu.Unlock()

		if age > maxAge {
			// Close idle connections before removing
			if entry.transport != nil {
				entry.transport.CloseIdleConnections()
			}
			upstreamCache.Delete(key)
			cleaned++
			logger.Debug("Removed stale transport from cache",
				"key", key,
				"age", age)
		}
		return true
	})

	if cleaned > 0 {
		logger.Info("Transport cache cleanup completed",
			"checked", checked,
			"cleaned", cleaned)
	}
}

// GetUpstreamTransport gets or creates transport for the given upstream
func GetUpstreamTransport(upstream *UpstreamInfo, config *TransportConfig, logger *slog.Logger) (*http.Transport, error) {
	// Create cache key
	cacheKey := fmt.Sprintf("%s:%s:%s", upstream.Type, upstream.Host, upstream.Port)

	// Check cache first
	if cached, ok := upstreamCache.Load(cacheKey); ok {
		entry := cached.(*transportCacheEntry)
		entry.mu.Lock()
		entry.lastUsed = time.Now()
		entry.mu.Unlock()
		
		logger.Debug("Using cached transport",
			"cache_key", cacheKey,
			"type", upstream.Type)
		return entry.transport, nil
	}

	logger.Debug("Creating new transport",
		"type", upstream.Type,
		"host", upstream.Host,
		"port", upstream.Port,
		"has_auth", upstream.Username != "")

	// Create new transport
	var transport *http.Transport
	var err error

	proxyURL := fmt.Sprintf("%s://%s:%s", upstream.Type, upstream.Host, upstream.Port)

	if upstream.Type == "socks5" {
		transport, err = CreateSOCKS5ProxyTransport(
			proxyURL,
			upstream.Username,
			upstream.Password,
			config,
			logger,
		)
	} else {
		transport, err = CreateHTTPProxyTransport(
			proxyURL,
			upstream.Username,
			upstream.Password,
			config,
			logger,
		)
	}

	if err != nil {
		logger.Debug("Failed to create transport",
			"type", upstream.Type,
			"error", err)
		return nil, err
	}

	// Cache the transport with timestamp
	entry := &transportCacheEntry{
		transport: transport,
		lastUsed:  time.Now(),
	}
	upstreamCache.Store(cacheKey, entry)
	logger.Debug("Transport cached", "cache_key", cacheKey)

	return transport, nil
}
