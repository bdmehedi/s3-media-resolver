package middleware

import (
	"net/http"
)

// ChainMiddleware chains multiple middleware functions together
func ChainMiddleware(handler http.HandlerFunc, middlewares ...func(http.HandlerFunc) http.HandlerFunc) http.HandlerFunc {
    for _, middleware := range middlewares {
        handler = middleware(handler)
    }
    return handler
}