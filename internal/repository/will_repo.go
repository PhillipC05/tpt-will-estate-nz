package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tpt-nz/tpt-will-estate-nz/internal/models"
)

var ErrWillNotFound = errors.New("will not found")

type WillRepo struct {
	pool *pgxpool.Pool
}

func NewWillRepo(pool *pgxpool.Pool) *WillRepo {
	return &WillRepo{pool: pool}
}

// CreateDraft persists a new will (or codicil) and its child rows inside a
// single transaction. If supersede is non-empty, those wills are marked
// superseded within the same transaction (atomic will-replacement).
func (r *WillRepo) CreateDraft(ctx context.Context, will models.Will, supersede []models.WillID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	// Convert nullable FK pointers to *string for pgx.
	var parentWillIDStr *string
	if will.ParentWillID != nil {
		s := string(*will.ParentWillID)
		parentWillIDStr = &s
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO will(
			id, status,
			testator_flt, testator_full_name, testator_assurance_level,
			vault_ciphertext, vault_nonce, vault_alg,
			audit_note,
			is_codicil, parent_will_id, vault_unlock_delay_days
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
	`,
		string(will.ID),
		string(will.Status),
		will.Testator.RealMeFLT,
		will.Testator.FullName,
		will.Testator.AssuranceLevel,
		will.VaultCiphertext,
		will.VaultNonce,
		will.VaultAlg,
		will.AuditNote,
		will.IsCodicil,
		parentWillIDStr,
		will.VaultUnlockDelayDays,
	)
	if err != nil {
		return err
	}

	for _, c := range will.Clauses {
		_, err = tx.Exec(ctx, `
			INSERT INTO will_clause(id, will_id, clause_index, clause_text) VALUES ($1,$2,$3,$4)
		`, c.ID, string(will.ID), c.Index, c.Text)
		if err != nil {
			return err
		}
	}
	for _, b := range will.Beneficiaries {
		_, err = tx.Exec(ctx, `
			INSERT INTO will_beneficiary(id, will_id, name, email, relation) VALUES ($1,$2,$3,$4,$5)
		`, b.ID, string(will.ID), b.Name, b.Email, b.Relation)
		if err != nil {
			return err
		}
	}
	for _, e := range will.Executors {
		var flt *string
		if e.FLT != "" {
			flt = &e.FLT
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO will_executor(id, will_id, name, email, flt) VALUES ($1,$2,$3,$4,$5)
		`, e.ID, string(will.ID), e.Name, e.Email, flt)
		if err != nil {
			return err
		}
	}

	// Supersede prior wills in the same transaction.
	newIDStr := string(will.ID)
	for _, sid := range supersede {
		_, err = tx.Exec(ctx, `
			UPDATE will SET status=$2, superseded_by=$3 WHERE id=$1
		`, string(sid), string(models.WillStatusSuperseded), newIDStr)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// GetByID returns the full will record including all child rows.
func (r *WillRepo) GetByID(ctx context.Context, id models.WillID) (models.Will, error) {
	var w models.Will
	row := r.pool.QueryRow(ctx, `
		SELECT
			id, status,
			testator_flt, testator_full_name, testator_assurance_level,
			testator_signed_at, testator_payload_hash,
			locked_at, unlocked_at,
			vault_ciphertext, vault_nonce, vault_alg,
			audit_note,
			is_codicil, parent_will_id, superseded_by,
			death_notified_at, vault_unlock_delay_days, unlock_after,
			last_confirmed_at, liveness_notified_at
		FROM will
		WHERE id = $1
	`, string(id))

	var (
		testatorSignedAt    *time.Time
		testatorPayloadHash *string
		lockedAt            *time.Time
		unlockedAt          *time.Time
		parentWillIDStr     *string
		supersededByStr     *string
		deathNotifiedAt     *time.Time
		unlockAfter         *time.Time
		lastConfirmedAt     *time.Time
		livenessNotifiedAt  *time.Time
	)

	if err := row.Scan(
		&w.ID,
		&w.Status,
		&w.Testator.RealMeFLT,
		&w.Testator.FullName,
		&w.Testator.AssuranceLevel,
		&testatorSignedAt,
		&testatorPayloadHash,
		&lockedAt,
		&unlockedAt,
		&w.VaultCiphertext,
		&w.VaultNonce,
		&w.VaultAlg,
		&w.AuditNote,
		&w.IsCodicil,
		&parentWillIDStr,
		&supersededByStr,
		&deathNotifiedAt,
		&w.VaultUnlockDelayDays,
		&unlockAfter,
		&lastConfirmedAt,
		&livenessNotifiedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Will{}, ErrWillNotFound
		}
		return models.Will{}, err
	}

	w.TestatorSignedAt = testatorSignedAt
	w.TestatorPayloadHash = testatorPayloadHash
	w.LockedAt = lockedAt
	w.UnlockedAt = unlockedAt
	w.DeathNotifiedAt = deathNotifiedAt
	w.UnlockAfter = unlockAfter
	w.LastConfirmedAt = lastConfirmedAt
	w.LivenessNotifiedAt = livenessNotifiedAt
	if parentWillIDStr != nil {
		pid := models.WillID(*parentWillIDStr)
		w.ParentWillID = &pid
	}
	if supersededByStr != nil {
		sid := models.WillID(*supersededByStr)
		w.SupersededBy = &sid
	}

	// Child rows
	w.Clauses = r.fetchClauses(ctx, id)
	w.Beneficiaries = r.fetchBeneficiaries(ctx, id)
	w.Executors = r.fetchExecutors(ctx, id)
	w.WitnessSignatures = r.fetchWitnessSignatures(ctx, id)

	return w, nil
}

func (r *WillRepo) fetchClauses(ctx context.Context, id models.WillID) []models.Clause {
	rows, err := r.pool.Query(ctx, `SELECT id, clause_index, clause_text FROM will_clause WHERE will_id=$1 ORDER BY clause_index`, string(id))
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []models.Clause
	for rows.Next() {
		var c models.Clause
		if err := rows.Scan(&c.ID, &c.Index, &c.Text); err != nil {
			return out
		}
		out = append(out, c)
	}
	return out
}

func (r *WillRepo) fetchBeneficiaries(ctx context.Context, id models.WillID) []models.Beneficiary {
	rows, err := r.pool.Query(ctx, `SELECT id, name, email, relation FROM will_beneficiary WHERE will_id=$1`, string(id))
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []models.Beneficiary
	for rows.Next() {
		var b models.Beneficiary
		if err := rows.Scan(&b.ID, &b.Name, &b.Email, &b.Relation); err != nil {
			return out
		}
		out = append(out, b)
	}
	return out
}

func (r *WillRepo) fetchExecutors(ctx context.Context, id models.WillID) []models.Executor {
	rows, err := r.pool.Query(ctx, `SELECT id, name, email, COALESCE(flt,'') FROM will_executor WHERE will_id=$1`, string(id))
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []models.Executor
	for rows.Next() {
		var e models.Executor
		if err := rows.Scan(&e.ID, &e.Name, &e.Email, &e.FLT); err != nil {
			return out
		}
		out = append(out, e)
	}
	return out
}

func (r *WillRepo) fetchWitnessSignatures(ctx context.Context, id models.WillID) []models.WitnessSignature {
	rows, err := r.pool.Query(ctx, `
		SELECT witness_flt, witness_full_name, witness_assurance_level, witness_signed_at, witness_payload_hash
		FROM will_witness_signature WHERE will_id=$1 ORDER BY witness_signed_at
	`, string(id))
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []models.WitnessSignature
	for rows.Next() {
		var ws models.WitnessSignature
		if err := rows.Scan(&ws.Witness.RealMeFLT, &ws.Witness.FullName, &ws.Witness.AssuranceLevel, &ws.SignedAt, &ws.PayloadHash); err != nil {
			return out
		}
		out = append(out, ws)
	}
	return out
}

// FindActiveByTestatorFLT returns IDs of wills for a testator that are still
// in draft or locked state (candidates for supersession on new will creation).
func (r *WillRepo) FindActiveByTestatorFLT(ctx context.Context, flt string) ([]models.WillID, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id FROM will
		WHERE testator_flt = $1
		  AND status IN ('draft','locked')
		  AND is_codicil = FALSE
	`, flt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []models.WillID
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, models.WillID(id))
	}
	return ids, rows.Err()
}

// UpdateStatus transitions a will to a new lifecycle state.
func (r *WillRepo) UpdateStatus(ctx context.Context, id models.WillID, status models.WillStatus, lockedAt, unlockedAt *time.Time) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE will SET status=$2, locked_at=$3, unlocked_at=$4 WHERE id=$1
	`, string(id), string(status), lockedAt, unlockedAt)
	return err
}

