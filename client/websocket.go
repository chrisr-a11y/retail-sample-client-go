// Package client provides WebSocket client for the Polymarket API.
// Doc: api-reference/websocket/overview.mdx, api-reference/websocket/private.mdx,
//
//	api-reference/websocket/markets.mdx
package client

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/polymarket/retail-sample-client-go/auth"
	"github.com/polymarket/retail-sample-client-go/config"
	"github.com/polymarket/retail-sample-client-go/models"
)

// WSClient is a WebSocket client for real-time data.
// Doc: api-reference/websocket/overview.mdx
type WSClient struct {
	config       *config.Config
	privateConn  *websocket.Conn
	marketsConn  *websocket.Conn
	privateURL   string
	marketsURL   string
	mu           sync.Mutex
	done         chan struct{}
	messages     chan *models.WSMessage
	requestID    int
	connected    bool
	reconnecting bool
}

// NewWSClient creates a new WebSocket client.
func NewWSClient(cfg *config.Config) *WSClient {
	return &WSClient{
		config:     cfg,
		privateURL: cfg.WSPrivateURL,
		marketsURL: cfg.WSMarketsURL,
		done:       make(chan struct{}),
		messages:   make(chan *models.WSMessage, 100),
	}
}

// Connect establishes WebSocket connections.
// Doc: api-reference/websocket/overview.mdx - Connection
func (c *WSClient) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Connect to private WebSocket
	// Doc: api-reference/websocket/private.mdx - Endpoint
	privateHeaders := auth.GenerateWSHeaders(c.config)
	privateDialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	privateConn, _, err := privateDialer.Dial(c.privateURL, privateHeaders)
	if err != nil {
		return fmt.Errorf("failed to connect to private WebSocket: %w", err)
	}
	c.privateConn = privateConn
	log.Printf("[WS] Connected to private WebSocket: %s", c.privateURL)

	// Connect to markets WebSocket
	// Doc: api-reference/websocket/markets.mdx - Endpoint
	marketsHeaders := auth.GenerateWSMarketsHeaders(c.config)
	marketsDialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	marketsConn, _, err := marketsDialer.Dial(c.marketsURL, marketsHeaders)
	if err != nil {
		c.privateConn.Close()
		return fmt.Errorf("failed to connect to markets WebSocket: %w", err)
	}
	c.marketsConn = marketsConn
	log.Printf("[WS] Connected to markets WebSocket: %s", c.marketsURL)

	c.connected = true

	// Start reading from both connections
	go c.readPrivate()
	go c.readMarkets()

	return nil
}

// Close closes WebSocket connections.
func (c *WSClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	close(c.done)

	var errs []error
	if c.privateConn != nil {
		if err := c.privateConn.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if c.marketsConn != nil {
		if err := c.marketsConn.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	c.connected = false

	if len(errs) > 0 {
		return fmt.Errorf("errors closing connections: %v", errs)
	}
	return nil
}

// Messages returns a channel for receiving WebSocket messages.
func (c *WSClient) Messages() <-chan *models.WSMessage {
	return c.messages
}

// nextRequestID generates a unique request ID.
func (c *WSClient) nextRequestID(prefix string) string {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.requestID++
	return fmt.Sprintf("%s-%d", prefix, c.requestID)
}

// readPrivate reads messages from the private WebSocket.
func (c *WSClient) readPrivate() {
	for {
		select {
		case <-c.done:
			return
		default:
			_, message, err := c.privateConn.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					log.Printf("[WS] Private connection closed normally")
					return
				}
				log.Printf("[WS] Error reading from private WebSocket: %v", err)
				return
			}

			var msg models.WSMessage
			if err := json.Unmarshal(message, &msg); err != nil {
				log.Printf("[WS] Failed to parse private message: %v", err)
				continue
			}

			// Handle heartbeat
			// Doc: api-reference/websocket/overview.mdx - Heartbeats
			if msg.Heartbeat != nil {
				log.Printf("[WS] Private heartbeat received")
				continue
			}

			// Send to channel
			select {
			case c.messages <- &msg:
			default:
				log.Printf("[WS] Message channel full, dropping message")
			}
		}
	}
}

// readMarkets reads messages from the markets WebSocket.
func (c *WSClient) readMarkets() {
	for {
		select {
		case <-c.done:
			return
		default:
			_, message, err := c.marketsConn.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					log.Printf("[WS] Markets connection closed normally")
					return
				}
				log.Printf("[WS] Error reading from markets WebSocket: %v", err)
				return
			}

			var msg models.WSMessage
			if err := json.Unmarshal(message, &msg); err != nil {
				log.Printf("[WS] Failed to parse markets message: %v", err)
				continue
			}

			// Handle heartbeat
			if msg.Heartbeat != nil {
				log.Printf("[WS] Markets heartbeat received")
				continue
			}

			// Send to channel
			select {
			case c.messages <- &msg:
			default:
				log.Printf("[WS] Message channel full, dropping message")
			}
		}
	}
}

