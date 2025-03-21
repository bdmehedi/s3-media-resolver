package controllers

import (
	"net/http"

	"github.com/bdmehedi/s3-media-resolver/internal/config"
)

type HomeController struct{}

func NewHomeController() *HomeController {
	return &HomeController{}
}

func (c *HomeController) HandleHome(w http.ResponseWriter, r *http.Request) {
	if !config.AppConfig.Limiter.Allow() {
		http.Error(w, "Too many requests, please try again later.", http.StatusTooManyRequests)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	_, err := w.Write([]byte(`
        <html>
            <head>
                <title>Welcome to the Media URL Generator</title>
                <style>
                    body { 
                        font-family: Arial, sans-serif; 
                        margin: 40px;
                        line-height: 1.6;
                    }
                    .container {
                        max-width: 800px;
                        margin: 0 auto;
                    }
                    code {
                        background: #f4f4f4;
                        padding: 2px 5px;
                        border-radius: 3px;
                    }
                </style>
            </head>
            <body>
                <div class="container">
                    <h1>Welcome to the Media URL Generator</h1>
                    <p>This application allows you to generate temporary URLs for media files stored in an S3-compatible storage.</p>
                    <h2>API Endpoints:</h2>
                    <ul>
                        <li>
                            <strong>Generate URL:</strong><br>
                            <code>GET /media?token=your_token&path=your_file_path</code>
                        </li>
                        <li>
                            <strong>Refresh Cache:</strong><br>
                            <code>GET /media/refresh?token=your_token&path=your_file_path</code>
                        </li>
                    </ul>
                </div>
            </body>
        </html>
    `))

	if err != nil {
		http.Error(w, "Failed to generate response", http.StatusInternalServerError)
	}
}
