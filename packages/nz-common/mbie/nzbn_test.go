package mbie_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tpt-nz/nz-common/mbie"
)

func TestGetEntityByNZBN(t *testing.T) {
	want := mbie.NZBNEntity{
		NZBN:             "9429039822327",
		EntityName:       "ACME LIMITED",
		EntityTypeCode:   "LTD",
		EntityStatusCode: "50",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v3/nzbn/entities/9429039822327" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(want)
	}))
	defer srv.Close()

	client := mbie.NewClient("test-key").WithBaseURL(srv.URL)
	got, err := client.GetEntityByNZBN(context.Background(), "9429039822327")
	if err != nil {
		t.Fatalf("GetEntityByNZBN: %v", err)
	}
	if got.NZBN != want.NZBN {
		t.Errorf("NZBN: got %q, want %q", got.NZBN, want.NZBN)
	}
	if got.EntityName != want.EntityName {
		t.Errorf("EntityName: got %q, want %q", got.EntityName, want.EntityName)
	}
}

func TestGetEntityByNZBN_Empty(t *testing.T) {
	client := mbie.NewClient("test-key")
	_, err := client.GetEntityByNZBN(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty NZBN")
	}
}

func TestSearchByNZBN(t *testing.T) {
	want := mbie.NZBNSearchResult{
		Items:      []mbie.NZBNEntity{{NZBN: "9429039822327", EntityName: "ACME LIMITED"}},
		TotalItems: 1,
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(want)
	}))
	defer srv.Close()

	client := mbie.NewClient("test-key").WithBaseURL(srv.URL)
	got, err := client.SearchByNZBN(context.Background(), "ACME", 10, 1)
	if err != nil {
		t.Fatalf("SearchByNZBN: %v", err)
	}
	if got.TotalItems != 1 {
		t.Errorf("TotalItems: got %d, want 1", got.TotalItems)
	}
}
