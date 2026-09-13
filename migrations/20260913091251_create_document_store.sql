-- +goose Up

CREATE TABLE document_updates (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  node_id uuid NOT NULL REFERENCES nodes (id) ON DELETE CASCADE,
  client_id text,
  clock bigint,
  schema_version integer NOT NULL DEFAULT 1,
  payload bytea NOT NULL,
  payload_size integer NOT NULL DEFAULT 0,
  created_by uuid REFERENCES users (id) ON DELETE SET NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX document_updates_node_id_idx ON document_updates (node_id, id);
CREATE INDEX document_updates_node_client_idx ON document_updates (node_id, client_id);

CREATE TABLE document_clients (
  node_id uuid NOT NULL REFERENCES nodes (id) ON DELETE CASCADE,
  client_id text NOT NULL,
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  clock bigint NOT NULL DEFAULT 0,
  user_id uuid REFERENCES users (id) ON DELETE SET NULL,
  last_seen_at timestamptz,
  PRIMARY KEY (node_id, client_id)
);

CREATE TABLE document_snapshots (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  node_id uuid NOT NULL REFERENCES nodes (id) ON DELETE CASCADE,
  last_update_id uuid REFERENCES document_updates (id) ON DELETE SET NULL,
  schema_version integer NOT NULL DEFAULT 1,
  payload bytea NOT NULL,
  payload_size integer NOT NULL DEFAULT 0,
  state_vector bytea,
  created_by uuid REFERENCES users (id) ON DELETE SET NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX document_snapshots_node_created_idx ON document_snapshots (node_id, created_at DESC);

CREATE TABLE document_snapshot_history (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  node_id uuid NOT NULL REFERENCES nodes (id) ON DELETE CASCADE,
  snapshot_id uuid NOT NULL REFERENCES document_snapshots (id) ON DELETE CASCADE,
  schema_version integer NOT NULL DEFAULT 1,
  payload bytea NOT NULL,
  payload_size integer NOT NULL DEFAULT 0,
  state_vector bytea,
  label text,
  expires_at timestamptz,
  created_by uuid REFERENCES users (id) ON DELETE SET NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX document_snapshot_history_node_created_idx ON document_snapshot_history (node_id, created_at DESC);
CREATE INDEX document_snapshot_history_expires_at_idx ON document_snapshot_history (expires_at) WHERE expires_at IS NOT NULL;

CREATE TABLE document_json_snapshots (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  node_id uuid NOT NULL REFERENCES nodes (id) ON DELETE CASCADE,
  snapshot_id uuid REFERENCES document_snapshots (id) ON DELETE SET NULL,
  schema_version integer NOT NULL DEFAULT 1,
  tree jsonb NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX document_json_snapshots_node_created_idx ON document_json_snapshots (node_id, created_at DESC);

CREATE TABLE document_compactions (
  node_id uuid PRIMARY KEY REFERENCES nodes (id) ON DELETE CASCADE,
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  last_compacted_update_id uuid REFERENCES document_updates (id) ON DELETE SET NULL,
  retained_update_count integer NOT NULL DEFAULT 0,
  last_compacted_at timestamptz
);

ALTER TABLE nodes
  ADD COLUMN current_snapshot_id uuid REFERENCES document_snapshots (id) ON DELETE SET NULL;
ALTER TABLE nodes
  ADD COLUMN current_json_snapshot_id uuid REFERENCES document_json_snapshots (id) ON DELETE SET NULL;

-- +goose Down
ALTER TABLE nodes DROP COLUMN IF EXISTS current_json_snapshot_id;
ALTER TABLE nodes DROP COLUMN IF EXISTS current_snapshot_id;
DROP TABLE IF EXISTS document_compactions;
DROP TABLE IF EXISTS document_json_snapshots;
DROP TABLE IF EXISTS document_snapshot_history;
DROP TABLE IF EXISTS document_snapshots;
DROP TABLE IF EXISTS document_clients;
DROP TABLE IF EXISTS document_updates;