// sendPrivate sends a message on the private WebSocket.
func (c *WSClient) sendPrivate(msg interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.privateConn == nil {
		return fmt.Errorf("private WebSocket not connected")
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	return c.privateConn.WriteMessage(websocket.TextMessage, data)
}

// sendMarkets sends a message on the markets WebSocket.
func (c *WSClient) sendMarkets(msg interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.marketsConn == nil {
		return fmt.Errorf("markets WebSocket not connected")
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	return c.marketsConn.WriteMessage(websocket.TextMessage, data)
}

// SubscribeOrders subscribes to order updates.
// Doc: api-reference/websocket/private.mdx - Order Subscriptions
func (c *WSClient) SubscribeOrders(marketSlugs []string) (string, error) {
	requestID := c.nextRequestID("order")

	// Doc: api-reference/websocket/private.mdx - Subscribe to Orders
	// "Leave marketSlugs empty to subscribe to all markets"
	msg := &models.WSSubscribeRequest{
		Subscribe: &models.WSSubscription{
			RequestID:        requestID,
			SubscriptionType: models.SubscriptionTypeOrder,
			MarketSlugs:      marketSlugs,
		},
	}

	if err := c.sendPrivate(msg); err != nil {
		return "", err
	}

	log.Printf("[WS] Subscribed to orders (requestId: %s, markets: %v)", requestID, marketSlugs)
	return requestID, nil
}

// SubscribePositions subscribes to position updates.
// Doc: api-reference/websocket/private.mdx - Position Subscriptions
func (c *WSClient) SubscribePositions(marketSlugs []string) (string, error) {
	requestID := c.nextRequestID("position")

	msg := &models.WSSubscribeRequest{
		Subscribe: &models.WSSubscription{
			RequestID:        requestID,
			SubscriptionType: models.SubscriptionTypePosition,
			MarketSlugs:      marketSlugs,
		},
	}

	if err := c.sendPrivate(msg); err != nil {
		return "", err
	}

	log.Printf("[WS] Subscribed to positions (requestId: %s, markets: %v)", requestID, marketSlugs)
	return requestID, nil
}

// SubscribeBalances subscribes to account balance updates.
// Doc: api-reference/websocket/private.mdx - Account Balance Subscriptions
func (c *WSClient) SubscribeBalances() (string, error) {
	requestID := c.nextRequestID("balance")

	msg := &models.WSSubscribeRequest{
		Subscribe: &models.WSSubscription{
			RequestID:        requestID,
			SubscriptionType: models.SubscriptionTypeAccountBalance,
		},
	}

	if err := c.sendPrivate(msg); err != nil {
		return "", err
	}

	log.Printf("[WS] Subscribed to account balances (requestId: %s)", requestID)
	return requestID, nil
}

// SubscribeMarketData subscribes to full market data (order book).
// Doc: api-reference/websocket/markets.mdx - Market Data Subscription
func (c *WSClient) SubscribeMarketData(marketSlugs []string, debounced bool) (string, error) {
	requestID := c.nextRequestID("marketdata")

	// Doc: api-reference/websocket/markets.mdx - Debouncing
	msg := &models.WSSubscribeRequest{
		Subscribe: &models.WSSubscription{
			RequestID:          requestID,
			SubscriptionType:   models.SubscriptionTypeMarketData,
			MarketSlugs:        marketSlugs,
			ResponsesDebounced: debounced,
		},
	}

	if err := c.sendMarkets(msg); err != nil {
		return "", err
	}

	log.Printf("[WS] Subscribed to market data (requestId: %s, markets: %v, debounced: %t)",
		requestID, marketSlugs, debounced)
	return requestID, nil
}

// SubscribeMarketDataLite subscribes to lightweight price data.
// Doc: api-reference/websocket/markets.mdx - Market Data Lite Subscription
func (c *WSClient) SubscribeMarketDataLite(marketSlugs []string) (string, error) {
	requestID := c.nextRequestID("marketdatalite")

	msg := &models.WSSubscribeRequest{
		Subscribe: &models.WSSubscription{
			RequestID:        requestID,
			SubscriptionType: models.SubscriptionTypeMarketDataLite,
			MarketSlugs:      marketSlugs,
		},
	}

	if err := c.sendMarkets(msg); err != nil {
		return "", err
	}

	log.Printf("[WS] Subscribed to market data lite (requestId: %s, markets: %v)", requestID, marketSlugs)
	return requestID, nil
}

// SubscribeTrades subscribes to trade notifications.
// Doc: api-reference/websocket/markets.mdx - Trade Subscription
func (c *WSClient) SubscribeTrades(marketSlugs []string) (string, error) {
	requestID := c.nextRequestID("trade")

	msg := &models.WSSubscribeRequest{
		Subscribe: &models.WSSubscription{
			RequestID:        requestID,
			SubscriptionType: models.SubscriptionTypeTrade,
			MarketSlugs:      marketSlugs,
		},
	}

	if err := c.sendMarkets(msg); err != nil {
		return "", err
	}

	log.Printf("[WS] Subscribed to trades (requestId: %s, markets: %v)", requestID, marketSlugs)
	return requestID, nil
}

// Unsubscribe cancels a subscription.
// Doc: api-reference/websocket/overview.mdx - Unsubscribing
func (c *WSClient) Unsubscribe(requestID string, isPrivate bool) error {
	msg := &models.WSUnsubscribeRequest{
		Unsubscribe: &models.WSUnsubscription{
			RequestID: requestID,
		},
	}

	if isPrivate {
		return c.sendPrivate(msg)
	}
	return c.sendMarkets(msg)
}

// IsConnected returns whether the client is connected.
func (c *WSClient) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.connected
}
