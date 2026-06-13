package linz

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// PropertyBoundary represents a land parcel boundary from the NZ Parcels dataset.
type PropertyBoundary struct {
	ParcelID    int64   `json:"id"`
	Appellation string  `json:"appellation"`
	AreaSqm     float64 `json:"area"`
	LegalDesc   string  `json:"legal_description"`
	TitleNo     string  `json:"title_no"`
}

// PropertyBoundaryResult holds LINZ parcel boundary results.
type PropertyBoundaryResult struct {
	Features []struct {
		Properties PropertyBoundary `json:"properties"`
	} `json:"features"`
}

// GetAddressByID retrieves a single NZ address by its LINZ address_id.
func (c *Client) GetAddressByID(ctx context.Context, addressID int64) (*Address, error) {
	params := url.Values{}
	params.Set("service", "WFS")
	params.Set("version", "2.0.0")
	params.Set("request", "GetFeature")
	params.Set("typeNames", "layer-53353")
	params.Set("outputFormat", "application/json")
	params.Set("cql_filter", fmt.Sprintf("address_id=%d", addressID))

	path := "/services;key=" + c.apiKey + "/wfs?" + params.Encode()
	var result AddressSearchResult
	if err := c.wfsGet(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("linz: get address by ID %d: %w", addressID, err)
	}
	if len(result.Features) == 0 {
		return nil, fmt.Errorf("linz: address %d not found", addressID)
	}
	addr := result.Features[0].Properties
	return &addr, nil
}

// GetPropertyBoundaries returns the land parcel boundaries that contain the
// given WGS84 coordinate. Used to verify a submitter's address against a
// property boundary for resource consent eligibility.
func (c *Client) GetPropertyBoundaries(ctx context.Context, lat, lng float64) ([]PropertyBoundary, error) {
	filter := fmt.Sprintf("INTERSECTS(shape,POINT(%f %f))", lng, lat)
	params := url.Values{}
	params.Set("service", "WFS")
	params.Set("version", "2.0.0")
	params.Set("request", "GetFeature")
	params.Set("typeNames", "layer-51564")
	params.Set("outputFormat", "application/json")
	params.Set("cql_filter", filter)

	path := "/services;key=" + c.apiKey + "/wfs?" + params.Encode()
	var result PropertyBoundaryResult
	if err := c.wfsGet(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("linz: property boundary query: %w", err)
	}

	out := make([]PropertyBoundary, 0, len(result.Features))
	for _, f := range result.Features {
		out = append(out, f.Properties)
	}
	return out, nil
}

// WithBaseURL overrides the base URL (useful for testing with a mock server).
func (c *Client) WithBaseURL(u string) *Client {
	c.baseURL = u
	return c
}

// wfsGet performs an authenticated GET against a LINZ WFS path and decodes the JSON response.
func (c *Client) wfsGet(ctx context.Context, path string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}
