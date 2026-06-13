package realme_test

import (
	"net/http/httptest"
	"testing"

	"github.com/tpt-nz/realme-go"
	"github.com/tpt-nz/realme-go/testenv"
)

func TestNewProvider_InvalidConfig(t *testing.T) {
	_, err := realme.NewProvider(realme.Config{})
	if err == nil {
		t.Fatal("expected error for empty config")
	}
}

func TestNewProvider_WithMockIdP(t *testing.T) {
	idp := testenv.NewMockIdP(t)

	cfg := realme.Config{
		Environment:    realme.MTS,
		EntityID:       "https://app.test/saml/metadata",
		ACSURL:         "https://app.test/auth/realme/callback",
		CertFile:       idp.SPCertFile(),
		KeyFile:        idp.SPKeyFile(),
		IdPMetadataURL: idp.MetadataURL(),
	}

	provider, err := realme.NewProvider(cfg)
	if err != nil {
		t.Fatalf("NewProvider: %v", err)
	}
	if provider == nil {
		t.Fatal("expected non-nil provider")
	}
}

func TestMetadataHandler(t *testing.T) {
	idp := testenv.NewMockIdP(t)

	cfg := realme.Config{
		Environment:    realme.MTS,
		EntityID:       "https://app.test/saml/metadata",
		ACSURL:         "https://app.test/auth/realme/callback",
		CertFile:       idp.SPCertFile(),
		KeyFile:        idp.SPKeyFile(),
		IdPMetadataURL: idp.MetadataURL(),
	}

	provider, err := realme.NewProvider(cfg)
	if err != nil {
		t.Fatalf("NewProvider: %v", err)
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/saml/metadata", nil)
	provider.MetadataHandler()(w, r)

	if w.Code != 200 {
		t.Errorf("MetadataHandler returned %d, want 200", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/xml; charset=utf-8" {
		t.Errorf("Content-Type = %q, want application/xml; charset=utf-8", ct)
	}
	body := w.Body.String()
	if len(body) < 100 {
		t.Errorf("metadata body too short (%d bytes)", len(body))
	}
}

func TestLoginHandler_RedirectsToIdP(t *testing.T) {
	idp := testenv.NewMockIdP(t)

	cfg := realme.Config{
		Environment:    realme.MTS,
		EntityID:       "https://app.test/saml/metadata",
		ACSURL:         "https://app.test/auth/realme/callback",
		CertFile:       idp.SPCertFile(),
		KeyFile:        idp.SPKeyFile(),
		IdPMetadataURL: idp.MetadataURL(),
	}

	provider, err := realme.NewProvider(cfg)
	if err != nil {
		t.Fatalf("NewProvider: %v", err)
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/auth/realme/login", nil)
	provider.LoginHandler()(w, r)

	// Should redirect to the IdP or return a SAML request page.
	if w.Code != 302 && w.Code != 200 {
		t.Errorf("LoginHandler returned %d, want 302 or 200", w.Code)
	}
}
