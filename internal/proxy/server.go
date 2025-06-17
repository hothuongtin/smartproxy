package proxy

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/elazarl/goproxy"
)

// Constants
const (
	ConnectBufferSize = 4096
	DefaultListenPort = 8888
	HTTPSDefaultPort  = ":443"
)

// Server represents the SmartProxy server
type Server struct {
	config          *Config
	routingConfig   *RoutingConfig
	transportConfig *TransportConfig
	logger          *slog.Logger
	proxyServer     *goproxy.ProxyHttpServer
	directTransport *http.Transport
	chromeTransport *http.Transport

	// Global state
	connectUpstreams sync.Map // map[string]*UpstreamInfo (keyed by remote addr)
	targetUpstreams  sync.Map // map[string]*UpstreamInfo (keyed by target addr)
}

// Config represents the server configuration
type Config struct {
	HTTPPort   int
	HTTPSMitm  bool
	CACert     string
	CAKey      string
	ListenAddr string
}

// NewServer creates a new SmartProxy server
func NewServer(config *Config, routingConfig *RoutingConfig, transportConfig *TransportConfig, logger *slog.Logger) *Server {
	return &Server{
		config:          config,
		routingConfig:   routingConfig,
		transportConfig: transportConfig,
		logger:          logger,
		proxyServer:     goproxy.NewProxyHttpServer(),
	}
}

// Start starts the proxy server
func (s *Server) Start() error {
	s.proxyServer.Verbose = false // Disable verbose logging for performance

	// Create optimized transport for direct connections
	s.directTransport = CreateOptimizedTransport(s.transportConfig)
	s.logger.Debug("Created direct transport",
		"max_idle_conns", s.transportConfig.MaxIdleConns,
		"max_idle_conns_per_host", s.transportConfig.MaxIdleConnsPerHost,
		"idle_conn_timeout", s.transportConfig.IdleConnTimeout)

	// Create Chrome-optimized transport
	s.chromeTransport = CreateChromeOptimizedTransport(s.transportConfig, s.logger)
	s.logger.Debug("Created Chrome-optimized transport",
		"max_idle_conns", s.chromeTransport.MaxIdleConns,
		"max_idle_conns_per_host", s.chromeTransport.MaxIdleConnsPerHost)

	// Initialize transport cache cleanup
	// Clean up transports not used for 5 minutes, check every minute
	InitTransportCacheCleanup(1*time.Minute, 5*time.Minute, s.logger)

	// Setup HTTPS handling
	s.setupHTTPS()

	// Setup authentication middleware
	s.setupAuthentication()

	// Setup ad blocking
	s.setupAdBlocking()

	// Setup response logging
	s.setupResponseLogging()

	// Setup routing logic
	s.setupRouting()

	// Create and start HTTP server
	return s.startHTTPServer()
}

