# EchoState: Tasks 4 + 1 — Funding Cleanup + Passive Reconnaissance Gatherers

## TL;DR

> **Quick Summary**: Clean up `.github/FUNDING.yml` with the correct GitHub Sponsors config for `notfixingit3`, then implement the three passive reconnaissance gatherers (WHOIS, ASN/BGP via Team Cymru, and chromedp web copyright scraping) in the existing scanner, with tests and docker-compose verification.
>
> **Deliverables**:
> - Updated `.github/FUNDING.yml`
> - `internal/scanner/normalize.go` + tests
> - `internal/scanner/whois.go` + tests
> - `internal/scanner/asn.go` + tests
> - `internal/scanner/web.go` + tests
> - Updated `internal/scanner/scanner.go` wiring + timeouts
> - Integration test and docker-compose end-to-end verification
>
> **Estimated Effort**: Medium
> **Parallel Execution**: YES — 3 waves + final verification
> **Critical Path**: 1 → 2 → 3 → 4-6 → 7 → 8 → 9 → F1-F4

---

## Context

### Original Request
Update EchoState's funding metadata and implement the passive reconnaissance gatherers: WHOIS, ASN/BGP, and chromedp DOM copyright scraping.

### Task ID Mapping
The original request numbered these as Task 4 (FUNDING.yml) and Task 1 (gatherers). This plan renumbers them sequentially for execution:

| Original Task | Plan Task(s) | Description |
|---------------|--------------|-------------|
| Task 4        | 1            | Clean up `.github/FUNDING.yml` |
| Task 1        | 2–9          | Implement gatherers, tests, and verification |

### Interview Summary
**Key Discussions**:
- Task 4 (FUNDING.yml) must be completed before Task 1 (gatherers).
- User's GitHub username is `notfixingit3`.
- Existing scanner skeleton already supports concurrent gatherers and per-gatherer graceful failure.

**Research Findings**:
- `internal/scanner/scanner.go` contains stub gatherers returning `{"status": "pending"}`.
- `docker-compose.yml` provides `browserless/chrome` at `ws://browser:3000/`.
- `config.go` exposes `BrowserWSURL` and the database URL.
- `handlers/scan.go` calls `scanner.Run` with a 60-second context timeout.

### Metis Review
**Identified Gaps** (addressed):
- WHOIS parsing may use `github.com/likexian/whois-parser` alongside the raw library.
- Gatherers return structured maps; failures are appended to `result.Errors` without failing the scan.
- Unit tests mock external dependencies where practical; agent QA uses live `example.com`.
- Per-gatherer timeout set to 10 seconds.
- Input normalization strips scheme, `www.`, path, and port; supports domains and IPs.
- `BROWSER_WS_URL=ws://browser:3000/` connects via `chromedp.NewRemoteAllocator`.
- Copyright scraping extracts visible DOM text containing `©`, `Copyright`, or 4-digit years.

---

## Work Objectives

### Core Objective
Clean up the GitHub funding file and implement production-ready WHOIS, ASN/BGP, and web copyright gatherers that integrate with the existing concurrent scanner.

### Concrete Deliverables
- `.github/FUNDING.yml` with only active funding lines.
- `internal/scanner/normalize.go` with `NormalizeHost` and tests.
- `internal/scanner/whois.go` using `likexian/whois` + `whois-parser`.
- `internal/scanner/asn.go` using Team Cymru DNS TXT lookups.
- `internal/scanner/web.go` using `chromedp` via browserless Chrome.
- Updated `internal/scanner/scanner.go` with real gatherers and 10s timeouts.
- Integration test verifying `scanner.Run` against `example.com`.

### Definition of Done
- [ ] `go test ./...` passes.
- [ ] `go vet ./...` passes.
- [ ] Docker Compose stack builds and runs.
- [ ] A scan of `example.com` returns non-pending WHOIS, ASN, and web data.
- [ ] FUNDING.yml validates against GitHub's schema.

### Must Have
- FUNDING.yml cleaned up with `github: [notfixingit3]` and the custom sponsor URL.
- WHOIS gatherer returns registrar, expiration, name servers, and raw text.
- ASN gatherer returns ASN and description via Team Cymru DNS.
- Web gatherer returns page title and copyright snippets via chromedp.
- Graceful failure when any individual gatherer fails.
- Clean logs with no reverse-DNS noise or `(dest)`-style strings.

