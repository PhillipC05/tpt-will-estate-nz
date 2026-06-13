package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims is the standard JWT claims structure used across TPT NZ apps.
type Claims struct {
	jwt.RegisteredClaims
	UserID         string `json:"uid"`
	FLT            string `json:"flt"`
	AssuranceLevel int    `json:"al"`
}

// TokenSigner creates and validates JWTs for service-to-service auth
// and short-lived sharing tokens (e.g., tenant application packs).
type TokenSigner struct {
	secret []byte
}

// NewTokenSigner creates a TokenSigner with the given HMAC secret.
func NewTokenSigner(secret string) *TokenSigner {
	return &TokenSigner{secret: []byte(secret)}
}

// Sign creates a signed JWT with the given claims and TTL.
func (ts *TokenSigner) Sign(userID, flt string, assuranceLevel int, ttl time.Duration) (string, error) {
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
		},
		UserID:         userID,
		FLT:            flt,
		AssuranceLevel: assuranceLevel,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(ts.secret)
}

// Parse validates a JWT and returns its claims.
func (ts *TokenSigner) Parse(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return ts.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("auth: parse token: %w", err)
	}
	return claims, nil
}
