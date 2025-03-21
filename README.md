# S3 Media URL Generator

A Go service that generates presigned URLs for media files stored in S3-compatible storage with caching support.

## Features

- Generates temporary presigned URLs for S3 media files
- Supports Redis and SQLite caching
- Rate limiting protection
- Token-based authentication 
- Configurable cache expiry
- API to refresh cached URLs

## Requirements

- Go 1.23+
- Redis (optional - for Redis caching)
- SQLite (optional - for SQLite caching)
- S3-compatible storage service

## Installation

1. Clone the repository:
```bash
git clone https://github.com/bdmehedi/s3-media-resolver.git
cd s3-media-resolver
```

2. Copy the example environment file:
```bash
cp .env.example .env
```

3. Configure the environment variables in .env:
```
APP_TOKEN=your_token_here
AWS_ACCESS_KEY=your_s3_access_key
AWS_SECRET_KEY=your_s3_secret_key
S3_BUCKET=your_bucket_name
S3_REGION=your_region
S3_ENDPOINT=https://your-s3-endpoint.com
CACHE_DRIVER=sqlite  # Options: redis, sqlite
CACHE_EXPIRY_HOURS=24
REDIS_HOST=localhost:6379  # Required if using Redis
SERVER_PORT=8080
RATE_LIMIT_REQUESTS_PER_SECOND=5
RATE_LIMIT_BURST_SIZE=10
```

4. Build and run:
```bash
go build ./cmd/main.go
./main
```

## API Endpoints

### Generate Media URL
```
GET /media?token=your_token&path=path/to/file.jpg
```
- `token`: Your authentication token
- `path`: Path to media file in S3
- `fresh` (optional): Set to "1" to bypass cache

### Refresh Cached URL
```
GET /media/refresh?token=your_token&path=path/to/file.jpg
```
- `token`: Your authentication token
- `path`: Path to media file in S3

## Configuration

### Cache Drivers

The service supports two caching mechanisms:

1. **Redis**
   - Set `CACHE_DRIVER=redis`
   - Configure `REDIS_HOST`

2. **SQLite**
   - Set `CACHE_DRIVER=sqlite`
   - SQLite database will be created as `cache.db`

### Rate Limiting

Configure rate limiting using:
- `RATE_LIMIT_REQUESTS_PER_SECOND`: Maximum requests per second
- `RATE_LIMIT_BURST_SIZE`: Maximum burst size for requests

## Development

The project structure follows standard Go project layout:

```
├── cmd/
│   └── main.go           # Application entry point
├── internal/
│   ├── config/           # Configuration management
│   ├── controllers/      # HTTP request handlers
│   ├── middleware/       # HTTP middleware
│   ├── routes/           # Route definitions
│   ├── services/         # Business logic
│   └── utils/           # Utility functions
```

## License

This project is licensed under the MIT License.


## Contributing

We welcome contributions to the S3 Media URL Generator project! Here's how you can help:

### Setting Up Development Environment

1. Fork the repository
2. Clone your fork:
```bash
git clone https://github.com/YOUR_USERNAME/s3-media-resolver.git
cd s3-media-resolver
```
3. Create a new branch:
```bash
git checkout -b feature/your-feature-name
```

### Development Guidelines

1. **Code Style**
   - Follow standard Go code formatting using `gofmt`
   - Use meaningful variable and function names
   - Add comments for complex logic
   - Run `go fmt ./...` before committing

2. **Commit Messages**
   - Use clear, descriptive commit messages
   - Format: `type(scope): description`
   - Example: `feat(cache): add Redis TTL support`

### Pull Request Process

1. Update documentation if necessary
2. Ensure CI/CD pipeline passes
3. Request review from maintainers
4. Address review comments

### Report Issues

- Use GitHub Issues to report bugs
- Include steps to reproduce
- Include environment details
- Provide logs if applicable

### Feature Requests

- Open a GitHub Issue with label "enhancement"
- Describe the feature and its benefits
- Discuss implementation approach if possible
