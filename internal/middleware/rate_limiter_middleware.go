package middleware

import (
	"github.com/bdmehedi/s3-media-resolver/internal/config"
	"net/http"
)

func RateLimiterMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !config.AppConfig.Limiter.Allow() {
			http.Error(w, "Too many requests, please try again later.", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	}
}
