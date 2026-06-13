package realme

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
	"github.com/golang-jwt/jwt/v5"
)

// contextKey is the type for values stored in request context by this package.
type contextKey string

const identityContextKey contextKey = "realme.identity"

// SessionStore is the interface implemented by session backends.
// The default implementation uses signed JWT cookies.
// Replace with a Redis-backed store for multi-instance deployments.
type SessionStore interface {
	samlsp.SessionProvider
	DeleteSession(w http.ResponseWriter, r *http.Request) error
}

// IdentityFromContext retrieves the authenticated Identity from the request context.
// Returns nil if no identity is present (unauthenticated request).
func IdentityFromContext(ctx context.Context) *Identity {
	v := ctx.Value(identityContextKey)
	if v == nil {
		return nil
	}
	id, _ := v.(*Identity)
	return id
}

// IdentityToContext stores an Identity in the request context.
// Useful for testing handlers that require authentication.
func IdentityToContext(ctx context.Context, identity *Identity) context.Context {
	return context.WithValue(ctx, identityContextKey, identity)
}

// cookieSessionStore is the default session store using signed JWT cookies.
// It implements the crewjam/saml SessionStore interface.
type cookieSessionStore struct {
	cfg    Config
	sigKey []byte
}

func newCookieSessionStore(cfg Config) *cookieSessionStore {
	// Derive a signing key from the entity ID for the dev/test case.
	// In production, set an explicit SESSION_SECRET environment variable.
	return &cookieSessionStore{
		cfg:    cfg,
		sigKey: []byte(cfg.EntityID),
	}
}

// identityClaims embeds jwt.RegisteredClaims so that the Identity can be
// round-tripped through a JWT cookie without exposing PII in plain text.
type identityClaims struct {
	jwt.RegisteredClaims
	FLT            string         `json:"flt"`
	AssuranceLevel AssuranceLevel `json:"al"`
	SessionIndex   string         `json:"si"`
	FullName       string         `json:"fn,omitempty"`
	DOB            string         `json:"dob,omitempty"`
	PlaceOfBirth   string         `json:"pob,omitempty"`
	Gender         string         `json:"g,omitempty"`
	AddressJSON    string         `json:"addr,omitempty"`
}

// GetSession reads the session cookie and returns the decoded identity.
// Implements samlsp.Session.
func (s *cookieSessionStore) GetSession(r *http.Request) (samlsp.Session, error) {
	cookie, err := r.Cookie(s.cfg.sessionCookieName())
	if err != nil {
		return nil, samlsp.ErrNoSession
	}

	claims := &identityClaims{}
	_, err = jwt.ParseWithClaims(cookie.Value, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.sigKey, nil
	})
	if err != nil {
		return nil, samlsp.ErrNoSession
	}

	id := &Identity{
		FLT:            claims.FLT,
		AssuranceLevel: claims.AssuranceLevel,
		SessionIndex:   claims.SessionIndex,
		FullName:       claims.FullName,
		PlaceOfBirth:   claims.PlaceOfBirth,
		Gender:         claims.Gender,
	}
	if claims.DOB != "" {
		t, _ := time.Parse("2006-01-02", claims.DOB)
		id.DateOfBirth = t
	}
	if claims.AddressJSON != "" {
		var addr Address
		if json.Unmarshal([]byte(claims.AddressJSON), &addr) == nil {
			id.Address = &addr
		}
	}

	return id, nil
}

// CreateSession serialises the Identity into a signed JWT cookie.
// Implements samlsp.Session.
func (s *cookieSessionStore) CreateSession(w http.ResponseWriter, r *http.Request, assertion *saml.Assertion) error {
	identity, err := extractIdentity(assertion)
	if err != nil {
		return fmt.Errorf("realme: extract identity from assertion: %w", err)
	}

	claims := identityClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.cfg.sessionMaxAge()) * time.Second)),
		},
		FLT:            identity.FLT,
		AssuranceLevel: identity.AssuranceLevel,
		SessionIndex:   identity.SessionIndex,
		FullName:       identity.FullName,
		PlaceOfBirth:   identity.PlaceOfBirth,
		Gender:         identity.Gender,
	}
	if !identity.DateOfBirth.IsZero() {
		claims.DOB = identity.DateOfBirth.Format("2006-01-02")
	}
	if identity.Address != nil {
		b, _ := json.Marshal(identity.Address)
		claims.AddressJSON = string(b)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.sigKey)
	if err != nil {
		return fmt.Errorf("realme: sign session JWT: %w", err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     s.cfg.sessionCookieName(),
		Value:    signed,
		Path:     "/",
		MaxAge:   s.cfg.sessionMaxAge(),
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

// DeleteSession clears the session cookie.
// Implements samlsp.Session.
func (s *cookieSessionStore) DeleteSession(w http.ResponseWriter, r *http.Request) error {
	http.SetCookie(w, &http.Cookie{
		Name:     s.cfg.sessionCookieName(),
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

// Identity implements samlsp.Session so that *Identity can be used directly
// as a session value where the samlsp.Session interface is expected.
func (i *Identity) GetAttributes() samlsp.Attributes {
	return samlsp.Attributes{
		"flt":            []string{i.FLT},
		"assuranceLevel": []string{i.AssuranceLevel.String()},
	}
}
