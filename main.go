// Package main demonstrates the Polymarket Retail API Go client.
// This demo exercises all major API capabilities:
// - WebSocket subscriptions for real-time data
// - Order management (place, query, cancel)
// - Account balance and position queries
// - Market data streaming
//
// See CLAUDE.md for documentation references and setup instructions.
// Adjacent docs: ../website-isv-docs/
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/polymarket/retail-sample-client-go/client"
	"github.com/polymarket/retail-sample-client-go/config"
	"github.com/polymarket/retail-sample-client-go/models"
)

func main() {
	log.SetFlags(log.Ltime | log.Lshortfile)
	log.Println("=== Polymarket Retail API Go Client Demo ===")
	log.Println("Doc: See CLAUDE.md for documentation references")

	// 1. Load configuration from environment
	// Doc: api/authentication.mdx - API key configuration
	log.Println("\n[STEP 1] Loading configuration...")
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	log.Printf("  API Key: %s...", cfg.APIKey[:8])
	log.Printf("  Symbol: %s (configurable via POLYMARKET_SYMBOL)", cfg.Symbol)
	log.Printf("  Base URL: %s", cfg.BaseURL)

	// 2. Initialize clients
	log.Println("\n[STEP 2] Initializing clients...")
	restClient := client.NewRestClient(cfg)
	wsClient := client.NewWSClient(cfg)
	log.Println("  REST client initialized")
	log.Println("  WebSocket client initialized")

	// Track market data received for summary
	var marketDataReceived []string
	var marketDataMu sync.Mutex

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("\n[SHUTDOWN] Received interrupt signal, cleaning up...")
		cancel()
	}()

	// 3. Fetch available markets
	// Doc: api-reference/market/overview.mdx - GET /v1/markets
	log.Println("\n[STEP 3] Fetching available markets...")
	activeTrue := true
	markets, err := restClient.GetMarkets(10, &activeTrue)
	if err != nil {
		log.Printf("  Warning: Failed to fetch markets: %v", err)
	} else {
		log.Printf("  Found %d active markets (showing first 10)", len(markets.Markets))
		for i, m := range markets.Markets {
			if i >= 5 {
				log.Printf("    ... and %d more", len(markets.Markets)-5)
				break
			}
			log.Printf("    - %s: %s (bid: %.2f, ask: %.2f)",
				m.Slug, truncate(m.Question, 40), m.BestBid, m.BestAsk)
		}
	}

	// 4. Get configured market details
	// Doc: api-reference/market/overview.mdx - GET /v1/market/slug/{slug}
	log.Printf("\n[STEP 4] Getting market details for '%s'...", cfg.Symbol)
	market, err := restClient.GetMarketBySlug(cfg.Symbol)
	if err != nil {
		log.Printf("  Warning: Failed to get market: %v", err)
		log.Printf("  (This is expected if the symbol doesn't exist)")
	} else {
		log.Printf("  Question: %s", market.Question)
		log.Printf("  Active: %t, Closed: %t", market.Active, market.Closed)
		log.Printf("  Best Bid: %.2f, Best Ask: %.2f", market.BestBid, market.BestAsk)
		log.Printf("  Volume 24h: %.2f", market.Volume24hr)
	}

	// 5. Check initial account balance
	// Doc: api-reference/account/overview.mdx - GET /v1/account/balances
	log.Println("\n[STEP 5] Checking account balance...")
	balances, err := restClient.GetBalances()
	if err != nil {
		log.Printf("  Warning: Failed to get balances: %v", err)
	} else {
		for _, b := range balances.Balances {
			log.Printf("  %s Balance: $%.2f", b.Currency, b.CurrentBalance)
			log.Printf("  Buying Power: $%.2f", b.BuyingPower)
			log.Printf("  Open Orders: $%.2f", b.OpenOrders)
		}
	}

	// 6. Get current positions
	// Doc: api-reference/portfolio/overview.mdx - GET /v1/portfolio/positions
	// Note: positions is a map of market slug to UserPosition
	log.Println("\n[STEP 6] Getting current positions...")
	positions, err := restClient.GetPositions("", 10, "")
	if err != nil {
		log.Printf("  Warning: Failed to get positions: %v", err)
	} else {
		if len(positions.Positions) == 0 {
			log.Println("  No open positions")
		} else {
			log.Printf("  Found %d positions:", len(positions.Positions))
			for slug, p := range positions.Positions {
				log.Printf("    - %s: qty=%s", slug, p.NetPosition)
			}
		}
	}

	// 7. Connect to WebSocket
	// Doc: api-reference/websocket/overview.mdx - Connection
	log.Println("\n[STEP 7] Connecting to WebSocket...")
	if err := wsClient.Connect(); err != nil {
		log.Printf("  Warning: Failed to connect WebSocket: %v", err)
		log.Println("  (Continuing without real-time updates)")
	} else {
		log.Println("  WebSocket connected successfully")

		// Start message handler
		go handleWSMessages(ctx, wsClient, &marketDataReceived, &marketDataMu)

		// 8. Subscribe to private streams
		// Doc: api-reference/websocket/private.mdx - Subscription Types
		log.Println("\n[STEP 8] Subscribing to private streams...")

		// Subscribe to orders for our symbol
		// Doc: api-reference/websocket/private.mdx - Order Subscriptions
		if _, err := wsClient.SubscribeOrders([]string{cfg.Symbol}); err != nil {
			log.Printf("  Warning: Failed to subscribe to orders: %v", err)
		}

		// Subscribe to positions
		// Doc: api-reference/websocket/private.mdx - Position Subscriptions
		if _, err := wsClient.SubscribePositions([]string{cfg.Symbol}); err != nil {
			log.Printf("  Warning: Failed to subscribe to positions: %v", err)
		}

		// Subscribe to balances
		// Doc: api-reference/websocket/private.mdx - Account Balance Subscriptions
		if _, err := wsClient.SubscribeBalances(); err != nil {
			log.Printf("  Warning: Failed to subscribe to balances: %v", err)
		}

		// 9. Subscribe to market streams
		// Doc: api-reference/websocket/markets.mdx - Subscription Types
		log.Println("\n[STEP 9] Subscribing to market streams...")

		// Subscribe to market data (order book)
		// Doc: api-reference/websocket/markets.mdx - Market Data Subscription
		if _, err := wsClient.SubscribeMarketData([]string{cfg.Symbol}, true); err != nil {
			log.Printf("  Warning: Failed to subscribe to market data: %v", err)
		}

		// Subscribe to trades
		// Doc: api-reference/websocket/markets.mdx - Trade Subscription
		if _, err := wsClient.SubscribeTrades([]string{cfg.Symbol}); err != nil {
			log.Printf("  Warning: Failed to subscribe to trades: %v", err)
		}

		// Give subscriptions time to initialize
		time.Sleep(2 * time.Second)
	}

	// 10. Place a limit order far from market
	// Doc: api-reference/orders/overview.mdx - POST /v1/orders
	// Schema: api-reference/oapi-schemas/orders-schema.json - CreateOrderRequest
	log.Println("\n[STEP 10] Placing limit order...")
	log.Printf("  Symbol: %s", cfg.Symbol)
	log.Printf("  Price: $0.01 (far from market to ensure it rests on book)")
	log.Printf("  Quantity: 10 shares")
	log.Printf("  Intent: ORDER_INTENT_BUY_LONG")

	orderReq := &models.CreateOrderRequest{
		MarketSlug: cfg.Symbol,
		Intent:     models.OrderIntentRequestBuyYes,  // 1 = Buy Yes
		Type:       models.OrderTypeRequestLimit,    // 1 = Limit
		Price: &models.Amount{
			Value:    "0.01",
			Currency: "USD",
		},
		Quantity: 10,
		TIF:      models.TIFRequestGTC, // 1 = Good Till Cancel
	}

	var orderID string
	orderResp, err := restClient.CreateOrder(orderReq)
	if err != nil {
		log.Printf("  Warning: Failed to create order: %v", err)
		log.Println("  (This may be due to invalid credentials or market)")
	} else {
		orderID = orderResp.ID
		log.Printf("  Order created successfully!")
		log.Printf("  Order ID: %s", orderID)
		if len(orderResp.Executions) > 0 {
			for _, exec := range orderResp.Executions {
				log.Printf("  Execution: %s - %s", exec.Type, exec.ID)
			}
		}

		// 11. Wait for order confirmation via WebSocket
		log.Println("\n[STEP 11] Waiting for WebSocket order confirmation...")
		time.Sleep(3 * time.Second)

		// 12. Get order details via REST
		// Doc: api-reference/orders/overview.mdx - GET /v1/order/{orderId}
		log.Println("\n[STEP 12] Getting order details...")
		orderDetail, err := restClient.GetOrder(orderID)
		if err != nil {
			log.Printf("  Warning: Failed to get order: %v", err)
		} else if orderDetail.Order != nil {
			o := orderDetail.Order
			log.Printf("  Order ID: %s", o.ID)
			log.Printf("  State: %s", o.State)
			log.Printf("  Side: %s, Type: %s", o.Side, o.Type)
			if o.Price != nil {
				log.Printf("  Price: %s %s", o.Price.Value, o.Price.Currency)
			}
			log.Printf("  Quantity: %.0f, Filled: %.0f, Remaining: %.0f",
				o.Quantity, o.CumQuantity, o.LeavesQuantity)
		}

		// 13. Get open orders
		// Doc: api-reference/orders/overview.mdx - GET /v1/orders/open
		log.Println("\n[STEP 13] Getting open orders...")
		openOrders, err := restClient.GetOpenOrders(nil)
		if err != nil {
			log.Printf("  Warning: Failed to get open orders: %v", err)
		} else {
			log.Printf("  Found %d open order(s)", len(openOrders.Orders))
			for _, o := range openOrders.Orders {
				log.Printf("    - %s: %s %s @ %s (qty: %.0f)",
					o.ID, o.Side, o.Intent, o.Price.Value, o.Quantity)
			}
		}

		// 14. Cancel the order
		// Doc: api-reference/orders/overview.mdx - POST /v1/order/{orderId}/cancel
		log.Println("\n[STEP 14] Canceling order...")
		if err := restClient.CancelOrder(orderID, cfg.Symbol); err != nil {
			log.Printf("  Warning: Failed to cancel order: %v", err)
		} else {
			log.Printf("  Order %s canceled successfully!", orderID)
		}

		// 15. Wait for cancel confirmation via WebSocket
		log.Println("\n[STEP 15] Waiting for WebSocket cancel confirmation...")
		time.Sleep(3 * time.Second)
	}

	// 16. Check final account balance
	// Doc: api-reference/account/overview.mdx - GET /v1/account/balances
	log.Println("\n[STEP 16] Checking final account balance...")
	finalBalances, err := restClient.GetBalances()
	if err != nil {
		log.Printf("  Warning: Failed to get balances: %v", err)
	} else {
		for _, b := range finalBalances.Balances {
			log.Printf("  %s Balance: $%.2f", b.Currency, b.CurrentBalance)
			log.Printf("  Buying Power: $%.2f", b.BuyingPower)
		}
	}

	// 17. Get activity history
	// Doc: api-reference/portfolio/overview.mdx - GET /v1/portfolio/activities
	// Schema: api-reference/oapi-schemas/portfolio-schema.json - Activity
	log.Println("\n[STEP 17] Getting recent activity...")
	activities, err := restClient.GetActivities("", nil, 5, "", "")
	if err != nil {
		log.Printf("  Warning: Failed to get activities: %v", err)
	} else {
		if len(activities.Activities) == 0 {
			log.Println("  No recent activity")
		} else {
			log.Printf("  Found %d activities:", len(activities.Activities))
			for _, a := range activities.Activities {
				detail := ""
				if a.Trade != nil {
					detail = fmt.Sprintf("market=%s, qty=%s", a.Trade.MarketSlug, a.Trade.Qty)
				} else if a.PositionResolution != nil {
					detail = fmt.Sprintf("market=%s", a.PositionResolution.MarketSlug)
				} else if a.AccountBalanceChange != nil {
					detail = fmt.Sprintf("txn=%s", a.AccountBalanceChange.TransactionID)
				}
				log.Printf("    - %s: %s", a.Type, detail)
			}
		}
	}

	// 18. Print market data summary
	log.Println("\n[STEP 18] Market data summary...")
	marketDataMu.Lock()
	if len(marketDataReceived) == 0 {
		log.Println("  No market data updates received")
	} else {
		log.Printf("  Received %d market data updates:", len(marketDataReceived))
		// Show last 10 updates
		start := 0
		if len(marketDataReceived) > 10 {
			start = len(marketDataReceived) - 10
		}
		for i := start; i < len(marketDataReceived); i++ {
			log.Printf("    %s", marketDataReceived[i])
		}
	}
	marketDataMu.Unlock()

	// 19. Clean shutdown
	log.Println("\n[STEP 19] Cleaning up...")
	if wsClient.IsConnected() {
		if err := wsClient.Close(); err != nil {
			log.Printf("  Warning: Error closing WebSocket: %v", err)
		} else {
			log.Println("  WebSocket connections closed")
		}
	}

	log.Println("\n=== Demo Complete ===")
	log.Println("See CLAUDE.md for documentation references")
}

