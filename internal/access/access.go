package access

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kashifxyz/flow-server/internal/httperr"
)

func RequireMember(ctx context.Context, db *pgxpool.Pool, workspaceID, userID uuid.UUID) error {
	var ok bool
	err := db.QueryRow(ctx, `
		SELECT true
		FROM workspace_members
		WHERE workspace_id = $1 AND user_id = $2 AND status = 'active' AND deactivated_at IS NULL
	`, workspaceID, userID).Scan(&ok)
	if errors.Is(err, pgx.ErrNoRows) {
		return httperr.ErrForbidden
	}
	return err
}

func WorkspaceIDForNode(ctx context.Context, db *pgxpool.Pool, nodeID uuid.UUID) (uuid.UUID, error) {
	var workspaceID uuid.UUID
	err := db.QueryRow(ctx, `
		SELECT workspace_id FROM nodes WHERE id = $1 AND purged_at IS NULL
	`, nodeID).Scan(&workspaceID)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, httperr.ErrNotFound
	}
	return workspaceID, err
}

func RequireNodeMember(ctx context.Context, db *pgxpool.Pool, userID, nodeID uuid.UUID) (uuid.UUID, error) {
	workspaceID, err := WorkspaceIDForNode(ctx, db, nodeID)
	if err != nil {
		return uuid.Nil, err
	}
	if err := RequireMember(ctx, db, workspaceID, userID); err != nil {
		return uuid.Nil, err
	}
	return workspaceID, nil
}
