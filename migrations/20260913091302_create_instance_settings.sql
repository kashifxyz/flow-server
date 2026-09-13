-- +goose Up

CREATE TABLE instance_settings (
  key text PRIMARY KEY,
  value jsonb NOT NULL DEFAULT '{}'::jsonb,
  updated_by uuid REFERENCES users (id) ON DELETE SET NULL,
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE instance_admins (
  user_id uuid PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
  created_by uuid REFERENCES users (id) ON DELETE SET NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS instance_admins;
DROP TABLE IF EXISTS instance_settings;