// setupHTTPS configures HTTPS handling
func (s *Server) setupHTTPS() {
	if s.config.HTTPSMitm {
		// Enable MITM for HTTPS interception
		if s.config.CACert != "" && s.config.CAKey != "" {
			// Load custom CA certificate
			ca, err := tls.LoadX509KeyPair(s.config.CACert, s.config.CAKey)
			if err != nil {
				s.logger.Error("Failed to load CA certificate", "error", err, "cert", s.config.CACert, "key", s.config.CAKey)
				os.Exit(1)
			}
			goproxy.GoproxyCa = ca
			s.logger.Info("Loaded custom CA certificate", "cert", s.config.CACert)
		} else {
			s.logger.Info("Using default goproxy CA certificate for HTTPS interception")
			s.logger.Warn("Clients must trust the goproxy CA certificate to avoid TLS errors")
		}
		// Add CONNECT handler with authentication for MITM
		s.proxyServer.OnRequest().HandleConnectFunc(func(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
			s.logger.Debug("HTTPS CONNECT request (MITM mode)", "host", host, "remote_addr", ctx.Req.RemoteAddr)

			// Check for authentication
			if ctx.Req == nil {
				s.logger.Debug("No request context for CONNECT (MITM)")
				return goproxy.RejectConnect, "No request context"
			}

			auth := ctx.Req.Header.Get("Proxy-Authorization")
			if auth == "" {
				s.logger.Debug("No authentication for CONNECT (MITM)", "host", host)
				// Set response headers for 407
				ctx.Resp = goproxy.NewResponse(ctx.Req, goproxy.ContentTypeText, http.StatusProxyAuthRequired, "Proxy Authentication Required")
				ctx.Resp.Header.Set("Proxy-Authenticate", `Basic realm="SmartProxy"`)
				return goproxy.RejectConnect, "Authentication required"
			}

			// Parse authentication
			if !strings.HasPrefix(auth, "Basic ") {
				s.logger.Debug("Invalid auth type for CONNECT (MITM)", "auth", auth[:min(10, len(auth))])
				return goproxy.RejectConnect, "Invalid authentication"
			}

			// Extract base64 part
			base64Auth := auth[6:]

			// Remove any whitespace/newlines that might have been inserted by the client
			base64Auth = strings.ReplaceAll(base64Auth, "\n", "")
			base64Auth = strings.ReplaceAll(base64Auth, "\r", "")
			base64Auth = strings.ReplaceAll(base64Auth, " ", "")
			base64Auth = strings.ReplaceAll(base64Auth, "\t", "")

			credentials, err := base64.StdEncoding.DecodeString(base64Auth)
			if err != nil {
				s.logger.Debug("Failed to decode CONNECT auth (MITM)",
					"error", err,
					"base64", base64Auth[:min(20, len(base64Auth))])
				return goproxy.RejectConnect, "Invalid authentication"
			}

			parts := strings.SplitN(string(credentials), ":", 2)
			if len(parts) != 2 {
				s.logger.Debug("Invalid CONNECT credential format (MITM)",
					"decoded", string(credentials),
					"parts", len(parts))
				return goproxy.RejectConnect, "Invalid authentication"
			}

			// Parse upstream from auth
			upstream, err := ParseUpstreamFromAuth(parts[0], parts[1], s.logger)
			if err != nil {
				s.logger.Error("Failed to parse upstream from CONNECT auth (MITM)", "error", err)
				ctx.Resp = goproxy.NewResponse(ctx.Req, goproxy.ContentTypeText, http.StatusForbidden,
					fmt.Sprintf("Account password authentication failed: %v", err))
				return goproxy.RejectConnect, "Invalid credentials"
			}

			// Store upstream info for later use
			ctx.UserData = upstream
			
			s.logger.Debug("CONNECT authentication successful (MITM)",
				"host", host,
				"upstream_type", upstream.Type,
				"upstream_host", upstream.Host)

			// Allow MITM after successful authentication
			return goproxy.MitmConnect, host
		})
	} else {
		// No MITM - setup tunneling with upstream proxy support
		s.setupHTTPSTunneling()
	}
}

