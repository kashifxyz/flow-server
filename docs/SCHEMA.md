# Flow database schema

Canonical persistence model for Flow. PostgreSQL 18 is the source of truth for durable state. Redis is transient. Garage/S3 holds bytes. This document describes the schema in `flow-server/migrations/`.

**Status:** Goose files are written and **not applied by agents**. Apply locally:

```text
goose -dir migrations postgres "host=localhost port=5432 dbname=flow sslmode=disable" up
```

**Inventory:** 23 SQL migrations, **121 tables**. IDs are `uuid` with `DEFAULT uuidv7()`. Timestamps are `timestamptz`. Enums are `text` (validated in Go). No `CHECK` constraints, no triggers, no PL/pgSQL functions.

---

## 1. Principles

1. **One identity graph.** Pages, blocks, databases, fields, records, views, projects, tasks, files, canvases, objects, comments, and dashboards are rows in `nodes` (or 1:1 satellite tables keyed by `nodes.id`).
2. **Workspace is the tenant.** Almost every owned row has `workspace_id`. Cross-workspace access is a bug.
3. **Database owns structure, Go owns rules.** Primary keys, foreign keys, and uniqueness that would corrupt data live in Postgres. Allowed roles, node types, and field types are validated in Go.
4. **Soft delete first.** `deleted_at` / `restore_until` / `trash_items`. Hard delete and object GC are jobs.
5. **Secrets are hashed or wrapped.** Session tokens, invite tokens, webhook secrets, recovery codes, TOTP, and wrapped DEKs never sit in plaintext.
6. **CRDT payloads are `bytea`.** JSON trees are a derived snapshot, not the collaborative source of truth.
7. **Presence live data is Redis.** `presence_states` is last-known recovery, not the pub/sub fabric.
8. **Rate-limit counters are Redis.** `rate_limit_policies` is configuration only.
9. **No OAuth/OIDC/SAML/SCIM tables.** Product rule. MFA is TOTP + recovery codes + WebAuthn passkeys.

---

## 2. Conventions

| Topic | Rule |
| --- | --- |
| Primary keys | `uuid`, `uuidv7()` |
| Time | UTC `timestamptz` |
| JSON | `jsonb` with empty-object/array defaults |
| Rank / order | `text` LexoRank for user-reorderable lists (`nodes.rank`, pins, fields). `integer` for system order (`z_index`, `automation_steps.rank`, `classification_labels.rank`, vertex `seq`). |
| Email uniqueness | `UNIQUE (lower(email))` where not deleted |
| Token storage | SHA-256 (or equivalent) hash in `*_hash` columns |
| 1:1 satellites | `id uuid PRIMARY KEY REFERENCES nodes(id) ON DELETE CASCADE` |
| Extensions | `pg_trgm` for title search |
| App-maintained | `updated_at`, `node_closure`, `search_documents` tsvectors, denormalized counts, `storage_used_bytes` |
| `created_by` NULL | Allowed. System-generated **or** user later deleted (`ON DELETE SET NULL`). Go sets it on user-created inserts. |

**Node `type` values (Go):** `page`, `block`, `database`, `field`, `record`, `view`, `project`, `task`, `file`, `canvas`, `object`, `comment`, `dashboard`, `frame`, `connector`, `group`.

**Membership `role` (Go):** `owner`, `admin`, `member`, `guest`.

**Node ACL `role` (Go):** `view`, `comment`, `edit`, `manage`.

---

## 3. Storage map

```text
Browser working state
        │
        ▼
   Flow server (Go)
        │
        ├── PostgreSQL 18     durable application state
        ├── Redis             cache, live presence, rate-limit counters
        └── Garage (S3)       file bytes, export artifacts, audit archives
```

If Redis disappears: documents, ACL, files metadata, and jobs remain correct. Live cursors and in-window rate counters reset.

---

## 4. Entity overview

