-- +goose Up

CREATE TABLE node_edges (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  from_id uuid NOT NULL REFERENCES nodes (id) ON DELETE CASCADE,
  to_id uuid NOT NULL REFERENCES nodes (id) ON DELETE CASCADE,
  field_id uuid REFERENCES nodes (id) ON DELETE CASCADE,
  inverse_edge_id uuid REFERENCES node_edges (id) ON DELETE SET NULL,
  kind text NOT NULL,
  rank text NOT NULL DEFAULT '0',
  start_anchor text,
  end_anchor text,
  label text,
  props jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_by uuid REFERENCES users (id) ON DELETE SET NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX node_edges_workspace_id_idx ON node_edges (workspace_id);
CREATE INDEX node_edges_from_kind_rank_idx ON node_edges (from_id, kind, rank);
CREATE INDEX node_edges_to_kind_idx ON node_edges (to_id, kind);
CREATE INDEX node_edges_field_id_idx ON node_edges (field_id);
CREATE UNIQUE INDEX node_edges_relation_uidx ON node_edges (from_id, to_id, field_id) WHERE kind = 'relation' AND field_id IS NOT NULL;

CREATE TABLE node_closure (
  ancestor_id uuid NOT NULL REFERENCES nodes (id) ON DELETE CASCADE,
  descendant_id uuid NOT NULL REFERENCES nodes (id) ON DELETE CASCADE,
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  depth integer NOT NULL,
  PRIMARY KEY (ancestor_id, descendant_id)
);

CREATE INDEX node_closure_descendant_idx ON node_closure (descendant_id);
CREATE INDEX node_closure_workspace_idx ON node_closure (workspace_id, ancestor_id);

-- +goose Down
DROP TABLE IF EXISTS node_closure;
DROP TABLE IF EXISTS node_edges;
