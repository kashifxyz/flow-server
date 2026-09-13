-- +goose Up

CREATE TABLE users (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  email text NOT NULL,
  password_hash text NOT NULL,
  password_algo text NOT NULL DEFAULT 'argon2id',
  password_params jsonb NOT NULL DEFAULT '{}'::jsonb,
  display_name text NOT NULL DEFAULT '',
  given_name text NOT NULL DEFAULT '',
  family_name text NOT NULL DEFAULT '',
  username text,
  locale text NOT NULL DEFAULT 'en',
  timezone text NOT NULL DEFAULT 'UTC',
  week_starts_on smallint NOT NULL DEFAULT 1,
  date_format text NOT NULL DEFAULT 'yyyy-mm-dd',
  time_format text NOT NULL DEFAULT '24h',
  theme text NOT NULL DEFAULT 'system',
  avatar_object_key text,
  bio text NOT NULL DEFAULT '',
  email_verified_at timestamptz,
  password_updated_at timestamptz,
  last_login_at timestamptz,
  last_login_ip inet,
  last_seen_at timestamptz,
  failed_login_count integer NOT NULL DEFAULT 0,
  locked_until timestamptz,
  mfa_required boolean NOT NULL DEFAULT false,
  mfa_enrolled_at timestamptz,
  tos_accepted_at timestamptz,
  privacy_accepted_at timestamptz,
  disabled_at timestamptz,
  disable_reason text,
  deleted_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX users_email_lower_uidx ON users (lower(email)) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX users_username_lower_uidx ON users (lower(username)) WHERE username IS NOT NULL AND deleted_at IS NULL;

CREATE TABLE user_preferences (
  user_id uuid PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
  editor jsonb NOT NULL DEFAULT '{}'::jsonb,
  notifications jsonb NOT NULL DEFAULT '{}'::jsonb,
  accessibility jsonb NOT NULL DEFAULT '{}'::jsonb,
  shortcuts jsonb NOT NULL DEFAULT '{}'::jsonb,
  extras jsonb NOT NULL DEFAULT '{}'::jsonb,
  updated_at timestamptz NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS user_preferences;
DROP TABLE IF EXISTS users;