### Must NOT Have (Guardrails)
- No active scanning, port scanning, crawling, screenshots, or login handling.
- No third-party ASN APIs unless Team Cymru DNS fails.
- No new HTTP API endpoints beyond existing `/health` and `POST /api/scan`.
- No PDF generation, authentication, or frontend work.
- No redesign of the scanner concurrency model.

---

## Verification Strategy (MANDATORY)

> **ZERO HUMAN INTERVENTION** - ALL verification is agent-executed. No exceptions.

### Test Decision
- **Infrastructure exists**: YES (Go 1.23, existing `go.mod`).
- **Automated tests**: YES (tests-after implementation).
- **Framework**: `go test`.
- **Agent-Executed QA**: ALWAYS — every task includes concrete QA scenarios.

### QA Policy
Every task MUST include agent-executed QA scenarios. Evidence saved to `.omo/evidence/task-{N}-{scenario-slug}.{ext}`.

- **Library/Module**: Use `Bash` (`go test`, `go vet`, `go build`).
- **API/Backend**: Use `Bash` (`curl`) against the running container.
- **Docker**: Use `Bash` (`docker compose`, `docker exec`).

---

## Execution Strategy

### Parallel Execution Waves

```
Wave 1 (Foundation - start immediately):
├── 1. Update .github/FUNDING.yml [quick]
├── 2. Add Go dependencies [quick]
└── 3. Host input normalization utility [quick]

Wave 2 (Core gatherers - after Wave 1):
├── 4. WHOIS gatherer + tests [unspecified-high]
├── 5. ASN/BGP gatherer + tests [unspecified-high]
└── 6. Web/copyright gatherer + tests [unspecified-high]

Wave 3 (Integration - after Wave 2):
├── 7. Wire gatherers into scanner + timeouts [quick]
├── 8. Integration test for scanner.Run [unspecified-high]
└── 9. Docker Compose end-to-end scan [unspecified-high]

Wave FINAL (After ALL tasks):
├── F1. Plan compliance audit (oracle)
├── F2. Code quality review (unspecified-high)
├── F3. Real manual QA (unspecified-high)
└── F4. Scope fidelity check (deep)
-> Present results -> Get explicit user okay
```

### Dependency Matrix

- **1**: - → F2, F3, F4
- **2**: - → 4, 5, 6, 7, 8, 9
- **3**: - → 4, 5, 6, 7, 8, 9
- **4**: 2, 3 → 7, 8, 9
- **5**: 2, 3 → 7, 8, 9
- **6**: 2, 3 → 7, 8, 9
- **7**: 4, 5, 6 → 8, 9
- **8**: 7 → 9, F3
- **9**: 7, 8 → F3
- **F1-F4**: 1-9 → user okay

### Agent Dispatch Summary

- **Wave 1**: 1 → `quick`, 2 → `quick`, 3 → `quick`
- **Wave 2**: 4 → `unspecified-high`, 5 → `unspecified-high`, 6 → `unspecified-high`
- **Wave 3**: 7 → `quick`, 8 → `unspecified-high`, 9 → `unspecified-high`
- **FINAL**: F1 → `oracle`, F2 → `unspecified-high`, F3 → `unspecified-high`, F4 → `deep`

---

## TODOs

- [ ] 1. Clean up `.github/FUNDING.yml`

  **What to do**:
  - Edit `.github/FUNDING.yml` so it contains only the active funding lines.
  - Keep `github: [notfixingit3]`.
  - Keep `custom: ["https://github.com/sponsors/notfixingit3"]`.
  - Remove the empty placeholder lines for `ko_fi`, `liberapay`, `issuehunt`, and `otechie`.
  - Ensure the file is valid YAML.

  **Must NOT do**:
  - Do not add funding links that are not confirmed by the user.
  - Do not modify README badges, sponsor links, or other project metadata.

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: none required

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 2, 3)
  - **Blocks**: F1, F2, F3, F4
  - **Blocked By**: None

  **References**:
  - `.github/FUNDING.yml` — current funding configuration.
  - GitHub Docs: `https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository/displaying-a-sponsor-button-in-your-repository` — FUNDING.yml schema.

  **Acceptance Criteria**:
  - [ ] `.github/FUNDING.yml` contains exactly two active lines.
  - [ ] File parses as valid YAML (verify with `go run` or online YAML parser).

  **QA Scenarios**:

  ```
  Scenario: FUNDING.yml is clean and valid
    Tool: Bash
    Preconditions: None
    Steps:
      1. Run `cat /Users/house/Documents/gitlab/echostate/.github/FUNDING.yml`
      2. Confirm content is:
         github: [notfixingit3]
         custom: ["https://github.com/sponsors/notfixingit3"]
      3. Run `python3 -c "import yaml, sys; yaml.safe_load(open('/Users/house/Documents/gitlab/echostate/.github/FUNDING.yml')); print('valid')"`
    Expected Result: Output contains "valid" and no empty placeholder keys remain.
    Failure Indicators: File contains `ko_fi:`, `liberapay:`, `issuehunt:`, or `otechie:` keys; YAML parse fails.
    Evidence: .omo/evidence/task-1-funding-valid.yml
  ```

  **Evidence to Capture**:
  - [ ] Screenshot or text capture of the validated FUNDING.yml contents.

  **Commit**: YES
  - Message: `chore(github): clean up FUNDING.yml — Jinkies!`
  - Files: `.github/FUNDING.yml`

