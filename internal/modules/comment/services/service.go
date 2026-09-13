package services

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kashifxyz/flow-server/internal/access"
	"github.com/kashifxyz/flow-server/internal/graph"
	"github.com/kashifxyz/flow-server/internal/modules/comment/models"
	"github.com/kashifxyz/flow-server/internal/utils/httperr"
)

const nodeTypeComment = "comment"

type Service struct {
	DB *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Service {
	return &Service{DB: db}
}

func (s *Service) ListDiscussions(ctx context.Context, nodeID, userID uuid.UUID) ([]models.Discussion, error) {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, nodeID); err != nil {
		return nil, err
	}
	rows, err := s.DB.Query(ctx, `
		SELECT id, node_id, created_by, resolved_at, resolved_by, selection_anchor, comment_count, created_at, updated_at
		FROM discussions WHERE node_id = $1 ORDER BY created_at
	`, nodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.Discussion, 0)
	for rows.Next() {
		d, err := scanDiscussion(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (s *Service) CreateDiscussion(ctx context.Context, nodeID, userID uuid.UUID, in models.CreateDiscussionRequest) (models.Discussion, error) {
	workspaceID, err := access.RequireNodeMember(ctx, s.DB, userID, nodeID)
	if err != nil {
		return models.Discussion{}, err
	}
	return scanDiscussion(s.DB.QueryRow(ctx, `
		INSERT INTO discussions (workspace_id, node_id, created_by, selection_anchor)
		VALUES ($1, $2, $3, $4)
		RETURNING id, node_id, created_by, resolved_at, resolved_by, selection_anchor, comment_count, created_at, updated_at
	`, workspaceID, nodeID, userID, nullIfEmptyJSON(in.SelectionAnchor)))
}

func (s *Service) GetDiscussion(ctx context.Context, discussionID, userID uuid.UUID) (models.Discussion, error) {
	d, err := s.getDiscussion(ctx, discussionID)
	if err != nil {
		return models.Discussion{}, err
	}
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, uuid.MustParse(d.NodeID)); err != nil {
		return models.Discussion{}, err
	}
	return d, nil
}

func (s *Service) ResolveDiscussion(ctx context.Context, discussionID, userID uuid.UUID) (models.Discussion, error) {
	d, err := s.getDiscussion(ctx, discussionID)
	if err != nil {
		return models.Discussion{}, err
	}
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, uuid.MustParse(d.NodeID)); err != nil {
		return models.Discussion{}, err
	}
	return scanDiscussion(s.DB.QueryRow(ctx, `
		UPDATE discussions SET resolved_at = now(), resolved_by = $2, updated_at = now()
		WHERE id = $1
		RETURNING id, node_id, created_by, resolved_at, resolved_by, selection_anchor, comment_count, created_at, updated_at
	`, discussionID, userID))
}

func (s *Service) ListComments(ctx context.Context, discussionID, userID uuid.UUID) ([]models.Comment, error) {
	d, err := s.getDiscussion(ctx, discussionID)
	if err != nil {
		return nil, err
	}
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, uuid.MustParse(d.NodeID)); err != nil {
		return nil, err
	}
	rows, err := s.DB.Query(ctx, `SELECT `+commentCols+` FROM comments WHERE discussion_id = $1 AND deleted_at IS NULL ORDER BY created_at`, discussionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.Comment, 0)
	for rows.Next() {
		c, err := scanComment(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Service) CreateComment(ctx context.Context, discussionID, userID uuid.UUID, in models.CreateCommentRequest) (models.Comment, error) {
	d, err := s.getDiscussion(ctx, discussionID)
	if err != nil {
		return models.Comment{}, err
	}
	nodeID := uuid.MustParse(d.NodeID)
	workspaceID, err := access.RequireNodeMember(ctx, s.DB, userID, nodeID)
	if err != nil {
		return models.Comment{}, err
	}
	body := strings.TrimSpace(in.Body)
	if body == "" {
		return models.Comment{}, httperr.ErrInvalid
	}
	var parentID *uuid.UUID
	if in.ParentID != "" {
		id, err := uuid.Parse(in.ParentID)
		if err != nil {
			return models.Comment{}, httperr.ErrInvalid
		}
		parentID = &id
	}
	richText := in.RichText
	if len(richText) == 0 {
		richText = []byte(`[]`)
	}
	var c models.Comment
	err = withTx(ctx, s.DB, func(tx pgx.Tx) error {
		n, err := graph.Create(ctx, tx, graph.Insert{
			WorkspaceID: workspaceID,
			Type:        nodeTypeComment,
			CreatedBy:   userID,
		})
		if err != nil {
			return err
		}
		c, err = scanComment(tx.QueryRow(ctx, `
			INSERT INTO comments (id, workspace_id, node_id, discussion_id, parent_id, created_by, body, rich_text)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			RETURNING `+commentCols, n.ID, workspaceID, nodeID, discussionID, parentID, userID, body, richText))
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `UPDATE discussions SET comment_count = comment_count + 1, updated_at = now() WHERE id = $1`, discussionID)
		return err
	})
	return c, err
}

func (s *Service) UpdateComment(ctx context.Context, commentID, userID uuid.UUID, in models.UpdateCommentRequest) (models.Comment, error) {
	nodeID, err := s.commentNodeID(ctx, commentID)
	if err != nil {
		return models.Comment{}, err
	}
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, nodeID); err != nil {
		return models.Comment{}, err
	}
	c, err := scanComment(s.DB.QueryRow(ctx, `
		UPDATE comments SET
			body = COALESCE($2, body),
			rich_text = COALESCE($3, rich_text),
			edited_at = now(),
			updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING `+commentCols, commentID, in.Body, in.RichText))
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Comment{}, httperr.ErrNotFound
	}
	return c, err
}

