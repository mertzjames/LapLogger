# API Reference

Base URL: `http://localhost:8080/api` (direct) or `http://localhost/api` (via nginx)

All authenticated endpoints require the `Authorization: Bearer <token>` header.

---

## Health

### `GET /api/health`

Returns the server health status. No authentication required.

**Response** `200 OK`

```json
{
  "status": "ok"
}
```

---

## Authentication

### `GET /api/auth/google`

Redirects the user to Google's OAuth 2.0 consent screen. No authentication required.

**Response** `307 Temporary Redirect`

Redirects to `https://accounts.google.com/o/oauth2/v2/auth?...`

---

### `GET /api/auth/google/callback`

Handles the OAuth callback from Google. Exchanges the authorization code for user info, upserts the user in the database, and issues a JWT.

**Query Parameters**

| Parameter | Type | Required | Description |
|---|---|---|---|
| `code` | string | Yes | Authorization code from Google |

**Response** `307 Temporary Redirect`

Redirects to `/?token=<jwt>` on success.

**Error Responses**

| Status | Body | Condition |
|---|---|---|
| 400 | `{"error": "missing code parameter"}` | No `code` query parameter |
| 401 | `{"error": "oauth exchange failed"}` | Invalid or expired authorization code |
| 500 | `{"error": "failed to get user info"}` | Could not fetch user info from Google |
| 500 | `{"error": "failed to decode user info"}` | Malformed response from Google |
| 500 | `{"error": "failed to save user"}` | Database error upserting user |
| 500 | `{"error": "failed to issue token"}` | JWT signing error |

---

### `GET /api/me`

Returns the authenticated user's profile. **Requires authentication.**

**Request Headers**

```
Authorization: Bearer <jwt>
```

**Response** `200 OK`

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "google_id": "117730123456789",
  "email": "coach@example.com",
  "name": "Coach Smith",
  "created_at": "2026-04-15T10:30:00Z"
}
```

**Error Responses**

| Status | Body | Condition |
|---|---|---|
| 401 | `{"error": "not authenticated"}` | Missing or invalid JWT |
| 404 | `{"error": "user not found"}` | User ID in JWT doesn't match any record |
| 500 | `{"error": "failed to get user"}` | Database error |

---

## Leagues

### `POST /api/leagues`

Creates a new league (tenant). Provisions an isolated PostgreSQL database and adds the requesting user as the league admin. **Requires authentication.**

**Request Headers**

```
Authorization: Bearer <jwt>
Content-Type: application/json
```

**Request Body**

```json
{
  "name": "Metro Swim League"
}
```

| Field | Type | Required | Constraints |
|---|---|---|---|
| `name` | string | Yes | Must be non-empty after trimming; must contain at least one alphanumeric character |

**Response** `201 Created`

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Metro Swim League",
  "slug": "metro-swim-league",
  "created_at": "2026-04-15T10:30:00Z"
}
```

Note: The `db_name` field is intentionally excluded from the JSON response (`json:"-"`).

**Slug Generation**

The slug is derived from the league name:
1. Convert to lowercase
2. Replace non-alphanumeric sequences with hyphens
3. Trim leading/trailing hyphens

Example: `"Metro Swim League!!!"` → `"metro-swim-league"`

**Side Effects**

1. Creates a `leagues` record in the control database
2. Creates a new PostgreSQL database named `laplogger_league_<uuid>` (with hyphens replaced by underscores)
3. Runs tenant migrations on the new database (creates `teams`, `swimmers`, `meets`, `events`, `times` tables)
4. Creates a `league_memberships` record linking the user as `admin`

**Error Responses**

| Status | Body | Condition |
|---|---|---|
| 400 | `{"error": "name is required"}` | Missing or invalid JSON body |
| 400 | `{"error": "name cannot be empty"}` | Name is empty or whitespace only |
| 400 | `{"error": "name must contain alphanumeric characters"}` | Name has no valid slug characters |
| 401 | `{"error": "missing authorization header"}` | No JWT provided |
| 409 | `{"error": "league name or slug already taken"}` | Duplicate name or slug |
| 500 | `{"error": "failed to configure league"}` | Database update error |
| 500 | `{"error": "failed to provision league database"}` | CREATE DATABASE or migration error |
| 500 | `{"error": "failed to add membership"}` | Membership insert error |

