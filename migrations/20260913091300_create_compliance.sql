-- +goose Up

CREATE TABLE audit_logs (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  workspace_id uuid REFERENCES workspaces (id) ON DELETE SET NULL,
  actor_id uuid REFERENCES users (id) ON DELETE SET NULL,
  actor_email text,
  session_id uuid,
  request_id text,
  ip inet,
  user_agent text,
  action text NOT NULL,
  resource_type text NOT NULL,
  resource_id uuid,
  node_id uuid REFERENCES nodes (id) ON DELETE SET NULL,
  before jsonb,
  after jsonb,
  metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX audit_logs_workspace_created_idx ON audit_logs (workspace_id, created_at DESC);
CREATE INDEX audit_logs_actor_created_idx ON audit_logs (actor_id, created_at DESC);
CREATE INDEX audit_logs_resource_idx ON audit_logs (resource_type, resource_id);
CREATE INDEX audit_logs_node_id_idx ON audit_logs (node_id);

CREATE TABLE retention_policies (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  name text NOT NULL,
  resource_type text NOT NULL,
  retain_days integer NOT NULL,
  action text NOT NULL DEFAULT 'purge',
  enabled boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX retention_policies_workspace_id_idx ON retention_policies (workspace_id);

CREATE TABLE legal_holds (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  name text NOT NULL,
  reason text NOT NULL DEFAULT '',
  created_by uuid REFERENCES users (id) ON DELETE SET NULL,
  started_at timestamptz NOT NULL DEFAULT now(),
  ended_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX legal_holds_workspace_id_idx ON legal_holds (workspace_id);

CREATE TABLE legal_hold_items (
  hold_id uuid NOT NULL REFERENCES legal_holds (id) ON DELETE CASCADE,
  node_id uuid NOT NULL REFERENCES nodes (id) ON DELETE CASCADE,
  created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (hold_id, node_id)
);

CREATE INDEX legal_hold_items_node_id_idx ON legal_hold_items (node_id);

CREATE TABLE classification_labels (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  workspace_id uuid REFERENCES workspaces (id) ON DELETE CASCADE,
  name text NOT NULL,
  color text,
  rank integer NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX classification_labels_workspace_name_uidx ON classification_labels (workspace_id, lower(name));

CREATE TABLE node_labels (
  node_id uuid NOT NULL REFERENCES nodes (id) ON DELETE CASCADE,
  label_id uuid NOT NULL REFERENCES classification_labels (id) ON DELETE CASCADE,
  applied_by uuid REFERENCES users (id) ON DELETE SET NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (node_id, label_id)
);

CREATE TABLE dlp_policies (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  name text NOT NULL,
  enabled boolean NOT NULL DEFAULT true,
  matchers jsonb NOT NULL DEFAULT '[]'::jsonb,
  actions jsonb NOT NULL DEFAULT '[]'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX dlp_policies_workspace_id_idx ON dlp_policies (workspace_id);

CREATE TABLE dlp_findings (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  policy_id uuid REFERENCES dlp_policies (id) ON DELETE SET NULL,
  node_id uuid REFERENCES nodes (id) ON DELETE SET NULL,
  severity text NOT NULL DEFAULT 'low',
  status text NOT NULL DEFAULT 'open',
  snippet text,
  created_at timestamptz NOT NULL DEFAULT now(),
  resolved_at timestamptz
);

CREATE INDEX dlp_findings_workspace_status_idx ON dlp_findings (workspace_id, status, created_at DESC);

CREATE TABLE discovery_cases (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  name text NOT NULL,
  status text NOT NULL DEFAULT 'open',
  created_by uuid REFERENCES users (id) ON DELETE SET NULL,
  legal_hold_id uuid REFERENCES legal_holds (id) ON DELETE SET NULL,
  query jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now(),
  closed_at timestamptz
);

CREATE TABLE discovery_custodians (
  case_id uuid NOT NULL REFERENCES discovery_cases (id) ON DELETE CASCADE,
  user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (case_id, user_id)
);

CREATE TABLE encryption_keys (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  workspace_id uuid REFERENCES workspaces (id) ON DELETE CASCADE,
  purpose text NOT NULL,
  algorithm text NOT NULL,
  wrapped_key bytea NOT NULL,
  rotated_at timestamptz,
  disabled_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX encryption_keys_workspace_purpose_idx ON encryption_keys (workspace_id, purpose);

CREATE TABLE audit_log_archives (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  workspace_id uuid REFERENCES workspaces (id) ON DELETE SET NULL,
  period_start timestamptz NOT NULL,
  period_end timestamptz NOT NULL,
  object_key text NOT NULL,
  row_count bigint NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS audit_log_archives;
DROP TABLE IF EXISTS encryption_keys;
DROP TABLE IF EXISTS discovery_custodians;
DROP TABLE IF EXISTS discovery_cases;
DROP TABLE IF EXISTS dlp_findings;
DROP TABLE IF EXISTS dlp_policies;
DROP TABLE IF EXISTS node_labels;
DROP TABLE IF EXISTS classification_labels;
DROP TABLE IF EXISTS legal_hold_items;
DROP TABLE IF EXISTS legal_holds;
DROP TABLE IF EXISTS retention_policies;
DROP TABLE IF EXISTS audit_logs;
