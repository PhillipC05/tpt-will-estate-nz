// Package realme provides a Go SAML 2.0 / OIDC client for New Zealand's RealMe
// identity service (https://developers.realme.govt.nz/).
//
// RealMe has two service tiers:
//   - Login Service: basic authentication, returns an opaque FLT identifier.
//   - Assertion Service: full verified identity — returns legal name, DOB,
//     place of birth, gender, photo, and verified residential address.
//
// Usage:
//
//	cfg := realme.Config{
//	    Environment:    realme.MTS,
//	    EntityID:       "https://myapp.example.nz/saml/metadata",
//	    ACSURL:         "https://myapp.example.nz/auth/realme/callback",
//	    CertFile:       "certs/mts.sp.crt",
//	    KeyFile:        "certs/mts.sp.key",
//	    IdPMetadataURL: realme.MTS.LoginServiceMetadataURL(),
//	}
//
//	sp, err := realme.NewProvider(cfg)
//	if err != nil { ... }
//
//	r.Get("/auth/realme/login",    sp.LoginHandler())
//	r.Get("/auth/realme/callback", sp.CallbackHandler(onSuccess))
//	r.Get("/saml/metadata",        sp.MetadataHandler())
//
//	// Protect routes:
//	r.With(sp.RequireLogin()).Get("/dashboard", dashboardHandler)
//	r.With(sp.RequireVerified()).Post("/incorporate", incorporateHandler)
package realme

import "time"

// AssuranceLevel represents the RealMe identity assurance level.
type AssuranceLevel int

const (
	// LevelNone indicates no identity verification.
	LevelNone AssuranceLevel = 0
	// LevelLogin indicates basic RealMe login (username + password + 2FA).
	// The user's identity is not government-verified.
	LevelLogin AssuranceLevel = 1
	// LevelVerified indicates RealMe Verified Identity — the user's identity
	// has been verified against government records by DIA.
	LevelVerified AssuranceLevel = 2
)

func (l AssuranceLevel) String() string {
	switch l {
	case LevelLogin:
		return "login"
	case LevelVerified:
		return "verified"
	default:
		return "none"
	}
}

// Identity holds the identity attributes extracted from a RealMe SAML assertion.
// FLT is always populated. Verified-only fields are only present when
// AssuranceLevel == LevelVerified.
type Identity struct {
	// FLT is the Federated Login Token — an opaque, persistent, per-service
	// identifier for the user. Use this as your internal user key.
	// Never expose the FLT to end users.
	FLT string

	// AssuranceLevel indicates whether this is a basic login or verified identity.
	AssuranceLevel AssuranceLevel

	// AuthnInstant is when the user authenticated at RealMe.
	AuthnInstant time.Time

	// SessionIndex is the RealMe session identifier (used for SLO).
	SessionIndex string

	// --- Verified Identity fields (LevelVerified only) ---

	// FullName is the legal full name as verified against government records.
	FullName string

	// DateOfBirth is the verified date of birth.
	DateOfBirth time.Time

	// PlaceOfBirth is the verified place of birth.
	PlaceOfBirth string

	// Gender is the verified gender ("male", "female", "unspecified").
	Gender string

	// Address is the verified residential address.
	Address *Address
}

// IsVerified returns true when the identity carries government-verified attributes.
func (i *Identity) IsVerified() bool {
	return i.AssuranceLevel >= LevelVerified
}

// Address is a New Zealand residential address as returned by the RealMe
// Assertion Service.
type Address struct {
	Unit     string
	Number   string
	Street   string
	Suburb   string
	City     string
	Postcode string
	Country  string
}

// Line1 returns the first address line (unit + number + street).
func (a *Address) Line1() string {
	if a.Unit != "" {
		return a.Unit + "/" + a.Number + " " + a.Street
	}
	return a.Number + " " + a.Street
}
