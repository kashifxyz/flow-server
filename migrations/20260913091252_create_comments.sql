-- +goose Up

CREATE TABLE discussions (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  node_id uuid NOT NULL REFERENCES nodes (id) ON DELETE CASCADE,
  created_by uuid REFERENCES users (id) ON DELETE SET NULL,
  resolved_at timestamptz,
  resolved_by uuid REFERENCES users (id) ON DELETE SET NULL,
  selection_anchor jsonb,
  comment_count integer NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX discussions_node_id_idx ON discussions (node_id, created_at);
CREATE INDEX discussions_unresolved_idx ON discussions (node_id) WHERE resolved_at IS NULL;

CREATE TABLE comments (
  id uuid PRIMARY KEY REFERENCES nodes (id) ON DELETE CASCADE,
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  node_id uuid NOT NULL REFERENCES nodes (id) ON DELETE CASCADE,
  discussion_id uuid NOT NULL REFERENCES discussions (id) ON DELETE CASCADE,
  parent_id uuid REFERENCES comments (id) ON DELETE CASCADE,
  created_by uuid REFERENCES users (id) ON DELETE SET NULL,
  body text NOT NULL DEFAULT '',
  rich_text jsonb NOT NULL DEFAULT '[]'::jsonb,
  body_format text NOT NULL DEFAULT 'plain',
  selection_anchor jsonb,
  edited_at timestamptz,
  deleted_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX comments_node_id_idx ON comments (node_id, created_at);
CREATE INDEX comments_discussion_id_idx ON comments (discussion_id, created_at);
CREATE INDEX comments_workspace_id_idx ON comments (workspace_id);

CREATE TABLE comment_attachments (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  comment_id uuid NOT NULL REFERENCES comments (id) ON DELETE CASCADE,
  file_id uuid REFERENCES nodes (id) ON DELETE SET NULL,
  object_key text,
  mime_type text,
  size_bytes bigint,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX comment_attachments_comment_id_idx ON comment_attachments (comment_id);

CREATE TABLE comment_reactions (
  comment_id uuid NOT NULL REFERENCES comments (id) ON DELETE CASCADE,
  user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  emoji text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (comment_id, user_id, emoji)
);

CREATE TABLE mentions (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  node_id uuid REFERENCES nodes (id) ON DELETE CASCADE,
  comment_id uuid REFERENCES comments (id) ON DELETE CASCADE,
  mentioned_user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  mentioned_by uuid REFERENCES users (id) ON DELETE SET NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX mentions_mentioned_user_idx ON mentions (mentioned_user_id, created_at DESC);
CREATE INDEX mentions_node_id_idx ON mentions (node_id);

-- +goose Down
DROP TABLE IF EXISTS mentions;
DROP TABLE IF EXISTS comment_reactions;
DROP TABLE IF EXISTS comment_attachments;
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS discussions;
