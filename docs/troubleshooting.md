# Troubleshooting

This guide covers common setup, auth, Docker, and runtime issues in LapLogger.

## Docker Compose Starts but nginx Is Unreachable

Symptoms:
- `http://localhost` does not load
- `docker compose ps` shows nginx unhealthy or restarting

Checks:

```bash
docker compose ps
docker compose logs nginx --tail=100
```

Fixes:
- Make sure port `80` is free on the host
- Verify the stack was started from the repo root
- Rebuild the stack with `docker compose up -d --build`

## Backend Healthy, Frontend Missing

Symptoms:
- `http://localhost/api/health` works
- The root page fails or returns gateway errors

Checks:

```bash
docker compose ps
docker compose logs frontend --tail=100
```

Fixes:
- Rebuild the frontend image after changing frontend code: `docker compose up -d --build frontend`
- Confirm the frontend container is healthy before nginx starts proxying

## Google Login Fails with redirect_uri_mismatch

Symptoms:
- Google shows `redirect_uri_mismatch`

Cause:
- `GOOGLE_REDIRECT_URI` in `.env` does not match an authorized redirect URI in Google Cloud Console

Fix:
- For Docker Compose local use: `http://localhost/api/auth/google/callback`
- For direct backend + Vite local use: `http://localhost:8080/api/auth/google/callback`
- Update Google Cloud Console and `.env` so both values match exactly

## Google Login Returns to the Wrong Host

Symptoms:
- OAuth succeeds but the browser lands on the wrong URL after login

Cause:
- `FRONTEND_URL` is set incorrectly

Fix:
- Docker Compose local use: `FRONTEND_URL=http://localhost`
- Direct backend + Vite use: `FRONTEND_URL=http://localhost:5173`

## Login Succeeds but API Calls Return 401

Symptoms:
- Login completes, but the app behaves as unauthenticated

Checks:
- Open the browser devtools and inspect `sessionStorage`
- Confirm a `token` entry exists
- Check `/api/me` in the network tab

Possible causes:
- Expired JWT
- Invalid `JWT_SECRET` between backend restarts
- OAuth callback redirected to the wrong frontend host

Fixes:
- Clear session storage and sign in again
- Ensure `.env` has a stable `JWT_SECRET`
- Verify `FRONTEND_URL` and `GOOGLE_REDIRECT_URI`

## Backend Cannot Connect to PostgreSQL

Symptoms:
- Backend exits on startup with database initialization errors

Checks:

```bash
docker compose logs backend --tail=100
docker compose logs postgres --tail=100
```

Fixes:
- Ensure `POSTGRES_USER`, `POSTGRES_PASSWORD`, and `POSTGRES_DB` are valid
- If you changed database credentials after initial startup, recreate the volume or update the existing Postgres instance accordingly
- For direct local development, ensure PostgreSQL is running on the configured host/port

## OAuth Works in Docker but Not in Direct Local Development

Cause:
- `.env.example` is Docker-oriented by default

Fix:
Use these direct-local values in `.env`:

```env
GOOGLE_REDIRECT_URI=http://localhost:8080/api/auth/google/callback
FRONTEND_URL=http://localhost:5173
```

Then run:

```bash
cd backend && go run .
cd frontend && npm run dev
```

## docker compose ps Shows Unhealthy Services

Checks:

```bash
docker compose ps
docker compose logs backend --tail=100
docker compose logs frontend --tail=100
docker compose logs nginx --tail=100
```

Fixes:
- Rebuild after Dockerfile or nginx config changes
- Confirm no local process is already occupying port `80`
- Ensure Docker has enough memory assigned on macOS

## Public Results Page Returns 404

Cause:
- The meet is private or the URL is incorrect

Fix:
- Mark the meet public in the app
- Verify the URL format:
  `http://localhost/results/<league-slug>/meets/<meet-id>`

## Resetting the Stack

If the local environment is too far out of sync:

```bash
docker compose down -v
docker compose up -d --build
```

This removes persisted Postgres data. Use only when you explicitly want a fresh local environment.
