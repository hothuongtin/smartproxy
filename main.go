package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/MatusOllah/slogcolor"
	"github.com/elazarl/goproxy"
	"golang.org/x/net/proxy"
)

// Global configuration variables
var (
	yamlConfig      *Config
	adDomainsConfig *AdDomainsConfig
	adDomainsMap    map[string]bool
	adDomainsMutex  sync.RWMutex
	logger          *slog.Logger
)

// Constants for better code maintainability
const (
	defaultTimeout       = 30 * time.Second
	connectBufferSize    = 4096
	defaultListenPort    = 8888
	httpsDefaultPort     = ":443"
)

// Check if URL is a static file
func isStaticFile(urlStr string) bool {
	// Parse URL to get path
	u, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	// Get path (already excludes query string and fragment)
	path := u.Path
	
	// Check extension
	lowerPath := strings.ToLower(path)
	if yamlConfig != nil {
		for _, ext := range yamlConfig.DirectExtensions {
			if strings.HasSuffix(lowerPath, ext) {
				return true
			}
		}
	}

	return false
}

// Check if domain is a CDN domain
func isCDNDomain(host string) bool {
	lowerHost := strings.ToLower(host)

	if yamlConfig != nil {
		for _, cdn := range yamlConfig.DirectDomains {
			if strings.Contains(lowerHost, cdn) {
				return true
			}
		}
	}

	return false
}

// Check if domain is in ad blocking list (optimized with map)
func isAdDomain(host string) bool {
	if yamlConfig == nil || !yamlConfig.AdBlocking.Enabled || adDomainsMap == nil {
		return false
	}

	lowerHost := strings.ToLower(host)
	
	// Use read lock for concurrent access
	adDomainsMutex.RLock()
	defer adDomainsMutex.RUnlock()
	
	// Check exact match first
	if adDomainsMap[lowerHost] {
		return true
	}
	
	// Check if any parent domain is blocked
	parts := strings.Split(lowerHost, ".")
	for i := range parts {
		domain := strings.Join(parts[i:], ".")
		if adDomainsMap[domain] {
			return true
		}
	}
	
	return false
}

// Create optimized transport for direct connections
func createOptimizedTransport(config *Config) *http.Transport {
	return &http.Transport{
		// Connection pooling settings
		MaxIdleConns:          config.Server.MaxIdleConns,
		MaxIdleConnsPerHost:   config.Server.MaxIdleConnsPerHost,
		IdleConnTimeout:       time.Duration(config.Server.IdleConnTimeout) * time.Second,
		
		// Timeouts
		TLSHandshakeTimeout:   time.Duration(config.Server.TLSHandshakeTimeout) * time.Second,
		ExpectContinueTimeout: time.Duration(config.Server.ExpectContinueTimeout) * time.Second,
		
		// Buffer settings
		ReadBufferSize:  config.Server.ReadBufferSize,
		WriteBufferSize: config.Server.WriteBufferSize,
		
		// Connection settings
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		
		// TLS settings
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false, // Fixed security issue
			SessionTicketsDisabled: false, // Enable TLS session resumption
		},
		
		// Enable HTTP/2
		ForceAttemptHTTP2: true,
		
		// Disable compression to reduce CPU usage
		DisableCompression: true,
	}
}

// Create transport for upstream HTTP proxy
func createHTTPProxyTransport(proxyURL string, username, password string, config *Config) (*http.Transport, error) {
	parsedURL, err := url.Parse(proxyURL)
	if err != nil {
		return nil, err
	}

	// Add authentication if provided
	if username != "" && password != "" {
		parsedURL.User = url.UserPassword(username, password)
	}

	transport := createOptimizedTransport(config)
	transport.Proxy = http.ProxyURL(parsedURL)

	return transport, nil
}

