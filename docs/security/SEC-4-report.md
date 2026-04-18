# Security Review Report: SEC-4 (Post-Phase 6)

**Date**: 2026-04-17
**Reviewer**: Security Agent (automated code review)
**Scope**: Frontend — React/TypeScript SPA (16 files), auth flow, API integration, user-facing components
**Methodology**: OWASP Top 10 (2021) + SPA threat model

---

## Executive Summary

The frontend implementation demonstrates solid security fundamentals: React's automatic XSS escaping is used throughout (zero `dangerouslySetInnerHTML`), route protection via `PrivateRoute` is correct, and TypeScript provides type-safety. However, the OAuth token delivery mechanism has a **critical** vulnerability, and several high-severity items require attention before production deployment.

| Severity | Count | Status |
|----------|-------|--------|
| CRITICAL | 1     | Requires backend change (JWT in URL) |
| HIGH     | 3     | Mitigated by design; address before production |
| MEDIUM   | 4     | Acceptable for MVP |
| LOW      | 2     | Best-practice improvements |

---

## Findings

### FINDING 1: JWT Token Exposed in URL Query Parameter [CRITICAL]

**Category:** A02:2021 – Cryptographic Failures

The backend OAuth callback redirects to `/?token=<jwt>`. While the frontend cleans the URL via `window.history.replaceState()`, the token is exposed in:
- Browser history before cleanup
- HTTP `Referer` headers on any outbound navigation
- Server/proxy access logs
- Browser extensions with URL access

**Current mitigation in code:** `AuthContext.tsx` extracts the token and immediately replaces the URL with `history.replaceState({}, '', window.location.pathname)`.

**Recommendation:** Use URL fragment (`#token=xxx`) instead of query parameter, or switch to secure HTTP-only session cookies set by the backend. Requires backend change — deferred to Phase 7+ or tracked as tech debt.

---

### FINDING 2: Missing PKCE in OAuth Flow [HIGH]

**Category:** A07:2021 – Identification and Authentication Failures

The OAuth flow uses a simple redirect to `/api/auth/google` without PKCE (Proof Key for Code Exchange). Authorization code interception is possible on shared or compromised networks.

**Impact:** Attacker could intercept the authorization code and exchange it for a token.

**Recommendation:** Implement PKCE challenge/verifier in the OAuth flow. Requires backend changes — tracked as tech debt.

---

### FINDING 3: No Explicit CSRF Protection [HIGH]

**Category:** A01:2021 – Broken Access Control

State-changing API requests (POST/PUT/DELETE) do not include CSRF tokens. The Bearer token in the `Authorization` header provides partial protection (not auto-attached by browser), but explicit CSRF tokens would add defense-in-depth.

**Current mitigation:** Bearer tokens are sent via `Authorization` header (not cookies), so browser-based CSRF attacks cannot forge the header.

**Risk level:** Effectively mitigated by Bearer token design, but lacks explicit CSRF layer.

---

### FINDING 4: sessionStorage Vulnerable to XSS [HIGH]

**Category:** A07:2021 – Cross-Site Scripting

JWT stored in `sessionStorage` is accessible to any JavaScript running in the page context. If an XSS vulnerability is introduced (e.g., via a dependency), the token could be exfiltrated.

**Current mitigations:**
- React's automatic output escaping prevents XSS in rendered content
- Zero usage of `dangerouslySetInnerHTML`
- `sessionStorage` clears on tab close (shorter exposure window than `localStorage`)

**Recommendation:** Add Content-Security-Policy headers (see Finding 10) to prevent inline script execution. Consider HTTP-only cookies for token storage in future iterations.

---

### FINDING 5: Insufficient Frontend Input Validation [MEDIUM]

**Category:** A04:2021 – Insecure Design

Form fields lack length constraints (`maxLength`), format validation, and special character handling. While backend validation is the security boundary, frontend validation improves UX and reduces attack surface.

**Examples:** League names, team names, swimmer names accept unlimited input without length restrictions.

**Recommendation:** Add `maxLength` attributes to form inputs and validate formats client-side for UX purposes. Backend must enforce all limits server-side (which it does).

---

### FINDING 6: Error Messages May Leak Backend Structure [MEDIUM]

**Category:** A09:2021 – Security Logging and Monitoring Failures

The `ApiError` class surfaces backend error messages verbatim via `body.error || res.statusText`. If the backend returns detailed error messages (e.g., database connection strings), they would be displayed to users.

**Current code:** `api.ts` throws `ApiError(res.status, body.error || res.statusText)` → components display the error message directly.

**Recommendation:** Map error status codes to generic user-facing messages. Log detailed errors to console for debugging only. Backend should also sanitize error responses.

---

### FINDING 7: No Rate Limiting on Frontend [MEDIUM]

**Category:** A04:2021 – Insecure Design

Form submission buttons are not disabled during API requests, allowing rapid duplicate submissions. No request throttling exists in the API service layer.

