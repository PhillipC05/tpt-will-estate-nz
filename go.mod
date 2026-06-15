module github.com/tpt-nz/tpt-will-estate-nz

go 1.23

require (
	github.com/go-chi/chi/v5 v5.0.12
	github.com/jackc/pgx/v5 v5.6.0
	github.com/tpt-nz/nz-common v0.0.0-00010101000000-000000000000
	github.com/tpt-nz/realme-go v0.0.0-00010101000000-000000000000
)

require (
	github.com/beevik/etree v1.1.0 // indirect
	github.com/crewjam/httperr v0.2.0 // indirect
	github.com/crewjam/saml v0.4.14 // indirect
	github.com/golang-jwt/jwt/v4 v4.4.3 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.1 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/jonboulle/clockwork v0.2.2 // indirect
	github.com/mattermost/xml-roundtrip-validator v0.1.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/russellhaering/goxmldsig v1.4.0 // indirect
	golang.org/x/crypto v0.21.0 // indirect
	golang.org/x/sync v0.1.0 // indirect
	golang.org/x/text v0.14.0 // indirect
)

replace github.com/tpt-nz/nz-common => ./packages/nz-common
replace github.com/tpt-nz/realme-go => ./packages/realme-go

