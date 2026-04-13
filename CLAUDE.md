# CLAUDE.md — Oleron Resource Management Platform (RMP) API

## Project Overview

This is the backend API for **Oleron RMP** — a resource management system that allows organizations to manage employees, track working hours via mobile punch-in/punch-out, and calculate salaries based on logged hours.

## Domain Model

### Core Entities

- **Branch** — An organizational unit (e.g. office, site, department). Users belong to a branch.
- **Manager** (`admin` role) — Manages employees within a branch. Can view attendance and payroll for their branch.
- **Employee** — A worker who punches in/out via the mobile app. Their salary is computed from logged working hours.
- **Super Admin** — Full system access across all branches.

### Attendance & Payroll

- Employees **punch in** and **punch out** from the mobile app.
- Each punch event is recorded with a timestamp and the employee's identity.
- **Working hours** are derived from paired punch-in / punch-out records.
- **Salary is calculated** from total working hours × the employee's configured hourly/daily rate.
- Managers can view, correct, and approve attendance records for their team.

### User Roles (current DB enum)

| Role | Access |
|---|---|
| `super_admin` | Full access across all branches |
| `admin` | Branch-level full access (acts as Manager) |
| `doctor` | Legacy — will be repurposed or removed |
| `receptionist` | Legacy — will be repurposed or removed |
| `billing_staff` | Legacy — will be repurposed or removed |
| `pharmacist` | Legacy — will be repurposed or removed |

> The role enum in the DB will be updated to reflect RMP-specific roles (e.g. `manager`, `employee`) as the schema evolves.

## Tech Stack

| Layer | Technology |
|---|---|
| Language | Go 1.25 |
| Router | `go-chi/chi` v5 |
| Database | PostgreSQL (via `jackc/pgx` v5) |
| Auth | JWT (`golang-jwt/jwt` v5) + API key for mobile |
| Config | `.env` via `joho/godotenv` |
| Passwords | bcrypt via `golang.org/x/crypto` |
| Validation | `go-playground/validator` v10 |
| Container | Docker + docker-compose |

## Project Structure

```
cmd/api/main.go             ← Entry point
internal/
  config/                   ← Env-based app config
  database/
    migrations/             ← Raw SQL schema files (not numbered migrations)
  middleware/               ← auth.go, apikey.go, cors.go, logger.go, ratelimit.go
  modules/
    auth/                   ← Login, token refresh (handler / service / repo / dto)
    admin/                  ← Branch, user, menu, role-permission CRUD (super_admin only)
    myprofile/              ← Authenticated user: update profile, change password, get menus
    patient/                ← Legacy — will be replaced by employee module
    billing/                ← Legacy — will be replaced by payroll module
    appointment/            ← Legacy — will be replaced by attendance/shift module
  models/                   ← DB struct types
  router/router.go          ← All routes registered here
  server/server.go          ← HTTP server setup
pkg/
  jwt/                      ← Token generation & validation helpers
  hash/                     ← bcrypt helpers
  response/                 ← Standard JSON response wrappers
  validator/                ← Request validation helpers
```

## API Overview

Base path: `/api/v1`

All routes except `/health` require the `X-API-Key` header (mobile app key).  
Protected routes additionally require a `Authorization: Bearer <jwt>` header.

### Public (API key only)
- `POST /auth/login`
- `POST /auth/refresh`

### Authenticated
- `GET/POST/PUT/DELETE /patients` — legacy, will become `/employees`
- `GET/POST/PUT /billing` — legacy, will become `/payroll`
- `GET/POST/PUT /appointments` — legacy, will become `/attendance`
- `PUT /profile/me`
- `PATCH /profile/me/password`
- `GET /menus/me`

### Admin only (`super_admin` role)
- `/admin/branches` — CRUD
- `/admin/users` — CRUD + password reset
- `/admin/menus` — CRUD
- `/admin/role-permissions` — CRUD

## Module Pattern

Every feature module follows the same 4-file pattern:

```
handler.go      ← HTTP handlers, parse request → call service → write response
service.go      ← Business logic, orchestrates repository calls
repository.go   ← Raw SQL queries against pgxpool
dto.go          ← Request/Response structs with validation tags
```

New modules (e.g. `attendance`, `payroll`, `employee`) must follow this same pattern.

## Development

```bash
# Run locally
go run cmd/api/main.go

# Run tests
go test ./...

# Docker
make docker-up
make docker-down
```

Server defaults to `:8080`. Configure via `.env` (copy `.env.example`).

## Database

PostgreSQL. Schema is managed via raw SQL files in `internal/database/migrations/`.  
Currently two schema files exist:
- `auth&superadmin_schema.sql` — extensions, enums, branches, users, tokens, sessions, login_audit, role_permissions, menus
- `admin(branch)_schema.sql` — fee_types, branch_fee_overrides, lab_tests (to be replaced with RMP-specific tables)

The `uuid-ossp` and `pgcrypto` extensions must be enabled.  
An `update_updated_at()` trigger function is shared across tables.

## Key Conventions

- All primary keys are UUIDs (`uuid_generate_v4()`).
- Timestamps use `TIMESTAMPTZ`.
- Passwords are stored as bcrypt hashes — never plain text.
- Tokens are stored as hashes — never raw values.
- Standard JSON response shape is wrapped via `pkg/response`.
- Do not skip the API key middleware on any route.
- Role-based access is enforced via `middleware.RequireRole(...)`.
