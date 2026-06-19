# Security Policy

## Supported Versions

| Version        | Supported |
| -------------- | --------- |
| v0.0.1         | Yes       |
| main           | Yes       |
| dev            | Preview   |
| v0.0.2-dev.x   | Dev       |
| v0.0.1-beta.x  | Beta      |

## Reporting a Vulnerability

Please report security vulnerabilities to [notfixingit@yahoo.com](mailto:notfixingit@yahoo.com).

Do not open public issues for security-sensitive bugs.

## Notes

- **Authentication (beta.20+)** — Passkey + enrollment-code auth. Full setup (Docker and local): [README § Authentication](README.md#authentication). Until the first admin is bootstrapped, the API behaves as before; after users exist, protected routes require a session cookie.
- **`ECHOSTATE_AUTH_PEPPER` is required in production** — the API refuses to start with `ECHOSTATE_ENV=production` unless this is set to a strong random value (not the dev default).
- Enrichment integration keys (Shodan, Censys, etc.) are stored in settings and may appear in snapshot `raw_data` when enrichment runs. Treat database backups and exports as sensitive; enrichment error messages are redacted before persistence.
- Set `ECHOSTATE_BREAK_GLASS_SECRET` for CLI recovery codes (`echostate auth issue-admin-code`).
- Auth defaults (code length/TTL, session lifetime, attempt limits, WebAuthn RP ID/origin) are admin-configurable in **Settings**.
- Deploy behind a reverse proxy with TLS in production.
- Configure Gin `TrustedProxies` appropriately when running behind load balancers.