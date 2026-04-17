## Plan: LapLogger — Multi-Tenant SaaS Swim Time Tracker

**Build a greenfield Go + React/TypeScript + PostgreSQL web app.** Each **League** is a tenant with its own isolated PostgreSQL database, provisioned automatically at signup. Auth is Google OAuth → JWT. Deployable via Docker Compose.

---

## Agent Roster

Each phase is executed by a designated agent. Use the agent picker or `@agent-name` to invoke them.

| Agent | Role | Invocation |
|---|---|---|
| `scrum-master` | Requirements, architecture, task decomposition, delegation | Primary orchestrator |
| `backend-go` | All Go source files — handlers, middleware, models, DB layer | Phases 1–5 |
| `frontend` | React/TypeScript — pages, components, API service layer | Phase 6 |
| `devsecops` | Dockerfiles, docker-compose, nginx, `.env` structure | Phases 0 & 7 |
| `test` | Unit + integration tests (Go `_test.go`, React Testing Library) | After each impl phase |
| `qa` | Static analysis, lint gates, acceptance review, definition-of-done sign-off | QA gates after each phase |

---

## Architecture

**Services (Docker Compose)**

| Service | Role |
|---|---|
| `backend` | Single Go service: control-plane routes + tenant-scoped API |
| `postgres` | One Postgres instance hosting `laplogger_control` + `laplogger_league_<uuid>` DBs |
| `frontend` | React/TS app served by nginx |
| `nginx` | Reverse proxy: `/api` → backend, `/` → frontend |

**Data Hierarchy**
```
League (tenant) ──► Teams ──► Swimmers
     │
     └──────────► Meets ──► Events ──► Times (Swimmer × Event × time_value)
```

**Control Plane DB** (`laplogger_control`): `users`, `leagues`, `league_memberships`
**Tenant DB** (`laplogger_league_<uuid>`): `teams`, `swimmers`, `meets`, `events`, `times`

### Data Model Detail

**Control Plane** (`laplogger_control`)
| Table | Key Fields |
|---|---|
| `users` | `id (UUID)`, `google_id`, `email`, `name`, `created_at` |
| `leagues` | `id (UUID)`, `name`, `slug (unique)`, `db_name`, `created_at` |
| `league_memberships` | `user_id`, `league_id`, `role (admin)`, `joined_at` |

**Tenant DB** (`laplogger_league_<uuid>`)
| Table | Key Fields |
|---|---|
| `teams` | `id (UUID)`, `name`, `short_name`, `created_at` |
| `swimmers` | `id (UUID)`, `team_id (FK)`, `first_name`, `last_name`, `date_of_birth`, `gender (M/F)`, `created_at` |
| `meets` | `id (UUID)`, `name`, `location`, `date`, `is_public`, `created_at` |
| `events` | `id (UUID)`, `meet_id (FK)`, `stroke`, `distance`, `unit (yards/meters)`, `gender (M/F/X)`, `age_group`, `is_custom`, `custom_name`, `sort_order` |
| `times` | `id (UUID)`, `event_id (FK)`, `swimmer_id (FK)`, `time_hundredths (INT)`, `is_exhibition`, `created_at` |

**Swim events**: Standard preset strokes (`Free`, `Back`, `Breast`, `Fly`, `IM`) × standard distances (`25`, `50`, `100`, `200`, `400`, `500`, `1650`) plus custom events defined by coaches.
**Age groups**: `8&U`, `9-10`, `11-12`, `13-14`, `15-18`, `Open` — derived from `swimmers.date_of_birth` relative to `meets.date`.
**Gender**: Events are gendered (`M`, `F`, or `X` for mixed). Separate results per gender.

---

## Steps

### Phase 0 — Project Foundation _(agent: `devsecops`)_ ✅
1. Create `docker-compose.yml` with postgres, backend, frontend, nginx
2. Initialize Go module (`go mod init github.com/user/laplogger`)
3. Initialize React + TypeScript frontend (Vite or CRA)
4. Set up `.env` structure (DB creds, Google OAuth secrets, JWT secret)

### Phase 1 — Database Layer _(agent: `backend-go`)_ ✅
5. Write control plane SQL migrations (`users`, `leagues`, `league_memberships`)
6. Write tenant SQL migrations (`teams`, `swimmers`, `meets`, `events`, `times`)
7. `database/control.go` — connect to control DB, run migrations on startup
8. `database/provisioner.go` — `CREATE DATABASE laplogger_league_<uuid>` + run tenant migrations
9. `database/tenant.go` — thread-safe connection pool map, lazy init per league_id

### Phase 2 — Auth _(agent: `backend-go`)_ ✅
10. Register Google OAuth app, configure redirect URI
11. `handlers/auth.go` — `GET /api/auth/google` (redirect) + `GET /api/auth/google/callback` (exchange code → upsert user → JWT)
12. `middleware/auth.go` — validate JWT, attach `userID` to context
13. `middleware/tenant.go` — resolve `league_id` from JWT, attach tenant DB to context

