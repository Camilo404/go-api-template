package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
)

// Recovery catches panics from any downstream handler or middleware
// and returns a generic 500 response. The panic and stack trace are
// logged so operators can diagnose the issue without taking the
// service down.
func Recovery(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger.Error("panic_recovered",
						slog.Any("error", rec),
						slog.String("path", r.URL.Path),
						slog.String("request_id", RequestIDFromContext(r.Context())),
						slog.String("stack", string(debug.Stack())),
					)
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = w.Write([]byte(`{"code":"internal_error","message":"internal server error"}`))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