// NotifyDeath records the BDM death notification and sets the time-lock.
func (r *WillRepo) NotifyDeath(ctx context.Context, id models.WillID, notifiedAt, unlockAfter time.Time) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE will
		SET status=$2, death_notified_at=$3, unlock_after=$4
		WHERE id=$1
	`, string(id), string(models.WillStatusDeathNotified), notifiedAt, unlockAfter)
	return err
}

// RecordTestatorSignature persists the testator's payload hash and signature timestamp.
func (r *WillRepo) RecordTestatorSignature(ctx context.Context, id models.WillID, payloadHash string, signedAt time.Time) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE will SET testator_signed_at=$2, testator_payload_hash=$3 WHERE id=$1
	`, string(id), signedAt, payloadHash)
	return err
}

// RecordWitnessSignature inserts a witness signature record.
func (r *WillRepo) RecordWitnessSignature(ctx context.Context, id models.WillID, payloadHash string, signedAt time.Time, witness models.IdentityClaims) error {
	idStr := string(id) + ":witness:" + signedAt.UTC().Format("20060102T150405.999999999Z")
	_, err := r.pool.Exec(ctx, `
		INSERT INTO will_witness_signature(id, will_id, witness_flt, witness_full_name, witness_assurance_level, witness_signed_at, witness_payload_hash)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
	`, idStr, string(id), witness.RealMeFLT, witness.FullName, witness.AssuranceLevel, signedAt, payloadHash)
	return err
}

// RecordLivenessConfirmation sets the last_confirmed_at timestamp.
func (r *WillRepo) RecordLivenessConfirmation(ctx context.Context, id models.WillID, confirmedAt time.Time) error {
	_, err := r.pool.Exec(ctx, `UPDATE will SET last_confirmed_at=$2 WHERE id=$1`, string(id), confirmedAt)
	return err
}

// RecordLivenessReminder sets the liveness_notified_at timestamp.
func (r *WillRepo) RecordLivenessReminder(ctx context.Context, id models.WillID, notifiedAt time.Time) error {
	_, err := r.pool.Exec(ctx, `UPDATE will SET liveness_notified_at=$2 WHERE id=$1`, string(id), notifiedAt)
	return err
}

// SetVault updates the encrypted vault fields (e.g. after re-encryption).
func (r *WillRepo) SetVault(ctx context.Context, id models.WillID, ciphertext, nonce, alg string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE will SET vault_ciphertext=$2, vault_nonce=$3, vault_alg=$4 WHERE id=$1
	`, string(id), ciphertext, nonce, alg)
	return err
}
