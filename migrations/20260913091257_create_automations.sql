-- +goose Up

CREATE TABLE automations (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  created_by uuid REFERENCES users (id) ON DELETE SET NULL,
  updated_by uuid REFERENCES users (id) ON DELETE SET NULL,
  name text NOT NULL,
  description text NOT NULL DEFAULT '',
  enabled boolean NOT NULL DEFAULT false,
  trigger_type text NOT NULL DEFAULT 'record.updated',
  trigger jsonb NOT NULL DEFAULT '{}'::jsonb,
  conditions jsonb NOT NULL DEFAULT '[]'::jsonb,
  last_run_at timestamptz,
  last_status text,
  last_error text,
  run_count bigint NOT NULL DEFAULT 0,
  version integer NOT NULL DEFAULT 1,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz
);

CREATE INDEX automations_workspace_id_idx ON automations (workspace_id) WHERE deleted_at IS NULL;
CREATE INDEX automations_trigger_type_idx ON automations (workspace_id, trigger_type) WHERE enabled AND deleted_at IS NULL;

CREATE TABLE automation_versions (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  automation_id uuid NOT NULL REFERENCES automations (id) ON DELETE CASCADE,
  version integer NOT NULL,
  trigger jsonb NOT NULL DEFAULT '{}'::jsonb,
  conditions jsonb NOT NULL DEFAULT '[]'::jsonb,
  created_by uuid REFERENCES users (id) ON DELETE SET NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (automation_id, version)
);

CREATE TABLE automation_steps (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  automation_id uuid NOT NULL REFERENCES automations (id) ON DELETE CASCADE,
  version integer NOT NULL DEFAULT 1,
  rank integer NOT NULL,
  type text NOT NULL,
  config jsonb NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX automation_steps_automation_rank_idx ON automation_steps (automation_id, version, rank);

CREATE TABLE automation_schedules (
  automation_id uuid PRIMARY KEY REFERENCES automations (id) ON DELETE CASCADE,
  cron_expr text NOT NULL,
  timezone text NOT NULL DEFAULT 'UTC',
  next_run_at timestamptz,
  last_run_at timestamptz
);

CREATE INDEX automation_schedules_next_run_idx ON automation_schedules (next_run_at);

CREATE TABLE automation_watches (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  automation_id uuid NOT NULL REFERENCES automations (id) ON DELETE CASCADE,
  node_id uuid REFERENCES nodes (id) ON DELETE CASCADE,
  database_id uuid REFERENCES databases (id) ON DELETE CASCADE,
  field_id uuid REFERENCES fields (id) ON DELETE CASCADE
);

CREATE INDEX automation_watches_automation_id_idx ON automation_watches (automation_id);

CREATE TABLE automation_runs (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  automation_id uuid NOT NULL REFERENCES automations (id) ON DELETE CASCADE,
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  version integer,
  status text NOT NULL DEFAULT 'queued',
  trigger_payload jsonb NOT NULL DEFAULT '{}'::jsonb,
  result jsonb,
  error text,
  started_at timestamptz,
  finished_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX automation_runs_automation_created_idx ON automation_runs (automation_id, created_at DESC);
CREATE INDEX automation_runs_status_idx ON automation_runs (status);

CREATE TABLE automation_step_runs (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  run_id uuid NOT NULL REFERENCES automation_runs (id) ON DELETE CASCADE,
  step_id uuid REFERENCES automation_steps (id) ON DELETE SET NULL,
  status text NOT NULL DEFAULT 'queued',
  input jsonb,
  output jsonb,
  error text,
  started_at timestamptz,
  finished_at timestamptz
);

CREATE INDEX automation_step_runs_run_id_idx ON automation_step_runs (run_id);

-- +goose Down
DROP TABLE IF EXISTS automation_step_runs;
DROP TABLE IF EXISTS automation_runs;
DROP TABLE IF EXISTS automation_watches;
DROP TABLE IF EXISTS automation_schedules;
DROP TABLE IF EXISTS automation_steps;
DROP TABLE IF EXISTS automation_versions;
DROP TABLE IF EXISTS automations;
