# EchoState: Add Missing Go Tests

## TL;DR

> **Quick Summary**: Backfill missing unit and integration tests for previously untested packages (`config`, `db`, `models`) and extend existing coverage gaps in `handlers/scan.go` and `scanner/scanner.go`, following the project's existing DB-backed test patterns.
>
> **Deliverables**:
> - `internal/config/config_test.go`
> - `internal/db/db_test.go`
> - `internal/models/models_test.go`
> - Extended `internal/handlers/scan_test.go`
> - `internal/scanner/scanner_test.go`
>
> **Estimated Effort**: Short
> **Parallel Execution**: YES — 2 waves + final verification
> **Critical Path**: 1 → 2 → 3 → 4 → 5 → 6 → F1-F4

---

## Context

### Original Request
Create any missing Go tests.

### Interview Summary
**Key Discussions**:
- Scope includes both files with zero tests and coverage gaps in existing test files.
- Approach: match existing project patterns (DB-backed tests skip if Postgres unavailable, env-var tests use `t.Setenv`, scanner tests avoid live network unless explicitly skippable).

**Research Findings**:
- Untested files: `internal/config/config.go`, `internal/db/db.go`, `internal/models/models.go`.
- Under-tested files: `internal/handlers/scan_test.go` only covers `upsertTarget`; `scanner/scanner.go` only has integration tests.
- Existing test pattern: `setupTestDB(t)` connects to real Postgres, runs migrations, and cleans tables.

### Metis Review
**Identified Gaps** (addressed):
- `main.go` excluded from scope as thin wiring unless explicitly requested.
- Production code changes for testability are out of scope by default.
- DB tests must follow skip-if-unavailable behavior and clean up after themselves.
- Scanner tests should use mock gatherers to avoid live network/browser dependencies.
- Acceptance criteria must be concrete `go test` commands, not vague coverage goals.

---

## Work Objectives

### Core Objective
Add behavior-focused tests for untested packages and under-tested functions so the entire suite can be run with `go test ./...`.

### Concrete Deliverables
- `internal/config/config_test.go`
- `internal/db/db_test.go`
- `internal/models/models_test.go`
- Extended `internal/handlers/scan_test.go`
- `internal/scanner/scanner_test.go`

### Definition of Done
- [ ] `go test ./...` passes.
- [ ] New tests cover the named files/functions.
- [ ] DB-dependent tests skip cleanly when Postgres is unavailable.
- [ ] No production behavior changes unless strictly necessary and documented.

### Must Have
- Tests for `config.Load()` default values, overrides, and `DATABASE_URL` assembly.
- Tests for `db.Connect()` success and failure paths.
- Tests for `db.Migrate()` idempotency and schema creation.
- Tests for `models.ReportStatus` constants.
- Tests for `handlers.health`, `createScan`, `storeSnapshot`, and `computeChanges`.
- Tests for `scanner.Hash`, `mergeMap`, and `Run` with mock gatherers.

### Must NOT Have (Guardrails)
- No production code refactors solely for test ergonomics.
- No live internet or browser-dependent tests unless explicitly marked/skipped.
- No new Docker Compose, CI, testcontainers, or coverage tooling.
- No tests for `main.go` wiring.
- No broad rewrites of existing tests.
- No mocks/interfaces unless required by the test itself.

---

## Verification Strategy (MANDATORY)

> **ZERO HUMAN INTERVENTION** - ALL verification is agent-executed.

### Test Decision
- **Infrastructure exists**: YES (Go `go test`, existing `*_test.go` files).
- **Automated tests**: YES (tests-after, since this is adding tests to existing code).
- **Framework**: `go test`.
- **Agent-Executed QA**: ALWAYS — every task includes concrete `go test` commands as acceptance criteria.

### QA Policy
Every task MUST include agent-executed QA scenarios. Evidence saved to `.omo/evidence/task-{N}-{scenario-slug}.{ext}`.

- **Library/Module**: Use `Bash` (`go test`, `go vet`, `go build`).

---

## Execution Strategy

### Parallel Execution Waves

```
Wave 1 (Pure unit tests - no external services):
├── 1. Add internal/config/config_test.go [quick]
├── 2. Add internal/models/models_test.go [quick]
└── 3. Add internal/scanner/scanner_test.go [quick]

Wave 2 (DB-backed tests - require Postgres, skip if unavailable):
├── 4. Add internal/db/db_test.go [unspecified-high]
└── 5. Extend internal/handlers/scan_test.go [unspecified-high]

Wave 3 (Suite-wide verification):
└── 6. Run full test suite and verify no regressions [quick]

Wave FINAL (After ALL tasks):
├── F1. Plan compliance audit (oracle)
├── F2. Code quality review (unspecified-high)
├── F3. Real manual QA — run targeted test commands (unspecified-high)
└── F4. Scope fidelity check (deep)
-> Present results -> Get explicit user okay
```

