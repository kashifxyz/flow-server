-- +goose Up

CREATE TABLE share_links (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  node_id uuid NOT NULL REFERENCES nodes (id) ON DELETE CASCADE,
  created_by uuid REFERENCES users (id) ON DELETE SET NULL,
  token_hash text NOT NULL,
  role text NOT NULL,
  password_hash text,
  include_subtree boolean NOT NULL DEFAULT true,
  allow_duplicate boolean NOT NULL DEFAULT false,
  domain_allowlist text[] NOT NULL DEFAULT '{}'::text[],
  max_uses integer,
  use_count integer NOT NULL DEFAULT 0,
  watermark boolean NOT NULL DEFAULT false,
  expires_at timestamptz,
  last_accessed_at timestamptz,
  revoked_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX share_links_token_hash_uidx ON share_links (token_hash);
CREATE INDEX share_links_node_id_idx ON share_links (node_id);

CREATE TABLE share_link_accesses (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  share_link_id uuid NOT NULL REFERENCES share_links (id) ON DELETE CASCADE,
  ip inet,
  user_agent text,
  user_id uuid REFERENCES users (id) ON DELETE SET NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX share_link_accesses_link_created_idx ON share_link_accesses (share_link_id, created_at DESC);

CREATE TABLE published_pages (
  node_id uuid PRIMARY KEY REFERENCES nodes (id) ON DELETE CASCADE,
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  slug text NOT NULL,
  title text NOT NULL DEFAULT '',
  indexed boolean NOT NULL DEFAULT false,
  published_by uuid REFERENCES users (id) ON DELETE SET NULL,
  unpublished_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX published_pages_workspace_slug_uidx ON published_pages (workspace_id, lower(slug));

-- +goose Down
DROP TABLE IF EXISTS published_pages;
DROP TABLE IF EXISTS share_link_accesses;
DROP TABLE IF EXISTS share_links;