---

### `GET /api/leagues`

Returns all leagues the authenticated user belongs to, ordered by name. **Requires authentication.**

**Request Headers**

```
Authorization: Bearer <jwt>
```

**Response** `200 OK`

```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Metro Swim League",
    "slug": "metro-swim-league",
    "created_at": "2026-04-15T10:30:00Z"
  },
  {
    "id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
    "name": "Summer Splash League",
    "slug": "summer-splash-league",
    "created_at": "2026-04-16T14:00:00Z"
  }
]
```

Returns an empty array `[]` if the user has no leagues.

**Error Responses**

| Status | Body | Condition |
|---|---|---|
| 401 | `{"error": "missing authorization header"}` | No JWT provided |
| 500 | `{"error": "failed to list leagues"}` | Database error |

---

## Authentication Details

### JWT Token

All authenticated endpoints validate the JWT using the `AuthRequired` middleware.

**Token Format**

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Token Claims**

| Claim | Type | Description |
|---|---|---|
| `sub` | string | User UUID (primary key in `users` table) |
| `email` | string | User's email from Google |
| `name` | string | User's display name from Google |
| `iat` | number | Issued-at timestamp (Unix seconds) |
| `exp` | number | Expiration timestamp (Unix seconds, 24h after issue) |

**Common Auth Error Responses**

| Status | Body | Condition |
|---|---|---|
| 401 | `{"error": "missing authorization header"}` | No `Authorization` header |
| 401 | `{"error": "invalid authorization format"}` | Header doesn't match `Bearer <token>` |
| 401 | `{"error": "invalid or expired token"}` | Token signature invalid, expired, or wrong algorithm |
| 401 | `{"error": "invalid token claims"}` | Claims couldn't be parsed |
| 401 | `{"error": "missing subject in token"}` | No `sub` claim in token |

---

## Tenant-Scoped Routes (Phase 4+)

Routes under `/api/leagues/:leagueId/` are protected by both `AuthRequired` and `TenantRequired` middleware. The tenant middleware:

1. Validates the `:leagueId` URL parameter
2. Verifies the authenticated user is a member of the league
3. Resolves the league's tenant database
4. Attaches the tenant `*sql.DB` to the request context

**Additional Error Responses (TenantRequired)**

| Status | Body | Condition |
|---|---|---|
| 400 | `{"error": "missing league id"}` | Empty `:leagueId` parameter |
| 403 | `{"error": "not a member of this league"}` | User is not in `league_memberships` for this league |
| 404 | `{"error": "league not found"}` | No league with this ID exists |
| 500 | `{"error": "membership check failed"}` | Database error during membership lookup |
| 500 | `{"error": "failed to resolve league"}` | Database error loading league record |
| 500 | `{"error": "failed to connect to league database"}` | Could not open connection to tenant DB |

CRUD endpoints for teams, swimmers, meets, events, and times are documented below.

---

## Teams

All team endpoints are under `/api/leagues/:leagueId/teams`. Require authentication + league membership.

### `GET /api/leagues/:leagueId/teams`

Returns all teams in the league, ordered by name.

**Response** `200 OK` — Array of team objects.

### `GET /api/leagues/:leagueId/teams/:teamId`

Returns a single team.

**Response** `200 OK` — Team object. `404` if not found. `400` if invalid UUID.

### `POST /api/leagues/:leagueId/teams`

Creates a new team.

| Field | Type | Required | Constraints |
|---|---|---|---|
| `name` | string | Yes | Non-empty, max 255 chars |
| `short_name` | string | No | Optional abbreviation |

**Response** `201 Created` — Created team object.

### `PUT /api/leagues/:leagueId/teams/:teamId`

Updates a team. Only provided fields are updated.

| Field | Type | Required |
|---|---|---|
| `name` | string | No |
| `short_name` | string | No |

**Response** `200 OK` — Updated team object. `404` if not found.

### `DELETE /api/leagues/:leagueId/teams/:teamId`

Deletes a team. **Response** `204 No Content`. `404` if not found.

---

## Swimmers

All swimmer endpoints are under `/api/leagues/:leagueId/swimmers`.

### `GET /api/leagues/:leagueId/swimmers`

Returns all swimmers. Optionally filter by `?team_id=<uuid>`.

**Response** `200 OK` — Array of swimmer objects ordered by last_name, first_name.

