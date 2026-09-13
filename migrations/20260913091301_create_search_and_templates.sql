-- +goose Up

CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE search_documents (
  node_id uuid PRIMARY KEY REFERENCES nodes (id) ON DELETE CASCADE,
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  space_id uuid,
  node_type text NOT NULL DEFAULT '',
  language text NOT NULL DEFAULT 'simple',
  title text NOT NULL DEFAULT '',
  body text NOT NULL DEFAULT '',
  document tsvector,
  title_vector tsvector,
  rank_boost real NOT NULL DEFAULT 1,
  permission_set_id uuid,
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX search_documents_workspace_id_idx ON search_documents (workspace_id);
CREATE INDEX search_documents_type_idx ON search_documents (workspace_id, node_type);
CREATE INDEX search_documents_fts_idx ON search_documents USING gin (document);
CREATE INDEX search_documents_title_fts_idx ON search_documents USING gin (title_vector);
CREATE INDEX search_documents_title_trgm_idx ON search_documents USING gin (title gin_trgm_ops);

CREATE TABLE search_synonyms (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  term text NOT NULL,
  synonym text NOT NULL
);

CREATE UNIQUE INDEX search_synonyms_workspace_term_uidx ON search_synonyms (workspace_id, lower(term), lower(synonym));

CREATE TABLE search_aliases (
  node_id uuid NOT NULL REFERENCES nodes (id) ON DELETE CASCADE,
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  alias text NOT NULL,
  PRIMARY KEY (node_id, alias)
);

CREATE INDEX search_aliases_workspace_alias_idx ON search_aliases (workspace_id, lower(alias));

CREATE TABLE templates (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  workspace_id uuid REFERENCES workspaces (id) ON DELETE CASCADE,
  node_id uuid NOT NULL REFERENCES nodes (id) ON DELETE CASCADE,
  created_by uuid REFERENCES users (id) ON DELETE SET NULL,
  name text NOT NULL,
  description text NOT NULL DEFAULT '',
  category text,
  is_instance_template boolean NOT NULL DEFAULT false,
  use_count integer NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX templates_workspace_id_idx ON templates (workspace_id);
CREATE INDEX templates_category_idx ON templates (category);

-- +goose Down
DROP TABLE IF EXISTS templates;
DROP TABLE IF EXISTS search_aliases;
DROP TABLE IF EXISTS search_synonyms;
DROP TABLE IF EXISTS search_documents;
