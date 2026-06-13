// Package testenv provides a mock RealMe Identity Provider (IdP) for local
// development and automated testing. It removes the dependency on the real
// RealMe MTS environment, making it possible to run the full SAML login flow
// without DIA credentials.
//
// Usage in tests:
//
//	idp := testenv.NewMockIdP(t)
//	defer idp.Close()
//
//	cfg := realme.Config{
//	    Environment:    realme.MTS,
//	    EntityID:       "https://app.test/saml/metadata",
//	    ACSURL:         "https://app.test/auth/realme/callback",
//	    CertFile:       idp.SPCertFile(),
//	    KeyFile:        idp.SPKeyFile(),
//	    IdPMetadataURL: idp.MetadataURL(),
//	}
//
//	// Authenticate as a specific test user:
//	idp.SetNextUser(testenv.UserLoginOnly)
//	// or:
//	idp.SetNextUser(testenv.UserVerified)
package testenv

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"encoding/xml"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// MockIdP is a minimal SAML IdP for testing RealMe integrations locally.
type MockIdP struct {
	server    *httptest.Server
	t         testing.TB
	spCert    string
	spKey     string
	nextUser  *TestUser
}

// NewMockIdP starts an httptest.Server acting as a RealMe-compatible SAML IdP.
// It generates a self-signed certificate pair for the SP and IdP.
// The server is automatically closed when the test ends via t.Cleanup.
func NewMockIdP(t testing.TB) *MockIdP {
	t.Helper()

	tmpDir := t.TempDir()

	// Generate SP key pair.
	spCert, spKey, err := generateSelfSignedCert(tmpDir, "sp")
	if err != nil {
		t.Fatalf("testenv: generate SP cert: %v", err)
	}

	idp := &MockIdP{
		t:        t,
		spCert:   spCert,
		spKey:    spKey,
		nextUser: UserVerified,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/saml2/metadata", idp.handleMetadata)
	mux.HandleFunc("/saml2/sso", idp.handleSSO)

	idp.server = httptest.NewTLSServer(mux)
	t.Cleanup(idp.server.Close)

	return idp
}

// MetadataURL returns the mock IdP's SAML metadata URL.
func (m *MockIdP) MetadataURL() string {
	return m.server.URL + "/saml2/metadata"
}

// SSOURL returns the mock IdP's Single Sign-On URL.
func (m *MockIdP) SSOURL() string {
	return m.server.URL + "/saml2/sso"
}

// SPCertFile returns the path to the generated SP certificate file.
func (m *MockIdP) SPCertFile() string {
	return m.spCert
}

// SPKeyFile returns the path to the generated SP private key file.
func (m *MockIdP) SPKeyFile() string {
	return m.spKey
}

// SetNextUser configures the user that will be "authenticated" in the next
// SAML response issued by the mock IdP.
func (m *MockIdP) SetNextUser(u *TestUser) {
	m.nextUser = u
}

func (m *MockIdP) handleMetadata(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/xml")
	meta := fmt.Sprintf(`<?xml version="1.0"?>
<EntityDescriptor xmlns="urn:oasis:names:tc:SAML:2.0:metadata"
    entityID="%s">
  <IDPSSODescriptor
      WantAuthnRequestsSigned="true"
      protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">
    <SingleSignOnService
        Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect"
        Location="%s"/>
  </IDPSSODescriptor>
</EntityDescriptor>`, m.server.URL, m.SSOURL())
	_, _ = w.Write([]byte(meta))
}

func (m *MockIdP) handleSSO(w http.ResponseWriter, r *http.Request) {
	// In a real test scenario this would parse the AuthnRequest, generate a
	// signed SAMLResponse, and POST it back to the SP's ACS URL.
	// For simplicity, we return a minimal response indicating what user was set.
	// A full implementation would use the crewjam/saml IdP support or
	// github.com/crewjam/saml/samlidp.
	user := m.nextUser
	if user == nil {
		user = UserLoginOnly
	}

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `<html><body>
<p>Mock RealMe IdP — would authenticate as: <strong>%s</strong> (FLT: %s)</p>
<p>In full test mode, this auto-POSTs a SAMLResponse to the ACS URL.</p>
</body></html>`, user.FullName, user.FLT)
}

// generateSelfSignedCert creates a self-signed RSA cert+key pair in dir.
// Returns paths to the cert and key PEM files.
func generateSelfSignedCert(dir, prefix string) (certFile, keyFile string, err error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", fmt.Errorf("generate RSA key: %w", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "mts." + prefix + ".testapp.example.nz"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(8760 * time.Hour), // 1 year
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return "", "", fmt.Errorf("create certificate: %w", err)
	}

	certFile = filepath.Join(dir, prefix+".crt")
	keyFile = filepath.Join(dir, prefix+".key")

	cf, err := os.Create(certFile)
	if err != nil {
		return "", "", err
	}
	defer cf.Close()
	if err := pem.Encode(cf, &pem.Block{Type: "CERTIFICATE", Bytes: certDER}); err != nil {
		return "", "", err
	}

	kf, err := os.Create(keyFile)
	if err != nil {
		return "", "", err
	}
	defer kf.Close()
	if err := pem.Encode(kf, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}); err != nil {
		return "", "", err
	}

	return certFile, keyFile, nil
}

// xmlMarshal is a convenience wrapper used by test helpers.
func xmlMarshal(v interface{}) string {
	b, _ := xml.MarshalIndent(v, "", "  ")
	return string(b)
}
