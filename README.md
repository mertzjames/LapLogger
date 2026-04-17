# LapLogger

A multi-tenant SaaS swim time tracker built with Go, React/TypeScript, and PostgreSQL. Each **League** is an isolated tenant with its own database, provisioned automatically at creation. Authentication is handled via Google OAuth with JWT session tokens.

## Features

- **Multi-tenant architecture** — each league gets a dedicated PostgreSQL database with full data isolation
- **Google OAuth** — users sign in with their Google account; sessions managed via JWT (HS256)
- **League management** — create leagues, automatic database provisioning, membership tracking
- **Swim data model** — teams, swimmers, meets, events (standard + custom), and time entries stored in hundredths of a second
- **Age groups** — derived at query time from swimmer date of birth relative to meet date (`8&U`, `9-10`, `11-12`, `13-14`, `15-18`, `Open`)
- **Gendered events** — events are `M`, `F`, or `X` (mixed) with separate results per gender
- **Docker Compose** — full-stack deployment with PostgreSQL, Go backend, React frontend, and nginx reverse proxy

## Architecture Overview

```
┌─────────┐     ┌──────────┐     ┌──────────────────────┐
│  nginx   │────▶│  backend  │────▶│  PostgreSQL           │
│  :80     │     │  (Go/Gin) │     │                       │
└─────────┘     │  :8080    │     │  laplogger_control     │
┌─────────┐     │           │     │  laplogger_league_<id> │
│ frontend │◀───│           │     │  laplogger_league_<id> │
│  (React) │    └──────────┘     └──────────────────────┘
└─────────┘
```

- **Control plane DB** (`laplogger_control`): stores users, leagues, and memberships
- **Tenant DBs** (`laplogger_league_<uuid>`): each league's teams, swimmers, meets, events, and times

See [docs/architecture.md](docs/architecture.md) for the full design.

## Quick Start

### Prerequisites

- Go 1.22+
- PostgreSQL 16
- Node.js 20+ (for frontend)
- Docker & Docker Compose (optional, for containerized setup)

### Local Development (without Docker)

```bash
# 1. Clone and enter the repo
git clone <repo-url> && cd LapLogger

# 2. Copy environment config
cp .env.example .env
# Edit .env with your PostgreSQL credentials and Google OAuth keys

# 3. Create the control database
createdb laplogger_control

# 4. Start the backend
cd backend
go run .

# 5. Start the frontend (separate terminal)
cd frontend
npm install && npm run dev
```

### Docker Compose

```bash
cp .env.example .env   # configure credentials
docker compose up --build
```

The app will be available at `http://localhost`.

### Environment Variables

| Variable | Description | Default |
|---|---|---|
| `POSTGRES_USER` | PostgreSQL username | `laplogger` |
| `POSTGRES_PASSWORD` | PostgreSQL password | `laplogger_dev` |
| `POSTGRES_HOST` | PostgreSQL host | `localhost` |
| `POSTGRES_PORT` | PostgreSQL port | `5432` |
| `POSTGRES_DB` | Control plane database name | `laplogger_control` |
| `GOOGLE_CLIENT_ID` | Google OAuth client ID | — |
| `GOOGLE_CLIENT_SECRET` | Google OAuth client secret | — |
| `GOOGLE_REDIRECT_URI` | OAuth redirect URI | — |
| `JWT_SECRET` | Secret for signing JWTs | `dev-secret-change-in-production` |
| `BACKEND_PORT` | Backend listen port | `8080` |

See [docs/developer-guide.md](docs/developer-guide.md) for detailed setup instructions.

## Documentation

| Document | Description |
|---|---|
| [Architecture](docs/architecture.md) | System design, data model, tenant isolation |
| [Developer Guide](docs/developer-guide.md) | Local setup, project structure, testing, conventions |
| [API Reference](docs/api-reference.md) | All REST endpoints (auth, leagues, health) |
| [QA Reports](docs/qa/) | QA gate reports for each phase |
| [Security Reports](docs/security/) | Security review reports |

## Tech Stack

| Component | Technology |
|---|---|
| Backend | Go 1.22+, Gin framework |
| Frontend | React, TypeScript, Vite |
| Database | PostgreSQL 16 |
| Auth | Google OAuth 2.0, JWT (HS256) |
| Migrations | golang-migrate (embedded SQL) |
| Reverse Proxy | nginx |
| Containerization | Docker Compose |

## License

MIT

