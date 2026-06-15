package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tpt-nz/tpt-will-estate-nz/internal/models"
)

// DigitalAssetRepo handles persistence for digital-asset clauses and property clauses.
type DigitalAssetRepo struct {
	pool *pgxpool.Pool
}

func NewDigitalAssetRepo(pool *pgxpool.Pool) *DigitalAssetRepo {
	return &DigitalAssetRepo{pool: pool}
}

// ── Digital asset clauses ────────────────────────────────────────────────────

func (r *DigitalAssetRepo) InsertDigitalAsset(ctx context.Context, c models.DigitalAssetClause) error {
	var beneficiaryID *string
	if c.BeneficiaryID != "" {
		beneficiaryID = &c.BeneficiaryID
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO will_digital_asset_clause(id, will_id, clause_index, asset_type, platform, identifier, instruction, beneficiary_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
	`, c.ID, string(c.WillID), c.ClauseIndex, string(c.AssetType), c.Platform, c.Identifier, c.Instruction, beneficiaryID)
	return err
}

func (r *DigitalAssetRepo) GetDigitalAssetsByWillID(ctx context.Context, willID models.WillID) ([]models.DigitalAssetClause, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, will_id, clause_index, asset_type, platform, COALESCE(identifier,''), instruction, COALESCE(beneficiary_id,'')
		FROM will_digital_asset_clause
		WHERE will_id = $1
		ORDER BY clause_index
	`, string(willID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.DigitalAssetClause
	for rows.Next() {
		var c models.DigitalAssetClause
		var willIDStr, assetType string
		if err := rows.Scan(&c.ID, &willIDStr, &c.ClauseIndex, &assetType, &c.Platform, &c.Identifier, &c.Instruction, &c.BeneficiaryID); err != nil {
			return nil, err
		}
		c.WillID = models.WillID(willIDStr)
		c.AssetType = models.DigitalAssetType(assetType)
		out = append(out, c)
	}
	return out, rows.Err()
}

// ── Property clauses ─────────────────────────────────────────────────────────

func (r *DigitalAssetRepo) InsertPropertyClause(ctx context.Context, c models.PropertyClause) error {
	var beneficiaryID *string
	if c.BeneficiaryID != "" {
		beneficiaryID = &c.BeneficiaryID
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO will_property_clause(id, will_id, clause_index, title_reference, linz_verified_at, linz_land_area_m2, legal_description, beneficiary_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
	`, c.ID, string(c.WillID), c.ClauseIndex, c.TitleReference, c.LINZVerifiedAt, nullableFloat(c.LINZLandAreaM2), c.LegalDescription, beneficiaryID)
	return err
}

func (r *DigitalAssetRepo) GetPropertyClausesByWillID(ctx context.Context, willID models.WillID) ([]models.PropertyClause, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, will_id, clause_index, title_reference, linz_verified_at, COALESCE(linz_land_area_m2,0), COALESCE(legal_description,''), COALESCE(beneficiary_id,'')
		FROM will_property_clause
		WHERE will_id = $1
		ORDER BY clause_index
	`, string(willID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.PropertyClause
	for rows.Next() {
		var c models.PropertyClause
		var willIDStr string
		var verifiedAt *time.Time
		if err := rows.Scan(&c.ID, &willIDStr, &c.ClauseIndex, &c.TitleReference, &verifiedAt, &c.LINZLandAreaM2, &c.LegalDescription, &c.BeneficiaryID); err != nil {
			return nil, err
		}
		c.WillID = models.WillID(willIDStr)
		c.LINZVerifiedAt = verifiedAt
		out = append(out, c)
	}
	return out, rows.Err()
}

// UpdatePropertyClauseLINZ stores LINZ verification data on an existing clause.
func (r *DigitalAssetRepo) UpdatePropertyClauseLINZ(ctx context.Context, clauseID string, areaM2 float64, legalDesc string, verifiedAt time.Time) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE will_property_clause
		SET linz_verified_at=$2, linz_land_area_m2=$3, legal_description=$4
		WHERE id=$1
	`, clauseID, verifiedAt, areaM2, legalDesc)
	return err
}

// nullableFloat returns nil for zero to avoid storing meaningless 0 values.
func nullableFloat(f float64) *float64 {
	if f == 0 {
		return nil
	}
	return &f
}
