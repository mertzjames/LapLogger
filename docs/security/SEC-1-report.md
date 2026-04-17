# Security Review Report: SEC-1 (Post-Phases 1–3)

**Date**: 2026-04-17
**Reviewer**: Security Agent (automated code review + security-focused tests)
**Scope**: Database layer, authentication (OAuth + JWT), league management, tenant isolation
**Methodology**: OWASP Top 10 (2021) + multi-tenancy threat model

---

## Executive Summary

The codebase demonstrates **strong fundamentals** in SQL injection prevention and tenant isolation design. Two **HIGH** severity issues were found in the authentication flow (hardcoded OAuth state, JWT in URL), and one **MEDIUM** issue in JWT validation (missing `exp` enforcement). No **CRITICAL** vulnerabilities that would allow immediate data breach were found.

| Severity | Count | Status |
|----------|-------|--------|
| CRITICAL | 0     | —      |
| HIGH     | 2     | Must fix before production |
| MEDIUM   | 3     | Fix before Phase 6 |
| LOW      | 3     | Acceptable for MVP |
| INFO     | 2     | Documented behavior |

---

## OWASP Top 10 Assessment

### A01:2021 — Broken Access Control

**Status: PASS (with caveats)**

- Tenant middleware (`middleware/tenant.go`) correctly enforces membership check before granting access to league-scoped routes.
- `CheckMembership` uses parameterized queries — no bypass via SQL injection.
- `League.DBName` has `json:"-"` tag — internal DB name is never exposed to API consumers.
- **Caveat**: No role-based access control within a league. All members are `admin`. This is a documented deferred design decision (acceptable for MVP).

### A02:2021 — Cryptographic Failures

**Status: PASS (with findings)**

- JWT uses HS256 (HMAC-SHA256) — appropriate for a single-service architecture.
- Auth middleware rejects non-HMAC signing methods, including `none` algorithm.
- **Finding SEC-1-F01**: Default JWT secret is `dev-secret-change-in-production`. No runtime validation enforces that this is changed. See MEDIUM-1 below.

### A03:2021 — Injection

**Status: PASS**

- **SQL Injection**: All database queries use parameterized statements (`$1`, `$2`, etc.). Verified across all 8 query functions in `database/queries.go`.
- **CREATE DATABASE**: The only non-parameterizable SQL statement uses regex validation (`^[a-zA-Z0-9_]+$`) before string interpolation. Validated against 18 injection patterns in tests.
- **SSRF**: The only outbound HTTP request is to Google's userinfo endpoint (`openidconnect.googleapis.com`), constructed by the OAuth library — not user-controlled.

### A04:2021 — Insecure Design

**Status: FINDINGS**

- **Finding SEC-1-F02**: OAuth state parameter is hardcoded to `"state"`, defeating CSRF protection. See HIGH-1 below.
- **Finding SEC-1-F03**: JWT delivered via URL query parameter. See HIGH-2 below.

### A05:2021 — Security Misconfiguration

**Status: PASS (with findings)**

- `sslmode=disable` on all PostgreSQL connections. Acceptable for Docker Compose local development but must be `require` or `verify-full` in production.
- Gin runs in default mode (debug). Production should use `gin.SetMode(gin.ReleaseMode)`.

### A06:2021 — Vulnerable and Outdated Components

**Status: NOT ASSESSED** — `go mod audit` not available in this Go version. Dependencies should be reviewed before production.

### A07:2021 — Identification and Authentication Failures

**Status: FINDINGS** — See HIGH-1 and MEDIUM-2.

### A08:2021 — Software and Data Integrity Failures

**Status: PASS** — JWT signatures are validated. Migration files are embedded at compile time.

### A09:2021 — Security Logging and Monitoring Failures

**Status: PASS (minimal)** — Errors are logged via `log.Printf`. No structured logging, no audit trail. Acceptable for MVP.

### A10:2021 — Server-Side Request Forgery (SSRF)

**Status: PASS** — No user-controlled URLs are used in server-side HTTP requests.

