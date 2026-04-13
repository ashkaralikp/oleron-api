# Clinic API()

Minimal Go API starter for the clinic project.

## Available endpoint

- `GET /health`

Example response:

```json
{
  "success": true,
  "data": {
    "status": "ok",
    "service": "clinic-api",
    "environment": "development"
  }
}
```

## Run locally

1. Copy `.env.example` values into `.env` if needed.
2. Start the API:

```bash
go run cmd/api/main.go
```

3. Test the health endpoint:

```bash
curl http://localhost:8080/health
```


# Project Structure
clinic-api/
│
├── cmd/
│   └── api/
│       └── main.go                 ← Entry point
│
├── internal/
│   ├── config/
│   │   └── config.go               ← Env vars, app config
│   │
│   ├── database/
│   │   ├── database.go             ← DB connection, pool setup
│   │   └── migrations/
│   │       ├── 001_create_users.sql
│   │       ├── 002_create_patients.sql
│   │       ├── 003_create_billing.sql
│   │       └── 004_create_appointments.sql
│   │
│   ├── middleware/
│   │   ├── auth.go                 ← JWT validation
│   │   ├── apikey.go               ← API key check (for mobile)
│   │   ├── cors.go                 ← CORS headers
│   │   ├── logger.go               ← Request logging
│   │   └── ratelimit.go            ← Rate limiting
│   │
│   ├── modules/
│   │   │
│   │   ├── auth/
│   │   │   ├── handler.go          ← HTTP handlers
│   │   │   ├── service.go          ← Business logic
│   │   │   ├── repository.go       ← DB queries
│   │   │   └── dto.go              ← Request/Response structs
│   │   │
│   │   ├── patient/
│   │   │   ├── handler.go
│   │   │   ├── service.go
│   │   │   ├── repository.go
│   │   │   └── dto.go
│   │   │
│   │   ├── billing/
│   │   │   ├── handler.go
│   │   │   ├── service.go
│   │   │   ├── repository.go
│   │   │   └── dto.go
│   │   │
│   │   ├── appointment/
│   │   │   ├── handler.go
│   │   │   ├── service.go
│   │   │   ├── repository.go
│   │   │   └── dto.go
│   │   │
│   │   ├── doctor/
│   │   │   ├── handler.go
│   │   │   ├── service.go
│   │   │   ├── repository.go
│   │   │   └── dto.go
│   │   │
│   │   └── report/
│   │       ├── handler.go
│   │       ├── service.go
│   │       ├── repository.go
│   │       └── dto.go
│   │
│   ├── models/
│   │   ├── user.go                 ← DB model structs
│   │   ├── patient.go
│   │   ├── billing.go
│   │   ├── appointment.go
│   │   └── doctor.go
│   │
│   ├── router/
│   │   └── router.go               ← All routes registered here
│   │
│   └── server/
│       └── server.go               ← HTTP server setup
│
├── pkg/
│   ├── jwt/
│   │   └── jwt.go                  ← JWT generate & validate
│   ├── hash/
│   │   └── hash.go                 ← Password hashing (bcrypt)
│   ├── response/
│   │   └── response.go             ← Standard API response helpers
│   └── validator/
│       └── validator.go            ← Request validation helpers
│
├── docker/
│   ├── Dockerfile
│   └── Dockerfile.dev
│
├── .env
├── .env.example
├── docker-compose.yml
├── go.mod
├── go.sum
└── Makefile


# Go Libraries For API
go get github.com/go-chi/chi/v5          # Router
go get github.com/jackc/pgx/v5           # PostgreSQL driver
go get github.com/golang-jwt/jwt/v5      # JWT
go get github.com/joho/godotenv          # .env loading
go get golang.org/x/crypto               # bcrypt
go get github.com/go-playground/validator/v10  # Validation
