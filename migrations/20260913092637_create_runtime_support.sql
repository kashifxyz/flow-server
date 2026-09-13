-- +goose Up

CREATE TABLE block_flavours (
  flavour text NOT NULL,
  version integer NOT NULL,
  node_type text NOT NULL,
  schema jsonb NOT NULL DEFAULT '{}'::jsonb,
  deprecated_at timestamptz,
  PRIMARY KEY (flavour, version)
);

CREATE TABLE trash_items (
  node_id uuid PRIMARY KEY REFERENCES nodes (id) ON DELETE CASCADE,
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  deleted_by uuid REFERENCES users (id) ON DELETE SET NULL,
  deleted_at timestamptz NOT NULL DEFAULT now(),
  restore_until timestamptz NOT NULL,
  original_parent_id uuid,
  original_rank text
);

CREATE INDEX trash_items_restore_until_idx ON trash_items (restore_until);

CREATE TABLE unique_field_values (
  field_id uuid NOT NULL REFERENCES fields (id) ON DELETE CASCADE,
  value_key text NOT NULL,
  record_id uuid NOT NULL REFERENCES nodes (id) ON DELETE CASCADE,
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  PRIMARY KEY (field_id, value_key)
);

CREATE INDEX unique_field_values_record_id_idx ON unique_field_values (record_id);

CREATE TABLE formula_functions (
  name text PRIMARY KEY,
  arity integer,
  result_type text,
  description text NOT NULL DEFAULT '',
  enabled boolean NOT NULL DEFAULT true
);

