// Package auth provides session management and JWT helpers for TPT NZ apps.
package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Session holds the application-level session data stored in Redis.
// FLT is the RealMe Federated Login Token — the canonical user identifier.
type Session struct {
	FLT            string    `json:"flt"`
	AssuranceLevel int       `json:"al"`
	UserID         string    `json:"uid,omitempty"` // app-internal user ID (UUID)
	CreatedAt      time.Time `json:"created_at"`
}

// SessionStore manages application sessions in Redis.
type SessionStore struct {
	rdb *redis.Client
	ttl time.Duration
}

// NewSessionStore creates a Redis-backed session store.
func NewSessionStore(rdb *redis.Client, ttl time.Duration) *SessionStore {
	return &SessionStore{rdb: rdb, ttl: ttl}
}

// Set stores a session under the given session ID.
func (s *SessionStore) Set(ctx context.Context, sessionID string, sess Session) error {
	b, err := json.Marshal(sess)
	if err != nil {
		return fmt.Errorf("auth: marshal session: %w", err)
	}
	return s.rdb.Set(ctx, "session:"+sessionID, b, s.ttl).Err()
}

// Get retrieves a session by ID. Returns (nil, nil) if not found.
func (s *SessionStore) Get(ctx context.Context, sessionID string) (*Session, error) {
	val, err := s.rdb.Get(ctx, "session:"+sessionID).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("auth: get session: %w", err)
	}
	var sess Session
	if err := json.Unmarshal(val, &sess); err != nil {
		return nil, fmt.Errorf("auth: unmarshal session: %w", err)
	}
	return &sess, nil
}

// Delete removes a session by ID.
func (s *SessionStore) Delete(ctx context.Context, sessionID string) error {
	return s.rdb.Del(ctx, "session:"+sessionID).Err()
}

// Refresh resets the session TTL.
func (s *SessionStore) Refresh(ctx context.Context, sessionID string) error {
	return s.rdb.Expire(ctx, "session:"+sessionID, s.ttl).Err()
}