- [ ] 2. Add reconnaissance Go dependencies

  **What to do**:
  - Add the required Go modules for the gatherers.
  - Run:
    - `go get github.com/likexian/whois`
    - `go get github.com/likexian/whois-parser`
    - `go get github.com/chromedp/chromedp`
  - Run `go mod tidy` to clean `go.mod` and `go.sum`.
  - Verify the project still builds.

  **Must NOT do**:
  - Do not add unused dependencies.
  - Do not upgrade unrelated dependencies.

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: none required

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1, 3)
  - **Blocks**: 4, 5, 6, 7, 8, 9
  - **Blocked By**: None

  **References**:
  - `go.mod` — current module dependencies.
  - `https://pkg.go.dev/github.com/likexian/whois` — WHOIS library API.
  - `https://pkg.go.dev/github.com/likexian/whois-parser` — WHOIS parser API.
  - `https://pkg.go.dev/github.com/chromedp/chromedp` — chromedp API.

  **Acceptance Criteria**:
  - [ ] `go.mod` lists the three new dependencies.
  - [ ] `go.sum` is updated.
  - [ ] `go build ./...` succeeds.

  **QA Scenarios**:

  ```
  Scenario: Dependencies added and project builds
    Tool: Bash
    Preconditions: Working directory is /Users/house/Documents/gitlab/echostate
    Steps:
      1. Run `grep -E 'likexian/whois|chromedp/chromedp' go.mod`
      2. Run `go mod tidy`
      3. Run `go build ./...`
    Expected Result: grep shows all three dependencies; build exits 0.
    Failure Indicators: Missing dependency in go.mod; build fails.
    Evidence: .omo/evidence/task-2-deps-build.txt
  ```

  **Evidence to Capture**:
  - [ ] Terminal output showing dependency additions and successful build.

  **Commit**: YES (group with Task 3)
  - Message: `chore(deps): add recon dependencies and host normalizer — Ruh-roh!`
  - Files: `go.mod`, `go.sum`

- [ ] 3. Implement host input normalization

  **What to do**:
  - Create `internal/scanner/normalize.go`.
  - Implement `NormalizeHost(host string) string` that:
    - Strips `http://`, `https://`, and any other scheme.
    - Removes `www.` prefix.
    - Removes path, query string, fragment, and username/password.
    - Removes port if present.
    - Converts to lowercase.
    - Returns the bare host or IP.
  - Create `internal/scanner/normalize_test.go` with table-driven tests covering:
    - `https://www.example.com/path?query=1#frag` → `example.com`
    - `http://Example.COM:8080/` → `example.com`
    - `www.example.com` → `example.com`
    - `192.168.1.1` → `192.168.1.1`
    - `example.com` → `example.com`
    - Empty string → empty string

  **Must NOT do**:
  - Do not validate TLDs or perform DNS lookups in the normalizer.
  - Do not reject invalid inputs; just strip known parts.

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: none required

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1, 2)
  - **Blocks**: 4, 5, 6, 7, 8, 9
  - **Blocked By**: None

  **References**:
  - `internal/scanner/scanner.go` — where `NormalizeHost` will be used.
  - `internal/models/models.go` — `ScanRequest.Host` is the raw input.

  **Acceptance Criteria**:
  - [ ] `internal/scanner/normalize.go` exists and exports `NormalizeHost`.
  - [ ] `go test ./internal/scanner/...` passes.
  - [ ] All table-driven test cases pass.

  **QA Scenarios**:

  ```
  Scenario: Normalizer handles all expected inputs
    Tool: Bash
    Preconditions: Task 2 dependencies present
    Steps:
      1. Run `go test ./internal/scanner/ -run TestNormalizeHost -v`
    Expected Result: All test cases PASS with correct normalized outputs.
    Failure Indicators: Any test FAIL; function not exported.
    Evidence: .omo/evidence/task-3-normalize-test.txt
  ```

  **Evidence to Capture**:
  - [ ] Test output showing all cases passing.

  **Commit**: YES (group with Task 2)
  - Message: `chore(deps): add recon dependencies and host normalizer — Ruh-roh!`
  - Files: `internal/scanner/normalize.go`, `internal/scanner/normalize_test.go`

