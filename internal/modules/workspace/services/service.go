package services

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kashifxyz/flow-server/internal/auth"
	"github.com/kashifxyz/flow-server/internal/modules/workspace/models"
	"github.com/kashifxyz/flow-server/internal/utils"
	"github.com/kashifxyz/flow-server/internal/utils/httperr"
	"github.com/kashifxyz/flow-server/pkg/mail"
)

type Service struct {
	DB        *pgxpool.Pool
	Mail      mail.Sender
	PublicURL string
}

func New(db *pgxpool.Pool, sender mail.Sender, publicURL string) *Service {
	return &Service{DB: db, Mail: sender, PublicURL: publicURL}
}

var memberRoles = map[string]bool{"owner": true, "admin": true, "member": true, "guest": true}
var seatTypes = map[string]bool{"full": true, "guest": true}
var memberStatuses = map[string]bool{"active": true, "suspended": true, "removed": true}

const defaultInviteTTL = 7 * 24 * time.Hour

// ---- workspaces ----

func (s *Service) Create(ctx context.Context, ownerID uuid.UUID, in models.CreateWorkspaceRequest) (models.Workspace, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return models.Workspace{}, httperr.ErrInvalid
	}
	id, err := utils.NewID()
	if err != nil {
		return models.Workspace{}, err
	}
	var w models.Workspace
	err = withTx(ctx, s.DB, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, `
			INSERT INTO workspaces (id, created_by, owner_user_id, name, slug, description, icon, color)
			VALUES ($1, $2, $2, $3, $4, $5, $6, $7)
			RETURNING `+workspaceCols, id, ownerID, name, nullIfEmpty(in.Slug), in.Description,
			nullIfEmpty(in.Icon), nullIfEmpty(in.Color))
		var scanErr error
		w, scanErr = scanWorkspace(row)
		if scanErr != nil {
			return scanErr
		}
		_, err := tx.Exec(ctx, `
			INSERT INTO workspace_members (id, workspace_id, user_id, role, status, joined_at)
			VALUES ($1, $2, $3, 'owner', 'active', now())
		`, utils.MustID(), id, ownerID)
		return err
	})
	return w, err
}

func (s *Service) List(ctx context.Context, userID uuid.UUID) ([]models.Workspace, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT `+workspaceColsW+`
		FROM workspaces w
		JOIN workspace_members m ON m.workspace_id = w.id
		WHERE m.user_id = $1 AND m.status = 'active' AND w.deleted_at IS NULL
		ORDER BY w.created_at
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.Workspace, 0)
	for rows.Next() {
		w, err := scanWorkspace(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, rows.Err()
}

func (s *Service) Get(ctx context.Context, workspaceID, userID uuid.UUID) (models.Workspace, error) {
	if err := s.requireMember(ctx, workspaceID, userID); err != nil {
		return models.Workspace{}, err
	}
	return s.getWorkspace(ctx, workspaceID)
}

func (s *Service) Update(ctx context.Context, workspaceID, userID uuid.UUID, in models.UpdateWorkspaceRequest) (models.Workspace, error) {
	if err := s.requireManage(ctx, workspaceID, userID); err != nil {
		return models.Workspace{}, err
	}
	w, err := scanWorkspace(s.DB.QueryRow(ctx, `
		UPDATE workspaces SET
			name = COALESCE($2, name),
			description = COALESCE($3, description),
			icon = COALESCE($4, icon),
			color = COALESCE($5, color),
			public_sharing_enabled = COALESCE($6, public_sharing_enabled),
			guest_access_enabled = COALESCE($7, guest_access_enabled),
			updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING `+workspaceCols, workspaceID, in.Name, in.Description, in.Icon, in.Color,
		in.PublicSharingEnabled, in.GuestAccessEnabled))
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Workspace{}, httperr.ErrNotFound
	}
	return w, err
}

