-- +goose Up

CREATE TABLE organizations (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  name text NOT NULL,
  slug text,
  created_by uuid REFERENCES users (id) ON DELETE SET NULL,
  locale text NOT NULL DEFAULT 'en',
  timezone text NOT NULL DEFAULT 'UTC',
  data_region text,
  settings jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz
);

CREATE UNIQUE INDEX organizations_slug_uidx ON organizations (lower(slug)) WHERE slug IS NOT NULL AND deleted_at IS NULL;

CREATE TABLE org_units (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  organization_id uuid NOT NULL REFERENCES organizations (id) ON DELETE CASCADE,
  parent_id uuid REFERENCES org_units (id) ON DELETE CASCADE,
  name text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX org_units_organization_id_idx ON org_units (organization_id);
CREATE INDEX org_units_parent_id_idx ON org_units (parent_id);

CREATE TABLE org_unit_members (
  org_unit_id uuid NOT NULL REFERENCES org_units (id) ON DELETE CASCADE,
  user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (org_unit_id, user_id)
);

CREATE INDEX org_unit_members_user_id_idx ON org_unit_members (user_id);

CREATE TABLE workspaces (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  organization_id uuid REFERENCES organizations (id) ON DELETE SET NULL,
  org_unit_id uuid REFERENCES org_units (id) ON DELETE SET NULL,
  created_by uuid REFERENCES users (id) ON DELETE SET NULL,
  owner_user_id uuid REFERENCES users (id) ON DELETE SET NULL,
  name text NOT NULL,
  slug text,
  description text NOT NULL DEFAULT '',
  icon text,
  color text,
  locale text NOT NULL DEFAULT 'en',
  timezone text NOT NULL DEFAULT 'UTC',
  status text NOT NULL DEFAULT 'active',
  settings jsonb NOT NULL DEFAULT '{}'::jsonb,
  feature_flags jsonb NOT NULL DEFAULT '{}'::jsonb,
  storage_used_bytes bigint NOT NULL DEFAULT 0,
  storage_quota_bytes bigint,
  member_quota integer,
  guest_quota integer,
  data_region text,
  retention_days integer,
  trash_retention_days integer NOT NULL DEFAULT 30,
  organization_name text,
  organization_domain text,
  billing_email text,
  public_sharing_enabled boolean NOT NULL DEFAULT false,
  guest_access_enabled boolean NOT NULL DEFAULT true,
  suspended_at timestamptz,
  suspend_reason text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz,
  purged_at timestamptz
);

CREATE INDEX workspaces_owner_user_id_idx ON workspaces (owner_user_id);
CREATE INDEX workspaces_organization_id_idx ON workspaces (organization_id);
CREATE UNIQUE INDEX workspaces_slug_uidx ON workspaces (lower(slug)) WHERE slug IS NOT NULL AND deleted_at IS NULL;

CREATE TABLE spaces (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  name text NOT NULL,
  description text NOT NULL DEFAULT '',
  icon text,
  is_private boolean NOT NULL DEFAULT false,
  is_default boolean NOT NULL DEFAULT false,
  created_by uuid REFERENCES users (id) ON DELETE SET NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz
);

CREATE INDEX spaces_workspace_id_idx ON spaces (workspace_id) WHERE deleted_at IS NULL;

CREATE TABLE space_members (
  space_id uuid NOT NULL REFERENCES spaces (id) ON DELETE CASCADE,
  user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  role text NOT NULL DEFAULT 'member',
  created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (space_id, user_id)
);

CREATE INDEX space_members_user_id_idx ON space_members (user_id);

CREATE TABLE workspace_domains (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  domain text NOT NULL,
  verified_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX workspace_domains_domain_uidx ON workspace_domains (lower(domain));
CREATE INDEX workspace_domains_workspace_id_idx ON workspace_domains (workspace_id);

CREATE TABLE workspace_roles (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  name text NOT NULL,
  description text NOT NULL DEFAULT '',
  is_system boolean NOT NULL DEFAULT false,
  permissions jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX workspace_roles_workspace_name_uidx ON workspace_roles (workspace_id, lower(name));

CREATE TABLE workspace_members (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  role text NOT NULL DEFAULT 'member',
  role_id uuid REFERENCES workspace_roles (id) ON DELETE SET NULL,
  seat_type text NOT NULL DEFAULT 'full',
  status text NOT NULL DEFAULT 'active',
  invited_by uuid REFERENCES users (id) ON DELETE SET NULL,
  joined_at timestamptz,
  last_seen_at timestamptz,
  guest_expires_at timestamptz,
  deactivated_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX workspace_members_workspace_user_uidx ON workspace_members (workspace_id, user_id);
CREATE INDEX workspace_members_user_id_idx ON workspace_members (user_id);
CREATE INDEX workspace_members_role_id_idx ON workspace_members (role_id);

CREATE TABLE workspace_usage_daily (
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  day date NOT NULL,
  storage_bytes bigint NOT NULL DEFAULT 0,
  file_count integer NOT NULL DEFAULT 0,
  member_count integer NOT NULL DEFAULT 0,
  api_request_count bigint NOT NULL DEFAULT 0,
  automation_run_count bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (workspace_id, day)
);

ALTER TABLE api_tokens
  ADD CONSTRAINT api_tokens_workspace_id_fkey
  FOREIGN KEY (workspace_id) REFERENCES workspaces (id) ON DELETE CASCADE;

-- +goose Down
ALTER TABLE api_tokens DROP CONSTRAINT IF EXISTS api_tokens_workspace_id_fkey;
DROP TABLE IF EXISTS workspace_usage_daily;
DROP TABLE IF EXISTS workspace_members;
DROP TABLE IF EXISTS workspace_roles;
DROP TABLE IF EXISTS workspace_domains;
DROP TABLE IF EXISTS space_members;
DROP TABLE IF EXISTS spaces;
DROP TABLE IF EXISTS workspaces;
DROP TABLE IF EXISTS org_unit_members;
DROP TABLE IF EXISTS org_units;
DROP TABLE IF EXISTS organizations;