// setupHTTPSTunneling configures HTTPS tunneling with upstream proxy support
func (s *Server) setupHTTPSTunneling() {
	s.logger.Info("HTTPS MITM disabled - tunneling HTTPS connections without interception")

	// Add CONNECT handler for authentication
	s.proxyServer.OnRequest().HandleConnectFunc(func(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
		s.logger.Debug("HTTPS CONNECT request", "host", host, "remote_addr", ctx.Req.RemoteAddr)

		// Check for authentication
		if ctx.Req == nil {
			s.logger.Debug("No request context for CONNECT")
			return goproxy.RejectConnect, "No request context"
		}

		auth := ctx.Req.Header.Get("Proxy-Authorization")
		if auth == "" {
			s.logger.Debug("No authentication for CONNECT", "host", host)
			// Set response headers for 407
			ctx.Resp = goproxy.NewResponse(ctx.Req, goproxy.ContentTypeText, http.StatusProxyAuthRequired, "Proxy Authentication Required")
			ctx.Resp.Header.Set("Proxy-Authenticate", `Basic realm="SmartProxy"`)
			return goproxy.RejectConnect, "Authentication required"
		}

		// Parse authentication
		if !strings.HasPrefix(auth, "Basic ") {
			s.logger.Debug("Invalid auth type for CONNECT", "auth", auth[:min(10, len(auth))])
			return goproxy.RejectConnect, "Invalid authentication"
		}

		// Extract base64 part
		base64Auth := auth[6:]

		// Remove any whitespace/newlines that might have been inserted by the client
		base64Auth = strings.ReplaceAll(base64Auth, "\n", "")
		base64Auth = strings.ReplaceAll(base64Auth, "\r", "")
		base64Auth = strings.ReplaceAll(base64Auth, " ", "")
		base64Auth = strings.ReplaceAll(base64Auth, "\t", "")

		s.logger.Debug("CONNECT auth base64 (after cleanup)",
			"length", len(base64Auth),
			"first_20", base64Auth[:min(20, len(base64Auth))],
			"last_20", base64Auth[max(0, len(base64Auth)-20):])

		credentials, err := base64.StdEncoding.DecodeString(base64Auth)
		if err != nil {
			s.logger.Debug("Failed to decode CONNECT auth",
				"error", err,
				"base64", base64Auth)
			return goproxy.RejectConnect, "Invalid authentication"
		}

		parts := strings.SplitN(string(credentials), ":", 2)
		if len(parts) != 2 {
			s.logger.Debug("Invalid CONNECT credential format",
				"decoded", string(credentials),
				"parts", len(parts))
			return goproxy.RejectConnect, "Invalid authentication"
		}

		// Parse upstream from auth
		upstream, err := ParseUpstreamFromAuth(parts[0], parts[1], s.logger)
		if err != nil {
			s.logger.Error("Failed to parse upstream from CONNECT auth", "error", err)
			ctx.Resp = goproxy.NewResponse(ctx.Req, goproxy.ContentTypeText, http.StatusForbidden,
				fmt.Sprintf("Account password authentication failed: %v", err))
			return goproxy.RejectConnect, "Invalid credentials"
		}

		// Store upstream info for later use
		ctx.UserData = upstream
		// Store by target address for ConnectDial
		targetKey := host
		if !strings.Contains(host, ":") {
			// Add default HTTPS port if not present
			targetKey = host + ":443"
		}
		s.targetUpstreams.Store(targetKey, upstream)
		// Clean up after some time to prevent memory leak
		go func(key string) {
			time.Sleep(5 * time.Minute)
			s.targetUpstreams.Delete(key)
		}(targetKey)

		// Also store by remote address for backwards compatibility
		if ctx.Req != nil {
			s.connectUpstreams.Store(ctx.Req.RemoteAddr, upstream)
			go func(addr string) {
				time.Sleep(5 * time.Minute)
				s.connectUpstreams.Delete(addr)
			}(ctx.Req.RemoteAddr)
		}

		s.logger.Debug("CONNECT authentication successful",
			"host", host,
			"upstream_type", upstream.Type,
			"upstream_host", upstream.Host)

		// Allow the connection
		return goproxy.OkConnect, host
	})

	// Set up custom dial function for HTTPS connections
	s.proxyServer.ConnectDial = func(network, addr string) (net.Conn, error) {
		s.logger.Debug("ConnectDial called", "network", network, "addr", addr)

		// Extract host for checking if it's a CDN or should use direct connection
		host, _, err := net.SplitHostPort(addr)
		if err != nil {
			// addr might not have a port
			host = addr
		}

		// Check if this should use direct connection
		if IsCDNDomain(host, s.routingConfig, s.logger) {
			s.logger.Debug("Using direct connection for CDN domain", "host", host)
			return net.DialTimeout(network, addr, DefaultTimeout)
		}

		// Look up upstream info by target address
		value, ok := s.targetUpstreams.Load(addr)
		if !ok {
			s.logger.Debug("No upstream found for target, using direct connection", "addr", addr)
			return net.DialTimeout(network, addr, DefaultTimeout)
		}

		upstream, ok := value.(*UpstreamInfo)
		if !ok {
			s.logger.Error("Invalid upstream info type", "addr", addr)
			return net.DialTimeout(network, addr, DefaultTimeout)
		}

		s.logger.Debug("Using upstream for HTTPS connection",
			"upstream_type", upstream.Type,
			"upstream_host", upstream.Host,
			"target_addr", addr)

		// Handle different upstream types
		switch upstream.Type {
		case "http":
			return DialThroughHTTPProxy(network, addr, upstream.Host, upstream.Port, upstream.Username, upstream.Password, s.logger)
		case "socks5":
			return DialThroughSOCKS5Proxy(network, addr, upstream.Host, upstream.Port, upstream.Username, upstream.Password, s.logger)
		default:
			s.logger.Error("Unknown upstream type", "type", upstream.Type)
			return net.DialTimeout(network, addr, DefaultTimeout)
		}
	}

	s.logger.Info("HTTPS tunneling configured with upstream proxy support")
}

