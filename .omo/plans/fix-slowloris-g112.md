# EchoState: Fix Slowloris Warning (G112)

## TL;DR

> **Quick Summary**: Add `ReadHeaderTimeout` to the `http.Server` in `main.go` to resolve gosec G112 (CWE-400) Slowloris warning.
>
> **Deliverables**:
> - `main.go` updated with `ReadHeaderTimeout`
> - `main_test.go` verifying the server config
> - gosec scan passes with zero issues
>
> **Estimated Effort**: Quick
> **Parallel Execution**: NO — single task + verification
> **Critical Path**: 1 → F1-F4

---

## Context

### Original Request
Fix the Slowloris warning reported by `gosec` on `main.go:50-53`.

### Research Findings
- `http.Server` in `main.go` currently has no `ReadHeaderTimeout` configured.
- Existing timeouts in the codebase: gatherers use 5-10s, scan handler uses 60s, worker render timeout is 30s.
- Go's `http.Server.ReadHeaderTimeout` guards against Slowloris-style attacks by limiting the time allowed to read request headers.

---

## Work Objectives

### Core Objective
Eliminate gosec G112 by configuring a reasonable `ReadHeaderTimeout` on the HTTP server.

### Concrete Deliverables
- Updated `main.go` with `ReadHeaderTimeout`.
- New `main_test.go` verifying the server is configured with the timeout.

### Definition of Done
- [ ] `gosec ./...` reports zero issues.
- [ ] `go test ./...` passes.
- [ ] `go vet ./...` passes.
- [ ] `go build ./...` passes.

### Must Have
- `ReadHeaderTimeout` set on `http.Server` in `main.go`.
- A test that confirms the timeout is non-zero.

### Must NOT Have (Guardrails)
- No changes to unrelated production code.
- No new dependencies.
- No overly aggressive timeout that breaks legitimate long requests (e.g., scan endpoint has 60s handler timeout, but header reads should be fast).

---

## Verification Strategy (MANDATORY)

> **ZERO HUMAN INTERVENTION** - ALL verification is agent-executed.

### Test Decision
- **Infrastructure exists**: YES.
- **Automated tests**: YES (tests-after).
- **Framework**: `go test`.
- **Agent-Executed QA**: ALWAYS.

---

## Execution Strategy

### Parallel Execution Waves

```
Wave 1:
└── 1. Add ReadHeaderTimeout to http.Server and add test [quick]

Wave FINAL:
├── F1. Plan compliance audit (oracle)
├── F2. Code quality review (unspecified-high)
├── F3. Real manual QA — run vet/test/gosec (unspecified-high)
└── F4. Scope fidelity check (deep)
-> Present results -> Get explicit user okay
```

---

## TODOs

- [x] 1. Add `ReadHeaderTimeout` to `http.Server` in `main.go` and add `main_test.go`

  **What to do**:
  - Edit `main.go`:
    - Add `ReadHeaderTimeout: 10 * time.Second` to the `http.Server` config.
    - Ensure `time` package is already imported (it is).
  - Create `main_test.go`:
    - Test that the configured server has a non-zero `ReadHeaderTimeout`.
    - Use `httptest` or inspect the server config via a small helper if needed.
    - Alternatively, refactor server creation into a small exported/internal helper `newServer(cfg *config.Config, handler http.Handler) *http.Server` so it can be tested without running `main()`.
    - If a helper is extracted, keep it minimal and in `main.go`.

  **Must NOT do**:
  - Do not change handler behavior.
  - Do not add external dependencies.
  - Do not set the timeout so low that normal requests fail.

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: none required

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 1
  - **Blocks**: F1-F4
  - **Blocked By**: None

  **References**:
  - `main.go:50-53` — `http.Server` configuration.
  - `https://pkg.go.dev/net/http#Server` — `ReadHeaderTimeout` documentation.

  **Acceptance Criteria**:
  - [ ] `gosec ./...` reports zero issues.
  - [ ] `go test ./... -count=1` passes.
  - [ ] `go vet ./...` passes.
  - [ ] `go build ./...` passes.

  **QA Scenarios**:

  ```
  Scenario: gosec reports no issues
    Tool: Bash
    Preconditions: gosec installed
    Steps:
      1. Run `gosec ./...`
    Expected Result: "Issues : 0"
    Failure Indicators: Any non-zero issue count.
    Evidence: .omo/evidence/task-1-gosec.txt

  Scenario: Full test suite still passes
    Tool: Bash
    Preconditions: Fix applied
    Steps:
      1. Run `go test ./... -count=1`
      2. Run `go vet ./...`
      3. Run `go build ./...`
    Expected Result: All commands exit 0.
    Failure Indicators: Any FAIL or non-zero exit.
    Evidence: .omo/evidence/task-1-tests.txt
  ```

  **Evidence to Capture**:
  - [ ] gosec output showing zero issues.
  - [ ] Full test/vet/build output.

  **Commit**: YES
  - Message: `fix(main): add ReadHeaderTimeout to mitigate Slowloris (G112)`
  - Files: `main.go`, `main_test.go`

---

## Final Verification Wave (MANDATORY — after ALL implementation tasks)

> 4 review agents run in PARALLEL. ALL must APPROVE.

- [x] F1. **Plan Compliance Audit** — `oracle`
  Verify `ReadHeaderTimeout` is set, test exists, and gosec passes.
  Output: `Must Have [N/N] | Must NOT Have [N/N] | Tasks [N/N] | VERDICT: APPROVE/REJECT`

- [x] F2. **Code Quality Review** — `unspecified-high`
  Run `go vet ./...` + `go test ./...` + `go build ./...` + `gofmt -l`. Check for unnecessary complexity.
  Output: `Build [PASS] | Lint [PASS] | Tests [8/8 pass] | Files [clean] | VERDICT: APPROVE`

- [x] F3. **Real Manual QA** — `unspecified-high`
  Run gosec and confirm zero issues. Run full test suite. Save evidence to `.omo/evidence/final-qa/`.
  Output: `Scenarios [pass] | Integration [pass] | Edge Cases [covered] | VERDICT: APPROVE`

- [x] F4. **Scope Fidelity Check** — `deep`
  Verify only `main.go` and `main_test.go` changed. Confirm timeout value is reasonable.
  Output: `Tasks [1/1 compliant] | Contamination [CLEAN] | Unaccounted [CLEAN] | VERDICT: APPROVE`

---

## Commit Strategy

- Single commit: `fix(main): add ReadHeaderTimeout to mitigate Slowloris (G112)`

---

## Success Criteria

### Verification Commands
```bash
gosec ./...              # Expected: Issues 0
go test ./... -count=1   # Expected: PASS
go vet ./...             # Expected: PASS
go build ./...           # Expected: PASS
```

### Final Checklist
- [x] `ReadHeaderTimeout` configured on `http.Server`.
- [x] Test verifies the timeout is set.
- [x] gosec reports zero issues.
- [x] Full suite passes.
