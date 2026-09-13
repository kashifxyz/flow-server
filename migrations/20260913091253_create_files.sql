-- +goose Up

CREATE TABLE files (
  id uuid PRIMARY KEY REFERENCES nodes (id) ON DELETE CASCADE,
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  uploader_id uuid REFERENCES users (id) ON DELETE SET NULL,
  bucket text,
  object_key text NOT NULL,
  original_name text NOT NULL,
  mime_type text NOT NULL,
  size_bytes bigint NOT NULL,
  checksum text,
  etag text,
  status text NOT NULL DEFAULT 'pending',
  scan_status text NOT NULL DEFAULT 'pending',
  scanned_at timestamptz,
  storage_class text,
  encryption_key_id text,
  thumbnail_object_key text,
  preview_object_key text,
  width integer,
  height integer,
  duration_ms integer,
  page_count integer,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX files_object_key_uidx ON files (object_key);
CREATE INDEX files_workspace_id_idx ON files (workspace_id);
CREATE INDEX files_status_idx ON files (status);

CREATE TABLE file_revisions (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  file_id uuid NOT NULL REFERENCES files (id) ON DELETE CASCADE,
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  object_key text NOT NULL,
  original_name text NOT NULL,
  mime_type text NOT NULL,
  size_bytes bigint NOT NULL,
  checksum text,
  created_by uuid REFERENCES users (id) ON DELETE SET NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX file_revisions_object_key_uidx ON file_revisions (object_key);
CREATE INDEX file_revisions_file_created_idx ON file_revisions (file_id, created_at DESC);

CREATE TABLE file_upload_sessions (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  file_id uuid REFERENCES files (id) ON DELETE CASCADE,
  created_by uuid REFERENCES users (id) ON DELETE SET NULL,
  upload_id text NOT NULL,
  object_key text NOT NULL,
  mime_type text NOT NULL,
  expected_size_bytes bigint,
  part_count integer NOT NULL DEFAULT 0,
  status text NOT NULL DEFAULT 'open',
  expires_at timestamptz NOT NULL,
  completed_at timestamptz,
  aborted_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX file_upload_sessions_upload_id_uidx ON file_upload_sessions (upload_id);
CREATE INDEX file_upload_sessions_expires_idx ON file_upload_sessions (expires_at) WHERE status = 'open';

CREATE TABLE file_tombstones (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  object_key text NOT NULL,
  reason text,
  delete_after timestamptz NOT NULL,
  deleted_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX file_tombstones_object_key_uidx ON file_tombstones (object_key);
CREATE INDEX file_tombstones_delete_after_idx ON file_tombstones (delete_after) WHERE deleted_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS file_tombstones;
DROP TABLE IF EXISTS file_upload_sessions;
DROP TABLE IF EXISTS file_revisions;
DROP TABLE IF EXISTS files;