func (s *Service) Delete(ctx context.Context, workspaceID, userID uuid.UUID) error {
	if err := s.requireRole(ctx, workspaceID, userID, "owner"); err != nil {
		return err
	}
	tag, err := s.DB.Exec(ctx, `
		UPDATE workspaces SET deleted_at = now(), updated_at = now() WHERE id = $1 AND deleted_at IS NULL
	`, workspaceID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return httperr.ErrNotFound
	}
	return nil
}

func (s *Service) Usage(ctx context.Context, workspaceID, userID uuid.UUID) (models.Usage, error) {
	if err := s.requireMember(ctx, workspaceID, userID); err != nil {
		return models.Usage{}, err
	}
	var u models.Usage
	err := s.DB.QueryRow(ctx, `
		SELECT storage_used_bytes, storage_quota_bytes, member_quota, guest_quota
		FROM workspaces WHERE id = $1 AND deleted_at IS NULL
	`, workspaceID).Scan(&u.StorageUsedBytes, &u.StorageQuotaBytes, &u.MemberQuota, &u.GuestQuota)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Usage{}, httperr.ErrNotFound
	}
	if err != nil {
		return models.Usage{}, err
	}
	err = s.DB.QueryRow(ctx, `
		SELECT
			count(*) FILTER (WHERE seat_type = 'full'),
			count(*) FILTER (WHERE seat_type = 'guest')
		FROM workspace_members WHERE workspace_id = $1 AND status = 'active'
	`, workspaceID).Scan(&u.MemberCount, &u.GuestCount)
	return u, err
}

// ---- members ----

