package testenv

import "time"

// TestUser represents a synthetic RealMe user for MTS/local testing.
type TestUser struct {
	FLT          string
	AssuranceLevel int // 1 = login, 2 = verified
	FullName     string
	DateOfBirth  time.Time
	PlaceOfBirth string
	Gender       string
	AddressUnit  string
	AddressNum   string
	Street       string
	Suburb       string
	City         string
	Postcode     string
}

// Pre-built test users matching the RealMe MTS test user catalogue.
var (
	// UserLoginOnly is a basic login user with no verified identity attributes.
	UserLoginOnly = &TestUser{
		FLT:          "FLT-TEST-LOGIN-001",
		AssuranceLevel: 1,
		FullName:     "",
	}

	// UserVerified is a fully verified user with all identity attributes present.
	UserVerified = &TestUser{
		FLT:          "FLT-TEST-VERIFIED-001",
		AssuranceLevel: 2,
		FullName:     "Jane Mary Doe",
		DateOfBirth:  time.Date(1985, 3, 22, 0, 0, 0, 0, time.UTC),
		PlaceOfBirth: "Wellington",
		Gender:       "female",
		AddressNum:   "12",
		Street:       "Lambton Quay",
		Suburb:       "Wellington Central",
		City:         "Wellington",
		Postcode:     "6011",
	}

	// UserVerified2 is a second verified test user for multi-user scenarios
	// (e.g., testing both granter and attorney in the POA registry).
	UserVerified2 = &TestUser{
		FLT:          "FLT-TEST-VERIFIED-002",
		AssuranceLevel: 2,
		FullName:     "Robert James Smith",
		DateOfBirth:  time.Date(1972, 11, 5, 0, 0, 0, 0, time.UTC),
		PlaceOfBirth: "Auckland",
		Gender:       "male",
		AddressNum:   "45",
		Street:       "Queen Street",
		Suburb:       "Auckland Central",
		City:         "Auckland",
		Postcode:     "1010",
	}

	// UserVerifiedNoAddress is a verified user with no address attribute —
	// for testing applications that can tolerate a missing address.
	UserVerifiedNoAddress = &TestUser{
		FLT:          "FLT-TEST-VERIFIED-NOADDR-001",
		AssuranceLevel: 2,
		FullName:     "Alex Taylor",
		DateOfBirth:  time.Date(1990, 7, 14, 0, 0, 0, 0, time.UTC),
		PlaceOfBirth: "Christchurch",
		Gender:       "unspecified",
	}
)
