package testenv_test

import (
	"testing"

	"github.com/tpt-nz/realme-go"
	"github.com/tpt-nz/realme-go/testenv"
)

func TestNewMockIdP_StartsAndServes(t *testing.T) {
	idp := testenv.NewMockIdP(t)

	if idp.MetadataURL() == "" {
		t.Fatal("MetadataURL is empty")
	}
	if idp.SPCertFile() == "" {
		t.Fatal("SPCertFile is empty")
	}
	if idp.SPKeyFile() == "" {
		t.Fatal("SPKeyFile is empty")
	}
}

func TestMockIdP_ProviderCreation(t *testing.T) {
	idp := testenv.NewMockIdP(t)

	cfg := realme.Config{
		Environment:    realme.MTS,
		EntityID:       "https://app.test/saml/metadata",
		ACSURL:         "https://app.test/auth/realme/callback",
		CertFile:       idp.SPCertFile(),
		KeyFile:        idp.SPKeyFile(),
		IdPMetadataURL: idp.MetadataURL(),
	}

	provider, err := realme.NewProvider(cfg)
	if err != nil {
		t.Fatalf("NewProvider with mock IdP: %v", err)
	}
	if provider == nil {
		t.Fatal("expected non-nil provider")
	}
}

func TestMockIdP_SetNextUser(t *testing.T) {
	idp := testenv.NewMockIdP(t)

	// Setting next user should not panic.
	idp.SetNextUser(testenv.UserLoginOnly)
	idp.SetNextUser(testenv.UserVerified)
	idp.SetNextUser(testenv.UserVerified2)
	idp.SetNextUser(testenv.UserVerifiedNoAddress)
}

func TestMockIdP_TestUsers_Fields(t *testing.T) {
	if testenv.UserLoginOnly.FLT == "" {
		t.Error("UserLoginOnly.FLT is empty")
	}
	if testenv.UserVerified.FLT == "" {
		t.Error("UserVerified.FLT is empty")
	}
	if testenv.UserVerified.FullName == "" {
		t.Error("UserVerified.FullName is empty")
	}
	if testenv.UserVerified.AssuranceLevel != realme.LevelVerified {
		t.Errorf("UserVerified.AssuranceLevel = %v, want LevelVerified", testenv.UserVerified.AssuranceLevel)
	}
	if testenv.UserLoginOnly.AssuranceLevel != realme.LevelLogin {
		t.Errorf("UserLoginOnly.AssuranceLevel = %v, want LevelLogin", testenv.UserLoginOnly.AssuranceLevel)
	}
}

func TestMockIdP_UserVerifiedNoAddress(t *testing.T) {
	user := testenv.UserVerifiedNoAddress
	if user.FLT == "" {
		t.Error("UserVerifiedNoAddress.FLT is empty")
	}
	// This user intentionally has no address.
	if user.Address != nil {
		t.Errorf("UserVerifiedNoAddress should have nil Address, got %+v", user.Address)
	}
}
