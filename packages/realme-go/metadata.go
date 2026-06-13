package realme

// metadata.go — SP metadata XML generation for DIA registration.
//
// SP metadata is served at runtime by MetadataHandler (see handler.go), which
// delegates to the crewjam/saml ServiceProvider.Metadata() method. This file
// documents the metadata format and the registration process.
//
// To obtain your SP metadata XML for DIA registration:
//
//  1. Start your app with the MTS environment configuration.
//  2. Request GET /saml/metadata (or wherever MetadataHandler is mounted).
//  3. Save the response XML.
//  4. Submit the XML to developers.realme.govt.nz along with your service
//     details and compliance checklist.
//
// The metadata XML includes:
//   - entityID: must match Config.EntityID exactly
//   - AssertionConsumerService URL: must match Config.ACSURL
//   - SP signing certificate (public key only, never the private key)
//   - NameIDFormat: urn:oasis:names:tc:SAML:2.0:nameid-format:persistent
//   - AuthnContextClassRef: at least ip-password-protected-transport (Login)
//     or urn:nzl:govt:ict:stds:authn:deployment:GLS:SAML:2.0:ac:classes:
//     ModStrength (Verified Identity Assertion Service)
//
// Metadata validity:
//   - DIA validates your metadata XML before approving integration.
//   - Update your DIA registration when the SP certificate is renewed.
//   - Certificate expiry within 30 days triggers a DIA notification.
