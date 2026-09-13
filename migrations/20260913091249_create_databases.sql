-- +goose Up

CREATE TABLE databases (
  id uuid PRIMARY KEY REFERENCES nodes (id) ON DELETE CASCADE,
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  is_inline boolean NOT NULL DEFAULT false,
  primary_field_id uuid,
  description text NOT NULL DEFAULT '',
  row_count integer NOT NULL DEFAULT 0,
  schema_version integer NOT NULL DEFAULT 1,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX databases_workspace_id_idx ON databases (workspace_id);

CREATE TABLE fields (
  id uuid PRIMARY KEY REFERENCES nodes (id) ON DELETE CASCADE,
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  database_id uuid NOT NULL REFERENCES databases (id) ON DELETE CASCADE,
  name text NOT NULL,
  type text NOT NULL,
  description text NOT NULL DEFAULT '',
  options jsonb NOT NULL DEFAULT '{}'::jsonb,
  default_value jsonb,
  is_primary boolean NOT NULL DEFAULT false,
  is_required boolean NOT NULL DEFAULT false,
  is_unique boolean NOT NULL DEFAULT false,
  is_computed boolean NOT NULL DEFAULT false,
  is_hidden boolean NOT NULL DEFAULT false,
  is_valid boolean NOT NULL DEFAULT true,
  is_reversed boolean NOT NULL DEFAULT false,
  prefers_single_record boolean NOT NULL DEFAULT false,
  formula text,
  formula_ast jsonb,
  result_type text,
  result_options jsonb,
  compute_error text,
  last_computed_at timestamptz,
  numeric_precision integer,
  currency text,
  duration_format text,
  rating_max integer,
  linked_database_id uuid REFERENCES databases (id) ON DELETE SET NULL,
  inverse_field_id uuid REFERENCES fields (id) ON DELETE SET NULL,
  relation_field_id uuid REFERENCES fields (id) ON DELETE SET NULL,
  lookup_field_id uuid REFERENCES fields (id) ON DELETE SET NULL,
  rollup_field_id uuid REFERENCES fields (id) ON DELETE SET NULL,
  rollup_function text,
  referenced_field_ids uuid[] NOT NULL DEFAULT '{}'::uuid[],
  rank text NOT NULL DEFAULT '0',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX fields_database_name_uidx ON fields (database_id, lower(name));
CREATE INDEX fields_workspace_id_idx ON fields (workspace_id);
CREATE INDEX fields_database_rank_idx ON fields (database_id, rank);

ALTER TABLE databases
  ADD CONSTRAINT databases_primary_field_id_fkey
  FOREIGN KEY (primary_field_id) REFERENCES fields (id) ON DELETE SET NULL;

CREATE TABLE field_options (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  field_id uuid NOT NULL REFERENCES fields (id) ON DELETE CASCADE,
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  key text NOT NULL,
  label text NOT NULL,
  color text,
  rank text NOT NULL DEFAULT '0',
  disabled_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX field_options_field_key_uidx ON field_options (field_id, key);
CREATE INDEX field_options_field_rank_idx ON field_options (field_id, rank);

CREATE TABLE field_dependencies (
  field_id uuid NOT NULL REFERENCES fields (id) ON DELETE CASCADE,
  depends_on_field_id uuid NOT NULL REFERENCES fields (id) ON DELETE CASCADE,
  kind text NOT NULL,
  PRIMARY KEY (field_id, depends_on_field_id, kind)
);

CREATE INDEX field_dependencies_depends_on_idx ON field_dependencies (depends_on_field_id);

CREATE TABLE relation_constraints (
  field_id uuid PRIMARY KEY REFERENCES fields (id) ON DELETE CASCADE,
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  from_database_id uuid NOT NULL REFERENCES databases (id) ON DELETE CASCADE,
  to_database_id uuid NOT NULL REFERENCES databases (id) ON DELETE CASCADE,
  allow_multiple boolean NOT NULL DEFAULT true,
  is_symmetric boolean NOT NULL DEFAULT false
);

CREATE INDEX relation_constraints_to_database_idx ON relation_constraints (to_database_id);

CREATE TABLE views (
  id uuid PRIMARY KEY REFERENCES nodes (id) ON DELETE CASCADE,
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  database_id uuid NOT NULL REFERENCES databases (id) ON DELETE CASCADE,
  type text NOT NULL,
  name text NOT NULL DEFAULT '',
  filter jsonb NOT NULL DEFAULT '[]'::jsonb,
  sorts jsonb NOT NULL DEFAULT '[]'::jsonb,
  groups jsonb NOT NULL DEFAULT '[]'::jsonb,
  visible_field_ids uuid[] NOT NULL DEFAULT '{}'::uuid[],
  frozen_field_ids uuid[] NOT NULL DEFAULT '{}'::uuid[],
  column_widths jsonb NOT NULL DEFAULT '{}'::jsonb,
  calendar_field_id uuid REFERENCES fields (id) ON DELETE SET NULL,
  board_field_id uuid REFERENCES fields (id) ON DELETE SET NULL,
  timeline_start_field_id uuid REFERENCES fields (id) ON DELETE SET NULL,
  timeline_end_field_id uuid REFERENCES fields (id) ON DELETE SET NULL,
  map_field_id uuid REFERENCES fields (id) ON DELETE SET NULL,
  owner_user_id uuid REFERENCES users (id) ON DELETE SET NULL,
  is_default boolean NOT NULL DEFAULT false,
  is_personal boolean NOT NULL DEFAULT false,
  is_locked boolean NOT NULL DEFAULT false,
  row_height text,
  color_rules jsonb NOT NULL DEFAULT '[]'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX views_database_id_idx ON views (database_id);

CREATE TABLE field_values (
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  record_id uuid NOT NULL REFERENCES nodes (id) ON DELETE CASCADE,
  field_id uuid NOT NULL REFERENCES nodes (id) ON DELETE CASCADE,
  value jsonb,
  computed_value jsonb,
  computed_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (record_id, field_id)
);

CREATE INDEX field_values_workspace_id_idx ON field_values (workspace_id);
CREATE INDEX field_values_field_id_idx ON field_values (field_id);
CREATE INDEX field_values_value_gin_idx ON field_values USING gin (value);

CREATE TABLE field_value_revisions (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  record_id uuid NOT NULL REFERENCES nodes (id) ON DELETE CASCADE,
  field_id uuid NOT NULL REFERENCES nodes (id) ON DELETE CASCADE,
  actor_id uuid REFERENCES users (id) ON DELETE SET NULL,
  value jsonb,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX field_value_revisions_record_field_idx ON field_value_revisions (record_id, field_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS field_value_revisions;
DROP TABLE IF EXISTS field_values;
DROP TABLE IF EXISTS views;
DROP TABLE IF EXISTS relation_constraints;
DROP TABLE IF EXISTS field_dependencies;
DROP TABLE IF EXISTS field_options;
ALTER TABLE databases DROP CONSTRAINT IF EXISTS databases_primary_field_id_fkey;
DROP TABLE IF EXISTS fields;
DROP TABLE IF EXISTS databases;
