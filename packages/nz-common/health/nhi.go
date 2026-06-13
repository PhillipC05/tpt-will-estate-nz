package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// WithBaseURL overrides the FHIR base URL (useful for testing with a mock server).
func (c *Client) WithBaseURL(u string) *Client {
	c.baseURL = u
	return c
}

// NHILookupResult contains the result of a National Health Index patient lookup.
type NHILookupResult struct {
	Patient   *Patient
	NHINumber string
	// Dormant indicates the NHI was merged into another record.
	Dormant bool
}

// GetPatientByNHI retrieves a patient from the National Health Index.
// Call only after the user has given explicit consent under the Health
// Information Privacy Code rule 10.
func (c *Client) GetPatientByNHI(ctx context.Context, nhiNumber string) (*NHILookupResult, error) {
	if nhiNumber == "" {
		return nil, fmt.Errorf("health: NHI number cannot be empty")
	}
	path := "/Patient?identifier=https://standards.digital.health.nz/ns/nhi-id|" + nhiNumber
	patients, err := c.searchBundle(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("health: NHI lookup %s: %w", nhiNumber, err)
	}
	if len(patients) == 0 {
		return nil, fmt.Errorf("health: patient not found for NHI %s", nhiNumber)
	}
	return &NHILookupResult{
		Patient:   &patients[0],
		NHINumber: nhiNumber,
	}, nil
}

// SearchPatientByDemographics performs a demographic search against the NHI
// using name and date of birth. Used when the NHI number is unknown.
func (c *Client) SearchPatientByDemographics(ctx context.Context, family, given, birthDate string) ([]Patient, error) {
	if family == "" || birthDate == "" {
		return nil, fmt.Errorf("health: family name and birthDate are required")
	}
	path := fmt.Sprintf("/Patient?family=%s&given=%s&birthdate=%s", family, given, birthDate)
	return c.searchBundle(ctx, path)
}

// searchBundle executes a FHIR search request and returns all Patient resources
// from the returned Bundle.
func (c *Client) searchBundle(ctx context.Context, path string) ([]Patient, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/fhir+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var bundle struct {
		Entry []struct {
			Resource Patient `json:"resource"`
		} `json:"entry"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&bundle); err != nil {
		return nil, fmt.Errorf("decode bundle: %w", err)
	}

	patients := make([]Patient, 0, len(bundle.Entry))
	for _, e := range bundle.Entry {
		if e.Resource.ResourceType == "Patient" {
			patients = append(patients, e.Resource)
		}
	}
	return patients, nil
}
