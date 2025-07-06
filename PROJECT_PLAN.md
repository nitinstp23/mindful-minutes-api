# Mindful Minutes API - Project Plan

## Project Overview

Build a Go-based REST API backend for the mindful-minutes meditation tracking application. The API will serve a Next.js frontend that uses Clerk for authentication and requires endpoints for user management, session tracking, and analytics.

## Frontend Analysis Summary

The frontend is a Next.js 15 app with:
- Clerk authentication integration
- Dashboard with streaks, weekly/yearly progress, and session history
- Session management with different meditation types
- Analytics and statistics visualization

## Architecture

- **Language**: Go 1.21+
- **Framework**: Gin (HTTP router)
- **Database**: PostgreSQL
- **Authentication**: Clerk webhook integration
- **Containerization**: Docker & Docker Compose
- **Environment**: Environment variables for configuration

## Database Schema

### Users Table
```sql
CREATE TABLE users (
    id CHAR(26) PRIMARY KEY, -- ULID
    clerk_user_id VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) NOT NULL,
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP DEFAULT NULL
);
```

### Sessions Table
```sql
CREATE TABLE sessions (
    id SERIAL PRIMARY KEY,
    user_id CHAR(26) REFERENCES users(id) ON DELETE CASCADE, -- ULID
    duration_seconds INTEGER NOT NULL, -- seconds
    session_type VARCHAR(50) NOT NULL, -- mindfulness, breathing, metta etc.
    notes TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP DEFAULT NULL
);
```

## API Endpoints

### Authentication
- `POST /api/webhooks/clerk` - Clerk webhook for user creation/updates
- `GET /api/user/profile` - Get current user profile

### Sessions
- `GET /api/sessions` - Get user's meditation sessions (paginated)
- `POST /api/sessions` - Create new meditation session
- `DELETE /api/sessions/:id` - Delete session

### Analytics
- `GET /api/dashboard` - Get all dashboard data in one call

## Implementation Plan

### Phase 1: Project Setup ✅ COMPLETED
- [x] Initialize Go module and project structure
- [x] Set up Gin HTTP server with clean architecture
- [x] Configure PostgreSQL connection with GORM
- [x] Set up Docker and Docker Compose
- [x] Create environment configuration
- [x] Set up database migrations with ULID support
- [x] Implement Kubernetes-ready health checks (liveness/readiness)
- [x] Test complete setup with Docker containers

### Phase 2: Authentication ✅ COMPLETED
- [x] Implement Clerk webhook handler
- [x] Create user management functionality
- [x] Add JWT token validation middleware
- [x] Test user creation and authentication flow

### Phase 3: Core API - Sessions ✅ COMPLETED
- [x] Implement create sessions endpoint
- [x] Add input validation and error handling
- [x] Create session model and database operations
- [x] Test create session endpoint
- [x] Implement delete sessions endpoint
- [x] Add error handling
- [x] Implement database operations
- [x] Test delete session endpoint
- [x] Implement get sessions endpoint with cursor based pagination support
- [x] Add error handling
- [x] Implement database operations
- [x] Test get sessions endpoint

### Phase 4: Analytics & Statistics
- [ ] Implement streak calculation logic
- [ ] Implement dashboard aggregation endpoint
- [ ] Test dashboard aggregation endpoint

### Phase 5: Testing & Documentation
- [ ] Write unit tests for core functionality
- [ ] Create API documentation
- [ ] Add proper error handling and logging
- [ ] Performance testing and optimization

### Phase 6: Deployment Preparation
- [ ] Finalize Docker configuration
- [ ] Add health check endpoints
- [ ] Configure production environment variables
- [ ] Test full integration with frontend

## File Structure
```
mindful-minutes-api/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── auth/
│   ├── handlers/
│   ├── models/
│   ├── services/
│   └── database/
├── migrations/
├── docker-compose.yml
├── Dockerfile
├── go.mod
├── go.sum
└── .env.example
```

## Expected Response Formats

### Dashboard Response
```json
{
  "user": {
    "id": "uuid",
    "name": "John Doe",
    "email": "john@example.com"
  },
  "streaks": {
    "current": 5,
    "longest": 23
  },
  "weekly_progress": [
    {"day": "Mon", "date": "2025-06-17", "minutes": 20}
  ],
  "yearly_progress": [
    {"month": "Jan", "hours": 8.5, "minutes": 510}
  ],
  "recent_sessions": [
    {
      "id": 1,
      "date": "2025-06-23",
      "duration": 10,
      "type": "Mindfulness",
      "notes": "Morning session"
    }
  ]
}
```

## Success Criteria

1. ✅ All authentication flows work with Clerk
2. ✅ Frontend can create, delete, list sessions
3. ✅ Dashboard displays real data from API
4. ✅ Streak calculations are accurate
5. ✅ API is containerized and ready for deployment
6. ✅ Basic error handling and logging in place

## Notes

- Follow Go best practices and clean architecture
- Keep database queries optimized
- Implement proper error handling
- Use environment variables for configuration
- Add basic logging for debugging
- Ensure API responses match frontend expectations exactly

---

## Review

### Phase 1 Completion Summary

**Completed Features:**
- ✅ Go module with proper structure and dependencies
- ✅ Gin HTTP server with clean separation of concerns
- ✅ PostgreSQL integration with GORM ORM
- ✅ Docker containerization with multi-stage builds
- ✅ Database migrations with ULID primary keys
- ✅ Kubernetes health checks (separate liveness/readiness endpoints)
- ✅ Environment configuration with .env support
- ✅ Complete Docker Compose setup with database

**Key Technical Decisions:**
- Used ULID instead of UUID for better database performance
- Implemented health-go package for production-ready health monitoring
- Separated server logic from main.go for better testability
- Added comprehensive database indexes for optimal query performance
- Configured soft deletes with deleted_at timestamps

**Files Created:**
- `cmd/server/main.go` - Application entry point
- `internal/http/server.go` - HTTP server and routing logic
- `internal/database/db.go` - Database connection and health checks
- `internal/models/user.go` - User model with ULID support
- `internal/models/session.go` - Session model with foreign keys
- `migrations/001_create_users_table.sql` - Users table schema
- `migrations/002_create_sessions_table.sql` - Sessions table schema
- `Dockerfile` - Multi-stage container build
- `docker-compose.yml` - Development environment setup
- `.env.example` - Environment configuration template

**Endpoints Available:**
- `GET /health/liveness` - Kubernetes liveness probe
- `GET /health/readiness` - Kubernetes readiness probe with DB check
- `GET /api/ping` - Basic API health check

**Next Phase:** Ready to begin Phase 2 (Authentication) with Clerk webhook integration.

---

**Next Steps**: Begin Phase 2 implementation or await further instructions.