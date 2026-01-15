// Package models defines API types and structures for the Polymarket Retail API.
// Doc: api-reference/oapi-schemas/orders-schema.json - components/schemas
package models

// Amount represents a monetary amount with currency.
// Doc: api-reference/oapi-schemas/orders-schema.json - Amount schema
type Amount struct {
	Value    string `json:"value"`    // Decimal string e.g. "0.55"
	Currency string `json:"currency"` // Currency code e.g. "USD"
}

// OrderType defines the type of order (string in responses).
// Note: Responses use strings, requests use integers
type OrderType string

const (
	OrderTypeLimit  OrderType = "ORDER_TYPE_LIMIT"
	OrderTypeMarket OrderType = "ORDER_TYPE_MARKET"
)

// Request integer constants for order type
const (
	OrderTypeRequestLimit  = 1
	OrderTypeRequestMarket = 2
)

// OrderSide defines buy or sell side (string in responses).
// Note: Responses use strings like "ORDER_SIDE_BUY"
type OrderSide string

const (
	OrderSideBuy  OrderSide = "ORDER_SIDE_BUY"
	OrderSideSell OrderSide = "ORDER_SIDE_SELL"
)

// OrderIntent indicates position direction (string in responses).
// Note: Responses use strings, requests use integers
type OrderIntent string

const (
	OrderIntentBuyLong   OrderIntent = "ORDER_INTENT_BUY_LONG"
	OrderIntentSellLong  OrderIntent = "ORDER_INTENT_SELL_LONG"
	OrderIntentBuyShort  OrderIntent = "ORDER_INTENT_BUY_SHORT"
	OrderIntentSellShort OrderIntent = "ORDER_INTENT_SELL_SHORT"
)

// Request integer constants for order intent
const (
	OrderIntentRequestBuyYes  = 1 // Buy Yes shares
	OrderIntentRequestSellYes = 2 // Sell Yes shares
	OrderIntentRequestBuyNo   = 3 // Buy No shares
	OrderIntentRequestSellNo  = 4 // Sell No shares
)

// TimeInForce defines order duration (string in responses).
// Note: Responses use strings, requests use integers
type TimeInForce string

const (
	TIFGoodTillCancel    TimeInForce = "TIME_IN_FORCE_GOOD_TILL_CANCEL"
	TIFGoodTillDate      TimeInForce = "TIME_IN_FORCE_GOOD_TILL_DATE"
	TIFImmediateOrCancel TimeInForce = "TIME_IN_FORCE_IMMEDIATE_OR_CANCEL"
	TIFFillOrKill        TimeInForce = "TIME_IN_FORCE_FILL_OR_KILL"
)

// Request integer constants for time in force
const (
	TIFRequestGTC = 1 // Good Till Cancel
	TIFRequestGTD = 2 // Good Till Date
	TIFRequestIOC = 3 // Immediate Or Cancel
	TIFRequestFOK = 4 // Fill Or Kill
)

// OrderState represents the current state of an order.
// Doc: api-reference/orders/overview.mdx - Order States
type OrderState string

const (
	OrderStatePendingNew      OrderState = "ORDER_STATE_PENDING_NEW"
	OrderStatePartiallyFilled OrderState = "ORDER_STATE_PARTIALLY_FILLED"
	OrderStateFilled          OrderState = "ORDER_STATE_FILLED"
	OrderStateCanceled        OrderState = "ORDER_STATE_CANCELED"
	OrderStateRejected        OrderState = "ORDER_STATE_REJECTED"
	OrderStateExpired         OrderState = "ORDER_STATE_EXPIRED"
	OrderStatePendingCancel   OrderState = "ORDER_STATE_PENDING_CANCEL"
	OrderStatePendingReplace  OrderState = "ORDER_STATE_PENDING_REPLACE"
	OrderStatePendingRisk     OrderState = "ORDER_STATE_PENDING_RISK"
	OrderStateReplaced        OrderState = "ORDER_STATE_REPLACED"
)

