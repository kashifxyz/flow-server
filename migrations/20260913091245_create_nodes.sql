-- +goose Up

CREATE TABLE nodes (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  space_id uuid REFERENCES spaces (id) ON DELETE SET NULL,
  parent_id uuid REFERENCES nodes (id) ON DELETE CASCADE,
  root_id uuid REFERENCES nodes (id) ON DELETE CASCADE,
  cover_node_id uuid REFERENCES nodes (id) ON DELETE SET NULL,
  created_by uuid REFERENCES users (id) ON DELETE SET NULL,
  last_edited_by uuid REFERENCES users (id) ON DELETE SET NULL,
  archived_by uuid REFERENCES users (id) ON DELETE SET NULL,
  deleted_by uuid REFERENCES users (id) ON DELETE SET NULL,
  locked_by uuid REFERENCES users (id) ON DELETE SET NULL,
  published_by uuid REFERENCES users (id) ON DELETE SET NULL,
  type text NOT NULL,
  subtype text,
  flavour text,
  flavour_version integer NOT NULL DEFAULT 1,
  title text NOT NULL DEFAULT '',
  description text NOT NULL DEFAULT '',
  icon text,
  icon_color text,
  cover_position double precision,
  rank text NOT NULL DEFAULT '0',
  props jsonb NOT NULL DEFAULT '{}'::jsonb,
  schema_version integer NOT NULL DEFAULT 1,
  content_format text NOT NULL DEFAULT 'yjs',
  is_inline boolean NOT NULL DEFAULT false,
  is_template boolean NOT NULL DEFAULT false,
  is_locked boolean NOT NULL DEFAULT false,
  locked_at timestamptz,
  published_at timestamptz,
  public_slug text,
  version bigint NOT NULL DEFAULT 1,
  comment_count integer NOT NULL DEFAULT 0,
  discussion_count integer NOT NULL DEFAULT 0,
  child_count integer NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  last_edited_at timestamptz,
  archived_at timestamptz,
  deleted_at timestamptz,
  restore_until timestamptz,
  purged_at timestamptz
);

CREATE INDEX nodes_workspace_id_idx ON nodes (workspace_id) WHERE deleted_at IS NULL;
CREATE INDEX nodes_space_id_idx ON nodes (space_id) WHERE deleted_at IS NULL;
CREATE INDEX nodes_workspace_type_idx ON nodes (workspace_id, type) WHERE deleted_at IS NULL;
CREATE INDEX nodes_parent_rank_idx ON nodes (parent_id, rank) WHERE deleted_at IS NULL;
CREATE INDEX nodes_root_id_idx ON nodes (root_id) WHERE deleted_at IS NULL;
CREATE INDEX nodes_updated_at_idx ON nodes (workspace_id, updated_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX nodes_deleted_restore_idx ON nodes (workspace_id, restore_until) WHERE deleted_at IS NOT NULL;
CREATE UNIQUE INDEX nodes_public_slug_uidx ON nodes (workspace_id, lower(public_slug)) WHERE public_slug IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX nodes_props_gin_idx ON nodes USING gin (props);

-- +goose Down
DROP TABLE IF EXISTS nodes;
