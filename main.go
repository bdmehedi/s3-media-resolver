package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
	"github.com/redis/go-redis/v9"
	"golang.org/x/time/rate"
)

var (
    ctx       = context.Background()
    cache     interface{}
    presigner *s3.PresignClient

    appToken   string
    s3Bucket   string
    s3Region   string
    s3Endpoint string
    expiry     time.Duration
    cacheDriver string
    db         *sql.DB
	limiter *rate.Limiter
)


func main() {
    loadEnv()

    initCache()
    initS3Presigner()
	initLimiter()

    http.HandleFunc("/", rootHandler)
    http.HandleFunc("/media", AuthMiddleware(mediaHandler))
    http.HandleFunc("/media/refresh", AuthMiddleware(refreshHandler))

    fmt.Println("Server running on port:", os.Getenv("SERVER_PORT"))
    log.Fatal(http.ListenAndServe(":"+os.Getenv("SERVER_PORT"), nil))
}

func loadEnv() {
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }

    appToken = os.Getenv("APP_TOKEN")
    s3Bucket = os.Getenv("S3_BUCKET")
    s3Region = os.Getenv("S3_REGION")
    s3Endpoint = os.Getenv("S3_ENDPOINT")
    cacheDriver = os.Getenv("CACHE_DRIVER")

    hours, err := strconv.Atoi(os.Getenv("CACHE_EXPIRY_HOURS"))
    if err != nil {
        hours = 24
    }
    expiry = time.Duration(hours) * time.Hour
}

func initCache() {
    if cacheDriver == "redis" {
        initRedis()
    } else if cacheDriver == "sqlite" {
        initSQLite()
    } else {
        log.Fatal("Invalid cache driver specified in .env")
    }
}

func initRedis() {
    cache = redis.NewClient(&redis.Options{
        Addr:     os.Getenv("REDIS_HOST"),
        Password: "", // optional
        DB:       0,
    })

    if err := cache.(*redis.Client).Ping(ctx).Err(); err != nil {
        log.Fatal("Failed to connect to Redis:", err)
    }
}

func initSQLite() {
    var err error
    db, err = sql.Open("sqlite3", "./cache.db")
    if err != nil {
        log.Fatal("Failed to connect to SQLite:", err)
    }

    // Create cache table if it doesn't exist
    _, err = db.Exec(`CREATE TABLE IF NOT EXISTS cache (
        key TEXT PRIMARY KEY,
        value TEXT,
        expiry INTEGER
    )`)
    if err != nil {
        log.Fatal("Failed to create SQLite table:", err)
    }
}

func initS3Presigner() {
    // Load AWS configuration
    cfg, err := config.LoadDefaultConfig(ctx,
        config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
            os.Getenv("AWS_ACCESS_KEY"), os.Getenv("AWS_SECRET_KEY"), "",
        )),
        config.WithRegion(s3Region),
        config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
            func(service, region string, options ...interface{}) (aws.Endpoint, error) {
                return aws.Endpoint{
                    URL:           s3Endpoint,
                    SigningRegion: s3Region,
                }, nil
            }),
        ),
    )
    if err != nil {
        log.Fatal("Failed to load AWS config:", err)
    }

    // Create S3 client
    client := s3.NewFromConfig(cfg)

    // Create a presigner
    presigner = s3.NewPresignClient(client)
}

func initLimiter() {
	// Get rate limiting settings from environment variables
	requestsPerSecond := os.Getenv("RATE_LIMIT_REQUESTS_PER_SECOND")
	burstSize := os.Getenv("RATE_LIMIT_BURST_SIZE")

	// Convert them to integers (with fallback values if empty)
	rateLimit, err := strconv.Atoi(requestsPerSecond)
	if err != nil {
		rateLimit = 5 // Default value
	}

	burst, err := strconv.Atoi(burstSize)
	if err != nil {
		burst = 10 // Default value
	}

	// Initialize the rate limiter with values from .env
	limiter = rate.NewLimiter(rate.Limit(rateLimit), burst)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	// Check if the rate limiter allows the request
	if !limiter.Allow() {
		// If rate limit is exceeded, respond with 429 status code
		http.Error(w, "Too many requests, please try again later.", http.StatusTooManyRequests)
		return
	}

    // Set Content-Type header for HTML response
    w.Header().Set("Content-Type", "text/html")

    // Write the response body for the root URL
    _, err := w.Write([]byte(`
        <html>
            <head>
                <title>Welcome to the Media URL Generator</title>
            </head>
            <body>
                <h1>Welcome to the Media URL Generator</h1>
                <p>This application allows you to generate temporary URLs for media files stored in an S3-compatible storage.</p>
                <p>To get started, use the endpoint <strong>/media</strong> with your token and file path:</p>
                <ul>
                    <li>Example: <strong>/media?token=your_token&path=your_file_path</strong></li>
                </ul>
                <p>For cache clearing, use the <strong>/media/refresh</strong> endpoint.</p>
            </body>
        </html>
    `))
    
    if err != nil {
        http.Error(w, "Failed to generate response", http.StatusInternalServerError)
    }
}