- [x] 4. Implement WHOIS gatherer

  **What to do**:
  - Create `internal/scanner/whois.go`.
  - Implement `gatherWHOIS(ctx context.Context, host string) (string, map[string]any, error)`.
  - Use `scanner.NormalizeHost(host)` first.
  - Use `github.com/likexian/whois` to fetch raw WHOIS text with a 10-second context-aware timeout.
  - Use `github.com/likexian/whois-parser` to parse the raw text into structured fields.
  - Return a map with stable keys:
    - `domain` (string)
    - `registrar` (string)
    - `expiration_date` (string)
    - `name_servers` ([]string)
    - `status` ([]string)
    - `raw` (string, truncated to 8KB to avoid huge JSONB)
  - If parsing fails but raw text is available, return the raw text under `raw` and record the parse error in `result.Errors` via the scanner.
  - If WHOIS lookup fails entirely, return an error so the scanner appends it to `result.Errors`.
  - Create `internal/scanner/whois_test.go` with:
    - A test using a known stable domain (e.g., `example.com`) verifying at least `domain` and `registrar` are populated when network is available.
    - A test with an invalid TLD to verify graceful error handling.
    - A unit test mocking raw WHOIS text to verify parser output shape.

  **Must NOT do**:
  - Do not perform recursive WHOIS server discovery manually.
  - Do not store more than 8KB of raw WHOIS text.
  - Do not fail the scan if WHOIS alone fails.

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: none required

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 5, 6)
  - **Blocks**: 7, 8, 9
  - **Blocked By**: 2, 3

  **References**:
  - `internal/scanner/scanner.go` — gatherer signature and error handling pattern.
  - `https://pkg.go.dev/github.com/likexian/whois` — WHOIS lookup API.
  - `https://pkg.go.dev/github.com/likexian/whois-parser` — parser output fields.

  **Acceptance Criteria**:
  - [ ] `internal/scanner/whois.go` exists and compiles.
  - [ ] `go test ./internal/scanner/...` passes.
  - [ ] WHOIS gatherer returns structured data for `example.com` when network is available.
  - [ ] Invalid/empty input returns an error without panicking.

  **QA Scenarios**:

  ```
  Scenario: WHOIS gatherer returns structured data for example.com
    Tool: Bash
    Preconditions: Network access available
    Steps:
      1. Run `go test ./internal/scanner/ -run TestWHOIS -v`
    Expected Result: Test passes; output map contains non-empty `domain` and `registrar`.
    Failure Indicators: Test fails; map is empty; panic.
    Evidence: .omo/evidence/task-4-whois-example.txt

  Scenario: WHOIS gatherer handles invalid input gracefully
    Tool: Bash
    Preconditions: None
    Steps:
      1. Run `go test ./internal/scanner/ -run TestWHOISError -v`
    Expected Result: Test passes; error is returned but no panic.
    Failure Indicators: Test fails; panic.
    Evidence: .omo/evidence/task-4-whois-error.txt
  ```

  **Evidence to Capture**:
  - [ ] Test output for both scenarios.

  **Commit**: YES
  - Message: `feat(scanner): add WHOIS gatherer — Zoinks!`
  - Files: `internal/scanner/whois.go`, `internal/scanner/whois_test.go`

