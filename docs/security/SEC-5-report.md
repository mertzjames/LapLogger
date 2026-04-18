# Security Review Report: SEC-5 (Post-Phase 7)

**Date**: 2026-04-17
**Reviewer**: Security Agent (automated review with remediation verification)
**Scope**: Dockerfiles, Docker Compose topology, nginx reverse proxy, `.env.example`, and deployment bootstrap script
**Methodology**: OWASP Top 10 (2021) + container hardening and reverse-proxy review

---

## Executive Summary

The Phase 7 deployment stack is in solid MVP shape after follow-up hardening. The final stack uses non-root runtime containers, internal-only service networking for backend and database traffic, health checks on all services, and security headers at the nginx edge. No CRITICAL findings remain in the current implementation.

| Severity | Count | Status |
|----------|-------|--------|
| CRITICAL | 0 | — |
| HIGH | 1 | Acceptable for MVP with production follow-up |
| MEDIUM | 2 | Acceptable for MVP |
| LOW | 2 | Best-practice follow-up |

---

## Findings

### FINDING 1: HTTPS/HSTS Not Configured [HIGH]

**Category:** A05:2021 – Security Misconfiguration

The reverse proxy serves HTTP on port 80 only and does not set `Strict-Transport-Security`. This is acceptable for local development and Docker Compose evaluation, but production deployment requires TLS termination and HSTS.

**Current state:**
- nginx listens on `80`
- No `Strict-Transport-Security` header is set
- Other security headers are present

**Recommendation:** For production, terminate TLS at nginx or an external load balancer and add:

```nginx
add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
```

---

### FINDING 2: Resource Limits Not Defined in Compose [MEDIUM]

**Category:** A04:2021 – Insecure Design

Services do not define CPU or memory limits. On a shared host, a runaway container could consume disproportionate resources.

**Recommendation:** Add deployment-specific memory and CPU constraints in the production Compose or orchestration manifests.

---

### FINDING 3: Container Runtime Hardening Is Partial [MEDIUM]

**Category:** A05:2021 – Security Misconfiguration

The backend and frontend images run as non-root, which is correct. However, the Compose services do not yet add defense-in-depth runtime controls such as `read_only`, `cap_drop`, or `no-new-privileges`.

**Recommendation:** In production, add:
- `read_only: true` where feasible
- `security_opt: ["no-new-privileges:true"]`
- `cap_drop: ["ALL"]` for nginx/frontend if validated against runtime needs

---

### FINDING 4: OAuth/JWT Session Flow Still Uses URL Query Token [LOW]

**Category:** A02:2021 – Cryptographic Failures

The deployment stack correctly routes OAuth through nginx, but the application-level auth flow still redirects with `/?token=<jwt>`. This issue originates in the backend auth design, not the Docker setup, but it remains relevant to final deployment posture.

**Recommendation:** Replace query-token delivery with secure cookies or fragment-based transfer in a future auth hardening phase.

---

### FINDING 5: `.env.example` Uses Development Secrets and Localhost Defaults [LOW]

**Category:** A05:2021 – Security Misconfiguration

The example environment file correctly avoids real secrets, but it still includes a weak development JWT secret placeholder and localhost-oriented URLs. This is acceptable for development, but operators must replace them for production.

**Recommendation:** Keep `.env.example` as a dev template, and document required production overrides clearly.

---

## Positive Security Observations

| Area | Assessment |
|------|-----------|
| **Backend image** | Multi-stage build, minimal Alpine runtime, non-root `app` user |
| **Frontend image** | Multi-stage build, unprivileged nginx runtime image |
| **No baked-in secrets** | Dockerfiles do not embed credentials or secrets |
| **Service exposure** | Only nginx is published to the host; backend and postgres are internal-only |
| **Network isolation** | Separate `app-tier` and `database-tier` networks reduce lateral exposure |
| **Health checks** | Postgres, backend, frontend, and nginx all have active health probes |
| **Reverse-proxy headers** | `X-Frame-Options`, `X-Content-Type-Options`, `Referrer-Policy`, `Permissions-Policy`, and CSP are active |
| **Bootstrap safety** | `start.sh` creates `.env` from `.env.example` only when absent |
| **Secret template hygiene** | `.env.example` uses placeholders instead of real Google OAuth credentials |

---

## Security Criteria Assessment

| Criterion | Status | Notes |
|----------|--------|-------|
| Non-root runtime containers | ✅ Pass | Backend and frontend runtimes are non-root |
| Minimal base images | ✅ Pass | Alpine-based runtime images |
| No secrets baked into images | ✅ Pass | Credentials injected only through env vars |
| nginx security headers | ✅ Mostly pass | Strong header set, HSTS still pending for TLS deployments |
| Host port exposure minimized | ✅ Pass | Only nginx exposes a host port |
| Internal service networking | ✅ Pass | Explicit app/database networks |
| `.env.example` placeholder-only | ✅ Pass | No real credentials in template |
| Health and readiness checks | ✅ Pass | All services have health checks |

---

## Conclusion

The Phase 7 Docker and deployment setup is **approved for MVP**. The most important deployment risks discovered during review were addressed in code: backend and Postgres are no longer exposed to the host, and explicit internal networks are in place. Remaining items are primarily production-hardening concerns rather than MVP blockers.

### Required Before Production

1. Add TLS termination and HSTS
2. Define resource limits appropriate to the deployment host
3. Add stronger runtime restrictions (`read_only`, `cap_drop`, `no-new-privileges`) where feasible
4. Replace JWT query-parameter delivery in the auth flow
