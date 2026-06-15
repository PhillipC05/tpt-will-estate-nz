package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tpt-nz/tpt-will-estate-nz/internal/models"
)

// ErrNoAuditEntries is returned when a will has no audit log entries yet.
var ErrNoAuditEntries = errors.New("no audit entries")

// AuditRepo persists the hash-chained tamper-evident audit log.
type AuditRepo struct {
	pool *pgxpool.Pool
}

func NewAuditRepo(pool *pgxpool.Pool) *AuditRepo {
	return &AuditRepo{pool: pool}
}

// Insert appends one entry to the audit log.
func (r *AuditRepo) Insert(ctx context.Context, e models.AuditEntry) error {
	var prevID *string
	if e.PreviousEntryID != "" {
		prevID = &e.PreviousEntryID
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO will_audit_log(id, will_id, previous_entry_id, event, actor_flt, payload_hash, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
	`, e.ID, string(e.WillID), prevID, string(e.Event), e.ActorFLT, e.PayloadHash, e.CreatedAt)
	return err
}

// GetLatestEntryID returns the ID of the most recent audit entry for a will.
// Returns ErrNoAuditEntries when no entries exist yet (start of chain).
func (r *AuditRepo) GetLatestEntryID(ctx context.Context, willID models.WillID) (string, error) {
	var id string
	err := r.pool.QueryRow(ctx, `
		SELECT id FROM will_audit_log
		WHERE will_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`, string(willID)).Scan(&id)
	if err != nil {
		if isNoRows(err) {
			return "", ErrNoAuditEntries
		}
		return "", err
	}
	return id, nil
}

// GetByWillID returns all audit entries for a will in chronological order.
func (r *AuditRepo) GetByWillID(ctx context.Context, willID models.WillID) ([]models.AuditEntry, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, will_id, COALESCE(previous_entry_id,''), event, COALESCE(actor_flt,''), COALESCE(payload_hash,''), created_at
		FROM will_audit_log
		WHERE will_id = $1
		ORDER BY created_at ASC
	`, string(willID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []models.AuditEntry
	for rows.Next() {
		var e models.AuditEntry
		var willIDStr string
		var event string
		if err := rows.Scan(&e.ID, &willIDStr, &e.PreviousEntryID, &event, &e.ActorFLT, &e.PayloadHash, &e.CreatedAt); err != nil {
			return nil, err
		}
		e.WillID = models.WillID(willIDStr)
		e.Event = models.AuditEvent(event)
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// VerifyChain re-derives every entry ID and confirms the chain is unbroken.
// Returns an error at the first tampered or missing entry.
func (r *AuditRepo) VerifyChain(ctx context.Context, willID models.WillID, deriveID func(prevID string, e models.AuditEntry) string) error {
	entries, err := r.GetByWillID(ctx, willID)
	if err != nil {
		return err
	}
	for _, e := range entries {
		expected := deriveID(e.PreviousEntryID, e)
		if expected != e.ID {
			return errors.New("audit chain broken at entry " + e.ID)
		}
	}
	return nil
}

// scanTimestamp is a helper for audit repo; isNoRows lives in will_repo so we
// use a local copy to avoid import cycles between repo files.
func isNoRows(err error) bool {
	return err != nil && err.Error() == "no rows in result set"
}

// AuditEntryForTime is a helper used by the audit service to build an entry
// with an explicit timestamp (needed for deterministic ID derivation in tests).
func AuditEntryForTime(willID models.WillID, prevID string, event models.AuditEvent, actorFLT, payloadHash string, at time.Time) models.AuditEntry {
	return models.AuditEntry{
		WillID:          willID,
		PreviousEntryID: prevID,
		Event:           event,
		ActorFLT:        actorFLT,
		PayloadHash:     payloadHash,
		CreatedAt:       at,
	}
}
