package middleware

import (
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

// RealIP just wraps the go-chi realip middleware.
func RealIP(next http.Handler) http.Handler {
	return middleware.RealIP(next)
}