---

## Detailed Findings

### HIGH-1: OAuth State Parameter Not Validated (CSRF)

| Field | Value |
|-------|-------|
| **Location** | `handlers/auth.go` — `GoogleLogin` (line ~50) |
| **OWASP** | A04 Insecure Design, A07 Auth Failures |
| **Impact** | An attacker can craft a link that, when clicked by a victim, associates the attacker's Google account with the victim's session (login CSRF). |
| **Test** | `TestSecurity_OAuth_HardcodedStateParam`, `TestSecurity_OAuth_CallbackNoStateValidation` |

**Issue**: The OAuth `state` parameter is hardcoded to the string `"state"`:
```go
url := h.oauthCfg.AuthCodeURL("state", oauth2.AccessTypeOffline)
```
The callback handler never validates this parameter. Any request to `/callback` with a valid Google authorization code will be accepted regardless of origin.

**Fix**:
1. Generate a cryptographically random state value per login attempt
2. Store it in a short-lived, HttpOnly, SameSite=Lax cookie
3. Validate the returned state matches the cookie before exchanging the code
4. Delete the cookie after validation

**Priority**: Fix before production. Acceptable for local development.

---

### HIGH-2: JWT Token Exposed in URL Query Parameter

| Field | Value |
|-------|-------|
| **Location** | `handlers/auth.go` — `GoogleCallback` (line ~75) |
| **OWASP** | A04 Insecure Design |
| **Impact** | JWT visible in browser history, HTTP Referer headers, proxy/CDN logs, and browser extensions with URL access. |
| **Test** | `TestSecurity_OAuth_TokenInURL` |

**Issue**: After successful OAuth, the JWT is placed in the URL:
```go
frontendURL := fmt.Sprintf("/?token=%s", jwtToken)
c.Redirect(http.StatusTemporaryRedirect, frontendURL)
```

**Fix Options**:
1. **HttpOnly cookie** (recommended): Set JWT as a `Secure; HttpOnly; SameSite=Lax` cookie. Frontend reads auth state from `/api/me`.
2. **POST with hidden form**: Return an HTML page with an auto-submitting form that POSTs the token to the frontend, which stores it in `sessionStorage`.

**Priority**: Fix in Phase 6 (Frontend) when cookie-based auth can be wired end-to-end.

---

### MEDIUM-1: No Validation of JWT Secret Strength

| Field | Value |
|-------|-------|
| **Location** | `config/config.go` — `Load()` |
| **Impact** | If `JWT_SECRET` env var is unset, the default `dev-secret-change-in-production` is used. A weak or default secret allows token forgery. |
| **Test** | `TestSecurity_JWT_EmptySecret` |

**Fix**: Add a startup check that rejects secrets shorter than 32 characters in non-development environments. Log a warning if the default is used.

---

### MEDIUM-2: JWT Tokens Without `exp` Claim Are Accepted

| Field | Value |
|-------|-------|
| **Location** | `middleware/auth.go` — `AuthRequired` |
| **Impact** | A token without an expiration claim is accepted indefinitely. If a token leaks, it can never expire. |
| **Test** | `TestSecurity_JWT_MissingExpClaim` — documents that status 200 is returned |

**Issue**: `jwt-go v5` does not reject tokens missing the `exp` claim by default. The current `jwt.Parse` call does not pass `jwt.WithExpirationRequired()`.

**Fix**: Add expiration requirement to the parser:
```go
token, err := jwt.Parse(tokenStr, keyFunc, jwt.WithExpirationRequired())
```

---

### MEDIUM-3: No DB Name Length Validation

| Field | Value |
|-------|-------|
| **Location** | `database/provisioner.go` — `ProvisionTenantDB` |
| **Impact** | PostgreSQL silently truncates identifiers longer than 63 bytes. Two UUIDs differing only after position 63 would produce the same DB name, causing data to be shared between tenants. |
| **Test** | `TestSecurity_ValidDBName_LengthLimits` |

