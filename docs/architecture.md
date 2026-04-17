# Architecture

LapLogger is a multi-tenant SaaS application where each **League** is a fully isolated tenant with its own PostgreSQL database. This document describes the system architecture, data model, and tenant isolation design.

## System Overview

```
                     ┌──────────────────────────────────────────────┐
                     │              Docker Compose                  │
                     │                                              │
   HTTP :80          │  ┌─────────┐                                 │
  ─────────────────▶ │  │  nginx  │                                 │
                     │  └────┬────┘                                 │
                     │       │                                      │
                     │  /api ▼         /                            │
                     │  ┌─────────┐   ┌──────────┐                  │
                     │  │ backend │   │ frontend │                  │
                     │  │ Go/Gin  │   │ React/TS │                  │
                     │  │  :8080  │   │  (Vite)  │                  │
                     │  └────┬────┘   └──────────┘                  │
                     │       │                                      │
                     │       ▼                                      │
                     │  ┌─────────────────────────────────┐         │
                     │  │         PostgreSQL 16            │         │
                     │  │                                  │         │
                     │  │  laplogger_control               │         │
                     │  │  laplogger_league_<uuid_1>       │         │
                     │  │  laplogger_league_<uuid_2>       │         │
                     │  │  ...                             │         │
                     │  └─────────────────────────────────┘         │
                     └──────────────────────────────────────────────┘
```

### Services

| Service | Role | Port |
|---|---|---|
| **nginx** | Reverse proxy — routes `/api` to backend, `/` to frontend | 80 |
| **backend** | Go (Gin) API server — control plane + tenant-scoped routes | 8080 |
| **frontend** | React/TypeScript SPA served by Vite (dev) or nginx (prod) | 5173 (dev) |
| **postgres** | Single PostgreSQL 16 instance hosting all databases | 5432 |

## Multi-Tenancy Design

LapLogger uses a **database-per-tenant** isolation model. Each league gets its own PostgreSQL database (`laplogger_league_<uuid>`), ensuring complete data separation at the database level.

### Why Database-per-Tenant?

- **Strong isolation**: no risk of cross-tenant data leakage through application bugs
- **Independent migrations**: tenant schemas can evolve independently if needed
- **Simple backup/restore**: individual league databases can be backed up or restored without affecting others
- **Clear resource boundaries**: each tenant's storage and query load is independently measurable

### Tenant Lifecycle

```
1. User calls POST /api/leagues with a league name
2. Backend creates a league record in laplogger_control.leagues
3. Backend generates db_name = "laplogger_league_<uuid>"
4. Backend executes CREATE DATABASE <db_name>
5. Backend runs tenant migrations on the new database
6. Backend adds the user as league admin in league_memberships
7. League is ready for use
```

### Connection Management

The `TenantManager` (in `database/tenant.go`) maintains a thread-safe map of `*sql.DB` connection pools, keyed by database name. Connections are **lazily initialized** — the first request to a league opens the connection — and reused for subsequent requests.

- Max open connections per tenant: **10**
- Max idle connections per tenant: **2**
- Double-checked locking pattern prevents duplicate pool creation under concurrency

## Data Model

### Control Plane Database (`laplogger_control`)

The control plane stores cross-tenant data: user accounts, league definitions, and membership mappings.

```
┌──────────────┐     ┌──────────────────┐     ┌──────────────┐
│    users      │     │ league_memberships│     │   leagues     │
├──────────────┤     ├──────────────────┤     ├──────────────┤
│ id (UUID PK) │◀────│ user_id (FK)     │────▶│ id (UUID PK) │
│ google_id    │     │ league_id (FK)   │     │ name          │
│ email        │     │ role             │     │ slug (unique) │
│ name         │     │ joined_at        │     │ db_name       │
│ created_at   │     └──────────────────┘     │ created_at    │
└──────────────┘                               └──────────────┘
```

| Table | Purpose |
|---|---|
| `users` | Registered accounts (linked to Google via `google_id`) |
| `leagues` | One row per tenant — holds `slug` for URLs and `db_name` for database routing |
| `league_memberships` | Many-to-many join with `role` (currently all members are `admin`) |

### Tenant Database (`laplogger_league_<uuid>`)

Each league's database contains the swim data hierarchy:

