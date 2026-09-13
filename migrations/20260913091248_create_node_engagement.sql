-- +goose Up

CREATE TABLE node_stars (
  user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  node_id uuid NOT NULL REFERENCES nodes (id) ON DELETE CASCADE,
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, node_id)
);

CREATE INDEX node_stars_workspace_user_idx ON node_stars (workspace_id, user_id, created_at DESC);

CREATE TABLE node_pins (
  user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  node_id uuid NOT NULL REFERENCES nodes (id) ON DELETE CASCADE,
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  rank text NOT NULL DEFAULT '0',
  created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, node_id)
);

CREATE INDEX node_pins_workspace_user_rank_idx ON node_pins (workspace_id, user_id, rank);

CREATE TABLE node_watches (
  user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  node_id uuid NOT NULL REFERENCES nodes (id) ON DELETE CASCADE,
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  include_children boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, node_id)
);

CREATE INDEX node_watches_node_id_idx ON node_watches (node_id);

CREATE TABLE node_visits (
  user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  node_id uuid NOT NULL REFERENCES nodes (id) ON DELETE CASCADE,
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  visit_count integer NOT NULL DEFAULT 1,
  last_visited_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, node_id)
);

CREATE INDEX node_visits_workspace_user_last_idx ON node_visits (workspace_id, user_id, last_visited_at DESC);

-- +goose Down
DROP TABLE IF EXISTS node_visits;
DROP TABLE IF EXISTS node_watches;
DROP TABLE IF EXISTS node_pins;
DROP TABLE IF EXISTS node_stars;