// ExecutionType defines the type of execution.
// Doc: api-reference/websocket/private.mdx - Execution Types
type ExecutionType string

const (
	ExecutionTypePartialFill ExecutionType = "EXECUTION_TYPE_PARTIAL_FILL"
	ExecutionTypeFill        ExecutionType = "EXECUTION_TYPE_FILL"
	ExecutionTypeCanceled    ExecutionType = "EXECUTION_TYPE_CANCELED"
	ExecutionTypeRejected    ExecutionType = "EXECUTION_TYPE_REJECTED"
	ExecutionTypeExpired     ExecutionType = "EXECUTION_TYPE_EXPIRED"
	ExecutionTypeReplace     ExecutionType = "EXECUTION_TYPE_REPLACE"
	ExecutionTypeDoneForDay  ExecutionType = "EXECUTION_TYPE_DONE_FOR_DAY"
)

// MarketMetadata contains market information.
// Doc: api-reference/oapi-schemas/orders-schema.json - MarketMetadata
type MarketMetadata struct {
	Slug      string `json:"slug"`
	Icon      string `json:"icon,omitempty"`
	Title     string `json:"title,omitempty"`
	Outcome   string `json:"outcome,omitempty"`
	EventSlug string `json:"eventSlug,omitempty"`
}

// Order represents an order in the system.
// Doc: api-reference/oapi-schemas/orders-schema.json - Order schema
type Order struct {
	ID             string          `json:"id"`
	MarketSlug     string          `json:"marketSlug"`
	Side           OrderSide       `json:"side"`
	Type           OrderType       `json:"type"`
	Price          *Amount         `json:"price,omitempty"`
	Quantity       float64         `json:"quantity"`
	CumQuantity    float64         `json:"cumQuantity,omitempty"`
	LeavesQuantity float64         `json:"leavesQuantity,omitempty"`
	TIF            TimeInForce     `json:"tif,omitempty"`
	GoodTillTime   string          `json:"goodTillTime,omitempty"`
	Intent         OrderIntent     `json:"intent"`
	MarketMetadata *MarketMetadata `json:"marketMetadata,omitempty"`
	State          OrderState      `json:"state"`
	AvgPx          *Amount         `json:"avgPx,omitempty"`
	InsertTime     string          `json:"insertTime,omitempty"`
	CreateTime     string          `json:"createTime,omitempty"`
}

// CreateOrderRequest is the request to create a new order.
// Note: API uses snake_case field names and integer enums for type/intent/tif
type CreateOrderRequest struct {
	MarketSlug           string  `json:"market_slug"`
	Type                 int     `json:"type,omitempty"`     // 1=LIMIT, 2=MARKET
	Price                *Amount `json:"price,omitempty"`
	Quantity             float64 `json:"quantity,omitempty"`
	TIF                  int     `json:"tif,omitempty"`      // 1=GTC, 2=GTD, 3=IOC, 4=FOK
	GoodTillTime         string  `json:"good_till_time,omitempty"`
	Intent               int     `json:"intent"`             // 1=BUY_YES, 2=SELL_YES, 3=BUY_NO, 4=SELL_NO
	CashOrderQty         *Amount `json:"cash_order_qty,omitempty"`
	ParticipateDoNotInit bool    `json:"participate_dont_initiate,omitempty"`
	SynchronousExecution bool    `json:"synchronous_execution,omitempty"`
	MaxBlockTime         string  `json:"max_block_time,omitempty"`
	ManualOrderIndicator string  `json:"manual_order_indicator,omitempty"`
}

