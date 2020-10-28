package middleware

import (
	"context"
	"net/http"
	"time"
)

// Timeout just copies the go-chi timeout middleware.
func Timeout(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)

		defer func() {
			cancel()
			if ctx.Err() == context.DeadlineExceeded {
				w.WriteHeader(http.StatusGatewayTimeout)
			}
		}()

		next.ServeHTTP(w, r.WithContext(ctx))
	}

	return http.HandlerFunc(fn)
}
