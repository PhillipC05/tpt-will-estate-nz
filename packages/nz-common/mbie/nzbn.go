package mbie

import (
	"context"
	"fmt"
)

// NZBNEntity is a NZ Business Number entity record.
type NZBNEntity struct {
	NZBN             string   `json:"nzbn"`
	EntityName       string   `json:"entityName"`
	EntityTypeCode   string   `json:"entityTypeCode"`
	EntityStatusCode string   `json:"entityStatusCode"`
	RegistrationDate string   `json:"registrationDate"`
	TradingNames     []string `json:"tradingNames,omitempty"`
}

// NZBNSearchResult is the paginated NZBN search response.
type NZBNSearchResult struct {
	Items      []NZBNEntity `json:"items"`
	TotalItems int          `json:"totalItems"`
}

// GetEntityByNZBN retrieves a business entity by its NZ Business Number.
// The NZBN is a 13-digit number uniquely identifying every NZ entity.
func (c *Client) GetEntityByNZBN(ctx context.Context, nzbn string) (*NZBNEntity, error) {
	if nzbn == "" {
		return nil, fmt.Errorf("mbie: nzbn cannot be empty")
	}
	var entity NZBNEntity
	if err := c.get(ctx, "/v3/nzbn/entities/"+nzbn, &entity); err != nil {
		return nil, fmt.Errorf("mbie: get entity by NZBN %s: %w", nzbn, err)
	}
	return &entity, nil
}

// SearchByNZBN searches the NZBN register by entity name or NZBN.
// Returns paginated results; page is 1-based.
func (c *Client) SearchByNZBN(ctx context.Context, q string, limit, page int) (*NZBNSearchResult, error) {
	if q == "" {
		return nil, fmt.Errorf("mbie: nzbn search query cannot be empty")
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if page <= 0 {
		page = 1
	}
	path := fmt.Sprintf("/v3/nzbn/entities?search-term=%s&page-size=%d&page=%d", q, limit, page)
	var result NZBNSearchResult
	if err := c.get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("mbie: nzbn search: %w", err)
	}
	return &result, nil
}