### `GET /api/leagues/:leagueId/swimmers/:swimmerId`

Returns a single swimmer. `404` if not found. `400` if invalid UUID.

### `POST /api/leagues/:leagueId/swimmers`

Creates a new swimmer.

| Field | Type | Required | Constraints |
|---|---|---|---|
| `team_id` | string (UUID) | Yes | Must reference existing team |
| `first_name` | string | Yes | Non-empty, max 255 chars |
| `last_name` | string | Yes | Non-empty, max 255 chars |
| `date_of_birth` | string | Yes | Format: `YYYY-MM-DD` |
| `gender` | string | Yes | `M` or `F` (case-insensitive) |

**Response** `201 Created` — Created swimmer object.

### `PUT /api/leagues/:leagueId/swimmers/:swimmerId`

Updates a swimmer. Only provided fields are updated. Same constraints as create.

**Response** `200 OK` — Updated swimmer object.

### `DELETE /api/leagues/:leagueId/swimmers/:swimmerId`

Deletes a swimmer. **Response** `204 No Content`.

---

## Meets

All meet endpoints are under `/api/leagues/:leagueId/meets`.

### `GET /api/leagues/:leagueId/meets`

Returns all meets, ordered by meet_date descending.

### `GET /api/leagues/:leagueId/meets/:meetId`

Returns a single meet. `404` if not found.

### `POST /api/leagues/:leagueId/meets`

| Field | Type | Required | Constraints |
|---|---|---|---|
| `name` | string | Yes | Non-empty, max 255 chars |
| `location` | string | No | Max 500 chars |
| `meet_date` | string | Yes | Format: `YYYY-MM-DD` |
| `is_public` | boolean | No | Defaults to `false` |

**Response** `201 Created`

### `PUT /api/leagues/:leagueId/meets/:meetId`

Updates a meet. Only provided fields are updated.

### `DELETE /api/leagues/:leagueId/meets/:meetId`

Deletes a meet. Cascades to events and times via FK constraints.

---

## Events

Events are scoped to meets: `/api/leagues/:leagueId/meets/:meetId/events`.

### `GET .../events`

Returns all events for a meet, ordered by sort_order.

### `GET .../events/:eventId`

Returns a single event. `404` if not found or doesn't belong to meet.

### `POST .../events`

| Field | Type | Required | Constraints |
|---|---|---|---|
| `stroke` | string | Yes | `Free`, `Back`, `Breast`, `Fly`, `IM` (unless `is_custom`) |
| `distance` | int | Yes | Must be positive |
| `unit` | string | No | `yards` (default) or `meters` |
| `gender` | string | No | `M`, `F`, or `X` (default) |
| `age_group` | string | No | Defaults to `Open` |
| `is_custom` | boolean | No | Bypasses stroke validation |
| `custom_name` | string | No | Name for custom events |
| `sort_order` | int | No | Display order |

**Response** `201 Created`

### `PUT .../events/:eventId`

Updates an event. Only provided fields are updated.

### `DELETE .../events/:eventId`

Deletes an event. Cascades to times.

---

## Times (Time Entries)

Times are scoped to events: `/api/leagues/:leagueId/meets/:meetId/events/:eventId/times`.

### `GET .../times`

Returns all times for an event, ordered by time_hundredths ascending (fastest first).

### `GET .../times/:timeId`

Returns a single time entry.

### `POST .../times`

| Field | Type | Required | Constraints |
|---|---|---|---|
| `swimmer_id` | string (UUID) | Yes | Must reference existing swimmer |
| `time_hundredths` | int | Yes | Must be positive (e.g., 6523 = 1:05.23) |
| `is_exhibition` | boolean | No | Defaults to `false` |

**Response** `201 Created`. Returns `409 Conflict` if swimmer already has a time in this event.

### `PUT .../times/:timeId`

Updates a time entry. Only provided fields are updated.

### `DELETE .../times/:timeId`

Deletes a time entry. **Response** `204 No Content`.

---

## Common Validation Errors

All CRUD endpoints return `400 Bad Request` for:
- Invalid UUID in path parameters
- Missing required fields
- Invalid enum values (gender, stroke, unit)
- Invalid date format (must be `YYYY-MM-DD`)
- String fields exceeding length limits (names: 255, location: 500)
