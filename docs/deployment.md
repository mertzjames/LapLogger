# Deployment Guide

This guide covers Docker Compose deployment for LapLogger, required environment variables, and production rollout considerations.

## Overview

LapLogger ships as a four-service Docker Compose stack:

- `nginx` — public reverse proxy and static frontend entrypoint
- `backend` — Go API server
- `frontend` — static React bundle served by internal nginx
- `postgres` — PostgreSQL 16 database instance

Only `nginx` is published to the host. Backend and Postgres stay internal to Docker networks.

## Prerequisites

- Docker Engine with Compose support
- A Google OAuth 2.0 client configured for your target hostname
- A secure JWT secret for non-development environments
- A host with port `80` available

## Environment Variables

Create `.env` from `.env.example`:

```bash
cp .env.example .env
```

Fill in at least these values:

| Variable | Required | Description |
|---|---|---|
| `POSTGRES_USER` | Yes | PostgreSQL username |
| `POSTGRES_PASSWORD` | Yes | PostgreSQL password |
| `POSTGRES_DB` | Yes | Control-plane database name |
| `GOOGLE_CLIENT_ID` | Yes | OAuth client ID |
| `GOOGLE_CLIENT_SECRET` | Yes | OAuth client secret |
| `GOOGLE_REDIRECT_URI` | Yes | OAuth callback URL |
| `FRONTEND_URL` | Yes | Frontend URL used after OAuth callback |
| `JWT_SECRET` | Yes | Secret for signing JWTs |
| `BACKEND_PORT` | Yes | Internal backend port, defaults to `8080` |

### Recommended Values

For Docker Compose on a single host:

```env
GOOGLE_REDIRECT_URI=http://localhost/api/auth/google/callback
FRONTEND_URL=http://localhost
```

For a deployed hostname such as `https://laplogger.example.com`:

```env
GOOGLE_REDIRECT_URI=https://laplogger.example.com/api/auth/google/callback
FRONTEND_URL=https://laplogger.example.com
```

## Starting the Stack

The simplest startup path is:

```bash
./start.sh
```

Equivalent manual command:

```bash
docker compose up --build
```

Detached mode:

```bash
docker compose up -d --build
```

## Health Verification

Check service state:

```bash
docker compose ps
```

Expected result:
- `postgres` healthy
- `backend` healthy
- `frontend` healthy
- `nginx` healthy

Verify the API through nginx:

```bash
curl http://localhost/api/health
```

Expected response:

```json
{"status":"ok"}
```

## Network Topology

The stack defines two internal networks:

- `app-tier` — nginx, backend, frontend
- `database-tier` — backend, postgres

This prevents direct host access to the backend and database while still allowing nginx to proxy the application.

## Google OAuth Configuration

In Google Cloud Console, configure your OAuth client with an authorized redirect URI matching `GOOGLE_REDIRECT_URI`.

Examples:
- `http://localhost/api/auth/google/callback` for Docker Compose local use
- `https://laplogger.example.com/api/auth/google/callback` for production

Make sure `FRONTEND_URL` points to the user-facing hostname where the app should land after OAuth callback.

## Production Notes

LapLogger’s current Compose setup is suitable for MVP deployment, but production rollout should also address:

1. TLS termination at nginx or an upstream load balancer
2. `Strict-Transport-Security` after HTTPS is in place
3. Strong `JWT_SECRET` values, rotated and stored securely
4. Database backups for the `pgdata` volume
5. Resource limits and monitoring
6. Log aggregation for nginx and backend containers

## Stopping and Updating

Stop the stack:

```bash
docker compose down
```

Stop and remove volumes:

```bash
docker compose down -v
```

Rebuild after code changes:

```bash
docker compose up -d --build
```

## Backups

At minimum, back up the Docker volume backing PostgreSQL. A simple logical backup example:

```bash
docker compose exec postgres pg_dump -U "$POSTGRES_USER" "$POSTGRES_DB" > control-plane.sql
```

You should also back up tenant databases created by the application.