CREATE TABLE formula_jobs (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  field_id uuid NOT NULL REFERENCES fields (id) ON DELETE CASCADE,
  record_id uuid REFERENCES nodes (id) ON DELETE CASCADE,
  status text NOT NULL DEFAULT 'queued',
  reason text,
  error text,
  run_at timestamptz NOT NULL DEFAULT now(),
  started_at timestamptz,
  finished_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX formula_jobs_status_run_at_idx ON formula_jobs (status, run_at);
CREATE INDEX formula_jobs_field_id_idx ON formula_jobs (field_id);

CREATE TABLE yjs_documents (
  node_id uuid PRIMARY KEY REFERENCES nodes (id) ON DELETE CASCADE,
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  guid text NOT NULL,
  collection text,
  gc_enabled boolean NOT NULL DEFAULT true,
  schema_name text,
  schema_version integer NOT NULL DEFAULT 1,
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX yjs_documents_guid_uidx ON yjs_documents (guid);

CREATE TABLE canvas_groups (
  id uuid PRIMARY KEY REFERENCES nodes (id) ON DELETE CASCADE,
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  canvas_id uuid NOT NULL REFERENCES nodes (id) ON DELETE CASCADE,
  title text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX canvas_groups_canvas_idx ON canvas_groups (canvas_id);

CREATE TABLE canvas_group_items (
  group_id uuid NOT NULL REFERENCES canvas_groups (id) ON DELETE CASCADE,
  object_id uuid NOT NULL REFERENCES canvas_objects (id) ON DELETE CASCADE,
  rank text NOT NULL DEFAULT '0',
  PRIMARY KEY (group_id, object_id)
);

CREATE TABLE canvas_connector_vertices (
  connector_id uuid NOT NULL REFERENCES canvas_connectors (id) ON DELETE CASCADE,
  seq integer NOT NULL,
  x double precision NOT NULL,
  y double precision NOT NULL,
  PRIMARY KEY (connector_id, seq)
);

CREATE TABLE rate_limit_policies (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  name text NOT NULL,
  scope text NOT NULL,
  key_template text NOT NULL,
  limit_count integer NOT NULL,
  window_seconds integer NOT NULL,
  enabled boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX rate_limit_policies_name_uidx ON rate_limit_policies (lower(name));

CREATE TABLE presence_states (
  user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  node_id uuid NOT NULL REFERENCES nodes (id) ON DELETE CASCADE,
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  session_id uuid,
  selection jsonb,
  cursor jsonb,
  viewport jsonb,
  last_seen_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, node_id)
);

CREATE INDEX presence_states_node_seen_idx ON presence_states (node_id, last_seen_at DESC);

CREATE TABLE saved_searches (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  name text NOT NULL,
  query text NOT NULL,
  filters jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX saved_searches_user_idx ON saved_searches (user_id, workspace_id);

CREATE TABLE dashboards (
  id uuid PRIMARY KEY REFERENCES nodes (id) ON DELETE CASCADE,
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  name text NOT NULL DEFAULT '',
  layout jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE dashboard_widgets (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  dashboard_id uuid NOT NULL REFERENCES dashboards (id) ON DELETE CASCADE,
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  type text NOT NULL,
  config jsonb NOT NULL DEFAULT '{}'::jsonb,
  x integer NOT NULL DEFAULT 0,
  y integer NOT NULL DEFAULT 0,
  width integer NOT NULL DEFAULT 1,
  height integer NOT NULL DEFAULT 1,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX dashboard_widgets_dashboard_idx ON dashboard_widgets (dashboard_id);

CREATE TABLE recurrence_rules (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  node_id uuid NOT NULL REFERENCES nodes (id) ON DELETE CASCADE,
  rrule text NOT NULL,
  dtstart timestamptz NOT NULL,
  timezone text NOT NULL DEFAULT 'UTC',
  until_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX recurrence_rules_node_id_idx ON recurrence_rules (node_id);

CREATE TABLE calendar_events (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  node_id uuid NOT NULL REFERENCES nodes (id) ON DELETE CASCADE,
  recurrence_id uuid REFERENCES recurrence_rules (id) ON DELETE SET NULL,
  start_at timestamptz NOT NULL,
  end_at timestamptz,
  all_day boolean NOT NULL DEFAULT false,
  timezone text NOT NULL DEFAULT 'UTC',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX calendar_events_workspace_start_idx ON calendar_events (workspace_id, start_at);
CREATE INDEX calendar_events_node_id_idx ON calendar_events (node_id);

CREATE TABLE event_outbox (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  workspace_id uuid REFERENCES workspaces (id) ON DELETE CASCADE,
  event_type text NOT NULL,
  payload jsonb NOT NULL DEFAULT '{}'::jsonb,
  status text NOT NULL DEFAULT 'pending',
  attempts integer NOT NULL DEFAULT 0,
  available_at timestamptz NOT NULL DEFAULT now(),
  published_at timestamptz,
  last_error text,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX event_outbox_status_available_idx ON event_outbox (status, available_at);

CREATE TABLE idempotency_keys (
  key text NOT NULL,
  user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  request_hash text NOT NULL,
  response_status integer,
  response_body jsonb,
  created_at timestamptz NOT NULL DEFAULT now(),
  expires_at timestamptz NOT NULL,
  PRIMARY KEY (key, user_id)
);

CREATE INDEX idempotency_keys_expires_at_idx ON idempotency_keys (expires_at);

CREATE TABLE notification_digests (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  channel text NOT NULL,
  period text NOT NULL,
  next_run_at timestamptz NOT NULL,
  last_run_at timestamptz
);

CREATE UNIQUE INDEX notification_digests_user_channel_period_uidx
  ON notification_digests (user_id, workspace_id, channel, period);

-- +goose Down
DROP TABLE IF EXISTS notification_digests;
DROP TABLE IF EXISTS idempotency_keys;
DROP TABLE IF EXISTS event_outbox;
DROP TABLE IF EXISTS calendar_events;
DROP TABLE IF EXISTS recurrence_rules;
DROP TABLE IF EXISTS dashboard_widgets;
DROP TABLE IF EXISTS dashboards;
DROP TABLE IF EXISTS saved_searches;
DROP TABLE IF EXISTS presence_states;
DROP TABLE IF EXISTS rate_limit_policies;
DROP TABLE IF EXISTS canvas_connector_vertices;
DROP TABLE IF EXISTS canvas_group_items;
DROP TABLE IF EXISTS canvas_groups;
DROP TABLE IF EXISTS yjs_documents;
DROP TABLE IF EXISTS formula_jobs;
DROP TABLE IF EXISTS formula_functions;
DROP TABLE IF EXISTS unique_field_values;
DROP TABLE IF EXISTS trash_items;
DROP TABLE IF EXISTS block_flavours;
