# Contributing to EchoState

Thank you for your interest in contributing.

## Getting Started

1. Fork and clone the repository.
2. Copy `.env.example` to `.env`.
3. Start the full stack:

   ```bash
   docker compose up --build
   ```

4. Open the UI at http://localhost:3001 and the API at http://localhost:8080.

### Local development (without full compose)

- **API:** `go run ./main.go` with `DATABASE_URL` and `BROWSER_WS_URL` set.
- **Frontend:** `cd frontend && npm install && npm run dev` with `NEXT_PUBLIC_API_URL=http://localhost:8080` in `.env.local`.

## Branching

- `dev` — integration branch for features and fixes.
- `main` — stable releases only.
- Branch from `dev` for all contributions.

## Pull Request Process

1. Create a feature branch from `dev`.
2. Make changes with tests where applicable.
3. Verify locally:

   ```bash
   go test $(go list ./... | grep -v '/frontend/')
   cd frontend && npm run typecheck && npm run build
   ```

4. Open a pull request targeting `dev`.
5. After review, changes merge to `dev` and are promoted to `main` for release.

## Code Style

### Go

- `gofmt` and `go vet` before committing.
- Keep handlers thin; business logic lives in `internal/` packages.
- Recon gatherers should fail gracefully — partial results are valuable.

### Frontend

- TypeScript strict mode; match existing shadcn/ui patterns.
- Use `data-testid` on interactive elements targeted by Playwright.
- Run `npm run format` if you touch TSX files.

## Commit Messages

- Use [Conventional Commits](https://www.conventionalcommits.org/) (`feat:`, `fix:`, `docs:`, `chore:`, `test:`, etc.).
- Write concise, natural messages — not robotic filler.
- A short Scooby-Doo quote at the end is optional project tradition.

## Releases

The canonical version lives in the root `VERSION` file. Before tagging, bump that file and sync the frontend manifest:

```bash
echo "0.0.1-beta.10" > VERSION
./scripts/sync-version.sh
```

Releases are tagged on `main` (or `dev` for betas) with a `v` prefix, e.g. `v0.0.1-beta.10`.

```bash
git tag -a v0.0.1-beta.10 -m "v0.0.1-beta.10"
git push origin v0.0.1-beta.10
```

Tag pushes trigger GitHub Release binaries and GHCR image builds via Actions.

Update `TODO.md` and `CHANGELOG.md` when completing roadmap items or shipping user-facing changes.