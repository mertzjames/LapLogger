# QA Gate Report: QA-3 (Post-Phase 3 — League Management)

**Date**: 2026-04-17 (updated after fixes)  
**Phase**: 3 — League Management (create league, provision tenant DB, list leagues)  
**Status**: **PASS**

---

## Automated Checks

| Check          | Result | Details                        |
|----------------|--------|--------------------------------|
| go vet         | PASS   | 0 warnings                    |
| golangci-lint  | PASS   | 0 findings                    |
| go test        | PASS   | 54/54 passed (all packages)   |
| eslint         | N/A    | No frontend code yet          |
| tsc            | N/A    | No frontend code yet          |
| frontend tests | N/A    | No frontend code yet          |

### Coverage

| Package    | Coverage |
|------------|----------|
| config     | 100.0%   |
| database   | 4.6%     |
| handlers   | 22.2%    |
| middleware | 46.0%    |
| models     | no statements |
| **Total**  | **17.6%** |

> Coverage is limited by the unit-test-only approach. Database and handler packages require integration tests with a live PostgreSQL instance. This is expected and acceptable for this stage.

---

## Acceptance Review Findings — Resolved

### ~~CRITICAL-2: Non-Atomic `CreateLeague` Operation~~ → **FIXED**
- **Fix applied**: Replaced 4 sequential operations with:
  1. `CreateLeagueWithMembership()` — wraps league INSERT + membership INSERT in a single SQL transaction with `defer tx.Rollback()` safety
  2. DB name generated up front (UUID created in handler, passed to transaction)
  3. On `ProvisionTenantDB` failure, `DeleteLeague()` is called to remove the orphan record (CASCADE removes memberships)
- **Verification**: Code reviewed — transaction commit only on success, `defer Rollback()` is no-op after commit, `DeleteLeague` rollback logs errors without panicking
- **Status**: ✅ Resolved

### ~~HIGH-1: `AddMembership` Silent Success on Duplicate~~ → **FIXED**
- **Fix applied**: `AddMembership` now returns `(bool, error)` — uses `RowsAffected()` to report whether a new row was inserted vs. duplicated
- **Note**: Primary league creation now uses the transactional `CreateLeagueWithMembership` (no `ON CONFLICT`). The standalone `AddMembership` with conflict handling remains for future use (e.g., inviting users to existing leagues).
- **Status**: ✅ Resolved

### HIGH-2: `UpsertUser` Generates UUID Unconditionally
- **Impact**: Minor — logic works correctly, `RETURNING` always returns the actual ID
- **Status**: ⚠️ Accepted as-is (low priority)

---

## Additional Fixes Applied (from QA-1, SEC-1)

| ID | Issue | Fix | Status |
|---|---|---|---|
| MEDIUM-2 | Missing FK indexes on tenant tables | New migration `000002_add_indexes.up.sql` — indexes on `swimmers.team_id`, `events.meet_id`, `times.event_id`, `times.swimmer_id` | ✅ Fixed |
| LOW-3 | Missing index on `league_memberships` | New migration `000002_add_indexes.up.sql` — index on `league_memberships(user_id)` | ✅ Fixed |
| MEDIUM-1 | Date fields as `string` | `Swimmer.DateOfBirth` and `Meet.MeetDate` changed to `*time.Time` | ✅ Fixed |
| SEC MEDIUM-2 | JWT without `exp` accepted | Added `jwt.WithExpirationRequired()` to auth middleware | ✅ Fixed |
| SEC MEDIUM-3 | No DB name length validation | Added `len(dbName) > 63` guard in `ProvisionTenantDB` | ✅ Fixed |
| — | 3 staticcheck QF1001 findings | Applied De Morgan's law in `league_security_test.go` | ✅ Fixed |

---

## Remaining Known Issues (Deferred)

| ID | Issue | Deferred To | Rationale |
|---|---|---|---|
| HIGH-3 | JWT token in URL query parameter | Phase 6 (Frontend) | Requires cookie-based auth wired end-to-end |
| SEC HIGH-1 | OAuth state parameter hardcoded | Pre-production | Acceptable for local dev; needs crypto-random state + cookie |
| SEC MEDIUM-1 | No JWT secret strength validation | Pre-production | Default secret acceptable for dev |
| LOW-1 | Hardcoded connection pool sizes | Post-MVP | Reasonable defaults for now |
| LOW-2 | League name validation ordering | Post-MVP | Cosmetic |
| HIGH-2 | `UpsertUser` generates UUID unconditionally | Post-MVP | Logic is correct |

---

## Verdict

- **Acceptance: PASS**
- **Overall: PASS — all blockers resolved, ready for Phase 4**
