package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kashifxyz/flow-server/internal/access"
	"github.com/kashifxyz/flow-server/internal/graph"
	"github.com/kashifxyz/flow-server/internal/modules/project/models"
	"github.com/kashifxyz/flow-server/internal/utils/httperr"
)

const nodeType = "project"

type Service struct {
	DB *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Service {
	return &Service{DB: db}
}

func (s *Service) List(ctx context.Context, workspaceID, userID uuid.UUID) ([]models.Project, error) {
	if err := access.RequireMember(ctx, s.DB, workspaceID, userID); err != nil {
		return nil, err
	}
	nodes, err := graph.ListByWorkspaceType(ctx, s.DB, workspaceID, nodeType)
	if err != nil {
		return nil, err
	}
	out := make([]models.Project, 0, len(nodes))
	for _, n := range nodes {
		out = append(out, models.FromGraph(n))
	}
	return out, nil
}

func (s *Service) Create(ctx context.Context, workspaceID, userID uuid.UUID, in models.CreateRequest) (models.Project, error) {
	if err := access.RequireMember(ctx, s.DB, workspaceID, userID); err != nil {
		return models.Project{}, err
	}
	if in.Title == "" {
		return models.Project{}, httperr.ErrInvalid
	}
	insert := graph.Insert{
		WorkspaceID: workspaceID,
		Type:        nodeType,
		Title:       in.Title,
		Description: in.Description,
		Props:       in.Props,
		CreatedBy:   userID,
	}
	if in.Icon != "" {
		icon := in.Icon
		insert.Icon = &icon
	}
	if in.SpaceID != "" {
		id, err := uuid.Parse(in.SpaceID)
		if err != nil {
			return models.Project{}, httperr.ErrInvalid
		}
		insert.SpaceID = &id
	}
	var n graph.Node
	err := withTx(ctx, s.DB, func(tx pgx.Tx) error {
		var err error
		n, err = graph.Create(ctx, tx, insert)
		return err
	})
	if err != nil {
		return models.Project{}, err
	}
	return models.FromGraph(n), nil
}

func (s *Service) Get(ctx context.Context, projectID, userID uuid.UUID) (models.Project, error) {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, projectID); err != nil {
		return models.Project{}, err
	}
	n, err := graph.GetOfType(ctx, s.DB, projectID, nodeType)
	if err != nil {
		return models.Project{}, err
	}
	return models.FromGraph(n), nil
}

func (s *Service) Update(ctx context.Context, projectID, userID uuid.UUID, in models.UpdateRequest) (models.Project, error) {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, projectID); err != nil {
		return models.Project{}, err
	}
	if _, err := graph.GetOfType(ctx, s.DB, projectID, nodeType); err != nil {
		return models.Project{}, err
	}
	n, err := graph.Update(ctx, s.DB, projectID, userID, graph.Patch{
		Title:       in.Title,
		Description: in.Description,
		Icon:        in.Icon,
		Props:       in.Props,
	})
	if err != nil {
		return models.Project{}, err
	}
	return models.FromGraph(n), nil
}

func (s *Service) Delete(ctx context.Context, projectID, userID uuid.UUID) error {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, projectID); err != nil {
		return err
	}
	if _, err := graph.GetOfType(ctx, s.DB, projectID, nodeType); err != nil {
		return err
	}
	return withTx(ctx, s.DB, func(tx pgx.Tx) error {
		return graph.SoftDelete(ctx, tx, projectID, userID, 0)
	})
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