- [x] 5. Implement ASN/BGP gatherer

  **What to do**:
  - Create `internal/scanner/asn.go`.
  - Implement `gatherASN(ctx context.Context, host string) (string, map[string]any, error)`.
  - Use `scanner.NormalizeHost(host)` first.
  - Resolve the host to an IPv4 address using `net.Resolver` with a 5-second timeout.
  - If the input is already an IP, use it directly.
  - Perform a Team Cymru DNS TXT lookup for the reversed IP under `origin.asn.cymru.com`.
  - Example: for IP `1.1.1.1`, query TXT for `1.1.1.1.origin.asn.cymru.com`.
  - Parse the TXT record format: `ASN | BGP Prefix | CC | Registry | Allocated | AS Name`.
  - Return a map with stable keys:
    - `asn` (string)
    - `prefix` (string)
    - `country` (string)
    - `registry` (string)
    - `allocated` (string)
    - `as_name` (string)
    - `ip` (string)
  - If Team Cymru returns no record, return an error so the scanner appends it to `result.Errors`.
  - If DNS resolution fails, return an error.
  - Create `internal/scanner/asn_test.go` with:
    - A unit test parsing a mocked Team Cymru TXT record.
    - A unit test for reverse-IP formatting.
    - A live test against `example.com` (network-dependent, skip if no network).

  **Must NOT do**:
  - Do not use active BGP looking-glass queries.
  - Do not call third-party ASN APIs unless Team Cymru DNS fails.
  - Do not perform port scanning or traceroute.

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: none required

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 4, 6)
  - **Blocks**: 7, 8, 9
  - **Blocked By**: 2, 3

  **References**:
  - `internal/scanner/scanner.go` — gatherer signature.
  - Team Cymru docs: `https://team-cymru.com/community-services/ip-asn-mapping/` — DNS TXT format.
  - `net.Resolver` — stdlib DNS resolver.

  **Acceptance Criteria**:
  - [ ] `internal/scanner/asn.go` exists and compiles.
  - [ ] `go test ./internal/scanner/...` passes.
  - [ ] Parser correctly extracts ASN and AS name from mocked Team Cymru TXT.
  - [ ] Live test against `example.com` returns an ASN when network is available.

  **QA Scenarios**:

  ```
  Scenario: ASN parser handles mocked Team Cymru TXT
    Tool: Bash
    Preconditions: None
    Steps:
      1. Run `go test ./internal/scanner/ -run TestASNParse -v`
    Expected Result: Test passes; parsed map contains expected `asn`, `prefix`, `as_name`.
    Failure Indicators: Parse fields are empty or incorrect.
    Evidence: .omo/evidence/task-5-asn-parse.txt

  Scenario: ASN gatherer resolves example.com when network available
    Tool: Bash
    Preconditions: Network access available
    Steps:
      1. Run `go test ./internal/scanner/ -run TestASNLive -v`
    Expected Result: Test passes or skips gracefully; returned map has non-empty `asn`.
    Failure Indicators: Test fails without skip; map missing `asn`.
    Evidence: .omo/evidence/task-5-asn-live.txt
  ```

  **Evidence to Capture**:
  - [ ] Test output for parser and live scenarios.

  **Commit**: YES
  - Message: `feat(scanner): add ASN/BGP gatherer — Jeepers!`
  - Files: `internal/scanner/asn.go`, `internal/scanner/asn_test.go`

