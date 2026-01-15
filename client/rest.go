// Package client provides HTTP and WebSocket clients for the Polymarket API.
// Doc: api-reference/orders/overview.mdx, api-reference/account/overview.mdx,
//
//	api-reference/portfolio/overview.mdx, api-reference/market/overview.mdx
package client

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/polymarket/retail-sample-client-go/auth"
	"github.com/polymarket/retail-sample-client-go/config"
	"github.com/polymarket/retail-sample-client-go/models"
)

// RestClient is an HTTP client for the Polymarket REST API.
type RestClient struct {
	config     *config.Config
	httpClient *http.Client
}

// NewRestClient creates a new REST API client.
func NewRestClient(cfg *config.Config) *RestClient {
	transport := &http.Transport{}

	// Configure TLS for staging/development with self-signed certs
	if cfg.InsecureSkipVerify {
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	return &RestClient{
		config: cfg,
		httpClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		},
	}
}

// doRequest performs an authenticated HTTP request.
func (c *RestClient) doRequest(method, path string, body interface{}) ([]byte, error) {
	// Build URL
	reqURL := c.config.BaseURL + path

	// Prepare body if provided
	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	// Create request
	req, err := http.NewRequest(method, reqURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set content type for POST requests
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Sign the request
	// Doc: api/authentication.mdx - Required Headers
	if err := auth.SignRequest(req, c.config); err != nil {
		return nil, fmt.Errorf("failed to sign request: %w", err)
	}

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for errors
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// ========== Markets API ==========
// Doc: api-reference/market/overview.mdx

// GetMarkets retrieves a list of markets with optional filters.
// Doc: api-reference/market/overview.mdx - GET /v1/markets
func (c *RestClient) GetMarkets(limit int, active *bool) (*models.GetMarketsResponse, error) {
	// Build query parameters
	// Doc: api-reference/market/overview.mdx - Filtering Markets
	params := url.Values{}
	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}
	if active != nil {
		params.Set("active", fmt.Sprintf("%t", *active))
	}

	path := "/v1/markets"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	respBody, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result models.GetMarketsResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// GetMarketBySlug retrieves a market by its slug.
// Doc: api-reference/market/overview.mdx - GET /v1/market/slug/{slug}
func (c *RestClient) GetMarketBySlug(slug string) (*models.Market, error) {
	path := "/v1/market/slug/" + url.PathEscape(slug)

	respBody, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result models.Market
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// GetMarketSettlement retrieves settlement data for a resolved market.
// Doc: api-reference/market/overview.mdx - Settlement
func (c *RestClient) GetMarketSettlement(slug string) (*models.MarketSettlement, error) {
	path := "/v1/markets/" + url.PathEscape(slug) + "/settlement"

	respBody, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result models.MarketSettlement
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// ========== Account API ==========
// Doc: api-reference/account/overview.mdx

// GetBalances retrieves account balances.
// Doc: api-reference/account/overview.mdx - GET /v1/account/balances
func (c *RestClient) GetBalances() (*models.GetBalancesResponse, error) {
	respBody, err := c.doRequest("GET", "/v1/account/balances", nil)
	if err != nil {
		return nil, err
	}

	var result models.GetBalancesResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// ========== Portfolio API ==========
// Doc: api-reference/portfolio/overview.mdx

// GetPositions retrieves trading positions.
// Doc: api-reference/portfolio/overview.mdx - GET /v1/portfolio/positions
func (c *RestClient) GetPositions(market string, limit int, cursor string) (*models.GetPositionsResponse, error) {
	params := url.Values{}
	if market != "" {
		params.Set("market", market)
	}
	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}
	if cursor != "" {
		params.Set("cursor", cursor)
	}

	path := "/v1/portfolio/positions"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	respBody, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result models.GetPositionsResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// GetActivities retrieves trading activity history.
// Doc: api-reference/portfolio/overview.mdx - GET /v1/portfolio/activities
func (c *RestClient) GetActivities(marketSlug string, types []string, limit int, cursor string, sortOrder string) (*models.GetActivitiesResponse, error) {
	params := url.Values{}
	if marketSlug != "" {
		params.Set("marketSlug", marketSlug)
	}
	if len(types) > 0 {
		params.Set("types", strings.Join(types, ","))
	}
	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}
	if cursor != "" {
		params.Set("cursor", cursor)
	}
	if sortOrder != "" {
		params.Set("sortOrder", sortOrder)
	}

	path := "/v1/portfolio/activities"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	respBody, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result models.GetActivitiesResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// ========== Orders API ==========
// Doc: api-reference/orders/overview.mdx
// Schema: api-reference/oapi-schemas/orders-schema.json

// CreateOrder creates a new order.
// Doc: api-reference/orders/overview.mdx - POST /v1/orders
// Schema: api-reference/oapi-schemas/orders-schema.json - CreateOrderRequest
func (c *RestClient) CreateOrder(req *models.CreateOrderRequest) (*models.CreateOrderResponse, error) {
	respBody, err := c.doRequest("POST", "/v1/orders", req)
	if err != nil {
		return nil, err
	}

	var result models.CreateOrderResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// PreviewOrder previews an order before submission.
// Doc: api-reference/orders/overview.mdx - POST /v1/order/preview
// Schema: api-reference/oapi-schemas/orders-schema.json - PreviewOrderRequest
func (c *RestClient) PreviewOrder(req *models.CreateOrderRequest) (*models.PreviewOrderResponse, error) {
	previewReq := &models.PreviewOrderRequest{
		Request: req,
	}

	respBody, err := c.doRequest("POST", "/v1/order/preview", previewReq)
	if err != nil {
		return nil, err
	}

	var result models.PreviewOrderResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// GetOpenOrders retrieves all open orders.
// Doc: api-reference/orders/overview.mdx - GET /v1/orders/open
// Schema: api-reference/oapi-schemas/orders-schema.json - GetOpenOrdersResponse
func (c *RestClient) GetOpenOrders(slugs []string) (*models.GetOpenOrdersResponse, error) {
	path := "/v1/orders/open"
	if len(slugs) > 0 {
		params := url.Values{}
		params.Set("slugs", strings.Join(slugs, ","))
		path += "?" + params.Encode()
	}

	respBody, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result models.GetOpenOrdersResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// GetOrder retrieves a specific order by ID.
// Doc: api-reference/orders/overview.mdx - GET /v1/order/{orderId}
// Schema: api-reference/oapi-schemas/orders-schema.json - GetOrderResponse
func (c *RestClient) GetOrder(orderID string) (*models.GetOrderResponse, error) {
	path := "/v1/order/" + url.PathEscape(orderID)

	respBody, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result models.GetOrderResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// CancelOrder cancels a specific order.
// Doc: api-reference/orders/overview.mdx - POST /v1/order/{orderId}/cancel
// Schema: api-reference/oapi-schemas/orders-schema.json - CancelOrderRequest
func (c *RestClient) CancelOrder(orderID string, marketSlug string) error {
	path := "/v1/order/" + url.PathEscape(orderID) + "/cancel"

	req := &models.CancelOrderRequest{
		MarketSlug: marketSlug,
	}

	_, err := c.doRequest("POST", path, req)
	return err
}

// CancelAllOpenOrders cancels all open orders.
// Doc: api-reference/orders/overview.mdx - POST /v1/orders/open/cancel
// Schema: api-reference/oapi-schemas/orders-schema.json - CancelOpenOrdersRequest
func (c *RestClient) CancelAllOpenOrders(slugs []string) (*models.CancelOpenOrdersResponse, error) {
	req := &models.CancelOpenOrdersRequest{
		Slugs: slugs,
	}

	respBody, err := c.doRequest("POST", "/v1/orders/open/cancel", req)
	if err != nil {
		return nil, err
	}

	var result models.CancelOpenOrdersResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}
