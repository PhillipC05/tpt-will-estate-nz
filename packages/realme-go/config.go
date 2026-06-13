package realme

import "fmt"

// Environment identifies which RealMe environment to target.
type Environment string

const (
	// MTS is the Messaging Test Site — used for initial component testing.
	// No real user identities; uses synthetic test users provided by DIA.
	MTS Environment = "mts"

	// ITE is the Integration Test Environment — pre-production staging.
	// Uses real RealMe accounts but flagged as test.
	ITE Environment = "ite"

	// Production is the live RealMe service.
	Production Environment = "production"
)

// RealMe IdP metadata URLs per environment.
// These point to the official DIA-hosted SAML metadata XML for each environment.
// You must download these and keep local copies; see the RealMe developer portal.
const (
	mtsLoginMetadataURL    = "https://mts.realme.govt.nz/saml2/metadata"
	mtsAssertionMetadataURL = "https://mts.realme.govt.nz/saml2/assertion/metadata"

	iteLoginMetadataURL    = "https://www.ite.logon.realme.govt.nz/saml2/metadata"
	iteAssertionMetadataURL = "https://www.ite.assertion.realme.govt.nz/saml2/metadata"

	prodLoginMetadataURL    = "https://www.realme.govt.nz/saml2/metadata"
	prodAssertionMetadataURL = "https://www.assertion.realme.govt.nz/saml2/metadata"
)

// LoginMetadataURL returns the IdP metadata URL for the RealMe Login Service.
func (e Environment) LoginMetadataURL() string {
	switch e {
	case MTS:
		return mtsLoginMetadataURL
	case ITE:
		return iteLoginMetadataURL
	default:
		return prodLoginMetadataURL
	}
}

// AssertionMetadataURL returns the IdP metadata URL for the RealMe Assertion Service
// (Verified Identity).
func (e Environment) AssertionMetadataURL() string {
	switch e {
	case MTS:
		return mtsAssertionMetadataURL
	case ITE:
		return iteAssertionMetadataURL
	default:
		return prodAssertionMetadataURL
	}
}

// Config holds the Service Provider configuration for a RealMe integration.
// One Config is needed per application registered with DIA.
type Config struct {
	// Environment selects MTS, ITE, or Production.
	Environment Environment

	// EntityID is the SP's unique SAML entity identifier.
	// Convention: "https://{your-domain}/saml/metadata"
	// Must match exactly what you registered with DIA.
	EntityID string

	// ACSURL is the Assertion Consumer Service URL — where RealMe POSTs the
	// SAML response after the user authenticates.
	// Convention: "https://{your-domain}/auth/realme/callback"
	ACSURL string

	// SLOUrl is the Single Logout URL (optional).
	// Convention: "https://{your-domain}/auth/realme/logout"
	SLOURL string

	// CertFile is the path to the SP's X.509 certificate (PEM).
	// Certificate naming convention required by DIA:
	//   {env}.{service}.{organisation_domain}
	// e.g. "mts.login.myapp.example.nz"
	CertFile string

	// KeyFile is the path to the SP's RSA private key (PEM).
	// Never commit this file — use environment variables or a secrets manager.
	KeyFile string

	// IdPMetadataFile is the path to a locally cached copy of the IdP's SAML
	// metadata XML. Download from Environment.LoginMetadataURL() and store
	// securely. Refresh before certificate expiry.
	IdPMetadataFile string

	// IdPMetadataURL may be used instead of IdPMetadataFile for dynamic
	// metadata loading (not recommended in production — network dependency at startup).
	IdPMetadataURL string

	// ForceAuthn, if true, forces RealMe to re-authenticate the user even if
	// they have an active RealMe session. Recommended for high-assurance flows.
	ForceAuthn bool

	// AllowIDPInitiated, if true, permits IdP-initiated SSO flows.
	// Disabled by default for security.
	AllowIDPInitiated bool

	// SessionCookieName is the name of the session cookie set after successful
	// authentication. Defaults to "realme_session".
	SessionCookieName string

	// SessionMaxAge is how long an authenticated session lasts.
	// Defaults to 8 hours.
	SessionMaxAge int // seconds
}

// Validate returns an error if any required Config fields are missing.
func (c *Config) Validate() error {
	if c.EntityID == "" {
		return fmt.Errorf("realme: Config.EntityID is required")
	}
	if c.ACSURL == "" {
		return fmt.Errorf("realme: Config.ACSURL is required")
	}
	if c.CertFile == "" {
		return fmt.Errorf("realme: Config.CertFile is required")
	}
	if c.KeyFile == "" {
		return fmt.Errorf("realme: Config.KeyFile is required")
	}
	if c.IdPMetadataFile == "" && c.IdPMetadataURL == "" {
		return fmt.Errorf("realme: one of Config.IdPMetadataFile or Config.IdPMetadataURL is required")
	}
	return nil
}

func (c *Config) sessionCookieName() string {
	if c.SessionCookieName != "" {
		return c.SessionCookieName
	}
	return "realme_session"
}

func (c *Config) sessionMaxAge() int {
	if c.SessionMaxAge > 0 {
		return c.SessionMaxAge
	}
	return 8 * 3600
}
