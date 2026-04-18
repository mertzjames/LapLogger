# Developer Guide

This guide covers local development setup, project structure, coding conventions, and testing for LapLogger.

## Prerequisites

- **Go** 1.22+ ([install](https://go.dev/doc/install))
- **PostgreSQL** 16 ([install](https://www.postgresql.org/download/))
- **Node.js** 20+ ([install](https://nodejs.org/))
- **golangci-lint** ([install](https://golangci-lint.run/welcome/install/)) — for running lint checks
- **Docker & Docker Compose** (optional) — for containerized development

## Local Setup

### 1. Clone the Repository

```bash
git clone <repo-url>
cd LapLogger
```

### 2. Configure Environment

```bash
cp .env.example .env
```

Edit `.env` with your local values:

| Variable | Description | Dev Default |
|---|---|---|
| `POSTGRES_USER` | PostgreSQL username | `laplogger` |
| `POSTGRES_PASSWORD` | PostgreSQL password | `laplogger_dev` |
| `POSTGRES_HOST` | PostgreSQL host | `localhost` |
| `POSTGRES_PORT` | PostgreSQL port | `5432` |
| `POSTGRES_DB` | Control plane DB name | `laplogger_control` |
| `GOOGLE_CLIENT_ID` | Google OAuth client ID | (set up in Google Cloud Console) |
| `GOOGLE_CLIENT_SECRET` | Google OAuth client secret | (set up in Google Cloud Console) |
| `GOOGLE_REDIRECT_URI` | OAuth callback URL | `http://localhost/api/auth/google/callback` |
| `JWT_SECRET` | HMAC secret for signing JWTs | `dev-secret-change-in-production` |
| `BACKEND_PORT` | Port the Go server listens on | `8080` |

### 3. Set Up the Database

```bash
# Create the control plane database
createdb laplogger_control

# The backend automatically runs migrations on startup
```

### 4. Start the Backend

```bash
cd backend
go run .
```

The server starts at `http://localhost:8080`. Control plane migrations run automatically.

### 5. Start the Frontend

```bash
cd frontend
npm install
npm run dev
```

The Vite dev server starts at `http://localhost:5173`.

### Docker Compose (Alternative)

```bash
docker compose up --build
```

This starts all services (postgres, backend, frontend, nginx) with the app available at `http://localhost`.

## Project Structure

```
LapLogger/
├── .env.example              # Template for environment variables
├── .golangci.yml             # golangci-lint configuration
├── docker-compose.yml        # Docker Compose service definitions
├── plan-lapLogger.prompt.md  # Project plan and architecture decisions
│
├── backend/
│   ├── main.go               # Entry point — wires config, DB, handlers, routes
│   ├── go.mod                # Go module definition
│   │
│   ├── config/
│   │   └── config.go         # Loads configuration from environment variables
│   │
│   ├── database/
│   │   ├── control.go        # Control plane DB connection + migration runner
│   │   ├── provisioner.go    # CREATE DATABASE + tenant migration for new leagues
│   │   ├── tenant.go         # Thread-safe tenant connection pool manager
│   │   ├── queries.go        # SQL queries for users, leagues, memberships
│   │   └── migrations/
│   │       ├── control/
│   │       │   ├── 000001_init.up.sql
│   │       │   └── 000001_init.down.sql
│   │       └── tenant/
│   │           ├── 000001_init.up.sql
│   │           └── 000001_init.down.sql
│   │
│   ├── handlers/
│   │   ├── auth.go           # Google OAuth login + JWT issuance + /me endpoint
│   │   ├── league.go         # Create and list leagues (tenant management)
│   │   ├── team.go           # Team CRUD within tenant DB
│   │   ├── swimmer.go        # Swimmer CRUD within tenant DB
│   │   ├── meet.go           # Meet CRUD within tenant DB
│   │   ├── event.go          # Event CRUD (scoped to meets)
│   │   ├── time_entry.go     # Time entry CRUD (scoped to events)
│   │   └── validation.go     # Shared validation helpers (UUID, date, length)
│   │
│   ├── middleware/
│   │   ├── auth.go           # JWT validation middleware
│   │   └── tenant.go         # Tenant resolution + membership check middleware
│   │
│   └── models/
│       ├── control.go        # User, League, LeagueMembership structs
│       └── tenant.go         # Team, Swimmer, Meet, Event, TimeEntry structs
│
├── frontend/
│   ├── package.json
│   ├── src/
│   │   ├── App.js
│   │   ├── index.js
│   │   ├── components/       # React components
│   │   ├── contexts/         # React context providers
│   │   └── services/         # API service layer
│   └── public/
│
└── docs/
    ├── architecture.md       # System design and data model
    ├── developer-guide.md    # This file
    ├── api-reference.md      # REST API documentation
    ├── qa/                   # QA gate reports
    └── security/             # Security review reports
```

## Coding Conventions

### Go Backend

- **Framework**: [Gin](https://gin-gonic.com/) for HTTP routing and middleware
- **Error handling**: Wrap errors with `fmt.Errorf("context: %w", err)`. Log server-side details; return generic messages to clients.
- **SQL**: All queries use parameterized statements (`$1`, `$2`, ...) — never string interpolation. The one exception is `CREATE DATABASE`, where the database name is validated against `^[a-zA-Z0-9_]+$` before interpolation.
- **Models**: Defined in `models/` with JSON tags for API responses and `db` tags for documentation. The `DBName` field on `League` uses `json:"-"` to prevent leaking internal database names.
- **UUIDs**: Generated with `github.com/google/uuid`. All primary keys are UUIDs.
- **Migrations**: SQL files in `database/migrations/{control,tenant}/` using golang-migrate naming (`000001_init.up.sql`, `000001_init.down.sql`). Embedded via `embed.FS`.

### Naming

- Files: `snake_case.go`
- Packages: single lowercase word (`config`, `database`, `handlers`, `middleware`, `models`)
- Exported types: `PascalCase` (`ControlDB`, `TenantManager`, `AuthHandler`)
- Handler constructors: `NewXxxHandler(cfg, db)` pattern
- Test files: `xxx_test.go` (unit), `xxx_security_test.go` (security-focused)

## Running Tests

### All Tests

```bash
cd backend
go test ./...
```

### With Coverage

```bash
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out     # summary by function
go tool cover -html=coverage.out     # interactive HTML report
```

### Specific Package

```bash
go test ./handlers/...
go test ./middleware/...
go test ./database/...
```

### Verbose Output

```bash
go test -v ./...
```

### Test Organization

| File Pattern | Purpose |
|---|---|
| `*_test.go` | Unit tests — functional correctness |
| `*_security_test.go` | Security tests — injection, bypass, boundary checks |

Current test inventory (54 tests):

| Package | Tests | Coverage |
|---|---|---|
| `config` | 4 | 100% |
| `database` | 9 (5 unit + 4 security) | 5.1% |
| `handlers` | 15 (8 unit + 7 security) | 15.6% |
| `middleware` | 18 (10 unit + 8 security) | 46% |
| `models` | 7+ | — |

## Static Analysis

### golangci-lint

```bash
golangci-lint run ./...
```

Configuration is in `.golangci.yml` at the repo root. The project enforces a **zero-findings policy** — all lint issues must be resolved before merging.

### go vet

```bash
go vet ./...
```

Run as part of every QA gate.

## Adding a New Handler

1. Create `handlers/yourresource.go` with a handler struct and constructor:
   ```go
   type YourHandler struct {
       controlDB *database.ControlDB
   }

   func NewYourHandler(db *database.ControlDB) *YourHandler {
       return &YourHandler{controlDB: db}
   }
   ```

2. Add methods for each endpoint (follow existing patterns in `auth.go` and `league.go`)

3. Register routes in `main.go` under the appropriate route group

4. Add tests in `handlers/yourresource_test.go`

## Adding a Migration

1. Create new SQL files following the naming convention:
   ```
   database/migrations/{control,tenant}/000002_description.up.sql
   database/migrations/{control,tenant}/000002_description.down.sql
   ```

2. The `up.sql` runs forward; the `down.sql` reverses it

3. Migrations are embedded automatically via `embed.FS` — no code changes needed

## Google OAuth Setup

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a project or select an existing one
3. Navigate to **APIs & Services → Credentials**
4. Create an **OAuth 2.0 Client ID** (Web application)
5. Add authorized redirect URI: `http://localhost/api/auth/google/callback`
6. Copy the Client ID and Client Secret into your `.env` file
