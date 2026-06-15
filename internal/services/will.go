package services

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/tpt-nz/tpt-will-estate-nz/internal/models"
	"github.com/tpt-nz/tpt-will-estate-nz/internal/repository"
)

var (
	ErrForbidden    = errors.New("forbidden")
	ErrInvalidState = errors.New("invalid state")
)

// WillService orchestrates the full will lifecycle.
type WillService struct {
	repo         *repository.WillRepo
	dalRepo      *repository.DigitalAssetRepo
	encSvc       *EncryptionService
	audit        *AuditService
	linzVerifier PropertyVerifier // nil = LINZ verification disabled
	log          *slog.Logger
}

func NewWillService(
	repo *repository.WillRepo,
	dalRepo *repository.DigitalAssetRepo,
	encSvc *EncryptionService,
	audit *AuditService,
	linzVerifier PropertyVerifier,
	log *slog.Logger,
) *WillService {
	return &WillService{repo: repo, dalRepo: dalRepo, encSvc: encSvc, audit: audit, linzVerifier: linzVerifier, log: log}
}

// CreateWillRequest is the payload for creating a draft will or codicil.
type CreateWillRequest struct {
	Clauses              []models.Clause      `json:"clauses"`
	Beneficiaries        []models.Beneficiary `json:"beneficiaries"`
	Executors            []models.Executor    `json:"executors"`
	VaultCiphertext      string               `json:"vaultCiphertext"`
	VaultNonce           string               `json:"vaultNonce"`
	VaultAlg             string               `json:"vaultAlg"`
	VaultUnlockDelayDays int                  `json:"vaultUnlockDelayDays"` // cooling-off days; 0 = immediate
}

// ── Core lifecycle ────────────────────────────────────────────────────────────

// CreateDraft creates a new will for the testator, automatically superseding any
// existing active (draft or locked) wills to comply with Wills Act 2007 s.17.
func (s *WillService) CreateDraft(ctx context.Context, testator models.IdentityClaims, req CreateWillRequest) (models.WillID, error) {
	if err := s.encSvc.ValidateVault(req.VaultCiphertext, req.VaultNonce, req.VaultAlg); err != nil {
		return "", err
	}

	id := newWillID()
	will := s.buildWill(id, testator, req, false, nil)

	// Feature 3: find existing active wills for this testator to supersede.
	supersede, err := s.repo.FindActiveByTestatorFLT(ctx, testator.RealMeFLT)
	if err != nil {
		return "", err
	}

	if err := s.repo.CreateDraft(ctx, will, supersede); err != nil {
		return "", err
	}

	s.audit.Record(ctx, id, models.AuditEventCreated, testator.RealMeFLT, "")
	for _, sid := range supersede {
		s.audit.Record(ctx, sid, models.AuditEventSuperseded, testator.RealMeFLT, string(id))
	}
	return id, nil
}

// GetByID returns the full will record.
func (s *WillService) GetByID(ctx context.Context, id models.WillID) (models.Will, error) {
	will, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return models.Will{}, err
	}
	will.DigitalAssetClauses, _ = s.dalRepo.GetDigitalAssetsByWillID(ctx, id)
	will.PropertyClauses, _ = s.dalRepo.GetPropertyClausesByWillID(ctx, id)
	return will, nil
}

// SignTestator records the testator's digital signature over the will payload hash.
func (s *WillService) SignTestator(ctx context.Context, id models.WillID, signer models.IdentityClaims, payloadHash string) error {
	will, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if will.Status != models.WillStatusDraft {
		return ErrInvalidState
	}
	if will.Testator.RealMeFLT != signer.RealMeFLT {
		return ErrForbidden
	}
	if err := s.repo.RecordTestatorSignature(ctx, id, payloadHash, time.Now().UTC()); err != nil {
		return err
	}
	s.audit.Record(ctx, id, models.AuditEventTestatorSigned, signer.RealMeFLT, payloadHash)
	return nil
}

// SignWitness records a witness signature. The testator must have signed first,
// and the witness cannot be the testator (Wills Act 2007 s.11).
func (s *WillService) SignWitness(ctx context.Context, id models.WillID, witness models.IdentityClaims, payloadHash string) error {
	will, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if will.Status != models.WillStatusDraft {
		return ErrInvalidState
	}
	if will.Testator.RealMeFLT == witness.RealMeFLT {
		return ErrForbidden
	}
	if will.TestatorSignedAt == nil {
		return ErrInvalidState
	}
	if err := s.repo.RecordWitnessSignature(ctx, id, payloadHash, time.Now().UTC(), witness); err != nil {
		return err
	}
	s.audit.Record(ctx, id, models.AuditEventWitnessSigned, witness.RealMeFLT, payloadHash)
	return nil
}