```text
organizations
  └── org_units
        └── workspaces  (tenant)
              ├── spaces (teamspaces)
              ├── members / groups / roles / invites
              └── nodes (identity)
                    ├── parent_id / node_closure
                    ├── node_edges / node_permissions
                    ├── databases → fields → field_values
                    ├── canvas_* / document_* / yjs_documents
                    ├── files → Garage object_key
                    └── comments / discussions
```

---

## 5. Migration order

| File | Domain |
| --- | --- |
| `20260913091240_create_users.sql` | users, preferences |
| `20260913091241_create_user_security.sql` | passwords, MFA, devices, API tokens, WebAuthn |
| `20260913091242_create_sessions.sql` | sessions, auth tokens, login telemetry |
| `20260913091243_create_workspaces.sql` | orgs, OUs, workspaces, spaces, roles, usage |
| `20260913091244_create_workspace_access.sql` | groups, invites |
| `20260913091245_create_nodes.sql` | identity graph |
| `20260913091246_create_node_graph.sql` | edges, closure |
| `20260913091247_create_node_permissions.sql` | ACL, guests, templates |
| `20260913091248_create_node_engagement.sql` | stars, pins, watches, visits |
| `20260913091249_create_databases.sql` | Airtable-class catalog |
| `20260913091250_create_canvas.sql` | objects, frames, connectors, tags |
| `20260913091251_create_document_store.sql` | Yjs log + snapshots |
| `20260913091252_create_comments.sql` | discussions |
| `20260913091253_create_files.sql` | object metadata |
| `20260913091254_create_sharing.sql` | links, published pages |
| `20260913091256_create_notifications.sql` | inbox + activity |
| `20260913091257_create_automations.sql` | workflows |
| `20260913091258_create_webhooks.sql` | outbound HTTP |
| `20260913091259_create_jobs.sql` | workers, mail, import/export |
| `20260913091300_create_compliance.sql` | audit, DLP, hold, discovery |
| `20260913091301_create_search_and_templates.sql` | FTS |
| `20260913091302_create_instance_settings.sql` | self-host admin |
| `20260913092637_create_runtime_support.sql` | formula jobs, Yjs registry, trash, calendars, outbox |

Cycle FKs (same-file `ALTER TABLE` only): `databases.primary_field_id`, `nodes.current_snapshot_id`, `nodes.current_json_snapshot_id`, `api_tokens.workspace_id`.

---

## 6. Table reference

Columns listed are the full stored set. `DEFAULT` and nullability match the SQL.

### 6.1 Identity

**users** — account. PK `id`. Unique `lower(email)` (not deleted), `lower(username)`.  
email, password_hash, password_algo, password_params, display_name, given_name, family_name, username, locale, timezone, week_starts_on, date_format, time_format, theme, avatar_object_key, bio, email_verified_at, password_updated_at, last_login_at, last_login_ip, last_seen_at, failed_login_count, locked_until, mfa_required, mfa_enrolled_at, tos_accepted_at, privacy_accepted_at, disabled_at, disable_reason, deleted_at, created_at, updated_at.

**user_preferences** — PK `user_id` → users. editor, notifications, accessibility, shortcuts, extras jsonb, updated_at.

**password_history** — reuse prevention. user_id, password_hash, password_algo, created_at.

**mfa_factors** — TOTP (and typed factors). user_id, type, name, secret_ciphertext, secret_nonce, digits, period_seconds, verified_at, last_used_at, disabled_at, timestamps.

**mfa_recovery_codes** — hashed one-time codes. Unique code_hash.

**user_devices** — trusted devices. fingerprint_hash unique per user while active.

**api_tokens** — hashed PAT. token_prefix, token_hash unique, scopes jsonb, optional workspace_id, expiry, revoke.

**webauthn_credentials** — passkeys. Unique credential_id bytea, public_key, sign_count, transports[], aaguid, backup flags.

**sessions** — hashed cookie session. device_id, csrf_hash, expires_at, idle_timeout_at, last_seen_at, revoked_at, revoke_reason, user_agent, ip, country.

**auth_tokens** — email verify / password reset. purpose, token_hash unique, expires, consumed, revoked.

