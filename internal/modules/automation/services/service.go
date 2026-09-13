package services

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kashifxyz/flow-server/internal/access"
	"github.com/kashifxyz/flow-server/internal/modules/automation/models"
	"github.com/kashifxyz/flow-server/internal/utils"
	"github.com/kashifxyz/flow-server/internal/utils/httperr"
)

type Service struct {
	DB *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Service {
	return &Service{DB: db}
}

const automationCols = `id, workspace_id, name, description, enabled, trigger_type, trigger, conditions, last_run_at, last_status, run_count, version, created_at, updated_at`

func (s *Service) List(ctx context.Context, workspaceID, userID uuid.UUID) ([]models.Automation, error) {
	if err := access.RequireMember(ctx, s.DB, workspaceID, userID); err != nil {
		return nil, err
	}
	rows, err := s.DB.Query(ctx, `SELECT `+automationCols+` FROM automations WHERE workspace_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.Automation, 0)
	for rows.Next() {
		a, err := scanAutomation(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (s *Service) Create(ctx context.Context, workspaceID, userID uuid.UUID, in models.CreateAutomationRequest) (models.Automation, error) {
	if err := access.RequireMember(ctx, s.DB, workspaceID, userID); err != nil {
		return models.Automation{}, err
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return models.Automation{}, httperr.ErrInvalid
	}
	triggerType := in.TriggerType
	if triggerType == "" {
		triggerType = "record.updated"
	}
	trigger := in.Trigger
	if len(trigger) == 0 {
		trigger = []byte(`{}`)
	}
	conditions := in.Conditions
	if len(conditions) == 0 {
		conditions = []byte(`[]`)
	}
	id, err := utils.NewID()
	if err != nil {
		return models.Automation{}, err
	}
	var a models.Automation
	err = withTx(ctx, s.DB, func(tx pgx.Tx) error {
		var scanErr error
		a, scanErr = scanAutomation(tx.QueryRow(ctx, `
			INSERT INTO automations (id, workspace_id, created_by, updated_by, name, description, trigger_type, trigger, conditions)
			VALUES ($1, $2, $3, $3, $4, $5, $6, $7, $8)
			RETURNING `+automationCols, id, workspaceID, userID, name, in.Description, triggerType, trigger, conditions))
		if scanErr != nil {
			return scanErr
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO automation_versions (automation_id, version, trigger, conditions, created_by)
			VALUES ($1, 1, $2, $3, $4)
		`, id, trigger, conditions, userID); err != nil {
			return err
		}
		for _, step := range in.Steps {
			stepID, err := utils.NewID()
			if err != nil {
				return err
			}
			config := step.Config
			if len(config) == 0 {
				config = []byte(`{}`)
			}
			if _, err := tx.Exec(ctx, `
				INSERT INTO automation_steps (id, automation_id, version, rank, type, config)
				VALUES ($1, $2, 1, $3, $4, $5)
			`, stepID, id, step.Rank, step.Type, config); err != nil {
				return err
			}
		}
		return nil
	})
	return a, err
}

func (s *Service) Get(ctx context.Context, automationID, userID uuid.UUID) (models.Automation, error) {
	a, err := s.getAutomation(ctx, automationID)
	if err != nil {
		return models.Automation{}, err
	}
	if err := access.RequireMember(ctx, s.DB, uuid.MustParse(a.WorkspaceID), userID); err != nil {
		return models.Automation{}, err
	}
	return a, nil
}

func (s *Service) Update(ctx context.Context, automationID, userID uuid.UUID, in models.UpdateAutomationRequest) (models.Automation, error) {
	a, err := s.getAutomation(ctx, automationID)
	if err != nil {
		return models.Automation{}, err
	}
	if err := access.RequireMember(ctx, s.DB, uuid.MustParse(a.WorkspaceID), userID); err != nil {
		return models.Automation{}, err
	}
	return scanAutomation(s.DB.QueryRow(ctx, `
		UPDATE automations SET
			name = COALESCE($2, name),
			description = COALESCE($3, description),
			trigger = COALESCE($4, trigger),
			conditions = COALESCE($5, conditions),
			updated_by = $6,
			version = version + 1,
			updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING `+automationCols, automationID, in.Name, in.Description, in.Trigger, in.Conditions, userID))
}

