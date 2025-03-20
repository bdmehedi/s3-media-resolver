package main

import (
	"net/http"
)

// AuthMiddleware validates the token before processing the request
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if token != appToken {
			http.Error(w, "Invalid token", http.StatusForbidden)
			return
		}
		next(w, r)
	}
}