**login_attempts** — brute-force log (email even if unknown user). success, failure_reason, ip.

**auth_events** — security event stream. event_type, session_id, metadata jsonb.

### 6.2 Organization and tenant

**organizations** — optional org above workspaces. name, slug unique, locale, timezone, data_region, settings.

**org_units** — tree (`parent_id`) inside an organization.

**org_unit_members** — PK (org_unit_id, user_id).

**workspaces** — tenant. organization_id, org_unit_id, owner_user_id, name, slug, description, icon, color, locale, timezone, status, settings, feature_flags, storage_used_bytes, storage_quota_bytes, member_quota, guest_quota, data_region, retention_days, trash_retention_days, organization_name, organization_domain, billing_email, public_sharing_enabled, guest_access_enabled, suspended_at, suspend_reason, deleted_at, purged_at.

**spaces** — Notion-style teamspace. workspace_id, is_private, is_default.

**space_members** — PK (space_id, user_id), role.

**workspace_domains** — verified join domains. Unique lower(domain).

**workspace_roles** — custom roles. permissions jsonb, is_system. Unique (workspace_id, lower(name)).

**workspace_members** — unique (workspace_id, user_id). role text, role_id, seat_type, status, guest_expires_at, last_seen_at.

**workspace_usage_daily** — PK (workspace_id, day). storage, file_count, member_count, api_request_count, automation_run_count.

**workspace_groups** / **workspace_group_members** — named principals.

**workspace_invites** — hashed token, email, role, group, seat_type, expiry, accepted/revoked.

### 6.3 Nodes and graph

**nodes** — universal identity. workspace_id, space_id, parent_id, root_id, cover_node_id, created_by, last_edited_by, archived_by, deleted_by, locked_by, published_by, type, subtype, flavour, flavour_version, title, description, icon, icon_color, cover_position, rank, props, schema_version, content_format, is_inline, is_template, is_locked, locked_at, published_at, public_slug, version (OCC), comment_count, discussion_count, child_count, timestamps, archived_at, deleted_at, restore_until, purged_at. Later columns: current_snapshot_id, current_json_snapshot_id.

**block_flavours** — PK (flavour, version). node_type, schema jsonb, deprecated_at. Registry for editor block schemas.

**node_edges** — relations, connectors (graph), mentions, assignees, dependencies, embeds. from_id, to_id, field_id, inverse_edge_id, kind, rank, start_anchor, end_anchor, label, props. Unique relation (from_id, to_id, field_id) where kind = `relation`.

**node_closure** — ancestor_id, descendant_id, depth, workspace_id. Maintained in Go on move/create/delete.

**trash_items** — PK node_id. restore_until, original_parent_id, original_rank.

### 6.4 ACL and engagement

**node_permissions** — PK (node_id, principal_type, principal_id). user_id / group_id FKs, role, inherit, public_role, expires_at.

**permission_changes** — sharing audit (before/after jsonb).

**permission_templates** — reusable role+inherit.

**node_permission_invites** — page guest by email before account exists. token_hash unique.

**node_stars**, **node_pins** (rank), **node_watches** (include_children), **node_visits** (visit_count, last_visited_at).

### 6.5 Databases (Airtable-class)

**databases** — 1:1 node. is_inline, primary_field_id (FK added after fields), row_count, schema_version.

**fields** — catalog. type, options jsonb, default_value, is_primary, is_required, is_unique, is_computed, is_hidden, is_valid, is_reversed, prefers_single_record, formula, formula_ast, result_type, result_options, compute_error, last_computed_at, precision, currency, duration_format, rating_max, linked_database_id, inverse_field_id, relation_field_id, lookup_field_id, rollup_field_id, rollup_function, referenced_field_ids[], rank.

**field_options** — select/status choices. Unique (field_id, key).

**field_dependencies** — PK (field_id, depends_on_field_id, kind). Formula/lookup/rollup DAG.

**relation_constraints** — PK field_id. from_database_id, to_database_id, allow_multiple, symmetric.

