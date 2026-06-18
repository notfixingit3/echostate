# Security Policy

## Supported Versions

| Version        | Supported |
| -------------- | --------- |
| main           | Yes       |
| dev            | Preview   |
| v0.0.1-beta.x  | Beta      |

## Reporting a Vulnerability

Please report security vulnerabilities to [notfixingit@yahoo.com](mailto:notfixingit@yahoo.com).

Do not open public issues for security-sensitive bugs.

## Notes

- **Authentication (beta.20+)** — Passkey + enrollment-code auth. Full setup (Docker and local): [README § Authentication](README.md#authentication). Until the first admin is bootstrapped, the API behaves as before; after users exist, protected routes require a session cookie.
- Set `ECHOSTATE_AUTH_PEPPER` in production for enrollment-code and session hashing.
- Set `ECHOSTATE_BREAK_GLASS_SECRET` for CLI recovery codes (`echostate auth issue-admin-code`).
- Auth defaults (code length/TTL, session lifetime, attempt limits, WebAuthn RP ID/origin) are admin-configurable in **Settings**.
- Deploy behind a reverse proxy with TLS in production.
- Configure Gin `TrustedProxies` appropriately when running behind load balancers.