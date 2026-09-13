-- +goose Up

CREATE TABLE jobs (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  workspace_id uuid REFERENCES workspaces (id) ON DELETE CASCADE,
  type text NOT NULL,
  status text NOT NULL DEFAULT 'queued',
  payload jsonb NOT NULL DEFAULT '{}'::jsonb,
  result jsonb,
  attempts integer NOT NULL DEFAULT 0,
  max_attempts integer NOT NULL DEFAULT 5,
  run_at timestamptz NOT NULL DEFAULT now(),
  started_at timestamptz,
  finished_at timestamptz,
  locked_by text,
  locked_at timestamptz,
  error text,
  created_by uuid REFERENCES users (id) ON DELETE SET NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX jobs_status_run_at_idx ON jobs (status, run_at);
CREATE INDEX jobs_workspace_id_idx ON jobs (workspace_id);
CREATE INDEX jobs_type_idx ON jobs (type);

CREATE TABLE email_outbox (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  workspace_id uuid REFERENCES workspaces (id) ON DELETE SET NULL,
  user_id uuid REFERENCES users (id) ON DELETE SET NULL,
  to_email text NOT NULL,
  template text NOT NULL,
  payload jsonb NOT NULL DEFAULT '{}'::jsonb,
  status text NOT NULL DEFAULT 'queued',
  attempts integer NOT NULL DEFAULT 0,
  last_error text,
  sent_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX email_outbox_status_created_idx ON email_outbox (status, created_at);

CREATE TABLE exports (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  requested_by uuid REFERENCES users (id) ON DELETE SET NULL,
  scope text NOT NULL,
  node_id uuid REFERENCES nodes (id) ON DELETE SET NULL,
  format text NOT NULL,
  status text NOT NULL DEFAULT 'queued',
  object_key text,
  error text,
  expires_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  finished_at timestamptz
);

CREATE INDEX exports_workspace_created_idx ON exports (workspace_id, created_at DESC);

CREATE TABLE imports (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  requested_by uuid REFERENCES users (id) ON DELETE SET NULL,
  source text NOT NULL,
  object_key text,
  status text NOT NULL DEFAULT 'queued',
  stats jsonb NOT NULL DEFAULT '{}'::jsonb,
  error text,
  created_at timestamptz NOT NULL DEFAULT now(),
  finished_at timestamptz
);

CREATE INDEX imports_workspace_created_idx ON imports (workspace_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS imports;
DROP TABLE IF EXISTS exports;
DROP TABLE IF EXISTS email_outbox;
DROP TABLE IF EXISTS jobs;