// Execution represents an order execution.
// Doc: api-reference/oapi-schemas/orders-schema.json - Execution schema
type Execution struct {
	ID                string        `json:"id"`
	Order             *Order        `json:"order,omitempty"`
	LastShares        string        `json:"lastShares,omitempty"`
	LastPx            *Amount       `json:"lastPx,omitempty"`
	Type              ExecutionType `json:"type"`
	Text              string        `json:"text,omitempty"`
	OrderRejectReason string        `json:"orderRejectReason,omitempty"`
	TransactTime      string        `json:"transactTime,omitempty"`
	TradeID           string        `json:"tradeId,omitempty"`
	Aggressor         bool          `json:"aggressor,omitempty"`
}

// CreateOrderResponse is the response from creating an order.
// Doc: api-reference/oapi-schemas/orders-schema.json - CreateOrderResponse
type CreateOrderResponse struct {
	ID         string      `json:"id"`
	Executions []Execution `json:"executions,omitempty"`
}

// GetOpenOrdersResponse is the response from getting open orders.
// Doc: api-reference/oapi-schemas/orders-schema.json - GetOpenOrdersResponse
type GetOpenOrdersResponse struct {
	Orders []Order `json:"orders"`
}

// GetOrderResponse is the response from getting a specific order.
// Doc: api-reference/oapi-schemas/orders-schema.json - GetOrderResponse
type GetOrderResponse struct {
	Order *Order `json:"order"`
}

// CancelOrderRequest is the request to cancel an order.
// Doc: api-reference/oapi-schemas/orders-schema.json - CancelOrderRequest
type CancelOrderRequest struct {
	MarketSlug string `json:"marketSlug,omitempty"`
}

// CancelOpenOrdersRequest cancels all open orders.
// Doc: api-reference/oapi-schemas/orders-schema.json - CancelOpenOrdersRequest
type CancelOpenOrdersRequest struct {
	Slugs []string `json:"slugs,omitempty"`
}

// CancelOpenOrdersResponse is the response from canceling open orders.
// Doc: api-reference/oapi-schemas/orders-schema.json - CancelOpenOrdersResponse
type CancelOpenOrdersResponse struct {
	CanceledOrderIDs []string `json:"canceledOrderIds"`
}

// PreviewOrderRequest previews an order before submission.
// Doc: api-reference/oapi-schemas/orders-schema.json - PreviewOrderRequest
type PreviewOrderRequest struct {
	Request *CreateOrderRequest `json:"request"`
}

// PreviewOrderResponse is the response from previewing an order.
// Doc: api-reference/oapi-schemas/orders-schema.json - PreviewOrderResponse
type PreviewOrderResponse struct {
	Order *Order `json:"order"`
}

// Balance represents account balance information.
// Doc: api-reference/account/overview.mdx - Balance Fields
type Balance struct {
	CurrentBalance    float64           `json:"currentBalance"`
	Currency          string            `json:"currency"`
	BuyingPower       float64           `json:"buyingPower"`
	AssetNotional     float64           `json:"assetNotional,omitempty"`
	AssetAvailable    float64           `json:"assetAvailable,omitempty"`
	PendingCredit     float64           `json:"pendingCredit,omitempty"`
	OpenOrders        float64           `json:"openOrders,omitempty"`
	UnsettledFunds    float64           `json:"unsettledFunds,omitempty"`
	MarginRequirement float64           `json:"marginRequirement,omitempty"`
	LastUpdated       string            `json:"lastUpdated,omitempty"`
	PendingWithdrawals []PendingWithdrawal `json:"pendingWithdrawals,omitempty"`
}

// PendingWithdrawal represents a pending withdrawal.
// Doc: api-reference/account/overview.mdx - Pending Withdrawals
type PendingWithdrawal struct {
	ID           string  `json:"id"`
	Balance      float64 `json:"balance"`
	Status       string  `json:"status"`
	CreationTime string  `json:"creationTime"`
}

// GetBalancesResponse is the response from getting account balances.
// Doc: api-reference/account/overview.mdx - Example Response
type GetBalancesResponse struct {
	Balances []Balance `json:"balances"`
}

