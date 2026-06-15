package models

import "time"

// WillID is a stable public identifier for a will (UUID v4 string).
type WillID string

// WillStatus represents the current lifecycle state of a will.
type WillStatus string

const (
	// WillStatusDraft — will is being prepared; not yet signed.
	WillStatusDraft WillStatus = "draft"

	// WillStatusLocked — will has been fully signed (testator + 2 witnesses) and
	// sealed. No further edits are permitted. Vault is inaccessible until death.
	WillStatusLocked WillStatus = "locked"

	// WillStatusDeathNotified — BDM death notification received; vault is in the
	// mandatory cooling-off period before the executor can access it.
	WillStatusDeathNotified WillStatus = "death_notified"

	// WillStatusUnlockedDead — cooling-off period has elapsed; vault is accessible
	// to nominated executors.
	WillStatusUnlockedDead WillStatus = "unlocked_at_death"

	// WillStatusSuperseded — this will has been replaced by a newer will created
	// by the same testator. The Wills Act 2007 treats a new will as revoking all
	// prior wills unless stated otherwise.
	WillStatusSuperseded WillStatus = "superseded"
)

// IdentityClaims holds the RealMe identity attributes captured at the time of
// a signature or access event. Stored alongside each action for audit purposes.
type IdentityClaims struct {
	RealMeFLT      string
	FullName       string
	AssuranceLevel string
	IsVerified     bool
}

// Clause is a single clause within a will.
type Clause struct {
	ID    string `json:"id"`
	Text  string `json:"text"`
	Index int    `json:"index"`
}

// Beneficiary is a person named to receive a bequest.
type Beneficiary struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Relation string `json:"relation"`
}

// Executor is a person nominated to administer the estate.
// FLT is the executor's RealMe Federated Login Token, used to verify identity
// at access time. Set by the executor claiming the role via RealMe before the
// will is locked, or captured on first access after death notification.
type Executor struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	FLT   string `json:"flt"` // RealMe FLT — used to gate executor access post-death
}

// WitnessSignature records one witness counter-signature as required by
// Wills Act 2007 s.11.
type WitnessSignature struct {
	Witness     IdentityClaims `json:"witness"`
	SignedAt    time.Time      `json:"signedAt"`
	PayloadHash string         `json:"payloadHash"`
}

// Will is the central will record.
type Will struct {
	ID        WillID         `json:"id"`
	Testator  IdentityClaims `json:"testator"`
	Clauses   []Clause       `json:"clauses"`

	// Signing fields
	TestatorSignedAt    *time.Time         `json:"testatorSignedAt,omitempty"`
	TestatorPayloadHash *string            `json:"testatorPayloadHash,omitempty"`
	WitnessSignatures   []WitnessSignature `json:"witnessSignatures"`

	// Clause collections
	Beneficiaries       []Beneficiary       `json:"beneficiaries"`
	Executors           []Executor          `json:"executors"`
	DigitalAssetClauses []DigitalAssetClause `json:"digitalAssetClauses"`
	PropertyClauses     []PropertyClause    `json:"propertyClauses"`

	// Lifecycle
	Status   WillStatus `json:"status"`
	LockedAt *time.Time `json:"lockedAt,omitempty"`

	// Feature: supersession — this will was replaced by a newer will.
	SupersededBy *WillID `json:"supersededBy,omitempty"`

	// Feature: codicil — this will is an amendment to an existing locked will.
	ParentWillID *WillID `json:"parentWillId,omitempty"`
	IsCodicil    bool    `json:"isCodicil"`

	// Feature: time-lock vault — mandatory cooling-off after death notification.
	DeathNotifiedAt      *time.Time `json:"deathNotifiedAt,omitempty"`
	VaultUnlockDelayDays int        `json:"vaultUnlockDelayDays"`
	UnlockAfter          *time.Time `json:"unlockAfter,omitempty"`
	UnlockedAt           *time.Time `json:"unlockedAt,omitempty"`

	// Feature: testator liveness check.
	LastConfirmedAt    *time.Time `json:"lastConfirmedAt,omitempty"`
	LivenessNotifiedAt *time.Time `json:"livenessNotifiedAt,omitempty"`

	// Encrypted vault metadata — server never stores or processes the plaintext.
	VaultCiphertext string `json:"vaultCiphertext"`
	VaultNonce      string `json:"vaultNonce"`
	VaultAlg        string `json:"vaultAlg"`

	AuditNote string `json:"auditNote"`
}