```
┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐
│  teams   │     │ swimmers │     │  meets   │     │  events  │
├──────────┤     ├──────────┤     ├──────────┤     ├──────────┤
│ id (PK)  │◀────│ team_id  │     │ id (PK)  │◀────│ meet_id  │
│ name     │     │ id (PK)  │     │ name     │     │ id (PK)  │
│ short_   │     │ first_   │     │ location │     │ stroke   │
│   name   │     │   name   │     │ meet_date│     │ distance │
│ created_ │     │ last_    │     │ is_public│     │ unit     │
│   at     │     │   name   │     │ created_ │     │ gender   │
└──────────┘     │ dob      │     │   at     │     │ age_group│
                 │ gender   │     └──────────┘     │ is_custom│
                 │ created_ │                      │ custom_  │
                 │   at     │                      │   name   │
                 └──────────┘                      │ sort_    │
                                                   │   order  │
                      ┌──────────┐                 │ created_ │
                      │  times   │                 │   at     │
                      ├──────────┤                 └──────────┘
                      │ id (PK)  │                       ▲
                      │ event_id │───────────────────────┘
                      │ swimmer_ │
                      │   id     │
                      │ time_    │
                      │ hundredths│
                      │ is_      │
                      │ exhibition│
                      │ created_ │
                      │   at     │
                      └──────────┘
```

| Table | Purpose |
|---|---|
| `teams` | Swim teams within the league |
| `swimmers` | Individual swimmers, each belonging to a team. Gender: `M`/`F`. |
| `meets` | Swim meets with date, location, and a public visibility flag |
| `events` | Events within a meet — defined by stroke, distance, unit, gender, age group. Supports custom events. |
| `times` | A swimmer's recorded time in an event, stored as integer hundredths of a second (`8345` = `1:23.45`). Unique per `(event_id, swimmer_id)`. |

### Key Design Decisions

| Decision | Choice | Rationale |
|---|---|---|
| Time storage | Integer hundredths of a second | Swim timing precision is hundredths; avoids floating-point issues |
| Age groups | Derived at query time | Computed from `swimmers.date_of_birth` relative to `meets.date` — no stale data |
| Event model | Standard presets + custom | Covers competitive meets and coach-defined drills |
| Roles | All members are `admin` | Simplified for MVP; role-based access is a post-MVP feature |

## Authentication Flow

```
┌────────┐     ┌─────────┐     ┌────────┐     ┌──────────┐
│ Browser │     │ Backend │     │ Google │     │ Control  │
│         │────▶│         │────▶│ OAuth  │     │   DB     │
│         │     │         │◀────│        │     │          │
│         │◀────│ JWT     │     │        │     │          │
│         │     │         │────▶│userinfo│     │          │
│         │     │         │◀────│        │     │ upsert   │
│         │     │         │────▶│        │     │  user    │
└────────┘     └─────────┘     └────────┘     └──────────┘
```

1. User visits `GET /api/auth/google` — redirected to Google consent screen
2. Google redirects back to `GET /api/auth/google/callback?code=...`
3. Backend exchanges the authorization code for an access token
4. Backend fetches user info from Google's userinfo endpoint
5. Backend upserts the user in the control database
6. Backend issues a JWT (HS256, 24h expiry) containing `sub` (user ID), `email`, `name`
7. User is redirected to the frontend with the JWT

### JWT Structure

- **Algorithm**: HS256 (HMAC-SHA256)
- **Expiry**: 24 hours
- **Claims**: `sub` (user UUID), `email`, `name`, `iat`, `exp`
- **Transport**: `Authorization: Bearer <token>` header

## Request Routing & Middleware

All API requests flow through middleware in this order:

```
Request → nginx → Gin Router
                    │
                    ├── Public routes (no middleware)
                    │   ├── GET /api/health
                    │   ├── GET /api/auth/google
                    │   └── GET /api/auth/google/callback
                    │
                    ├── Authenticated routes (AuthRequired)
                    │   ├── GET /api/me
                    │   ├── POST /api/leagues
                    │   └── GET /api/leagues
                    │
                    └── Tenant-scoped routes (AuthRequired + TenantRequired)
                        └── /api/leagues/:leagueId/* (Phase 4+)
```

### Middleware Stack

| Middleware | Sets in Context | Purpose |
|---|---|---|
| `AuthRequired` | `userID`, `email`, `name` | Validates JWT from `Authorization` header; rejects unsigned/expired tokens |
| `TenantRequired` | `leagueID`, `tenantDB` | Resolves `:leagueId` param, verifies membership, attaches tenant `*sql.DB` |

## Database Migrations

Migrations use **golang-migrate** with SQL files embedded in the Go binary via `embed.FS`.

| Migration Set | Location | Applied To |
|---|---|---|
| Control plane | `database/migrations/control/` | `laplogger_control` on startup |
| Tenant | `database/migrations/tenant/` | Each `laplogger_league_<uuid>` at provisioning time |

Migrations run automatically — the control plane migrations execute at backend startup, and tenant migrations execute when a new league is created.