**Fix**: Add a length check: `if len(dbName) > 63 { return error }`. The current naming scheme `laplogger_league_<uuid_with_underscores>` produces ~53-char names, so this is not currently exploitable, but should be validated defensively.

---

### LOW-1: sslmode=disable on PostgreSQL Connections

| Field | Value |
|-------|-------|
| **Location** | `database/control.go`, `database/tenant.go`, `database/provisioner.go` |
| **Impact** | Database credentials and data transmitted in plaintext between backend and PostgreSQL. |

Acceptable in Docker Compose (same host network). Must be changed for production.

---

### LOW-2: Error Messages May Leak Internal State

| Field | Value |
|-------|-------|
| **Location** | Various handlers |
| **Impact** | Error messages like `"failed to provision league database"` are generic (good), but `log.Printf` calls log full error details to stdout. In production, ensure logs are not exposed to end users. |
| **Test** | `TestSecurity_JWT_ErrorMessageLeakage` — verified no stack traces in API responses |

API responses are safe. Log output should be restricted in production.

---

### LOW-3: Hardcoded Connection Pool Sizes

| Field | Value |
|-------|-------|
| **Location** | `database/tenant.go` — `GetDB` |
| **Impact** | 10 max connections per tenant DB. With many leagues, could exhaust PostgreSQL `max_connections`. Not a security vulnerability per se, but could cause denial of service. |

---

### INFO-1: Future `iat` Claims Accepted

Tokens with `iat` (issued-at) set in the future are accepted. This is standard behavior for `jwt-go` and not a security risk, but notable.

### INFO-2: Gin Debug Mode Active

Gin runs in debug mode by default, which prints route tables and other debug info to stdout. Set `gin.SetMode(gin.ReleaseMode)` in production.

---

## Tenant Isolation Analysis

| Vector | Status | Details |
|--------|--------|---------|
| Separate databases | PASS | Each league gets its own `laplogger_league_<uuid>` database |
| Membership gate | PASS | `TenantRequired` middleware checks `league_memberships` before granting access |
| DB name injection | PASS | `validDBName` regex rejects all non-alphanumeric characters except underscore |
| Connection pool isolation | PASS | Each tenant DB gets its own `*sql.DB` with independent pool settings |
| SQL parameterization | PASS | All queries use `$N` placeholders — no string interpolation |
| Cross-tenant data leak | PASS | No shared tables between tenants; control plane only stores metadata |
| `DBName` in API responses | PASS | `json:"-"` tag prevents DB name from appearing in JSON responses |

---

## Security Tests Added

| File | Tests | Coverage Area |
|------|-------|---------------|
| `middleware/auth_security_test.go` | 8 tests (35 sub-cases) | JWT algorithm confusion, tampering, empty secrets, missing exp, SQL injection in claims, error message leakage |
| `handlers/auth_security_test.go` | 4 tests (7 sub-cases) | OAuth state CSRF, callback state validation, token in URL, code injection |
| `handlers/league_security_test.go` | 4 tests (19 sub-cases) | Slug XSS, slug SQL injection, path traversal, DB name generation safety |
| `database/database_security_test.go` | 4 tests (22 sub-cases) | DB name injection patterns, length limits, empty name, tenant isolation |

**Total**: 20 new security tests with 83 sub-cases. All passing.

---

## Recommendations Priority

### Must Fix Before Production
1. **HIGH-1**: Implement proper OAuth state parameter (random + cookie + validate)
2. **HIGH-2**: Move JWT from URL to HttpOnly cookie (Phase 6)
3. **MEDIUM-2**: Add `jwt.WithExpirationRequired()` to JWT parser

### Fix Before Phase 6
4. **MEDIUM-1**: Validate JWT secret strength at startup
5. **MEDIUM-3**: Add length check on DB names (max 63 chars)

### Production Hardening
6. **LOW-1**: Enable `sslmode=require` for PostgreSQL in production
7. **INFO-2**: Set `gin.SetMode(gin.ReleaseMode)` in production