// Create transport for upstream SOCKS5 proxy
func createSOCKS5ProxyTransport(proxyAddr string, username, password string, config *Config) (*http.Transport, error) {
	var auth *proxy.Auth
	if username != "" && password != "" {
		auth = &proxy.Auth{
			User:     username,
			Password: password,
		}
	}

	// Parse SOCKS5 URL if needed
	if strings.HasPrefix(proxyAddr, "socks5://") {
		u, err := url.Parse(proxyAddr)
		if err != nil {
			return nil, err
		}
		proxyAddr = u.Host
	}

	dialer, err := proxy.SOCKS5("tcp", proxyAddr, auth, proxy.Direct)
	if err != nil {
		return nil, err
	}

	transport := createOptimizedTransport(config)
	transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		return dialer.Dial(network, addr)
	}

	return transport, nil
}

// Load ad domains into map for O(1) lookup
func loadAdDomainsMap(adDomains []string) {
	newMap := make(map[string]bool, len(adDomains))
	for _, domain := range adDomains {
		newMap[strings.ToLower(domain)] = true
	}
	
	adDomainsMutex.Lock()
	adDomainsMap = newMap
	adDomainsMutex.Unlock()
}

// setupLogger configures slog with colored output using slogcolor
func setupLogger() *slog.Logger {
	opts := &slogcolor.Options{
		Level: slog.LevelInfo,
		TimeFormat: "15:04:05",
		SrcFileMode: slogcolor.ShortFile,
		SrcFileLength: 0,
		MsgPrefix: "",
	}
	
	handler := slogcolor.NewHandler(os.Stdout, opts)
	return slog.New(handler)
}

