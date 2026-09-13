package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kashifxyz/flow-server/internal/access"
	"github.com/kashifxyz/flow-server/internal/modules/collaboration/models"
	"github.com/kashifxyz/flow-server/internal/utils/httperr"
)

type Service struct {
	DB *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Service {
	return &Service{DB: db}
}

// recencyWindow limits presence rows to viewers active in the last few minutes.
// presence_states is reconnect/hydration data, not the live pub/sub fabric (SCHEMA.md §6.14).
const recencyWindow = `5 minutes`

func (s *Service) ListByWorkspace(ctx context.Context, workspaceID, userID uuid.UUID) ([]models.Presence, error) {
	if err := access.RequireMember(ctx, s.DB, workspaceID, userID); err != nil {
		return nil, err
	}
	rows, err := s.DB.Query(ctx, `
		SELECT user_id, node_id, selection, cursor, viewport, last_seen_at
		FROM presence_states
		WHERE workspace_id = $1 AND last_seen_at > now() - interval '`+recencyWindow+`'
		ORDER BY last_seen_at DESC
		LIMIT 200
	`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collect(rows)
}

func (s *Service) ListByNode(ctx context.Context, nodeID, userID uuid.UUID) ([]models.Presence, error) {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, nodeID); err != nil {
		return nil, err
	}
	rows, err := s.DB.Query(ctx, `
		SELECT user_id, node_id, selection, cursor, viewport, last_seen_at
		FROM presence_states
		WHERE node_id = $1 AND last_seen_at > now() - interval '`+recencyWindow+`'
		ORDER BY last_seen_at DESC
	`, nodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collect(rows)
}

func (s *Service) Put(ctx context.Context, nodeID, userID uuid.UUID, in models.PutPresenceRequest) (models.Presence, error) {
	workspaceID, err := access.RequireNodeMember(ctx, s.DB, userID, nodeID)
	if err != nil {
		return models.Presence{}, err
	}
	var p models.Presence
	var uID, nID uuid.UUID
	err = s.DB.QueryRow(ctx, `
		INSERT INTO presence_states (user_id, node_id, workspace_id, selection, cursor, viewport, last_seen_at)
		VALUES ($1, $2, $3, $4, $5, $6, now())
		ON CONFLICT (user_id, node_id) DO UPDATE SET
			selection = EXCLUDED.selection, cursor = EXCLUDED.cursor, viewport = EXCLUDED.viewport, last_seen_at = now()
		RETURNING user_id, node_id, selection, cursor, viewport, last_seen_at
	`, userID, nodeID, workspaceID, in.Selection, in.Cursor, in.Viewport).Scan(&uID, &nID, &p.Selection, &p.Cursor, &p.Viewport, &p.LastSeenAt)
	if err != nil {
		return models.Presence{}, err
	}
	p.UserID = uID.String()
	p.NodeID = nID.String()
	return p, nil
}

func (s *Service) Delete(ctx context.Context, nodeID, userID uuid.UUID) error {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, nodeID); err != nil {
		return err
	}
	tag, err := s.DB.Exec(ctx, `DELETE FROM presence_states WHERE user_id = $1 AND node_id = $2`, userID, nodeID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return httperr.ErrNotFound
	}
	return nil
}

func collect(rows pgx.Rows) ([]models.Presence, error) {
	out := make([]models.Presence, 0)
	for rows.Next() {
		var p models.Presence
		var userID, nodeID uuid.UUID
		if err := rows.Scan(&userID, &nodeID, &p.Selection, &p.Cursor, &p.Viewport, &p.LastSeenAt); err != nil {
			return nil, err
		}
		p.UserID = userID.String()
		p.NodeID = nodeID.String()
		out = append(out, p)
	}
	return out, rows.Err()
}
