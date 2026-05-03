# URL Shortener

A production-ready URL shortener service built with Go, following senior-level architecture patterns.

## Features

- **CRUD Operations**: Create, Read, Update, Delete URLs
- **Redirect**: Short codes redirect to original URLs (302 Found)
- **Pagination**: List URLs with page/per_page support
- **Validation**: URL format validation (http/https required)
- **Graceful Shutdown**: Proper server cleanup on signals
- **Health Check**: `/health` endpoint for monitoring
- **API Key Authentication**: Secure all API endpoints
- **Rate Limiting**: 100 req/min per IP
- **OpenAPI Docs**: `/docs/openapi.yaml`
- **Docker Support**: Run anywhere with Docker

## Architecture

```
cmd/server/main.go       # Entry point
internal/
├── config/             # Configuration management
├── models/             # DTOs and domain models
├── repository/         # Database layer (SQLite)
├── service/            # Business logic
├── handler/            # HTTP handlers
├── middleware/         # Logging, recovery, auth, rate limit
├── errors/             # Custom error types
pkg/validator/         # URL validation
docs/                  # OpenAPI specification
```

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/health` | Health check |
| `GET` | `/metrics` | Prometheus metrics |
| `POST` | `/api/urls` | Create short URL |
| `GET` | `/api/urls/{id}` | Get URL by ID |
| `PUT` | `/api/urls/{id}` | Update URL |
| `DELETE` | `/api/urls/{id}` | Delete URL |
| `GET` | `/api/urls?page=1&per_page=10` | List URLs |
| `GET` | `/{shortCode}` | Redirect to URL |

## Quick Start

### Local

```bash
# Install dependencies
go mod download

# Run server
go run ./cmd/server

# Or build and run
make build
./bin/server
```

### Docker

```bash
docker-compose up -d
```

## Authentication

All API endpoints (except `/health`, `/metrics` and `/{shortCode}`) require an API key:

```bash
# Header
curl -X POST http://localhost:8080/api/urls \
  -H "X-API-Key: your-secure-api-key" \
  -H "Content-Type: application/json" \
  -d '{"original": "https://google.com"}'

# Or query param
curl "http://localhost:8080/api/urls?api_key=your-secure-api-key"
```

## Usage

### Create URL

```bash
curl -X POST http://localhost:8080/api/urls \
  -H "X-API-Key: your-secure-api-key" \
  -H "Content-Type: application/json" \
  -d '{"original": "https://google.com"}'
```

Response:
```json
{
  "id": "9f8e7d6c5b4a39281706f5e4d3c2b1a0",
  "original": "https://google.com",
  "short_code": "a1b2c3",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### Redirect

```bash
curl -I http://localhost:8080/a1b2c3
# HTTP/1.1 302 Found
# Location: https://google.com
```

### List URLs

```bash
curl "http://localhost:8080/api/urls?page=1&per_page=10" \
  -H "X-API-Key: your-secure-api-key"
```

## Environment Variables

| Variable | Default | Description |
|-----------|---------|-------------|
| `SERVER_PORT` | `8080` | Server port |
| `DB_HOST` | `./data` | Database directory |
| `DB_NAME` | `database.sqlite` | Database filename |
| `API_KEY` | `default-key` | API authentication key |

## Testing

```bash
# Run all tests
make test

# Run with coverage
go test -cover ./...
```

## OpenAPI Documentation

The OpenAPI specification is available at `docs/openapi.yaml`. You can view it with:
- Swagger Editor: https://editor.swagger.io
- VS Code: OpenAPI extension

## Tech Stack

- **Go 1.26** - Programming language
- **SQLite** - Embedded database
- **slog** - Structured logging
- **stdlib** - HTTP server

## Best Practices Demonstrated

- Dependency Injection via interfaces
- Context propagation for cancellation
- Proper error handling
- Graceful shutdown
- Structured logging
- Unit + HTTP tests with mocks
- API Key authentication
- Rate limiting middleware
- Prometheus metrics
- Custom error types
- Docker multi-stage build
- GitHub Actions CI/CD
- OpenAPI documentation