func main() {
	// Setup logger
	logger = setupLogger()
	
	// Load configuration from YAML file
	configFile := "config.yaml"
	if envConfig := os.Getenv("SMARTPROXY_CONFIG"); envConfig != "" {
		configFile = envConfig
	}

	config, err := LoadConfig(configFile)
	if err != nil {
		logger.Error("Failed to load configuration", "error", err, "config", configFile)
		os.Exit(1)
	}
	yamlConfig = config

	// Set default values
	yamlConfig.SetDefaults()

	logger.Info("Loaded configuration", "config", configFile)

	// Load ad domains if ad blocking is enabled
	if yamlConfig.AdBlocking.Enabled {
		adDomainsConfig, err = LoadAdDomains(yamlConfig.AdBlocking.DomainsFile)
		if err != nil {
			logger.Warn("Failed to load ad domains", "error", err, "file", yamlConfig.AdBlocking.DomainsFile)
		} else {
			loadAdDomainsMap(adDomainsConfig.AdDomains)
			logger.Info("Loaded ad domains", "count", len(adDomainsConfig.AdDomains))
		}
	}

	// Create proxy server
	proxyServer := goproxy.NewProxyHttpServer()
	proxyServer.Verbose = false // Disable verbose logging for performance

	// Create optimized transports
	directTransport := createOptimizedTransport(yamlConfig)
	
	// Create upstream proxy transport (REQUIRED)
	if yamlConfig.Upstream.ProxyURL == "" {
		logger.Error("Upstream proxy is required", "help", "Please configure upstream.proxy_url in config.yaml")
		os.Exit(1)
	}
	
	var upstreamTransport *http.Transport
	
	// Determine proxy type from URL
	if strings.HasPrefix(yamlConfig.Upstream.ProxyURL, "socks5://") {
		var err error
		upstreamTransport, err = createSOCKS5ProxyTransport(
			yamlConfig.Upstream.ProxyURL,
			yamlConfig.Upstream.Username,
			yamlConfig.Upstream.Password,
			yamlConfig,
		)
		if err != nil {
			logger.Error("Failed to create upstream proxy transport", "error", err, "type", "socks5")
			os.Exit(1)
		}
	} else {
		var err error
		upstreamTransport, err = createHTTPProxyTransport(
			yamlConfig.Upstream.ProxyURL,
			yamlConfig.Upstream.Username,
			yamlConfig.Upstream.Password,
			yamlConfig,
		)
		if err != nil {
			logger.Error("Failed to create upstream proxy transport", "error", err, "type", "socks5")
			os.Exit(1)
		}
	}

	logger.Info("Using upstream proxy", "url", yamlConfig.Upstream.ProxyURL)

	// Handle HTTPS CONNECT
	if yamlConfig.Server.HTTPSMitm {
		// Enable MITM for HTTPS interception
		if yamlConfig.Server.CACert != "" && yamlConfig.Server.CAKey != "" {
			// Load custom CA certificate
			ca, err := tls.LoadX509KeyPair(yamlConfig.Server.CACert, yamlConfig.Server.CAKey)
			if err != nil {
				logger.Error("Failed to load CA certificate", "error", err, "cert", yamlConfig.Server.CACert, "key", yamlConfig.Server.CAKey)
				os.Exit(1)
			}
			goproxy.GoproxyCa = ca
			logger.Info("Loaded custom CA certificate", "cert", yamlConfig.Server.CACert)
		} else {
			logger.Info("Using default goproxy CA certificate for HTTPS interception")
			logger.Warn("Clients must trust the goproxy CA certificate to avoid TLS errors")
		}
		proxyServer.OnRequest().HandleConnect(goproxy.AlwaysMitm)
	} else {
		// No MITM - just tunnel HTTPS connections
		logger.Info("HTTPS MITM disabled - tunneling HTTPS connections without interception")
		
		// Configure proper HTTPS tunneling
		if yamlConfig.Upstream.ProxyURL != "" && strings.HasPrefix(yamlConfig.Upstream.ProxyURL, "socks5://") {
			// For SOCKS5 proxy
			var auth *proxy.Auth
			if yamlConfig.Upstream.Username != "" && yamlConfig.Upstream.Password != "" {
				auth = &proxy.Auth{
					User:     yamlConfig.Upstream.Username,
					Password: yamlConfig.Upstream.Password,
				}
			}
			u, _ := url.Parse(yamlConfig.Upstream.ProxyURL)
			socks5Dialer, err := proxy.SOCKS5("tcp", u.Host, auth, proxy.Direct)
			if err != nil {
				logger.Error("Failed to create SOCKS5 dialer", "error", err)
				os.Exit(1)
			}
			proxyServer.ConnectDial = socks5Dialer.Dial
		} else if yamlConfig.Upstream.ProxyURL != "" {
			// For HTTP proxy with CONNECT support
			proxyURL, err := url.Parse(yamlConfig.Upstream.ProxyURL)
			if err != nil {
				logger.Error("Failed to parse upstream proxy URL", "error", err, "url", yamlConfig.Upstream.ProxyURL)
				os.Exit(1)
			}
			
			// Create custom dialer for HTTP CONNECT proxy
			proxyServer.ConnectDial = func(network, addr string) (net.Conn, error) {
				// Connect to the HTTP proxy
				proxyConn, err := net.DialTimeout("tcp", proxyURL.Host, 30*time.Second)
				if err != nil {
					return nil, err
				}
				
				// Send CONNECT request
				connectReq := fmt.Sprintf("CONNECT %s HTTP/1.1\r\nHost: %s\r\n", addr, addr)
				
				// Add proxy authentication if needed
				if yamlConfig.Upstream.Username != "" && yamlConfig.Upstream.Password != "" {
					auth := yamlConfig.Upstream.Username + ":" + yamlConfig.Upstream.Password
					encoded := base64.StdEncoding.EncodeToString([]byte(auth))
					connectReq += fmt.Sprintf("Proxy-Authorization: Basic %s\r\n", encoded)
				}
				
				connectReq += "\r\n"
				
				_, err = proxyConn.Write([]byte(connectReq))
				if err != nil {
					proxyConn.Close()
					return nil, err
				}
				
				// Read response using buffered reader
				reader := bufio.NewReader(proxyConn)
				statusLine, err := reader.ReadString('\n')
				if err != nil {
					proxyConn.Close()
					return nil, fmt.Errorf("failed to read CONNECT response: %w", err)
				}
				
				// Parse status line (e.g., "HTTP/1.1 200 Connection established")
				parts := strings.Fields(statusLine)
				if len(parts) < 2 || parts[1] != "200" {
					proxyConn.Close()
					return nil, fmt.Errorf("proxy CONNECT failed: %s", strings.TrimSpace(statusLine))
				}
				
				// Read headers until empty line
				for {
					line, err := reader.ReadString('\n')
					if err != nil {
						proxyConn.Close()
						return nil, fmt.Errorf("failed to read CONNECT headers: %w", err)
					}
					if line == "\r\n" || line == "\n" {
						break
					}
				}
				
				return proxyConn, nil
			}
		}
	}

	// Block ad domains with optimized handler
	if yamlConfig.AdBlocking.Enabled && adDomainsMap != nil {
		proxyServer.OnRequest().DoFunc(
			func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
				if isAdDomain(r.Host) {
					// Return minimal blocking response
					return r, goproxy.NewResponse(r, goproxy.ContentTypeText, http.StatusNoContent, "")
				}
				return r, nil
			})
	}

	// Main routing logic (only applies when MITM is enabled or for HTTP requests)
	if yamlConfig.Server.HTTPSMitm {
		proxyServer.OnRequest().DoFunc(
			func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
				// Build full URL for checking
				fullURL := r.URL.String()
				if r.URL.Scheme == "" {
					fullURL = "http://" + r.Host + r.URL.Path
					if r.URL.RawQuery != "" {
						fullURL += "?" + r.URL.RawQuery
					}
				}

				// Determine which transport to use
				if isStaticFile(fullURL) || isCDNDomain(r.Host) {
					// Use direct connection for static files and CDNs
					ctx.RoundTripper = goproxy.RoundTripperFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Response, error) {
						return directTransport.RoundTrip(req)
					})
				} else {
					// Use upstream proxy for other requests
					ctx.RoundTripper = goproxy.RoundTripperFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Response, error) {
						return upstreamTransport.RoundTrip(req)
					})
				}

				return r, nil
			})
	} else {
		// For non-MITM mode, only handle HTTP requests (not HTTPS CONNECT)
		proxyServer.OnRequest().DoFunc(
			func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
				// Build full URL for checking
				fullURL := r.URL.String()
				if r.URL.Scheme == "" {
					fullURL = "http://" + r.Host + r.URL.Path
					if r.URL.RawQuery != "" {
						fullURL += "?" + r.URL.RawQuery
					}
				}
				
				// Check if it's a static file or CDN
				if isStaticFile(fullURL) || isCDNDomain(r.Host) {
					// Use direct connection
					ctx.RoundTripper = goproxy.RoundTripperFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Response, error) {
						return directTransport.RoundTrip(req)
					})
				} else {
					// Use upstream proxy
					ctx.RoundTripper = goproxy.RoundTripperFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Response, error) {
						return upstreamTransport.RoundTrip(req)
					})
				}
				return r, nil
			})
	}

	// Create HTTP server with optimized settings
	server := &http.Server{
		Addr:    yamlConfig.GetListenAddr(),
		Handler: proxyServer,
		
		// Timeouts to prevent slow clients from holding connections
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
		
		// Buffer sizes
		ReadHeaderTimeout: 10 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1MB
	}

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Shutting down proxy server...")
		
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		
		if err := server.Shutdown(ctx); err != nil {
			logger.Error("Server shutdown error", "error", err)
		}
	}()

	// Start server
	logger.Info("Starting high-performance proxy server", "address", yamlConfig.GetListenAddr())
	logger.Info("Configuration loaded", 
		"directExtensions", len(yamlConfig.DirectExtensions), 
		"cdnDomains", len(yamlConfig.DirectDomains))
	
	if yamlConfig.AdBlocking.Enabled {
		logger.Info("Ad blocking enabled", "domains", len(adDomainsMap))
	}
	
	
	logger.Info("Performance settings", 
		"maxIdleConns", yamlConfig.Server.MaxIdleConns, 
		"maxIdleConnsPerHost", yamlConfig.Server.MaxIdleConnsPerHost)

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		logger.Error("Server error", "error", err)
	}
	
	logger.Info("Server gracefully stopped")
}