# mindful-minutes-api

Backend API for the mindful-minutes meditation tracking application. Built with Go, PostgreSQL, and Docker.

## Features

- User authentication via Clerk webhooks
- Session management for meditation tracking
- Analytics and streak calculations
- Dashboard data aggregation
- RESTful API with comprehensive validation

## Tech Stack

- **Language**: Go 1.21+
- **Framework**: Gin HTTP router
- **Database**: PostgreSQL 15
- **Authentication**: Clerk webhook integration
- **Containerization**: Docker & Docker Compose
- **Testing**: Go testing with testify

## Development Setup

### Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- Clerk account for authentication

### Environment Variables

Create a `.env` file in the root directory:

```bash
# Database Configuration
DATABASE_URL=postgres://mindful_user:mindful_pass@localhost:5432/mindful_minutes?sslmode=disable

# Clerk Configuration
CLERK_SECRET_KEY=your_clerk_secret_key_here

# Server Configuration
GIN_MODE=debug
PORT=8080
```

### Running the Application

1. **Clone the repository**
   ```bash
   git clone https://github.com/nitinstp23/mindful-minutes-api.git
   cd mindful-minutes-api
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Start the database**
   ```bash
   docker-compose up postgres -d
   ```

4. **Run database migrations**
   ```bash
   # Migrations are automatically applied when the database starts
   # Check docker-compose.yml for migration setup
   ```

5. **Run the application**
   ```bash
   go run cmd/server/main.go
   ```

   The API will be available at `http://localhost:8080`

### Running with Docker Compose

```bash
# Set your Clerk secret key
export CLERK_SECRET_KEY=your_clerk_secret_key_here

# Start all services
docker-compose up --build
```

## API Documentation

### Authentication

All API endpoints (except health check) require authentication via Clerk. Include the authorization header:

```bash
Authorization: Bearer <clerk_session_token>
```

### Base URL

```
http://localhost:8080
```

### Endpoints

#### Health Check

```bash
GET /health
```

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2025-07-08T10:00:00Z"
}
```

#### User Management

##### Create/Update User (Webhook)

```bash
POST /webhooks/clerk
```

**Request Body:**
```json
{
  "type": "user.created",
  "data": {
    "id": "user_clerk_id",
    "email_addresses": [
      {
        "email_address": "user@example.com"
      }
    ],
    "first_name": "John",
    "last_name": "Doe"
  }
}
```

#### Session Management

##### Create Session

```bash
POST /api/sessions
```

**Request Body:**
```json
{
  "duration_seconds": 600,
  "session_type": "mindfulness",
  "notes": "Morning meditation session"
}
```

**Response:**
```json
{
  "id": "session_id",
  "user_id": "user_ulid",
  "duration_seconds": 600,
  "session_type": "mindfulness",
  "notes": "Morning meditation session",
  "created_at": "2025-07-08T10:00:00Z"
}
```

##### Get Sessions

```bash
GET /api/sessions?limit=10&cursor=cursor_value
```

**Response:**
```json
{
  "sessions": [
    {
      "id": "session_id",
      "user_id": "user_ulid",
      "duration_seconds": 600,
      "session_type": "mindfulness",
      "notes": "Morning meditation session",
      "created_at": "2025-07-08T10:00:00Z"
    }
  ],
  "pagination": {
    "next_cursor": "next_cursor_value",
    "has_more": true
  }
}
```

##### Delete Session

```bash
DELETE /api/sessions/{session_id}
```

**Response:**
```json
{
  "message": "Session deleted successfully"
}
```

#### Analytics & Dashboard

##### Get Dashboard Data

```bash
GET /api/dashboard?year=2025&sessions=5
```

**Response:**
```json
{
  "current_streak": 5,
  "longest_streak": 12,
  "total_sessions": 45,
  "total_minutes": 2700,
  "weekly_progress": [
    {
      "date": "2025-07-01",
      "sessions": 2,
      "total_minutes": 30
    }
  ],
  "yearly_progress": [
    {
      "month": "January",
      "sessions": 15,
      "total_minutes": 450
    }
  ],
  "recent_sessions": [
    {
      "id": "session_id",
      "duration_seconds": 600,
      "session_type": "mindfulness",
      "notes": "Evening session",
      "created_at": "2025-07-08T20:00:00Z"
    }
  ]
}
```

### Session Types

Valid session types:
- `mindfulness`
- `breathing`
- `metta`
- `body_scan`
- `walking`
- `other`

### Error Responses

All endpoints return consistent error responses:

```json
{
  "error": "Error message",
  "details": "Additional error details (in development mode)"
}
```

Common HTTP status codes:
- `400` - Bad Request (validation errors)
- `401` - Unauthorized (authentication required)
- `404` - Not Found
- `500` - Internal Server Error

## Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests for specific package
go test -v ./internal/handlers

# Run specific test
go test -v ./internal/handlers -run TestCreateSession
```

### Test Coverage

```bash
# Generate coverage report
go test -cover ./...

# Generate HTML coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Database Testing

Tests use an in-memory SQLite database for fast, isolated testing. The test utilities automatically:
- Set up a fresh database for each test
- Clean up data after each test
- Migrate database schema

### Test Structure

Tests follow the pattern:
```go
func TestFunctionName(t *testing.T) {
    t.Run("return expected result when condition", func(t *testing.T) {
        // Test implementation
    })
}
```

## Code Quality

### Linting

```bash
# Run linter (requires golangci-lint)
golangci-lint run

# Format code
go fmt ./...

# Build project
go build ./...
```

### Code Organization

- `cmd/server/` - Application entry point
- `internal/handlers/` - HTTP request handlers
- `internal/services/` - Business logic
- `internal/models/` - Database models
- `internal/auth/` - Authentication middleware
- `internal/database/` - Database connection and utilities
- `internal/config/` - Configuration management
- `internal/testutils/` - Test utilities

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run tests and linting
6. Submit a pull request

## License

This project is licensed under the MIT License.
