# QA Gate Report: QA-3 (Post-Phase 3 — League Management)

**Date**: 2026-04-17  
**Phase**: 3 — League Management (create league, provision tenant DB, list leagues)  
**Status**: **BLOCKED**

---

## Automated Checks

| Check          | Result | Details                        |
|----------------|--------|--------------------------------|
| go vet         | PASS   | 0 warnings                    |
| golangci-lint  | PASS   | 0 findings                    |
| go test        | PASS   | 34/34 passed                  |
| eslint         | N/A    | No frontend code yet          |
| tsc            | N/A    | No frontend code yet          |
| frontend tests | N/A    | No frontend code yet          |

### Coverage

| Package  | Coverage |
|----------|----------|
| handlers | 15.6%    |

---

## Acceptance Review Findings

### CRITICAL-2: Non-Atomic `CreateLeague` Operation (BLOCKER)
- **Location**: `handlers/league.go` — `CreateLeague`
- **Issue**: The handler performs 4 sequential operations with no transaction or rollback:
  1. Insert league record into control DB
  2. Update league record with `db_name`
  3. Provision tenant DB (`CREATE DATABASE` + run migrations)
  4. Insert membership record
- **Failure scenario**: If step 3 fails (e.g. DNS timeout, disk full), an orphan league record exists with no backing database. Retry creates a duplicate league.
- **Impact**: Unrecoverable data inconsistency.
- **Fix options**:
  1. **Reorder**: Provision DB first, then insert league record (if DB creation fails, nothing to clean up).
  2. **Add rollback**: On `ProvisionTenantDB` failure, call a `DeleteLeague()` function to remove the orphan.
  3. **Wrap steps 1+2+4 in a SQL transaction**: At minimum, the control DB operations should be atomic.
- **Action**: **Must fix before Phase 4.**

### HIGH-1: `AddMembership` Silent Success on Duplicate
- **Location**: `database/queries.go` — `AddMembership`
- **Issue**: Uses `ON CONFLICT DO NOTHING`, so the caller cannot distinguish between a successful insert and a silent no-op on duplicate.
- **Impact**: Retry logic or idempotency checks are unreliable.
- **Fix**: Use `RETURNING user_id` and check for `sql.ErrNoRows` to detect duplicates.
- **Action**: Should fix before Phase 4.

### HIGH-2: `UpsertUser` Generates UUID Unconditionally
- **Location**: `database/queries.go` — `UpsertUser`
- **Issue**: A new UUID is generated on every call even when the user already exists (conflict path). The returned ID is correct (from `RETURNING`), but the wasted UUID generation is fragile.
- **Impact**: Minor — logic works correctly but is misleading to read.
- **Action**: Low priority; acceptable as-is.

---

## Verdict

- **Acceptance: FAIL**
- **Overall: BLOCKED**
  - **Blocker**: CRITICAL-2 — `CreateLeague` must include rollback or reordering to prevent orphan records

---

## Required Actions Before Phase 4

| Priority | ID         | Action                                                    |
|----------|------------|-----------------------------------------------------------|
| MUST     | CRITICAL-2 | Make `CreateLeague` atomic or add rollback on failure     |
| MUST     | MEDIUM-2   | Add FK indexes on tenant tables (from QA-1)               |
| MUST     | LOW-3      | Add composite index on `league_memberships` (from QA-1)   |
| SHOULD   | MEDIUM-1   | Change `DateOfBirth`/`MeetDate` to `*time.Time` (from QA-1) |
| SHOULD   | HIGH-1     | Return insertion status from `AddMembership`              |
| DEFER    | HIGH-3     | Move JWT from URL param to HttpOnly cookie (Phase 6)      |
