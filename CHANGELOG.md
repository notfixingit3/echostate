# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.0.1-beta.0] - 2026-06-17

First public beta — API, web UI, PDF reports, and Docker Compose stack.

### Added

- Go/Gin JSON API with concurrent WHOIS, ASN/BGP, and web reconnaissance gatherers.
- PostgreSQL persistence for targets, snapshots, and reports.
- Async PDF report worker with headless Chrome rendering.
- Browse API: targets, snapshots, and reports with pagination.
- Per-IP rate limiting (30 scans/min) on `POST /api/scan`.
- pWhois async worker for submitter IP enrichment.
- Next.js web UI: scan form, targets, snapshots, reports, snapshot/target detail pages.
- Dark mode toggle (light / dark / system).
- Country flags and expanded intel display (ASN, web title, registrar, etc.).
- `GET /api/targets/:id` target detail endpoint with latest snapshot.
- Docker Compose: API, frontend, PostgreSQL, browserless Chrome.
- GitHub Actions CI, release binaries, and GHCR container publishing.
- Playwright end-to-end tests for scan and report flows.

### Security

- Trusted proxy configuration for accurate client IP behind reverse proxies.
- Report error responses sanitized; Slowloris mitigation on HTTP server.

[0.0.1-beta.0]: https://github.com/notfixingit3/echostate/releases/tag/v0.0.1-beta.0