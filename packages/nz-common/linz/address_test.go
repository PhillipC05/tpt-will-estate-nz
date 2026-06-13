package linz_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tpt-nz/nz-common/linz"
)

func TestSearchAddresses(t *testing.T) {
	fixture := map[string]any{
		"features": []map[string]any{
			{"properties": map[string]any{
				"address_id":   1,
				"full_address": "1 Queen Street, Auckland",
			}},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(fixture)
	}))
	defer srv.Close()

	client := linz.NewClient("test-key").WithBaseURL(srv.URL)
	addrs, err := client.SearchAddresses(context.Background(), "Queen Street", 5)
	if err != nil {
		t.Fatalf("SearchAddresses: %v", err)
	}
	if len(addrs) != 1 {
		t.Fatalf("expected 1 result, got %d", len(addrs))
	}
	if addrs[0].FullAddress != "1 Queen Street, Auckland" {
		t.Errorf("unexpected address: %q", addrs[0].FullAddress)
	}
}

func TestSearchAddresses_EmptyQuery(t *testing.T) {
	client := linz.NewClient("test-key")
	_, err := client.SearchAddresses(context.Background(), "", 5)
	if err == nil {
		t.Fatal("expected error for empty query")
	}
}

func TestGetPropertyBoundaries(t *testing.T) {
	fixture := map[string]any{
		"features": []map[string]any{
			{"properties": map[string]any{
				"id":          42,
				"appellation": "Lot 1 DP 12345",
			}},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(fixture)
	}))
	defer srv.Close()

	client := linz.NewClient("test-key").WithBaseURL(srv.URL)
	boundaries, err := client.GetPropertyBoundaries(context.Background(), -36.8485, 174.7633)
	if err != nil {
		t.Fatalf("GetPropertyBoundaries: %v", err)
	}
	if len(boundaries) != 1 {
		t.Fatalf("expected 1 boundary, got %d", len(boundaries))
	}
	if boundaries[0].ParcelID != 42 {
		t.Errorf("ParcelID: got %d, want 42", boundaries[0].ParcelID)
	}
}
