package routes

import (
	"github.com/bdmehedi/s3-media-resolver/internal/controllers"
	"github.com/bdmehedi/s3-media-resolver/internal/middleware"
	"net/http"
)

func SetupRoutes() http.Handler {
	mux := http.NewServeMux()

	// Initialize controllers
	homeController := controllers.NewHomeController()
	mediaController := controllers.NewMediaController()

	// Apply middleware to handlers
	home := middleware.RateLimiterMiddleware(homeController.HandleHome)
	// media := middleware.RateLimiterMiddleware(middleware.AuthMiddleware(mediaController.HandleMedia))
	media := middleware.ChainMiddleware(mediaController.HandleMedia, middleware.AuthMiddleware, middleware.RateLimiterMiddleware)
	refresh := middleware.RateLimiterMiddleware(middleware.AuthMiddleware(mediaController.HandleRefresh))

	// Register routes
	mux.HandleFunc("/", home)
	mux.HandleFunc("/media", media)
	mux.HandleFunc("/media/refresh", refresh)

	return mux
}
