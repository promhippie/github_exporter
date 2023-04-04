package middleware

import (
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

// Profiler just wraps the go-chi profiler middleware.
func Profiler() http.Handler {
	return middleware.Profiler()
}