func mediaHandler(w http.ResponseWriter, r *http.Request) {
    token := r.URL.Query().Get("token")
    path := r.URL.Query().Get("path")
    fresh := r.URL.Query().Get("fresh")

    if token != appToken {
        http.Error(w, "Invalid token", http.StatusForbidden)
        return
    }
    if path == "" {
        http.Error(w, "Missing path", http.StatusBadRequest)
        return
    }

    cacheKey := fmt.Sprintf("media_cache:%s:%s", token, path)

    if fresh == "1" {
        clearCache(cacheKey)
    }

    url, err := getCache(cacheKey)
    if err == nil {
        http.Redirect(w, r, url, http.StatusFound)
        return
    }

    newURL, err := createPresignedURL(path)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    setCache(cacheKey, newURL)
    http.Redirect(w, r, newURL, http.StatusFound)
}

func refreshHandler(w http.ResponseWriter, r *http.Request) {
    token := r.URL.Query().Get("token")
    path := r.URL.Query().Get("path")

    if token != appToken {
        http.Error(w, "Invalid token", http.StatusForbidden)
        return
    }
    if path == "" {
        http.Error(w, "Missing path", http.StatusBadRequest)
        return
    }

    cacheKey := fmt.Sprintf("media_cache:%s:%s", token, path)
    clearCache(cacheKey)

    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"message":"Cache cleared"}`))
}

func createPresignedURL(path string) (string, error) {
    if strings.HasPrefix(path, "/") {
        path = path[1:]
    }

    input := &s3.GetObjectInput{
        Bucket: aws.String(s3Bucket),
        Key:    aws.String(path),
    }

    presigned, err := presigner.PresignGetObject(ctx, input, s3.WithPresignExpires(expiry))
    if err != nil {
        return "", err
    }

    return presigned.URL, nil
}

func getCache(key string) (string, error) {
    if cacheDriver == "redis" {
        return getRedisCache(key)
    } else if cacheDriver == "sqlite" {
        return getSQLiteCache(key)
    }
    return "", fmt.Errorf("cache driver not configured")
}

func setCache(key, value string) {
    if cacheDriver == "redis" {
        setRedisCache(key, value)
    } else if cacheDriver == "sqlite" {
        setSQLiteCache(key, value)
    }
}

func clearCache(key string) {
    if cacheDriver == "redis" {
        clearRedisCache(key)
    } else if cacheDriver == "sqlite" {
        clearSQLiteCache(key)
    }
}

func getRedisCache(key string) (string, error) {
    result, err := cache.(*redis.Client).Get(ctx, key).Result()
    if err == redis.Nil {
        return "", fmt.Errorf("cache not found")
    }
    return result, err
}

func setRedisCache(key, value string) {
    cache.(*redis.Client).Set(ctx, key, value, expiry)
}

func clearRedisCache(key string) {
    cache.(*redis.Client).Del(ctx, key)
}

func getSQLiteCache(key string) (string, error) {
    var value string
    row := db.QueryRow("SELECT value FROM cache WHERE key = ?", key)
    err := row.Scan(&value)
    if err == sql.ErrNoRows {
        return "", fmt.Errorf("cache not found")
    }
    return value, err
}

func setSQLiteCache(key, value string) {
    expiryUnix := time.Now().Add(expiry).Unix()
    _, err := db.Exec("INSERT OR REPLACE INTO cache (key, value, expiry) VALUES (?, ?, ?)", key, value, expiryUnix)
    if err != nil {
        log.Println("Failed to set SQLite cache:", err)
    }
}

func clearSQLiteCache(key string) {
    _, err := db.Exec("DELETE FROM cache WHERE key = ?", key)
    if err != nil {
        log.Println("Failed to clear SQLite cache:", err)
    }
}
