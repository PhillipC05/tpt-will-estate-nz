package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"log/slog"

	"github.com/tpt-nz/tpt-will-estate-nz/internal/handlers"
	"github.com/tpt-nz/tpt-will-estate-nz/internal/models"
	"github.com/tpt-nz/tpt-will-estate-nz/internal/services"
	"github.com/tpt-nz/realme-go"
)

// mockWillService stubs the WillService for handler tests.
type mockWillRepo struct{}

func (m *mockWillRepo) Save(_ interface{}) (models.WillID, error) { return "will_test_001", nil }

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestWillHandler_Create_Unauthenticated(t *testing.T) {
	encSvc := services.NewEncryptionService()
	// Use nil repo — unauthenticated requests should be rejected before DB access.
	willSvc := services.NewWillService(nil, encSvc, testLogger())
	h := handlers.NewWillHandler(willSvc, testLogger())

	body, _ := json.Marshal(services.CreateWillRequest{
		VaultCiphertext: "abc",
		VaultNonce:      "nonce",
		VaultAlg:        "AES-GCM",
	})
	r := httptest.NewRequest(http.MethodPost, "/wills", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.Create(w, r) // No identity in context.

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestWillHandler_Create_Authenticated(t *testing.T) {
	encSvc := services.NewEncryptionService()
	willSvc := services.NewWillService(nil, encSvc, testLogger())
	h := handlers.NewWillHandler(willSvc, testLogger())

	body, _ := json.Marshal(services.CreateWillRequest{
		VaultCiphertext: "ciphertext",
		VaultNonce:      "nonce",
		VaultAlg:        "AES-GCM",
	})
	r := httptest.NewRequest(http.MethodPost, "/wills", bytes.NewReader(body))
	r = r.WithContext(realme.IdentityToContext(r.Context(), &realme.Identity{
		FLT:            "flt-will-testator-001",
		AssuranceLevel: realme.LevelVerified,
		FullName:       "Jane Smith",
	}))
	w := httptest.NewRecorder()

	h.Create(w, r)

	// Service returns an error because repo is nil, but the handler
	// should not panic and should return a proper error status.
	if w.Code == http.StatusUnauthorized {
		t.Error("authenticated request should not return 401")
	}
}