### Phase 3 — League (Tenant) Management _(agent: `backend-go`)_ ✅
14. `handlers/league.go` — `POST /api/leagues` (creates league + provisions DB), `GET /api/leagues` (user's leagues)

### QA-0 — QA Tooling Setup _(agent: `qa`)_
*Run once before Phase 4 begins.*

- Install & configure `golangci-lint` (`.golangci.yml` in repo root)
- Configure Go test coverage output (`-coverprofile`)
- Add ESLint + Prettier for frontend (`eslint.config.js`)
- Add `tsc --noEmit` check script to `frontend/package.json`
- Verify all tools run cleanly against current codebase

### QA-1 — QA Gate: Post-Phase 1 _(agent: `qa`)_
- `go vet ./...`
- `golangci-lint run ./...`
- `go test ./... -coverprofile=coverage.out`
- Review: migration SQL files, model structs, DB connection code
- Confirm: all Phase 1 acceptance criteria met

### QA-2 — QA Gate: Post-Phase 2 _(agent: `qa`)_
- `go vet ./...`
- `golangci-lint run ./...`
- `go test ./... -coverprofile=coverage.out`
- Review: auth handler, JWT middleware, tenant middleware
- Confirm: all Phase 2 acceptance criteria met

### QA-3 — QA Gate: Post-Phase 3 _(agent: `qa`)_
- `go vet ./...`
- `golangci-lint run ./...`
- `go test ./... -coverprofile=coverage.out`
- Review: league handler, provisioning flow
- Confirm: all Phase 3 acceptance criteria met

### Phase 4 — Tenant CRUD API _(agent: `backend-go`)_
*All routes under `/api/leagues/:leagueId/`, require auth + tenant middleware*

15. `handlers/team.go` — full CRUD for teams
16. `handlers/swimmer.go` — full CRUD, filterable by team
17. `handlers/meet.go` — full CRUD with `is_public` flag
18. `handlers/event.go` — full CRUD under `/meets/:meetId/events`
19. `handlers/time.go` — full CRUD under `/meets/:meetId/events/:eventId/times`

### QA-4 — QA Gate: Post-Phase 4 _(agent: `qa`)_
- `go vet ./...`
- `golangci-lint run ./...`
- `go test ./... -coverprofile=coverage.out`
- Review: all CRUD handlers, input validation, SQL injection safety
- Confirm: all Phase 4 acceptance criteria met

### Phase 5 — Public Results _(agent: `backend-go`)_
20. `handlers/results.go` — `GET /public/:leagueSlug/meets/:meetId` — returns structured results if `is_public=true`

### QA-5 — QA Gate: Post-Phase 5 _(agent: `qa`)_
- `go vet ./...`
- `golangci-lint run ./...`
- `go test ./... -coverprofile=coverage.out`
- Review: public results handler, auth bypass safety, tenant isolation
- Confirm: all Phase 5 acceptance criteria met

### Phase 6 — Frontend _(agent: `frontend`)_
21. Set up React Router v6 routes
22. `AuthContext.tsx` — Google OAuth redirect, JWT storage, `useAuth()` hook
23. `services/api.ts` — typed fetch wrapper with JWT header + league prefix
24. `Login.tsx` → `LeagueSelect.tsx` → `Dashboard.tsx`
25. `Teams.tsx` + `Swimmers.tsx` (list + add/edit/delete)
26. `Meets.tsx` + `Events.tsx` (create meets, add stroke/distance events)
27. `TimeEntry.tsx` — select swimmer + event, `MM:SS.ms` formatted input
28. `PublicResults.tsx` — unauthenticated, displays results table
29. `Layout.tsx` + `PrivateRoute.tsx`

### QA-6 — QA Gate: Post-Phase 6 _(agent: `qa`)_
- `npx eslint . --max-warnings 0`
- `npx tsc --noEmit`
- `npm test -- --coverage`
- Review: component structure, API integration, auth flow
- Confirm: all Phase 6 acceptance criteria met

### Phase 7 — Docker & Dev Setup _(agent: `devsecops`)_
30. Multi-stage `Dockerfile` for Go backend
31. `Dockerfile` for frontend (build → nginx)
32. `nginx.conf` (proxy + static file serving)
33. Finalize `docker-compose.yml` with healthchecks
34. Write `start.sh`

---

## QA Gate Process

Each QA gate follows a standard sequence. The `qa` agent runs these steps and produces a pass/fail report.

| # | Step | Tool / Command | Pass Criteria |
|---|---|---|---|
| 1 | Go vet | `go vet ./...` | Zero warnings |
| 2 | Lint (Go) | `golangci-lint run ./...` | Zero findings |
| 3 | Unit tests (Go) | `go test ./... -coverprofile=coverage.out` | All pass |
| 4 | Coverage check | `go tool cover -func=coverage.out` | ≥ 60% on new code |
| 5 | Lint (TS) | `npx eslint . --max-warnings 0` | Zero warnings |
| 6 | Type check (TS) | `npx tsc --noEmit` | Zero errors |
| 7 | Unit tests (TS) | `npm test -- --coverage` | All pass |
| 8 | Acceptance review | Manual review of phase deliverables | All criteria met |

*Steps 5–7 apply only when frontend code exists (Phase 6+).*

**Report template** (qa agent outputs after each gate):
```
## QA Gate Report: QA-N (Post-Phase N)
- go vet: PASS/FAIL
- golangci-lint: PASS/FAIL (N findings)
- go test: PASS/FAIL (N/M passed, coverage XX%)
- eslint: PASS/FAIL/N/A
- tsc: PASS/FAIL/N/A
- frontend tests: PASS/FAIL/N/A
- Acceptance: PASS/FAIL
- **Overall: PASS / BLOCKED (list blockers)**
```

---

## Verification _(agent: `qa`)_
1. `docker compose up --build` — all services healthy
2. Google OAuth login flow completes, JWT returned
3. Create a league → tenant DB provisioned, visible in `pg_database`
4. Team/swimmer CRUD → data persists and lists correctly
5. Create meet + events + times → saved correctly
6. Toggle meet public → public URL returns results without auth
7. Create second league → zero data bleed between tenant DBs
8. `go vet ./...` — zero warnings across all backend packages
9. `golangci-lint run ./...` — zero lint findings
10. `go test ./... -coverprofile=coverage.out` — all tests pass, ≥ 60% coverage on new code
11. `npx eslint . && npx tsc --noEmit` — zero frontend lint/type errors
12. QA gate reports archived for all phases (QA-0 through QA-6)

---

## Locked Decisions

| # | Decision | Choice | Rationale |
|---|---|---|---|
| D1 | Go web framework | **Gin** | Mature middleware ecosystem, structured route groups for tenant isolation |
| D2 | Go module path | `github.com/laplogger/laplogger` | Neutral placeholder; change anytime in `go.mod` |
| D3 | Frontend bundler | **Vite** | CRA unmaintained; Vite is the modern standard |
| D4 | JWT signing | **HS256** shared secret | Simplest for monolith; RS256 only needed for multi-service validation |
| D5 | Time storage | **Hundredths of a second** as `INTEGER` | Swim timing precision is hundredths; `8345` = `1:23.45` |
| D6 | Migration tool | **golang-migrate** with embedded SQL files | Lightweight, no ORM, version-controllable |
| D7 | Dev ports | Backend `:8080`, Vite `:5173`, Postgres `:5432`, Nginx `:80` | Standard defaults |
| D8 | Event model | Standard presets + custom events | Covers competitive meets and coach-defined events |
| D9 | Age groups | Derived from `date_of_birth` vs `meet.date` | `8&U`, `9-10`, `11-12`, `13-14`, `15-18`, `Open` |
| D10 | Gender | Events are gendered (`M`/`F`/`X`) | Separate results per gender |
| D11 | Google OAuth | **Placeholder credentials** for now | User will configure real credentials later |
| D12 | Go linter | **golangci-lint** | Industry standard, configurable, fast |
| D13 | Frontend linter | **ESLint** + Prettier | Standard React/TS toolchain |
| D14 | Type checking | **tsc --noEmit** | Catches type errors without emitting JS |
| D15 | Test coverage target | **≥ 60% on new code** | Pragmatic floor; raise after MVP |
| D16 | QA gate policy | **Block on lint/vet failures** | Zero-warning policy for static analysis |
| D17 | QA report format | **Markdown in PR/commit** | Lightweight, no external tooling needed |

---

## Scope Boundaries
- **In scope (MVP)**: Auth, league/team/swimmer/meet/event/time CRUD, public results page, age groups, gendered events, custom events
- **In scope (QA)**: `golangci-lint`, `go vet`, `go test` with coverage, ESLint, `tsc --noEmit`, QA gate reports after each phase
- **Out of scope (future)**: Stripe billing, coach/viewer roles, relay events, split times, push notifications, native mobile
- **Out of scope (QA)**: End-to-end browser tests (Playwright/Cypress), CI/CD pipeline, automated deploy gates
- **Deferred design choice**: Role-based access within a league (currently all members are league admin) — add after MVP validates the core workflow

---

## Technical Notes

1. **Tenant DB connection limits**: Each league uses a persistent `*sql.DB` pool. With many tenants, this could exhaust Postgres `max_connections`. Mitigation: use `pgbouncer` in the Docker Compose stack, or lazy-evict idle connections.

2. **Age group calculation**: Computed at query time from `swimmers.date_of_birth` relative to `meets.date`, not stored — avoids stale data as swimmers age.

3. **Test coverage threshold**: The 60% floor applies to *new code* introduced in each phase, not cumulative project coverage. Use `go tool cover -func=coverage.out` to spot untested functions. Raise the threshold post-MVP once the codebase stabilizes.