**views** — filter/sorts/groups jsonb, visible/frozen field id arrays, column_widths, calendar/board/timeline/map field FKs, owner_user_id, is_default, is_personal, is_locked, row_height, color_rules.

**field_values** — PK (record_id, field_id). value jsonb, computed_value, computed_at. GIN on value.

**field_value_revisions** — cell history.

**unique_field_values** — PK (field_id, value_key) → record_id. Enforces `fields.is_unique` in SQL without CHECKs.

**formula_functions** — PK name. arity, result_type, enabled. Catalog the evaluator may call.

**formula_jobs** — durable recompute queue. field_id, optional record_id, status, run_at.

### 6.6 Canvas

**canvas_objects** — 1:1 object node. canvas_id, object_kind, x, y, width, height, rotation, z_index, locked, visible, opacity, style. GiST on `box(point(x,y), point(x+width,y+height))`.

**canvas_frames** — titled regions.

**canvas_cameras** — PK (user_id, canvas_id). last pan/zoom.

**canvas_connectors** — 1:1 connector node. start/end object, shape, snap_to, offsets, captions, style.

**canvas_connector_vertices** — PK (connector_id, seq). routed path points.

**canvas_tags** / **canvas_tag_items** — Miro-style tags.

**canvas_groups** / **canvas_group_items** — grouped objects.

### 6.7 Documents / CRDT

**yjs_documents** — 1:1 node. guid unique, collection, gc_enabled, schema_name, schema_version.

**document_updates** — Yjs update log. client_id, clock, payload bytea, payload_size.

**document_clients** — PK (node_id, client_id). last clock, last_seen_at.

**document_snapshots** — compacted binary + state_vector.

**document_snapshot_history** — labeled/expiring history.

**document_json_snapshots** — BlockSuite-style JSON tree (derived).

**document_compactions** — PK node_id. last_compacted_update_id, retained_update_count.

Go/Yjs still performs `applyUpdate` / encode. Postgres stores every durable input and output of that pipeline.

### 6.8 Comments

**discussions** — thread on a node. selection_anchor, resolved_at, comment_count.

**comments** — 1:1 comment node. discussion_id, parent_id, body, rich_text jsonb, body_format.

**comment_attachments**, **comment_reactions** PK (comment_id, user_id, emoji), **mentions**.

### 6.9 Files

**files** — 1:1 file node. bucket, object_key unique, original_name, mime_type, size_bytes, checksum, etag, status, scan_status, scanned_at, storage_class, encryption_key_id, thumbnail/preview keys, width, height, duration_ms, page_count.

**file_revisions** — immutable prior objects. Unique object_key.

**file_upload_sessions** — multipart. upload_id unique, expires_at.

**file_tombstones** — GC queue. Unique object_key, delete_after.

Bytes live in Garage. Postgres never stores file bodies.

### 6.10 Sharing

**share_links** — token_hash unique, role, password_hash, include_subtree, allow_duplicate, domain_allowlist[], max_uses, use_count, watermark, expires, last_accessed, revoked.

**share_link_accesses** — access audit.

**published_pages** — PK node_id. slug unique per workspace, indexed, unpublished_at.

### 6.11 Notifications and search

**notifications** — inbox. type, title, body, payload, read_at, archived_at.

**notification_preferences** — PK (user_id, workspace_id, channel, event_type).

**notification_digests** — unique (user, workspace, channel, period). next_run_at.

**activity_events** — user-facing feed (not compliance).

**search_documents** — PK node_id. language, title, body, document tsvector, title_vector, rank_boost, node_type. GIN FTS + trigram title. Go writes vectors on change.

**search_synonyms**, **search_aliases**, **saved_searches**.

**templates** — workspace or instance templates pointing at a node.

### 6.12 Automation, webhooks, jobs

**automations** — trigger_type, trigger/conditions jsonb, enabled, version, run stats.

**automation_versions**, **automation_steps** (rank, type, config), **automation_schedules** (cron, next_run_at), **automation_watches**, **automation_runs**, **automation_step_runs**.

