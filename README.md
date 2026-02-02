# Chirpy

A lightweight Twitter-like API built with Go and PostgreSQL.

## Overview

Chirpy is a simple social media API that allows users to:
- Create accounts with email and password
- Post short messages ("chirps")
- View all chirps
- JWT authentication for protected endpoints

## Tech Stack

- **Language:** Go 1.24+
- **Database:** PostgreSQL
- **SQL Generation:** SQLC
- **Migrations:** Goose
- **Authentication:** JWT with bcrypt password hashing

## Project Structure

```
.
├── internal/
│   ├── api/           # Shared API utilities (config, responses)
│   ├── auth/          # Authentication (JWT, password hashing)
│   ├── database/      # SQLC generated code
│   └── handlers/      # HTTP handlers
├── sql/
│   ├── queries/       # SQL queries for SQLC
│   └── schema/        # Database migrations
├── main.go            # Application entry point
└── go.mod             # Dependencies
```

## Setup

### Prerequisites

- Go 1.24 or later
- PostgreSQL database
- Environment variables configured

### Environment Variables

Create a `.env` file with:

```
DB_URL=postgres://username:password@localhost:5432/chirpy?sslmode=disable
PLATFORM=dev
SECRET=your-secret-key-for-jwt-signing
```

### Database Setup

1. Create the database:
```bash
createdb chirpy
```

2. Run migrations:
```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
goose -dir sql/schema postgres "${DB_URL}" up
```

3. Generate SQLC code (if needed):
```bash
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
sqlc generate
```

### Running the Server

```bash
go build -o out
./out
```

The server will start on port 8080.

## API Endpoints

### Health Check
- `GET /api/healthz` - Check server status

### Users
- `POST /api/users` - Create new user
- `POST /api/login` - Authenticate user, receive JWT

### Chirps
- `GET /api/chirps` - List all chirps
- `GET /api/chirps/{id}` - Get specific chirp
- `POST /api/chirps` - Create chirp (requires authentication)

### Admin
- `GET /admin/metrics` - View file server hit count
- `POST /admin/reset` - Reset users table (dev only)

### Static Files
- `GET /app/*` - Serve static files with metrics tracking

## Request Examples

### Create User
```bash
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "secret123"}'
```

### Login
```bash
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "secret123"}'
```

### Create Chirp (with JWT)
```bash
curl -X POST http://localhost:8080/api/chirps \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <your-jwt-token>" \
  -d '{"body": "Hello, Chirpy!"}'
```

## Testing

Run the tests:
```bash
go test ./...
```

## Security Notes

- Passwords are hashed using bcrypt with a cost of 10
- JWT tokens expire after 1 hour by default
- The `/admin/reset` endpoint is only available in dev environment
