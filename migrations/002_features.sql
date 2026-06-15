-- Migration 002: feature additions
-- Applies on top of 001_init.sql.
-- Run with: atlas schema apply --dir "file://migrations" --url "$DATABASE_URL"

-- ── Feature 1: Executor RealMe FLT verification ───────────────────────────────
-- Store the executor's RealMe FLT at will-creation time so their identity can
-- be verified at executor-access time after death notification.
ALTER TABLE will_executor ADD COLUMN IF NOT EXISTS flt TEXT;

-- ── Feature 2: Testator liveness check ───────────────────────────────────────
-- Track when the testator last confirmed the will is still current, and when
-- a liveness reminder email was last sent.
ALTER TABLE will ADD COLUMN IF NOT EXISTS last_confirmed_at      TIMESTAMPTZ;
ALTER TABLE will ADD COLUMN IF NOT EXISTS liveness_notified_at   TIMESTAMPTZ;

-- ── Feature 3: Will supersession ─────────────────────────────────────────────
-- When a testator creates a new will, prior active wills are automatically
-- superseded (Wills Act 2007 s.17). Records which newer will superseded this one.
ALTER TABLE will ADD COLUMN IF NOT EXISTS superseded_by TEXT REFERENCES will(id);

-- ── Feature 4: Codicil support ───────────────────────────────────────────────
-- A codicil is an amendment to an existing locked will. It goes through the same
-- signing lifecycle as a will but does not supersede the parent.
ALTER TABLE will ADD COLUMN IF NOT EXISTS parent_will_id TEXT REFERENCES will(id);
ALTER TABLE will ADD COLUMN IF NOT EXISTS is_codicil BOOLEAN NOT NULL DEFAULT FALSE;

-- ── Feature 5: Time-lock vault ───────────────────────────────────────────────
-- After a BDM death notification, the vault remains locked for a configurable
-- cooling-off period before executors can access it.
ALTER TABLE will ADD COLUMN IF NOT EXISTS death_notified_at      TIMESTAMPTZ;
ALTER TABLE will ADD COLUMN IF NOT EXISTS vault_unlock_delay_days INT NOT NULL DEFAULT 0;
ALTER TABLE will ADD COLUMN IF NOT EXISTS unlock_after            TIMESTAMPTZ;

-- ── Feature 6: Hash-chained audit trail ──────────────────────────────────────
-- Tamper-evident log of every state transition.
-- Entry ID = SHA-256(previous_entry_id || created_at || event || will_id || actor_flt)
-- Tampering with any past entry invalidates all subsequent entry IDs.
CREATE TABLE IF NOT EXISTS will_audit_log (
    id               TEXT        PRIMARY KEY,
    will_id          TEXT        NOT NULL REFERENCES will(id),
    previous_entry_id TEXT       REFERENCES will_audit_log(id),
    event            TEXT        NOT NULL,
    actor_flt        TEXT,
    payload_hash     TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS will_audit_log_will_id_idx ON will_audit_log(will_id, created_at);

-- ── Feature 7: Digital asset clauses ─────────────────────────────────────────
-- Structured records for digital assets (crypto wallets, social media, email).
-- NEVER stores passwords, seed phrases, or private keys — only access instructions.
CREATE TABLE IF NOT EXISTS will_digital_asset_clause (
    id             TEXT    PRIMARY KEY,
    will_id        TEXT    NOT NULL REFERENCES will(id),
    clause_index   INT     NOT NULL DEFAULT 0,
    asset_type     TEXT    NOT NULL,  -- crypto_wallet | social_media | email | password_manager | cloud_storage | other
    platform       TEXT    NOT NULL,  -- "Coinbase", "Gmail", etc.
    identifier     TEXT,              -- wallet address or username (not credentials)
    instruction    TEXT    NOT NULL,  -- how the executor should access or transfer
    beneficiary_id TEXT    REFERENCES will_beneficiary(id)
);

-- ── Feature 8: LINZ property cross-reference ──────────────────────────────────
-- Real property clauses verified against LINZ (Land Information New Zealand)
-- land title data to confirm ownership at the time the will was created.
CREATE TABLE IF NOT EXISTS will_property_clause (
    id                TEXT        PRIMARY KEY,
    will_id           TEXT        NOT NULL REFERENCES will(id),
    clause_index      INT         NOT NULL DEFAULT 0,
    title_reference   TEXT        NOT NULL,   -- e.g. "NA1234/1"
    linz_verified_at  TIMESTAMPTZ,            -- NULL = unverified
    linz_land_area_m2 NUMERIC,
    legal_description TEXT,
    beneficiary_id    TEXT        REFERENCES will_beneficiary(id)
);
