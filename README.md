# app-will-estate — Digital Will & Estate Platform

> **Legal disclaimer:** This service is not a substitute for legal advice. For complex estates, assets held in trust, or wills involving minors, consult a solicitor. Wills created through this service are intended to comply with the Wills Act 2007 and the Electronic Transactions Act 2002, but DIA / TPT NZ accept no liability for their legal validity.

A secure digital will creation and estate administration platform using RealMe Verified Identity. Wills are encrypted client-side; the service cannot read the content of any will.

---

## Features

- **Client-side encryption** — will content is encrypted in the browser before upload; the server stores only ciphertext.
- **RealMe Verified Identity** for testators and witnesses (both must hold a verified identity).
- **Two-witness signing** — witnesses counter-sign using their own RealMe verified identity, satisfying the Wills Act 2007 section 11.
- **BDM death notification hook** — the vault is unlocked only after a registered death notification is received from Births, Deaths and Marriages.
- **Executor access** — nominated executors access the will via their RealMe verified identity after death notification.
- **Beneficiary notification** — beneficiaries receive an email and can view their specific bequest.

---

## Setup

### Prerequisites

- Go 1.23+
- Node 20+ / pnpm 9+
- Docker Compose (for local infra)
- RealMe MTS credentials (see [realme-go README](../../realme-go/README.md))

### Local development

```bash
# From project root — start postgres, redis, nats
make dev

# Start app + mock RealMe IdP
cd packages/app-will-estate
docker compose up -d

# API: http://localhost:8083
# Frontend: cd web && pnpm dev
```

### Environment variables

| Variable | Default | Description |
|----------|---------|-------------|
| `LISTEN_ADDR` | `:8080` | HTTP listen address |
| `DATABASE_URL` | — | PostgreSQL connection string |
| `REALME_ENVIRONMENT` | `mts` | `mts`, `ite`, or `production` |
| `REALME_CERT_FILE` | — | Path to SP signing certificate |
| `REALME_KEY_FILE` | — | Path to SP private key |
| `REALME_ENTITY_ID` | — | SP entity ID (must match DIA registration) |
| `REALME_ACS_URL` | — | Assertion Consumer Service URL |
| `BDM_WEBHOOK_SECRET` | — | HMAC secret for BDM death notification webhook |

### Database migration

```bash
atlas schema apply \
  --dir "file://packages/app-will-estate/migrations" \
  --url "$DATABASE_URL" \
  --auto-approve
```

---

## RealMe Registration

This app requires the **Assertion Service** (Verified Identity) tier. Before registering with DIA:

1. Complete your MTS integration and automated tests.
2. Request ITE access at [developers.realme.govt.nz](https://developers.realme.govt.nz/).
3. Submit a Privacy Impact Assessment (PIA) — a template is in `docs/pia.md`.
4. Obtain Production approval — requires CA-signed certificates and DIA compliance sign-off.

### Applicable legislation

- Wills Act 2007
- Electronic Transactions Act 2002
- Privacy Act 2020 (Principle 3 consent before collecting identity)
- Administration Act 1969 (estate administration)
