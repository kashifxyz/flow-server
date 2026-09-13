-- +goose Up

CREATE TABLE workspace_groups (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  name text NOT NULL,
  description text NOT NULL DEFAULT '',
  created_by uuid REFERENCES users (id) ON DELETE SET NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz
);

CREATE UNIQUE INDEX workspace_groups_workspace_name_uidx ON workspace_groups (workspace_id, lower(name)) WHERE deleted_at IS NULL;

CREATE TABLE workspace_group_members (
  group_id uuid NOT NULL REFERENCES workspace_groups (id) ON DELETE CASCADE,
  user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  added_by uuid REFERENCES users (id) ON DELETE SET NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (group_id, user_id)
);

CREATE INDEX workspace_group_members_user_id_idx ON workspace_group_members (user_id);

CREATE TABLE workspace_invites (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  invited_by uuid REFERENCES users (id) ON DELETE SET NULL,
  email text NOT NULL,
  role text NOT NULL DEFAULT 'member',
  role_id uuid REFERENCES workspace_roles (id) ON DELETE SET NULL,
  group_id uuid REFERENCES workspace_groups (id) ON DELETE SET NULL,
  seat_type text NOT NULL DEFAULT 'full',
  token_hash text NOT NULL,
  message text,
  expires_at timestamptz NOT NULL,
  accepted_at timestamptz,
  accepted_by uuid REFERENCES users (id) ON DELETE SET NULL,
  revoked_at timestamptz,
  reminder_sent_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX workspace_invites_token_hash_uidx ON workspace_invites (token_hash);
CREATE INDEX workspace_invites_workspace_id_idx ON workspace_invites (workspace_id);
CREATE INDEX workspace_invites_email_idx ON workspace_invites (workspace_id, lower(email));

-- +goose Down
DROP TABLE IF EXISTS workspace_invites;
DROP TABLE IF EXISTS workspace_group_members;
DROP TABLE IF EXISTS workspace_groups;
