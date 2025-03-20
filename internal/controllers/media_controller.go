package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/bdmehedi/s3-media-resolver/internal/services"
)

type MediaController struct {
    cacheService *services.CacheService
    s3Service    *services.S3Service
}

func NewMediaController() *MediaController {
    return &MediaController{
        cacheService: services.NewCacheService(),
        s3Service:    services.NewS3Service(),
    }
}

func (c *MediaController) HandleMedia(w http.ResponseWriter, r *http.Request) {
    path := r.URL.Query().Get("path")
    fresh := r.URL.Query().Get("fresh")

    if path == "" {
        http.Error(w, "Missing path", http.StatusBadRequest)
        return
    }

    cacheKey := "media_cache:" + path

    if fresh == "1" {
        if err := c.cacheService.Clear(cacheKey); err != nil {
            http.Error(w, "Failed to clear cache", http.StatusInternalServerError)
            return
        }
    }

    // Try to get URL from cache
    url, err := c.cacheService.Get(cacheKey)
    if err == nil {
        http.Redirect(w, r, url, http.StatusFound)
        return
    }

    // Generate new URL
    newURL, err := c.s3Service.CreatePresignedURL(path)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Cache the new URL
    if err := c.cacheService.Set(cacheKey, newURL); err != nil {
        // Log the error but continue as the URL is still valid
        // log.Printf("Failed to cache URL: %v", err)
    }

    http.Redirect(w, r, newURL, http.StatusFound)
}

func (c *MediaController) HandleRefresh(w http.ResponseWriter, r *http.Request) {
    path := r.URL.Query().Get("path")
    if path == "" {
        http.Error(w, "Missing path", http.StatusBadRequest)
        return
    }

    cacheKey := "media_cache:" + path
    if err := c.cacheService.Clear(cacheKey); err != nil {
        http.Error(w, "Failed to clear cache", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "message": "Cache cleared successfully",
    })
}