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

### Phase 1: Project Setup
- [ ] Initialize Go module and project structure
- [ ] Set up Gin HTTP server
- [ ] Configure PostgreSQL connection
- [ ] Set up Docker and Docker Compose
- [ ] Create environment configuration
- [ ] Set up database migrations

### Phase 2: Authentication
- [ ] Implement Clerk webhook handler
- [ ] Create user management functionality
- [ ] Add JWT token validation middleware
- [ ] Test user creation and authentication flow

### Phase 3: Core API - Sessions
- [ ] Implement create sessions endpoint
- [ ] Add input validation and error handling
- [ ] Create session model and database operations
- [ ] Test create session endpoint
- [ ] Implement delete sessions endpoint
- [ ] Add error handling
- [ ] Implement database operations
- [ ] Test delete session endpoint
- [ ] Implement get sessions endpoint with cursor based pagination support
- [ ] Add error handling
- [ ] Implement database operations
- [ ] Test get sessions endpoint

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

**Next Steps**: Review this plan and get approval before implementation begins.