// Lock finalises the will after two-witness signing. The will cannot be modified
// after locking; executors and beneficiaries are notified only after death.
func (s *WillService) Lock(ctx context.Context, id models.WillID, lockedAt time.Time) error {
	will, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if will.Status != models.WillStatusDraft {
		return ErrInvalidState
	}
	if will.TestatorSignedAt == nil || len(will.WitnessSignatures) < 2 {
		return ErrInvalidState
	}
	if err := s.repo.UpdateStatus(ctx, id, models.WillStatusLocked, &lockedAt, nil); err != nil {
		return err
	}
	s.audit.Record(ctx, id, models.AuditEventLocked, will.Testator.RealMeFLT, "")
	return nil
}

// NotifyDeath records the BDM death notification and applies the time-lock vault.
// When delayDays == 0 the vault is unlocked immediately (backward-compatible).
// When delayDays > 0 the status transitions to death_notified; CompleteUnlock
// must be called (or triggered lazily by the executor access handler) once the
// cooling-off period has elapsed.
func (s *WillService) NotifyDeath(ctx context.Context, id models.WillID, delayDays int, notifiedAt time.Time) error {
	will, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if will.Status != models.WillStatusLocked {
		return ErrInvalidState
	}
	unlockAfter := notifiedAt.AddDate(0, 0, delayDays)
	if delayDays == 0 {
		if err := s.repo.UpdateStatus(ctx, id, models.WillStatusUnlockedDead, nil, &notifiedAt); err != nil {
			return err
		}
		s.audit.Record(ctx, id, models.AuditEventVaultUnlocked, "", "")
		return nil
	}
	if err := s.repo.NotifyDeath(ctx, id, notifiedAt, unlockAfter); err != nil {
		return err
	}
	s.audit.Record(ctx, id, models.AuditEventDeathNotified, "", "")
	return nil
}

// CompleteUnlock transitions a death-notified will to unlocked_at_death once
// the cooling-off period has elapsed. Called lazily by the executor access handler.
func (s *WillService) CompleteUnlock(ctx context.Context, id models.WillID, now time.Time) error {
	will, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if will.Status != models.WillStatusDeathNotified {
		return ErrInvalidState
	}
	if will.UnlockAfter != nil && now.Before(*will.UnlockAfter) {
		return ErrInvalidState // cooling-off not yet elapsed
	}
	if err := s.repo.UpdateStatus(ctx, id, models.WillStatusUnlockedDead, nil, &now); err != nil {
		return err
	}
	s.audit.Record(ctx, id, models.AuditEventVaultUnlocked, "", "")
	return nil
}

// ── Feature 4: Codicil support ───────────────────────────────────────────────

// CreateCodicil creates an amendment will (codicil) attached to an existing locked
// will. The codicil goes through the same signing lifecycle as the parent will but
// does NOT supersede the parent.
func (s *WillService) CreateCodicil(ctx context.Context, parentID models.WillID, testator models.IdentityClaims, req CreateWillRequest) (models.WillID, error) {
	parent, err := s.repo.GetByID(ctx, parentID)
	if err != nil {
		return "", err
	}
	if parent.Status != models.WillStatusLocked {
		return "", ErrInvalidState // codicils attach to locked wills only
	}
	if parent.Testator.RealMeFLT != testator.RealMeFLT {
		return "", ErrForbidden
	}
	if err := s.encSvc.ValidateVault(req.VaultCiphertext, req.VaultNonce, req.VaultAlg); err != nil {
		return "", err
	}

	id := newWillID()
	will := s.buildWill(id, testator, req, true, &parentID)
	will.AuditNote = "codicil to " + string(parentID)

	if err := s.repo.CreateDraft(ctx, will, nil); err != nil { // no supersession for codicils
		return "", err
	}
	s.audit.Record(ctx, id, models.AuditEventCodicilCreated, testator.RealMeFLT, "")
	s.audit.Record(ctx, parentID, models.AuditEventCodicilAppended, testator.RealMeFLT, string(id))
	return id, nil
}

// ── Feature 2: Testator liveness check ───────────────────────────────────────

// ConfirmLiveness records that the testator has confirmed the will is still current.
// Resets the liveness reminder clock.
func (s *WillService) ConfirmLiveness(ctx context.Context, id models.WillID, testator models.IdentityClaims) error {
	will, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if will.Testator.RealMeFLT != testator.RealMeFLT {
		return ErrForbidden
	}
	if will.Status == models.WillStatusSuperseded || will.Status == models.WillStatusUnlockedDead {
		return ErrInvalidState
	}
	if err := s.repo.RecordLivenessConfirmation(ctx, id, time.Now().UTC()); err != nil {
		return err
	}
	s.audit.Record(ctx, id, models.AuditEventLivenessConfirmed, testator.RealMeFLT, "")
	return nil
}