func (s *Service) Delete(ctx context.Context, automationID, userID uuid.UUID) error {
	a, err := s.getAutomation(ctx, automationID)
	if err != nil {
		return err
	}
	if err := access.RequireMember(ctx, s.DB, uuid.MustParse(a.WorkspaceID), userID); err != nil {
		return err
	}
	tag, err := s.DB.Exec(ctx, `UPDATE automations SET deleted_at = now(), updated_at = now() WHERE id = $1 AND deleted_at IS NULL`, automationID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return httperr.ErrNotFound
	}
	return nil
}

func (s *Service) SetEnabled(ctx context.Context, automationID, userID uuid.UUID, enabled bool) (models.Automation, error) {
	a, err := s.getAutomation(ctx, automationID)
	if err != nil {
		return models.Automation{}, err
	}
	if err := access.RequireMember(ctx, s.DB, uuid.MustParse(a.WorkspaceID), userID); err != nil {
		return models.Automation{}, err
	}
	return scanAutomation(s.DB.QueryRow(ctx, `
		UPDATE automations SET enabled = $2, updated_by = $3, updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING `+automationCols, automationID, enabled, userID))
}

func (s *Service) Run(ctx context.Context, automationID, userID uuid.UUID, in models.RunRequest) (models.Run, error) {
	a, err := s.getAutomation(ctx, automationID)
	if err != nil {
		return models.Run{}, err
	}
	workspaceID := uuid.MustParse(a.WorkspaceID)
	if err := access.RequireMember(ctx, s.DB, workspaceID, userID); err != nil {
		return models.Run{}, err
	}
	payload := in.Payload
	if len(payload) == 0 {
		payload = []byte(`{}`)
	}
	id, err := utils.NewID()
	if err != nil {
		return models.Run{}, err
	}
	var r models.Run
	err = withTx(ctx, s.DB, func(tx pgx.Tx) error {
		var scanErr error
		r, scanErr = scanRun(tx.QueryRow(ctx, `
			INSERT INTO automation_runs (id, automation_id, workspace_id, version, status, trigger_payload)
			VALUES ($1, $2, $3, $4, 'queued', $5)
			RETURNING id, automation_id, status, trigger_payload, error, started_at, finished_at, created_at
		`, id, automationID, workspaceID, a.Version, payload))
		if scanErr != nil {
			return scanErr
		}
		jobID, err := utils.NewID()
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO jobs (id, workspace_id, type, payload) VALUES ($1, $2, 'automation.run', $3)
		`, jobID, workspaceID, []byte(`{"run_id":"`+id.String()+`"}`))
		return err
	})
	return r, err
}

func (s *Service) ListRuns(ctx context.Context, automationID, userID uuid.UUID) ([]models.Run, error) {
	a, err := s.getAutomation(ctx, automationID)
	if err != nil {
		return nil, err
	}
	if err := access.RequireMember(ctx, s.DB, uuid.MustParse(a.WorkspaceID), userID); err != nil {
		return nil, err
	}
	rows, err := s.DB.Query(ctx, `
		SELECT id, automation_id, status, trigger_payload, error, started_at, finished_at, created_at
		FROM automation_runs WHERE automation_id = $1
		ORDER BY created_at DESC
		LIMIT 50
	`, automationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.Run, 0)
	for rows.Next() {
		r, err := scanRun(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Service) getAutomation(ctx context.Context, automationID uuid.UUID) (models.Automation, error) {
	a, err := scanAutomation(s.DB.QueryRow(ctx, `SELECT `+automationCols+` FROM automations WHERE id = $1 AND deleted_at IS NULL`, automationID))
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Automation{}, httperr.ErrNotFound
	}
	return a, err
}

func scanAutomation(row pgx.Row) (models.Automation, error) {
	var a models.Automation
	var id, workspaceID uuid.UUID
	err := row.Scan(&id, &workspaceID, &a.Name, &a.Description, &a.Enabled, &a.TriggerType, &a.Trigger, &a.Conditions,
		&a.LastRunAt, &a.LastStatus, &a.RunCount, &a.Version, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return models.Automation{}, err
	}
	a.ID = id.String()
	a.WorkspaceID = workspaceID.String()
	return a, nil
}

func scanRun(row pgx.Row) (models.Run, error) {
	var r models.Run
	var id, automationID uuid.UUID
	err := row.Scan(&id, &automationID, &r.Status, &r.TriggerPayload, &r.Error, &r.StartedAt, &r.FinishedAt, &r.CreatedAt)
	if err != nil {
		return models.Run{}, err
	}
	r.ID = id.String()
	r.AutomationID = automationID.String()
	return r, nil
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
