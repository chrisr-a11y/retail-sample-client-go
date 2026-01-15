// Package auth provides Ed25519 signature-based authentication for the Polymarket API.
// Doc: api/authentication.mdx
package auth

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/crypto/ed25519"

	"github.com/polymarket/retail-sample-client-go/config"
)

// SignRequest signs an HTTP request with Ed25519 authentication headers.
// Doc: api/authentication.mdx - Required Headers
//
// Headers set:
//   - X-PM-Access-Key: API key ID (UUID)
//   - X-PM-Timestamp: Current Unix timestamp in milliseconds
//   - X-PM-Signature: Ed25519 signature (base64 encoded)
//
// Signature format: {timestamp}{HTTP_METHOD}{URL_PATH}
// Example: "1704067200000GET/v1/portfolio/positions"
func SignRequest(req *http.Request, cfg *config.Config) error {
	// Generate timestamp in milliseconds
	// Doc: api/authentication.mdx - "Current Unix timestamp in milliseconds"
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)

	// Build message to sign
	// Doc: api/authentication.mdx - Signature Format
	// Format: {timestamp}{HTTP_METHOD}{URL_PATH}
	// Note: URL_PATH does NOT include query parameters per the docs example
	method := req.Method
	path := req.URL.Path
	message := timestamp + method + path

	// Sign the message with Ed25519
	signature := ed25519.Sign(cfg.PrivateKey, []byte(message))
	signatureB64 := base64.StdEncoding.EncodeToString(signature)

	// Set authentication headers
	// Doc: api/authentication.mdx - Required Headers table
	req.Header.Set("X-PM-Access-Key", cfg.APIKey)
	req.Header.Set("X-PM-Timestamp", timestamp)
	req.Header.Set("X-PM-Signature", signatureB64)

	return nil
}

// GenerateWSHeaders generates authentication headers for WebSocket connections.
// Doc: api-reference/websocket/overview.mdx - Authentication
//
// WebSocket headers (different from REST):
//   - X-API-Key: API key
//   - X-API-Signature: Signature for the connection
//   - X-API-Timestamp: Current timestamp in milliseconds
//   - X-API-Passphrase: Derived from signing the API key
func GenerateWSHeaders(cfg *config.Config) http.Header {
	headers := make(http.Header)

	// Generate timestamp in milliseconds
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)

	// For WebSocket, sign the timestamp + GET + path
	// The path for WebSocket connection is typically just the endpoint
	message := timestamp + "GET" + "/v1/ws/private"
	signature := ed25519.Sign(cfg.PrivateKey, []byte(message))
	signatureB64 := base64.StdEncoding.EncodeToString(signature)

	// Generate passphrase by signing the API key
	// This derives the passphrase from the API key and private key
	passphraseMessage := cfg.APIKey
	passphraseSignature := ed25519.Sign(cfg.PrivateKey, []byte(passphraseMessage))
	passphraseB64 := base64.StdEncoding.EncodeToString(passphraseSignature)

	// Set WebSocket authentication headers
	// Doc: api-reference/websocket/overview.mdx - Authentication section
	headers.Set("X-API-Key", cfg.APIKey)
	headers.Set("X-API-Timestamp", timestamp)
	headers.Set("X-API-Signature", signatureB64)
	headers.Set("X-API-Passphrase", passphraseB64)

	return headers
}

// GenerateWSMarketsHeaders generates authentication headers for the markets WebSocket.
// Doc: api-reference/websocket/markets.mdx - endpoint
func GenerateWSMarketsHeaders(cfg *config.Config) http.Header {
	headers := make(http.Header)

	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)

	// Sign for markets endpoint
	message := timestamp + "GET" + "/v1/ws/markets"
	signature := ed25519.Sign(cfg.PrivateKey, []byte(message))
	signatureB64 := base64.StdEncoding.EncodeToString(signature)

	// Generate passphrase
	passphraseSignature := ed25519.Sign(cfg.PrivateKey, []byte(cfg.APIKey))
	passphraseB64 := base64.StdEncoding.EncodeToString(passphraseSignature)

	headers.Set("X-API-Key", cfg.APIKey)
	headers.Set("X-API-Timestamp", timestamp)
	headers.Set("X-API-Signature", signatureB64)
	headers.Set("X-API-Passphrase", passphraseB64)

	return headers
}

// ValidateTimestamp checks if a timestamp is within the allowed window.
// Doc: api/authentication.mdx - Timestamp Validation
// "Timestamps must be within ±5 minutes of server time"
func ValidateTimestamp(timestampMs int64) error {
	now := time.Now().UnixMilli()
	diff := now - timestampMs
	if diff < 0 {
		diff = -diff
	}

	// 5 minutes in milliseconds = 5 * 60 * 1000 = 300000
	maxDiff := int64(5 * 60 * 1000)
	if diff > maxDiff {
		return fmt.Errorf("timestamp outside valid window: difference of %d ms exceeds %d ms", diff, maxDiff)
	}

	return nil
}
