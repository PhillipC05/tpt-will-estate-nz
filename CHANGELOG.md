# Changelog — Digital Will & Estate Platform

All notable changes to this package are documented here.
Format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).
This project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

---

## [0.1.0] — 2026-06-15

### Added

- Core will lifecycle: `CreateDraft`, `SignTestator`, `SignWitness`, `Lock`, `UnlockAtDeath`.
- RealMe SAML 2.0 Verified Identity integration (MTS environment).
- Client-side encryption vault — server stores only ciphertext/nonce/alg.
- BDM death notification webhook with HMAC-SHA256 signature verification.
- Executor access endpoint (post-death, vault-unlocked state).
- Beneficiary notification endpoint (stub — wire in an email provider).
- PostgreSQL schema with Atlas migrations and `updated_at` trigger.
- Docker Compose local development setup (PostgreSQL 16, Redis 7, NATS 2.10).
- Next.js 14 frontend: multi-step will wizard, executor page, beneficiary page.
- `packages/realme-go` — standalone Go SAML 2.0 client for NZ RealMe.
- `packages/nz-common` — shared NZ civic utilities (MBIE, LINZ, Health NZ, BDM).
- MIT license, README, CONTRIBUTING guide, `.env.example`, GitHub Actions CI.

### Fixed

- SQL placeholder count mismatch in `WillRepo.CreateDraft` (`$8` → `$9`).
- Missing address-of operator (`&`) on nullable pointer variables in `WillRepo.GetByID` Scan.
- `CreateDraft` repository inserts now run inside a single transaction.
- `pgxpool.New` replaced with `pgxpool.NewWithConfig` + `context.Background()`.
- Import conflict between `chi/middleware` and `nz-common/middleware` resolved.
- `realme.NewProvider` wired from environment variables; removed non-existent `NewProviderFromEnv`.
- `DeathTrigger.now()` return type corrected from `models.WillStatus` to `time.Time`.
- Encryption service now validates algorithm against an allowlist (`AES-GCM-256`).

[Unreleased]: https://github.com/tpt-nz/tpt-will-estate-nz/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/tpt-nz/tpt-will-estate-nz/releases/tag/v0.1.0