// setupAuthentication configures authentication middleware for non-CONNECT requests
func (s *Server) setupAuthentication() {
	s.proxyServer.OnRequest().DoFunc(
		func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			startTime := time.Now()
			s.logger.Debug("Incoming request",
				"method", r.Method,
				"url", r.URL.String(),
				"host", r.Host,
				"remote_addr", r.RemoteAddr,
				"user_agent", r.Header.Get("User-Agent"))

			// For MITM requests, check if we already have upstream info from CONNECT
			if s.config.HTTPSMitm && ctx.UserData != nil {
				// Already authenticated during CONNECT phase
				s.logger.Debug("Using upstream from CONNECT phase (MITM)",
					"method", r.Method,
					"url", r.URL.String())
				return r, nil
			}

			// Check for Proxy-Authorization header
			auth := r.Header.Get("Proxy-Authorization")
			if auth == "" {
				s.logger.Debug("No authentication provided",
					"remote_addr", r.RemoteAddr,
					"response", "407 Proxy Authentication Required")
				// Return 407 Proxy Authentication Required
				resp := goproxy.NewResponse(r, goproxy.ContentTypeText, http.StatusProxyAuthRequired, "Proxy Authentication Required")
				resp.Header.Set("Proxy-Authenticate", `Basic realm="SmartProxy"`)
				return r, resp
			}

			// Parse Basic auth
			if !strings.HasPrefix(auth, "Basic ") {
				s.logger.Debug("Invalid authentication type",
					"auth_type", strings.Split(auth, " ")[0],
					"response", "400 Bad Request")
				return r, goproxy.NewResponse(r, goproxy.ContentTypeText, http.StatusBadRequest, "Invalid authentication")
			}

			// Extract and clean base64 part
			base64Auth := auth[6:]
			// Remove any whitespace/newlines that might have been inserted
			base64Auth = strings.ReplaceAll(base64Auth, "\n", "")
			base64Auth = strings.ReplaceAll(base64Auth, "\r", "")
			base64Auth = strings.ReplaceAll(base64Auth, " ", "")
			base64Auth = strings.ReplaceAll(base64Auth, "\t", "")

			credentials, err := base64.StdEncoding.DecodeString(base64Auth)
			if err != nil {
				s.logger.Debug("Failed to decode credentials",
					"error", err,
					"response", "400 Bad Request")
				return r, goproxy.NewResponse(r, goproxy.ContentTypeText, http.StatusBadRequest, "Invalid authentication")
			}

			// Split username:password
			parts := strings.SplitN(string(credentials), ":", 2)
			if len(parts) != 2 {
				s.logger.Debug("Invalid credential format",
					"parts", len(parts),
					"response", "400 Bad Request")
				return r, goproxy.NewResponse(r, goproxy.ContentTypeText, http.StatusBadRequest, "Invalid authentication")
			}

			username := parts[0]
			password := parts[1]

			// Parse upstream from auth
			upstream, err := ParseUpstreamFromAuth(username, password, s.logger)
			if err != nil {
				s.logger.Error("Failed to parse upstream from auth", "error", err)
				return r, goproxy.NewResponse(r, goproxy.ContentTypeText, http.StatusForbidden,
					fmt.Sprintf("Account password authentication failed: %v", err))
			}

			// Store upstream info in context for later use
			ctx.UserData = upstream

			// Remove Proxy-Authorization header before forwarding
			r.Header.Del("Proxy-Authorization")

			s.logger.Debug("Authentication successful",
				"upstream_type", upstream.Type,
				"upstream_host", upstream.Host,
				"upstream_port", upstream.Port,
				"duration", time.Since(startTime))

			return r, nil
		})
}

