-- +goose Up

CREATE TABLE canvas_objects (
  id uuid PRIMARY KEY REFERENCES nodes (id) ON DELETE CASCADE,
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  canvas_id uuid NOT NULL REFERENCES nodes (id) ON DELETE CASCADE,
  object_kind text NOT NULL DEFAULT 'shape',
  x double precision NOT NULL DEFAULT 0,
  y double precision NOT NULL DEFAULT 0,
  width double precision NOT NULL DEFAULT 0,
  height double precision NOT NULL DEFAULT 0,
  rotation double precision NOT NULL DEFAULT 0,
  z_index integer NOT NULL DEFAULT 0,
  locked boolean NOT NULL DEFAULT false,
  visible boolean NOT NULL DEFAULT true,
  opacity double precision NOT NULL DEFAULT 1,
  style jsonb NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX canvas_objects_canvas_z_idx ON canvas_objects (canvas_id, z_index);
CREATE INDEX canvas_objects_extent_gist_idx
  ON canvas_objects
  USING gist (box(point(x, y), point(x + width, y + height)));

CREATE TABLE canvas_frames (
  id uuid PRIMARY KEY REFERENCES nodes (id) ON DELETE CASCADE,
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  canvas_id uuid NOT NULL REFERENCES nodes (id) ON DELETE CASCADE,
  title text NOT NULL DEFAULT '',
  x double precision NOT NULL DEFAULT 0,
  y double precision NOT NULL DEFAULT 0,
  width double precision NOT NULL DEFAULT 0,
  height double precision NOT NULL DEFAULT 0,
  collapsed boolean NOT NULL DEFAULT false,
  z_index integer NOT NULL DEFAULT 0
);

CREATE INDEX canvas_frames_canvas_idx ON canvas_frames (canvas_id, z_index);

CREATE TABLE canvas_cameras (
  user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  canvas_id uuid NOT NULL REFERENCES nodes (id) ON DELETE CASCADE,
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  x double precision NOT NULL DEFAULT 0,
  y double precision NOT NULL DEFAULT 0,
  zoom double precision NOT NULL DEFAULT 1,
  updated_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, canvas_id)
);

CREATE INDEX canvas_cameras_canvas_idx ON canvas_cameras (canvas_id);

CREATE TABLE canvas_connectors (
  id uuid PRIMARY KEY REFERENCES nodes (id) ON DELETE CASCADE,
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  canvas_id uuid NOT NULL REFERENCES nodes (id) ON DELETE CASCADE,
  start_object_id uuid NOT NULL REFERENCES canvas_objects (id) ON DELETE CASCADE,
  end_object_id uuid NOT NULL REFERENCES canvas_objects (id) ON DELETE CASCADE,
  shape text NOT NULL DEFAULT 'straight',
  start_snap_to text NOT NULL DEFAULT 'auto',
  end_snap_to text NOT NULL DEFAULT 'auto',
  start_offset_x double precision,
  start_offset_y double precision,
  end_offset_x double precision,
  end_offset_y double precision,
  captions jsonb NOT NULL DEFAULT '[]'::jsonb,
  style jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX canvas_connectors_canvas_idx ON canvas_connectors (canvas_id);
CREATE INDEX canvas_connectors_start_idx ON canvas_connectors (start_object_id);
CREATE INDEX canvas_connectors_end_idx ON canvas_connectors (end_object_id);

CREATE TABLE canvas_tags (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  canvas_id uuid NOT NULL REFERENCES nodes (id) ON DELETE CASCADE,
  name text NOT NULL,
  color text,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX canvas_tags_canvas_name_uidx ON canvas_tags (canvas_id, lower(name));

CREATE TABLE canvas_tag_items (
  tag_id uuid NOT NULL REFERENCES canvas_tags (id) ON DELETE CASCADE,
  object_id uuid NOT NULL REFERENCES canvas_objects (id) ON DELETE CASCADE,
  created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (tag_id, object_id)
);

CREATE INDEX canvas_tag_items_object_idx ON canvas_tag_items (object_id);

-- +goose Down
DROP TABLE IF EXISTS canvas_tag_items;
DROP TABLE IF EXISTS canvas_tags;
DROP TABLE IF EXISTS canvas_connectors;
DROP TABLE IF EXISTS canvas_cameras;
DROP TABLE IF EXISTS canvas_frames;
DROP TABLE IF EXISTS canvas_objects;
