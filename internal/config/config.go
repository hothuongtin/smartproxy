package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the complete configuration structure
type Config struct {
	Server           ServerConfig  `yaml:"server"`
	AdBlocking       AdBlockConfig `yaml:"ad_blocking"`
	DirectExtensions []string      `yaml:"direct_extensions"`
	DirectDomains    []string      `yaml:"direct_domains"`
	Logging          LoggingConfig `yaml:"logging"`
}

// ServerConfig represents server configuration
type ServerConfig struct {
	HTTPPort              int    `yaml:"http_port"`
	HTTPSMitm             bool   `yaml:"https_mitm"`
	CACert                string `yaml:"ca_cert"`
	CAKey                 string `yaml:"ca_key"`
	MaxIdleConns          int    `yaml:"max_idle_conns"`
	MaxIdleConnsPerHost   int    `yaml:"max_idle_conns_per_host"`
	IdleConnTimeout       int    `yaml:"idle_conn_timeout"`
	TLSHandshakeTimeout   int    `yaml:"tls_handshake_timeout"`
	ExpectContinueTimeout int    `yaml:"expect_continue_timeout"`
	ReadBufferSize        int    `yaml:"read_buffer_size"`
	WriteBufferSize       int    `yaml:"write_buffer_size"`
}

// AdBlockConfig represents ad blocking configuration
type AdBlockConfig struct {
	Enabled     bool   `yaml:"enabled"`
	DomainsFile string `yaml:"domains_file"`
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

// AdDomainsConfig represents the ad domains configuration
type AdDomainsConfig struct {
	AdDomains []string `yaml:"ad_domains"`
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// LoadAdDomains loads ad domains from a YAML file
func LoadAdDomains(adDomainsPath string) (*AdDomainsConfig, error) {
	data, err := os.ReadFile(adDomainsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read ad domains file: %w", err)
	}

	var adDomains AdDomainsConfig
	if err := yaml.Unmarshal(data, &adDomains); err != nil {
		return nil, fmt.Errorf("failed to parse ad domains file: %w", err)
	}

	return &adDomains, nil
}

// GetListenAddr returns the listen address based on config
func (c *Config) GetListenAddr() string {
	if c.Server.HTTPPort > 0 {
		return fmt.Sprintf(":%d", c.Server.HTTPPort)
	}
	return ":8080" // default
}

// SetDefaults sets default values for performance settings
func (c *Config) SetDefaults() {
	// Server defaults
	if c.Server.HTTPPort == 0 {
		c.Server.HTTPPort = 8888
	}
	if c.Server.MaxIdleConns == 0 {
		c.Server.MaxIdleConns = 10000
	}
	if c.Server.MaxIdleConnsPerHost == 0 {
		c.Server.MaxIdleConnsPerHost = 100
	}
	if c.Server.IdleConnTimeout == 0 {
		c.Server.IdleConnTimeout = 90
	}
	if c.Server.TLSHandshakeTimeout == 0 {
		c.Server.TLSHandshakeTimeout = 10
	}
	if c.Server.ExpectContinueTimeout == 0 {
		c.Server.ExpectContinueTimeout = 1
	}
	if c.Server.ReadBufferSize == 0 {
		c.Server.ReadBufferSize = 65536
	}
	if c.Server.WriteBufferSize == 0 {
		c.Server.WriteBufferSize = 65536
	}

	// Ad blocking defaults
	if c.AdBlocking.DomainsFile == "" {
		c.AdBlocking.DomainsFile = "ad_domains.yaml"
	}

	// Logging defaults
	if c.Logging.Level == "" {
		c.Logging.Level = "info"
	}
	if c.Logging.Format == "" {
		c.Logging.Format = "text"
	}

	// Set default extensions if empty
	if len(c.DirectExtensions) == 0 {
		c.DirectExtensions = []string{
			".js", ".css", ".jpg", ".jpeg", ".png", ".gif", ".ico", ".svg", ".webp",
			".woff", ".woff2", ".ttf", ".eot", ".mp4", ".webm", ".mp3", ".wav",
			".pdf", ".zip", ".gz", ".tar", ".rar", ".7z",
		}
	}

	// Set default CDN domains if empty
	if len(c.DirectDomains) == 0 {
		c.DirectDomains = []string{
			"cdn.", "cdnjs.", "cloudflare.", "googleapis.", "gstatic.",
			"unpkg.com", "jsdelivr.net", "bootstrapcdn.com", "jquery.com",
			"staticfile.org", "akamai.", "fastly.", "cloudfront.",
		}
	}
}

// CreateAdDomainsMap creates a map for O(1) ad domain lookup
func CreateAdDomainsMap(adDomains []string) map[string]bool {
	adMap := make(map[string]bool, len(adDomains))
	for _, domain := range adDomains {
		adMap[strings.ToLower(domain)] = true
	}
	return adMap
}
