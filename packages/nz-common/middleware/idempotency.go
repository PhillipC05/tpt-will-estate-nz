package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// IdempotencyResult is the stored response for a completed idempotent request.
type IdempotencyResult struct {
	StatusCode int             `json:"status_code"`
	Body       json.RawMessage `json:"body"`
	CreatedAt  time.Time       `json:"created_at"`
}

// Idempotency returns middleware that enforces idempotency via the
// Idempotency-Key request header. Results are persisted in PostgreSQL.
//
// The caller's app schema must contain the idempotency_keys table created by:
//
//	CREATE TABLE IF NOT EXISTS idempotency_keys (
//	    key         TEXT        PRIMARY KEY,
//	    status_code INT         NOT NULL,
//	    body        JSONB       NOT NULL,
//	    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
//	);
//
// Only POST requests are subject to idempotency checks.
func Idempotency(db *pgxpool.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				next.ServeHTTP(w, r)
				return
			}

			key := r.Header.Get("Idempotency-Key")
			if key == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Check for a previously completed result.
			if result, ok := lookupKey(r.Context(), db, key); ok {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Idempotency-Key-Replay", "true")
				w.WriteHeader(result.StatusCode)
				_, _ = w.Write(result.Body)
				return
			}

			// Capture the response so it can be stored.
			rw := &capturingWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rw, r)

			// Only persist 2xx responses.
			if rw.status >= 200 && rw.status < 300 {
				storeKey(r.Context(), db, key, rw.status, rw.body)
			}
		})
	}
}

func lookupKey(ctx context.Context, db *pgxpool.Pool, key string) (*IdempotencyResult, bool) {
	const q = `SELECT status_code, body, created_at FROM idempotency_keys WHERE key = $1`
	var r IdempotencyResult
	err := db.QueryRow(ctx, q, key).Scan(&r.StatusCode, &r.Body, &r.CreatedAt)
	if err != nil {
		return nil, false
	}
	return &r, true
}

func storeKey(ctx context.Context, db *pgxpool.Pool, key string, status int, body []byte) {
	const q = `INSERT INTO idempotency_keys (key, status_code, body) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`
	_, _ = db.Exec(ctx, q, key, status, json.RawMessage(body))
}

// capturingWriter captures the response status and body for storage.
type capturingWriter struct {
	http.ResponseWriter
	status int
	body   []byte
}

func (cw *capturingWriter) WriteHeader(status int) {
	cw.status = status
	cw.ResponseWriter.WriteHeader(status)
}

func (cw *capturingWriter) Write(b []byte) (int, error) {
	cw.body = append(cw.body, b...)
	return cw.ResponseWriter.Write(b)
}