// UserPosition represents a trading position.
// Doc: api-reference/portfolio/overview.mdx - Position Fields
// Schema: api-reference/oapi-schemas/portfolio-schema.json - UserPosition
type UserPosition struct {
	NetPosition    string          `json:"netPosition"`
	QtyBought      string          `json:"qtyBought,omitempty"`
	QtySold        string          `json:"qtySold,omitempty"`
	Cost           *Amount         `json:"cost,omitempty"`
	Realized       *Amount         `json:"realized,omitempty"`
	BodPosition    string          `json:"bodPosition,omitempty"`
	Expired        bool            `json:"expired,omitempty"`
	UpdateTime     string          `json:"updateTime,omitempty"`
	CashValue      *Amount         `json:"cashValue,omitempty"`
	QtyAvailable   string          `json:"qtyAvailable,omitempty"`
	MarketMetadata *MarketMetadata `json:"marketMetadata,omitempty"`
}

// GetPositionsResponse is the response from getting positions.
// Doc: api-reference/portfolio/overview.mdx - Pagination
// Schema: api-reference/oapi-schemas/portfolio-schema.json - GetUserPositionsResponse
// Note: positions is a map of market slug to UserPosition, not an array
type GetPositionsResponse struct {
	Positions          map[string]UserPosition `json:"positions"`
	NextCursor         string                  `json:"nextCursor,omitempty"`
	EOF                bool                    `json:"eof"`
	AvailablePositions []string                `json:"availablePositions,omitempty"`
}

// Activity represents a trading activity.
// Doc: api-reference/portfolio/overview.mdx - Activity Types
// Schema: api-reference/oapi-schemas/portfolio-schema.json - Activity
type Activity struct {
	Type               string               `json:"type"`
	Trade              *Trade               `json:"trade,omitempty"`
	PositionResolution *PositionResolution  `json:"positionResolution,omitempty"`
	AccountBalanceChange *AccountBalanceChange `json:"accountBalanceChange,omitempty"`
}

// Trade represents a trade execution.
// Schema: api-reference/oapi-schemas/portfolio-schema.json - Trade
type Trade struct {
	ID          string  `json:"id"`
	MarketSlug  string  `json:"marketSlug"`
	State       string  `json:"state"`
	CreateTime  string  `json:"createTime"`
	UpdateTime  string  `json:"updateTime,omitempty"`
	Price       *Amount `json:"price"`
	Qty         string  `json:"qty"`
	IsAggressor bool    `json:"isAggressor"`
	CostBasis   *Amount `json:"costBasis,omitempty"`
	RealizedPnl *Amount `json:"realizedPnl,omitempty"`
}

// PositionResolution represents a position resolution event.
// Schema: api-reference/oapi-schemas/portfolio-schema.json - PositionResolution
type PositionResolution struct {
	MarketSlug     string        `json:"marketSlug"`
	BeforePosition *UserPosition `json:"beforePosition,omitempty"`
	AfterPosition  *UserPosition `json:"afterPosition,omitempty"`
	UpdateTime     string        `json:"updateTime,omitempty"`
	TradeID        string        `json:"tradeId,omitempty"`
	Side           string        `json:"side,omitempty"`
}

// AccountBalanceChange represents a balance change event.
// Schema: api-reference/oapi-schemas/portfolio-schema.json - AccountBalanceChange
type AccountBalanceChange struct {
	TransactionID string  `json:"transactionId"`
	Status        string  `json:"status"`
	Amount        *Amount `json:"amount"`
	UpdateTime    string  `json:"updateTime,omitempty"`
	CreateTime    string  `json:"createTime,omitempty"`
}

// GetActivitiesResponse is the response from getting activities.
// Doc: api-reference/portfolio/overview.mdx - Pagination
// Schema: api-reference/oapi-schemas/portfolio-schema.json - GetActivitiesResponse
type GetActivitiesResponse struct {
	Activities []Activity `json:"activities"`
	NextCursor string     `json:"nextCursor,omitempty"`
	EOF        bool       `json:"eof"`
}

