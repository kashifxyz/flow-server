-- +goose Up

CREATE TABLE node_permissions (
  node_id uuid NOT NULL REFERENCES nodes (id) ON DELETE CASCADE,
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  principal_type text NOT NULL,
  principal_id uuid NOT NULL,
  user_id uuid REFERENCES users (id) ON DELETE CASCADE,
  group_id uuid REFERENCES workspace_groups (id) ON DELETE CASCADE,
  role text NOT NULL,
  inherit boolean NOT NULL DEFAULT true,
  public_role text,
  expires_at timestamptz,
  created_by uuid REFERENCES users (id) ON DELETE SET NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (node_id, principal_type, principal_id)
);

CREATE INDEX node_permissions_workspace_id_idx ON node_permissions (workspace_id);
CREATE INDEX node_permissions_principal_idx ON node_permissions (principal_type, principal_id);
CREATE INDEX node_permissions_user_id_idx ON node_permissions (user_id);
CREATE INDEX node_permissions_group_id_idx ON node_permissions (group_id);

CREATE TABLE permission_changes (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  node_id uuid REFERENCES nodes (id) ON DELETE SET NULL,
  actor_id uuid REFERENCES users (id) ON DELETE SET NULL,
  action text NOT NULL,
  principal_type text,
  principal_id uuid,
  before jsonb,
  after jsonb,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX permission_changes_node_created_idx ON permission_changes (node_id, created_at DESC);

CREATE TABLE permission_templates (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  name text NOT NULL,
  role text NOT NULL,
  inherit boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE node_permission_invites (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  node_id uuid NOT NULL REFERENCES nodes (id) ON DELETE CASCADE,
  email text NOT NULL,
  role text NOT NULL,
  invited_by uuid REFERENCES users (id) ON DELETE SET NULL,
  token_hash text NOT NULL,
  expires_at timestamptz NOT NULL,
  accepted_at timestamptz,
  revoked_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX node_permission_invites_token_hash_uidx ON node_permission_invites (token_hash);
CREATE INDEX node_permission_invites_node_email_idx ON node_permission_invites (node_id, lower(email));

-- +goose Down
DROP TABLE IF EXISTS node_permission_invites;
DROP TABLE IF EXISTS permission_templates;
DROP TABLE IF EXISTS permission_changes;
DROP TABLE IF EXISTS node_permissions;