func (s *Service) DeleteComment(ctx context.Context, commentID, userID uuid.UUID) error {
	nodeID, err := s.commentNodeID(ctx, commentID)
	if err != nil {
		return err
	}
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, nodeID); err != nil {
		return err
	}
	var discussionID uuid.UUID
	tag, err := s.DB.Exec(ctx, `UPDATE comments SET deleted_at = now(), updated_at = now() WHERE id = $1 AND deleted_at IS NULL`, commentID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return httperr.ErrNotFound
	}
	if err := s.DB.QueryRow(ctx, `SELECT discussion_id FROM comments WHERE id = $1`, commentID).Scan(&discussionID); err == nil {
		_, _ = s.DB.Exec(ctx, `UPDATE discussions SET comment_count = GREATEST(comment_count - 1, 0), updated_at = now() WHERE id = $1`, discussionID)
	}
	return nil
}

func (s *Service) AddReaction(ctx context.Context, commentID, userID uuid.UUID, in models.CreateReactionRequest) (models.Reaction, error) {
	nodeID, err := s.commentNodeID(ctx, commentID)
	if err != nil {
		return models.Reaction{}, err
	}
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, nodeID); err != nil {
		return models.Reaction{}, err
	}
	emoji := strings.TrimSpace(in.Emoji)
	if emoji == "" {
		return models.Reaction{}, httperr.ErrInvalid
	}
	var r models.Reaction
	var cID, uID uuid.UUID
	err = s.DB.QueryRow(ctx, `
		INSERT INTO comment_reactions (comment_id, user_id, emoji) VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING
		RETURNING comment_id, user_id, emoji, created_at
	`, commentID, userID, emoji).Scan(&cID, &uID, &r.Emoji, &r.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		r = models.Reaction{CommentID: commentID.String(), UserID: userID.String(), Emoji: emoji}
		return r, nil
	}
	if err != nil {
		return models.Reaction{}, err
	}
	r.CommentID = cID.String()
	r.UserID = uID.String()
	return r, nil
}

func (s *Service) RemoveReaction(ctx context.Context, commentID, userID uuid.UUID, emoji string) error {
	nodeID, err := s.commentNodeID(ctx, commentID)
	if err != nil {
		return err
	}
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, nodeID); err != nil {
		return err
	}
	tag, err := s.DB.Exec(ctx, `DELETE FROM comment_reactions WHERE comment_id = $1 AND user_id = $2 AND emoji = $3`, commentID, userID, emoji)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return httperr.ErrNotFound
	}
	return nil
}

func (s *Service) commentNodeID(ctx context.Context, commentID uuid.UUID) (uuid.UUID, error) {
	var nodeID uuid.UUID
	err := s.DB.QueryRow(ctx, `SELECT node_id FROM comments WHERE id = $1`, commentID).Scan(&nodeID)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, httperr.ErrNotFound
	}
	return nodeID, err
}

func (s *Service) getDiscussion(ctx context.Context, discussionID uuid.UUID) (models.Discussion, error) {
	d, err := scanDiscussion(s.DB.QueryRow(ctx, `
		SELECT id, node_id, created_by, resolved_at, resolved_by, selection_anchor, comment_count, created_at, updated_at
		FROM discussions WHERE id = $1
	`, discussionID))
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Discussion{}, httperr.ErrNotFound
	}
	return d, err
}

func scanDiscussion(row pgx.Row) (models.Discussion, error) {
	var d models.Discussion
	var id, nodeID uuid.UUID
	var createdBy, resolvedBy *uuid.UUID
	err := row.Scan(&id, &nodeID, &createdBy, &d.ResolvedAt, &resolvedBy, &d.SelectionAnchor, &d.CommentCount, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return models.Discussion{}, err
	}
	d.ID = id.String()
	d.NodeID = nodeID.String()
	if createdBy != nil {
		s := createdBy.String()
		d.CreatedBy = &s
	}
	if resolvedBy != nil {
		s := resolvedBy.String()
		d.ResolvedBy = &s
	}
	return d, nil
}

const commentCols = `id, discussion_id, node_id, parent_id, body, rich_text, body_format, created_by, edited_at, created_at, updated_at`

func scanComment(row pgx.Row) (models.Comment, error) {
	var c models.Comment
	var id, discussionID, nodeID uuid.UUID
	var parentID, createdBy *uuid.UUID
	err := row.Scan(&id, &discussionID, &nodeID, &parentID, &c.Body, &c.RichText, &c.BodyFormat, &createdBy, &c.EditedAt, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return models.Comment{}, err
	}
	c.ID = id.String()
	c.DiscussionID = discussionID.String()
	c.NodeID = nodeID.String()
	if parentID != nil {
		s := parentID.String()
		c.ParentID = &s
	}
	if createdBy != nil {
		s := createdBy.String()
		c.CreatedBy = &s
	}
	return c, nil
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

func nullIfEmptyJSON(b []byte) any {
	if len(b) == 0 {
		return nil
	}
	return b
}
