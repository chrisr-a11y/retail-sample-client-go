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
}

// Load loads configuration from environment variables.
// Returns an error if required variables are missing or invalid.
func Load() (*Config, error) {
	apiKey := os.Getenv("POLYMARKET_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("POLYMARKET_API_KEY environment variable is required")
	}

	privateKeyB64 := os.Getenv("POLYMARKET_PRIVATE_KEY")
	if privateKeyB64 == "" {
		return nil, fmt.Errorf("POLYMARKET_PRIVATE_KEY environment variable is required")
	}

	// Decode the base64-encoded private key
	// Doc: api/authentication.mdx - "base64-encoded Ed25519 private key"
	privateKeyBytes, err := base64.StdEncoding.DecodeString(privateKeyB64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode POLYMARKET_PRIVATE_KEY: %w", err)
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

	symbol := os.Getenv("POLYMARKET_SYMBOL")
	if symbol == "" {
		return nil, fmt.Errorf("POLYMARKET_SYMBOL environment variable is required")
	}

	baseURL := os.Getenv("POLYMARKET_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.polymarket.us"
	}

	// WebSocket URLs derived from base URL
	// Doc: api-reference/websocket/overview.mdx - endpoints
	wsBaseURL := "wss://api.polymarket.us"

	return &Config{
		APIKey:       apiKey,
		PrivateKey:   privateKey,
		Symbol:       symbol,
		BaseURL:      baseURL,
		WSPrivateURL: wsBaseURL + "/v1/ws/private",
		WSMarketsURL: wsBaseURL + "/v1/ws/markets",
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
