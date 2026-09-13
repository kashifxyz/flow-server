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
	"github.com/kashifxyz/flow-server/internal/modules/database/models"
	"github.com/kashifxyz/flow-server/internal/utils"
	"github.com/kashifxyz/flow-server/internal/utils/httperr"
)

const (
	nodeTypeDatabase = "database"
	nodeTypeField    = "field"
)

type Service struct {
	DB *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Service {
	return &Service{DB: db}
}

func (s *Service) List(ctx context.Context, workspaceID, userID uuid.UUID) ([]models.Database, error) {
	if err := access.RequireMember(ctx, s.DB, workspaceID, userID); err != nil {
		return nil, err
	}
	rows, err := s.DB.Query(ctx, `
		SELECT d.id, d.workspace_id, n.title, d.description, d.is_inline, d.primary_field_id, d.row_count, d.created_at, d.updated_at
		FROM databases d JOIN nodes n ON n.id = d.id
		WHERE d.workspace_id = $1 AND n.deleted_at IS NULL
		ORDER BY n.created_at
	`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.Database, 0)
	for rows.Next() {
		d, err := scanDatabase(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (s *Service) Create(ctx context.Context, workspaceID, userID uuid.UUID, in models.CreateDatabaseRequest) (models.Database, error) {
	if err := access.RequireMember(ctx, s.DB, workspaceID, userID); err != nil {
		return models.Database{}, err
	}
	title := strings.TrimSpace(in.Title)
	if title == "" {
		return models.Database{}, httperr.ErrInvalid
	}
	insert := graph.Insert{
		WorkspaceID: workspaceID,
		Type:        nodeTypeDatabase,
		Title:       title,
		CreatedBy:   userID,
	}
	if in.SpaceID != "" {
		id, err := uuid.Parse(in.SpaceID)
		if err != nil {
			return models.Database{}, httperr.ErrInvalid
		}
		insert.SpaceID = &id
	}
	if in.ParentID != "" {
		id, err := uuid.Parse(in.ParentID)
		if err != nil {
			return models.Database{}, httperr.ErrInvalid
		}
		insert.ParentID = &id
	}
	var d models.Database
	err := withTx(ctx, s.DB, func(tx pgx.Tx) error {
		n, err := graph.Create(ctx, tx, insert)
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO databases (id, workspace_id, is_inline) VALUES ($1, $2, $3)
		`, n.ID, workspaceID, in.IsInline)
		if err != nil {
			return err
		}
		d = models.Database{
			ID:          n.ID.String(),
			WorkspaceID: workspaceID.String(),
			Title:       n.Title,
			IsInline:    in.IsInline,
			CreatedAt:   n.CreatedAt,
			UpdatedAt:   n.UpdatedAt,
		}
		return nil
	})
	return d, err
}

func (s *Service) Get(ctx context.Context, databaseID, userID uuid.UUID) (models.Database, error) {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, databaseID); err != nil {
		return models.Database{}, err
	}
	return s.getDatabase(ctx, databaseID)
}

func (s *Service) Update(ctx context.Context, databaseID, userID uuid.UUID, in models.UpdateDatabaseRequest) (models.Database, error) {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, databaseID); err != nil {
		return models.Database{}, err
	}
	if in.Title != nil {
		if _, err := graph.Update(ctx, s.DB, databaseID, userID, graph.Patch{Title: in.Title}); err != nil {
			return models.Database{}, err
		}
	}
	var primaryFieldID *uuid.UUID
	if in.PrimaryFieldID != nil && *in.PrimaryFieldID != "" {
		id, err := uuid.Parse(*in.PrimaryFieldID)
		if err != nil {
			return models.Database{}, httperr.ErrInvalid
		}
		primaryFieldID = &id
	}
	tag, err := s.DB.Exec(ctx, `
		UPDATE databases SET
			description = COALESCE($2, description),
			primary_field_id = COALESCE($3, primary_field_id),
			updated_at = now()
		WHERE id = $1
	`, databaseID, in.Description, primaryFieldID)
	if err != nil {
		return models.Database{}, err
	}
	if tag.RowsAffected() == 0 {
		return models.Database{}, httperr.ErrNotFound
	}
	return s.getDatabase(ctx, databaseID)
}

func (s *Service) Delete(ctx context.Context, databaseID, userID uuid.UUID) error {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, databaseID); err != nil {
		return err
	}
	if _, err := graph.GetOfType(ctx, s.DB, databaseID, nodeTypeDatabase); err != nil {
		return err
	}
	return withTx(ctx, s.DB, func(tx pgx.Tx) error {
		return graph.SoftDelete(ctx, tx, databaseID, userID, 0)
	})
}

func (s *Service) ListFields(ctx context.Context, databaseID, userID uuid.UUID) ([]models.Field, error) {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, databaseID); err != nil {
		return nil, err
	}
	rows, err := s.DB.Query(ctx, fieldCols+` FROM fields WHERE database_id = $1 ORDER BY rank`, databaseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.Field, 0)
	for rows.Next() {
		f, err := scanField(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

func (s *Service) CreateField(ctx context.Context, databaseID, userID uuid.UUID, in models.CreateFieldRequest) (models.Field, error) {
	workspaceID, err := access.RequireNodeMember(ctx, s.DB, userID, databaseID)
	if err != nil {
		return models.Field{}, err
	}
	if _, err := graph.GetOfType(ctx, s.DB, databaseID, nodeTypeDatabase); err != nil {
		return models.Field{}, err
	}
	name := strings.TrimSpace(in.Name)
	if name == "" || !models.FieldTypes[in.Type] {
		return models.Field{}, httperr.ErrInvalid
	}
	rank := in.Rank
	if rank == "" {
		rank = "0"
	}
	opts := in.Options
	if len(opts) == 0 {
		opts = []byte(`{}`)
	}
	var f models.Field
	err = withTx(ctx, s.DB, func(tx pgx.Tx) error {
		n, err := graph.Create(ctx, tx, graph.Insert{
			WorkspaceID: workspaceID,
			ParentID:    &databaseID,
			Type:        nodeTypeField,
			Title:       name,
			CreatedBy:   userID,
		})
		if err != nil {
			return err
		}
		var existing int
		if err := tx.QueryRow(ctx, `SELECT count(*) FROM fields WHERE database_id = $1`, databaseID).Scan(&existing); err != nil {
			return err
		}
		isPrimary := existing == 0
		var formula *string
		if in.Formula != "" {
			formula = &in.Formula
		}
		var id, dbID uuid.UUID
		if err := tx.QueryRow(ctx, `
			INSERT INTO fields (id, workspace_id, database_id, name, type, description, options, is_required, is_unique, is_hidden, is_primary, formula, is_computed, rank)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
			RETURNING id, database_id, name, type, description, options, is_primary, is_required, is_unique, is_computed, is_hidden, formula, rank, created_at, updated_at
		`, n.ID, workspaceID, databaseID, name, in.Type, in.Description, opts, in.IsRequired, in.IsUnique, in.IsHidden,
			isPrimary, formula, formula != nil, rank).Scan(
			&id, &dbID, &f.Name, &f.Type, &f.Description, &f.Options, &f.IsPrimary, &f.IsRequired,
			&f.IsUnique, &f.IsComputed, &f.IsHidden, &f.Formula, &f.Rank, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return err
		}
		f.ID = id.String()
		f.DatabaseID = dbID.String()
		if isPrimary {
			if _, err := tx.Exec(ctx, `UPDATE databases SET primary_field_id = $2, updated_at = now() WHERE id = $1`, databaseID, n.ID); err != nil {
				return err
			}
		}
		return nil
	})
	return f, err
}

func (s *Service) UpdateField(ctx context.Context, fieldID, userID uuid.UUID, in models.UpdateFieldRequest) (models.Field, error) {
	databaseID, _, err := s.fieldLocation(ctx, fieldID)
	if err != nil {
		return models.Field{}, err
	}
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, databaseID); err != nil {
		return models.Field{}, err
	}
	tag, err := s.DB.Exec(ctx, `
		UPDATE fields SET
			name = COALESCE($2, name),
			description = COALESCE($3, description),
			options = COALESCE($4, options),
			is_required = COALESCE($5, is_required),
			is_unique = COALESCE($6, is_unique),
			is_hidden = COALESCE($7, is_hidden),
			formula = COALESCE($8, formula),
			rank = COALESCE($9, rank),
			updated_at = now()
		WHERE id = $1
	`, fieldID, in.Name, in.Description, in.Options, in.IsRequired, in.IsUnique, in.IsHidden, in.Formula, in.Rank)
	if err != nil {
		return models.Field{}, err
	}
	if tag.RowsAffected() == 0 {
		return models.Field{}, httperr.ErrNotFound
	}
	if in.Name != nil {
		if _, err := s.DB.Exec(ctx, `UPDATE nodes SET title = $2, updated_at = now() WHERE id = $1`, fieldID, *in.Name); err != nil {
			return models.Field{}, err
		}
	}
	return scanField(s.DB.QueryRow(ctx, fieldCols+` FROM fields WHERE id = $1`, fieldID))
}

func (s *Service) DeleteField(ctx context.Context, fieldID, userID uuid.UUID) error {
	databaseID, _, err := s.fieldLocation(ctx, fieldID)
	if err != nil {
		return err
	}
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, databaseID); err != nil {
		return err
	}
	return withTx(ctx, s.DB, func(tx pgx.Tx) error {
		return graph.SoftDelete(ctx, tx, fieldID, userID, 0)
	})
}

func (s *Service) ListFieldOptions(ctx context.Context, fieldID, userID uuid.UUID) ([]models.FieldOption, error) {
	databaseID, _, err := s.fieldLocation(ctx, fieldID)
	if err != nil {
		return nil, err
	}
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, databaseID); err != nil {
		return nil, err
	}
	rows, err := s.DB.Query(ctx, `
		SELECT id, field_id, key, label, color, rank, disabled_at, created_at
		FROM field_options WHERE field_id = $1 ORDER BY rank
	`, fieldID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.FieldOption, 0)
	for rows.Next() {
		o, err := scanFieldOption(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

func (s *Service) CreateFieldOption(ctx context.Context, fieldID, userID uuid.UUID, in models.CreateFieldOptionRequest) (models.FieldOption, error) {
	databaseID, workspaceID, err := s.fieldLocation(ctx, fieldID)
	if err != nil {
		return models.FieldOption{}, err
	}
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, databaseID); err != nil {
		return models.FieldOption{}, err
	}
	key := strings.TrimSpace(in.Key)
	label := strings.TrimSpace(in.Label)
	if key == "" || label == "" {
		return models.FieldOption{}, httperr.ErrInvalid
	}
	rank := in.Rank
	if rank == "" {
		rank = "0"
	}
	id, err := utils.NewID()
	if err != nil {
		return models.FieldOption{}, err
	}
	return scanFieldOption(s.DB.QueryRow(ctx, `
		INSERT INTO field_options (id, field_id, workspace_id, key, label, color, rank)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, field_id, key, label, color, rank, disabled_at, created_at
	`, id, fieldID, workspaceID, key, label, nullIfEmpty(in.Color), rank))
}

func (s *Service) UpdateFieldOption(ctx context.Context, optionID, userID uuid.UUID, in models.UpdateFieldOptionRequest) (models.FieldOption, error) {
	var fieldID uuid.UUID
	if err := s.DB.QueryRow(ctx, `SELECT field_id FROM field_options WHERE id = $1`, optionID).Scan(&fieldID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.FieldOption{}, httperr.ErrNotFound
		}
		return models.FieldOption{}, err
	}
	databaseID, _, err := s.fieldLocation(ctx, fieldID)
	if err != nil {
		return models.FieldOption{}, err
	}
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, databaseID); err != nil {
		return models.FieldOption{}, err
	}
	row := s.DB.QueryRow(ctx, `
		UPDATE field_options SET
			label = COALESCE($2, label),
			color = COALESCE($3, color),
			rank = COALESCE($4, rank),
			disabled_at = CASE
				WHEN $5::boolean IS NULL THEN disabled_at
				WHEN $5::boolean THEN now()
				ELSE NULL
			END
		WHERE id = $1
		RETURNING id, field_id, key, label, color, rank, disabled_at, created_at
	`, optionID, in.Label, in.Color, in.Rank, in.Disabled)
	return scanFieldOption(row)
}

func (s *Service) DeleteFieldOption(ctx context.Context, optionID, userID uuid.UUID) error {
	var fieldID uuid.UUID
	if err := s.DB.QueryRow(ctx, `SELECT field_id FROM field_options WHERE id = $1`, optionID).Scan(&fieldID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return httperr.ErrNotFound
		}
		return err
	}
	databaseID, _, err := s.fieldLocation(ctx, fieldID)
	if err != nil {
		return err
	}
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, databaseID); err != nil {
		return err
	}
	tag, err := s.DB.Exec(ctx, `DELETE FROM field_options WHERE id = $1`, optionID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return httperr.ErrNotFound
	}
	return nil
}

func (s *Service) fieldLocation(ctx context.Context, fieldID uuid.UUID) (databaseID, workspaceID uuid.UUID, err error) {
	err = s.DB.QueryRow(ctx, `SELECT database_id, workspace_id FROM fields WHERE id = $1`, fieldID).Scan(&databaseID, &workspaceID)
	if errors.Is(err, pgx.ErrNoRows) {
		err = httperr.ErrNotFound
	}
	return
}

func (s *Service) getDatabase(ctx context.Context, databaseID uuid.UUID) (models.Database, error) {
	d, err := scanDatabase(s.DB.QueryRow(ctx, `
		SELECT d.id, d.workspace_id, n.title, d.description, d.is_inline, d.primary_field_id, d.row_count, d.created_at, d.updated_at
		FROM databases d JOIN nodes n ON n.id = d.id
		WHERE d.id = $1 AND n.deleted_at IS NULL
	`, databaseID))
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Database{}, httperr.ErrNotFound
	}
	return d, err
}

func scanDatabase(row pgx.Row) (models.Database, error) {
	var d models.Database
	var id, workspaceID uuid.UUID
	var primaryFieldID *uuid.UUID
	err := row.Scan(&id, &workspaceID, &d.Title, &d.Description, &d.IsInline, &primaryFieldID, &d.RowCount, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return models.Database{}, err
	}
	d.ID = id.String()
	d.WorkspaceID = workspaceID.String()
	if primaryFieldID != nil {
		s := primaryFieldID.String()
		d.PrimaryFieldID = &s
	}
	return d, nil
}

const fieldCols = `SELECT id, database_id, name, type, description, options, is_primary, is_required, is_unique, is_computed, is_hidden, formula, rank, created_at, updated_at`

func scanField(row pgx.Row) (models.Field, error) {
	var f models.Field
	var id, databaseID uuid.UUID
	err := row.Scan(&id, &databaseID, &f.Name, &f.Type, &f.Description, &f.Options, &f.IsPrimary, &f.IsRequired,
		&f.IsUnique, &f.IsComputed, &f.IsHidden, &f.Formula, &f.Rank, &f.CreatedAt, &f.UpdatedAt)
	if err != nil {
		return models.Field{}, err
	}
	f.ID = id.String()
	f.DatabaseID = databaseID.String()
	return f, nil
}

func scanFieldOption(row pgx.Row) (models.FieldOption, error) {
	var o models.FieldOption
	var id, fieldID uuid.UUID
	err := row.Scan(&id, &fieldID, &o.Key, &o.Label, &o.Color, &o.Rank, &o.DisabledAt, &o.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.FieldOption{}, httperr.ErrNotFound
	}
	if err != nil {
		return models.FieldOption{}, err
	}
	o.ID = id.String()
	o.FieldID = fieldID.String()
	return o, nil
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
