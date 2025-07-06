# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is the backend API for the mindful-minutes application. The project is intended to be built with:
- **Language**: Golang
- **Database**: PostgreSQL
- **Containerization**: Docker

## Current State

This is a newly initialized repository with minimal structure:
- Only contains README.md, LICENSE, and .gitignore
- No Go modules, source code, or configuration files have been implemented yet
- The .gitignore is configured for Go development with standard exclusions

## Development Setup

Since this is a greenfield Go project, typical development commands will include:
- `go mod init` - Initialize Go modules (when creating the project structure)
- `go build` - Build the application
- `go test ./...` - Run all tests
- `go run main.go` - Run the application (once main.go exists)
- `docker build` - Build Docker image (once Dockerfile exists)
- `docker-compose up` - Run with database (once docker-compose.yml exists)

## Code Quality
- Go has built-in type checking, so no separate typecheck command is needed
- Make sure to add newline to the end of each file using UNIX line endings
- Run `go fmt ./...` to format the code
- Use `go build` to verify code compiles correctly
- Use `go test ./...` to run all tests and ensure code quality

## Architecture Notes

This project will likely follow standard Go API patterns:
- Main application entry point in `main.go`
- Handlers/controllers for HTTP endpoints
- Database models and migrations
- Middleware for authentication, logging, etc.
- Configuration management for database connections and environment variables

# Workflow
1. First think through the problem, read the codebase for relevant files, and write a plan to PROJECT_PLAN.md
2. The plan should have a list of todo items that you can check off as you complete them
3. Before you begin working, check in with me and I will verify the plan
4. Then, begin working on the todo items, marking them as complete as you go
5. In every step of the way just give me a high level explanation of what changes you made
6. Make every task and code change you do as simple as possible. We want to avoid making any massive or complex changes. Every change should impact as little code as possible. Everything is about simplicity
7. Finally, add a review section to the PROJECT_PLAN.md file with a summary of the changes you made and any other relevant information

# Testing Standards

## Test Structure
- Use `t.Run()` with descriptive test names that describe expected behavior
- Test names should follow the format: "return <expected_result> when <condition>"
- Examples:
  - "return bad request when signature header is missing"
  - "successfully create user when user.created event is received"
  - "return true when signature is valid"

## Test Organization
- Group related tests under a main test function (e.g., `TestVerifyClerkWebhook`)
- Use helper functions like `cleanDB()` for setup/teardown operations
- Prefer calling helper functions in each test rather than using `t.Cleanup()`

## Test Coverage Guidelines
- Only write tests for public functions (exported functions that start with uppercase)
- Do not write tests for private functions (functions that start with lowercase)
- Private functions are tested indirectly through their public function callers

## Testing Dependencies
- Use `github.com/stretchr/testify/assert` for assertions
- Use `github.com/samber/lo` for utility functions like `lo.ToPtr()`
- Create test utilities in `internal/testutils/` package for reusable functions

# Code Organization

## Function Order
- Public functions should be placed at the top of the file
- Private functions should be placed at the bottom of the file
- This helps with readability and follows Go conventions

# important-instruction-reminders
Do what has been asked; nothing more, nothing less.
NEVER create files unless they're absolutely necessary for achieving your goal.
ALWAYS prefer editing an existing file to creating a new one.
NEVER proactively create documentation files (*.md) or README files. Only create documentation files if explicitly requested by the User.
