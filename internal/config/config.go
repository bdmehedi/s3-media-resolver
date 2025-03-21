package config

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"github.com/redis/go-redis/v9"
	"golang.org/x/time/rate"
)

type AppConfiguration struct {
	Context     context.Context
	AppToken    string
	ServerPort  string
	S3Bucket    string
	S3Region    string
	S3Endpoint  string
	CacheDriver string
	Expiry      time.Duration
	DB          *sql.DB
	Cache       interface{}
	Presigner   *s3.PresignClient
	Limiter     *rate.Limiter
}

var AppConfig AppConfiguration

func Load() error {
	AppConfig.Context = context.Background()

	if err := loadEnv(); err != nil {
		return err
	}

	if err := initCache(); err != nil {
		return err
	}

	if err := initS3Presigner(); err != nil {
		return err
	}

	initRateLimiter()
	return nil
}

func loadEnv() error {
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}

	AppConfig.AppToken = os.Getenv("APP_TOKEN")
	AppConfig.ServerPort = os.Getenv("SERVER_PORT")
	AppConfig.S3Bucket = os.Getenv("S3_BUCKET")
	AppConfig.S3Region = os.Getenv("S3_REGION")
	AppConfig.S3Endpoint = os.Getenv("S3_ENDPOINT")
	AppConfig.CacheDriver = os.Getenv("CACHE_DRIVER")

	hours, err := strconv.Atoi(os.Getenv("CACHE_EXPIRY_HOURS"))
	if err != nil {
		hours = 24
	}
	AppConfig.Expiry = time.Duration(hours) * time.Hour

	return nil
}

func initCache() error {
	switch AppConfig.CacheDriver {
	case "redis":
		return initRedis()
	case "sqlite":
		return initSQLite()
	default:
		return fmt.Errorf("invalid cache driver specified")
	}
}

func initRedis() error {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST"),
		Password: "",
		DB:       0,
	})

	if err := redisClient.Ping(AppConfig.Context).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	AppConfig.Cache = redisClient
	return nil
}

func initSQLite() error {
	db, err := sql.Open("sqlite3", "./cache.db")
	if err != nil {
		return fmt.Errorf("failed to connect to SQLite: %w", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS cache (
        key TEXT PRIMARY KEY,
        value TEXT,
        expiry INTEGER
    )`)
	if err != nil {
		return fmt.Errorf("failed to create SQLite table: %w", err)
	}

	AppConfig.DB = db
	AppConfig.Cache = db
	return nil
}

func initS3Presigner() error {
	cfg, err := config.LoadDefaultConfig(AppConfig.Context,
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			os.Getenv("AWS_ACCESS_KEY"),
			os.Getenv("AWS_SECRET_KEY"),
			"",
		)),
		config.WithRegion(AppConfig.S3Region),
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{
					URL:           AppConfig.S3Endpoint,
					SigningRegion: AppConfig.S3Region,
				}, nil
			}),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg)
	AppConfig.Presigner = s3.NewPresignClient(client)
	return nil
}

func initRateLimiter() {
	requestsPerSecond := os.Getenv("RATE_LIMIT_REQUESTS_PER_SECOND")
	burstSize := os.Getenv("RATE_LIMIT_BURST_SIZE")

	rateLimit, err := strconv.Atoi(requestsPerSecond)
	if err != nil {
		rateLimit = 5
	}

	burst, err := strconv.Atoi(burstSize)
	if err != nil {
		burst = 10
	}

	AppConfig.Limiter = rate.NewLimiter(rate.Limit(rateLimit), burst)
}
