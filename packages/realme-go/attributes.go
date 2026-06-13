package realme

import (
	"fmt"
	"strings"
	"time"

	"github.com/crewjam/saml"
)

// RealMe SAML attribute names (OID URNs).
// These are defined in the RealMe SAML attribute profile available from
// developers.realme.govt.nz after service registration.
const (
	// AttrFLT is the Federated Login Token — the opaque persistent user identifier.
	// Present in all RealMe assertions (Login and Verified).
	AttrFLT = "urn:nzl:govt:ict:stds:authn:safeb64:attribute:igovt:IVS:Assertion:Identity:FLT"

	// AttrFullName is the verified legal full name.
	AttrFullName = "urn:nzl:govt:ict:stds:authn:safeb64:attribute:igovt:IVS:Assertion:Identity:FullName"

	// AttrDateOfBirth is the verified date of birth (YYYY-MM-DD).
	AttrDateOfBirth = "urn:nzl:govt:ict:stds:authn:safeb64:attribute:igovt:IVS:Assertion:Identity:DateOfBirth"

	// AttrPlaceOfBirth is the verified place of birth (town/city name).
	AttrPlaceOfBirth = "urn:nzl:govt:ict:stds:authn:safeb64:attribute:igovt:IVS:Assertion:Identity:PlaceOfBirth"

	// AttrGender is the verified gender ("male", "female", "unspecified").
	AttrGender = "urn:nzl:govt:ict:stds:authn:safeb64:attribute:igovt:IVS:Assertion:Identity:Gender"

	// AttrAddressUnit is the unit/apartment number of the verified address.
	AttrAddressUnit = "urn:nzl:govt:ict:stds:authn:safeb64:attribute:igovt:IVS:Assertion:Address:Unit"

	// AttrAddressNumber is the street number of the verified address.
	AttrAddressNumber = "urn:nzl:govt:ict:stds:authn:safeb64:attribute:igovt:IVS:Assertion:Address:Number"

	// AttrAddressStreet is the street name of the verified address.
	AttrAddressStreet = "urn:nzl:govt:ict:stds:authn:safeb64:attribute:igovt:IVS:Assertion:Address:Street"

	// AttrAddressSuburb is the suburb of the verified address.
	AttrAddressSuburb = "urn:nzl:govt:ict:stds:authn:safeb64:attribute:igovt:IVS:Assertion:Address:Suburb"

	// AttrAddressCity is the city of the verified address.
	AttrAddressCity = "urn:nzl:govt:ict:stds:authn:safeb64:attribute:igovt:IVS:Assertion:Address:City"

	// AttrAddressPostcode is the postcode of the verified address.
	AttrAddressPostcode = "urn:nzl:govt:ict:stds:authn:safeb64:attribute:igovt:IVS:Assertion:Address:Postcode"

	// AttrAddressCountry is the country of the verified address.
	AttrAddressCountry = "urn:nzl:govt:ict:stds:authn:safeb64:attribute:igovt:IVS:Assertion:Address:Country"
)

// RealMe Authentication Context Classes — indicate the strength of authentication.
const (
	// ACLowStrength is the authn context for RealMe Login Service (basic login).
	ACLowStrength = "urn:nzl:govt:ict:stds:authn:deployment:GLS:SAML:2.0:ac:classes:LowStrength"

	// ACModStrengthOTP is the authn context for RealMe Verified Identity.
	ACModStrengthOTP = "urn:nzl:govt:ict:stds:authn:deployment:GLS:SAML:2.0:ac:classes:ModStrength+OTP"

	// ACModStrength is the authn context for RealMe Verified Identity (SMS fallback).
	ACModStrength = "urn:nzl:govt:ict:stds:authn:deployment:GLS:SAML:2.0:ac:classes:ModStrength"
)

// extractIdentity parses a crewjam/saml Assertion into an Identity struct.
func extractIdentity(assertion *saml.Assertion) (*Identity, error) {
	if assertion == nil {
		return nil, fmt.Errorf("realme: nil assertion")
	}

	id := &Identity{}

	// Extract authn instant and session index from AuthnStatement.
	if len(assertion.AuthnStatements) > 0 {
		stmt := assertion.AuthnStatements[0]
		id.AuthnInstant = stmt.AuthnInstant
		id.SessionIndex = stmt.SessionIndex

		// Determine assurance level from authentication context class.
		if stmt.AuthnContext.AuthnContextClassRef != nil {
			ref := stmt.AuthnContext.AuthnContextClassRef.Value
			switch {
			case strings.Contains(ref, "ModStrength"):
				id.AssuranceLevel = LevelVerified
			case strings.Contains(ref, "LowStrength"):
				id.AssuranceLevel = LevelLogin
			default:
				id.AssuranceLevel = LevelLogin
			}
		}
	}

	// Build attribute map for easy lookup.
	attrs := make(map[string]string)
	for _, stmt := range assertion.AttributeStatements {
		for _, attr := range stmt.Attributes {
			if len(attr.Values) > 0 {
				attrs[attr.Name] = attr.Values[0].Value
			}
		}
	}

	// FLT is required in all assertions.
	id.FLT = attrs[AttrFLT]
	if id.FLT == "" {
		return nil, fmt.Errorf("realme: assertion missing FLT attribute")
	}

	// Verified-only attributes.
	if id.AssuranceLevel >= LevelVerified {
		id.FullName = attrs[AttrFullName]
		id.PlaceOfBirth = attrs[AttrPlaceOfBirth]
		id.Gender = strings.ToLower(attrs[AttrGender])

		if dob := attrs[AttrDateOfBirth]; dob != "" {
			t, err := time.Parse("2006-01-02", dob)
			if err == nil {
				id.DateOfBirth = t
			}
		}

		addr := &Address{
			Unit:     attrs[AttrAddressUnit],
			Number:   attrs[AttrAddressNumber],
			Street:   attrs[AttrAddressStreet],
			Suburb:   attrs[AttrAddressSuburb],
			City:     attrs[AttrAddressCity],
			Postcode: attrs[AttrAddressPostcode],
			Country:  attrs[AttrAddressCountry],
		}
		if addr.Street != "" || addr.City != "" {
			id.Address = addr
		}
	}

	return id, nil
}
