# Contributing to EchoState

Thank you for your interest in contributing to EchoState.

## Getting Started

1. Clone the repository.
2. Copy `.env.example` to `.env` and adjust values as needed.
3. Run `docker compose up --build` to start the API, database, and browserless Chrome services.

## Pull Request Process

1. Create a feature branch from `dev`.
2. Make your changes and add tests where applicable.
3. Ensure `go test ./...` passes locally.
4. Open a pull request back to `dev`.
5. After review, changes are merged to `dev` and promoted to `main` for release.

## Code Style

- Follow standard Go conventions (`gofmt`, `go vet`).
- Keep handlers thin; business logic belongs in `internal/` packages.
- Fail gracefully on individual reconnaissance tasks.
