package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// EnsureDocument creates the yjs_documents row for a node on first use.
// Idempotent — safe to call on every connection open.
func (s *Service) EnsureDocument(ctx context.Context, nodeID, workspaceID uuid.UUID) error {
	_, err := s.DB.Exec(ctx, `
		INSERT INTO yjs_documents (node_id, workspace_id, guid)
		VALUES ($1, $2, $1::text)
		ON CONFLICT (node_id) DO NOTHING
	`, nodeID, workspaceID)
	return err
}

// LoadSync returns the latest compacted snapshot (if any) plus every update
// recorded after it, in order — everything a joining client needs to
// reconstruct current state. There is no compaction job yet (SCHEMA.md
// §11.5), so in practice this is usually "no snapshot, all updates".
func (s *Service) LoadSync(ctx context.Context, nodeID uuid.UUID) (snapshot []byte, updates [][]byte, err error) {
	var snapshotID *uuid.UUID
	err = s.DB.QueryRow(ctx, `
		SELECT id, payload FROM document_snapshots
		WHERE node_id = $1 ORDER BY created_at DESC LIMIT 1
	`, nodeID).Scan(&snapshotID, &snapshot)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, nil, err
	}
	err = nil

	query := `SELECT payload FROM document_updates WHERE node_id = $1`
	args := []any{nodeID}
	if snapshotID != nil {
		query += ` AND id > (SELECT last_update_id FROM document_snapshots WHERE id = $2)`
		args = append(args, *snapshotID)
	}
	query += ` ORDER BY id`
	rows, err := s.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	updates = make([][]byte, 0)
	for rows.Next() {
		var payload []byte
		if err := rows.Scan(&payload); err != nil {
			return nil, nil, err
		}
		updates = append(updates, payload)
	}
	return snapshot, updates, rows.Err()
}

// AppendUpdate persists one Yjs update and advances the sending client's
// server-side clock. It does not interpret the update's contents — merging
// happens client-side in Yjs (SCHEMA.md §8).
func (s *Service) AppendUpdate(ctx context.Context, nodeID, workspaceID, userID uuid.UUID, clientID string, payload []byte) error {
	var clock int64
	err := s.DB.QueryRow(ctx, `
		INSERT INTO document_clients (node_id, client_id, workspace_id, user_id, clock, last_seen_at)
		VALUES ($1, $2, $3, $4, 1, now())
		ON CONFLICT (node_id, client_id) DO UPDATE SET
			clock = document_clients.clock + 1, last_seen_at = now(), user_id = EXCLUDED.user_id
		RETURNING clock
	`, nodeID, clientID, workspaceID, userID).Scan(&clock)
	if err != nil {
		return err
	}
	_, err = s.DB.Exec(ctx, `
		INSERT INTO document_updates (workspace_id, node_id, client_id, clock, payload, payload_size, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, workspaceID, nodeID, clientID, clock, payload, len(payload), userID)
	return err
}