// RecordLivenessReminder marks that a liveness reminder was dispatched.
// Call this after sending the reminder email so the timestamp is accurate.
func (s *WillService) RecordLivenessReminder(ctx context.Context, id models.WillID) error {
	if err := s.repo.RecordLivenessReminder(ctx, id, time.Now().UTC()); err != nil {
		return err
	}
	s.audit.Record(ctx, id, models.AuditEventLivenessReminder, "", "")
	return nil
}

// ── Feature 7: Digital asset clauses ─────────────────────────────────────────

// AddDigitalAssetClause adds a digital-asset bequest clause to a draft will.
// SECURITY: the Instruction field must describe HOW to access the asset (e.g.
// "recovery seed is in the fireproof safe"), never the credential itself.
func (s *WillService) AddDigitalAssetClause(ctx context.Context, willID models.WillID, testator models.IdentityClaims, clause models.DigitalAssetClause) (models.DigitalAssetClause, error) {
	will, err := s.repo.GetByID(ctx, willID)
	if err != nil {
		return models.DigitalAssetClause{}, err
	}
	if will.Status != models.WillStatusDraft {
		return models.DigitalAssetClause{}, ErrInvalidState
	}
	if will.Testator.RealMeFLT != testator.RealMeFLT {
		return models.DigitalAssetClause{}, ErrForbidden
	}
	clause.ID = newChildID(string(willID), "da")
	clause.WillID = willID
	return clause, s.dalRepo.InsertDigitalAsset(ctx, clause)
}

// ── Feature 8: LINZ property cross-reference ──────────────────────────────────

// AddPropertyClause adds a real-property bequest, optionally verified against LINZ.
// If LINZ_API_KEY is configured, the title reference is validated against the NZ
// Parcels dataset and legal description / area are stored alongside the clause.
func (s *WillService) AddPropertyClause(ctx context.Context, willID models.WillID, testator models.IdentityClaims, titleRef, beneficiaryID string) (models.PropertyClause, error) {
	will, err := s.repo.GetByID(ctx, willID)
	if err != nil {
		return models.PropertyClause{}, err
	}
	if will.Status != models.WillStatusDraft {
		return models.PropertyClause{}, ErrInvalidState
	}
	if will.Testator.RealMeFLT != testator.RealMeFLT {
		return models.PropertyClause{}, ErrForbidden
	}

	clause := models.PropertyClause{
		ID:             newChildID(string(willID), "prop"),
		WillID:         willID,
		TitleReference: titleRef,
		BeneficiaryID:  beneficiaryID,
	}

	if s.linzVerifier != nil {
		data, verifyErr := s.linzVerifier.VerifyTitle(ctx, titleRef)
		if verifyErr != nil {
			s.log.Warn("LINZ title verification failed — storing clause unverified", "err", verifyErr, "title_ref", titleRef)
		} else {
			now := time.Now().UTC()
			clause.LINZVerifiedAt = &now
			clause.LINZLandAreaM2 = data.LINZLandAreaM2
			clause.LegalDescription = data.LegalDescription
		}
	}

	return clause, s.dalRepo.InsertPropertyClause(ctx, clause)
}

// ── Internal helpers ──────────────────────────────────────────────────────────

func (s *WillService) buildWill(id models.WillID, testator models.IdentityClaims, req CreateWillRequest, isCodicil bool, parentWillID *models.WillID) models.Will {
	clauses := make([]models.Clause, len(req.Clauses))
	for i, c := range req.Clauses {
		if c.ID == "" {
			c.ID = newChildID(string(id), "clause")
		}
		clauses[i] = c
	}
	beneficiaries := make([]models.Beneficiary, len(req.Beneficiaries))
	for i, b := range req.Beneficiaries {
		if b.ID == "" {
			b.ID = newChildID(string(id), "bene")
		}
		beneficiaries[i] = b
	}
	executors := make([]models.Executor, len(req.Executors))
	for i, e := range req.Executors {
		if e.ID == "" {
			e.ID = newChildID(string(id), "exec")
		}
		executors[i] = e
	}
	return models.Will{
		ID:                   id,
		Testator:             testator,
		Clauses:              clauses,
		Beneficiaries:        beneficiaries,
		Executors:            executors,
		Status:               models.WillStatusDraft,
		IsCodicil:            isCodicil,
		ParentWillID:         parentWillID,
		VaultCiphertext:      req.VaultCiphertext,
		VaultNonce:           req.VaultNonce,
		VaultAlg:             req.VaultAlg,
		VaultUnlockDelayDays: req.VaultUnlockDelayDays,
		AuditNote:            "created draft",
	}
}

func newWillID() models.WillID {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant bits
	return models.WillID(fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:]))
}

func newChildID(willID, kind string) string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s:%s:%x", willID, kind, b)
}
