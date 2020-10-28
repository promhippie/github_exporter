package middleware

import (
	"net/http"

	"github.com/go-chi/chi/middleware"
)

// RealIP just wraps the go-chi realip middleware.
func RealIP(next http.Handler) http.Handler {
	return middleware.RealIP(next)
}
