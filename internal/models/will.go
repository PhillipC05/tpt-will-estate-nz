package models

import "time"

// WillID is a stable public identifier for a will (UUID-like string).
type WillID string

type WillStatus string

const (
	WillStatusDraft        WillStatus = "draft"
	WillStatusSigned       WillStatus = "signed"
	WillStatusLocked       WillStatus = "locked"
	WillStatusUnlockedDead WillStatus = "unlocked_at_death"
)

type IdentityClaims struct {
	RealMeFLT        string
	FullName        string
	AssuranceLevel  string
	IsVerified      bool
}

type Clause struct {
	ID      string `json:"id"`
	Text    string `json:"text"`
	Index   int    `json:"index"`
}

type Beneficiary struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Relation string `json:"relation"`
}

type Executor struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type WitnessSignature struct {
	Witness      IdentityClaims `json:"witness"`
	SignedAt     time.Time       `json:"signedAt"`
	PayloadHash  string          `json:"payloadHash"`
}

// Will is a digital will record.
type Will struct {
	ID        WillID      `json:"id"`
	Testator  IdentityClaims `json:"testator"`
	Clauses   []Clause   `json:"clauses"`
	Beneficiaries []Beneficiary `json:"beneficiaries"`
	Executors []Executor `json:"executors"`

	TestatorSignedAt  *time.Time         `json:"testatorSignedAt"`
	TestatorPayloadHash *string          `json:"testatorPayloadHash"`
	WitnessSignatures []WitnessSignature `json:"witnessSignatures"`

	Status    WillStatus `json:"status"`
	LockedAt  *time.Time `json:"lockedAt"`
	UnlockedAt *time.Time `json:"unlockedAt"`

	// Encrypted vault metadata (server does not store plaintext).
	VaultCiphertext string `json:"vaultCiphertext"`
	VaultNonce      string `json:"vaultNonce"`
	VaultAlg         string `json:"vaultAlg"`

	AuditNote string `json:"auditNote"`
}

