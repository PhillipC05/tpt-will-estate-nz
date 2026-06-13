// Package health provides a client for the Health New Zealand (Te Whatu Ora)
// FHIR R4 APIs, including the National Health Index (NHI), Shared Digital
// Health Record (SDHR), Health Provider Index (HPI), and Medical Warning
// System (MWS).
//
// Register for access: https://www.tewhatuora.govt.nz/health-services-and-programmes/digital-health/digital-services-hub/explore-apis-digital-services/
// FHIR standards: https://fhir.org.nz/
package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Endpoint URLs for Health NZ FHIR APIs.
// These require registration and OAuth2 client credentials.
const (
	NHIBaseURL  = "https://api.hip.digital.health.nz/fhir/R4"
	SDHRBaseURL = "https://api.sdhb.digital.health.nz/fhir/R4"
	HPIBaseURL  = "https://api.hpi.digital.health.nz/fhir/R4"
	MWSBaseURL  = "https://api.mws.digital.health.nz/fhir/R4"
)

// Client is the Health NZ FHIR API client.
type Client struct {
	baseURL    string
	token      string // OAuth2 bearer token
	httpClient *http.Client
}

// NewNHIClient creates a client for the National Health Index (NHI) API.
func NewNHIClient(bearerToken string) *Client {
	return &Client{
		baseURL: NHIBaseURL,
		token:   bearerToken,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// NewSDHRClient creates a client for the Shared Digital Health Record API.
func NewSDHRClient(bearerToken string) *Client {
	return &Client{
		baseURL: SDHRBaseURL,
		token:   bearerToken,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// Patient is a minimal FHIR Patient resource, representing a NZ patient.
type Patient struct {
	ResourceType string      `json:"resourceType"`
	ID           string      `json:"id"`
	Identifier   []Identifier `json:"identifier"`
	Name         []HumanName  `json:"name"`
	BirthDate    string       `json:"birthDate"`
	Gender       string       `json:"gender"`
	Address      []FHIRAddress `json:"address"`
}

// Identifier is a FHIR Identifier (e.g. NHI number).
type Identifier struct {
	System string `json:"system"`
	Value  string `json:"value"`
}

// HumanName is a FHIR HumanName.
type HumanName struct {
	Family string   `json:"family"`
	Given  []string `json:"given"`
	Use    string   `json:"use"`
}

// FHIRAddress is a FHIR Address.
type FHIRAddress struct {
	Line       []string `json:"line"`
	City       string   `json:"city"`
	PostalCode string   `json:"postalCode"`
	Country    string   `json:"country"`
	Use        string   `json:"use"`
}

// NHINumber extracts the NHI number from a Patient's identifiers.
// The NHI system URI is https://standards.digital.health.nz/ns/nhi-id.
func (p *Patient) NHINumber() string {
	const nhiSystem = "https://standards.digital.health.nz/ns/nhi-id"
	for _, id := range p.Identifier {
		if id.System == nhiSystem {
			return id.Value
		}
	}
	return ""
}

// GetPatient retrieves a FHIR Patient by NHI number.
func (c *Client) GetPatient(ctx context.Context, nhiNumber string) (*Patient, error) {
	if nhiNumber == "" {
		return nil, fmt.Errorf("health: NHI number is required")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.baseURL+"/Patient?identifier=https://standards.digital.health.nz/ns/nhi-id|"+nhiNumber, nil)
	if err != nil {
		return nil, fmt.Errorf("health: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/fhir+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("health: get patient: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("health: patient with NHI %q not found", nhiNumber)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("health: get patient returned HTTP %d", resp.StatusCode)
	}

	// The NHI API returns a Bundle; extract the first Patient entry.
	var bundle struct {
		Entry []struct {
			Resource Patient `json:"resource"`
		} `json:"entry"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&bundle); err != nil {
		return nil, fmt.Errorf("health: decode patient response: %w", err)
	}
	if len(bundle.Entry) == 0 {
		return nil, fmt.Errorf("health: patient with NHI %q not found", nhiNumber)
	}
	p := bundle.Entry[0].Resource
	return &p, nil
}