// setupAdBlocking configures ad blocking functionality
func (s *Server) setupAdBlocking() {
	if s.routingConfig.AdBlocking.Enabled {
		s.proxyServer.OnRequest().DoFunc(
			func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
				if IsAdDomain(r.Host, s.routingConfig, s.logger) {
					s.logger.Debug("Blocking ad domain request",
						"host", r.Host,
						"method", r.Method,
						"url", r.URL.String())
					// Return minimal blocking response
					return r, goproxy.NewResponse(r, goproxy.ContentTypeText, http.StatusNoContent, "")
				}
				return r, nil
			})
	}
}

// setupResponseLogging configures response logging in debug mode
func (s *Server) setupResponseLogging() {
	if s.logger.Enabled(nil, slog.LevelDebug) {
		s.proxyServer.OnResponse().DoFunc(
			func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
				if resp != nil {
					s.logger.Debug("Response received",
						"status", resp.StatusCode,
						"method", ctx.Req.Method,
						"url", ctx.Req.URL.String(),
						"content_length", resp.ContentLength,
						"content_type", resp.Header.Get("Content-Type"))
				}
				return resp
			})
	}
}

// setupRouting configures the main routing logic
func (s *Server) setupRouting() {
	if s.config.HTTPSMitm {
		// MITM enabled - handle all requests
		s.setupMITMRouting()
	} else {
		// Non-MITM mode - only handle HTTP requests
		s.setupNonMITMRouting()
	}
}

// setupMITMRouting configures routing when MITM is enabled
func (s *Server) setupMITMRouting() {
	s.proxyServer.OnRequest().DoFunc(
		func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			startTime := time.Now()

			// Build full URL for checking
			fullURL := r.URL.String()
			if r.URL.Scheme == "" {
				fullURL = "http://" + r.Host + r.URL.Path
				if r.URL.RawQuery != "" {
					fullURL += "?" + r.URL.RawQuery
				}
			}

			s.logger.Debug("Routing decision for request",
				"url", fullURL,
				"host", r.Host,
				"method", r.Method)

			// Check if this is a Chrome browser
			userAgent := r.Header.Get("User-Agent")
			isChrome := IsChromeBrowser(userAgent)
			
			// Determine which transport to use
			if IsStaticFile(fullURL, s.routingConfig, s.logger) || IsCDNDomain(r.Host, s.routingConfig, s.logger) {
				// Use direct connection for static files and CDNs
				s.logger.Debug("Using direct connection",
					"reason", "static_file_or_cdn",
					"url", fullURL,
					"is_chrome", isChrome,
					"duration", time.Since(startTime))

				// Select appropriate transport based on browser
				var transport *http.Transport
				if isChrome {
					transport = s.chromeTransport
					s.logger.Debug("Using Chrome-optimized transport for direct connection")
				} else {
					transport = s.directTransport
				}

				ctx.RoundTripper = goproxy.RoundTripperFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Response, error) {
					return transport.RoundTrip(req)
				})
			} else {
				// Get upstream from context
				upstream, ok := ctx.UserData.(*UpstreamInfo)
				if !ok || upstream == nil {
					s.logger.Error("No upstream info in context")
					return r, goproxy.NewResponse(r, goproxy.ContentTypeText, http.StatusInternalServerError, "No upstream configured")
				}

				s.logger.Debug("Using upstream proxy",
					"upstream_type", upstream.Type,
					"upstream_host", upstream.Host,
					"upstream_port", upstream.Port,
					"url", fullURL)

				// Get or create transport for this upstream
				upstreamTransport, err := GetUpstreamTransport(upstream, s.transportConfig, s.logger)
				if err != nil {
					s.logger.Error("Failed to get upstream transport", "error", err)
					return r, goproxy.NewResponse(r, goproxy.ContentTypeText, http.StatusBadGateway, "Upstream connection failed")
				}

				// Use upstream proxy for other requests
				ctx.RoundTripper = goproxy.RoundTripperFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Response, error) {
					respStart := time.Now()
					resp, err := upstreamTransport.RoundTrip(req)

					if err != nil {
						s.logger.Debug("Upstream request failed",
							"error", err,
							"duration", time.Since(respStart))
					} else {
						s.logger.Debug("Upstream request completed",
							"status", resp.StatusCode,
							"duration", time.Since(respStart))
					}

					return resp, err
				})
			}

			s.logger.Debug("Request routing configured",
				"total_duration", time.Since(startTime))

			return r, nil
		})
}

