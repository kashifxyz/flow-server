package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kashifxyz/flow-server/internal/access"
	"github.com/kashifxyz/flow-server/internal/graph"
	"github.com/kashifxyz/flow-server/internal/modules/node/models"
	"github.com/kashifxyz/flow-server/internal/utils/httperr"
)

type Service struct {
	DB *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Service {
	return &Service{DB: db}
}

var permissionRoles = map[string]bool{"view": true, "comment": true, "edit": true, "manage": true}
var principalTypes = map[string]bool{"user": true, "group": true, "public": true}

func (s *Service) Get(ctx context.Context, nodeID, userID uuid.UUID) (models.Node, error) {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, nodeID); err != nil {
		return models.Node{}, err
	}
	n, err := graph.Get(ctx, s.DB, nodeID)
	if err != nil {
		return models.Node{}, err
	}
	return models.FromGraph(n), nil
}

func (s *Service) Update(ctx context.Context, nodeID, userID uuid.UUID, in models.UpdateRequest) (models.Node, error) {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, nodeID); err != nil {
		return models.Node{}, err
	}
	patch := graph.Patch{
		Title:       in.Title,
		Description: in.Description,
		Icon:        in.Icon,
		Props:       in.Props,
		Rank:        in.Rank,
	}
	if in.SpaceID != nil {
		id, err := uuid.Parse(*in.SpaceID)
		if err != nil {
			return models.Node{}, httperr.ErrInvalid
		}
		patch.SpaceID = &id
	}
	n, err := graph.Update(ctx, s.DB, nodeID, userID, patch)
	if err != nil {
		return models.Node{}, err
	}
	return models.FromGraph(n), nil
}

func (s *Service) Delete(ctx context.Context, nodeID, userID uuid.UUID) error {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, nodeID); err != nil {
		return err
	}
	return withTx(ctx, s.DB, func(tx pgx.Tx) error {
		return graph.SoftDelete(ctx, tx, nodeID, userID, 0)
	})
}

func (s *Service) Restore(ctx context.Context, nodeID, userID uuid.UUID) error {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, nodeID); err != nil {
		return err
	}
	return withTx(ctx, s.DB, func(tx pgx.Tx) error {
		return graph.Restore(ctx, tx, nodeID)
	})
}

func (s *Service) Archive(ctx context.Context, nodeID, userID uuid.UUID) error {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, nodeID); err != nil {
		return err
	}
	return graph.Archive(ctx, s.DB, nodeID, userID)
}

func (s *Service) Unarchive(ctx context.Context, nodeID, userID uuid.UUID) error {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, nodeID); err != nil {
		return err
	}
	return graph.Unarchive(ctx, s.DB, nodeID)
}

func (s *Service) Move(ctx context.Context, nodeID, userID uuid.UUID, in models.MoveRequest) error {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, nodeID); err != nil {
		return err
	}
	var parentID *uuid.UUID
	if in.ParentID != nil && *in.ParentID != "" {
		id, err := uuid.Parse(*in.ParentID)
		if err != nil {
			return httperr.ErrInvalid
		}
		parentID = &id
	}
	return withTx(ctx, s.DB, func(tx pgx.Tx) error {
		return graph.Move(ctx, tx, nodeID, parentID, in.Rank)
	})
}

