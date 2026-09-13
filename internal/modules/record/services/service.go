package services

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kashifxyz/flow-server/internal/access"
	"github.com/kashifxyz/flow-server/internal/graph"
	"github.com/kashifxyz/flow-server/internal/modules/record/models"
	"github.com/kashifxyz/flow-server/internal/utils/httperr"
)

const (
	nodeTypeDatabase = "database"
	nodeTypeRecord   = "record"
)

type Service struct {
	DB *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Service {
	return &Service{DB: db}
}

func (s *Service) List(ctx context.Context, databaseID, userID uuid.UUID) ([]models.Record, error) {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, databaseID); err != nil {
		return nil, err
	}
	rows, err := s.DB.Query(ctx, listRecordsSQL, databaseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.Record, 0)
	for rows.Next() {
		r, err := scanRecord(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Service) Create(ctx context.Context, databaseID, userID uuid.UUID, in models.CreateRecordRequest) (models.Record, error) {
	workspaceID, err := access.RequireNodeMember(ctx, s.DB, userID, databaseID)
	if err != nil {
		return models.Record{}, err
	}
	if _, err := graph.GetOfType(ctx, s.DB, databaseID, nodeTypeDatabase); err != nil {
		return models.Record{}, err
	}
	rank := in.Rank
	if rank == "" {
		rank = "0"
	}
	var recordID uuid.UUID
	err = withTx(ctx, s.DB, func(tx pgx.Tx) error {
		n, err := graph.Create(ctx, tx, graph.Insert{
			WorkspaceID: workspaceID,
			ParentID:    &databaseID,
			Type:        nodeTypeRecord,
			Title:       in.Title,
			Rank:        rank,
			CreatedBy:   userID,
		})
		if err != nil {
			return err
		}
		recordID = n.ID
		for fieldIDStr, value := range in.Values {
			fieldID, err := uuid.Parse(fieldIDStr)
			if err != nil {
				return httperr.ErrInvalid
			}
			if err := setFieldValueTx(ctx, tx, databaseID, workspaceID, recordID, fieldID, userID, value); err != nil {
				return err
			}
		}
		_, err = tx.Exec(ctx, `UPDATE databases SET row_count = row_count + 1, updated_at = now() WHERE id = $1`, databaseID)
		return err
	})
	if err != nil {
		return models.Record{}, err
	}
	return s.getRecord(ctx, recordID)
}

func (s *Service) Get(ctx context.Context, recordID, userID uuid.UUID) (models.Record, error) {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, recordID); err != nil {
		return models.Record{}, err
	}
	return s.getRecord(ctx, recordID)
}

func (s *Service) Update(ctx context.Context, recordID, userID uuid.UUID, in models.UpdateRecordRequest) (models.Record, error) {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, recordID); err != nil {
		return models.Record{}, err
	}
	if _, err := graph.GetOfType(ctx, s.DB, recordID, nodeTypeRecord); err != nil {
		return models.Record{}, err
	}
	if _, err := graph.Update(ctx, s.DB, recordID, userID, graph.Patch{Title: in.Title, Rank: in.Rank}); err != nil {
		return models.Record{}, err
	}
	return s.getRecord(ctx, recordID)
}

func (s *Service) Delete(ctx context.Context, recordID, userID uuid.UUID) error {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, recordID); err != nil {
		return err
	}
	n, err := graph.GetOfType(ctx, s.DB, recordID, nodeTypeRecord)
	if err != nil {
		return err
	}
	return withTx(ctx, s.DB, func(tx pgx.Tx) error {
		if err := graph.SoftDelete(ctx, tx, recordID, userID, 0); err != nil {
			return err
		}
		if n.ParentID != nil {
			_, err := tx.Exec(ctx, `UPDATE databases SET row_count = GREATEST(row_count - 1, 0), updated_at = now() WHERE id = $1`, *n.ParentID)
			return err
		}
		return nil
	})
}

func (s *Service) SetFieldValue(ctx context.Context, recordID, fieldID, userID uuid.UUID, in models.SetFieldValueRequest) (models.Record, error) {
	workspaceID, err := access.RequireNodeMember(ctx, s.DB, userID, recordID)
	if err != nil {
		return models.Record{}, err
	}
	n, err := graph.GetOfType(ctx, s.DB, recordID, nodeTypeRecord)
	if err != nil {
		return models.Record{}, err
	}
	if n.ParentID == nil {
		return models.Record{}, httperr.ErrInvalid
	}
	databaseID := *n.ParentID
	err = withTx(ctx, s.DB, func(tx pgx.Tx) error {
		return setFieldValueTx(ctx, tx, databaseID, workspaceID, recordID, fieldID, userID, in.Value)
	})
	if err != nil {
		return models.Record{}, err
	}
	return s.getRecord(ctx, recordID)
}

