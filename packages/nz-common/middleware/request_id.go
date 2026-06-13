package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type contextKeyType string

const requestIDKey contextKeyType = "request_id"

// RequestID injects a unique X-Request-ID header and context value into
// every request. If the incoming request already has an X-Request-ID header
// it is propagated unchanged (trusting upstream proxies).
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = uuid.New().String()
		}
		ctx := context.WithValue(r.Context(), requestIDKey, id)
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequestIDFromContext retrieves the request ID from context.
// Returns "" if no request ID is present.
func RequestIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(requestIDKey).(string)
	return v
}
