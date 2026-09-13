package services

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kashifxyz/flow-server/internal/access"
	"github.com/kashifxyz/flow-server/internal/auth"
	"github.com/kashifxyz/flow-server/internal/modules/webhook/models"
	"github.com/kashifxyz/flow-server/internal/utils"
	"github.com/kashifxyz/flow-server/internal/utils/httperr"
)

type Service struct {
	DB *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Service {
	return &Service{DB: db}
}

const webhookCols = `id, workspace_id, name, url, events, enabled, failure_count, last_success_at, last_failure_at, created_at, updated_at`

func (s *Service) List(ctx context.Context, workspaceID, userID uuid.UUID) ([]models.Webhook, error) {
	if err := access.RequireMember(ctx, s.DB, workspaceID, userID); err != nil {
		return nil, err
	}
	rows, err := s.DB.Query(ctx, `SELECT `+webhookCols+` FROM webhooks WHERE workspace_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.Webhook, 0)
	for rows.Next() {
		wh, err := scanWebhook(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, wh)
	}
	return out, rows.Err()
}

func (s *Service) Create(ctx context.Context, workspaceID, userID uuid.UUID, in models.CreateWebhookRequest) (models.CreateWebhookResponse, error) {
	if err := access.RequireMember(ctx, s.DB, workspaceID, userID); err != nil {
		return models.CreateWebhookResponse{}, err
	}
	name := strings.TrimSpace(in.Name)
	url := strings.TrimSpace(in.URL)
	if name == "" || !strings.HasPrefix(url, "https://") {
		return models.CreateWebhookResponse{}, httperr.ErrInvalid
	}
	secret, err := auth.RandomToken()
	if err != nil {
		return models.CreateWebhookResponse{}, err
	}
	id, err := utils.NewID()
	if err != nil {
		return models.CreateWebhookResponse{}, err
	}
	wh, err := scanWebhook(s.DB.QueryRow(ctx, `
		INSERT INTO webhooks (id, workspace_id, created_by, name, url, secret_hash, events)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING `+webhookCols, id, workspaceID, userID, name, url, auth.HashToken(secret), in.Events))
	if err != nil {
		return models.CreateWebhookResponse{}, err
	}
	return models.CreateWebhookResponse{Webhook: wh, Secret: secret}, nil
}

func (s *Service) Get(ctx context.Context, webhookID, userID uuid.UUID) (models.Webhook, error) {
	wh, err := s.getWebhook(ctx, webhookID)
	if err != nil {
		return models.Webhook{}, err
	}
	if err := access.RequireMember(ctx, s.DB, uuid.MustParse(wh.WorkspaceID), userID); err != nil {
		return models.Webhook{}, err
	}
	return wh, nil
}

func (s *Service) Update(ctx context.Context, webhookID, userID uuid.UUID, in models.UpdateWebhookRequest) (models.Webhook, error) {
	wh, err := s.getWebhook(ctx, webhookID)
	if err != nil {
		return models.Webhook{}, err
	}
	if err := access.RequireMember(ctx, s.DB, uuid.MustParse(wh.WorkspaceID), userID); err != nil {
		return models.Webhook{}, err
	}
	if in.URL != nil && !strings.HasPrefix(*in.URL, "https://") {
		return models.Webhook{}, httperr.ErrInvalid
	}
	return scanWebhook(s.DB.QueryRow(ctx, `
		UPDATE webhooks SET
			name = COALESCE($2, name),
			url = COALESCE($3, url),
			events = COALESCE($4, events),
			enabled = COALESCE($5, enabled),
			updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING `+webhookCols, webhookID, in.Name, in.URL, in.Events, in.Enabled))
}

func (s *Service) Delete(ctx context.Context, webhookID, userID uuid.UUID) error {
	wh, err := s.getWebhook(ctx, webhookID)
	if err != nil {
		return err
	}
	if err := access.RequireMember(ctx, s.DB, uuid.MustParse(wh.WorkspaceID), userID); err != nil {
		return err
	}
	tag, err := s.DB.Exec(ctx, `UPDATE webhooks SET deleted_at = now(), updated_at = now() WHERE id = $1 AND deleted_at IS NULL`, webhookID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return httperr.ErrNotFound
	}
	return nil
}

func (s *Service) ListDeliveries(ctx context.Context, webhookID, userID uuid.UUID) ([]models.Delivery, error) {
	wh, err := s.getWebhook(ctx, webhookID)
	if err != nil {
		return nil, err
	}
	if err := access.RequireMember(ctx, s.DB, uuid.MustParse(wh.WorkspaceID), userID); err != nil {
		return nil, err
	}
	rows, err := s.DB.Query(ctx, `
		SELECT id, event_type, payload, status, attempt_count, response_status, delivered_at, created_at
		FROM webhook_deliveries WHERE webhook_id = $1
		ORDER BY created_at DESC
		LIMIT 50
	`, webhookID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.Delivery, 0)
	for rows.Next() {
		var d models.Delivery
		var id uuid.UUID
		if err := rows.Scan(&id, &d.EventType, &d.Payload, &d.Status, &d.AttemptCount, &d.ResponseStatus, &d.DeliveredAt, &d.CreatedAt); err != nil {
			return nil, err
		}
		d.ID = id.String()
		out = append(out, d)
	}
	return out, rows.Err()
}

func (s *Service) getWebhook(ctx context.Context, webhookID uuid.UUID) (models.Webhook, error) {
	wh, err := scanWebhook(s.DB.QueryRow(ctx, `SELECT `+webhookCols+` FROM webhooks WHERE id = $1 AND deleted_at IS NULL`, webhookID))
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Webhook{}, httperr.ErrNotFound
	}
	return wh, err
}

func scanWebhook(row pgx.Row) (models.Webhook, error) {
	var wh models.Webhook
	var id, workspaceID uuid.UUID
	err := row.Scan(&id, &workspaceID, &wh.Name, &wh.URL, &wh.Events, &wh.Enabled, &wh.FailureCount,
		&wh.LastSuccessAt, &wh.LastFailureAt, &wh.CreatedAt, &wh.UpdatedAt)
	if err != nil {
		return models.Webhook{}, err
	}
	wh.ID = id.String()
	wh.WorkspaceID = workspaceID.String()
	return wh, nil
}
