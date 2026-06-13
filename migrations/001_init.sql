SET search_path TO will_estate;

-- app-will-estate 001_init.sql

-- Minimal schema scaffold for digital wills.
-- NOTE: This is a scaffold; field coverage should be expanded once encryption/envelope design is finalized.

CREATE TABLE IF NOT EXISTS will (
  id TEXT PRIMARY KEY,
  status TEXT NOT NULL,

  testator_flt TEXT NOT NULL,
  testator_full_name TEXT NOT NULL,
  testator_assurance_level TEXT NOT NULL,

  testator_signed_at TIMESTAMPTZ NULL,
  testator_payload_hash TEXT NULL,

  locked_at TIMESTAMPTZ NULL,
  unlocked_at TIMESTAMPTZ NULL,

  vault_ciphertext TEXT NOT NULL,
  vault_nonce TEXT NOT NULL,
  vault_alg TEXT NOT NULL,

  audit_note TEXT NULL,

  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS will_clause (
  id TEXT PRIMARY KEY,
  will_id TEXT NOT NULL REFERENCES will(id) ON DELETE CASCADE,
  clause_index INT NOT NULL,
  clause_text TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS will_beneficiary (
  id TEXT PRIMARY KEY,
  will_id TEXT NOT NULL REFERENCES will(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  email TEXT NULL,
  relation TEXT NULL
);

CREATE TABLE IF NOT EXISTS will_executor (
  id TEXT PRIMARY KEY,
  will_id TEXT NOT NULL REFERENCES will(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  email TEXT NULL
);

CREATE TABLE IF NOT EXISTS will_witness_signature (
  id TEXT PRIMARY KEY,
  will_id TEXT NOT NULL REFERENCES will(id) ON DELETE CASCADE,
  witness_flt TEXT NOT NULL,
  witness_full_name TEXT NOT NULL,
  witness_assurance_level TEXT NOT NULL,
  witness_signed_at TIMESTAMPTZ NOT NULL,
  witness_payload_hash TEXT NOT NULL
);

-- Basic trigger to update updated_at.
CREATE OR REPLACE FUNCTION set_updated_at() RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS will_set_updated_at ON will;
CREATE TRIGGER will_set_updated_at
BEFORE UPDATE ON will
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

