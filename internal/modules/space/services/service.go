package services

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kashifxyz/flow-server/internal/access"
	"github.com/kashifxyz/flow-server/internal/modules/space/models"
	"github.com/kashifxyz/flow-server/internal/utils"
	"github.com/kashifxyz/flow-server/internal/utils/httperr"
)

type Service struct {
	DB *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Service {
	return &Service{DB: db}
}

var spaceMemberRoles = map[string]bool{"member": true, "admin": true}

const spaceCols = `id, workspace_id, name, description, icon, is_private, is_default, created_at, updated_at`

func (s *Service) List(ctx context.Context, workspaceID, userID uuid.UUID) ([]models.Space, error) {
	if err := access.RequireMember(ctx, s.DB, workspaceID, userID); err != nil {
		return nil, err
	}
	rows, err := s.DB.Query(ctx, `
		SELECT `+spaceCols+` FROM spaces
		WHERE workspace_id = $1 AND deleted_at IS NULL
		AND (is_private = false OR id IN (SELECT space_id FROM space_members WHERE user_id = $2))
		ORDER BY is_default DESC, created_at
	`, workspaceID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.Space, 0)
	for rows.Next() {
		sp, err := scanSpace(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, sp)
	}
	return out, rows.Err()
}

func (s *Service) Create(ctx context.Context, workspaceID, userID uuid.UUID, in models.CreateSpaceRequest) (models.Space, error) {
	if err := access.RequireMember(ctx, s.DB, workspaceID, userID); err != nil {
		return models.Space{}, err
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return models.Space{}, httperr.ErrInvalid
	}
	id, err := utils.NewID()
	if err != nil {
		return models.Space{}, err
	}
	sp, err := scanSpace(s.DB.QueryRow(ctx, `
		INSERT INTO spaces (id, workspace_id, name, description, icon, is_private, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING `+spaceCols, id, workspaceID, name, in.Description, nullIfEmpty(in.Icon), in.IsPrivate, userID))
	if err != nil {
		return models.Space{}, err
	}
	if _, err := s.DB.Exec(ctx, `
		INSERT INTO space_members (space_id, user_id, role) VALUES ($1, $2, 'admin')
	`, id, userID); err != nil {
		return models.Space{}, err
	}
	return sp, nil
}

func (s *Service) Get(ctx context.Context, spaceID, userID uuid.UUID) (models.Space, error) {
	sp, err := s.getSpace(ctx, spaceID)
	if err != nil {
		return models.Space{}, err
	}
	if err := access.RequireMember(ctx, s.DB, uuid.MustParse(sp.WorkspaceID), userID); err != nil {
		return models.Space{}, err
	}
	return sp, nil
}

func (s *Service) Update(ctx context.Context, spaceID, userID uuid.UUID, in models.UpdateSpaceRequest) (models.Space, error) {
	if err := s.requireManage(ctx, spaceID, userID); err != nil {
		return models.Space{}, err
	}
	sp, err := scanSpace(s.DB.QueryRow(ctx, `
		UPDATE spaces SET
			name = COALESCE($2, name),
			description = COALESCE($3, description),
			icon = COALESCE($4, icon),
			is_private = COALESCE($5, is_private),
			updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING `+spaceCols, spaceID, in.Name, in.Description, in.Icon, in.IsPrivate))
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Space{}, httperr.ErrNotFound
	}
	return sp, err
}

func (s *Service) Delete(ctx context.Context, spaceID, userID uuid.UUID) error {
	if err := s.requireManage(ctx, spaceID, userID); err != nil {
		return err
	}
	tag, err := s.DB.Exec(ctx, `
		UPDATE spaces SET deleted_at = now(), updated_at = now() WHERE id = $1 AND deleted_at IS NULL AND is_default = false
	`, spaceID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return httperr.ErrNotFound
	}
	return nil
}

