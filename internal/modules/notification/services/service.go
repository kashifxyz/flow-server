package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kashifxyz/flow-server/internal/access"
	"github.com/kashifxyz/flow-server/internal/modules/notification/models"
	"github.com/kashifxyz/flow-server/internal/utils/httperr"
)

type Service struct {
	DB *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Service {
	return &Service{DB: db}
}

func (s *Service) List(ctx context.Context, userID uuid.UUID, workspaceID *uuid.UUID) ([]models.Notification, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT id, workspace_id, actor_id, subject_node_id, type, title, body, payload, read_at, created_at
		FROM notifications
		WHERE user_id = $1 AND archived_at IS NULL AND ($2::uuid IS NULL OR workspace_id = $2)
		ORDER BY created_at DESC
		LIMIT 100
	`, userID, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.Notification, 0)
	for rows.Next() {
		n, err := scanNotification(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

func (s *Service) MarkAllRead(ctx context.Context, userID uuid.UUID, workspaceID *uuid.UUID) error {
	_, err := s.DB.Exec(ctx, `
		UPDATE notifications SET read_at = now()
		WHERE user_id = $1 AND read_at IS NULL AND ($2::uuid IS NULL OR workspace_id = $2)
	`, userID, workspaceID)
	return err
}

func (s *Service) MarkRead(ctx context.Context, notificationID, userID uuid.UUID) error {
	tag, err := s.DB.Exec(ctx, `
		UPDATE notifications SET read_at = now() WHERE id = $1 AND user_id = $2 AND read_at IS NULL
	`, notificationID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return httperr.ErrNotFound
	}
	return nil
}

func (s *Service) ListActivity(ctx context.Context, workspaceID, userID uuid.UUID) ([]models.ActivityEvent, error) {
	if err := access.RequireMember(ctx, s.DB, workspaceID, userID); err != nil {
		return nil, err
	}
	rows, err := s.DB.Query(ctx, `
		SELECT id, workspace_id, node_id, actor_id, event_type, summary, payload, created_at
		FROM activity_events WHERE workspace_id = $1
		ORDER BY created_at DESC
		LIMIT 100
	`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.ActivityEvent, 0)
	for rows.Next() {
		var e models.ActivityEvent
		var id, wsID uuid.UUID
		var nodeID, actorID *uuid.UUID
		if err := rows.Scan(&id, &wsID, &nodeID, &actorID, &e.EventType, &e.Summary, &e.Payload, &e.CreatedAt); err != nil {
			return nil, err
		}
		e.ID = id.String()
		e.WorkspaceID = wsID.String()
		if nodeID != nil {
			s := nodeID.String()
			e.NodeID = &s
		}
		if actorID != nil {
			s := actorID.String()
			e.ActorID = &s
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func scanNotification(row pgx.Row) (models.Notification, error) {
	var n models.Notification
	var id, workspaceID uuid.UUID
	var actorID, subjectNodeID *uuid.UUID
	err := row.Scan(&id, &workspaceID, &actorID, &subjectNodeID, &n.Type, &n.Title, &n.Body, &n.Payload, &n.ReadAt, &n.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Notification{}, httperr.ErrNotFound
	}
	if err != nil {
		return models.Notification{}, err
	}
	n.ID = id.String()
	n.WorkspaceID = workspaceID.String()
	if actorID != nil {
		s := actorID.String()
		n.ActorID = &s
	}
	if subjectNodeID != nil {
		s := subjectNodeID.String()
		n.SubjectNodeID = &s
	}
	return n, nil
}
