package services

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/tpt-nz/tpt-will-estate-nz/internal/models"
	"github.com/tpt-nz/tpt-will-estate-nz/internal/repository"
)

var (
	ErrForbidden = errors.New("forbidden")
	ErrInvalidState = errors.New("invalid state")
)

type WillService struct {
	repo   *repository.WillRepo
	encSvc *EncryptionService
	log    *slog.Logger
}

func NewWillService(repo *repository.WillRepo, encSvc *EncryptionService, log *slog.Logger) *WillService {
	return &WillService{repo: repo, encSvc: encSvc, log: log}
}

type CreateWillRequest struct {
	Clauses        []models.Clause        `json:"clauses"`
	Beneficiaries  []models.Beneficiary  `json:"beneficiaries"`
	Executors      []models.Executor      `json:"executors"`
	VaultCiphertext string               `json:"vaultCiphertext"`
	VaultNonce      string               `json:"vaultNonce"`
	VaultAlg         string               `json:"vaultAlg"`
}

func (s *WillService) CreateDraft(ctx context.Context, testator models.IdentityClaims, req CreateWillRequest) (models.WillID, error) {
	if err := s.encSvc.ValidateVault(req.VaultCiphertext, req.VaultNonce, req.VaultAlg); err != nil {
		return "", err
	}

	// Deterministic scaffold ID. Replace with UUID generation once required.
	id := models.WillID("will_scaffold_" + time.Now().UTC().Format(time.RFC3339Nano))

	// Ensure child row IDs exist.
	clauses := make([]models.Clause, 0, len(req.Clauses))
	for _, c := range req.Clauses {
		if c.ID == "" {
			c.ID = string(id) + ":clause:" + time.Now().UTC().Format(time.RFC3339Nano)
		}
		clauses = append(clauses, c)
	}
	beneficiaries := make([]models.Beneficiary, 0, len(req.Beneficiaries))
	for _, b := range req.Beneficiaries {
		if b.ID == "" {
			b.ID = string(id) + ":beneficiary:" + time.Now().UTC().Format(time.RFC3339Nano)
		}
		beneficiaries = append(beneficiaries, b)
	}
	executors := make([]models.Executor, 0, len(req.Executors))
	for _, e := range req.Executors {
		if e.ID == "" {
			e.ID = string(id) + ":executor:" + time.Now().UTC().Format(time.RFC3339Nano)
		}
		executors = append(executors, e)
	}

	will := models.Will{
		ID:               id,
		Testator:         testator,
		Clauses:          clauses,
		Beneficiaries:    beneficiaries,
		Executors:        executors,
		Status:           models.WillStatusDraft,
		VaultCiphertext:  req.VaultCiphertext,
		VaultNonce:       req.VaultNonce,
		VaultAlg:          req.VaultAlg,
		AuditNote:        "created draft",
	}


	if err := s.repo.CreateDraft(ctx, will); err != nil {
		return "", err
	}
	return id, nil
}

func (s *WillService) SignTestator(ctx context.Context, id models.WillID, signer models.IdentityClaims, payloadHash string) error {
	// TODO: fetch will, validate state, authorization, then record signature.
	return ErrInvalidState
}

func (s *WillService) SignWitness(ctx context.Context, id models.WillID, witness models.IdentityClaims, payloadHash string) error {
	return ErrInvalidState
}

func (s *WillService) Lock(ctx context.Context, id models.WillID, lockedAt time.Time) error {
	return ErrInvalidState
}

func (s *WillService) UnlockAtDeath(ctx context.Context, id models.WillID, unlockedAt time.Time) error {
	return ErrInvalidState
}

