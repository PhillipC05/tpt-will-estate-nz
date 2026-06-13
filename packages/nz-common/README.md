# nz-common

Shared Go utilities for all TPT NZ civic apps. Import path: `github.com/tpt-nz/nz-common`.

---

## Packages

### `mbie/` — MBIE Business APIs

Client for the Ministry of Business, Innovation and Employment APIs at `api.business.govt.nz`.

```go
client := mbie.NewClient(os.Getenv("MBIE_API_KEY"))

// Companies Register
result, err := client.SearchCompanies(ctx, "Acme", 20)
company, err := client.GetCompanyByNumber(ctx, "1234567")

// NZ Business Number
entity, err := client.GetEntityByNZBN(ctx, "9429039822327")
results, err := client.SearchByNZBN(ctx, "Acme", 20, 1)
```

Register for an API key: https://portal.api.business.govt.nz/

### `linz/` — LINZ Data Service

Client for Land Information New Zealand WFS/WMS datasets.

```go
client := linz.NewClient(os.Getenv("LINZ_API_KEY"))

// Address search (NZ Physical Addresses, layer 53353)
addresses, err := client.SearchAddresses(ctx, "1 Queen Street", 10)
address, err := client.GetAddressByID(ctx, 2229143)

// Property boundary query (NZ Parcels, layer 51564)
boundaries, err := client.GetPropertyBoundaries(ctx, -36.8485, 174.7633)
```

Register for an API key: https://data.linz.govt.nz/

### `health/` — Health NZ FHIR APIs

Client for Health New Zealand FHIR R4 APIs (NHI, SDHR, HPI, MWS). Requires OAuth2 client credentials — register at the Health NZ Digital Services Hub.

```go
nhiClient := health.NewNHIClient(bearerToken)

// NHI lookup
result, err := nhiClient.GetPatientByNHI(ctx, "ZAA0001")

// Demographic search (when NHI unknown)
patients, err := nhiClient.SearchPatientByDemographics(ctx, "Smith", "John", "1985-03-15")
```

Only call after obtaining explicit consent per the Health Information Privacy Code rule 10. Never log NHI numbers at INFO level or above.

### `bdm/` — Births Deaths Marriages

Stub client for BDM death notification hooks. No public BDM API exists; this package provides the interface for when one becomes available.

### `auth/`

- `session.go` — Redis-backed session store (UUID session ID → user data)
- `token.go` — JWT creation/validation helpers

### `middleware/`

Chi-compatible HTTP middleware:

| Middleware | Header/Behaviour |
|------------|-----------------|
| `Logger` | Structured slog request logging |
| `RequestID` | `X-Request-ID` injection |
| `CORS` | Configurable CORS headers |
| `Idempotency` | `Idempotency-Key` enforcement (PostgreSQL-backed) |

The `Idempotency` middleware requires an `idempotency_keys` table in your app's schema:

```sql
CREATE TABLE IF NOT EXISTS idempotency_keys (
    key         TEXT        PRIMARY KEY,
    status_code INT         NOT NULL,
    body        JSONB       NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

### `pagination/`

Cursor-based pagination helpers compatible with `pgx/v5`.

### `money/`

`int64` NZD money type. Never use `float64` for currency.

```go
price := money.NZD(1099) // $10.99
fmt.Println(price.Format()) // "$10.99"
```

---

## Development

```bash
go test ./...
go vet ./...
```

Tests use real PostgreSQL where needed (see CI). Do not mock the database.
