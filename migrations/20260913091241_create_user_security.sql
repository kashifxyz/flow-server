-- +goose Up

CREATE TABLE password_history (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  password_hash text NOT NULL,
  password_algo text NOT NULL DEFAULT 'argon2id',
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX password_history_user_id_idx ON password_history (user_id, created_at DESC);

CREATE TABLE mfa_factors (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  type text NOT NULL,
  name text NOT NULL DEFAULT '',
  secret_ciphertext bytea,
  secret_nonce bytea,
  digits smallint NOT NULL DEFAULT 6,
  period_seconds smallint NOT NULL DEFAULT 30,
  verified_at timestamptz,
  last_used_at timestamptz,
  disabled_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX mfa_factors_user_id_idx ON mfa_factors (user_id);

CREATE TABLE mfa_recovery_codes (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  code_hash text NOT NULL,
  used_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX mfa_recovery_codes_code_hash_uidx ON mfa_recovery_codes (code_hash);
CREATE INDEX mfa_recovery_codes_user_id_idx ON mfa_recovery_codes (user_id);

CREATE TABLE user_devices (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  device_label text NOT NULL DEFAULT '',
  fingerprint_hash text,
  platform text,
  user_agent text,
  last_ip inet,
  last_seen_at timestamptz,
  trusted_at timestamptz,
  revoked_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX user_devices_user_id_idx ON user_devices (user_id);
CREATE UNIQUE INDEX user_devices_fingerprint_uidx ON user_devices (user_id, fingerprint_hash) WHERE fingerprint_hash IS NOT NULL AND revoked_at IS NULL;

CREATE TABLE api_tokens (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  workspace_id uuid,
  name text NOT NULL,
  token_prefix text NOT NULL,
  token_hash text NOT NULL,
  scopes jsonb NOT NULL DEFAULT '[]'::jsonb,
  last_used_at timestamptz,
  last_used_ip inet,
  expires_at timestamptz,
  revoked_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX api_tokens_token_hash_uidx ON api_tokens (token_hash);
CREATE INDEX api_tokens_user_id_idx ON api_tokens (user_id);

CREATE TABLE webauthn_credentials (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  device_id uuid REFERENCES user_devices (id) ON DELETE SET NULL,
  credential_id bytea NOT NULL,
  public_key bytea NOT NULL,
  attestation_type text,
  aaguid bytea,
  sign_count bigint NOT NULL DEFAULT 0,
  transports text[] NOT NULL DEFAULT '{}'::text[],
  backup_eligible boolean NOT NULL DEFAULT false,
  backed_up boolean NOT NULL DEFAULT false,
  name text NOT NULL DEFAULT '',
  last_used_at timestamptz,
  revoked_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX webauthn_credentials_credential_id_uidx ON webauthn_credentials (credential_id);
CREATE INDEX webauthn_credentials_user_id_idx ON webauthn_credentials (user_id);

-- +goose Down
DROP TABLE IF EXISTS webauthn_credentials;
DROP TABLE IF EXISTS api_tokens;
DROP TABLE IF EXISTS user_devices;
DROP TABLE IF EXISTS mfa_recovery_codes;
DROP TABLE IF EXISTS mfa_factors;
DROP TABLE IF EXISTS password_history;