func (s *Service) ListMembers(ctx context.Context, spaceID, userID uuid.UUID) ([]models.Member, error) {
	sp, err := s.getSpace(ctx, spaceID)
	if err != nil {
		return nil, err
	}
	if err := access.RequireMember(ctx, s.DB, uuid.MustParse(sp.WorkspaceID), userID); err != nil {
		return nil, err
	}
	rows, err := s.DB.Query(ctx, `
		SELECT sm.user_id, u.email, u.display_name, sm.role, sm.created_at
		FROM space_members sm JOIN users u ON u.id = sm.user_id
		WHERE sm.space_id = $1 ORDER BY sm.created_at
	`, spaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.Member, 0)
	for rows.Next() {
		var m models.Member
		if err := rows.Scan(&m.UserID, &m.Email, &m.DisplayName, &m.Role, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (s *Service) AddMember(ctx context.Context, spaceID, userID uuid.UUID, in models.AddMemberRequest) (models.Member, error) {
	if err := s.requireManage(ctx, spaceID, userID); err != nil {
		return models.Member{}, err
	}
	targetID, err := uuid.Parse(in.UserID)
	if err != nil {
		return models.Member{}, httperr.ErrInvalid
	}
	role := in.Role
	if role == "" {
		role = "member"
	}
	if !spaceMemberRoles[role] {
		return models.Member{}, httperr.ErrInvalid
	}
	if _, err := s.DB.Exec(ctx, `
		INSERT INTO space_members (space_id, user_id, role) VALUES ($1, $2, $3)
		ON CONFLICT (space_id, user_id) DO UPDATE SET role = EXCLUDED.role
	`, spaceID, targetID, role); err != nil {
		return models.Member{}, err
	}
	var m models.Member
	err = s.DB.QueryRow(ctx, `
		SELECT sm.user_id, u.email, u.display_name, sm.role, sm.created_at
		FROM space_members sm JOIN users u ON u.id = sm.user_id
		WHERE sm.space_id = $1 AND sm.user_id = $2
	`, spaceID, targetID).Scan(&m.UserID, &m.Email, &m.DisplayName, &m.Role, &m.CreatedAt)
	return m, err
}

func (s *Service) UpdateMember(ctx context.Context, spaceID, targetUserID, userID uuid.UUID, in models.UpdateMemberRequest) (models.Member, error) {
	if err := s.requireManage(ctx, spaceID, userID); err != nil {
		return models.Member{}, err
	}
	if !spaceMemberRoles[in.Role] {
		return models.Member{}, httperr.ErrInvalid
	}
	tag, err := s.DB.Exec(ctx, `UPDATE space_members SET role = $3 WHERE space_id = $1 AND user_id = $2`, spaceID, targetUserID, in.Role)
	if err != nil {
		return models.Member{}, err
	}
	if tag.RowsAffected() == 0 {
		return models.Member{}, httperr.ErrNotFound
	}
	var m models.Member
	err = s.DB.QueryRow(ctx, `
		SELECT sm.user_id, u.email, u.display_name, sm.role, sm.created_at
		FROM space_members sm JOIN users u ON u.id = sm.user_id
		WHERE sm.space_id = $1 AND sm.user_id = $2
	`, spaceID, targetUserID).Scan(&m.UserID, &m.Email, &m.DisplayName, &m.Role, &m.CreatedAt)
	return m, err
}

func (s *Service) RemoveMember(ctx context.Context, spaceID, targetUserID, userID uuid.UUID) error {
	if err := s.requireManage(ctx, spaceID, userID); err != nil {
		return err
	}
	tag, err := s.DB.Exec(ctx, `DELETE FROM space_members WHERE space_id = $1 AND user_id = $2`, spaceID, targetUserID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return httperr.ErrNotFound
	}
	return nil
}

func (s *Service) getSpace(ctx context.Context, spaceID uuid.UUID) (models.Space, error) {
	sp, err := scanSpace(s.DB.QueryRow(ctx, `SELECT `+spaceCols+` FROM spaces WHERE id = $1 AND deleted_at IS NULL`, spaceID))
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Space{}, httperr.ErrNotFound
	}
	return sp, err
}

func (s *Service) requireManage(ctx context.Context, spaceID, userID uuid.UUID) error {
	sp, err := s.getSpace(ctx, spaceID)
	if err != nil {
		return err
	}
	var ok bool
	err = s.DB.QueryRow(ctx, `
		SELECT true WHERE
			EXISTS (SELECT 1 FROM space_members WHERE space_id = $1 AND user_id = $2 AND role = 'admin')
			OR EXISTS (SELECT 1 FROM workspace_members WHERE workspace_id = $3 AND user_id = $2 AND status = 'active' AND role IN ('owner','admin'))
	`, spaceID, userID, sp.WorkspaceID).Scan(&ok)
	if errors.Is(err, pgx.ErrNoRows) {
		return httperr.ErrForbidden
	}
	return err
}

func scanSpace(row pgx.Row) (models.Space, error) {
	var sp models.Space
	var id, workspaceID uuid.UUID
	var icon *string
	err := row.Scan(&id, &workspaceID, &sp.Name, &sp.Description, &icon, &sp.IsPrivate, &sp.IsDefault, &sp.CreatedAt, &sp.UpdatedAt)
	if err != nil {
		return models.Space{}, err
	}
	sp.ID = id.String()
	sp.WorkspaceID = workspaceID.String()
	sp.Icon = icon
	return sp, nil
}

func nullIfEmpty(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}
