package middleware

import (
    "net/http"
    "github.com/bdmehedi/s3-media-resolver/internal/config"
)

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        token := r.URL.Query().Get("token")
        
        if token == "" {
            http.Error(w, "Token is required", http.StatusUnauthorized)
            return
        }

        if token != config.AppConfig.AppToken {
            http.Error(w, "Invalid token", http.StatusForbidden)
            return
        }

        next.ServeHTTP(w, r)
    }
}