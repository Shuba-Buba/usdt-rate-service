package grinex

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type OrderBookEntry struct {
	Price  string `json:"price"`
	Volume string `json:"volume"`
	Amount string `json:"amount"`
	Factor string `json:"factor"`
	Type   string `json:"type"`
}

type DepthResponse struct {
	Timestamp int64            `json:"timestamp"`
	Asks      []OrderBookEntry `json:"asks"`
	Bids      []OrderBookEntry `json:"bids"`
}

type Client struct {
	baseURL    string
	httpClient *http.Client
	logger     *zap.Logger
}

func NewClient(baseURL string, logger *zap.Logger) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

func (c *Client) GetDepth(ctx context.Context) (*DepthResponse, error) {
	url := fmt.Sprintf("%s/api/v2/depth?market=usdtrub", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.Fatal("Failed to close connection", zap.Error(err))
		}
	}()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d and data: %s", resp.StatusCode, string(data))
	}

	var depthResp DepthResponse
	if err := json.Unmarshal(data, &depthResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &depthResp, nil
}
