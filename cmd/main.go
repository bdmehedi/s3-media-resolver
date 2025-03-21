package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/bdmehedi/s3-media-resolver/internal/config"
	"github.com/bdmehedi/s3-media-resolver/internal/routes"
)

func main() {
	// Initialize configurations
	if err := config.Load(); err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Setup routes
	router := routes.SetupRoutes()

	// Start server
	port := config.AppConfig.ServerPort
	fmt.Printf("Server running on port: %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
