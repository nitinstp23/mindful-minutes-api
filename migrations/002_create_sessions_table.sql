-- Create sessions table
CREATE TABLE IF NOT EXISTS sessions (
    id SERIAL PRIMARY KEY,
    user_id CHAR(26) REFERENCES users(id) ON DELETE CASCADE,
    duration_seconds INTEGER NOT NULL,
    session_type VARCHAR(50) NOT NULL,
    notes TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP DEFAULT NULL
);

-- Create index for soft deletes
CREATE INDEX IF NOT EXISTS idx_sessions_deleted_at ON sessions(deleted_at);

-- Create index for user_id for faster lookups
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);

-- Create index for created_at for date-based queries
CREATE INDEX IF NOT EXISTS idx_sessions_created_at ON sessions(created_at);