func (s *Service) ListMembers(ctx context.Context, workspaceID, userID uuid.UUID) ([]models.Member, error) {
	if err := s.requireMember(ctx, workspaceID, userID); err != nil {
		return nil, err
	}
	rows, err := s.DB.Query(ctx, `
		SELECT m.id, m.user_id, u.email, u.display_name, m.role, m.role_id, m.seat_type, m.status, m.joined_at, m.created_at
		FROM workspace_members m
		JOIN users u ON u.id = m.user_id
		WHERE m.workspace_id = $1 AND m.status <> 'removed'
		ORDER BY m.created_at
	`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.Member, 0)
	for rows.Next() {
		var m models.Member
		var roleID *uuid.UUID
		if err := rows.Scan(&m.ID, &m.UserID, &m.Email, &m.DisplayName, &m.Role, &roleID, &m.SeatType, &m.Status, &m.JoinedAt, &m.CreatedAt); err != nil {
			return nil, err
		}
		if roleID != nil {
			id := roleID.String()
			m.RoleID = &id
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (s *Service) AddMember(ctx context.Context, workspaceID, actorID uuid.UUID, in models.AddMemberRequest) (models.Member, error) {
	if err := s.requireManage(ctx, workspaceID, actorID); err != nil {
		return models.Member{}, err
	}
	role := in.Role
	if role == "" {
		role = "member"
	}
	if !memberRoles[role] {
		return models.Member{}, httperr.ErrInvalid
	}
	email := strings.ToLower(strings.TrimSpace(in.Email))
	if email == "" {
		return models.Member{}, httperr.ErrInvalid
	}
	var targetID uuid.UUID
	err := s.DB.QueryRow(ctx, `SELECT id FROM users WHERE lower(email) = $1 AND deleted_at IS NULL`, email).Scan(&targetID)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Member{}, httperr.ErrNotFound
	}
	if err != nil {
		return models.Member{}, err
	}
	id, err := utils.NewID()
	if err != nil {
		return models.Member{}, err
	}
	_, err = s.DB.Exec(ctx, `
		INSERT INTO workspace_members (id, workspace_id, user_id, role, status, invited_by, joined_at)
		VALUES ($1, $2, $3, $4, 'active', $5, now())
	`, id, workspaceID, targetID, role, actorID)
	if err != nil {
		return models.Member{}, err
	}
	return s.getMember(ctx, workspaceID, targetID)
}

func (s *Service) UpdateMember(ctx context.Context, workspaceID, targetUserID, actorID uuid.UUID, in models.UpdateMemberRequest) (models.Member, error) {
	if err := s.requireManage(ctx, workspaceID, actorID); err != nil {
		return models.Member{}, err
	}
	if in.Role != nil {
		if !memberRoles[*in.Role] {
			return models.Member{}, httperr.ErrInvalid
		}
		if *in.Role == "owner" {
			if err := s.requireRole(ctx, workspaceID, actorID, "owner"); err != nil {
				return models.Member{}, err
			}
		}
	}
	if in.SeatType != nil && !seatTypes[*in.SeatType] {
		return models.Member{}, httperr.ErrInvalid
	}
	if in.Status != nil && !memberStatuses[*in.Status] {
		return models.Member{}, httperr.ErrInvalid
	}
	tag, err := s.DB.Exec(ctx, `
		UPDATE workspace_members SET
			role = COALESCE($3, role),
			seat_type = COALESCE($4, seat_type),
			status = COALESCE($5, status),
			deactivated_at = CASE WHEN $5 = 'removed' THEN now() ELSE deactivated_at END,
			updated_at = now()
		WHERE workspace_id = $1 AND user_id = $2
	`, workspaceID, targetUserID, in.Role, in.SeatType, in.Status)
	if err != nil {
		return models.Member{}, err
	}
	if tag.RowsAffected() == 0 {
		return models.Member{}, httperr.ErrNotFound
	}
	return s.getMember(ctx, workspaceID, targetUserID)
}

func (s *Service) RemoveMember(ctx context.Context, workspaceID, targetUserID, actorID uuid.UUID) error {
	if err := s.requireManage(ctx, workspaceID, actorID); err != nil {
		return err
	}
	tag, err := s.DB.Exec(ctx, `
		UPDATE workspace_members SET status = 'removed', deactivated_at = now(), updated_at = now()
		WHERE workspace_id = $1 AND user_id = $2 AND status <> 'removed'
	`, workspaceID, targetUserID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return httperr.ErrNotFound
	}
	return nil
}

func (s *Service) getMember(ctx context.Context, workspaceID, userID uuid.UUID) (models.Member, error) {
	var m models.Member
	var roleID *uuid.UUID
	err := s.DB.QueryRow(ctx, `
		SELECT m.id, m.user_id, u.email, u.display_name, m.role, m.role_id, m.seat_type, m.status, m.joined_at, m.created_at
		FROM workspace_members m
		JOIN users u ON u.id = m.user_id
		WHERE m.workspace_id = $1 AND m.user_id = $2
	`, workspaceID, userID).Scan(&m.ID, &m.UserID, &m.Email, &m.DisplayName, &m.Role, &roleID, &m.SeatType, &m.Status, &m.JoinedAt, &m.CreatedAt)
	if roleID != nil {
		id := roleID.String()
		m.RoleID = &id
	}
	return m, err
}

// ---- invites ----

func (s *Service) ListInvites(ctx context.Context, workspaceID, userID uuid.UUID) ([]models.Invite, error) {
	if err := s.requireManage(ctx, workspaceID, userID); err != nil {
		return nil, err
	}
	rows, err := s.DB.Query(ctx, `
		SELECT id, email, role, group_id, seat_type, expires_at, accepted_at, revoked_at, created_at
		FROM workspace_invites WHERE workspace_id = $1 ORDER BY created_at DESC
	`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.Invite, 0)
	for rows.Next() {
		var inv models.Invite
		var groupID *uuid.UUID
		if err := rows.Scan(&inv.ID, &inv.Email, &inv.Role, &groupID, &inv.SeatType, &inv.ExpiresAt, &inv.AcceptedAt, &inv.RevokedAt, &inv.CreatedAt); err != nil {
			return nil, err
		}
		if groupID != nil {
			id := groupID.String()
			inv.GroupID = &id
		}
		out = append(out, inv)
	}
	return out, rows.Err()
}

func (s *Service) CreateInvite(ctx context.Context, workspaceID, actorID uuid.UUID, in models.CreateInviteRequest) (models.Invite, string, error) {
	if err := s.requireManage(ctx, workspaceID, actorID); err != nil {
		return models.Invite{}, "", err
	}
	email := strings.ToLower(strings.TrimSpace(in.Email))
	role := in.Role
	if role == "" {
		role = "member"
	}
	seat := in.SeatType
	if seat == "" {
		seat = "full"
	}
	if email == "" || !memberRoles[role] || !seatTypes[seat] {
		return models.Invite{}, "", httperr.ErrInvalid
	}
	var groupID *uuid.UUID
	if in.GroupID != "" {
		id, err := uuid.Parse(in.GroupID)
		if err != nil {
			return models.Invite{}, "", httperr.ErrInvalid
		}
		groupID = &id
	}
	raw, err := auth.RandomToken()
	if err != nil {
		return models.Invite{}, "", err
	}
	id, err := utils.NewID()
	if err != nil {
		return models.Invite{}, "", err
	}
	expiresAt := time.Now().Add(defaultInviteTTL)
	var inv models.Invite
	err = s.DB.QueryRow(ctx, `
		INSERT INTO workspace_invites (id, workspace_id, invited_by, email, role, group_id, seat_type, token_hash, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, email, role, group_id, seat_type, expires_at, accepted_at, revoked_at, created_at
	`, id, workspaceID, actorID, email, role, groupID, seat, auth.HashToken(raw), expiresAt).Scan(
		&inv.ID, &inv.Email, &inv.Role, &groupID, &inv.SeatType, &inv.ExpiresAt, &inv.AcceptedAt, &inv.RevokedAt, &inv.CreatedAt,
	)
	if groupID != nil {
		gid := groupID.String()
		inv.GroupID = &gid
	}
	if err != nil {
		return inv, raw, err
	}
	s.sendInviteMail(ctx, workspaceID, email, raw)
	return inv, raw, nil
}

// sendInviteMail is best-effort: a delivery failure must not fail invite creation
// (the raw token is already returned to the admin in the API response either way).
func (s *Service) sendInviteMail(ctx context.Context, workspaceID uuid.UUID, toEmail, rawToken string) {
	if s.Mail == nil {
		return
	}
	var workspaceName string
	if err := s.DB.QueryRow(ctx, `SELECT name FROM workspaces WHERE id = $1`, workspaceID).Scan(&workspaceName); err != nil {
		workspaceName = "a Flow workspace"
	}
	link := s.PublicURL + "/invite?token=" + url.QueryEscape(rawToken)
	_ = s.Mail.Send(ctx, mail.Message{
		To:       toEmail,
		Template: mail.TemplateWorkspaceInvite,
		Payload: map[string]any{
			"link":      link,
			"workspace": workspaceName,
		},
	})
}

func (s *Service) RevokeInvite(ctx context.Context, workspaceID, inviteID, actorID uuid.UUID) error {
	if err := s.requireManage(ctx, workspaceID, actorID); err != nil {
		return err
	}
	tag, err := s.DB.Exec(ctx, `
		UPDATE workspace_invites SET revoked_at = now()
		WHERE id = $1 AND workspace_id = $2 AND accepted_at IS NULL AND revoked_at IS NULL
	`, inviteID, workspaceID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return httperr.ErrNotFound
	}
	return nil
}

func (s *Service) AcceptInvite(ctx context.Context, token string, userID uuid.UUID) error {
	tokenHash := auth.HashToken(token)
	var inviteID, workspaceID uuid.UUID
	var role, seat, email string
	var groupID *uuid.UUID
	var expiresAt time.Time
	var acceptedAt, revokedAt *time.Time
	err := s.DB.QueryRow(ctx, `
		SELECT id, workspace_id, role, group_id, seat_type, email, expires_at, accepted_at, revoked_at
		FROM workspace_invites WHERE token_hash = $1
	`, tokenHash).Scan(&inviteID, &workspaceID, &role, &groupID, &seat, &email, &expiresAt, &acceptedAt, &revokedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return httperr.ErrNotFound
	}
	if err != nil {
		return err
	}
	if acceptedAt != nil || revokedAt != nil || time.Now().After(expiresAt) {
		return httperr.ErrInvalid
	}
	return s.finalizeAcceptedInvite(ctx, inviteID, workspaceID, role, seat, groupID, userID)
}

// ListMyInvites returns pending invites addressed to the current user's email —
// this is what makes an invite visible in-app without depending on the invitee
// having received or clicked the emailed link.
func (s *Service) ListMyInvites(ctx context.Context, userEmail string) ([]models.MyInvite, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT i.id, i.workspace_id, w.name, i.role, u.display_name, i.expires_at, i.created_at
		FROM workspace_invites i
		JOIN workspaces w ON w.id = i.workspace_id
		LEFT JOIN users u ON u.id = i.invited_by
		WHERE lower(i.email) = lower($1) AND i.accepted_at IS NULL AND i.revoked_at IS NULL AND i.expires_at > now()
		ORDER BY i.created_at DESC
	`, userEmail)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.MyInvite, 0)
	for rows.Next() {
		var inv models.MyInvite
		if err := rows.Scan(&inv.ID, &inv.WorkspaceID, &inv.WorkspaceName, &inv.Role, &inv.InvitedByName, &inv.ExpiresAt, &inv.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, inv)
	}
	return out, rows.Err()
}

// AcceptInviteByID accepts a pending invite the current user can see via
// ListMyInvites — authorized by the invite's email matching the signed-in
// user's own email, rather than possession of the (never-stored-in-plaintext)
// invite token. This is what a "click to accept" button in the product itself
// can use, as opposed to the emailed link's /invites/{token}/accept.
func (s *Service) AcceptInviteByID(ctx context.Context, workspaceID, inviteID, userID uuid.UUID, userEmail string) error {
	var role, seat, email string
	var groupID *uuid.UUID
	var expiresAt time.Time
	var acceptedAt, revokedAt *time.Time
	err := s.DB.QueryRow(ctx, `
		SELECT role, group_id, seat_type, email, expires_at, accepted_at, revoked_at
		FROM workspace_invites WHERE id = $1 AND workspace_id = $2
	`, inviteID, workspaceID).Scan(&role, &groupID, &seat, &email, &expiresAt, &acceptedAt, &revokedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return httperr.ErrNotFound
	}
	if err != nil {
		return err
	}
	if !strings.EqualFold(email, userEmail) {
		return httperr.ErrForbidden
	}
	if acceptedAt != nil || revokedAt != nil || time.Now().After(expiresAt) {
		return httperr.ErrInvalid
	}
	return s.finalizeAcceptedInvite(ctx, inviteID, workspaceID, role, seat, groupID, userID)
}

func (s *Service) finalizeAcceptedInvite(ctx context.Context, inviteID, workspaceID uuid.UUID, role, seat string, groupID *uuid.UUID, userID uuid.UUID) error {
	return withTx(ctx, s.DB, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `
			INSERT INTO workspace_members (id, workspace_id, user_id, role, seat_type, status, invited_by, joined_at)
			VALUES ($1, $2, $3, $4, $5, 'active', NULL, now())
			ON CONFLICT (workspace_id, user_id) DO UPDATE SET status = 'active', role = EXCLUDED.role, updated_at = now()
		`, utils.MustID(), workspaceID, userID, role, seat); err != nil {
			return err
		}
		if groupID != nil {
			if _, err := tx.Exec(ctx, `
				INSERT INTO workspace_group_members (group_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING
			`, *groupID, userID); err != nil {
				return err
			}
		}
		_, err := tx.Exec(ctx, `UPDATE workspace_invites SET accepted_at = now(), accepted_by = $2 WHERE id = $1`, inviteID, userID)
		return err
	})
}

// ---- roles ----

func (s *Service) ListRoles(ctx context.Context, workspaceID, userID uuid.UUID) ([]models.Role, error) {
	if err := s.requireMember(ctx, workspaceID, userID); err != nil {
		return nil, err
	}
	rows, err := s.DB.Query(ctx, `
		SELECT id, name, description, is_system, permissions, created_at
		FROM workspace_roles WHERE workspace_id = $1 ORDER BY created_at
	`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.Role, 0)
	for rows.Next() {
		var r models.Role
		if err := rows.Scan(&r.ID, &r.Name, &r.Description, &r.IsSystem, &r.Permissions, &r.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Service) CreateRole(ctx context.Context, workspaceID, userID uuid.UUID, in models.CreateRoleRequest) (models.Role, error) {
	if err := s.requireManage(ctx, workspaceID, userID); err != nil {
		return models.Role{}, err
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return models.Role{}, httperr.ErrInvalid
	}
	perms := in.Permissions
	if len(perms) == 0 {
		perms = json.RawMessage(`{}`)
	}
	id, err := utils.NewID()
	if err != nil {
		return models.Role{}, err
	}
	var r models.Role
	err = s.DB.QueryRow(ctx, `
		INSERT INTO workspace_roles (id, workspace_id, name, description, permissions)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, name, description, is_system, permissions, created_at
	`, id, workspaceID, name, in.Description, perms).Scan(&r.ID, &r.Name, &r.Description, &r.IsSystem, &r.Permissions, &r.CreatedAt)
	return r, err
}

func (s *Service) UpdateRole(ctx context.Context, workspaceID, roleID, userID uuid.UUID, in models.UpdateRoleRequest) (models.Role, error) {
	if err := s.requireManage(ctx, workspaceID, userID); err != nil {
		return models.Role{}, err
	}
	var perms any
	if len(in.Permissions) > 0 {
		perms = in.Permissions
	}
	var r models.Role
	err := s.DB.QueryRow(ctx, `
		UPDATE workspace_roles SET
			name = COALESCE($3, name),
			description = COALESCE($4, description),
			permissions = COALESCE($5, permissions),
			updated_at = now()
		WHERE id = $1 AND workspace_id = $2 AND is_system = false
		RETURNING id, name, description, is_system, permissions, created_at
	`, roleID, workspaceID, in.Name, in.Description, perms).Scan(&r.ID, &r.Name, &r.Description, &r.IsSystem, &r.Permissions, &r.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Role{}, httperr.ErrNotFound
	}
	return r, err
}

func (s *Service) DeleteRole(ctx context.Context, workspaceID, roleID, userID uuid.UUID) error {
	if err := s.requireManage(ctx, workspaceID, userID); err != nil {
		return err
	}
	tag, err := s.DB.Exec(ctx, `
		DELETE FROM workspace_roles WHERE id = $1 AND workspace_id = $2 AND is_system = false
	`, roleID, workspaceID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return httperr.ErrNotFound
	}
	return nil
}

// ---- groups ----

func (s *Service) ListGroups(ctx context.Context, workspaceID, userID uuid.UUID) ([]models.Group, error) {
	if err := s.requireMember(ctx, workspaceID, userID); err != nil {
		return nil, err
	}
	rows, err := s.DB.Query(ctx, `
		SELECT id, name, description, created_at
		FROM workspace_groups WHERE workspace_id = $1 AND deleted_at IS NULL ORDER BY created_at
	`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.Group, 0)
	for rows.Next() {
		var g models.Group
		if err := rows.Scan(&g.ID, &g.Name, &g.Description, &g.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, rows.Err()
}

func (s *Service) CreateGroup(ctx context.Context, workspaceID, userID uuid.UUID, in models.CreateGroupRequest) (models.Group, error) {
	if err := s.requireManage(ctx, workspaceID, userID); err != nil {
		return models.Group{}, err
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return models.Group{}, httperr.ErrInvalid
	}
	id, err := utils.NewID()
	if err != nil {
		return models.Group{}, err
	}
	var g models.Group
	err = s.DB.QueryRow(ctx, `
		INSERT INTO workspace_groups (id, workspace_id, name, description, created_by)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, name, description, created_at
	`, id, workspaceID, name, in.Description, userID).Scan(&g.ID, &g.Name, &g.Description, &g.CreatedAt)
	return g, err
}

func (s *Service) UpdateGroup(ctx context.Context, workspaceID, groupID, userID uuid.UUID, in models.UpdateGroupRequest) (models.Group, error) {
	if err := s.requireManage(ctx, workspaceID, userID); err != nil {
		return models.Group{}, err
	}
	var g models.Group
	err := s.DB.QueryRow(ctx, `
		UPDATE workspace_groups SET
			name = COALESCE($3, name),
			description = COALESCE($4, description),
			updated_at = now()
		WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL
		RETURNING id, name, description, created_at
	`, groupID, workspaceID, in.Name, in.Description).Scan(&g.ID, &g.Name, &g.Description, &g.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Group{}, httperr.ErrNotFound
	}
	return g, err
}

func (s *Service) DeleteGroup(ctx context.Context, workspaceID, groupID, userID uuid.UUID) error {
	if err := s.requireManage(ctx, workspaceID, userID); err != nil {
		return err
	}
	tag, err := s.DB.Exec(ctx, `
		UPDATE workspace_groups SET deleted_at = now(), updated_at = now()
		WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL
	`, groupID, workspaceID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return httperr.ErrNotFound
	}
	return nil
}

func (s *Service) ListGroupMembers(ctx context.Context, workspaceID, groupID, userID uuid.UUID) ([]models.GroupMember, error) {
	if err := s.requireMember(ctx, workspaceID, userID); err != nil {
		return nil, err
	}
	rows, err := s.DB.Query(ctx, `
		SELECT gm.user_id, u.email, u.display_name, gm.created_at
		FROM workspace_group_members gm
		JOIN workspace_groups g ON g.id = gm.group_id
		JOIN users u ON u.id = gm.user_id
		WHERE gm.group_id = $1 AND g.workspace_id = $2
		ORDER BY gm.created_at
	`, groupID, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.GroupMember, 0)
	for rows.Next() {
		var m models.GroupMember
		if err := rows.Scan(&m.UserID, &m.Email, &m.DisplayName, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (s *Service) AddGroupMember(ctx context.Context, workspaceID, groupID, userID uuid.UUID, in models.AddGroupMemberRequest) error {
	if err := s.requireManage(ctx, workspaceID, userID); err != nil {
		return err
	}
	targetID, err := uuid.Parse(in.UserID)
	if err != nil {
		return httperr.ErrInvalid
	}
	tag, err := s.DB.Exec(ctx, `
		INSERT INTO workspace_group_members (group_id, user_id, added_by)
		SELECT $1, $2, $3 WHERE EXISTS (SELECT 1 FROM workspace_groups WHERE id = $1 AND workspace_id = $4)
		ON CONFLICT DO NOTHING
	`, groupID, targetID, userID, workspaceID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return httperr.ErrNotFound
	}
	return nil
}

func (s *Service) RemoveGroupMember(ctx context.Context, workspaceID, groupID, targetUserID, userID uuid.UUID) error {
	if err := s.requireManage(ctx, workspaceID, userID); err != nil {
		return err
	}
	tag, err := s.DB.Exec(ctx, `
		DELETE FROM workspace_group_members gm
		USING workspace_groups g
		WHERE gm.group_id = g.id AND g.workspace_id = $1 AND gm.group_id = $2 AND gm.user_id = $3
	`, workspaceID, groupID, targetUserID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return httperr.ErrNotFound
	}
	return nil
}

// ---- domains ----

func (s *Service) ListDomains(ctx context.Context, workspaceID, userID uuid.UUID) ([]models.Domain, error) {
	if err := s.requireMember(ctx, workspaceID, userID); err != nil {
		return nil, err
	}
	rows, err := s.DB.Query(ctx, `
		SELECT id, domain, verified_at, created_at FROM workspace_domains WHERE workspace_id = $1 ORDER BY created_at
	`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.Domain, 0)
	for rows.Next() {
		var d models.Domain
		if err := rows.Scan(&d.ID, &d.Domain, &d.VerifiedAt, &d.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (s *Service) CreateDomain(ctx context.Context, workspaceID, userID uuid.UUID, in models.CreateDomainRequest) (models.Domain, error) {
	if err := s.requireManage(ctx, workspaceID, userID); err != nil {
		return models.Domain{}, err
	}
	domain := strings.ToLower(strings.TrimSpace(in.Domain))
	if domain == "" {
		return models.Domain{}, httperr.ErrInvalid
	}
	id, err := utils.NewID()
	if err != nil {
		return models.Domain{}, err
	}
	var d models.Domain
	err = s.DB.QueryRow(ctx, `
		INSERT INTO workspace_domains (id, workspace_id, domain) VALUES ($1, $2, $3)
		RETURNING id, domain, verified_at, created_at
	`, id, workspaceID, domain).Scan(&d.ID, &d.Domain, &d.VerifiedAt, &d.CreatedAt)
	return d, err
}

func (s *Service) DeleteDomain(ctx context.Context, workspaceID, domainID, userID uuid.UUID) error {
	if err := s.requireManage(ctx, workspaceID, userID); err != nil {
		return err
	}
	tag, err := s.DB.Exec(ctx, `DELETE FROM workspace_domains WHERE id = $1 AND workspace_id = $2`, domainID, workspaceID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return httperr.ErrNotFound
	}
	return nil
}

// ---- helpers ----

const workspaceCols = `id, name, slug, description, icon, color, status, settings, owner_user_id,
	public_sharing_enabled, guest_access_enabled, created_at, updated_at`

// workspaceColsW is workspaceCols qualified with the `w` alias, for queries that join
// workspaces against another table (e.g. workspace_members) which also has an `id` column.
const workspaceColsW = `w.id, w.name, w.slug, w.description, w.icon, w.color, w.status, w.settings, w.owner_user_id,
	w.public_sharing_enabled, w.guest_access_enabled, w.created_at, w.updated_at`

func scanWorkspace(row pgx.Row) (models.Workspace, error) {
	var w models.Workspace
	var slug, icon, color *string
	var ownerID *uuid.UUID
	var id uuid.UUID
	err := row.Scan(&id, &w.Name, &slug, &w.Description, &icon, &color, &w.Status, &w.Settings, &ownerID,
		&w.PublicSharingEnabled, &w.GuestAccessEnabled, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		return models.Workspace{}, err
	}
	w.ID = id.String()
	w.Slug = slug
	w.Icon = icon
	w.Color = color
	if ownerID != nil {
		s := ownerID.String()
		w.OwnerUserID = &s
	}
	return w, nil
}

func (s *Service) getWorkspace(ctx context.Context, workspaceID uuid.UUID) (models.Workspace, error) {
	w, err := scanWorkspace(s.DB.QueryRow(ctx, `SELECT `+workspaceCols+` FROM workspaces WHERE id = $1 AND deleted_at IS NULL`, workspaceID))
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Workspace{}, httperr.ErrNotFound
	}
	return w, err
}

func (s *Service) requireMember(ctx context.Context, workspaceID, userID uuid.UUID) error {
	var ok bool
	err := s.DB.QueryRow(ctx, `
		SELECT true FROM workspace_members WHERE workspace_id = $1 AND user_id = $2 AND status = 'active'
	`, workspaceID, userID).Scan(&ok)
	if errors.Is(err, pgx.ErrNoRows) {
		return httperr.ErrForbidden
	}
	return err
}

func (s *Service) requireRole(ctx context.Context, workspaceID, userID uuid.UUID, roles ...string) error {
	var role string
	err := s.DB.QueryRow(ctx, `
		SELECT role FROM workspace_members WHERE workspace_id = $1 AND user_id = $2 AND status = 'active'
	`, workspaceID, userID).Scan(&role)
	if errors.Is(err, pgx.ErrNoRows) {
		return httperr.ErrForbidden
	}
	if err != nil {
		return err
	}
	for _, r := range roles {
		if role == r {
			return nil
		}
	}
	return httperr.ErrForbidden
}

func (s *Service) requireManage(ctx context.Context, workspaceID, userID uuid.UUID) error {
	return s.requireRole(ctx, workspaceID, userID, "owner", "admin")
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

func nullIfEmpty(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}
