package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tpt-nz/tpt-will-estate-nz/internal/models"
)

var (
	ErrWillNotFound = errors.New("will not found")
)

type WillRepo struct {
	pool *pgxpool.Pool
}

func NewWillRepo(pool *pgxpool.Pool) *WillRepo {
	return &WillRepo{pool: pool}
}

func (r *WillRepo) CreateDraft(ctx context.Context, will models.Will) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO will(
			id, status,
			testator_flt, testator_full_name, testator_assurance_level,
			vault_ciphertext, vault_nonce, vault_alg,
			audit_note
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
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
	)
	if err != nil {
		return err
	}

	// Clauses/beneficiaries/executors inserted as separate tables.
	for _, c := range will.Clauses {
		_, err = r.pool.Exec(ctx, `
			INSERT INTO will_clause(id, will_id, clause_index, clause_text)
			VALUES ($1,$2,$3,$4)
		`, c.ID, string(will.ID), c.Index, c.Text)
		if err != nil {
			return err
		}
	}
	for _, b := range will.Beneficiaries {
		_, err = r.pool.Exec(ctx, `
			INSERT INTO will_beneficiary(id, will_id, name, email, relation)
			VALUES ($1,$2,$3,$4,$5)
		`, b.ID, string(will.ID), b.Name, b.Email, b.Relation)
		if err != nil {
			return err
		}
	}
	for _, e := range will.Executors {
		_, err = r.pool.Exec(ctx, `
			INSERT INTO will_executor(id, will_id, name, email)
			VALUES ($1,$2,$3,$4)
		`, e.ID, string(will.ID), e.Name, e.Email)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *WillRepo) GetByID(ctx context.Context, id models.WillID) (models.Will, error) {
	var w models.Will
	row := r.pool.QueryRow(ctx, `
		SELECT
			id, status,
			testator_flt, testator_full_name, testator_assurance_level,
			testator_signed_at, testator_payload_hash,
			locked_at, unlocked_at,
			vault_ciphertext, vault_nonce, vault_alg,
			audit_note
		FROM will
		WHERE id = $1
	`, string(id))

	var (
		testatorSignedAt *time.Time
		testatorPayloadHash *string
		lockedAt *time.Time
		unlockedAt *time.Time
	)

	if err := row.Scan(
		&w.ID,
		&w.Status,
		&w.Testator.RealMeFLT,
		&w.Testator.FullName,
		&w.Testator.AssuranceLevel,
		testatorSignedAt,
		testatorPayloadHash,
		lockedAt,
		unlockedAt,
		&w.VaultCiphertext,
		&w.VaultNonce,
		&w.VaultAlg,
		&w.AuditNote,
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

	// Fetch clauses/beneficiaries/executors/witnesses
	w.Clauses = []models.Clause{}
	rows, err := r.pool.Query(ctx, `SELECT id, clause_index, clause_text FROM will_clause WHERE will_id=$1 ORDER BY clause_index`, string(id))
	if err != nil {
		return models.Will{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var c models.Clause
		if err := rows.Scan(&c.ID, &c.Index, &c.Text); err != nil {
			return models.Will{}, err
		}
		w.Clauses = append(w.Clauses, c)
	}

	w.Beneficiaries = []models.Beneficiary{}
	rows, err = r.pool.Query(ctx, `SELECT id, name, email, relation FROM will_beneficiary WHERE will_id=$1`, string(id))
	if err != nil {
		return models.Will{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var b models.Beneficiary
		if err := rows.Scan(&b.ID, &b.Name, &b.Email, &b.Relation); err != nil {
			return models.Will{}, err
		}
		w.Beneficiaries = append(w.Beneficiaries, b)
	}

	w.Executors = []models.Executor{}
	rows, err = r.pool.Query(ctx, `SELECT id, name, email FROM will_executor WHERE will_id=$1`, string(id))
	if err != nil {
		return models.Will{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var e models.Executor
		if err := rows.Scan(&e.ID, &e.Name, &e.Email); err != nil {
			return models.Will{}, err
		}
		w.Executors = append(w.Executors, e)
	}

	w.WitnessSignatures = []models.WitnessSignature{}
	rows, err = r.pool.Query(ctx, `
		SELECT
			witness_flt, witness_full_name, witness_assurance_level,
			witness_signed_at, witness_payload_hash
		FROM will_witness_signature
		WHERE will_id=$1
		ORDER BY witness_signed_at
	`, string(id))
	if err != nil {
		return models.Will{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var ws models.WitnessSignature
		var assurance string
		if err := rows.Scan(&ws.Witness.RealMeFLT, &ws.Witness.FullName, &assurance, &ws.SignedAt, &ws.PayloadHash); err != nil {
			return models.Will{}, err
		}
		ws.Witness.AssuranceLevel = assurance
		w.WitnessSignatures = append(w.WitnessSignatures, ws)
	}

	return w, nil
}

func (r *WillRepo) UpdateStatus(ctx context.Context, id models.WillID, status models.WillStatus, lockedAt, unlockedAt *time.Time) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE will
		SET status=$2,
			locked_at=$3,
			unlocked_at=$4
		WHERE id=$1
	`, string(id), string(status), lockedAt, unlockedAt)
	return err
}

func (r *WillRepo) RecordTestatorSignature(ctx context.Context, id models.WillID, payloadHash string, signedAt time.Time) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE will
		SET testator_signed_at=$2,
			testator_payload_hash=$3
		WHERE id=$1
	`, string(id), signedAt, payloadHash)
	return err
}

func (r *WillRepo) RecordWitnessSignature(ctx context.Context, id models.WillID, payloadHash string, signedAt time.Time, witness models.IdentityClaims) error {
	idStr := string(id) + ":witness:" + signedAt.UTC().Format(time.RFC3339Nano)
	_, err := r.pool.Exec(ctx, `
		INSERT INTO will_witness_signature(id, will_id, witness_flt, witness_full_name, witness_assurance_level, witness_signed_at, witness_payload_hash)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
	`, idStr, string(id), witness.RealMeFLT, witness.FullName, witness.AssuranceLevel, signedAt, payloadHash)
	return err
}

func (r *WillRepo) SetVault(ctx context.Context, id models.WillID, ciphertext, nonce, alg string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE will
		SET vault_ciphertext=$2,
			vault_nonce=$3,
			vault_alg=$4
		WHERE id=$1
	`, string(id), ciphertext, nonce, alg)
	return err
}