### Dependency Matrix

- **1**: - → 6
- **2**: - → 6
- **3**: - → 6
- **4**: - → 6
- **5**: 4 → 6 (can run in parallel with 4 if DB isolation is respected)
- **6**: 1-5 → F1-F4
- **F1-F4**: 1-6 → user okay

### Agent Dispatch Summary

- **Wave 1**: 1 → `quick`, 2 → `quick`, 3 → `quick`
- **Wave 2**: 4 → `unspecified-high`, 5 → `unspecified-high`
- **Wave 3**: 6 → `quick`
- **FINAL**: F1 → `oracle`, F2 → `unspecified-high`, F3 → `unspecified-high`, F4 → `deep`

---

## TODOs

- [x] 1. Add `internal/config/config_test.go`

  **What to do**:
  - Create `internal/config/config_test.go` in package `config`.
  - Test `Load()` defaults:
    - `ECHOSTATE_ENV` defaults to `"development"`.
    - `ECHOSTATE_PORT` defaults to `"8080"`.
    - `BROWSER_WS_URL` defaults to `"ws://localhost:3000/"`.
    - `DATABASE_URL` is assembled from `POSTGRES_*` env vars when absent.
  - Test `Load()` overrides:
    - Setting `ECHOSTATE_ENV`, `ECHOSTATE_PORT`, `BROWSER_WS_URL` returns those values.
    - Setting `DATABASE_URL` uses it directly and ignores `POSTGRES_*` vars.
  - Test `getEnv()` helper:
    - Returns env value when set.
    - Returns fallback when unset or empty.
  - Use `t.Setenv` and `t.Unsetenv` (or `t.Setenv(key, "")`) to avoid leaking env state.

  **Must NOT do**:
  - Do not change `config.go` production code unless a small extraction is unavoidable (document in commit).
  - Do not assert exact internal error strings from `os`/`fmt` unless stable.
  - Do not require a real database.

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: none required

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 2, 3)
  - **Blocks**: 6
  - **Blocked By**: None

  **References**:
  - `internal/config/config.go` — functions under test.
  - `internal/handlers/reports_test.go:30-55` — existing testify-based test style.

  **Acceptance Criteria**:
  - [ ] `internal/config/config_test.go` exists.
  - [ ] `go test ./internal/config/... -count=1` passes.
  - [ ] `go test ./internal/config/... -run TestLoad -count=1` passes.

  **QA Scenarios**:

  ```
  Scenario: Config defaults are applied when env vars are unset
    Tool: Bash
    Preconditions: No ECHOSTATE_* or POSTGRES_* env vars set
    Steps:
      1. Run `go test ./internal/config/... -run TestLoadDefaults -count=1 -v`
    Expected Result: Test passes; Load returns defaults.
    Evidence: .omo/evidence/task-1-config-defaults.txt

  Scenario: Config env overrides take precedence
    Tool: Bash
    Preconditions: Env vars set via test
    Steps:
      1. Run `go test ./internal/config/... -run TestLoadOverrides -count=1 -v`
    Expected Result: Test passes; Load returns overridden values.
    Evidence: .omo/evidence/task-1-config-overrides.txt
  ```

  **Evidence to Capture**:
  - [ ] Test output for default and override scenarios.

  **Commit**: YES
  - Message: `test(config): cover env loading and DATABASE_URL assembly`
  - Files: `internal/config/config_test.go`

- [x] 2. Add `internal/models/models_test.go`

  **What to do**:
  - Create `internal/models/models_test.go` in package `models`.
  - Test `ReportStatus` constants equal the expected string values:
    - `ReportPending == "pending"`
    - `ReportRunning == "running"`
    - `ReportCompleted == "completed"`
    - `ReportFailed == "failed"`
  - Test that `ReportStatus` is a distinct string type and constants are comparable.
  - Optionally verify JSON marshaling/unmarshaling of `ReportResponse` and `ScanRequest` with sample values.

  **Must NOT do**:
  - Do not test plain struct field presence only (no value in that).
  - Do not add JSON tags beyond what already exists.

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: none required

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1, 3)
  - **Blocks**: 6
  - **Blocked By**: None

  **References**:
  - `internal/models/models.go` — types and constants under test.

  **Acceptance Criteria**:
  - [ ] `internal/models/models_test.go` exists.
  - [ ] `go test ./internal/models/... -count=1` passes.

  **QA Scenarios**:

  ```
  Scenario: Report status constants have expected values
    Tool: Bash
    Preconditions: None
    Steps:
      1. Run `go test ./internal/models/... -run TestReportStatus -count=1 -v`
    Expected Result: Test passes; constants match expected strings.
    Evidence: .omo/evidence/task-2-models-status.txt
  ```

  **Evidence to Capture**:
  - [ ] Test output showing constants validation.

  **Commit**: YES
  - Message: `test(models): validate report status constants`
  - Files: `internal/models/models_test.go`