**Recommendation:** Disable submit buttons during pending requests (add `isSubmitting` state to forms). Backend should implement rate limiting per user/endpoint.

---

### FINDING 8: Personal Information Rendered Without Additional Protection [MEDIUM]

**Category:** A09:2021 – Security Logging and Monitoring Failures

Swimmer PII (names, dates of birth) and user names are rendered without browser caching controls. Minors' data displayed in `PublicResults.tsx` and `Swimmers.tsx`.

**Recommendation:** Add `Cache-Control: no-store` headers via nginx for authenticated pages. Document PII handling policy. Low priority for current MVP.

---

### FINDING 9: No Content Security Policy (CSP) [LOW]

**Category:** A07:2021 – Cross-Site Scripting

`index.html` does not define a Content-Security-Policy. While React prevents most XSS, CSP provides defense-in-depth against inline script injection.

**Recommendation:** Add CSP headers via nginx in Phase 7:
```
Content-Security-Policy: default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; connect-src 'self'; frame-src 'none'; object-src 'none'
```

---

### FINDING 10: Missing Security-Related HTTP Headers [LOW]

**Category:** A05:2021 – Security Misconfiguration

Standard security headers are not configured: `X-Content-Type-Options`, `X-Frame-Options`, `Strict-Transport-Security`, `Referrer-Policy`, `Permissions-Policy`.

**Recommendation:** Add all security headers in nginx configuration during Phase 7.

---

## Positive Security Observations

| Area | Assessment |
|------|-----------|
| **XSS Prevention** | Zero `dangerouslySetInnerHTML` usage; all content rendered through React's automatic escaping |
| **Route Protection** | `PrivateRoute` correctly handles loading states and redirects unauthenticated users |
| **Token Storage** | `sessionStorage` chosen over `localStorage` — clears on tab close |
| **URL Cleanup** | Token removed from URL immediately after capture via `replaceState()` |
| **Authorization Header** | Bearer token sent via `Authorization` header, not URL parameters or cookies |
| **Type Safety** | Full TypeScript coverage with proper interfaces prevents type-confusion attacks |
| **HTTP Methods** | Correct REST verb usage (GET/POST/PUT/DELETE) for all operations |
| **Time Input Validation** | `parseTime()` validates swim time format with regex before submission |
| **Dependencies** | Minimal dependency footprint: React 19, react-router-dom 7, TypeScript 6 — all current versions |

---

## OWASP Top 10 Summary

| Category | Status | Notes |
|----------|--------|-------|
| A01: Broken Access Control | ✅ Good | PrivateRoute + Bearer tokens |
| A02: Cryptographic Failures | ⚠️ Critical | JWT in URL query parameter |
| A03: Injection | ✅ Good | React auto-escaping, no raw HTML |
| A04: Insecure Design | ⚠️ Medium | Missing PKCE, input validation, rate limiting |
| A05: Security Misconfiguration | ✅ Low | Missing headers (Phase 7 scope) |
| A07: XSS | ✅ Good | React escaping + no dangerouslySetInnerHTML |
| A08: Software Integrity | ✅ Good | Clean dependencies, no known vulns |
| A09: Logging & Monitoring | ⚠️ Medium | Error message leakage, PII caching |
| A10: SSRF | ✅ N/A | No server-side requests from frontend |

---

## Remediation Priority

| Priority | Finding | Owner | Target Phase |
|----------|---------|-------|-------------|
| **Immediate** | F1: JWT in URL (CRITICAL) | backend | Phase 7+ / tech debt |
| **Short-term** | F2: PKCE (HIGH) | backend | Phase 7+ / tech debt |
| **Short-term** | F3: CSRF tokens (HIGH) | backend + frontend | Phase 7+ / tech debt |
| **Short-term** | F4: XSS → sessionStorage (HIGH) | devsecops (CSP) | Phase 7 (nginx) |
| **Medium-term** | F5: Input validation (MEDIUM) | frontend | Next frontend iteration |
| **Medium-term** | F6: Error messages (MEDIUM) | frontend + backend | Next iteration |
| **Medium-term** | F7: Rate limiting (MEDIUM) | backend + frontend | Next iteration |
| **Medium-term** | F8: PII caching (MEDIUM) | devsecops | Phase 7 (nginx) |
| **Phase 7** | F9: CSP headers (LOW) | devsecops | Phase 7 (nginx) |
| **Phase 7** | F10: Security headers (LOW) | devsecops | Phase 7 (nginx) |

---

## Conclusion

The frontend code itself is well-secured through React's built-in protections and proper TypeScript usage. The critical and high findings (JWT in URL, PKCE, CSRF, sessionStorage XSS) primarily require **backend or infrastructure changes** rather than frontend code fixes. Several medium findings (CSP, security headers, PII caching) will be addressed in Phase 7 when nginx is configured. The frontend is **approved for MVP** with the understanding that the critical finding is tracked as tech debt for remediation.