// ── Digital asset clause ──────────────────────────────────────────────────────

// DigitalAssetType classifies the kind of digital asset being bequeathed.
type DigitalAssetType string

const (
	DigitalAssetCryptoWallet    DigitalAssetType = "crypto_wallet"
	DigitalAssetSocialMedia     DigitalAssetType = "social_media"
	DigitalAssetEmail           DigitalAssetType = "email"
	DigitalAssetPasswordManager DigitalAssetType = "password_manager"
	DigitalAssetCloudStorage    DigitalAssetType = "cloud_storage"
	DigitalAssetOther           DigitalAssetType = "other"
)

// DigitalAssetClause describes a digital asset bequest.
// SECURITY: never store passwords, seed phrases, or private keys in this record.
// The Instruction field must contain access guidance only (e.g., "recovery key is
// in the fireproof safe at the family home"), not the credentials themselves.
type DigitalAssetClause struct {
	ID            string           `json:"id"`
	WillID        WillID           `json:"willId"`
	ClauseIndex   int              `json:"clauseIndex"`
	AssetType     DigitalAssetType `json:"assetType"`
	Platform      string           `json:"platform"`      // "Coinbase", "Gmail", "1Password", etc.
	Identifier    string           `json:"identifier"`    // wallet address or username — NOT password
	Instruction   string           `json:"instruction"`   // how the executor should access or transfer
	BeneficiaryID string           `json:"beneficiaryId"` // empty = estate generally
}

// ── Property clause ───────────────────────────────────────────────────────────

// PropertyClause describes a real property bequest, optionally verified against
// LINZ (Land Information New Zealand / Toitū Te Whenua) title data.
type PropertyClause struct {
	ID               string     `json:"id"`
	WillID           WillID     `json:"willId"`
	ClauseIndex      int        `json:"clauseIndex"`
	TitleReference   string     `json:"titleReference"`           // e.g. "NA1234/1"
	LINZVerifiedAt   *time.Time `json:"linzVerifiedAt,omitempty"` // nil = unverified
	LINZLandAreaM2   float64    `json:"linzLandAreaM2,omitempty"`
	LegalDescription string     `json:"legalDescription,omitempty"`
	BeneficiaryID    string     `json:"beneficiaryId"`
}

// ── Audit log ─────────────────────────────────────────────────────────────────

// AuditEvent is the event type recorded in the hash-chained audit log.
type AuditEvent string

const (
	AuditEventCreated          AuditEvent = "will_created"
	AuditEventCodicilCreated   AuditEvent = "codicil_created"
	AuditEventCodicilAppended  AuditEvent = "codicil_appended"
	AuditEventTestatorSigned   AuditEvent = "testator_signed"
	AuditEventWitnessSigned    AuditEvent = "witness_signed"
	AuditEventLocked           AuditEvent = "locked"
	AuditEventSuperseded       AuditEvent = "superseded"
	AuditEventDeathNotified    AuditEvent = "death_notified"
	AuditEventVaultUnlocked    AuditEvent = "vault_unlocked"
	AuditEventLivenessConfirmed AuditEvent = "liveness_confirmed"
	AuditEventLivenessReminder AuditEvent = "liveness_reminder_sent"
)

// AuditEntry is one entry in the hash-chained, tamper-evident audit log.
//
// The entry ID is derived as: SHA-256(previous_entry_id | created_at | event | will_id | actor_flt)
// This means that any alteration of a past entry invalidates the IDs of all
// subsequent entries, making tampering detectable.
type AuditEntry struct {
	ID              string     `json:"id"`              // SHA-256 fingerprint
	WillID          WillID     `json:"willId"`
	PreviousEntryID string     `json:"previousEntryId"` // empty string for first entry
	Event           AuditEvent `json:"event"`
	ActorFLT        string     `json:"actorFlt,omitempty"`
	PayloadHash     string     `json:"payloadHash,omitempty"` // optional hash of will content at this point
	CreatedAt       time.Time  `json:"createdAt"`
}