- [x] 3. Add `internal/scanner/scanner_test.go`

  **What to do**:
  - Create `internal/scanner/scanner_test.go` in package `scanner`.
  - Test `Hash`:
    - Returns deterministic SHA256 hex string for same input.
    - Returns different hash for different input.
    - Handles empty `ScanResult`.
  - Test `mergeMap`:
    - Merges source into destination.
    - Overwrites existing keys.
    - Handles nil destination and source.
    - Does not mutate source map.
  - Test `Run` with a custom `Scanner` built with mock gatherers:
    - Successful gatherer populates the correct field (`whois`, `asn`, `web`).
    - Failing gatherer records an error but does not panic.
    - Multiple gatherers aggregate correctly.
    - Overlapping keys merge on error path (see `mergeMap` usage in `Run`).
    - Context cancellation/timeout is respected.
  - Use the existing `Scanner` struct fields directly (same package) to inject mock gatherers without changing production code.

  **Must NOT do**:
  - Do not modify `scanner.go` production code unless unavoidable.
  - Do not use live network/browser gatherers.
  - Do not duplicate the existing integration test in `scanner_integration_test.go`.

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: none required

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1, 2)
  - **Blocks**: 6
  - **Blocked By**: None

  **References**:
  - `internal/scanner/scanner.go` — functions under test.
  - `internal/scanner/scanner_integration_test.go` — existing scanner test style.

  **Acceptance Criteria**:
  - [ ] `internal/scanner/scanner_test.go` exists.
  - [ ] `go test ./internal/scanner/... -count=1` passes.

  **QA Scenarios**:

  ```
  Scenario: Hash is deterministic and sensitive to input
    Tool: Bash
    Preconditions: None
    Steps:
      1. Run `go test ./internal/scanner/... -run TestHash -count=1 -v`
    Expected Result: Test passes; same input -> same hash, different input -> different hash.
    Evidence: .omo/evidence/task-3-scanner-hash.txt

  Scenario: Run aggregates mock gatherers and handles failures
    Tool: Bash
    Preconditions: None
    Steps:
      1. Run `go test ./internal/scanner/... -run TestRun -count=1 -v`
    Expected Result: Test passes; success/failure/mixed gatherer cases covered.
    Evidence: .omo/evidence/task-3-scanner-run.txt
  ```

  **Evidence to Capture**:
  - [ ] Test output for Hash and Run scenarios.

  **Commit**: YES
  - Message: `test(scanner): cover hash, merge, and run with mock gatherers`
  - Files: `internal/scanner/scanner_test.go`

- [x] 4. Add `internal/db/db_test.go`

  **What to do**:
  - Create `internal/db/db_test.go` in package `db`.
  - Test `Connect`:
    - Success path: connect to Postgres using `testDatabaseURL()` (reuse the helper pattern from `internal/handlers/reports_test.go:23-28` or define a local helper).
    - Failure path: invalid URL returns an error.
    - Failure path: unreachable host returns a ping error.
  - Test `Migrate`:
    - Running migrations creates `targets`, `snapshots`, and `reports` tables with expected columns/indexes.
    - Running migrations twice is idempotent (no error).
  - Test `Close`:
    - Closing a connected DB does not panic.
  - Use `t.Skipf` if Postgres is unavailable, matching the existing pattern.
  - Clean up any test-created rows or use unique identifiers where needed.

  **Must NOT do**:
  - Do not require a fresh database instance; tests must be idempotent and clean up.
  - Do not assert exact internal pgx error strings.
  - Do not modify `db.go` production code.

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: none required

  **Parallelization**:
  - **Can Run In Parallel**: YES (with Task 5 if DB isolation is respected)
  - **Parallel Group**: Wave 2 (with Task 5)
  - **Blocks**: 6
  - **Blocked By**: None

  **References**:
  - `internal/db/db.go` — functions under test.
  - `internal/handlers/reports_test.go:23-55` — existing DB test setup and cleanup pattern.

  **Acceptance Criteria**:
  - [ ] `internal/db/db_test.go` exists.
  - [ ] `go test ./internal/db/... -count=1` passes (or skips cleanly if DB unavailable).

  **QA Scenarios**:

  ```
  Scenario: Connect to Postgres and run migrations idempotently
    Tool: Bash
    Preconditions: Postgres running at localhost:5432
    Steps:
      1. Run `go test ./internal/db/... -run TestConnectAndMigrate -count=1 -v`
    Expected Result: Test passes; tables exist; second migration run succeeds.
    Evidence: .omo/evidence/task-4-db-connect.txt

  Scenario: Connect fails for invalid database URL
    Tool: Bash
    Preconditions: None
    Steps:
      1. Run `go test ./internal/db/... -run TestConnectInvalidURL -count=1 -v`
    Expected Result: Test passes; Connect returns an error.
    Evidence: .omo/evidence/task-4-db-invalid.txt
  ```

  **Evidence to Capture**:
  - [ ] Test output for connect and migration scenarios.

  **Commit**: YES
  - Message: `test(db): cover connection and migration idempotency`
  - Files: `internal/db/db_test.go`

