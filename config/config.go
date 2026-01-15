// Package config provides configuration loading from environment variables.
// Doc: api/authentication.mdx - API key configuration
package config

import (
	"encoding/base64"
	"fmt"
	"os"

	"golang.org/x/crypto/ed25519"
)

// Config holds all configuration for the Polymarket API client.
// Environment variables are documented in CLAUDE.md.
type Config struct {
	// APIKey is the API key ID (UUID) for authentication.
	// Env: POLYMARKET_API_KEY
	// Doc: api/authentication.mdx - X-PM-Access-Key header
	APIKey string

	// PrivateKey is the Ed25519 private key for signing requests.
	// Env: POLYMARKET_PRIVATE_KEY (base64 encoded)
	// Doc: api/authentication.mdx - Ed25519 signature generation
	PrivateKey ed25519.PrivateKey

	// Symbol is the market slug to trade.
	// Env: POLYMARKET_SYMBOL
	// Doc: api-reference/market/overview.mdx - market slug identifier
	Symbol string

	// BaseURL is the API base URL.
	// Env: POLYMARKET_BASE_URL (default: https://api.polymarket.us)
	// Doc: api-reference/oapi-schemas/orders-schema.json - servers section
	BaseURL string

	// WSPrivateURL is the WebSocket URL for private data.
	// Doc: api-reference/websocket/private.mdx - endpoint
	WSPrivateURL string

	// WSMarketsURL is the WebSocket URL for market data.
	// Doc: api-reference/websocket/markets.mdx - endpoint
	WSMarketsURL string

	// InsecureSkipVerify disables TLS certificate verification.
	// Use only for staging/development with self-signed certs.
	// Env: INSECURE_SKIP_VERIFY=true
	InsecureSkipVerify bool
}

// getEnvWithFallback returns the first non-empty value from the given env var names.
// This allows the harness to set variables only if not already set.
func getEnvWithFallback(names ...string) string {
	for _, name := range names {
		if val := os.Getenv(name); val != "" {
			return val
		}
	}
	return ""
}

// Load loads configuration from environment variables.
// Variables are checked with fallbacks to support both direct usage and harness integration:
//   - POLYMARKET_API_KEY or TEST_API_KEY_ID
//   - POLYMARKET_PRIVATE_KEY or TEST_API_SECRET_KEY
//   - POLYMARKET_SYMBOL or TEST_MARKET_SLUG
//   - POLYMARKET_BASE_URL or RETAIL_API_URL (default: https://api.polymarket.us)
//   - POLYMARKET_WS_URL or RETAIL_WS_URL (default: derived from base URL)
func Load() (*Config, error) {
	// API Key: check POLYMARKET_API_KEY first, fall back to TEST_API_KEY_ID
	apiKey := getEnvWithFallback("POLYMARKET_API_KEY", "TEST_API_KEY_ID")
	if apiKey == "" {
		return nil, fmt.Errorf("POLYMARKET_API_KEY or TEST_API_KEY_ID environment variable is required")
	}

	// Private Key: check POLYMARKET_PRIVATE_KEY first, fall back to TEST_API_SECRET_KEY
	privateKeyB64 := getEnvWithFallback("POLYMARKET_PRIVATE_KEY", "TEST_API_SECRET_KEY")
	if privateKeyB64 == "" {
		return nil, fmt.Errorf("POLYMARKET_PRIVATE_KEY or TEST_API_SECRET_KEY environment variable is required")
	}

	// Decode the base64-encoded private key
	// Doc: api/authentication.mdx - "base64-encoded Ed25519 private key"
	privateKeyBytes, err := base64.StdEncoding.DecodeString(privateKeyB64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key: %w", err)
	}

	// Ed25519 private keys are 64 bytes (32 byte seed + 32 byte public key)
	// or 32 bytes (seed only). Handle both cases.
	var privateKey ed25519.PrivateKey
	switch len(privateKeyBytes) {
	case ed25519.PrivateKeySize: // 64 bytes
		privateKey = ed25519.PrivateKey(privateKeyBytes)
	case ed25519.SeedSize: // 32 bytes
		privateKey = ed25519.NewKeyFromSeed(privateKeyBytes)
	default:
		return nil, fmt.Errorf("invalid private key length: expected %d or %d bytes, got %d",
			ed25519.PrivateKeySize, ed25519.SeedSize, len(privateKeyBytes))
	}

	// Symbol: check POLYMARKET_SYMBOL first, fall back to TEST_MARKET_SLUG
	symbol := getEnvWithFallback("POLYMARKET_SYMBOL", "TEST_MARKET_SLUG")
	if symbol == "" {
		return nil, fmt.Errorf("POLYMARKET_SYMBOL or TEST_MARKET_SLUG environment variable is required")
	}

	// Base URL: check POLYMARKET_BASE_URL first, fall back to RETAIL_API_URL
	baseURL := getEnvWithFallback("POLYMARKET_BASE_URL", "RETAIL_API_URL")
	if baseURL == "" {
		baseURL = "https://api.polymarket.us"
	}

	// WebSocket URL: check POLYMARKET_WS_URL first, fall back to RETAIL_WS_URL
	// Doc: api-reference/websocket/overview.mdx - endpoints
	wsBaseURL := getEnvWithFallback("POLYMARKET_WS_URL", "RETAIL_WS_URL")
	if wsBaseURL == "" {
		// Derive from base URL by replacing https with wss
		wsBaseURL = baseURL
		if len(wsBaseURL) > 5 && wsBaseURL[:5] == "https" {
			wsBaseURL = "wss" + wsBaseURL[5:]
		} else if len(wsBaseURL) > 4 && wsBaseURL[:4] == "http" {
			wsBaseURL = "ws" + wsBaseURL[4:]
		}
	}

	// Check if TLS verification should be skipped (for staging with self-signed certs)
	insecureSkipVerify := getEnvWithFallback("INSECURE_SKIP_VERIFY") == "true"

	return &Config{
		APIKey:             apiKey,
		PrivateKey:         privateKey,
		Symbol:             symbol,
		BaseURL:            baseURL,
		WSPrivateURL:       wsBaseURL + "/v1/ws/private",
		WSMarketsURL:       wsBaseURL + "/v1/ws/markets",
		InsecureSkipVerify: insecureSkipVerify,
	}, nil
}

// MustLoad loads configuration or panics on error.
// Use this in main() for cleaner error handling.
func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}
	return cfg
}
