package realme

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
)

// Provider is a configured RealMe Service Provider. Create one per application.
type Provider struct {
	cfg    Config
	sp     *saml.ServiceProvider
	store  SessionStore
}

// NewProvider creates a Provider from the given Config.
// It loads the SP certificate and private key, and fetches or reads IdP metadata.
func NewProvider(cfg Config, opts ...Option) (*Provider, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	// Load SP certificate and private key.
	keyPair, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("realme: load SP certificate: %w", err)
	}
	keyPair.Leaf, err = x509.ParseCertificate(keyPair.Certificate[0])
	if err != nil {
		return nil, fmt.Errorf("realme: parse SP certificate: %w", err)
	}

	privateKey, ok := keyPair.PrivateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("realme: SP private key must be RSA")
	}

	// Load IdP metadata.
	idpMeta, err := loadIdPMetadata(cfg)
	if err != nil {
		return nil, fmt.Errorf("realme: load IdP metadata: %w", err)
	}

	entityURL, err := url.Parse(cfg.EntityID)
	if err != nil {
		return nil, fmt.Errorf("realme: invalid EntityID URL: %w", err)
	}

	acsURL, err := url.Parse(cfg.ACSURL)
	if err != nil {
		return nil, fmt.Errorf("realme: invalid ACS URL: %w", err)
	}

	sp := &saml.ServiceProvider{
		EntityID:          cfg.EntityID,
		Key:               privateKey,
		Certificate:       keyPair.Leaf,
		MetadataURL:       *entityURL,
		AcsURL:            *acsURL,
		IDPMetadata:       idpMeta,
		ForceAuthn:        &cfg.ForceAuthn,
		AllowIDPInitiated: cfg.AllowIDPInitiated,
		AuthnNameIDFormat: saml.TransientNameIDFormat,
	}

	p := &Provider{
		cfg:   cfg,
		sp:    sp,
		store: newCookieSessionStore(cfg),
	}

	for _, opt := range opts {
		opt(p)
	}

	return p, nil
}

// Option is a functional option for Provider.
type Option func(*Provider)

// WithSessionStore replaces the default cookie-based session store.
func WithSessionStore(store SessionStore) Option {
	return func(p *Provider) {
		p.store = store
	}
}

// Middleware returns a samlsp.Middleware for use with net/http or Chi.
// This wraps the underlying crewjam/saml middleware with RealMe-specific logic.
func (p *Provider) middleware() *samlsp.Middleware {
	return &samlsp.Middleware{
		ServiceProvider: *p.sp,
		OnError:         samlsp.DefaultOnError,
		Session:         p.store,
	}
}

// loadIdPMetadata loads the IdP metadata from a local file or URL.
func loadIdPMetadata(cfg Config) (*saml.EntityDescriptor, error) {
	var data []byte
	var err error

	if cfg.IdPMetadataFile != "" {
		data, err = os.ReadFile(cfg.IdPMetadataFile)
		if err != nil {
			return nil, fmt.Errorf("read IdP metadata file %q: %w", cfg.IdPMetadataFile, err)
		}
	} else {
		resp, err := http.Get(cfg.IdPMetadataURL) //nolint:noctx
		if err != nil {
			return nil, fmt.Errorf("fetch IdP metadata from %q: %w", cfg.IdPMetadataURL, err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("IdP metadata returned HTTP %d from %q", resp.StatusCode, cfg.IdPMetadataURL)
		}
		data = make([]byte, 0, 65536)
		buf := make([]byte, 4096)
		for {
			n, readErr := resp.Body.Read(buf)
			data = append(data, buf[:n]...)
			if readErr != nil {
				break
			}
		}
	}

	meta := &saml.EntityDescriptor{}
	if err := xml.Unmarshal(data, meta); err != nil {
		return nil, fmt.Errorf("parse IdP metadata XML: %w", err)
	}
	return meta, nil
}
