package proxy

import (
	"encoding/base64"
	"fmt"
	"log/slog"
	"strings"
)

// UpstreamInfo holds parsed upstream configuration
type UpstreamInfo struct {
	Host     string
	Port     string
	Username string
	Password string
	Type     string // http or socks5
}

// Helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// ParseUpstreamFromAuth parses upstream info from authentication credentials
func ParseUpstreamFromAuth(username, password string, logger *slog.Logger) (*UpstreamInfo, error) {
	logger.Debug("Parsing upstream from authentication",
		"username", username,
		"password_length", len(password),
		"password_first_20", password[:min(20, len(password))],
		"password_last_20", password[max(0, len(password)-20):])

	// Check for any trailing whitespace or special characters
	if len(password) > 0 {
		lastChar := password[len(password)-1]
		logger.Debug("Password details",
			"last_char_byte", lastChar,
			"last_char_printable", string(lastChar),
			"is_whitespace", lastChar == ' ' || lastChar == '\t' || lastChar == '\n' || lastChar == '\r',
			"full_password", fmt.Sprintf("%q", password)) // %q shows escaped chars
	}

	// Username is the schema (http or socks5)
	schema := strings.ToLower(username)
	if schema != "http" && schema != "socks5" {
		logger.Debug("Invalid schema in authentication", "schema", schema)
		return nil, fmt.Errorf("invalid schema: %s, must be http or socks5", schema)
	}

	// Clean the password - remove any whitespace/newlines that might have been added
	// Some base64 implementations wrap at 76 characters
	password = strings.ReplaceAll(password, "\n", "")
	password = strings.ReplaceAll(password, "\r", "")
	password = strings.ReplaceAll(password, " ", "")
	password = strings.ReplaceAll(password, "\t", "")
	logger.Debug("After cleaning whitespace", "password_length", len(password))

	// Decode password from base64
	decoded, err := base64.StdEncoding.DecodeString(password)
	if err != nil {
		logger.Debug("Failed to decode base64 password",
			"error", err,
			"password", password,
			"password_bytes", []byte(password))

		// Try to identify the problematic character
		if len(password) >= 72 {
			logger.Debug("Character at position 72",
				"char", password[71],
				"char_byte", password[71],
				"char_printable", string(password[71]))
		}

		return nil, fmt.Errorf("failed to decode password: %w", err)
	}

	decodedStr := string(decoded)
	logger.Debug("Decoded upstream configuration",
		"decoded", decodedStr,
		"schema", schema)

	// Parse decoded string
	// Format: host:port or host:port:username:password
	parts := strings.Split(decodedStr, ":")
	if len(parts) < 2 {
		logger.Debug("Invalid upstream format", "parts", len(parts), "decoded", decodedStr)
		return nil, fmt.Errorf("invalid upstream format, expected host:port")
	}

	upstream := &UpstreamInfo{
		Host: parts[0],
		Port: parts[1],
		Type: schema,
	}

	// Check if authentication is provided
	if len(parts) >= 4 {
		upstream.Username = parts[2]
		upstream.Password = parts[3]
		logger.Debug("Upstream includes authentication",
			"host", upstream.Host,
			"port", upstream.Port,
			"has_auth", true)
	} else {
		logger.Debug("Upstream without authentication",
			"host", upstream.Host,
			"port", upstream.Port,
			"has_auth", false)
	}

	return upstream, nil
}