- [x] 6. Implement web/copyright gatherer with chromedp

  **What to do**:
  - Create `internal/scanner/web.go`.
  - Implement `gatherWeb(ctx context.Context, host string) (string, map[string]any, error)`.
  - Use `scanner.NormalizeHost(host)` first.
  - Read `BROWSER_WS_URL` from the config via a package-level variable set during scanner construction, or pass it as a scanner option. Prefer adding a `ScannerOption` or constructor parameter so the gatherer is testable.
  - Connect to the browserless Chrome instance using `chromedp.NewRemoteAllocator`.
  - Create a 10-second context for the chromedp operations.
  - Navigate to `https://<normalized-host>` first; if that fails, fall back to `http://<normalized-host>`.
  - Wait for the page body to be present.
  - Extract:
    - `title` (string) — page title.
    - `copyrights` ([]string) — visible text nodes containing `©`, `Copyright` (case-insensitive), or a 4-digit year (`19xx` or `20xx`).
    - `url` (string) — final URL after navigation/fallback.
  - Deduplicate copyright snippets.
  - Return an error if neither HTTPS nor HTTP succeeds.
  - Create `internal/scanner/web_test.go` with:
    - A unit test verifying the copyright regex/matcher against sample HTML strings.
    - An integration test against `example.com` if a browser endpoint is available (skip otherwise).

  **Must NOT do**:
  - Do not take screenshots.
  - Do not follow links or crawl beyond the landing page.
  - Do not execute arbitrary JavaScript beyond what is needed to render the page.
  - Do not handle logins, cookies, or session state.

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: none required

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 4, 5)
  - **Blocks**: 7, 8, 9
  - **Blocked By**: 2, 3

  **References**:
  - `internal/scanner/scanner.go` — gatherer signature.
  - `internal/config/config.go` — `BrowserWSURL` field.
  - `docker-compose.yml` — browserless/chrome service configuration.
  - `https://pkg.go.dev/github.com/chromedp/chromedp` — chromedp actions.

  **Acceptance Criteria**:
  - [ ] `internal/scanner/web.go` exists and compiles.
  - [ ] `go test ./internal/scanner/...` passes.
  - [ ] Copyright matcher unit test passes.
  - [ ] Integration test against `example.com` returns a title when browser endpoint is available.

  **QA Scenarios**:

  ```
  Scenario: Copyright matcher extracts text from sample HTML
    Tool: Bash
    Preconditions: None
    Steps:
      1. Run `go test ./internal/scanner/ -run TestCopyrightMatcher -v`
    Expected Result: Test passes; expected copyright strings are extracted.
    Failure Indicators: Matcher misses obvious copyright text or extracts unrelated text.
    Evidence: .omo/evidence/task-6-copyright-matcher.txt

  Scenario: chromedp gatherer fetches example.com title via browserless
    Tool: Bash
    Preconditions: Browser endpoint reachable at BROWSER_WS_URL (default ws://localhost:3000/)
    Steps:
      1. Run `go test ./internal/scanner/ -run TestWebLive -v`
    Expected Result: Test passes or skips; returned map has non-empty `title`.
    Failure Indicators: Test fails without skip; connection refused.
    Evidence: .omo/evidence/task-6-web-live.txt
  ```

  **Evidence to Capture**:
  - [ ] Test output for matcher and live scenarios.

  **Commit**: YES
  - Message: `feat(scanner): add chromedp copyright gatherer — Scooby-Dooby-Doo!`
  - Files: `internal/scanner/web.go`, `internal/scanner/web_test.go`

- [x] 7. Wire gatherers into scanner with per-gatherer timeouts

  **What to do**:
  - Refactor `internal/scanner/scanner.go`:
    - Replace the stub `gatherWHOIS`, `gatherASN`, and `gatherWeb` functions with calls to the implementations from Tasks 4, 5, and 6.
    - Pass `BrowserWSURL` into the `Scanner` struct or constructor so the web gatherer can use it.
    - Update `NewScanner` to accept a config struct or functional options (`WithBrowserWSURL`).
    - In `Run`, wrap each gatherer call with `context.WithTimeout(ctx, 10*time.Second)` so a slow gatherer does not starve the others.
    - Call `scanner.NormalizeHost(host)` once at the start of `Run` and pass the normalized host to each gatherer.
  - Update `internal/handlers/scan.go` to pass `cfg.BrowserWSURL` when constructing the scanner.
  - Ensure `result.Errors` still collects per-gatherer failures cleanly.

  **Must NOT do**:
  - Do not change the `Gatherer` function signature.
  - Do not remove the existing concurrency model.
  - Do not introduce global state for configuration.

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: none required

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 3 (sequential after Wave 2)
  - **Blocks**: 8, 9
  - **Blocked By**: 4, 5, 6

  **References**:
  - `internal/scanner/scanner.go` — current orchestration code.
  - `internal/handlers/scan.go` — where `NewScanner()` is called.
  - `internal/config/config.go` — `BrowserWSURL` field.

  **Acceptance Criteria**:
  - [ ] `go build ./...` succeeds.
  - [ ] `go vet ./...` passes.
  - [ ] Stub gatherers no longer return `{"status": "pending"}`.
  - [ ] `Scanner` accepts the browser WebSocket URL.

  **QA Scenarios**:

  ```
  Scenario: Scanner compiles and uses real gatherers
    Tool: Bash
    Preconditions: Tasks 4, 5, 6 completed
    Steps:
      1. Run `go build ./...`
      2. Run `go vet ./...`
      3. Grep scanner.go for "status.*pending" and confirm zero matches.
    Expected Result: Build and vet pass; no pending stubs remain.
    Failure Indicators: Build/vet fails; pending stubs still present.
    Evidence: .omo/evidence/task-7-wire-scanner.txt
  ```

  **Evidence to Capture**:
  - [ ] Build/vet output.
  - [ ] Grep output confirming stubs removed.

  **Commit**: YES
  - Message: `refactor(scanner): wire real gatherers with timeouts — Would you do it for a Scooby Snack?`
  - Files: `internal/scanner/scanner.go`, `internal/handlers/scan.go`