// Market represents market information.
// Doc: api-reference/market/overview.mdx - Key Market Fields
type Market struct {
	ID                 string  `json:"id"`
	Slug               string  `json:"slug"`
	Question           string  `json:"question"`
	Description        string  `json:"description,omitempty"`
	Category           string  `json:"category,omitempty"`
	Subcategory        string  `json:"subcategory,omitempty"`
	Active             bool    `json:"active"`
	Closed             bool    `json:"closed"`
	Archived           bool    `json:"archived"`
	LastTradePrice     float64 `json:"lastTradePrice,omitempty"`
	BestBid            float64 `json:"bestBid,omitempty"`
	BestAsk            float64 `json:"bestAsk,omitempty"`
	Spread             float64 `json:"spread,omitempty"`
	OneDayPriceChange  float64 `json:"oneDayPriceChange,omitempty"`
	OneWeekPriceChange float64 `json:"oneWeekPriceChange,omitempty"`
	Liquidity          string  `json:"liquidity,omitempty"`
	LiquidityNum       float64 `json:"liquidityNum,omitempty"`
	Volume             string  `json:"volume,omitempty"`
	VolumeNum          float64 `json:"volumeNum,omitempty"`
	Volume24hr         float64 `json:"volume24hr,omitempty"`
	Volume1wk          float64 `json:"volume1wk,omitempty"`
	Volume1mo          float64 `json:"volume1mo,omitempty"`
	// Sports market fields
	// Doc: api-reference/market/overview.mdx - Sports Market Fields
	// Schema: api-reference/oapi-schemas/market-schema.json - Market schema
	SportsMarketTypeV2 string   `json:"sportsMarketTypeV2,omitempty"`
	GameID             string   `json:"gameId,omitempty"`
	Line               *float64 `json:"line,omitempty"`       // number in schema
	PropType           string   `json:"propType,omitempty"`
	OutcomeTeamA       *int     `json:"outcomeTeamA,omitempty"` // integer in schema
	OutcomeTeamB       *int     `json:"outcomeTeamB,omitempty"` // integer in schema
}

// GetMarketsResponse is the response from listing markets.
// Doc: api-reference/market/overview.mdx - Pagination & Ordering
type GetMarketsResponse struct {
	Markets []Market `json:"markets"`
}

// GetMarketResponse is the response from getting a single market.
type GetMarketResponse struct {
	Market *Market `json:"market"`
}

// MarketSettlement represents market settlement data.
// Doc: api-reference/market/overview.mdx - Settlement
type MarketSettlement struct {
	Slug       string  `json:"slug"`
	Settlement float64 `json:"settlement"`
}

// ========== WebSocket Types ==========
// Note: API uses snake_case field names and integer subscription types

// WSSubscribeRequest is a WebSocket subscription request.
type WSSubscribeRequest struct {
	Subscribe *WSSubscription `json:"subscribe"`
}

// WSSubscription defines what to subscribe to.
// Note: API uses snake_case and integer subscription_type
type WSSubscription struct {
	RequestID          string   `json:"request_id"`
	SubscriptionType   int      `json:"subscription_type"`
	MarketSlugs        []string `json:"market_slugs,omitempty"`
	ResponsesDebounced bool     `json:"responses_debounced,omitempty"`
}

// WSUnsubscribeRequest unsubscribes from a stream.
type WSUnsubscribeRequest struct {
	Unsubscribe *WSUnsubscription `json:"unsubscribe"`
}

// WSUnsubscription identifies the subscription to cancel.
type WSUnsubscription struct {
	RequestID string `json:"request_id"`
}

