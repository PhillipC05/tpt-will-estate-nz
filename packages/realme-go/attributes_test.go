package realme

import (
	"testing"
	"time"

	"github.com/crewjam/saml"
)

func TestExtractIdentity_Login(t *testing.T) {
	assertion := &saml.Assertion{
		AuthnStatements: []saml.AuthnStatement{
			{
				AuthnInstant: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
				SessionIndex: "sess-001",
				AuthnContext: saml.AuthnContext{
					AuthnContextClassRef: ptr(saml.AuthnContextClassRef(ACLowStrength)),
				},
			},
		},
		AttributeStatements: []saml.AttributeStatement{
			{
				Attributes: []saml.Attribute{
					{Name: AttrFLT, Values: []saml.AttributeValue{{Value: "FLT-ABC123"}}},
				},
			},
		},
	}

	id, err := extractIdentity(assertion)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if id.FLT != "FLT-ABC123" {
		t.Errorf("FLT = %q, want %q", id.FLT, "FLT-ABC123")
	}
	if id.AssuranceLevel != LevelLogin {
		t.Errorf("AssuranceLevel = %v, want LevelLogin", id.AssuranceLevel)
	}
	if id.IsVerified() {
		t.Error("IsVerified() = true, want false for login-level identity")
	}
	if id.SessionIndex != "sess-001" {
		t.Errorf("SessionIndex = %q, want %q", id.SessionIndex, "sess-001")
	}
}

func TestExtractIdentity_Verified(t *testing.T) {
	assertion := &saml.Assertion{
		AuthnStatements: []saml.AuthnStatement{
			{
				AuthnInstant: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
				SessionIndex: "sess-002",
				AuthnContext: saml.AuthnContext{
					AuthnContextClassRef: ptr(saml.AuthnContextClassRef(ACModStrengthOTP)),
				},
			},
		},
		AttributeStatements: []saml.AttributeStatement{
			{
				Attributes: []saml.Attribute{
					{Name: AttrFLT, Values: []saml.AttributeValue{{Value: "FLT-VERIFIED-XYZ"}}},
					{Name: AttrFullName, Values: []saml.AttributeValue{{Value: "Jane Mary Doe"}}},
					{Name: AttrDateOfBirth, Values: []saml.AttributeValue{{Value: "1985-03-22"}}},
					{Name: AttrPlaceOfBirth, Values: []saml.AttributeValue{{Value: "Wellington"}}},
					{Name: AttrGender, Values: []saml.AttributeValue{{Value: "Female"}}},
					{Name: AttrAddressNumber, Values: []saml.AttributeValue{{Value: "12"}}},
					{Name: AttrAddressStreet, Values: []saml.AttributeValue{{Value: "Lambton Quay"}}},
					{Name: AttrAddressCity, Values: []saml.AttributeValue{{Value: "Wellington"}}},
					{Name: AttrAddressPostcode, Values: []saml.AttributeValue{{Value: "6011"}}},
					{Name: AttrAddressCountry, Values: []saml.AttributeValue{{Value: "NZ"}}},
				},
			},
		},
	}

	id, err := extractIdentity(assertion)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if id.FLT != "FLT-VERIFIED-XYZ" {
		t.Errorf("FLT = %q", id.FLT)
	}
	if id.AssuranceLevel != LevelVerified {
		t.Errorf("AssuranceLevel = %v, want LevelVerified", id.AssuranceLevel)
	}
	if !id.IsVerified() {
		t.Error("IsVerified() = false, want true")
	}
	if id.FullName != "Jane Mary Doe" {
		t.Errorf("FullName = %q", id.FullName)
	}
	if id.DateOfBirth.Format("2006-01-02") != "1985-03-22" {
		t.Errorf("DateOfBirth = %v", id.DateOfBirth)
	}
	if id.Gender != "female" {
		t.Errorf("Gender = %q, want lowercase %q", id.Gender, "female")
	}
	if id.Address == nil {
		t.Fatal("Address is nil")
	}
	if id.Address.City != "Wellington" {
		t.Errorf("Address.City = %q", id.Address.City)
	}
	line1 := id.Address.Line1()
	if line1 != "12 Lambton Quay" {
		t.Errorf("Address.Line1() = %q, want %q", line1, "12 Lambton Quay")
	}
}

func TestExtractIdentity_MissingFLT(t *testing.T) {
	assertion := &saml.Assertion{
		AuthnStatements: []saml.AuthnStatement{
			{
				AuthnContext: saml.AuthnContext{
					AuthnContextClassRef: ptr(saml.AuthnContextClassRef(ACLowStrength)),
				},
			},
		},
		AttributeStatements: []saml.AttributeStatement{
			{Attributes: []saml.Attribute{}},
		},
	}

	_, err := extractIdentity(assertion)
	if err == nil {
		t.Error("expected error for missing FLT, got nil")
	}
}

func TestExtractIdentity_NilAssertion(t *testing.T) {
	_, err := extractIdentity(nil)
	if err == nil {
		t.Error("expected error for nil assertion, got nil")
	}
}

func ptr[T any](v T) *T { return &v }