- [x] 8. Add integration test for scanner.Run

  **What to do**:
  - Create `internal/scanner/scanner_integration_test.go` (build-tagged `integration` or environment-gated).
  - The test should:
    - Construct a `Scanner` with the local browserless endpoint if available.
    - Call `scanner.Run(ctx, "example.com")`.
    - Assert that the returned `ScanResult` has non-empty `WHOIS`, `ASN`, or `Web` data (at least one of them) when network/browser are available.
    - Assert that `result.Errors` does not contain panics or unexpected fatal errors.
    - Use `t.Skip` if network or browser is unavailable.
  - Run the test with and without the browser endpoint to verify skip behavior.

  **Must NOT do**:
  - Do not make CI depend on live external services unless a skip fallback exists.
  - Do not assert exact values that depend on third-party data changing.

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: none required

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 3 (after Task 7)
  - **Blocks**: 9
  - **Blocked By**: 7

  **References**:
  - `internal/scanner/scanner.go` — `Run` method signature.
  - `internal/models/models.go` — `ScanResult` structure.

  **Acceptance Criteria**:
  - [ ] Integration test file exists.
  - [ ] `go test ./internal/scanner/ -tags=integration` runs and passes or skips gracefully.
  - [ ] Test asserts non-pending data when services are available.

  **QA Scenarios**:

  ```
  Scenario: Integration test runs against example.com
    Tool: Bash
    Preconditions: Network available; browserless may or may not be running
    Steps:
      1. Run `go test ./internal/scanner/ -tags=integration -v`
    Expected Result: Tests pass or skip; if browser/network available, WHOIS/ASN/Web data is non-empty.
    Failure Indicators: Test fails with panic or timeout.
    Evidence: .omo/evidence/task-8-integration-test.txt
  ```

  **Evidence to Capture**:
  - [ ] Integration test output.

  **Commit**: YES
  - Message: `test(scanner): add integration test for scanner.Run — And I would have gotten away with it too!`
  - Files: `internal/scanner/scanner_integration_test.go`

- [x] 9. Verify full stack with Docker Compose

  **What to do**:
  - Start the full stack:
    - `docker compose down -v`
    - `docker compose up --build -d`
  - Wait for all services healthy (use `docker compose ps` or healthchecks).
  - Send a scan request:
    - `curl -X POST http://localhost:8080/api/scan -H "Content-Type: application/json" -d '{"host":"example.com"}'`
  - Capture the response and verify it contains a snapshot with a `data_hash` and non-empty `raw_data`.
  - Query PostgreSQL directly:
    - `docker exec echostate-db psql -U echostate -d echostate -c "SELECT id, host, data_hash FROM snapshots;"`
  - Verify a row was inserted.
  - Stop the stack:
    - `docker compose down`

  **Must NOT do**:
  - Do not leave containers running after verification.
  - Do not modify production data.

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: none required

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 3 (after Tasks 7, 8)
  - **Blocks**: F1, F2, F3, F4
  - **Blocked By**: 7, 8

  **References**:
  - `docker-compose.yml` — service definitions.
  - `internal/handlers/scan.go` — `POST /api/scan` handler.
  - `internal/db/db.go` — snapshot schema.

  **Acceptance Criteria**:
  - [ ] All three services start and reach healthy state.
  - [ ] `POST /api/scan` returns HTTP 200 with a snapshot object.
  - [ ] PostgreSQL contains at least one snapshot row.
  - [ ] Containers are stopped after verification.

  **QA Scenarios**:

  ```
  Scenario: Full docker-compose scan stores a snapshot
    Tool: Bash
    Preconditions: Docker Desktop or Docker Engine running
    Steps:
      1. Run `docker compose down -v`
      2. Run `docker compose up --build -d`
      3. Wait for health: `docker compose ps`
      4. Run `curl -s -X POST http://localhost:8080/api/scan -H "Content-Type: application/json" -d '{"host":"example.com"}'`
      5. Query DB: `docker exec echostate-db psql -U echostate -d echostate -c "SELECT COUNT(*) FROM snapshots;"`
      6. Run `docker compose down`
    Expected Result: curl returns JSON with `data_hash`; DB count is 1 or more; containers stop cleanly.
    Failure Indicators: Service unhealthy; curl returns non-200; DB count is 0; containers fail to stop.
    Evidence: .omo/evidence/task-9-docker-e2e.txt
  ```

  **Evidence to Capture**:
  - [ ] `docker compose ps` output.
  - [ ] curl response body.
  - [ ] DB query output.

  **Commit**: YES
  - Message: `test(e2e): verify docker-compose scan end-to-end — Let's split up, gang!`
  - Files: no source changes (evidence only)

