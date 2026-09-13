-- +goose Up

CREATE TABLE webhooks (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  created_by uuid REFERENCES users (id) ON DELETE SET NULL,
  name text NOT NULL,
  url text NOT NULL,
  secret_hash text NOT NULL,
  events text[] NOT NULL DEFAULT '{}'::text[],
  enabled boolean NOT NULL DEFAULT true,
  failure_count integer NOT NULL DEFAULT 0,
  last_success_at timestamptz,
  last_failure_at timestamptz,
  last_error text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz
);

CREATE INDEX webhooks_workspace_id_idx ON webhooks (workspace_id) WHERE deleted_at IS NULL;

CREATE TABLE webhook_deliveries (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  webhook_id uuid NOT NULL REFERENCES webhooks (id) ON DELETE CASCADE,
  workspace_id uuid NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
  event_type text NOT NULL,
  payload jsonb NOT NULL DEFAULT '{}'::jsonb,
  status text NOT NULL DEFAULT 'pending',
  attempt_count integer NOT NULL DEFAULT 0,
  response_status integer,
  response_body text,
  next_attempt_at timestamptz,
  delivered_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX webhook_deliveries_webhook_created_idx ON webhook_deliveries (webhook_id, created_at DESC);
CREATE INDEX webhook_deliveries_pending_idx ON webhook_deliveries (next_attempt_at) WHERE status = 'pending';

-- +goose Down
DROP TABLE IF EXISTS webhook_deliveries;
DROP TABLE IF EXISTS webhooks;