// setupNonMITMRouting configures routing when MITM is disabled
func (s *Server) setupNonMITMRouting() {
	s.proxyServer.OnRequest().DoFunc(
		func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			startTime := time.Now()

			// Build full URL for checking
			fullURL := r.URL.String()
			if r.URL.Scheme == "" {
				fullURL = "http://" + r.Host + r.URL.Path
				if r.URL.RawQuery != "" {
					fullURL += "?" + r.URL.RawQuery
				}
			}

			s.logger.Debug("Non-MITM routing decision",
				"url", fullURL,
				"host", r.Host,
				"method", r.Method)

			// Check if this is a Chrome browser
			userAgent := r.Header.Get("User-Agent")
			isChrome := IsChromeBrowser(userAgent)
			
			// Check if it's a static file or CDN
			if IsStaticFile(fullURL, s.routingConfig, s.logger) || IsCDNDomain(r.Host, s.routingConfig, s.logger) {
				// Use direct connection
				s.logger.Debug("Using direct connection (non-MITM)",
					"reason", "static_file_or_cdn",
					"url", fullURL,
					"is_chrome", isChrome,
					"duration", time.Since(startTime))

				// Select appropriate transport based on browser
				var transport *http.Transport
				if isChrome {
					transport = s.chromeTransport
					s.logger.Debug("Using Chrome-optimized transport for direct connection (non-MITM)")
				} else {
					transport = s.directTransport
				}
				
				ctx.RoundTripper = goproxy.RoundTripperFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Response, error) {
					respStart := time.Now()
					resp, err := transport.RoundTrip(req)

					if err != nil {
						s.logger.Debug("Direct request failed",
							"error", err,
							"duration", time.Since(respStart))
					} else {
						s.logger.Debug("Direct request completed",
							"status", resp.StatusCode,
							"duration", time.Since(respStart))
					}

					return resp, err
				})
			} else {
				// Get upstream from context
				upstream, ok := ctx.UserData.(*UpstreamInfo)
				if !ok || upstream == nil {
					s.logger.Error("No upstream info in context (non-MITM)")
					return r, goproxy.NewResponse(r, goproxy.ContentTypeText, http.StatusInternalServerError, "No upstream configured")
				}

				s.logger.Debug("Using upstream proxy (non-MITM)",
					"upstream_type", upstream.Type,
					"upstream_host", upstream.Host,
					"upstream_port", upstream.Port,
					"url", fullURL)

				// Get or create transport for this upstream
				upstreamTransport, err := GetUpstreamTransport(upstream, s.transportConfig, s.logger)
				if err != nil {
					s.logger.Error("Failed to get upstream transport", "error", err)
					return r, goproxy.NewResponse(r, goproxy.ContentTypeText, http.StatusBadGateway, "Upstream connection failed")
				}

				// Use upstream proxy
				ctx.RoundTripper = goproxy.RoundTripperFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Response, error) {
					respStart := time.Now()
					resp, err := upstreamTransport.RoundTrip(req)

					if err != nil {
						s.logger.Debug("Upstream request failed (non-MITM)",
							"error", err,
							"duration", time.Since(respStart))
					} else {
						s.logger.Debug("Upstream request completed (non-MITM)",
							"status", resp.StatusCode,
							"duration", time.Since(respStart))
					}

					return resp, err
				})
			}

			s.logger.Debug("Non-MITM routing configured",
				"total_duration", time.Since(startTime))

			return r, nil
		})
}

// startHTTPServer creates and starts the HTTP server
func (s *Server) startHTTPServer() error {
	// Create HTTP server with optimized settings
	server := &http.Server{
		Addr:    s.config.ListenAddr,
		Handler: s.proxyServer,

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
		s.logger.Info("Shutting down proxy server...")

		// Stop transport cache cleanup
		StopTransportCacheCleanup()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			s.logger.Error("Server shutdown error", "error", err)
		}
	}()

	// Start server
	s.logger.Info("Starting high-performance proxy server",
		"address", s.config.ListenAddr,
		"mode", "smart_proxy_auth")

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		s.logger.Error("Server error", "error", err)
		return err
	}

	s.logger.Info("Server gracefully stopped")
	return nil
}