**webhooks** — url, secret_hash, events[], failure_count.

**webhook_deliveries** — retry via next_attempt_at.

**jobs** — generic worker. type, status, payload, attempts, run_at, locked_by.

**email_outbox** — reliable mail.

**exports** / **imports** — artifact object_key + stats.

**event_outbox** — domain events for automations/webhooks (status, available_at).

**idempotency_keys** — PK (key, user_id). request_hash, cached response, expires_at.

### 6.13 Calendars and dashboards

**recurrence_rules** — rrule, dtstart, timezone, until_at, node_id.

**calendar_events** — start_at, end_at, all_day, recurrence_id. Indexed (workspace_id, start_at).

**dashboards** — 1:1 dashboard node. layout jsonb.

**dashboard_widgets** — type, config, grid x/y/width/height.

### 6.14 Collaboration recovery and limits

**presence_states** — PK (user_id, node_id). selection, cursor, viewport jsonb, last_seen_at. **Live** presence remains Redis; this row is reconnect/hydration.

**rate_limit_policies** — name, scope, key_template, limit_count, window_seconds. **Counters** remain Redis.

### 6.15 Compliance and instance

**audit_logs** — actor snapshot, request_id, ip, action, resource_type/id, before/after. workspace_id nullable for instance events.

**audit_log_archives** — cold export to object storage.

**retention_policies** — resource_type, retain_days, action.

**legal_holds** / **legal_hold_items**.

**classification_labels** / **node_labels**.

**dlp_policies** (matchers/actions jsonb) / **dlp_findings**.

**discovery_cases** / **discovery_custodians**.

**encryption_keys** — wrapped_key bytea, purpose, algorithm, rotation.

**instance_settings** — PK key, value jsonb.

**instance_admins** — PK user_id.

---

## 7. Integrity the database enforces

- Foreign keys with `ON DELETE CASCADE` or `SET NULL` as declared.
- Unique emails, usernames, slugs, token hashes, object keys, field names per database, membership pairs, WebAuthn credential_id, Yjs guid, unique cell keys.
- Partial unique indexes for soft-deleted users/workspaces and open share/public slugs.

Go must still: reject illegal enums; keep `node_closure` and search vectors current; enqueue `formula_jobs` and `event_outbox`; refuse cross-workspace IDs even if a client sends them.

---

## 8. What is not in PostgreSQL (on purpose)

| Concern | Where | Why |
| --- | --- | --- |
| Live cursors / typing | Redis | High-churn, not business truth |
| Rate-limit counters | Redis | Need atomic INCR with TTL |
| File bytes | Garage | Size and streaming |
| Formula **evaluation** | Go worker | Uses `formula_jobs` + `field_dependencies` |
| Yjs **applyUpdate** | Realtime process | Uses `document_updates` + `yjs_documents` |
| Canvas hit-testing | Client | Uses geometry + vertices + GiST |
| OAuth / SAML / SCIM | Nowhere | Product forbids third-party IdP |

Those last four rows are why a honest schema score is **94/100**, not 100. Persistence for engines is complete; the engines themselves are not SQL.

---

## 9. Typical write paths (application)

**Create a page:** insert `nodes`; insert self-row in `node_closure`; optional `yjs_documents` + empty snapshot; index `search_documents`.

**Move a node:** update `parent_id`/`rank`; rebuild `node_closure` for the subtree in the same transaction (see §11.2).

**Edit a cell:** update `field_values`; insert `field_value_revisions`; upsert `unique_field_values` if unique; enqueue `formula_jobs` for dependents; `event_outbox`.

**Upload a file:** `file_upload_sessions` → complete → `files` ready; bytes in Garage; `storage_used_bytes` / usage daily.

**Share a page:** `node_permissions` or `node_permission_invites` / `share_links`; `permission_changes` + `audit_logs`.

**Delete:** set `nodes.deleted_at` + `trash_items`; do not CASCADE purge until `restore_until` and no `legal_hold_items`.

---

## 10. Indexing notes

Hot paths expected:

- Workspace lists: `nodes (workspace_id) WHERE deleted_at IS NULL`
- Tree: `(parent_id, rank)`
- Recents: `(workspace_id, updated_at DESC)`
- ACL: `(principal_type, principal_id)`
- Record filter: GIN `field_values.value`
- Canvas viewport: GiST box on `canvas_objects`
- FTS: GIN `search_documents.document` and `title_vector`
- Jobs: `(status, run_at)` on `jobs`, `formula_jobs`, `event_outbox`

Do not add more indexes without `EXPLAIN ANALYZE` on a real workload (`PROJECT-INFO.md` §62).

---

## 11. Application maintenance contracts

These are **not** PostgreSQL triggers. Flow does not use triggers, trigger functions, or `CHECK` constraints. Go owns the following. If a writer skips them, Postgres will still accept the row.

### 11.1 `updated_at`

`DEFAULT now()` fires on **INSERT only**. Every `UPDATE` must set `updated_at` in SQL.

Do not add `set_updated_at()` triggers. They hide missed writes, override intentional timestamps (imports, CRDT apply time, tests), and violate the no-trigger rule.

`nodes.last_edited_at` is the human edit time. `nodes.updated_at` is any row mutation. Set both on user edits; set only `updated_at` for bookkeeping.

### 11.2 `node_closure`

`parent_id` is the source-of-truth edge. Closure is the materialized ancestor index.

**Self row:** every node has `(ancestor_id = id, descendant_id = id, depth = 0)`.

**Create** node N under parent P, same transaction as `INSERT nodes`:

1. Insert the self row for N.
2. Copy P’s ancestors onto N:  
   `INSERT INTO node_closure (ancestor_id, descendant_id, workspace_id, depth)`  
   `SELECT ancestor_id, N, workspace_id, depth + 1 FROM node_closure WHERE descendant_id = P`.

**Move** subtree N to parent P′:

1. Let D = descendants of N including N.
2. Delete closure rows where `descendant_id IN D` and `ancestor_id NOT IN D`.
3. If P′ is set, insert every ancestor of P′ (including P′) crossed with every D, with depths added.
4. Update `nodes.parent_id` and `root_id` if the tree root changed.

**Hard delete:** `ON DELETE CASCADE` from `nodes`. **Soft delete** leaves closure intact so trash restore still has a parent.

**Test invariant:** if `parent_id = P`, closure contains `(P, N, depth = 1)` and every ancestor of P is an ancestor of N at `depth + 1`.

### 11.3 `search_documents` tsvectors

Go writes `title`, `body`, `language`, then:

```text
title_vector = to_tsvector(language::regconfig, title)
document     = to_tsvector(language::regconfig, title || ' ' || body)
```

`language` must be a real Postgres text-search config (`simple`, `english`, …). No tsvector trigger.

### 11.4 `block_flavours` vs `nodes.flavour`

`nodes.flavour` + `flavour_version` name the editor schema. `block_flavours` holds the JSON schema. Go inserts registry rows before production nodes use that flavour. Not an FK so experiments do not need a migration.

### 11.5 High-growth tables (partitioning)

Do **not** range-partition `document_updates` while `document_snapshots.last_update_id` and compaction reference `document_updates(id)`. Partitioned unique constraints must include the partition key, which breaks a simple `id` FK.

Archive instead: `audit_log_archives`, snapshot `expires_at`, `document_compactions`, `file_tombstones`. When `audit_logs` / `login_attempts` / `activity_events` are large, partition **those** (no inbound FKs on `id`) by `RANGE (created_at)`, or export to object storage.

### 11.6 Status `text` columns

No `CHECK`. Go validates `files.status`, `workspace_members.seat_type`, `jobs.status`, and the rest. Invalid values are application bugs.

---

## 12. Related documents

- `PROJECT-INFO.md` — product and architecture
- `DEPENDENCIES.md` — pgx, goose, no ORM, email+password only
- Competitive review canvas (schema scores)

When the SQL changes, update this file in the same change.
