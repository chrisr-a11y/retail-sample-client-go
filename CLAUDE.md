# Polymarket Retail API Go Client

## Purpose

This is a comprehensive Go client example for the Polymarket Retail API. It demonstrates all major API capabilities including:

- WebSocket subscriptions for real-time data
- Order management (place, query, cancel)
- Account balance and position queries
- Market data streaming

## Adjacent Documentation

The API documentation is located at `../website-isv-docs/`:

| Doc File | Description |
|----------|-------------|
| `api/authentication.mdx` | Ed25519 signature authentication |
| `api-reference/websocket/overview.mdx` | WebSocket connection details |
| `api-reference/websocket/private.mdx` | Private subscriptions (orders, positions, balances) |
| `api-reference/websocket/markets.mdx` | Market data subscriptions |
| `api-reference/orders/overview.mdx` | Order management endpoints |
| `api-reference/account/overview.mdx` | Account balance endpoint |
| `api-reference/portfolio/overview.mdx` | Positions and activities |
| `api-reference/market/overview.mdx` | Market query endpoints |
| `api-reference/oapi-schemas/orders-schema.json` | Complete order API OpenAPI schema |

## Environment Setup

Set the following environment variables before running:

```bash
# Required
export POLYMARKET_API_KEY="your-api-key-uuid"
export POLYMARKET_PRIVATE_KEY="base64-encoded-ed25519-private-key"
export POLYMARKET_SYMBOL="market-slug-to-trade"

# Optional
export POLYMARKET_BASE_URL="https://api.polymarket.us"  # default
```

## How to Run

```bash
# Install dependencies
go mod download

# Run the demo
go run main.go
```

## What the Demo Does

1. Loads configuration from environment variables
2. Connects to WebSocket endpoints (private + markets)
3. Subscribes to: orders, positions, balances, market data, trades
4. Fetches available markets and account balance
5. Places a limit order far from market ($0.01) to ensure it rests on book
6. Waits for order confirmation via WebSocket
7. Cancels the order
8. Displays all market data observed during the session
9. Cleans up WebSocket connections

## Code Organization

```
├── main.go           # Demo orchestrator
├── config/
│   └── config.go     # Environment configuration loader
├── auth/
│   └── auth.go       # Ed25519 signature authentication
├── client/
│   ├── rest.go       # REST API client
│   └── websocket.go  # WebSocket client
└── models/
    └── types.go      # API types and models
```

## API Reference Comments

All code includes documentation references in comments:

```go
// CreateOrder creates a new order
// Doc: api-reference/orders/overview.mdx - POST /v1/orders
// Schema: api-reference/oapi-schemas/orders-schema.json - CreateOrderRequest
```

## Goal

**Primary purpose**: This project tests whether the API documentation enables building a correct implementation from docs alone.

This client serves as:
1. A validation tool to verify the accuracy and completeness of API documentation
2. A working example for developers integrating with the Polymarket Retail API
3. A test harness to exercise and validate API functionality

## Documentation Accuracy Findings

During implementation, we discovered the following about documentation quality:

### OpenAPI Schemas (Accurate ✓)
The JSON schemas at `api-reference/oapi-schemas/*.json` are authoritative and accurate:
- `portfolio-schema.json` correctly documents `positions` as a map (not array)
- `Activity` type correctly shows nested `trade`, `positionResolution`, `accountBalanceChange` objects
- `market-schema.json` correctly shows `line` as number and `outcomeTeamA/B` as integer

### Authentication Docs (Minor Ambiguity)
`api/authentication.mdx` - Signature format is correct but could be clearer:
- States: `{timestamp}{HTTP_METHOD}{URL_PATH}`
- Example only shows path without query params: `/v1/portfolio/positions`
- **Clarification needed**: Explicitly state that URL_PATH does NOT include query parameters

### Recommendation
When implementing, prioritize OpenAPI JSON schemas over narrative `.mdx` docs. The schemas are the source of truth for data structures.
