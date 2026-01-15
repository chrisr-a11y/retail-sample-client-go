# Polymarket Retail API Go Client

A sample Go client demonstrating the Polymarket Retail API. This project validates that the API documentation enables building a correct implementation from docs alone.

## Quick Start

```bash
# Set environment variables
export POLYMARKET_API_KEY="your-api-key-id"
export POLYMARKET_PRIVATE_KEY="your-base64-ed25519-private-key"
export POLYMARKET_SYMBOL="market-slug-to-trade"

# Run
go run main.go
```

For staging environment:
```bash
./run-staging.sh
```

## Documentation vs Implementation Comparison

This section documents how the API documentation aligns with the actual API behavior, tested against the staging environment.

---

## REST API Endpoints

### Authentication

| Aspect | Documentation | Implementation | Status |
|--------|---------------|----------------|--------|
| Headers | `X-PM-Access-Key`, `X-PM-Timestamp`, `X-PM-Signature` | Same | ✅ Match |
| Signature Format | `{timestamp}{METHOD}{PATH}` | Same | ✅ Match |
| Query Params in Sig | Explicitly excluded | Same | ✅ Match |
| Timestamp | Milliseconds, ±5min window | Same | ✅ Match |

**File**: `auth/auth.go`

---

### Orders API

#### POST /v1/orders - Create Order

| Field | Documentation | Implementation | Status |
|-------|---------------|----------------|--------|
| `market_slug` | snake_case | snake_case | ✅ Match |
| `type` | integer (1=LIMIT, 2=MARKET) | integer | ✅ Match |
| `intent` | integer (1=BUY_YES, 2=SELL_YES, 3=BUY_NO, 4=SELL_NO) | integer | ✅ Match |
| `tif` | integer (1=GTC, 2=GTD, 3=IOC, 4=FOK) | integer | ✅ Match |
| `price` | Amount object | Amount object | ✅ Match |
| `quantity` | number | number | ✅ Match |

**File**: `client/rest.go:271` - `CreateOrder()`

#### GET /v1/orders/open - Get Open Orders

| Aspect | Documentation | Implementation | Status |
|--------|---------------|----------------|--------|
| Query param | `slugs` (comma-separated) | Same | ✅ Match |
| Response | Array of Order objects | Same | ✅ Match |

**File**: `client/rest.go:309` - `GetOpenOrders()`

#### GET /v1/order/{orderId} - Get Order

| Aspect | Documentation | Implementation | Status |
|--------|---------------|----------------|--------|
| Path param | `orderId` | Same | ✅ Match |
| Response | Order object | Same | ✅ Match |

**File**: `client/rest.go:333` - `GetOrder()`

#### POST /v1/order/{orderId}/cancel - Cancel Order

| Aspect | Documentation | Implementation | Status |
|--------|---------------|----------------|--------|
| Path param | `orderId` | Same | ✅ Match |
| Body | `marketSlug` (optional) | Same | ✅ Match |

**File**: `client/rest.go:352` - `CancelOrder()`

#### POST /v1/orders/open/cancel - Cancel All Orders

| Aspect | Documentation | Implementation | Status |
|--------|---------------|----------------|--------|
| Body | `slugs` array (optional) | Same | ✅ Match |
| Response | `canceledOrderIds` array | Same | ✅ Match |

**File**: `client/rest.go:366` - `CancelAllOpenOrders()`

---

### Portfolio API

#### GET /v1/portfolio/positions - Get Positions

| Aspect | Documentation | Implementation | Status |
|--------|---------------|----------------|--------|
| Response type | Map of slug → UserPosition | Same | ✅ Match |
| Pagination | cursor-based (nextCursor, eof) | Same | ✅ Match |
| UserPosition fields | All documented fields present | Same | ✅ Match |

**File**: `client/rest.go:196` - `GetPositions()`

#### GET /v1/portfolio/activities - Get Activities

| Aspect | Documentation | Implementation | Status |
|--------|---------------|----------------|--------|
| Response type | Array of Activity objects | Same | ✅ Match |
| Activity structure | Nested trade/positionResolution/accountBalanceChange | Same | ✅ Match |
| Pagination | cursor-based | Same | ✅ Match |

**File**: `client/rest.go:228` - `GetActivities()`

---

### Account API

#### GET /v1/account/balances - Get Balances

