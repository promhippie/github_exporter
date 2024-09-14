package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
)

// Recoverer initializes a recoverer middleware.
func Recoverer(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rvr := recover(); rvr != nil {
					logger.Error(rvr.(string),
						"trace", string(debug.Stack()),
					)

					http.Error(
						w,
						http.StatusText(http.StatusInternalServerError),
						http.StatusInternalServerError,
					)
				}
			}()

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}
