package main

import (
	"fmt"
	"os"

	"github.com/hothuongtin/smartproxy/internal/config"
	"github.com/hothuongtin/smartproxy/internal/logger"
	"github.com/hothuongtin/smartproxy/internal/proxy"
)

func main() {
	// Load configuration from YAML file first to get logging settings
	configFile := "configs/config.yaml"
	if envConfig := os.Getenv("SMARTPROXY_CONFIG"); envConfig != "" {
		configFile = envConfig
	}

	yamlConfig, err := config.LoadConfig(configFile)
	if err != nil {
		// Use basic logger for error
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Set default values
	yamlConfig.SetDefaults()

	// Setup logger with config
	loggerConfig := &logger.Config{
		Level: yamlConfig.Logging.Level,
	}
	log := logger.SetupLogger(loggerConfig)

	log.Info("Loaded configuration",
		"config", configFile,
		"log_level", yamlConfig.Logging.Level)

	// Log configuration details in debug mode
	log.Debug("Configuration details",
		"http_port", yamlConfig.Server.HTTPPort,
		"https_mitm", yamlConfig.Server.HTTPSMitm,
		"max_idle_conns", yamlConfig.Server.MaxIdleConns,
		"max_idle_conns_per_host", yamlConfig.Server.MaxIdleConnsPerHost,
		"read_buffer_size", yamlConfig.Server.ReadBufferSize,
		"write_buffer_size", yamlConfig.Server.WriteBufferSize,
		"direct_extensions", len(yamlConfig.DirectExtensions),
		"direct_domains", len(yamlConfig.DirectDomains),
		"ad_blocking_enabled", yamlConfig.AdBlocking.Enabled)

	// Load ad domains if ad blocking is enabled
	var adDomainsMap map[string]bool
	if yamlConfig.AdBlocking.Enabled {
		log.Debug("Loading ad domains", "file", yamlConfig.AdBlocking.DomainsFile)
		adDomainsConfig, err := config.LoadAdDomains(yamlConfig.AdBlocking.DomainsFile)
		if err != nil {
			log.Warn("Failed to load ad domains", "error", err, "file", yamlConfig.AdBlocking.DomainsFile)
		} else {
			adDomainsMap = config.CreateAdDomainsMap(adDomainsConfig.AdDomains)
			proxy.SetAdDomainsMap(adDomainsMap)
			log.Info("Loaded ad domains", "count", len(adDomainsConfig.AdDomains))

			// Log sample domains in debug mode
			if len(adDomainsConfig.AdDomains) > 0 {
				sampleSize := 5
				if len(adDomainsConfig.AdDomains) < sampleSize {
					sampleSize = len(adDomainsConfig.AdDomains)
				}
				log.Debug("Sample ad domains",
					"samples", adDomainsConfig.AdDomains[:sampleSize],
					"total", len(adDomainsConfig.AdDomains))
			}
		}
	}

	// Smart proxy mode - upstream will be determined by auth credentials
	log.Info("Starting in smart proxy mode - upstream configured via authentication")
	log.Debug("Smart proxy authentication format",
		"username", "schema (http or socks5)",
		"password", "base64(host:port) or base64(host:port:user:pass)")

	// Create server configurations
	serverConfig := &proxy.Config{
		HTTPPort:   yamlConfig.Server.HTTPPort,
		HTTPSMitm:  yamlConfig.Server.HTTPSMitm,
		CACert:     yamlConfig.Server.CACert,
		CAKey:      yamlConfig.Server.CAKey,
		ListenAddr: yamlConfig.GetListenAddr(),
	}

	routingConfig := &proxy.RoutingConfig{
		DirectExtensions: yamlConfig.DirectExtensions,
		DirectDomains:    yamlConfig.DirectDomains,
		AdBlocking: struct {
			Enabled bool
		}{
			Enabled: yamlConfig.AdBlocking.Enabled,
		},
	}

	transportConfig := &proxy.TransportConfig{
		MaxIdleConns:          yamlConfig.Server.MaxIdleConns,
		MaxIdleConnsPerHost:   yamlConfig.Server.MaxIdleConnsPerHost,
		IdleConnTimeout:       yamlConfig.Server.IdleConnTimeout,
		TLSHandshakeTimeout:   yamlConfig.Server.TLSHandshakeTimeout,
		ExpectContinueTimeout: yamlConfig.Server.ExpectContinueTimeout,
		ReadBufferSize:        yamlConfig.Server.ReadBufferSize,
		WriteBufferSize:       yamlConfig.Server.WriteBufferSize,
	}

	// Initialize static extensions map for O(1) lookup
	proxy.InitStaticExtensions(yamlConfig.DirectExtensions)
	log.Debug("Initialized static extensions map", "count", len(yamlConfig.DirectExtensions))

	// Create and start the proxy server
	server := proxy.NewServer(serverConfig, routingConfig, transportConfig, log)

	log.Info("Configuration summary",
		"directExtensions", len(yamlConfig.DirectExtensions),
		"cdnDomains", len(yamlConfig.DirectDomains),
		"adBlockingEnabled", yamlConfig.AdBlocking.Enabled)

	if yamlConfig.AdBlocking.Enabled {
		log.Info("Ad blocking active", "domains", len(adDomainsMap))
	}

	log.Info("Performance settings",
		"maxIdleConns", yamlConfig.Server.MaxIdleConns,
		"maxIdleConnsPerHost", yamlConfig.Server.MaxIdleConnsPerHost)

	// Log authentication instructions in debug mode
	log.Debug("Authentication instructions",
		"format", "Basic Auth",
		"username", "schema (http|socks5)",
		"password", "base64(host:port[:user:pass])",
		"example_http", "curl -x http://http:$(echo -n 'proxy.example.com:8080' | base64)@localhost:8888",
		"example_socks5", "curl -x http://socks5:$(echo -n 'socks.example.com:1080:user:pass' | base64)@localhost:8888")

	log.Debug("Routing rules",
		"static_files", "direct connection",
		"cdn_domains", "direct connection",
		"ad_domains", "blocked (204)",
		"other", "upstream proxy (via auth)")

	// Start the server
	if err := server.Start(); err != nil {
		log.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
}