- [x] 5. Extend `internal/handlers/scan_test.go`

  **What to do**:
  - Add tests to the existing `internal/handlers/scan_test.go` file.
  - Test `health`:
    - Returns HTTP 200 with `{"status":"ok","env":"..."}`.
  - Test `createScan`:
    - Invalid JSON returns 400.
    - Missing/empty host returns 400.
    - Valid host triggers scan, persists target/snapshot, and returns snapshot JSON.
    - Scanner failure returns 500.
    - Use a test handler with a mockable scanner. If `Handler.scanner` cannot be replaced without production changes, test through `httptest` and either accept real network behavior or mark the test to skip if browser/network is unavailable.
  - Test `storeSnapshot` directly:
    - New target creates a snapshot.
    - Same hash updates `last_seen` only.
    - Different hash inserts a new snapshot with computed changes.
  - Test `computeChanges` directly:
    - No previous snapshot returns empty changes.
    - Identical current/previous returns empty changes.
    - Added/removed/changed keys are detected.
    - JSON marshal/unmarshal errors produce graceful error strings in changes.

  **Must NOT do**:
  - Do not refactor `scan.go` unless a tiny change is unavoidable (document it).
  - Do not require live browser/network for tests that can be unit-tested directly (`storeSnapshot`, `computeChanges`, `health`).
  - Do not duplicate existing `TestUpsertTargetStoresNormalizedHost`.

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: none required

  **Parallelization**:
  - **Can Run In Parallel**: YES (with Task 4 if DB isolation is respected)
  - **Parallel Group**: Wave 2 (with Task 4)
  - **Blocks**: 6
  - **Blocked By**: None

  **References**:
  - `internal/handlers/scan.go` — functions under test.
  - `internal/handlers/reports_test.go:30-90` — existing DB setup and httptest patterns.
  - `internal/handlers/reports_test.go:90-130` — `newTestHandler` helper.

  **Acceptance Criteria**:
  - [ ] New tests added to `internal/handlers/scan_test.go`.
  - [ ] `go test ./internal/handlers/... -count=1` passes.

  **QA Scenarios**:

  ```
  Scenario: Health endpoint returns ok
    Tool: Bash
    Preconditions: None
    Steps:
      1. Run `go test ./internal/handlers/... -run TestHealth -count=1 -v`
    Expected Result: Test passes; returns 200 and expected JSON.
    Evidence: .omo/evidence/task-5-scan-health.txt

  Scenario: ComputeChanges detects added, removed, and changed keys
    Tool: Bash
    Preconditions: None
    Steps:
      1. Run `go test ./internal/handlers/... -run TestComputeChanges -count=1 -v`
    Expected Result: Test passes; changes list matches expectations.
    Evidence: .omo/evidence/task-5-scan-changes.txt

  Scenario: StoreSnapshot reuses existing snapshot for identical hash
    Tool: Bash
    Preconditions: Postgres running
    Steps:
      1. Run `go test ./internal/handlers/... -run TestStoreSnapshot -count=1 -v`
    Expected Result: Test passes; duplicate hash updates last_seen, different hash inserts new row.
    Evidence: .omo/evidence/task-5-scan-snapshot.txt
  ```

  **Evidence to Capture**:
  - [ ] Test output for health, computeChanges, and storeSnapshot scenarios.

  **Commit**: YES
  - Message: `test(handlers): cover scan handler, snapshot storage, and change detection`
  - Files: `internal/handlers/scan_test.go`

