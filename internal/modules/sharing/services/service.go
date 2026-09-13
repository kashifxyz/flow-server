package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kashifxyz/flow-server/internal/access"
	"github.com/kashifxyz/flow-server/internal/auth"
	"github.com/kashifxyz/flow-server/internal/modules/sharing/models"
	"github.com/kashifxyz/flow-server/internal/utils"
	"github.com/kashifxyz/flow-server/internal/utils/httperr"
)

type Service struct {
	DB *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Service {
	return &Service{DB: db}
}

const shareLinkCols = `id, node_id, role, (password_hash IS NOT NULL), include_subtree, allow_duplicate, max_uses, use_count, watermark, expires_at, revoked_at, created_at`

func (s *Service) List(ctx context.Context, nodeID, userID uuid.UUID) ([]models.ShareLink, error) {
	if err := s.requireManage(ctx, nodeID, userID); err != nil {
		return nil, err
	}
	rows, err := s.DB.Query(ctx, `SELECT `+shareLinkCols+` FROM share_links WHERE node_id = $1 ORDER BY created_at DESC`, nodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.ShareLink, 0)
	for rows.Next() {
		l, err := scanShareLink(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

func (s *Service) Create(ctx context.Context, nodeID, userID uuid.UUID, in models.CreateShareLinkRequest) (models.CreateShareLinkResponse, error) {
	workspaceID, err := access.RequireNodeMember(ctx, s.DB, userID, nodeID)
	if err != nil {
		return models.CreateShareLinkResponse{}, err
	}
	if err := s.requireManage(ctx, nodeID, userID); err != nil {
		return models.CreateShareLinkResponse{}, err
	}
	role := in.Role
	if role == "" {
		role = "view"
	}
	if !models.Roles[role] {
		return models.CreateShareLinkResponse{}, httperr.ErrInvalid
	}
	var passwordHash *string
	if in.Password != "" {
		hash, err := auth.HashPassword(in.Password)
		if err != nil {
			return models.CreateShareLinkResponse{}, err
		}
		passwordHash = &hash
	}
	raw, err := auth.RandomToken()
	if err != nil {
		return models.CreateShareLinkResponse{}, err
	}
	id, err := utils.NewID()
	if err != nil {
		return models.CreateShareLinkResponse{}, err
	}
	l, err := scanShareLink(s.DB.QueryRow(ctx, `
		INSERT INTO share_links (id, workspace_id, node_id, created_by, token_hash, role, password_hash, include_subtree, allow_duplicate, max_uses, watermark, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING `+shareLinkCols, id, workspaceID, nodeID, userID, auth.HashToken(raw), role, passwordHash,
		in.IncludeSubtree, in.AllowDuplicate, in.MaxUses, in.Watermark, in.ExpiresAt))
	if err != nil {
		return models.CreateShareLinkResponse{}, err
	}
	return models.CreateShareLinkResponse{ShareLink: l, Token: raw}, nil
}

func (s *Service) Update(ctx context.Context, linkID, userID uuid.UUID, in models.UpdateShareLinkRequest) (models.ShareLink, error) {
	nodeID, err := s.linkNodeID(ctx, linkID)
	if err != nil {
		return models.ShareLink{}, err
	}
	if err := s.requireManage(ctx, nodeID, userID); err != nil {
		return models.ShareLink{}, err
	}
	if in.Role != nil && !models.Roles[*in.Role] {
		return models.ShareLink{}, httperr.ErrInvalid
	}
	l, err := scanShareLink(s.DB.QueryRow(ctx, `
		UPDATE share_links SET
			role = COALESCE($2, role),
			include_subtree = COALESCE($3, include_subtree),
			allow_duplicate = COALESCE($4, allow_duplicate),
			max_uses = COALESCE($5, max_uses),
			watermark = COALESCE($6, watermark),
			expires_at = COALESCE($7, expires_at)
		WHERE id = $1 AND revoked_at IS NULL
		RETURNING `+shareLinkCols, linkID, in.Role, in.IncludeSubtree, in.AllowDuplicate, in.MaxUses, in.Watermark, in.ExpiresAt))
	if errors.Is(err, pgx.ErrNoRows) {
		return models.ShareLink{}, httperr.ErrNotFound
	}
	return l, err
}

func (s *Service) Revoke(ctx context.Context, linkID, userID uuid.UUID) error {
	nodeID, err := s.linkNodeID(ctx, linkID)
	if err != nil {
		return err
	}
	if err := s.requireManage(ctx, nodeID, userID); err != nil {
		return err
	}
	tag, err := s.DB.Exec(ctx, `UPDATE share_links SET revoked_at = now() WHERE id = $1 AND revoked_at IS NULL`, linkID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return httperr.ErrNotFound
	}
	return nil
}

func (s *Service) ResolvePublic(ctx context.Context, token, password, ip, userAgent string) (models.PublicResolveResponse, error) {
	tokenHash := auth.HashToken(token)
	var linkID, nodeID uuid.UUID
	var role string
	var passwordHash *string
	var maxUses, useCount *int
	var expiresAt, revokedAt *time.Time
	err := s.DB.QueryRow(ctx, `
		SELECT id, node_id, role, password_hash, max_uses, use_count, expires_at, revoked_at
		FROM share_links WHERE token_hash = $1
	`, tokenHash).Scan(&linkID, &nodeID, &role, &passwordHash, &maxUses, &useCount, &expiresAt, &revokedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.PublicResolveResponse{}, httperr.ErrNotFound
	}
	if err != nil {
		return models.PublicResolveResponse{}, err
	}
	if revokedAt != nil || (expiresAt != nil && time.Now().After(*expiresAt)) {
		return models.PublicResolveResponse{}, httperr.ErrNotFound
	}
	if maxUses != nil && useCount != nil && *useCount >= *maxUses {
		return models.PublicResolveResponse{}, httperr.ErrNotFound
	}
	if passwordHash != nil {
		ok, err := auth.ComparePassword(*passwordHash, password)
		if err != nil || !ok {
			return models.PublicResolveResponse{}, httperr.ErrForbidden
		}
	}
	var nodeType, title string
	err = s.DB.QueryRow(ctx, `SELECT type, title FROM nodes WHERE id = $1 AND deleted_at IS NULL`, nodeID).Scan(&nodeType, &title)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.PublicResolveResponse{}, httperr.ErrNotFound
	}
	if err != nil {
		return models.PublicResolveResponse{}, err
	}
	_, _ = s.DB.Exec(ctx, `UPDATE share_links SET use_count = use_count + 1, last_accessed_at = now() WHERE id = $1`, linkID)
	_, _ = s.DB.Exec(ctx, `
		INSERT INTO share_link_accesses (share_link_id, ip, user_agent) VALUES ($1, $2, $3)
	`, linkID, nullIfEmpty(ip), nullIfEmpty(userAgent))
	return models.PublicResolveResponse{NodeID: nodeID.String(), Type: nodeType, Title: title, Role: role}, nil
}

func (s *Service) Publish(ctx context.Context, nodeID, userID uuid.UUID, in models.PublishRequest) (models.PublishedPage, error) {
	workspaceID, err := access.RequireNodeMember(ctx, s.DB, userID, nodeID)
	if err != nil {
		return models.PublishedPage{}, err
	}
	if err := s.requireManage(ctx, nodeID, userID); err != nil {
		return models.PublishedPage{}, err
	}
	slug := strings.TrimSpace(in.Slug)
	if slug == "" {
		return models.PublishedPage{}, httperr.ErrInvalid
	}
	var title string
	if err := s.DB.QueryRow(ctx, `SELECT title FROM nodes WHERE id = $1`, nodeID).Scan(&title); err != nil {
		return models.PublishedPage{}, err
	}
	var p models.PublishedPage
	var nID uuid.UUID
	err = s.DB.QueryRow(ctx, `
		INSERT INTO published_pages (node_id, workspace_id, slug, title, indexed, published_by)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (node_id) DO UPDATE SET
			slug = EXCLUDED.slug, title = EXCLUDED.title, indexed = EXCLUDED.indexed,
			published_by = EXCLUDED.published_by, unpublished_at = NULL, updated_at = now()
		RETURNING node_id, slug, title, indexed, created_at, updated_at
	`, nodeID, workspaceID, slug, title, in.Indexed, userID).Scan(&nID, &p.Slug, &p.Title, &p.Indexed, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return models.PublishedPage{}, err
	}
	p.NodeID = nID.String()
	if _, err := s.DB.Exec(ctx, `UPDATE nodes SET published_at = now(), published_by = $2, public_slug = $3, updated_at = now() WHERE id = $1`, nodeID, userID, slug); err != nil {
		return models.PublishedPage{}, err
	}
	return p, nil
}

func (s *Service) Unpublish(ctx context.Context, nodeID, userID uuid.UUID) error {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, nodeID); err != nil {
		return err
	}
	if err := s.requireManage(ctx, nodeID, userID); err != nil {
		return err
	}
	tag, err := s.DB.Exec(ctx, `
		UPDATE published_pages SET unpublished_at = now(), updated_at = now() WHERE node_id = $1 AND unpublished_at IS NULL
	`, nodeID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return httperr.ErrNotFound
	}
	_, err = s.DB.Exec(ctx, `UPDATE nodes SET published_at = NULL, public_slug = NULL, updated_at = now() WHERE id = $1`, nodeID)
	return err
}

func (s *Service) linkNodeID(ctx context.Context, linkID uuid.UUID) (uuid.UUID, error) {
	var nodeID uuid.UUID
	err := s.DB.QueryRow(ctx, `SELECT node_id FROM share_links WHERE id = $1`, linkID).Scan(&nodeID)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, httperr.ErrNotFound
	}
	return nodeID, err
}

func (s *Service) requireManage(ctx context.Context, nodeID, userID uuid.UUID) error {
	workspaceID, err := access.WorkspaceIDForNode(ctx, s.DB, nodeID)
	if err != nil {
		return err
	}
	var ok bool
	err = s.DB.QueryRow(ctx, `
		SELECT true WHERE
			EXISTS (SELECT 1 FROM workspace_members WHERE workspace_id = $1 AND user_id = $2 AND status = 'active' AND role IN ('owner','admin'))
			OR EXISTS (SELECT 1 FROM node_permissions WHERE node_id = $3 AND principal_type = 'user' AND principal_id = $2 AND role = 'manage')
	`, workspaceID, userID, nodeID).Scan(&ok)
	if errors.Is(err, pgx.ErrNoRows) {
		return httperr.ErrForbidden
	}
	return err
}

func scanShareLink(row pgx.Row) (models.ShareLink, error) {
	var l models.ShareLink
	var id, nodeID uuid.UUID
	err := row.Scan(&id, &nodeID, &l.Role, &l.PasswordProtected, &l.IncludeSubtree, &l.AllowDuplicate,
		&l.MaxUses, &l.UseCount, &l.Watermark, &l.ExpiresAt, &l.RevokedAt, &l.CreatedAt)
	if err != nil {
		return models.ShareLink{}, err
	}
	l.ID = id.String()
	l.NodeID = nodeID.String()
	return l, nil
}

func nullIfEmpty(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}
