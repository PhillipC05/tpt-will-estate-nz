// Command mock-idp is a standalone mock RealMe Identity Provider for local development.
// It simulates the RealMe SAML IdP for testing applications without DIA credentials.
//
// Usage:
//
//	go run ./cmd/mock-idp
//
// Listens on :8081 by default (set IDP_ADDR env var to change).
// Serves metadata at /metadata and SSO at /sso.
package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func main() {
	addr := os.Getenv("IDP_ADDR")
	if addr == "" {
		addr = ":8081"
	}

	tmpDir, err := os.MkdirTemp("", "mock-idp-*")
	if err != nil {
		log.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	spCert, spKey, err := generateCert(tmpDir, "sp")
	if err != nil {
		log.Fatalf("failed to generate SP cert: %v", err)
	}
	log.Printf("SP certificate: %s", spCert)
	log.Printf("SP key: %s", spKey)

	mux := http.NewServeMux()
	mux.HandleFunc("/metadata", handleMetadata)
	mux.HandleFunc("/sso", handleSSO)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	log.Printf("Mock RealMe IdP starting on %s", addr)
	log.Printf("Metadata URL: http://localhost%s/metadata", addr)
	log.Printf("SSO URL: http://localhost%s/sso", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func handleMetadata(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/xml")
	meta := fmt.Sprintf(`<?xml version="1.0"?>
<EntityDescriptor xmlns="urn:oasis:names:tc:SAML:2.0:metadata"
    entityID="http://localhost%s">
  <IDPSSODescriptor
      WantAuthnRequestsSigned="true"
      protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">
    <SingleSignOnService
        Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect"
        Location="http://localhost%s/sso"/>
  </IDPSSODescriptor>
</EntityDescriptor>`, r.Host, r.Host)
	w.Write([]byte(meta))
}

func handleSSO(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, `<!DOCTYPE html>
<html>
<head><title>Mock RealMe IdP</title></head>
<body style="font-family: system-ui, sans-serif; max-width: 600px; margin: 40px auto; padding: 20px;">
  <h1>Mock RealMe IdP</h1>
  <p>This is a simulated RealMe login page for local development.</p>
  <form method="post" action="/sso/callback">
    <p>Select a test user:</p>
    <p><button type="submit" name="user" value="login" style="padding: 10px 20px; font-size: 16px;">Login (Basic Auth)</button></p>
    <p><button type="submit" name="user" value="verified" style="padding: 10px 20px; font-size: 16px;">Login (Verified Identity)</button></p>
    <p><button type="submit" name="user" value="verified2" style="padding: 10px 20px; font-size: 16px;">Login (Verified User 2)</button></p>
  </form>
  <p style="color: #666; font-size: 12px;">In a full implementation, this would POST a signed SAMLResponse to the ACS URL.</p>
</body>
</html>`)
}

func generateCert(dir, prefix string) (certFile, keyFile string, err error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", fmt.Errorf("generate RSA key: %w", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "mock-idp.localhost"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(8760 * time.Hour),
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