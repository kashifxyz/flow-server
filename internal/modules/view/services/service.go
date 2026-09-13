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
	"github.com/kashifxyz/flow-server/internal/modules/view/models"
	"github.com/kashifxyz/flow-server/internal/utils/httperr"
)

const (
	nodeTypeDatabase = "database"
	nodeTypeView     = "view"
)

type Service struct {
	DB *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Service {
	return &Service{DB: db}
}

const viewCols = `id, database_id, type, name, filter, sorts, groups, visible_field_ids, frozen_field_ids,
	column_widths, is_default, is_personal, is_locked, row_height, created_at, updated_at`

func (s *Service) List(ctx context.Context, databaseID, userID uuid.UUID) ([]models.View, error) {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, databaseID); err != nil {
		return nil, err
	}
	rows, err := s.DB.Query(ctx, `SELECT `+viewCols+` FROM views WHERE database_id = $1 ORDER BY created_at`, databaseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.View, 0)
	for rows.Next() {
		v, err := scanView(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (s *Service) Create(ctx context.Context, databaseID, userID uuid.UUID, in models.CreateViewRequest) (models.View, error) {
	workspaceID, err := access.RequireNodeMember(ctx, s.DB, userID, databaseID)
	if err != nil {
		return models.View{}, err
	}
	if _, err := graph.GetOfType(ctx, s.DB, databaseID, nodeTypeDatabase); err != nil {
		return models.View{}, err
	}
	viewType := in.Type
	if viewType == "" {
		viewType = "table"
	}
	if !models.ViewTypes[viewType] {
		return models.View{}, httperr.ErrInvalid
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		name = strings.ToUpper(viewType[:1]) + viewType[1:]
	}
	var v models.View
	err = withTx(ctx, s.DB, func(tx pgx.Tx) error {
		n, err := graph.Create(ctx, tx, graph.Insert{
			WorkspaceID: workspaceID,
			ParentID:    &databaseID,
			Type:        nodeTypeView,
			Title:       name,
			CreatedBy:   userID,
		})
		if err != nil {
			return err
		}
		var ownerID *uuid.UUID
		if in.IsPersonal {
			ownerID = &userID
		}
		v, err = scanView(tx.QueryRow(ctx, `
			INSERT INTO views (id, workspace_id, database_id, type, name, owner_user_id, is_personal)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING `+viewCols, n.ID, workspaceID, databaseID, viewType, name, ownerID, in.IsPersonal))
		return err
	})
	return v, err
}

func (s *Service) Get(ctx context.Context, viewID, userID uuid.UUID) (models.View, error) {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, viewID); err != nil {
		return models.View{}, err
	}
	return s.getView(ctx, viewID)
}

func (s *Service) Update(ctx context.Context, viewID, userID uuid.UUID, in models.UpdateViewRequest) (models.View, error) {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, viewID); err != nil {
		return models.View{}, err
	}
	v, err := scanView(s.DB.QueryRow(ctx, `
		UPDATE views SET
			name = COALESCE($2, name),
			filter = COALESCE($3, filter),
			sorts = COALESCE($4, sorts),
			groups = COALESCE($5, groups),
			visible_field_ids = COALESCE($6, visible_field_ids),
			frozen_field_ids = COALESCE($7, frozen_field_ids),
			column_widths = COALESCE($8, column_widths),
			is_locked = COALESCE($9, is_locked),
			row_height = COALESCE($10, row_height),
			updated_at = now()
		WHERE id = $1
		RETURNING `+viewCols, viewID, in.Name, in.Filter, in.Sorts, in.Groups, uuidStrings(in.VisibleFieldIDs),
		uuidStrings(in.FrozenFieldIDs), in.ColumnWidths, in.IsLocked, in.RowHeight))
	if errors.Is(err, pgx.ErrNoRows) {
		return models.View{}, httperr.ErrNotFound
	}
	if err != nil {
		return models.View{}, err
	}
	if in.Name != nil {
		_, _ = s.DB.Exec(ctx, `UPDATE nodes SET title = $2, updated_at = now() WHERE id = $1`, viewID, *in.Name)
	}
	return v, nil
}

func (s *Service) Delete(ctx context.Context, viewID, userID uuid.UUID) error {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, viewID); err != nil {
		return err
	}
	var isDefault bool
	if err := s.DB.QueryRow(ctx, `SELECT is_default FROM views WHERE id = $1`, viewID).Scan(&isDefault); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return httperr.ErrNotFound
		}
		return err
	}
	if isDefault {
		return httperr.ErrInvalid
	}
	return withTx(ctx, s.DB, func(tx pgx.Tx) error {
		return graph.SoftDelete(ctx, tx, viewID, userID, 0)
	})
}

func (s *Service) getView(ctx context.Context, viewID uuid.UUID) (models.View, error) {
	v, err := scanView(s.DB.QueryRow(ctx, `SELECT `+viewCols+` FROM views WHERE id = $1`, viewID))
	if errors.Is(err, pgx.ErrNoRows) {
		return models.View{}, httperr.ErrNotFound
	}
	return v, err
}

func scanView(row pgx.Row) (models.View, error) {
	var v models.View
	var id, databaseID uuid.UUID
	var visible, frozen []uuid.UUID
	err := row.Scan(&id, &databaseID, &v.Type, &v.Name, &v.Filter, &v.Sorts, &v.Groups, &visible, &frozen,
		&v.ColumnWidths, &v.IsDefault, &v.IsPersonal, &v.IsLocked, &v.RowHeight, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		return models.View{}, err
	}
	v.ID = id.String()
	v.DatabaseID = databaseID.String()
	v.VisibleFieldIDs = stringify(visible)
	v.FrozenFieldIDs = stringify(frozen)
	return v, nil
}

func stringify(ids []uuid.UUID) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, id.String())
	}
	return out
}

func uuidStrings(ids []string) any {
	if ids == nil {
		return nil
	}
	out := make([]uuid.UUID, 0, len(ids))
	for _, s := range ids {
		id, err := uuid.Parse(s)
		if err != nil {
			continue
		}
		out = append(out, id)
	}
	return out
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
