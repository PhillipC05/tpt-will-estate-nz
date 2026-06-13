package health_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tpt-nz/nz-common/health"
)

func TestGetPatientByNHI(t *testing.T) {
	bundle := map[string]any{
		"resourceType": "Bundle",
		"entry": []map[string]any{
			{
				"resource": map[string]any{
					"resourceType": "Patient",
					"id":          "nhi-test-001",
					"identifier": []map[string]any{
						{
							"system": "https://standards.digital.health.nz/ns/nhi-id",
							"value":  "ZAA0001",
						},
					},
					"name": []map[string]any{
						{"family": "Smith", "given": []string{"John"}, "use": "official"},
					},
					"birthDate": "1985-03-15",
				},
			},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/fhir+json")
		_ = json.NewEncoder(w).Encode(bundle)
	}))
	defer srv.Close()

	client := health.NewNHIClient("test-bearer-token")
	client.WithBaseURL(srv.URL)

	result, err := client.GetPatientByNHI(context.Background(), "ZAA0001")
	if err != nil {
		t.Fatalf("GetPatientByNHI: %v", err)
	}
	if result.NHINumber != "ZAA0001" {
		t.Errorf("NHINumber: got %q, want ZAA0001", result.NHINumber)
	}
	if result.Patient == nil {
		t.Fatal("expected non-nil patient")
	}
	if result.Patient.ID != "nhi-test-001" {
		t.Errorf("Patient.ID: got %q, want nhi-test-001", result.Patient.ID)
	}
}

func TestGetPatientByNHI_Empty(t *testing.T) {
	client := health.NewNHIClient("test-token")
	_, err := client.GetPatientByNHI(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty NHI number")
	}
}

func TestSearchPatientByDemographics_RequiresFields(t *testing.T) {
	client := health.NewNHIClient("test-token")
	_, err := client.SearchPatientByDemographics(context.Background(), "", "", "1985-01-01")
	if err == nil {
		t.Fatal("expected error when family name is missing")
	}
}
