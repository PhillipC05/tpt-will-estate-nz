// Package mbie provides a client for the MBIE (Ministry of Business,
// Innovation and Employment) business APIs at api.business.govt.nz.
//
// Register for an API key at: https://portal.api.business.govt.nz/
package mbie

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	defaultBaseURL = "https://api.business.govt.nz"
	defaultTimeout = 15 * time.Second
)

// Client is the MBIE API client.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewClient creates an MBIE API client.
// apiKey is obtained from https://portal.api.business.govt.nz/
func NewClient(apiKey string) *Client {
	return &Client{
		baseURL: defaultBaseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

// WithBaseURL overrides the base URL (useful for testing with a mock server).
func (c *Client) WithBaseURL(url string) *Client {
	c.baseURL = url
	return c
}

func (c *Client) get(ctx context.Context, path string, out interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("mbie: build request: %w", err)
	}
	req.Header.Set("Authorization", "ApiKey "+c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("mbie: request %s: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("mbie: %s returned HTTP %d", path, resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("mbie: decode response from %s: %w", path, err)
	}
	return nil
}

// ErrNotFound is returned when the MBIE API returns HTTP 404.
var ErrNotFound = fmt.Errorf("mbie: not found")
