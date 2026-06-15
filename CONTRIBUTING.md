# Contributing to tpt-will-estate-nz

Thank you for your interest in contributing. This document covers the development setup, PR workflow, and coding standards for this project.

## Development setup

### Prerequisites

- Go 1.23+
- Node 20+ / pnpm 9+
- Docker Compose

### Start local infrastructure

```bash
# PostgreSQL 16, Redis 7, NATS 2.10 + app container
docker compose up -d

# Apply DB migrations
make migrate

# Run the API server locally (without Docker)
make dev

# Run the Next.js frontend
cd web && pnpm install && pnpm dev
```

### Environment variables

Copy `.env.example` to `.env` and fill in the required values. For local development the Docker Compose file sets sensible defaults; you only need a valid RealMe MTS certificate pair.

### Mock RealMe IdP

The `packages/realme-go/cmd/mock-idp` binary simulates a RealMe IdP for local development. It is started automatically via `docker compose up` using the `realme-go` test environment.

## Workflow

1. **Fork** the repository and create a branch from `main`.
2. **One concern per PR** — keep changes focused. Large PRs are hard to review.
3. **Write tests** for new behaviour. The project uses standard `go test`; run `make test` before pushing.
4. **Run linting** with `make lint` (golangci-lint). Fix all errors before opening a PR.
5. **Open a PR** against `main` with a clear description of *why* the change is needed, not just *what* it does.

## Commit style

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add beneficiary email notification
fix: correct SQL placeholder count in CreateDraft
docs: clarify RealMe registration steps
refactor: extract identityClaims helper
test: add will signing integration test
```

## Code standards

- **Go**: idiomatic Go, no unused imports, `gofmt`/`goimports` formatting.
- **Error handling**: return errors; do not `log.Fatal` inside packages.
- **No plaintext PII in logs** — log will IDs and FLTs only; never log full names or addresses at INFO level.
- **Security**: no hardcoded secrets. Use environment variables. Never commit `.env` or certificate files.

## Sensitive areas

Changes to any of the following require extra care and a security-focused review:

| Area | Why |
|------|-----|
| `internal/services/encryption.go` | Validates vault ciphertext; any weakening breaks client-side encryption guarantees |
| `internal/services/death_trigger.go` | Controls will-unlock timing; premature unlock exposes private estate data |
| `packages/realme-go/` | RealMe SAML integration; auth bugs could allow identity spoofing |
| `migrations/` | Schema changes are irreversible in production |

## Reporting security issues

Do **not** open a public GitHub issue for security vulnerabilities. Email **security@tpt.nz** with a description and reproduction steps. We aim to respond within 48 hours.

## License

By submitting a pull request you agree that your contribution will be licensed under the [MIT License](LICENSE).
