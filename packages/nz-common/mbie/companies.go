package mbie

import (
	"context"
	"fmt"
	"net/url"
)

// Company is a New Zealand registered company from the Companies Register.
type Company struct {
	CompanyNumber    string     `json:"companyNumber"`
	CompanyName      string     `json:"companyName"`
	CompanyStatus    string     `json:"companyStatus"` // "Registered", "Removed", etc.
	EntityTypeCode   string     `json:"entityTypeCode"`
	RegistrationDate string     `json:"registrationDate"`
	NZBN             string     `json:"nzbn"`
	Directors        []Director `json:"directors,omitempty"`
}

// Director is a company director as recorded in the Companies Register.
type Director struct {
	FullName        string `json:"fullName"`
	ResidentialCity string `json:"residentialCity"`
	ConsentDate     string `json:"consentDate"`
}

// CompaniesSearchResult is the paginated response from the companies search API.
type CompaniesSearchResult struct {
	Items      []Company `json:"items"`
	TotalItems int       `json:"totalItems"`
}

// SearchCompanies searches the NZ Companies Register by name.
// q is the search string. Returns up to limit results.
func (c *Client) SearchCompanies(ctx context.Context, q string, limit int) (*CompaniesSearchResult, error) {
	if q == "" {
		return nil, fmt.Errorf("mbie: companies search query cannot be empty")
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	path := fmt.Sprintf("/v3/companies?q=%s&pageSize=%d", url.QueryEscape(q), limit)
	var result CompaniesSearchResult
	if err := c.get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetCompany retrieves a company by its company number.
func (c *Client) GetCompany(ctx context.Context, companyNumber string) (*Company, error) {
	if companyNumber == "" {
		return nil, fmt.Errorf("mbie: companyNumber is required")
	}
	var company Company
	if err := c.get(ctx, "/v3/companies/"+url.PathEscape(companyNumber), &company); err != nil {
		return nil, err
	}
	return &company, nil
}

// GetCompaniesByNZBN retrieves a company by its NZ Business Number.
func (c *Client) GetCompaniesByNZBN(ctx context.Context, nzbn string) (*Company, error) {
	if nzbn == "" {
		return nil, fmt.Errorf("mbie: nzbn is required")
	}
	var company Company
	if err := c.get(ctx, "/v3/companies?nzbn="+url.QueryEscape(nzbn), &company); err != nil {
		return nil, err
	}
	return &company, nil
}