func (s *Service) ListRevisions(ctx context.Context, recordID, userID uuid.UUID) ([]models.FieldValueRevision, error) {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, recordID); err != nil {
		return nil, err
	}
	rows, err := s.DB.Query(ctx, `
		SELECT id, field_id, actor_id, value, created_at
		FROM field_value_revisions WHERE record_id = $1
		ORDER BY created_at DESC
		LIMIT 200
	`, recordID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.FieldValueRevision, 0)
	for rows.Next() {
		var rev models.FieldValueRevision
		var id, fieldID uuid.UUID
		var actorID *uuid.UUID
		if err := rows.Scan(&id, &fieldID, &actorID, &rev.Value, &rev.CreatedAt); err != nil {
			return nil, err
		}
		rev.ID = id.String()
		rev.FieldID = fieldID.String()
		if actorID != nil {
			a := actorID.String()
			rev.ActorID = &a
		}
		out = append(out, rev)
	}
	return out, rows.Err()
}

func setFieldValueTx(ctx context.Context, tx pgx.Tx, databaseID, workspaceID, recordID, fieldID, userID uuid.UUID, value json.RawMessage) error {
	var isUnique bool
	err := tx.QueryRow(ctx, `SELECT is_unique FROM fields WHERE id = $1 AND database_id = $2`, fieldID, databaseID).Scan(&isUnique)
	if errors.Is(err, pgx.ErrNoRows) {
		return httperr.ErrInvalid
	}
	if err != nil {
		return err
	}
	if len(value) == 0 {
		value = []byte("null")
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO field_values (workspace_id, record_id, field_id, value, updated_at)
		VALUES ($1, $2, $3, $4, now())
		ON CONFLICT (record_id, field_id) DO UPDATE SET value = EXCLUDED.value, updated_at = now()
	`, workspaceID, recordID, fieldID, value); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO field_value_revisions (workspace_id, record_id, field_id, actor_id, value)
		VALUES ($1, $2, $3, $4, $5)
	`, workspaceID, recordID, fieldID, userID, value); err != nil {
		return err
	}
	if isUnique {
		if _, err := tx.Exec(ctx, `DELETE FROM unique_field_values WHERE field_id = $1 AND record_id = $2`, fieldID, recordID); err != nil {
			return err
		}
		valueKey := string(value)
		tag, err := tx.Exec(ctx, `
			INSERT INTO unique_field_values (field_id, value_key, record_id, workspace_id)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (field_id, value_key) DO NOTHING
		`, fieldID, valueKey, recordID, workspaceID)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return httperr.ErrConflict
		}
	}
	dep, err := tx.Query(ctx, `SELECT field_id FROM field_dependencies WHERE depends_on_field_id = $1`, fieldID)
	if err != nil {
		return err
	}
	defer dep.Close()
	var dependents []uuid.UUID
	for dep.Next() {
		var depFieldID uuid.UUID
		if err := dep.Scan(&depFieldID); err != nil {
			return err
		}
		dependents = append(dependents, depFieldID)
	}
	if err := dep.Err(); err != nil {
		return err
	}
	for _, depFieldID := range dependents {
		if _, err := tx.Exec(ctx, `
			INSERT INTO formula_jobs (workspace_id, field_id, record_id, status, reason)
			VALUES ($1, $2, $3, 'queued', 'dependency_changed')
		`, workspaceID, depFieldID, recordID); err != nil {
			return err
		}
	}
	return nil
}

const listRecordsSQL = `
	SELECT n.id, n.parent_id, n.title, n.rank, n.created_at, n.updated_at,
	       COALESCE(
	         (SELECT jsonb_object_agg(fv.field_id::text, fv.value) FROM field_values fv WHERE fv.record_id = n.id),
	         '{}'::jsonb
	       )
	FROM nodes n
	WHERE n.parent_id = $1 AND n.type = 'record' AND n.deleted_at IS NULL
	ORDER BY n.rank, n.created_at
`

func (s *Service) getRecord(ctx context.Context, recordID uuid.UUID) (models.Record, error) {
	r, err := scanRecord(s.DB.QueryRow(ctx, `
		SELECT n.id, n.parent_id, n.title, n.rank, n.created_at, n.updated_at,
		       COALESCE(
		         (SELECT jsonb_object_agg(fv.field_id::text, fv.value) FROM field_values fv WHERE fv.record_id = n.id),
		         '{}'::jsonb
		       )
		FROM nodes n WHERE n.id = $1 AND n.type = 'record' AND n.deleted_at IS NULL
	`, recordID))
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Record{}, httperr.ErrNotFound
	}
	return r, err
}

func scanRecord(row pgx.Row) (models.Record, error) {
	var r models.Record
	var id uuid.UUID
	var parentID *uuid.UUID
	var values json.RawMessage
	err := row.Scan(&id, &parentID, &r.Title, &r.Rank, &r.CreatedAt, &r.UpdatedAt, &values)
	if err != nil {
		return models.Record{}, err
	}
	r.ID = id.String()
	if parentID != nil {
		r.DatabaseID = parentID.String()
	}
	var vals map[string]json.RawMessage
	if err := json.Unmarshal(values, &vals); err != nil {
		return models.Record{}, err
	}
	r.Values = vals
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
