-- +goose Up

CREATE TABLE notifications (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  actor_id uuid REFERENCES users (id) ON DELETE SET NULL,
  subject_node_id uuid REFERENCES nodes (id) ON DELETE SET NULL,
  type text NOT NULL,
  title text NOT NULL DEFAULT '',
  body text NOT NULL DEFAULT '',
  payload jsonb NOT NULL DEFAULT '{}'::jsonb,
  read_at timestamptz,
  archived_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX notifications_user_unread_idx ON notifications (user_id, created_at DESC) WHERE read_at IS NULL AND archived_at IS NULL;
CREATE INDEX notifications_user_created_idx ON notifications (user_id, created_at DESC);

CREATE TABLE notification_preferences (
  user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  channel text NOT NULL,
  event_type text NOT NULL,
  enabled boolean NOT NULL DEFAULT true,
  updated_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, workspace_id, channel, event_type)
);

CREATE TABLE activity_events (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  node_id uuid REFERENCES nodes (id) ON DELETE SET NULL,
  actor_id uuid REFERENCES users (id) ON DELETE SET NULL,
  event_type text NOT NULL,
  summary text NOT NULL DEFAULT '',
  payload jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX activity_events_workspace_created_idx ON activity_events (workspace_id, created_at DESC);
CREATE INDEX activity_events_node_created_idx ON activity_events (node_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS activity_events;
DROP TABLE IF EXISTS notification_preferences;
DROP TABLE IF EXISTS notifications;
