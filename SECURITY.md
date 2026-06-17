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

- EchoState has **no authentication** — browse endpoints and submitter IPs are publicly visible by design.
- Deploy behind a reverse proxy with TLS in production.
- Configure Gin `TrustedProxies` appropriately when running behind load balancers.