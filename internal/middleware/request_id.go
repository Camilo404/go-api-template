package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

// Header used to read/write the request id.
const RequestIDHeader = "X-Request-ID"

type ctxKey int

const requestIDKey ctxKey = 1

// RequestID assigns or propagates a per-request identifier. The id is
// echoed back in the response header and stored in the request context
// so logs and downstream calls can be correlated.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(RequestIDHeader)
		if id == "" {
			id = newRequestID()
		}
		w.Header().Set(RequestIDHeader, id)
		ctx := context.WithValue(r.Context(), requestIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequestIDFromContext returns the request id stored in ctx, or "".
func RequestIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(requestIDKey).(string); ok {
		return v
	}
	return ""
}

func newRequestID() string {
	var b [12]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "req"
	}
	return hex.EncodeToString(b[:])
}
