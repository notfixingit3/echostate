# Contributing to EchoState

Thank you for your interest in contributing.

## Getting Started

1. Fork and clone the repository.
2. Copy `.env.example` to `.env`.
3. Start the full stack:

   ```bash
   docker compose pull
   docker compose up -d
   ```

   Or build from source: `docker compose -f docker-compose.yml -f docker-compose.dev.yml up --build`

4. Open the UI at http://localhost:3001 and the API at http://localhost:8080.

5. Optional — enable passkey auth: see [Authentication](README.md#authentication) in the README (`docker compose run --rm api auth bootstrap-admin`).

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
3. If you add a dependency, gatherer, API route, or worker, update [appmap.md](appmap.md) and [THIRD_PARTY.md](THIRD_PARTY.md) as appropriate.

4. Verify locally:

   ```bash
   go test $(go list ./... | grep -v '/frontend/')
   cd frontend && npm run typecheck && npm run build
   cd frontend && npm run test:e2e
   ```

   E2E boots the full Docker Compose stack (see `frontend/e2e/global-setup.ts`), bootstraps an admin, registers a virtual passkey, then runs smoke, auth, delete, and report specs. CI runs the same suite on every `dev` / `main` push.

5. Open a pull request targeting `dev`.
6. After review, changes merge to `dev` and are promoted to `main` for release.

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