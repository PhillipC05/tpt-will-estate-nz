package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/tpt-nz/tpt-will-estate-nz/internal/models"
	"github.com/tpt-nz/tpt-will-estate-nz/internal/repository"
)

// AuditService records hash-chained, tamper-evident audit events for each will.
//
// Chain integrity: each entry's ID is derived as:
//
//	SHA-256(previous_entry_id | created_at.RFC3339Nano | event | will_id | actor_flt)
//
// Replacing or reordering any historical entry invalidates all subsequent IDs,
// making tampering immediately detectable by re-running the chain derivation.
type AuditService struct {
	repo *repository.AuditRepo
	log  *slog.Logger
}

func NewAuditService(repo *repository.AuditRepo, log *slog.Logger) *AuditService {
	return &AuditService{repo: repo, log: log}
}

// Record appends one event to the audit log for a will.
// payloadHash may be empty for events that don't change the will content.
// Errors are logged but not returned — audit failure must not abort estate operations.
func (s *AuditService) Record(ctx context.Context, willID models.WillID, event models.AuditEvent, actorFLT, payloadHash string) {
	if err := s.record(ctx, willID, event, actorFLT, payloadHash, time.Now().UTC()); err != nil {
		s.log.Error("audit record failed", "will_id", string(willID), "event", string(event), "err", err)
	}
}

func (s *AuditService) record(ctx context.Context, willID models.WillID, event models.AuditEvent, actorFLT, payloadHash string, at time.Time) error {
	prevID, err := s.repo.GetLatestEntryID(ctx, willID)
	if err != nil && !errors.Is(err, repository.ErrNoAuditEntries) {
		return err
	}

	entry := repository.AuditEntryForTime(willID, prevID, event, actorFLT, payloadHash, at)
	entry.ID = deriveEntryID(prevID, at, event, willID, actorFLT)

	return s.repo.Insert(ctx, entry)
}

// GetByWillID returns all audit entries for a will in chronological order.
func (s *AuditService) GetByWillID(ctx context.Context, willID models.WillID) ([]models.AuditEntry, error) {
	return s.repo.GetByWillID(ctx, willID)
}

// VerifyChain re-derives every entry ID and returns an error if the chain is broken.
func (s *AuditService) VerifyChain(ctx context.Context, willID models.WillID) error {
	return s.repo.VerifyChain(ctx, willID, func(prevID string, e models.AuditEntry) string {
		return deriveEntryID(prevID, e.CreatedAt, e.Event, e.WillID, e.ActorFLT)
	})
}

// deriveEntryID computes the chain-link hash for an audit entry.
func deriveEntryID(prevID string, at time.Time, event models.AuditEvent, willID models.WillID, actorFLT string) string {
	raw := strings.Join([]string{
		prevID,
		at.UTC().Format(time.RFC3339Nano),
		string(event),
		string(willID),
		actorFLT,
	}, "|")
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}
