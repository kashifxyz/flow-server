-- +goose Up

CREATE TABLE sessions (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  device_id uuid REFERENCES user_devices (id) ON DELETE SET NULL,
  token_hash text NOT NULL,
  csrf_hash text,
  expires_at timestamptz NOT NULL,
  idle_timeout_at timestamptz,
  last_seen_at timestamptz NOT NULL DEFAULT now(),
  revoked_at timestamptz,
  revoke_reason text,
  user_agent text,
  ip inet,
  country text,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX sessions_token_hash_uidx ON sessions (token_hash);
CREATE INDEX sessions_user_id_idx ON sessions (user_id);
CREATE INDEX sessions_expires_at_idx ON sessions (expires_at) WHERE revoked_at IS NULL;

CREATE TABLE auth_tokens (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  purpose text NOT NULL,
  token_hash text NOT NULL,
  expires_at timestamptz NOT NULL,
  consumed_at timestamptz,
  revoked_at timestamptz,
  request_ip inet,
  user_agent text,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX auth_tokens_token_hash_uidx ON auth_tokens (token_hash);
CREATE INDEX auth_tokens_user_id_purpose_idx ON auth_tokens (user_id, purpose);

CREATE TABLE login_attempts (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  user_id uuid REFERENCES users (id) ON DELETE SET NULL,
  email text NOT NULL,
  ip inet,
  user_agent text,
  success boolean NOT NULL DEFAULT false,
  failure_reason text,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX login_attempts_email_created_idx ON login_attempts (lower(email), created_at DESC);
CREATE INDEX login_attempts_ip_created_idx ON login_attempts (ip, created_at DESC);

CREATE TABLE auth_events (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  user_id uuid REFERENCES users (id) ON DELETE SET NULL,
  session_id uuid REFERENCES sessions (id) ON DELETE SET NULL,
  event_type text NOT NULL,
  ip inet,
  user_agent text,
  metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX auth_events_user_created_idx ON auth_events (user_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS auth_events;
DROP TABLE IF EXISTS login_attempts;
DROP TABLE IF EXISTS auth_tokens;
DROP TABLE IF EXISTS sessions;