// WSMessage is a generic WebSocket message.
type WSMessage struct {
	// Common fields
	RequestID        string `json:"requestId,omitempty"`
	SubscriptionType string `json:"subscriptionType,omitempty"`
	Error            string `json:"error,omitempty"`

	// Heartbeat
	// Doc: api-reference/websocket/overview.mdx - Heartbeats
	Heartbeat *struct{} `json:"heartbeat,omitempty"`

	// Order subscription responses
	// Doc: api-reference/websocket/private.mdx - Order Subscriptions
	OrderSubscriptionSnapshot *OrderSnapshot `json:"orderSubscriptionSnapshot,omitempty"`
	OrderSubscriptionUpdate   *OrderUpdate   `json:"orderSubscriptionUpdate,omitempty"`

	// Position subscription responses
	// Doc: api-reference/websocket/private.mdx - Position Subscriptions
	PositionSubscription *PositionUpdate `json:"positionSubscription,omitempty"`

	// Balance subscription responses
	// Doc: api-reference/websocket/private.mdx - Account Balance Subscriptions
	AccountBalancesSnapshot *BalanceSnapshot `json:"accountBalancesSnapshot,omitempty"`
	AccountBalancesUpdate   *BalanceUpdate   `json:"accountBalancesUpdate,omitempty"`

	// Market data responses
	// Doc: api-reference/websocket/markets.mdx - Market Data Subscription
	MarketData     *MarketDataUpdate     `json:"marketData,omitempty"`
	MarketDataLite *MarketDataLiteUpdate `json:"marketDataLite,omitempty"`
	Trade          *TradeUpdate          `json:"trade,omitempty"`
}

// OrderSnapshot is the initial snapshot of open orders.
// Doc: api-reference/websocket/private.mdx - Order Snapshot Response
type OrderSnapshot struct {
	Orders []Order `json:"orders"`
	EOF    bool    `json:"eof"`
}

// OrderUpdate is a real-time order execution update.
// Doc: api-reference/websocket/private.mdx - Order Update Response
type OrderUpdate struct {
	Execution *Execution `json:"execution"`
}

// PositionUpdate is a position change notification.
// Doc: api-reference/websocket/private.mdx - Position Update Response
type PositionUpdate struct {
	BeforePosition *UserPosition `json:"beforePosition,omitempty"`
	AfterPosition  *UserPosition `json:"afterPosition,omitempty"`
	UpdateTime     string        `json:"updateTime,omitempty"`
	EntryType      string        `json:"entryType,omitempty"`
	TradeID        string        `json:"tradeId,omitempty"`
}

// BalanceSnapshot is the initial balance snapshot.
// Doc: api-reference/websocket/private.mdx - Balance Snapshot Response
type BalanceSnapshot struct {
	Balances []Balance `json:"balances"`
}

// BalanceUpdate is a balance change notification.
// Doc: api-reference/websocket/private.mdx - Balance Update Response
type BalanceUpdate struct {
	BalanceChange *BalanceChange `json:"balanceChange"`
}

// BalanceChange represents a change in balance.
type BalanceChange struct {
	BeforeBalance *Balance `json:"beforeBalance,omitempty"`
	AfterBalance  *Balance `json:"afterBalance,omitempty"`
	Description   string   `json:"description,omitempty"`
	UpdateTime    string   `json:"updateTime,omitempty"`
	EntryType     string   `json:"entryType,omitempty"`
}

// PriceLevel represents a level in the order book.
// Doc: api-reference/websocket/markets.mdx - Order Book Depth
type PriceLevel struct {
	Px  *Amount `json:"px"`
	Qty string  `json:"qty"`
}

// MarketStats contains market statistics.
// Doc: api-reference/websocket/markets.mdx - Market Data Response
type MarketStats struct {
	LastTradePx   *Amount `json:"lastTradePx,omitempty"`
	SharesTraded  string  `json:"sharesTraded,omitempty"`
	OpenInterest  string  `json:"openInterest,omitempty"`
	HighPx        *Amount `json:"highPx,omitempty"`
	LowPx         *Amount `json:"lowPx,omitempty"`
}

