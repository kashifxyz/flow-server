package services

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kashifxyz/flow-server/internal/access"
	"github.com/kashifxyz/flow-server/internal/modules/search/models"
	"github.com/kashifxyz/flow-server/internal/utils"
	"github.com/kashifxyz/flow-server/internal/utils/httperr"
)

type Service struct {
	DB *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Service {
	return &Service{DB: db}
}

func (s *Service) Search(ctx context.Context, workspaceID, userID uuid.UUID, query string) ([]models.Result, error) {
	if err := access.RequireMember(ctx, s.DB, workspaceID, userID); err != nil {
		return nil, err
	}
	query = strings.TrimSpace(query)
	if query == "" {
		return []models.Result{}, nil
	}
	rows, err := s.DB.Query(ctx, `
		SELECT node_id, node_type, title, ts_rank(document, websearch_to_tsquery('simple', $2)) AS rank
		FROM search_documents
		WHERE workspace_id = $1 AND document @@ websearch_to_tsquery('simple', $2)
		ORDER BY rank DESC
		LIMIT 50
	`, workspaceID, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.Result, 0)
	for rows.Next() {
		var r models.Result
		var nodeID uuid.UUID
		if err := rows.Scan(&nodeID, &r.Type, &r.Title, &r.Rank); err != nil {
			return nil, err
		}
		r.NodeID = nodeID.String()
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Service) ListSavedSearches(ctx context.Context, workspaceID, userID uuid.UUID) ([]models.SavedSearch, error) {
	if err := access.RequireMember(ctx, s.DB, workspaceID, userID); err != nil {
		return nil, err
	}
	rows, err := s.DB.Query(ctx, `
		SELECT id, name, query, filters, created_at, updated_at
		FROM saved_searches WHERE workspace_id = $1 AND user_id = $2
		ORDER BY created_at DESC
	`, workspaceID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.SavedSearch, 0)
	for rows.Next() {
		var sv models.SavedSearch
		var id uuid.UUID
		if err := rows.Scan(&id, &sv.Name, &sv.Query, &sv.Filters, &sv.CreatedAt, &sv.UpdatedAt); err != nil {
			return nil, err
		}
		sv.ID = id.String()
		out = append(out, sv)
	}
	return out, rows.Err()
}

func (s *Service) CreateSavedSearch(ctx context.Context, workspaceID, userID uuid.UUID, in models.CreateSavedSearchRequest) (models.SavedSearch, error) {
	if err := access.RequireMember(ctx, s.DB, workspaceID, userID); err != nil {
		return models.SavedSearch{}, err
	}
	name := strings.TrimSpace(in.Name)
	query := strings.TrimSpace(in.Query)
	if name == "" || query == "" {
		return models.SavedSearch{}, httperr.ErrInvalid
	}
	filters := in.Filters
	if len(filters) == 0 {
		filters = []byte(`{}`)
	}
	id, err := utils.NewID()
	if err != nil {
		return models.SavedSearch{}, err
	}
	var sv models.SavedSearch
	var savedID uuid.UUID
	err = s.DB.QueryRow(ctx, `
		INSERT INTO saved_searches (id, workspace_id, user_id, name, query, filters)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, name, query, filters, created_at, updated_at
	`, id, workspaceID, userID, name, query, filters).Scan(&savedID, &sv.Name, &sv.Query, &sv.Filters, &sv.CreatedAt, &sv.UpdatedAt)
	if err != nil {
		return models.SavedSearch{}, err
	}
	sv.ID = savedID.String()
	return sv, nil
}

func (s *Service) DeleteSavedSearch(ctx context.Context, searchID, userID uuid.UUID) error {
	tag, err := s.DB.Exec(ctx, `DELETE FROM saved_searches WHERE id = $1 AND user_id = $2`, searchID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return httperr.ErrNotFound
	}
	return nil
}