- [x] 6. Run full test suite and verify no regressions

  **What to do**:
  - Run `go test ./... -count=1`.
  - Run `go vet ./...`.
  - Run `go build ./...`.
  - Confirm all new test files are included and no existing tests fail.
  - If any test is flaky or fails due to environment, document it in `.omo/notepads/missing-go-tests/issues.md`.

  **Must NOT do**:
  - Do not ignore failing tests.
  - Do not skip verification if only one environment is unavailable.

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: none required

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 3
  - **Blocks**: F1-F4
  - **Blocked By**: 1-5

  **References**:
  - `.omo/plans/pdf-report-endpoint.md` — previous plan's verification strategy.

  **Acceptance Criteria**:
  - [ ] `go test ./... -count=1` passes.
  - [ ] `go vet ./...` passes.
  - [ ] `go build ./...` passes.

  **QA Scenarios**:

  ```
  Scenario: Full suite passes after adding tests
    Tool: Bash
    Preconditions: Tasks 1-5 complete
    Steps:
      1. Run `go test ./... -count=1`
      2. Run `go vet ./...`
      3. Run `go build ./...`
    Expected Result: All commands exit 0.
    Evidence: .omo/evidence/task-6-full-suite.txt
  ```

  **Evidence to Capture**:
  - [ ] Terminal output of all three commands.

  **Commit**: YES (if any fixes are needed) or NO
  - Message: `test: verify full suite passes with new tests`
  - Files: any fixes needed

---

## Final Verification Wave (MANDATORY — after ALL implementation tasks)

> 4 review agents run in PARALLEL. ALL must APPROVE. Present consolidated results to user and get explicit "okay" before completing.

- [x] F1. **Plan Compliance Audit** — `oracle`
  Read the plan end-to-end. For each "Must Have": verify tests exist and pass. For each "Must NOT Have": search codebase for forbidden patterns — reject with file:line if found. Check evidence files exist in `.omo/evidence/`.
  Output: `Must Have [N/N] | Must NOT Have [N/N] | Tasks [N/N] | VERDICT: APPROVE/REJECT`

- [x] F2. **Code Quality Review** — `unspecified-high`
  Run `go vet ./...` + `go test ./...` + `go build ./...`. Review new test files for: empty catches, `fmt.Println`, dead code, unused imports, AI slop, brittle assertions, flaky tests.
  Output: `Build [PASS/FAIL] | Lint [PASS/FAIL] | Tests [N pass/N fail] | Files [N clean/N issues] | VERDICT`

- [x] F3. **Real Manual QA** — `unspecified-high`
  Execute EVERY QA scenario from EVERY task. Verify targeted test commands pass and `go test ./...` passes. Run with Postgres available and verify DB tests do not skip. Run with Postgres unavailable and verify DB tests skip cleanly. Save evidence to `.omo/evidence/final-qa/`.
  Output: `Scenarios [N/N pass] | Integration [N/N] | Edge Cases [N tested] | VERDICT`

- [x] F4. **Scope Fidelity Check** — `deep`
  For each task: read "What to do", read actual diff. Verify 1:1 — everything in spec was built, nothing beyond spec was built. Check "Must NOT do" compliance. Detect cross-task contamination and unaccounted production code changes.
  Output: `Tasks [N/N compliant] | Contamination [CLEAN/N issues] | Unaccounted [CLEAN/N files] | VERDICT`

---

## Commit Strategy

- Use conventional commit prefixes (`test:`, `refactor:` if minor testability tweak is unavoidable).
- Keep messages natural and concise.
- Example commits:
  - `test(config): cover env loading and DATABASE_URL assembly`
  - `test(db): cover connection and migration idempotency`
  - `test(models): validate report status constants`
  - `test(handlers): cover scan handler, snapshot storage, and change detection`
  - `test(scanner): cover hash, merge, and run with mock gatherers`

---

## Success Criteria

### Verification Commands
```bash
go test ./...           # Expected: PASS
go vet ./...            # Expected: PASS
go build ./...          # Expected: PASS
```

### Final Checklist
- [ ] All new test files listed under Deliverables exist.
- [ ] `go test ./...` passes.
- [ ] No production code changes beyond what is strictly necessary.
- [ ] DB-dependent tests skip cleanly when Postgres is unavailable.
- [ ] No live network/browser tests unless explicitly marked/skipped.
- [ ] Evidence captured for every QA scenario.
