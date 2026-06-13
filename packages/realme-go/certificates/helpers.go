// Package certificates provides helpers for loading and validating the X.509
// certificates required for RealMe SAML integration.
//
// Certificate naming convention required by DIA:
//
//	{environment}.{service}.{organisation_domain}
//
// Examples:
//
//	mts.login.myapp.example.nz      ← MTS Login Service
//	mts.assertion.myapp.example.nz  ← MTS Assertion Service
//	ite.login.myapp.example.nz      ← ITE Login Service
//	login.myapp.example.nz          ← Production Login Service
//
// Certificates must be obtained from an accredited Certificate Authority.
// Self-signed certificates are only accepted in the MTS environment.
package certificates

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"time"
)

// LoadKeyPair loads an X.509 certificate and RSA private key from PEM files
// and validates they are suitable for RealMe SAML signing.
func LoadKeyPair(certFile, keyFile string) (tls.Certificate, error) {
	pair, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("load key pair: %w", err)
	}

	pair.Leaf, err = x509.ParseCertificate(pair.Certificate[0])
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("parse certificate: %w", err)
	}

	if _, ok := pair.PrivateKey.(*rsa.PrivateKey); !ok {
		return tls.Certificate{}, fmt.Errorf("private key must be RSA (got %T)", pair.PrivateKey)
	}

	return pair, nil
}

// CertInfo holds human-readable information about a certificate for diagnostics.
type CertInfo struct {
	Subject    string
	Issuer     string
	NotBefore  time.Time
	NotAfter   time.Time
	DaysLeft   int
	IsExpired  bool
}

// InspectCertFile reads a PEM certificate file and returns its metadata.
func InspectCertFile(certFile string) (*CertInfo, error) {
	data, err := os.ReadFile(certFile)
	if err != nil {
		return nil, fmt.Errorf("read cert file: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("no PEM block found in %q", certFile)
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse certificate: %w", err)
	}

	now := time.Now()
	daysLeft := int(cert.NotAfter.Sub(now).Hours() / 24)

	return &CertInfo{
		Subject:   cert.Subject.CommonName,
		Issuer:    cert.Issuer.CommonName,
		NotBefore: cert.NotBefore,
		NotAfter:  cert.NotAfter,
		DaysLeft:  daysLeft,
		IsExpired: now.After(cert.NotAfter),
	}, nil
}

// WarnIfExpiringSoon returns a warning string if the certificate at certFile
// expires within the given threshold (days). Returns "" if not near expiry.
func WarnIfExpiringSoon(certFile string, thresholdDays int) string {
	info, err := InspectCertFile(certFile)
	if err != nil {
		return fmt.Sprintf("WARNING: could not inspect certificate %q: %v", certFile, err)
	}
	if info.IsExpired {
		return fmt.Sprintf("WARNING: certificate %q (CN=%s) has EXPIRED", certFile, info.Subject)
	}
	if info.DaysLeft <= thresholdDays {
		return fmt.Sprintf("WARNING: certificate %q (CN=%s) expires in %d days", certFile, info.Subject, info.DaysLeft)
	}
	return ""
}