---

## Final Verification Wave (MANDATORY — after ALL implementation tasks)

> 4 review agents run in PARALLEL. ALL must APPROVE. Present consolidated results to user and get explicit "okay" before completing.

- [ ] F1. **Plan Compliance Audit** — `oracle`
  Read the plan end-to-end. For each "Must Have": verify implementation exists (read file, curl endpoint, run command). For each "Must NOT Have": search codebase for forbidden patterns — reject with file:line if found. Check evidence files exist in `.omo/evidence/`. Compare deliverables against plan.
  Output: `Must Have [N/N] | Must NOT Have [N/N] | Tasks [N/N] | VERDICT: APPROVE/REJECT`

- [ ] F2. **Code Quality Review** — `unspecified-high`
  Run `go vet ./...` + `go test ./...` + `go build ./...`. Review all changed files for: `as any` (N/A in Go but check unsafe/blank imports), empty catches, `fmt.Println` in prod, dead code, unused imports. Check AI slop: excessive comments, over-abstraction, generic names.
  Output: `Build [PASS/FAIL] | Vet [PASS/FAIL] | Tests [N pass/N fail] | Files [N clean/N issues] | VERDICT`

- [ ] F3. **Real Manual QA** — `unspecified-high`
  Start from clean state. Execute EVERY QA scenario from EVERY task — follow exact steps, capture evidence. Test cross-task integration (features working together, not isolation). Test edge cases: empty host, invalid host, gatherer timeout, browser unavailable. Save to `.omo/evidence/final-qa/`.
  Output: `Scenarios [N/N pass] | Integration [N/N] | Edge Cases [N tested] | VERDICT`

- [ ] F4. **Scope Fidelity Check** — `deep`
  For each task: read "What to do", read actual diff (`git log/dev..HEAD`). Verify 1:1 — everything in spec was built (no missing), nothing beyond spec was built (no creep). Check "Must NOT do" compliance. Detect cross-task contamination.
  Output: `Tasks [N/N compliant] | Contamination [CLEAN/N issues] | Unaccounted [CLEAN/N files] | VERDICT`

---

## Commit Strategy

- **Task 1**: `chore(github): clean up FUNDING.yml — Jinkies!`
- **Tasks 2-3**: `chore(deps): add recon dependencies and host normalizer — Ruh-roh!`
- **Task 4**: `feat(scanner): add WHOIS gatherer — Zoinks!`
- **Task 5**: `feat(scanner): add ASN/BGP gatherer — Jeepers!`
- **Task 6**: `feat(scanner): add chromedp copyright gatherer — Scooby-Dooby-Doo!`
- **Task 7**: `refactor(scanner): wire real gatherers with timeouts — Would you do it for a Scooby Snack?`
- **Task 8**: `test(scanner): add integration test for scanner.Run — And I would have gotten away with it too!`
- **Task 9**: `test(e2e): verify docker-compose scan end-to-end — Let's split up, gang!`

---

## Success Criteria

### Verification Commands
```bash
go test ./...           # Expected: PASS
go vet ./...            # Expected: PASS
go build -o echostate ./main.go  # Expected: success
```

### Final Checklist
- [ ] FUNDING.yml contains only active funding lines.
- [ ] WHOIS, ASN, and web gatherers return structured data for `example.com`.
- [ ] Graceful failure is verified for each gatherer.
- [ ] Docker Compose stack runs and stores a snapshot.
- [ ] All evidence files exist in `.omo/evidence/`.
- [ ] No reverse-DNS noise or `(dest)`-style strings in logs.