func (s *Service) ListTrash(ctx context.Context, workspaceID, userID uuid.UUID) ([]models.TrashItem, error) {
	if err := access.RequireMember(ctx, s.DB, workspaceID, userID); err != nil {
		return nil, err
	}
	rows, err := s.DB.Query(ctx, `
		SELECT t.node_id, n.title, n.type, t.deleted_at, t.restore_until, t.original_parent_id
		FROM trash_items t JOIN nodes n ON n.id = t.node_id
		WHERE t.workspace_id = $1
		ORDER BY t.deleted_at DESC
	`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.TrashItem, 0)
	for rows.Next() {
		var it models.TrashItem
		var nodeID uuid.UUID
		var parentID *uuid.UUID
		if err := rows.Scan(&nodeID, &it.Title, &it.Type, &it.DeletedAt, &it.RestoreUntil, &parentID); err != nil {
			return nil, err
		}
		it.NodeID = nodeID.String()
		if parentID != nil {
			s := parentID.String()
			it.OriginalParentID = &s
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func (s *Service) ListPermissions(ctx context.Context, nodeID, userID uuid.UUID) ([]models.Permission, error) {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, nodeID); err != nil {
		return nil, err
	}
	rows, err := s.DB.Query(ctx, `
		SELECT principal_type, principal_id, role, inherit, public_role, expires_at
		FROM node_permissions WHERE node_id = $1
	`, nodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.Permission, 0)
	for rows.Next() {
		var p models.Permission
		var principalID uuid.UUID
		if err := rows.Scan(&p.PrincipalType, &principalID, &p.Role, &p.Inherit, &p.PublicRole, &p.ExpiresAt); err != nil {
			return nil, err
		}
		p.PrincipalID = principalID.String()
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *Service) SetPermissions(ctx context.Context, nodeID, userID uuid.UUID, in models.SetPermissionsRequest) ([]models.Permission, error) {
	workspaceID, err := access.RequireNodeMember(ctx, s.DB, userID, nodeID)
	if err != nil {
		return nil, err
	}
	if err := s.requireManage(ctx, nodeID, workspaceID, userID); err != nil {
		return nil, err
	}
	for _, p := range in.Permissions {
		if !principalTypes[p.PrincipalType] || !permissionRoles[p.Role] {
			return nil, httperr.ErrInvalid
		}
		if _, err := uuid.Parse(p.PrincipalID); err != nil {
			return nil, httperr.ErrInvalid
		}
	}
	err = withTx(ctx, s.DB, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `DELETE FROM node_permissions WHERE node_id = $1`, nodeID); err != nil {
			return err
		}
		for _, p := range in.Permissions {
			principalID := uuid.MustParse(p.PrincipalID)
			inherit := true
			if p.Inherit != nil {
				inherit = *p.Inherit
			}
			var uid, gid *uuid.UUID
			switch p.PrincipalType {
			case "user":
				uid = &principalID
			case "group":
				gid = &principalID
			}
			if _, err := tx.Exec(ctx, `
				INSERT INTO node_permissions (node_id, workspace_id, principal_type, principal_id, user_id, group_id, role, inherit, expires_at, created_by)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			`, nodeID, workspaceID, p.PrincipalType, principalID, uid, gid, p.Role, inherit, p.ExpiresAt, userID); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.ListPermissions(ctx, nodeID, userID)
}

func (s *Service) Star(ctx context.Context, nodeID, userID uuid.UUID) error {
	workspaceID, err := access.RequireNodeMember(ctx, s.DB, userID, nodeID)
	if err != nil {
		return err
	}
	_, err = s.DB.Exec(ctx, `
		INSERT INTO node_stars (user_id, node_id, workspace_id) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING
	`, userID, nodeID, workspaceID)
	return err
}

func (s *Service) Unstar(ctx context.Context, nodeID, userID uuid.UUID) error {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, nodeID); err != nil {
		return err
	}
	_, err := s.DB.Exec(ctx, `DELETE FROM node_stars WHERE user_id = $1 AND node_id = $2`, userID, nodeID)
	return err
}

func (s *Service) Pin(ctx context.Context, nodeID, userID uuid.UUID) error {
	workspaceID, err := access.RequireNodeMember(ctx, s.DB, userID, nodeID)
	if err != nil {
		return err
	}
	_, err = s.DB.Exec(ctx, `
		INSERT INTO node_pins (user_id, node_id, workspace_id) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING
	`, userID, nodeID, workspaceID)
	return err
}

func (s *Service) Unpin(ctx context.Context, nodeID, userID uuid.UUID) error {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, nodeID); err != nil {
		return err
	}
	_, err := s.DB.Exec(ctx, `DELETE FROM node_pins WHERE user_id = $1 AND node_id = $2`, userID, nodeID)
	return err
}

func (s *Service) Watch(ctx context.Context, nodeID, userID uuid.UUID) error {
	workspaceID, err := access.RequireNodeMember(ctx, s.DB, userID, nodeID)
	if err != nil {
		return err
	}
	_, err = s.DB.Exec(ctx, `
		INSERT INTO node_watches (user_id, node_id, workspace_id) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING
	`, userID, nodeID, workspaceID)
	return err
}

func (s *Service) Unwatch(ctx context.Context, nodeID, userID uuid.UUID) error {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, nodeID); err != nil {
		return err
	}
	_, err := s.DB.Exec(ctx, `DELETE FROM node_watches WHERE user_id = $1 AND node_id = $2`, userID, nodeID)
	return err
}

func (s *Service) Visit(ctx context.Context, nodeID, userID uuid.UUID) error {
	workspaceID, err := access.RequireNodeMember(ctx, s.DB, userID, nodeID)
	if err != nil {
		return err
	}
	_, err = s.DB.Exec(ctx, `
		INSERT INTO node_visits (user_id, node_id, workspace_id, visit_count, last_visited_at)
		VALUES ($1, $2, $3, 1, now())
		ON CONFLICT (user_id, node_id) DO UPDATE SET visit_count = node_visits.visit_count + 1, last_visited_at = now()
	`, userID, nodeID, workspaceID)
	return err
}

func (s *Service) requireManage(ctx context.Context, nodeID, workspaceID, userID uuid.UUID) error {
	var ok bool
	err := s.DB.QueryRow(ctx, `
		SELECT true WHERE
			EXISTS (SELECT 1 FROM workspace_members WHERE workspace_id = $1 AND user_id = $2 AND status = 'active' AND role IN ('owner','admin'))
			OR EXISTS (SELECT 1 FROM node_permissions WHERE node_id = $3 AND principal_type = 'user' AND principal_id = $2 AND role = 'manage')
	`, workspaceID, userID, nodeID).Scan(&ok)
	if errors.Is(err, pgx.ErrNoRows) {
		return httperr.ErrForbidden
	}
	return err
}

func withTx(ctx context.Context, db *pgxpool.Pool, fn func(pgx.Tx) error) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