| Aspect | Documentation | Implementation | Status |
|--------|---------------|----------------|--------|
| Response | Array of Balance objects | Same | ✅ Match |
| Balance fields | currentBalance, buyingPower, openOrders, etc. | Same | ✅ Match |

**File**: `client/rest.go:177` - `GetBalances()`

---

### Market API

#### GET /v1/markets - Get Markets

| Aspect | Documentation | Implementation | Status |
|--------|---------------|----------------|--------|
| Query params | limit, active, etc. | Same | ✅ Match |
| Response | Array of Market objects | Same | ✅ Match |
| Market.line | number | *float64 | ✅ Match |
| Market.outcomeTeamA/B | integer | *int | ✅ Match |

**File**: `client/rest.go:107` - `GetMarkets()`

#### GET /v1/market/slug/{slug} - Get Market by Slug

| Aspect | Documentation | Implementation | Status |
|--------|---------------|----------------|--------|
| Path param | slug | Same | ✅ Match |
| Response | Market object | Same | ✅ Match |

**File**: `client/rest.go:138` - `GetMarketBySlug()`

---

## WebSocket API

### Authentication

| Aspect | Documentation | Implementation | Status |
|--------|---------------|----------------|--------|
| Headers | `X-PM-Access-Key`, `X-PM-Timestamp`, `X-PM-Signature` | Same | ✅ Match |
| Same as REST | Yes | Yes | ✅ Match |
| Passphrase | Not used | Not used | ✅ Match |

**File**: `auth/auth.go:55` - `GenerateWSHeaders()`

### Message Format

| Aspect | Documentation | Implementation | Status |
|--------|---------------|----------------|--------|
| Field names | snake_case | snake_case | ✅ Match |
| `request_id` | string | string | ✅ Match |
| `subscription_type` | integer | integer | ✅ Match |
| `market_slugs` | array | array | ✅ Match |

**File**: `models/types.go:371` - `WSSubscription`

### Private WebSocket Subscription Types

| Value | Documentation | Implementation | Status |
|-------|---------------|----------------|--------|
| 1 | ORDER | ORDER | ✅ Match |
| 3 | POSITION | POSITION | ✅ Match |
| 4 | ACCOUNT_BALANCE | ACCOUNT_BALANCE | ✅ Match |

**Note**: Type 2 is unused/reserved on the private WebSocket.

**File**: `models/types.go:547`

### Markets WebSocket Subscription Types

| Value | Documentation | Implementation | Status |
|-------|---------------|----------------|--------|
| 1 | MARKET_DATA | MARKET_DATA | ✅ Match |
| 2 | MARKET_DATA_LITE | MARKET_DATA_LITE | ✅ Match |
| 3 | TRADE | TRADE | ✅ Match |

**File**: `models/types.go:554`

---

## Summary

### Documentation Accuracy

| API | Status | Notes |
|-----|--------|-------|
| REST Authentication | ✅ Accurate | Query params excluded from signature |
| Orders API | ✅ Accurate | snake_case, integer enums |
| Portfolio API | ✅ Accurate | Map response, nested Activity objects |
| Account API | ✅ Accurate | All fields documented |
| Market API | ✅ Accurate | Field types correct |
| WebSocket Auth | ✅ Accurate | Same as REST |
| WebSocket Messages | ✅ Accurate | snake_case, integers |
| WS Subscription Types | ✅ Accurate | All subscription type numbers match |

All documentation is now accurate and matches the API implementation.

---

## Project Structure

```
retail-sample-client-go/
├── main.go              # Demo orchestrator
├── CLAUDE.md            # Project documentation
├── README.md            # This file
├── auth/
│   └── auth.go          # Ed25519 authentication
├── client/
│   ├── rest.go          # REST API client
│   └── websocket.go     # WebSocket client
├── config/
│   └── config.go        # Configuration
└── models/
    └── types.go         # API types
```

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `POLYMARKET_API_KEY` | Yes | API key ID (UUID) |
| `POLYMARKET_PRIVATE_KEY` | Yes | Base64-encoded Ed25519 private key |
| `POLYMARKET_SYMBOL` | Yes | Market slug to trade |
| `POLYMARKET_BASE_URL` | No | API base URL (default: https://api.polymarket.us) |
| `INSECURE_SKIP_VERIFY` | No | Skip TLS verification for staging |

## License

MIT