// handleWSMessages processes WebSocket messages.
func handleWSMessages(ctx context.Context, wsClient *client.WSClient, marketData *[]string, mu *sync.Mutex) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-wsClient.Messages():
			if msg == nil {
				continue
			}

			// Handle errors
			if msg.Error != "" {
				log.Printf("[WS] Error: %s (requestId: %s)", msg.Error, msg.RequestID)
				continue
			}

			// Handle order snapshot
			// Doc: api-reference/websocket/private.mdx - Order Snapshot Response
			if msg.OrderSubscriptionSnapshot != nil {
				log.Printf("[WS] Order snapshot: %d orders", len(msg.OrderSubscriptionSnapshot.Orders))
				for _, o := range msg.OrderSubscriptionSnapshot.Orders {
					log.Printf("[WS]   - %s: %s %s @ %s", o.ID, o.State, o.Side, o.Price.Value)
				}
			}

			// Handle order update
			// Doc: api-reference/websocket/private.mdx - Order Update Response
			if msg.OrderSubscriptionUpdate != nil && msg.OrderSubscriptionUpdate.Execution != nil {
				exec := msg.OrderSubscriptionUpdate.Execution
				log.Printf("[WS] Order update: %s - %s", exec.Type, exec.ID)
				if exec.Order != nil {
					log.Printf("[WS]   Order state: %s", exec.Order.State)
				}
			}

			// Handle position update
			// Doc: api-reference/websocket/private.mdx - Position Update Response
			if msg.PositionSubscription != nil {
				log.Printf("[WS] Position update: entry=%s", msg.PositionSubscription.EntryType)
				if msg.PositionSubscription.AfterPosition != nil {
					log.Printf("[WS]   Net position: %s", msg.PositionSubscription.AfterPosition.NetPosition)
				}
			}

			// Handle balance snapshot
			// Doc: api-reference/websocket/private.mdx - Balance Snapshot Response
			if msg.AccountBalancesSnapshot != nil {
				log.Printf("[WS] Balance snapshot: %d balances", len(msg.AccountBalancesSnapshot.Balances))
				for _, b := range msg.AccountBalancesSnapshot.Balances {
					log.Printf("[WS]   %s: $%.2f (buying power: $%.2f)", b.Currency, b.CurrentBalance, b.BuyingPower)
				}
			}

			// Handle balance update
			// Doc: api-reference/websocket/private.mdx - Balance Update Response
			if msg.AccountBalancesUpdate != nil && msg.AccountBalancesUpdate.BalanceChange != nil {
				change := msg.AccountBalancesUpdate.BalanceChange
				log.Printf("[WS] Balance update: %s", change.Description)
				if change.AfterBalance != nil {
					log.Printf("[WS]   New balance: $%.2f", change.AfterBalance.CurrentBalance)
				}
			}

			// Handle market data
			// Doc: api-reference/websocket/markets.mdx - Market Data Response
			if msg.MarketData != nil {
				md := msg.MarketData
				summary := fmt.Sprintf("%s: %d bids, %d offers, state=%s",
					md.MarketSlug, len(md.Bids), len(md.Offers), md.State)
				log.Printf("[WS] Market data: %s", summary)

				mu.Lock()
				*marketData = append(*marketData, summary)
				mu.Unlock()

				// Print top of book
				if len(md.Bids) > 0 {
					log.Printf("[WS]   Best bid: %s @ %s", md.Bids[0].Qty, md.Bids[0].Px.Value)
				}
				if len(md.Offers) > 0 {
					log.Printf("[WS]   Best ask: %s @ %s", md.Offers[0].Qty, md.Offers[0].Px.Value)
				}
			}

			// Handle market data lite
			// Doc: api-reference/websocket/markets.mdx - Market Data Lite Response
			if msg.MarketDataLite != nil {
				mdl := msg.MarketDataLite
				summary := fmt.Sprintf("%s: bid=%s ask=%s", mdl.MarketSlug,
					safeAmountValue(mdl.BestBid), safeAmountValue(mdl.BestAsk))
				log.Printf("[WS] Market data lite: %s", summary)

				mu.Lock()
				*marketData = append(*marketData, summary)
				mu.Unlock()
			}

			// Handle trade
			// Doc: api-reference/websocket/markets.mdx - Trade Response
			if msg.Trade != nil {
				t := msg.Trade
				summary := fmt.Sprintf("%s: trade @ %s qty=%s at %s",
					t.MarketSlug, t.Price.Value, t.Quantity.Value, t.TradeTime)
				log.Printf("[WS] Trade: %s", summary)

				mu.Lock()
				*marketData = append(*marketData, summary)
				mu.Unlock()
			}
		}
	}
}

// truncate shortens a string if it exceeds maxLen.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// safeAmountValue safely extracts value from Amount pointer.
func safeAmountValue(a *models.Amount) string {
	if a == nil {
		return "N/A"
	}
	return a.Value
}
