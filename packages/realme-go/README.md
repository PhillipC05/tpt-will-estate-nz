# realme-go

A Go SAML 2.0 / OIDC client for New Zealand's [RealMe](https://www.realme.govt.nz/) identity service.

No official Go library exists for RealMe. This package wraps [`crewjam/saml`](https://github.com/crewjam/saml) with RealMe-specific SAML attribute mappings, environment configuration, Chi middleware, and a mock IdP for local development.

## Features

- SAML 2.0 SP implementation for RealMe Login and Assertion (Verified Identity) services
- Chi-compatible `RequireLogin` and `RequireVerified` middleware
- Signed JWT cookie session store (swap for Redis store in multi-instance deployments)
- Mock IdP for local development — no DIA credentials required
- Certificate expiry warnings
- MTS / ITE / Production environment management

## Quick Start

```go
import (
    "github.com/tpt-nz/realme-go"
    "github.com/go-chi/chi/v5"
)

cfg := realme.Config{
    Environment:     realme.MTS,
    EntityID:        "https://myapp.example.nz/saml/metadata",
    ACSURL:          "https://myapp.example.nz/auth/realme/callback",
    CertFile:        "certs/mts.login.myapp.example.nz.crt",
    KeyFile:         "certs/mts.login.myapp.example.nz.key",
    IdPMetadataFile: "certs/mts-idp-metadata.xml",
    ForceAuthn:      true,
}

sp, err := realme.NewProvider(cfg)
if err != nil {
    log.Fatal(err)
}

r := chi.NewRouter()

// Auth routes (unprotected)
r.Get("/saml/metadata",         sp.MetadataHandler())
r.Get("/auth/realme/login",     sp.LoginHandler())
r.Post("/auth/realme/callback", sp.CallbackHandler(onLogin))
r.Get("/auth/realme/logout",    sp.LogoutHandler())

// Protected routes
r.With(sp.RequireLogin()).Get("/dashboard", dashboardHandler)
r.With(sp.RequireVerified()).Post("/incorporate", incorporateHandler)
```

```go
func onLogin(w http.ResponseWriter, r *http.Request, identity *realme.Identity) {
    // Upsert your application user from the FLT.
    // identity.FLT is the stable identifier — never use name/DOB as the key.
    http.Redirect(w, r, "/dashboard", http.StatusFound)
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
    identity := realme.IdentityFromContext(r.Context())
    fmt.Fprintf(w, "Hello, FLT=%s (level=%s)", identity.FLT, identity.AssuranceLevel)
}
```

## RealMe Integration Steps

### 1. Register your service

1. Go to [developers.realme.govt.nz](https://developers.realme.govt.nz/) and register
2. Create a service integration project for your app
3. Choose the service(s) you need: **Login Service** and/or **Assertion Service**

### 2. Generate MTS certificates

DIA requires your certificates to follow this naming convention:

```
{env}.{service}.{organisation_domain}
```

Example:
- `mts.login.myapp.example.nz` ← MTS Login Service
- `mts.assertion.myapp.example.nz` ← MTS Assertion Service
- `login.myapp.example.nz` ← Production Login Service

For MTS, self-signed certificates are accepted. For ITE and Production, obtain
certificates from an accredited CA (at least 1-year validity required).

### 3. Download IdP metadata

Download the IdP metadata XML for your environment and store it locally:

| Environment | Login Service | Assertion Service |
|-------------|--------------|------------------|
| MTS | https://mts.realme.govt.nz/saml2/metadata | https://mts.realme.govt.nz/saml2/assertion/metadata |
| ITE | https://www.ite.logon.realme.govt.nz/saml2/metadata | https://www.ite.assertion.realme.govt.nz/saml2/metadata |
| Production | https://www.realme.govt.nz/saml2/metadata | https://www.assertion.realme.govt.nz/saml2/metadata |

### 4. Generate and submit SP metadata

Start your app and download the SP metadata:

```bash
curl https://myapp.example.nz/saml/metadata > sp-metadata.xml
```

Upload `sp-metadata.xml` to the DIA developer portal for your service registration.

### 5. Test in MTS

Set `REALME_ENV=mts` and use the test users from the RealMe MTS test user catalogue.

### 6. Apply for ITE and Production

After successful MTS testing, request ITE access from DIA via integrations@realme.govt.nz.
Production access requires formal approval.

## Local Development (Mock IdP)

No DIA credentials needed for local development:

```go
import "github.com/tpt-nz/realme-go/testenv"

// In your test:
idp := testenv.NewMockIdP(t)
idp.SetNextUser(testenv.UserVerified)

cfg := realme.Config{
    IdPMetadataURL: idp.MetadataURL(),
    CertFile:       idp.SPCertFile(),
    KeyFile:        idp.SPKeyFile(),
    // ...
}
```

## Privacy

- Store only the FLT as your user identifier, not name/DOB
- The FLT is opaque and per-service — the same person gets a different FLT for each registered service
- RealMe deliberately does not include the user's email address (Privacy Act Principle 13)
- Only request the Assertion Service (Verified Identity) when legally necessary

## Licence

MIT