// MarketDataUpdate is full order book and market stats.
// Doc: api-reference/websocket/markets.mdx - Market Data Response
type MarketDataUpdate struct {
	MarketSlug   string       `json:"marketSlug"`
	Bids         []PriceLevel `json:"bids,omitempty"`
	Offers       []PriceLevel `json:"offers,omitempty"`
	State        string       `json:"state,omitempty"`
	Stats        *MarketStats `json:"stats,omitempty"`
	TransactTime string       `json:"transactTime,omitempty"`
}

// MarketDataLiteUpdate is lightweight price data.
// Doc: api-reference/websocket/markets.mdx - Market Data Lite Response
type MarketDataLiteUpdate struct {
	MarketSlug   string  `json:"marketSlug"`
	CurrentPx    *Amount `json:"currentPx,omitempty"`
	LastTradePx  *Amount `json:"lastTradePx,omitempty"`
	BestBid      *Amount `json:"bestBid,omitempty"`
	BestAsk      *Amount `json:"bestAsk,omitempty"`
	BidDepth     int     `json:"bidDepth,omitempty"`
	AskDepth     int     `json:"askDepth,omitempty"`
	SharesTraded string  `json:"sharesTraded,omitempty"`
	OpenInterest string  `json:"openInterest,omitempty"`
}

// TradeUpdate is a real-time trade notification.
// Doc: api-reference/websocket/markets.mdx - Trade Response
type TradeUpdate struct {
	MarketSlug string    `json:"marketSlug"`
	Price      *Amount   `json:"price"`
	Quantity   *Amount   `json:"quantity"`
	TradeTime  string    `json:"tradeTime"`
	Maker      *TradeSide `json:"maker,omitempty"`
	Taker      *TradeSide `json:"taker,omitempty"`
}

// TradeSide represents one side of a trade.
type TradeSide struct {
	Side   OrderSide   `json:"side"`
	Intent OrderIntent `json:"intent"`
}

// Subscription type constants (integer values used by API).
// Private WebSocket subscription types:
const (
	SubscriptionTypeOrder          = 1 // Order drop copy
	SubscriptionTypeOrderSnapshot  = 2 // Order snapshot
	SubscriptionTypePosition       = 3 // Position updates
	SubscriptionTypeAccountBalance = 4 // Account balance updates
)

// Markets WebSocket subscription types:
const (
	SubscriptionTypeMarketData     = 1 // Full order book
	SubscriptionTypeMarketDataLite = 2 // Price summary only
	SubscriptionTypeTrade          = 3 // Trade feed
)

// Market state constants.
// Doc: api-reference/websocket/markets.mdx - Market States
const (
	MarketStateOpen       = "MARKET_STATE_OPEN"
	MarketStatePreopen    = "MARKET_STATE_PREOPEN"
	MarketStateSuspended  = "MARKET_STATE_SUSPENDED"
	MarketStateHalted     = "MARKET_STATE_HALTED"
	MarketStateExpired    = "MARKET_STATE_EXPIRED"
	MarketStateTerminated = "MARKET_STATE_TERMINATED"
)

// Ledger entry types.
// Doc: api-reference/websocket/private.mdx - Ledger Entry Types
const (
	LedgerEntryTypeOrderExecution = "LEDGER_ENTRY_TYPE_ORDER_EXECUTION"
	LedgerEntryTypeDeposit        = "LEDGER_ENTRY_TYPE_DEPOSIT"
	LedgerEntryTypeWithdrawal     = "LEDGER_ENTRY_TYPE_WITHDRAWAL"
	LedgerEntryTypeResolution     = "LEDGER_ENTRY_TYPE_RESOLUTION"
	LedgerEntryTypeCommission     = "LEDGER_ENTRY_TYPE_COMMISSION"
)
