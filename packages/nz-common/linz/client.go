// Package linz provides a client for the LINZ (Land Information New Zealand /
// Toitū Te Whenua) Data Service APIs.
//
// Register for an API key at: https://data.linz.govt.nz/
// Documentation: https://www.linz.govt.nz/guidance/data-service/linz-data-service-guide/web-services/lds-apis-and-web-services
package linz

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const (
	defaultBaseURL = "https://data.linz.govt.nz"
	defaultTimeout = 20 * time.Second
)

// Client is the LINZ Data Service client.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a LINZ Data Service client.
// apiKey is obtained from https://data.linz.govt.nz/
func NewClient(apiKey string) *Client {
	return &Client{
		baseURL: defaultBaseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

// Address represents a NZ address from the LINZ address dataset.
type Address struct {
	AddressID    int64   `json:"address_id"`
	FullAddress  string  `json:"full_address"`
	UnitValue    string  `json:"unit_value"`
	AddressNumber int    `json:"address_number"`
	RoadName     string  `json:"road_name"`
	RoadTypeName string  `json:"road_type_name"`
	SuburbLocality string `json:"suburb_locality"`
	TownCity     string  `json:"town_city"`
	Postcode     string  `json:"postcode"`
	Latitude     float64 `json:"shape_y"`
	Longitude    float64 `json:"shape_x"`
}

// AddressSearchResult holds LINZ address search results.
type AddressSearchResult struct {
	Features []struct {
		Properties Address `json:"properties"`
	} `json:"features"`
}

// SearchAddresses searches the NZ Physical Addresses dataset (layer 53353).
// q is a partial address string. Returns up to limit results.
func (c *Client) SearchAddresses(ctx context.Context, q string, limit int) ([]Address, error) {
	if q == "" {
		return nil, fmt.Errorf("linz: address search query cannot be empty")
	}
	if limit <= 0 {
		limit = 10
	}

	// LINZ WFS endpoint for NZ Physical Addresses
	params := url.Values{}
	params.Set("service", "WFS")
	params.Set("version", "2.0.0")
	params.Set("request", "GetFeature")
	params.Set("typeNames", "layer-53353")
	params.Set("outputFormat", "application/json")
	params.Set("count", fmt.Sprintf("%d", limit))
	params.Set("cql_filter", fmt.Sprintf("full_address ILIKE '%%%s%%'", q))

	path := "/services;key=" + c.apiKey + "/wfs?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("linz: build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("linz: request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("linz: address search returned HTTP %d", resp.StatusCode)
	}

	var result AddressSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("linz: decode response: %w", err)
	}

	addresses := make([]Address, 0, len(result.Features))
	for _, f := range result.Features {
		addresses = append(addresses, f.Properties)
	}
	return addresses, nil
}
